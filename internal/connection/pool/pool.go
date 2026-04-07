package pool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"MystiSql/internal/connection"
	"MystiSql/internal/connection/monitor"
	"MystiSql/pkg/types"
)

type connectionWrapper struct {
	conn       connection.Connection
	pool       *ConnectionPoolImpl
	createdAt  time.Time
	lastUsedAt time.Time
	instance   *types.DatabaseInstance
}

type ConnectionPoolImpl struct {
	config            *connection.PoolConfig
	factory           connection.ConnectionFactory
	instance          *types.DatabaseInstance
	idleConnections   chan *connectionWrapper
	allConnections    []*connectionWrapper
	mu                sync.RWMutex
	stats             connection.PoolStats
	closed            bool
	healthCheckTicker *time.Ticker
	ctx               context.Context
	cancel            context.CancelFunc

	metricsCollector *monitor.Collector
	instanceName     string
}

type PoolOption func(*ConnectionPoolImpl)

func WithMetricsCollector(collector *monitor.Collector) PoolOption {
	return func(p *ConnectionPoolImpl) {
		p.metricsCollector = collector
	}
}

func NewConnectionPool(
	instance *types.DatabaseInstance,
	factory connection.ConnectionFactory,
	config *connection.PoolConfig,
	opts ...PoolOption,
) (connection.ConnectionPool, error) {
	if config.MaxConnections <= 0 {
		config.MaxConnections = 10
	}
	if config.MinConnections < 0 {
		config.MinConnections = 0
	}
	if config.MinConnections > config.MaxConnections {
		config.MinConnections = config.MaxConnections
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &ConnectionPoolImpl{
		config:          config,
		factory:         factory,
		instance:        instance,
		instanceName:    instance.Name,
		idleConnections: make(chan *connectionWrapper, config.MaxConnections),
		allConnections:  make([]*connectionWrapper, 0, config.MaxConnections),
		stats: connection.PoolStats{
			MaxConnections: config.MaxConnections,
			MinConnections: config.MinConnections,
		},
		closed:           false,
		ctx:              ctx,
		cancel:           cancel,
		metricsCollector: monitor.DefaultCollector(),
	}

	for _, opt := range opts {
		opt(pool)
	}

	if config.MinConnections > 0 {
		if err := pool.initializeMinConnections(); err != nil {
			cancel()
			return nil, fmt.Errorf("failed to initialize min connections: %w", err)
		}
	}

	pool.startHealthCheck()

	return pool, nil
}

func (p *ConnectionPoolImpl) initializeMinConnections() error {
	for i := 0; i < p.config.MinConnections; i++ {
		conn, err := p.createConnection()
		if err != nil {
			return err
		}
		p.idleConnections <- conn
	}
	return nil
}

func (p *ConnectionPoolImpl) createConnection() (*connectionWrapper, error) {
	start := time.Now()
	conn, err := p.factory.CreateConnection(p.instance)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(p.ctx, 10*time.Second)
	defer cancel()
	if err := conn.Connect(ctx); err != nil {
		conn.Close()
		return nil, err
	}

	wrapper := &connectionWrapper{
		conn:       conn,
		pool:       p,
		createdAt:  time.Now(),
		lastUsedAt: time.Now(),
		instance:   p.instance,
	}

	p.mu.Lock()
	p.allConnections = append(p.allConnections, wrapper)
	p.stats.TotalConnections++
	p.mu.Unlock()

	if p.metricsCollector != nil {
		p.metricsCollector.RecordConnectionCreated(p.instanceName)
		p.metricsCollector.UpdatePoolStats(p.instanceName, &connection.PoolStats{
			TotalConnections:  p.stats.TotalConnections,
			IdleConnections:   p.stats.IdleConnections,
			ActiveConnections: p.stats.ActiveConnections,
			MaxConnections:    p.config.MaxConnections,
			MinConnections:    p.config.MinConnections,
		})
	}

	if p.metricsCollector != nil {
		p.metricsCollector.RecordAcquire(p.instanceName, time.Since(start).Nanoseconds(), true)
	}

	return wrapper, nil
}

func (p *ConnectionPoolImpl) GetConnection(ctx context.Context) (connection.Connection, error) {
	const maxRetries = 3

	for retry := 0; retry < maxRetries; retry++ {
		start := time.Now()

		p.mu.Lock()
		if p.closed {
			p.mu.Unlock()
			return nil, fmt.Errorf("connection pool is closed")
		}

		p.stats.AcquireCount++
		p.mu.Unlock()

		var waitStart time.Time

		select {
		case conn := <-p.idleConnections:
			p.mu.Lock()
			p.stats.ActiveConnections++
			p.stats.IdleConnections--
			p.mu.Unlock()

			if err := conn.conn.Ping(ctx); err != nil {
				p.mu.Lock()
				p.stats.AcquireFailed++
				p.stats.ActiveConnections--
				p.mu.Unlock()

				conn.conn.Close()
				p.removeConnectionFromList(conn)

				if p.metricsCollector != nil {
					p.metricsCollector.RecordAcquire(p.instanceName, time.Since(start).Nanoseconds(), false)
					p.metricsCollector.RecordConnectionClosed(p.instanceName)
				}

				continue
			}

			conn.lastUsedAt = time.Now()

			if p.metricsCollector != nil {
				p.metricsCollector.RecordAcquire(p.instanceName, time.Since(start).Nanoseconds(), true)
				p.metricsCollector.UpdatePoolStats(p.instanceName, &connection.PoolStats{
					TotalConnections:  p.stats.TotalConnections,
					IdleConnections:   p.stats.IdleConnections,
					ActiveConnections: p.stats.ActiveConnections,
					MaxConnections:    p.config.MaxConnections,
					MinConnections:    p.config.MinConnections,
				})
			}

			return conn, nil

		default:
			p.mu.Lock()
			totalConnections := p.stats.TotalConnections
			p.mu.Unlock()

			if totalConnections >= p.config.MaxConnections {
				waitStart = time.Now()

				select {
				case conn := <-p.idleConnections:
					p.mu.Lock()
					p.stats.ActiveConnections++
					p.stats.IdleConnections--
					p.mu.Unlock()

					if err := conn.conn.Ping(ctx); err != nil {
						p.mu.Lock()
						p.stats.AcquireFailed++
						p.stats.ActiveConnections--
						p.mu.Unlock()

						conn.conn.Close()
						p.removeConnectionFromList(conn)

						if p.metricsCollector != nil {
							p.metricsCollector.RecordAcquire(p.instanceName, time.Since(start).Nanoseconds(), false)
							p.metricsCollector.RecordConnectionClosed(p.instanceName)
						}

						continue
					}

					conn.lastUsedAt = time.Now()

					if p.metricsCollector != nil {
						p.metricsCollector.RecordWait(p.instanceName, time.Since(waitStart).Nanoseconds())
						p.metricsCollector.RecordAcquire(p.instanceName, time.Since(start).Nanoseconds(), true)
						p.metricsCollector.UpdatePoolStats(p.instanceName, &connection.PoolStats{
							TotalConnections:  p.stats.TotalConnections,
							IdleConnections:   p.stats.IdleConnections,
							ActiveConnections: p.stats.ActiveConnections,
							MaxConnections:    p.config.MaxConnections,
							MinConnections:    p.config.MinConnections,
						})
					}

					return conn, nil

				case <-ctx.Done():
					p.mu.Lock()
					p.stats.AcquireFailed++
					p.mu.Unlock()

					if p.metricsCollector != nil {
						p.metricsCollector.RecordWait(p.instanceName, time.Since(waitStart).Nanoseconds())
						p.metricsCollector.RecordAcquire(p.instanceName, time.Since(start).Nanoseconds(), false)
					}

					return nil, ctx.Err()
				}
			}

			conn, err := p.createConnection()
			if err != nil {
				p.mu.Lock()
				p.stats.AcquireFailed++
				p.mu.Unlock()

				if p.metricsCollector != nil {
					p.metricsCollector.RecordAcquire(p.instanceName, time.Since(start).Nanoseconds(), false)
				}

				return nil, err
			}

			p.mu.Lock()
			p.stats.ActiveConnections++
			p.mu.Unlock()

			if p.metricsCollector != nil {
				p.metricsCollector.UpdatePoolStats(p.instanceName, &connection.PoolStats{
					TotalConnections:  p.stats.TotalConnections,
					IdleConnections:   p.stats.IdleConnections,
					ActiveConnections: p.stats.ActiveConnections,
					MaxConnections:    p.config.MaxConnections,
					MinConnections:    p.config.MinConnections,
				})
			}

			return conn, nil
		}
	}

	return nil, fmt.Errorf("failed to get connection after %d retries", maxRetries)
}

func (p *ConnectionPoolImpl) removeConnectionFromList(wrapper *connectionWrapper) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, c := range p.allConnections {
		if c == wrapper {
			p.allConnections = append(p.allConnections[:i], p.allConnections[i+1:]...)
			break
		}
	}
}

