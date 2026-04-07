package redis

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"MystiSql/internal/connection"
	"MystiSql/pkg/errors"
	"MystiSql/pkg/types"

	"github.com/redis/go-redis/v9"
)

type Connection struct {
	instance *types.DatabaseInstance
	client   *redis.Client
}

type Factory struct{}

func NewFactory() connection.ConnectionFactory {
	return &Factory{}
}

func (f *Factory) CreateConnection(instance *types.DatabaseInstance) (connection.Connection, error) {
	return NewConnection(instance), nil
}

func NewConnection(instance *types.DatabaseInstance) connection.Connection {
	return &Connection{
		instance: instance,
	}
}

func (c *Connection) Connect(ctx context.Context) error {
	opts := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", c.instance.Host, c.instance.Port),
		Username: c.instance.Username,
		Password: c.instance.Password,
	}

	if c.instance.Database != "" {
		if db, err := strconv.Atoi(c.instance.Database); err == nil {
			opts.DB = db
		}
	}

	c.client = redis.NewClient(opts)

	if err := c.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("%w: ping failed: %v", errors.ErrConnectionFailed, err)
	}

	c.instance.SetStatus(types.InstanceStatusHealthy)
	return nil
}

func (c *Connection) Query(ctx context.Context, query string) (*types.QueryResult, error) {
	if c.client == nil {
		return nil, errors.ErrConnectionClosed
	}

	start := time.Now()
	query = strings.TrimSpace(query)
	upperQuery := strings.ToUpper(query)

	switch {
	case strings.HasPrefix(upperQuery, "GET "):
		return c.executeGet(ctx, query, start)
	case strings.HasPrefix(upperQuery, "KEYS "):
		return c.executeKeys(ctx, query, start)
	case strings.HasPrefix(upperQuery, "TYPE "):
		return c.executeType(ctx, query, start)
	case strings.HasPrefix(upperQuery, "HGETALL "):
		return c.executeHGetAll(ctx, query, start)
	case strings.HasPrefix(upperQuery, "LRANGE "):
		return c.executeLRange(ctx, query, start)
	case strings.HasPrefix(upperQuery, "SMEMBERS "):
		return c.executeSMembers(ctx, query, start)
	case upperQuery == "DBSIZE":
		return c.executeDBSize(ctx, start)
	case upperQuery == "INFO":
		return c.executeInfo(ctx, start)
	default:
		return nil, fmt.Errorf("%w: unsupported query: %s", errors.ErrQueryFailed, query)
	}
}

func (c *Connection) Exec(ctx context.Context, query string) (*types.ExecResult, error) {
	if c.client == nil {
		return nil, errors.ErrConnectionClosed
	}

	start := time.Now()
	query = strings.TrimSpace(query)
	upperQuery := strings.ToUpper(query)

	switch {
	case strings.HasPrefix(upperQuery, "SET ") || strings.HasPrefix(upperQuery, "SETEX "):
		return c.executeSet(ctx, query, start)
	case strings.HasPrefix(upperQuery, "DEL "):
		return c.executeDel(ctx, query, start)
	case strings.HasPrefix(upperQuery, "HSET "):
		return c.executeHSet(ctx, query, start)
	case strings.HasPrefix(upperQuery, "LPUSH "):
		return c.executeLPush(ctx, query, start)
	case strings.HasPrefix(upperQuery, "RPUSH "):
		return c.executeRPush(ctx, query, start)
	case strings.HasPrefix(upperQuery, "SADD "):
		return c.executeSAdd(ctx, query, start)
	case strings.HasPrefix(upperQuery, "INCR "):
		return c.executeIncr(ctx, query, start)
	default:
		return nil, fmt.Errorf("%w: unsupported command: %s", errors.ErrQueryFailed, query)
	}
}

func (c *Connection) Ping(ctx context.Context) error {
	if c.client == nil {
		return errors.ErrConnectionClosed
	}

	if err := c.client.Ping(ctx).Err(); err != nil {
		c.instance.SetStatus(types.InstanceStatusUnhealthy)
		return fmt.Errorf("%w: ping failed: %v", errors.ErrConnectionFailed, err)
	}

	c.instance.SetStatus(types.InstanceStatusHealthy)
	return nil
}

func (c *Connection) Close() error {
	if c.client == nil {
		return nil
	}

	err := c.client.Close()
	c.client = nil
	c.instance.SetStatus(types.InstanceStatusUnknown)

	if err != nil {
		return fmt.Errorf("close connection failed: %v", err)
	}

	return nil
}

