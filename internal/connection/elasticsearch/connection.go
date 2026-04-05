package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"MystiSql/internal/connection"
	"MystiSql/pkg/errors"
	"MystiSql/pkg/types"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
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

	// Read the entire body before closing so the caller can access the data.
	bodyBytes, readErr := io.ReadAll(res.Body)
	res.Body.Close()
	if readErr != nil {
		return nil, fmt.Errorf("%w: read response body failed: %v", errors.ErrQueryFailed, readErr)
	}

	if res.IsError() {
		return nil, fmt.Errorf("%w: search error: %s", errors.ErrQueryFailed, string(bodyBytes))
	}

	return parseSearchResponse(bodyBytes, time.Since(start))
}

// esSearchResponse maps the relevant fields of an Elasticsearch search response.
type esSearchResponse struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source map[string]interface{} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

// parseSearchResponse converts raw ES JSON into a QueryResult.
// Columns are derived from the first hit's _source keys; each _source becomes a row.
// If there are no hits, an empty result set is returned.
// If JSON parsing fails, the raw body is returned as a single-row fallback.
func parseSearchResponse(body []byte, execTime time.Duration) (*types.QueryResult, error) {
	var esResp esSearchResponse
	if err := json.Unmarshal(body, &esResp); err != nil {
		columns := []types.ColumnInfo{{Name: "response", Type: "json"}}
		rows := []types.Row{{string(body)}}
		return types.NewQueryResult(columns, rows, execTime), nil
	}

	hits := esResp.Hits.Hits
	if len(hits) == 0 {
		return types.NewQueryResult(nil, nil, execTime), nil
	}

	firstSource := hits[0].Source
	columns := make([]types.ColumnInfo, 0, len(firstSource))
	colIndex := make(map[string]int, len(firstSource))
	for key := range firstSource {
		colIndex[key] = len(columns)
		columns = append(columns, types.ColumnInfo{Name: key, Type: "text"})
	}

	rows := make([]types.Row, 0, len(hits))
	for _, hit := range hits {
		row := make(types.Row, len(columns))
		for key, val := range hit.Source {
			if idx, ok := colIndex[key]; ok {
				row[idx] = val
			}
		}
		rows = append(rows, row)
	}

	return types.NewQueryResult(columns, rows, execTime), nil
}

func (c *Connection) Exec(ctx context.Context, query string) (*types.ExecResult, error) {
	if c.client == nil {
		return nil, errors.ErrConnectionClosed
	}

	start := time.Now()

	req := esapi.IndexRequest{
		Index: c.index,
		Body:  strings.NewReader(query),
	}
	res, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, fmt.Errorf("%w: index failed: %v", errors.ErrQueryFailed, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("%w: index error: %s", errors.ErrQueryFailed, res.String())
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