func (p *ConnectionPoolImpl) ReturnConnection(conn connection.Connection) {
	if p.closed {
		conn.Close()
		return
	}

	wrapper, ok := conn.(*connectionWrapper)
	if !ok {
		conn.Close()
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	ctx, cancel := context.WithTimeout(p.ctx, 5*time.Second)
	defer cancel()
	if err := wrapper.conn.Ping(ctx); err != nil {
		wrapper.conn.Close()
		p.stats.ActiveConnections--
		p.stats.TotalConnections--

		for i, c := range p.allConnections {
			if c == wrapper {
				p.allConnections = append(p.allConnections[:i], p.allConnections[i+1:]...)
				break
			}
		}

		if p.metricsCollector != nil {
			p.metricsCollector.RecordConnectionClosed(p.instanceName)
			p.metricsCollector.UpdatePoolStats(p.instanceName, &connection.PoolStats{
				TotalConnections:  p.stats.TotalConnections,
				IdleConnections:   p.stats.IdleConnections,
				ActiveConnections: p.stats.ActiveConnections,
				MaxConnections:    p.config.MaxConnections,
				MinConnections:    p.config.MinConnections,
			})
		}
		return
	}

	maxLifetime, _ := time.ParseDuration(p.config.MaxLifetime)
	if maxLifetime > 0 && time.Since(wrapper.createdAt) > maxLifetime {
		wrapper.conn.Close()
		p.stats.ActiveConnections--
		p.stats.TotalConnections--

		for i, c := range p.allConnections {
			if c == wrapper {
				p.allConnections = append(p.allConnections[:i], p.allConnections[i+1:]...)
				break
			}
		}

		if p.metricsCollector != nil {
			p.metricsCollector.RecordConnectionClosed(p.instanceName)
			p.metricsCollector.UpdatePoolStats(p.instanceName, &connection.PoolStats{
				TotalConnections:  p.stats.TotalConnections,
				IdleConnections:   p.stats.IdleConnections,
				ActiveConnections: p.stats.ActiveConnections,
				MaxConnections:    p.config.MaxConnections,
				MinConnections:    p.config.MinConnections,
			})
		}
		return
	}

	p.stats.ActiveConnections--
	p.stats.IdleConnections++
	p.stats.ReleaseCount++

	select {
	case p.idleConnections <- wrapper:
		if p.metricsCollector != nil {
			p.metricsCollector.UpdatePoolStats(p.instanceName, &connection.PoolStats{
				TotalConnections:  p.stats.TotalConnections,
				IdleConnections:   p.stats.IdleConnections,
				ActiveConnections: p.stats.ActiveConnections,
				MaxConnections:    p.config.MaxConnections,
				MinConnections:    p.config.MinConnections,
			})
		}
	default:
		wrapper.conn.Close()
		p.stats.TotalConnections--
		p.stats.IdleConnections--

		for i, c := range p.allConnections {
			if c == wrapper {
				p.allConnections = append(p.allConnections[:i], p.allConnections[i+1:]...)
				break
			}
		}

		if p.metricsCollector != nil {
			p.metricsCollector.RecordConnectionClosed(p.instanceName)
		}
	}
}

func (p *ConnectionPoolImpl) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	p.mu.Unlock()

	p.cancel()

	if p.healthCheckTicker != nil {
		p.healthCheckTicker.Stop()
	}

	p.mu.Lock()
	connections := make([]*connectionWrapper, len(p.allConnections))
	copy(connections, p.allConnections)
	p.mu.Unlock()

	for _, conn := range connections {
		conn.conn.Close()
		if p.metricsCollector != nil {
			p.metricsCollector.RecordConnectionClosed(p.instanceName)
		}
	}

	close(p.idleConnections)

	return nil
}