func (c *Connection) executeGet(ctx context.Context, query string, start time.Time) (*types.QueryResult, error) {
	parts := strings.Fields(query)
	if len(parts) < 2 {
		return nil, fmt.Errorf("GET requires a key")
	}
	key := parts[1]

	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return types.NewQueryResult([]types.ColumnInfo{{Name: "value", Type: "string"}}, []types.Row{{nil}}, time.Since(start)), nil
	}
	if err != nil {
		return nil, fmt.Errorf("GET failed: %v", err)
	}

	return types.NewQueryResult(
		[]types.ColumnInfo{{Name: "value", Type: "string"}},
		[]types.Row{{val}},
		time.Since(start),
	), nil
}

func (c *Connection) executeKeys(ctx context.Context, query string, start time.Time) (*types.QueryResult, error) {
	parts := strings.Fields(query)
	pattern := "*"
	if len(parts) >= 2 {
		pattern = parts[1]
	}

	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("KEYS failed: %v", err)
	}

	rows := make([]types.Row, len(keys))
	for i, key := range keys {
		rows[i] = types.Row{key}
	}

	return types.NewQueryResult(
		[]types.ColumnInfo{{Name: "key", Type: "string"}},
		rows,
		time.Since(start),
	), nil
}

func (c *Connection) executeType(ctx context.Context, query string, start time.Time) (*types.QueryResult, error) {
	parts := strings.Fields(query)
	if len(parts) < 2 {
		return nil, fmt.Errorf("TYPE requires a key")
	}
	key := parts[1]

	val, err := c.client.Type(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("TYPE failed: %v", err)
	}

	return types.NewQueryResult(
		[]types.ColumnInfo{{Name: "type", Type: "string"}},
		[]types.Row{{val}},
		time.Since(start),
	), nil
}

func (c *Connection) executeHGetAll(ctx context.Context, query string, start time.Time) (*types.QueryResult, error) {
	parts := strings.Fields(query)
	if len(parts) < 2 {
		return nil, fmt.Errorf("HGETALL requires a key")
	}
	key := parts[1]

	vals, err := c.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("HGETALL failed: %v", err)
	}

	rows := make([]types.Row, 0, len(vals))
	for k, v := range vals {
		rows = append(rows, types.Row{k, v})
	}

	return types.NewQueryResult(
		[]types.ColumnInfo{{Name: "field", Type: "string"}, {Name: "value", Type: "string"}},
		rows,
		time.Since(start),
	), nil
}

func (c *Connection) executeLRange(ctx context.Context, query string, start time.Time) (*types.QueryResult, error) {
	parts := strings.Fields(query)
	if len(parts) < 4 {
		return nil, fmt.Errorf("LRANGE requires key, start, stop")
	}
	key := parts[1]
	startIdx, _ := strconv.ParseInt(parts[2], 10, 64)
	stopIdx, _ := strconv.ParseInt(parts[3], 10, 64)

	vals, err := c.client.LRange(ctx, key, startIdx, stopIdx).Result()
	if err != nil {
		return nil, fmt.Errorf("LRANGE failed: %v", err)
	}

	rows := make([]types.Row, len(vals))
	for i, v := range vals {
		rows[i] = types.Row{v}
	}

	return types.NewQueryResult(
		[]types.ColumnInfo{{Name: "value", Type: "string"}},
		rows,
		time.Since(start),
	), nil
}

func (c *Connection) executeSMembers(ctx context.Context, query string, start time.Time) (*types.QueryResult, error) {
	parts := strings.Fields(query)
	if len(parts) < 2 {
		return nil, fmt.Errorf("SMEMBERS requires a key")
	}
	key := parts[1]

	vals, err := c.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("SMEMBERS failed: %v", err)
	}

	rows := make([]types.Row, len(vals))
	for i, v := range vals {
		rows[i] = types.Row{v}
	}

	return types.NewQueryResult(
		[]types.ColumnInfo{{Name: "member", Type: "string"}},
		rows,
		time.Since(start),
	), nil
}

func (c *Connection) executeDBSize(ctx context.Context, start time.Time) (*types.QueryResult, error) {
	val, err := c.client.DBSize(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("DBSIZE failed: %v", err)
	}

	return types.NewQueryResult(
		[]types.ColumnInfo{{Name: "size", Type: "int64"}},
		[]types.Row{{val}},
		time.Since(start),
	), nil
}

func (c *Connection) executeInfo(ctx context.Context, start time.Time) (*types.QueryResult, error) {
	val, err := c.client.Info(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("INFO failed: %v", err)
	}

	return types.NewQueryResult(
		[]types.ColumnInfo{{Name: "info", Type: "string"}},
		[]types.Row{{val}},
		time.Since(start),
	), nil
}

