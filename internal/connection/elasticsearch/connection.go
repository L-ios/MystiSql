package elasticsearch

import (
	"context"
	"fmt"
	"strings"
	"time"

	"MystiSql/internal/connection"
	"MystiSql/pkg/errors"
	"MystiSql/pkg/types"

	"github.com/elastic/go-elasticsearch/v8"
)

type Connection struct {
	instance *types.DatabaseInstance
	client   *elasticsearch.Client
	index    string
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
	address := fmt.Sprintf("http://%s:%d", c.instance.Host, c.instance.Port)

	cfg := elasticsearch.Config{
		Addresses: []string{address},
	}

	if c.instance.Username != "" && c.instance.Password != "" {
		cfg.Username = c.instance.Username
		cfg.Password = c.instance.Password
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("%w: create client failed: %v", errors.ErrConnectionFailed, err)
	}

	c.client = client
	c.index = c.instance.Database
	c.instance.SetStatus(types.InstanceStatusHealthy)
	return nil
}

func (c *Connection) Query(ctx context.Context, query string) (*types.QueryResult, error) {
	if c.client == nil {
		return nil, errors.ErrConnectionClosed
	}

	start := time.Now()

	res, err := c.client.Search(
		c.client.Search.WithContext(ctx),
		c.client.Search.WithIndex(c.index),
		c.client.Search.WithBody(strings.NewReader(query)),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: search failed: %v", errors.ErrQueryFailed, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("%w: search error: %s", errors.ErrQueryFailed, res.String())
	}

	var resultRows []types.Row
	resultRows = append(resultRows, types.Row{res.Body})

	columnInfos := []types.ColumnInfo{{Name: "response", Type: "json"}}

	return types.NewQueryResult(columnInfos, resultRows, time.Since(start)), nil
}

func (c *Connection) Exec(ctx context.Context, query string) (*types.ExecResult, error) {
	if c.client == nil {
		return nil, errors.ErrConnectionClosed
	}

	start := time.Now()

	res, err := c.client.Index(c.index).BodyString(query).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: index failed: %v", errors.ErrQueryFailed, err)
	}

	return types.NewExecResult(1, 0, time.Since(start)), nil
}

func (c *Connection) Ping(ctx context.Context) error {
	if c.client == nil {
		return errors.ErrConnectionClosed
	}

	res, err := c.client.Info(c.client.Info.WithContext(ctx))
	if err != nil {
		c.instance.SetStatus(types.InstanceStatusUnhealthy)
		return fmt.Errorf("%w: ping failed: %v", errors.ErrConnectionFailed, err)
	}

	if res.IsError() {
		c.instance.SetStatus(types.InstanceStatusUnhealthy)
		return fmt.Errorf("%w: ping failed: %s", errors.ErrConnectionFailed, res.String())
	}

	c.instance.SetStatus(types.InstanceStatusHealthy)
	return nil
}

func (c *Connection) Close() error {
	c.client = nil
	c.instance.SetStatus(types.InstanceStatusUnknown)
	return nil
}
