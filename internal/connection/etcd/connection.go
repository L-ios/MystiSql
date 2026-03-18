package etcd

import (
	"context"
	"fmt"
	"time"

	"MystiSql/internal/connection"
	"MystiSql/pkg/errors"
	"MystiSql/pkg/types"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Connection struct {
	instance *types.DatabaseInstance
	client   *clientv3.Client
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
	endpoints := []string{fmt.Sprintf("http://%s:%d", c.instance.Host, c.instance.Port)}

	cfg := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 10 * time.Second,
	}

	if c.instance.Username != "" && c.instance.Password != "" {
		cfg.Username = c.instance.Username
		cfg.Password = c.instance.Password
	}

	client, err := clientv3.New(cfg)
	if err != nil {
		return fmt.Errorf("%w: create client failed: %v", errors.ErrConnectionFailed, err)
	}

	c.client = client
	c.instance.SetStatus(types.InstanceStatusHealthy)
	return nil
}

func (c *Connection) Query(ctx context.Context, query string) (*types.QueryResult, error) {
	if c.client == nil {
		return nil, errors.ErrConnectionClosed
	}

	start := time.Now()

	if len(query) > 4 && query[:4] == "GET " {
		key := query[4:]
		resp, err := c.client.Get(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("%w: get failed: %v", errors.ErrQueryFailed, err)
		}

		var resultRows []types.Row
		for _, kv := range resp.Kvs {
			resultRows = append(resultRows, types.Row{string(kv.Key), string(kv.Value)})
		}

		columnInfos := []types.ColumnInfo{{Name: "key", Type: "string"}, {Name: "value", Type: "string"}}
		return types.NewQueryResult(columnInfos, resultRows, time.Since(start)), nil
	}

	return nil, fmt.Errorf("%w: unsupported query: %s", errors.ErrQueryFailed, query)
}

func (c *Connection) Exec(ctx context.Context, query string) (*types.ExecResult, error) {
	if c.client == nil {
		return nil, errors.ErrConnectionClosed
	}

	start := time.Now()

	if len(query) > 4 && query[:4] == "PUT " {
		parts := []byte(query[4:])
		spaceIdx := -1
		for i, b := range parts {
			if b == ' ' {
				spaceIdx = i
				break
			}
		}
		if spaceIdx == -1 {
			return nil, fmt.Errorf("%w: invalid PUT format", errors.ErrQueryFailed)
		}
		key := string(parts[:spaceIdx])
		value := string(parts[spaceIdx+1:])

		_, err := c.client.Put(ctx, key, value)
		if err != nil {
			return nil, fmt.Errorf("%w: put failed: %v", errors.ErrQueryFailed, err)
		}
		return types.NewExecResult(1, 0, time.Since(start)), nil
	}

	if len(query) > 7 && query[:7] == "DELETE " {
		key := query[7:]
		_, err := c.client.Delete(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("%w: delete failed: %v", errors.ErrQueryFailed, err)
		}
		return types.NewExecResult(1, 0, time.Since(start)), nil
	}

	return nil, fmt.Errorf("%w: unsupported command: %s", errors.ErrQueryFailed, query)
}

func (c *Connection) Ping(ctx context.Context) error {
	if c.client == nil {
		return errors.ErrConnectionClosed
	}

	_, err := c.client.Get(ctx, "health-check-key")
	if err != nil {
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
