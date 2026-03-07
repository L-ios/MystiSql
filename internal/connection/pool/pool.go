package pool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"MystiSql/internal/connection"
	"MystiSql/pkg/types"
)

// connectionWrapper 包装连接，添加连接池相关信息
type connectionWrapper struct {
	conn       connection.Connection
	pool       *ConnectionPoolImpl
	createdAt  time.Time
	lastUsedAt time.Time
	instance   *types.DatabaseInstance
}

// ConnectionPoolImpl 实现 ConnectionPool 接口
type ConnectionPoolImpl struct {
	// 配置
	config *connection.PoolConfig

	// 连接工厂
	factory connection.ConnectionFactory

	// 数据库实例
	instance *types.DatabaseInstance

	// 空闲连接池
	idleConnections chan *connectionWrapper

	// 所有连接
	allConnections []*connectionWrapper

	// 互斥锁
	mu sync.RWMutex

	// 统计信息
	stats connection.PoolStats

	// 关闭标志
	closed bool

	// 健康检查定时器
	healthCheckTicker *time.Ticker

	// 上下文
	ctx    context.Context
	cancel context.CancelFunc
}

// NewConnectionPool 创建一个新的连接池
func NewConnectionPool(
	instance *types.DatabaseInstance,
	factory connection.ConnectionFactory,
	config *connection.PoolConfig,
) (connection.ConnectionPool, error) {
	// 设置默认值
	if config.MaxConnections <= 0 {
		config.MaxConnections = 10
	}
	if config.MinConnections < 0 {
		config.MinConnections = 0
	}
	if config.MinConnections > config.MaxConnections {
		config.MinConnections = config.MaxConnections
	}

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 创建连接池
	pool := &ConnectionPoolImpl{
		config:          config,
		factory:         factory,
		instance:        instance,
		idleConnections: make(chan *connectionWrapper, config.MaxConnections),
		allConnections:  make([]*connectionWrapper, 0, config.MaxConnections),
		stats: connection.PoolStats{
			MaxConnections: config.MaxConnections,
			MinConnections: config.MinConnections,
		},
		closed: false,
		ctx:    ctx,
		cancel: cancel,
	}

	// 初始化最小连接数
	if config.MinConnections > 0 {
		if err := pool.initializeMinConnections(); err != nil {
			cancel()
			return nil, fmt.Errorf("failed to initialize min connections: %w", err)
		}
	}

	// 启动健康检查
	pool.startHealthCheck()

	return pool, nil
}

// initializeMinConnections 初始化最小连接数
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

// createConnection 创建一个新的连接
func (p *ConnectionPoolImpl) createConnection() (*connectionWrapper, error) {
	conn, err := p.factory.CreateConnection(p.instance)
	if err != nil {
		return nil, err
	}

	// 连接到数据库
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

	return wrapper, nil
}

// GetConnection 从连接池中获取一个连接
func (p *ConnectionPoolImpl) GetConnection(ctx context.Context) (connection.Connection, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, fmt.Errorf("connection pool is closed")
	}

	// 增加获取连接计数
	p.stats.AcquireCount++
	p.mu.Unlock()

	// 尝试从空闲连接池获取
	select {
	case conn := <-p.idleConnections:
		p.mu.Lock()
		p.stats.ActiveConnections++
		p.stats.IdleConnections--
		p.mu.Unlock()

		// 检查连接是否有效
		if err := conn.conn.Ping(ctx); err != nil {
			// 连接无效，创建新连接
			p.mu.Lock()
			p.stats.AcquireFailed++
			p.stats.ActiveConnections--
			p.mu.Unlock()

			// 关闭无效连接
			conn.conn.Close()

			// 重新获取连接
			return p.GetConnection(ctx)
		}

		// 更新最后使用时间
		conn.lastUsedAt = time.Now()
		return conn, nil

	default:
		// 没有空闲连接，检查是否达到最大连接数
		p.mu.Lock()
		totalConnections := p.stats.TotalConnections
		p.mu.Unlock()

		if totalConnections >= p.config.MaxConnections {
			// 等待空闲连接
			select {
			case conn := <-p.idleConnections:
				p.mu.Lock()
				p.stats.ActiveConnections++
				p.stats.IdleConnections--
				p.mu.Unlock()

				// 检查连接是否有效
				if err := conn.conn.Ping(ctx); err != nil {
					// 连接无效，创建新连接
					p.mu.Lock()
					p.stats.AcquireFailed++
					p.stats.ActiveConnections--
					p.mu.Unlock()

					// 关闭无效连接
					conn.conn.Close()

					// 重新获取连接
					return p.GetConnection(ctx)
				}

				// 更新最后使用时间
				conn.lastUsedAt = time.Now()
				return conn, nil

			case <-ctx.Done():
				p.mu.Lock()
				p.stats.AcquireFailed++
				p.mu.Unlock()
				return nil, ctx.Err()
			}
		}

		// 创建新连接
		conn, err := p.createConnection()
		if err != nil {
			p.mu.Lock()
			p.stats.AcquireFailed++
			p.mu.Unlock()
			return nil, err
		}

		p.mu.Lock()
		p.stats.ActiveConnections++
		p.mu.Unlock()

		return conn, nil
	}
}