func (p *ConnectionPoolImpl) GetStats() *connection.PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := p.stats
	stats.MaxConnections = p.config.MaxConnections
	stats.MinConnections = p.config.MinConnections

	if stats.AcquireCount > 0 && p.metricsCollector != nil {
		metrics := p.metricsCollector.GetInstanceMetrics(p.instanceName)
		if metrics != nil {
			stats.AvgAcquireDuration = metrics.AvgAcquireDuration
			stats.MaxAcquireDuration = metrics.MaxAcquireDuration
			stats.WaitCount = metrics.WaitCount
			stats.WaitDuration = metrics.WaitDuration
			stats.MaxWaitDuration = metrics.MaxWaitDuration
			stats.AvgWaitDuration = metrics.AvgWaitDuration
			stats.QueryCount = metrics.QueryCount
			stats.QueryFailed = metrics.QueryFailed
			stats.ExecCount = metrics.ExecCount
			stats.ExecFailed = metrics.ExecFailed
			stats.HealthCheckCount = metrics.HealthCheckCount
			stats.HealthCheckFailed = metrics.HealthCheckFailed
			stats.ConnectionsCreated = metrics.ConnectionsCreated
			stats.ConnectionsClosed = metrics.ConnectionsClosed
		}
	}

	return &stats
}