func (c *Connection) executeSet(ctx context.Context, query string, start time.Time) (*types.ExecResult, error) {
	upperQuery := strings.ToUpper(query)

	if strings.HasPrefix(upperQuery, "SETEX ") {
		parts := strings.Fields(query)
		if len(parts) < 4 {
			return nil, fmt.Errorf("SETEX requires key, seconds, value")
		}
		key := parts[1]
		seconds, _ := strconv.ParseInt(parts[2], 10, 64)
		value := strings.Join(parts[3:], " ")

		ok, err := c.client.Set(ctx, key, value, time.Duration(seconds)*time.Second).Result()
		if err != nil {
			return nil, fmt.Errorf("SETEX failed: %v", err)
		}
		if ok != "OK" {
			return nil, fmt.Errorf("SETEX failed")
		}

		return types.NewExecResult(1, 0, time.Since(start)), nil
	}

	parts := strings.Fields(query)
	if len(parts) < 3 {
		return nil, fmt.Errorf("SET requires key and value")
	}
	key := parts[1]
	value := strings.Join(parts[2:], " ")

	ok, err := c.client.Set(ctx, key, value, 0).Result()
	if err != nil {
		return nil, fmt.Errorf("SET failed: %v", err)
	}
	if ok != "OK" {
		return nil, fmt.Errorf("SET failed")
	}

	return types.NewExecResult(1, 0, time.Since(start)), nil
}

func (c *Connection) executeDel(ctx context.Context, query string, start time.Time) (*types.ExecResult, error) {
	parts := strings.Fields(query)
	if len(parts) < 2 {
		return nil, fmt.Errorf("DEL requires at least one key")
	}
	keys := parts[1:]

	count, err := c.client.Del(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("DEL failed: %v", err)
	}

	return types.NewExecResult(count, 0, time.Since(start)), nil
}

func (c *Connection) executeHSet(ctx context.Context, query string, start time.Time) (*types.ExecResult, error) {
	parts := strings.Fields(query)
	if len(parts) < 4 {
		return nil, fmt.Errorf("HSET requires key, field, value")
	}
	key := parts[1]
	field := parts[2]
	value := strings.Join(parts[3:], " ")

	count, err := c.client.HSet(ctx, key, field, value).Result()
	if err != nil {
		return nil, fmt.Errorf("HSET failed: %v", err)
	}

	return types.NewExecResult(count, 0, time.Since(start)), nil
}

func (c *Connection) executeLPush(ctx context.Context, query string, start time.Time) (*types.ExecResult, error) {
	parts := strings.Fields(query)
	if len(parts) < 3 {
		return nil, fmt.Errorf("LPUSH requires key and value")
	}
	key := parts[1]
	value := strings.Join(parts[2:], " ")

	count, err := c.client.LPush(ctx, key, value).Result()
	if err != nil {
		return nil, fmt.Errorf("LPUSH failed: %v", err)
	}

	return types.NewExecResult(count, 0, time.Since(start)), nil
}

func (c *Connection) executeRPush(ctx context.Context, query string, start time.Time) (*types.ExecResult, error) {
	parts := strings.Fields(query)
	if len(parts) < 3 {
		return nil, fmt.Errorf("RPUSH requires key and value")
	}
	key := parts[1]
	value := strings.Join(parts[2:], " ")

	count, err := c.client.RPush(ctx, key, value).Result()
	if err != nil {
		return nil, fmt.Errorf("RPUSH failed: %v", err)
	}

	return types.NewExecResult(count, 0, time.Since(start)), nil
}

func (c *Connection) executeSAdd(ctx context.Context, query string, start time.Time) (*types.ExecResult, error) {
	parts := strings.Fields(query)
	if len(parts) < 3 {
		return nil, fmt.Errorf("SADD requires key and member")
	}
	key := parts[1]
	member := strings.Join(parts[2:], " ")

	count, err := c.client.SAdd(ctx, key, member).Result()
	if err != nil {
		return nil, fmt.Errorf("SADD failed: %v", err)
	}

	return types.NewExecResult(count, 0, time.Since(start)), nil
}

func (c *Connection) executeIncr(ctx context.Context, query string, start time.Time) (*types.ExecResult, error) {
	parts := strings.Fields(query)
	if len(parts) < 2 {
		return nil, fmt.Errorf("INCR requires a key")
	}
	key := parts[1]

	val, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("INCR failed: %v", err)
	}

	return types.NewExecResult(1, val, time.Since(start)), nil
}