// ReturnConnection 将连接归还到连接池
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

	// 检查连接是否有效
	ctx, cancel := context.WithTimeout(p.ctx, 5*time.Second)
	defer cancel()
	if err := wrapper.conn.Ping(ctx); err != nil {
		// 连接无效，关闭并从连接池移除
		wrapper.conn.Close()
		p.stats.ActiveConnections--
		p.stats.TotalConnections--

		// 从所有连接列表中移除
		for i, c := range p.allConnections {
			if c == wrapper {
				p.allConnections = append(p.allConnections[:i], p.allConnections[i+1:]...)
				break
			}
		}
		return
	}

	// 检查连接是否超过最大生命周期
	maxLifetime, _ := time.ParseDuration(p.config.MaxLifetime)
	if maxLifetime > 0 && time.Since(wrapper.createdAt) > maxLifetime {
		// 连接已超过最大生命周期，关闭并从连接池移除
		wrapper.conn.Close()
		p.stats.ActiveConnections--
		p.stats.TotalConnections--

		// 从所有连接列表中移除
		for i, c := range p.allConnections {
			if c == wrapper {
				p.allConnections = append(p.allConnections[:i], p.allConnections[i+1:]...)
				break
			}
		}
		return
	}

	// 将连接归还到空闲连接池
	p.stats.ActiveConnections--
	p.stats.IdleConnections++
	p.stats.ReleaseCount++

	select {
	case p.idleConnections <- wrapper:
		// 连接成功归还
	default:
		// 空闲连接池已满，关闭连接
		wrapper.conn.Close()
		p.stats.TotalConnections--
		p.stats.IdleConnections--

		// 从所有连接列表中移除
		for i, c := range p.allConnections {
			if c == wrapper {
				p.allConnections = append(p.allConnections[:i], p.allConnections[i+1:]...)
				break
			}
		}
	}
}

// Close 关闭连接池并释放所有连接
func (p *ConnectionPoolImpl) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	p.mu.Unlock()

	// 取消上下文
	p.cancel()

	// 关闭健康检查定时器
	if p.healthCheckTicker != nil {
		p.healthCheckTicker.Stop()
	}

	// 关闭所有连接
	p.mu.Lock()
	connections := make([]*connectionWrapper, len(p.allConnections))
	copy(connections, p.allConnections)
	p.mu.Unlock()

	for _, conn := range connections {
		conn.conn.Close()
	}

	// 清空连接池
	close(p.idleConnections)

	return nil
}

// GetStats 获取连接池统计信息
func (p *ConnectionPoolImpl) GetStats() *connection.PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := p.stats
	return &stats
}