func (p *ConnectionPoolImpl) SetMaxConnections(max int) {
	if max <= 0 {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.config.MaxConnections = max
	p.stats.MaxConnections = max

	oldPool := p.idleConnections
	p.idleConnections = make(chan *connectionWrapper, max)

	close(oldPool)
	for conn := range oldPool {
		select {
		case p.idleConnections <- conn:
		default:
			conn.conn.Close()
			p.stats.TotalConnections--
			p.stats.IdleConnections--

			for i, c := range p.allConnections {
				if c == conn {
					p.allConnections = append(p.allConnections[:i], p.allConnections[i+1:]...)
					break
				}
			}

			if p.metricsCollector != nil {
				p.metricsCollector.RecordConnectionClosed(p.instanceName)
			}
		}
	}
}

func (p *ConnectionPoolImpl) SetMinConnections(min int) {
	if min < 0 {
		min = 0
	}

	p.mu.Lock()
	if min > p.config.MaxConnections {
		min = p.config.MaxConnections
	}
	p.config.MinConnections = min
	p.stats.MinConnections = min
	p.mu.Unlock()

	for {
		p.mu.Lock()
		totalCount := p.stats.TotalConnections
		p.mu.Unlock()

		if totalCount >= min {
			break
		}

		conn, err := p.createConnection()
		if err != nil {
			break
		}

		p.mu.Lock()
		p.stats.IdleConnections++
		p.mu.Unlock()

		select {
		case p.idleConnections <- conn:
		default:
			conn.conn.Close()
			p.mu.Lock()
			p.stats.TotalConnections--
			p.mu.Unlock()
		}
	}
}

func (p *ConnectionPoolImpl) SetMaxIdleTime(duration string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.config.MaxIdleTime = duration
}

func (p *ConnectionPoolImpl) SetMaxLifetime(duration string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.config.MaxLifetime = duration
}

func (p *ConnectionPoolImpl) startHealthCheck() {
	pingInterval, err := time.ParseDuration(p.config.PingInterval)
	if err != nil || pingInterval <= 0 {
		pingInterval = 30 * time.Second
	}

	p.healthCheckTicker = time.NewTicker(pingInterval)

	go func() {
		for {
			select {
			case <-p.healthCheckTicker.C:
				p.checkConnectionsHealth()
			case <-p.ctx.Done():
				return
			}
		}
	}()
}

func (p *ConnectionPoolImpl) checkConnectionsHealth() {
	if p.closed {
		return
	}

	idleConnCount := len(p.idleConnections)
healthCheck:
	for i := 0; i < idleConnCount; i++ {
		select {
		case conn := <-p.idleConnections:
			maxIdleTime, _ := time.ParseDuration(p.config.MaxIdleTime)
			if maxIdleTime > 0 && time.Since(conn.lastUsedAt) > maxIdleTime {
				conn.conn.Close()

				p.mu.Lock()
				p.stats.TotalConnections--
				p.stats.IdleConnections--

				for j, c := range p.allConnections {
					if c == conn {
						p.allConnections = append(p.allConnections[:j], p.allConnections[j+1:]...)
						break
					}
				}
				p.mu.Unlock()

				if p.metricsCollector != nil {
					p.metricsCollector.RecordConnectionClosed(p.instanceName)
				}
				continue
			}

			ctx, cancel := context.WithTimeout(p.ctx, 5*time.Second)
			if err := conn.conn.Ping(ctx); err != nil {
				conn.conn.Close()

				p.mu.Lock()
				p.stats.TotalConnections--
				p.stats.IdleConnections--

				for j, c := range p.allConnections {
					if c == conn {
						p.allConnections = append(p.allConnections[:j], p.allConnections[j+1:]...)
						break
					}
				}
				p.mu.Unlock()

				if p.metricsCollector != nil {
					p.metricsCollector.RecordHealthCheck(p.instanceName, false)
					p.metricsCollector.RecordConnectionClosed(p.instanceName)
				}
			} else {
				p.idleConnections <- conn

				if p.metricsCollector != nil {
					p.metricsCollector.RecordHealthCheck(p.instanceName, true)
				}
			}
			cancel()

		default:
			break healthCheck
		}
	}

	p.mu.Lock()
	minConnections := p.config.MinConnections
	totalConnections := p.stats.TotalConnections
	p.mu.Unlock()

	if totalConnections < minConnections {
		for i := totalConnections; i < minConnections; i++ {
			conn, err := p.createConnection()
			if err != nil {
				break
			}

			p.mu.Lock()
			p.stats.IdleConnections++
			p.mu.Unlock()

			select {
			case p.idleConnections <- conn:
			default:
				conn.conn.Close()
				p.mu.Lock()
				p.stats.TotalConnections--
				p.mu.Unlock()
			}
		}
	}
}

func (c *connectionWrapper) Query(ctx context.Context, sql string) (*types.QueryResult, error) {
	start := time.Now()
	result, err := c.conn.Query(ctx, sql)
	duration := time.Since(start).Nanoseconds()

	if c.pool.metricsCollector != nil {
		c.pool.metricsCollector.RecordQuery(c.pool.instanceName, duration, err == nil)
	}

	return result, err
}

func (c *connectionWrapper) Exec(ctx context.Context, sql string) (*types.ExecResult, error) {
	start := time.Now()
	result, err := c.conn.Exec(ctx, sql)
	duration := time.Since(start).Nanoseconds()

	if c.pool.metricsCollector != nil {
		c.pool.metricsCollector.RecordExec(c.pool.instanceName, duration, err == nil)
	}

	return result, err
}

func (c *connectionWrapper) Ping(ctx context.Context) error {
	return c.conn.Ping(ctx)
}

func (c *connectionWrapper) Close() error {
	c.pool.ReturnConnection(c)
	return nil
}

func (c *connectionWrapper) Connect(ctx context.Context) error {
	return c.conn.Connect(ctx)
}