// SetMaxConnections 设置最大连接数
func (p *ConnectionPoolImpl) SetMaxConnections(max int) {
	if max <= 0 {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.config.MaxConnections = max
	p.stats.MaxConnections = max

	// 重新创建空闲连接池
	oldPool := p.idleConnections
	p.idleConnections = make(chan *connectionWrapper, max)

	// 将旧连接池中的连接转移到新连接池
	close(oldPool)
	for conn := range oldPool {
		select {
		case p.idleConnections <- conn:
		default:
			// 新连接池已满，关闭多余连接
			conn.conn.Close()
			p.stats.TotalConnections--
			p.stats.IdleConnections--

			// 从所有连接列表中移除
			for i, c := range p.allConnections {
				if c == conn {
					p.allConnections = append(p.allConnections[:i], p.allConnections[i+1:]...)
					break
				}
			}
		}
	}
}

// SetMinConnections 设置最小连接数
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

	// 确保达到最小连接数
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
			// 空闲连接池已满，关闭连接
			conn.conn.Close()
			p.mu.Lock()
			p.stats.TotalConnections--
			p.mu.Unlock()
		}
	}
}

// SetMaxIdleTime 设置连接最大空闲时间
func (p *ConnectionPoolImpl) SetMaxIdleTime(duration string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.config.MaxIdleTime = duration
}

// SetMaxLifetime 设置连接最大生命周期
func (p *ConnectionPoolImpl) SetMaxLifetime(duration string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.config.MaxLifetime = duration
}

// startHealthCheck 启动健康检查
func (p *ConnectionPoolImpl) startHealthCheck() {
	// 解析健康检查间隔
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

// checkConnectionsHealth 检查连接健康状态
func (p *ConnectionPoolImpl) checkConnectionsHealth() {
	if p.closed {
		return
	}

	// 检查空闲连接
	idleConnCount := len(p.idleConnections)
	for i := 0; i < idleConnCount; i++ {
		select {
		case conn := <-p.idleConnections:
			// 检查连接是否超过最大空闲时间
			maxIdleTime, _ := time.ParseDuration(p.config.MaxIdleTime)
			if maxIdleTime > 0 && time.Since(conn.lastUsedAt) > maxIdleTime {
				// 连接已超过最大空闲时间，关闭并从连接池移除
				conn.conn.Close()

				p.mu.Lock()
				p.stats.TotalConnections--
				p.stats.IdleConnections--

				// 从所有连接列表中移除
				for j, c := range p.allConnections {
					if c == conn {
						p.allConnections = append(p.allConnections[:j], p.allConnections[j+1:]...)
						break
					}
				}
				p.mu.Unlock()
				continue
			}

			// 检查连接是否有效
			ctx, cancel := context.WithTimeout(p.ctx, 5*time.Second)
			if err := conn.conn.Ping(ctx); err != nil {
				// 连接无效，关闭并从连接池移除
				conn.conn.Close()

				p.mu.Lock()
				p.stats.TotalConnections--
				p.stats.IdleConnections--

				// 从所有连接列表中移除
				for j, c := range p.allConnections {
					if c == conn {
						p.allConnections = append(p.allConnections[:j], p.allConnections[j+1:]...)
						break
					}
				}
				p.mu.Unlock()
			} else {
				// 连接有效，放回空闲连接池
				p.idleConnections <- conn
			}
			cancel()

		default:
			break
		}
	}

	// 确保达到最小连接数
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
				// 空闲连接池已满，关闭连接
				conn.conn.Close()
				p.mu.Lock()
				p.stats.TotalConnections--
				p.mu.Unlock()
			}
		}
	}
}

// Query 实现 Connection 接口的 Query 方法
func (c *connectionWrapper) Query(ctx context.Context, sql string) (*types.QueryResult, error) {
	return c.conn.Query(ctx, sql)
}

// Exec 实现 Connection 接口的 Exec 方法
func (c *connectionWrapper) Exec(ctx context.Context, sql string) (*types.ExecResult, error) {
	return c.conn.Exec(ctx, sql)
}

// Ping 实现 Connection 接口的 Ping 方法
func (c *connectionWrapper) Ping(ctx context.Context) error {
	return c.conn.Ping(ctx)
}

// Close 实现 Connection 接口的 Close 方法
// 注意：这里不直接关闭连接，而是将连接归还到连接池
func (c *connectionWrapper) Close() error {
	c.pool.ReturnConnection(c)
	return nil
}

// Connect 实现 Connection 接口的 Connect 方法
func (c *connectionWrapper) Connect(ctx context.Context) error {
	return c.conn.Connect(ctx)
}
