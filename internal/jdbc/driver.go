package jdbc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"MystiSql/pkg/errors"
	"MystiSql/pkg/types"
)

// Driver JDBC 驱动实现
type Driver struct {
	// 服务器地址
	serverURL string

	// HTTP 客户端
	httpClient *http.Client

	// 连接池
	connections map[string]*Connection

	// 互斥锁
	mu sync.RWMutex
}

// Connection JDBC 连接实现
type Connection struct {
	// 驱动实例
	driver *Driver

	// 数据库实例名称
	instanceName string

	// 用户名
	username string

	// 密码
	password string

	// 连接状态
	closed bool

	// 连接创建时间
	createdAt time.Time

	// 最后使用时间
	lastUsedAt time.Time
}

// Statement JDBC 语句实现
type Statement struct {
	// 连接实例
	conn *Connection

	// SQL 语句
	sql string

	// 结果集
	resultSet *ResultSet
}

// ResultSet JDBC 结果集实现
type ResultSet struct {
	// 列信息
	columns []types.ColumnInfo

	// 行数据
	rows []types.Row

	// 当前行索引
	currentRow int

	// 结果集是否关闭
	closed bool
}

// PreparedStatement JDBC 预编译语句实现
type PreparedStatement struct {
	// 连接实例
	conn *Connection

	// SQL 语句
	sql string

	// 参数
	parameters []interface{}
}

// ResultSetMetaData 结果集元数据实现
type ResultSetMetaData struct {
	// 列信息
	columns []types.ColumnInfo
}

// DatabaseMetaData 数据库元数据实现
type DatabaseMetaData struct {
	// 连接实例
	conn *Connection
}

// NewDriver 创建一个新的 JDBC 驱动实例
func NewDriver(serverURL string) *Driver {
	return &Driver{
		serverURL: strings.TrimRight(serverURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		connections: make(map[string]*Connection),
	}
}

// Connect 创建一个新的 JDBC 连接
func (d *Driver) Connect(url string) (*Connection, error) {
	// 解析连接 URL
	instanceName, username, password, err := d.parseURL(url)
	if err != nil {
		return nil, err
	}

	// 创建连接实例
	conn := &Connection{
		driver:       d,
		instanceName: instanceName,
		username:     username,
		password:     password,
		closed:       false,
		createdAt:    time.Now(),
		lastUsedAt:   time.Now(),
	}

	// 缓存连接
	connID := fmt.Sprintf("%s:%s", instanceName, username)
	d.mu.Lock()
	d.connections[connID] = conn
	d.mu.Unlock()

	return conn, nil
}

// parseURL 解析 JDBC 连接 URL
// 格式: jdbc:mystisql://<server-url>/<instance-name>?user=<username>&password=<password>
func (d *Driver) parseURL(url string) (instanceName, username, password string, err error) {
	// 检查 URL 前缀
	prefix := "jdbc:mystisql://"
	if !strings.HasPrefix(url, prefix) {
		return "", "", "", fmt.Errorf("invalid JDBC URL format")
	}

	// 提取服务器地址和实例名
	remainder := strings.TrimPrefix(url, prefix)
	parts := strings.SplitN(remainder, "/", 2)
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("invalid JDBC URL format: missing instance name")
	}

	// 提取实例名和参数
	instanceAndParams := parts[1]
	paramsParts := strings.SplitN(instanceAndParams, "?", 2)
	instanceName = paramsParts[0]

	// 解析参数
	if len(paramsParts) == 2 {
		params := paramsParts[1]
		paramPairs := strings.Split(params, "&")
		for _, pair := range paramPairs {
			keyValue := strings.SplitN(pair, "=", 2)
			if len(keyValue) == 2 {
				key := strings.TrimSpace(keyValue[0])
				value := strings.TrimSpace(keyValue[1])
				switch key {
				case "user":
					username = value
				case "password":
					password = value
				}
			}
		}
	}

	if instanceName == "" {
		return "", "", "", fmt.Errorf("invalid JDBC URL format: missing instance name")
	}

	return instanceName, username, password, nil
}

// Close 关闭驱动并释放所有连接
func (d *Driver) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	var errs []error
	for id, conn := range d.connections {
		if err := conn.Close(); err != nil {
			errs = append(errs, err)
		}
		delete(d.connections, id)
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}

	return nil
}

// GetConnection 获取一个连接
func (d *Driver) GetConnection(instanceName, username, password string) (*Connection, error) {
	connID := fmt.Sprintf("%s:%s", instanceName, username)

	d.mu.RLock()
	if conn, exists := d.connections[connID]; exists && !conn.closed {
		conn.lastUsedAt = time.Now()
		d.mu.RUnlock()
		return conn, nil
	}
	d.mu.RUnlock()

	// 创建新连接
	return d.Connect(fmt.Sprintf("jdbc:mystisql://%s/%s?user=%s&password=%s", d.serverURL, instanceName, username, password))
}

// Close 关闭连接
func (c *Connection) Close() error {
	if c.closed {
		return nil
	}

	c.closed = true

	// 从驱动的连接池中移除
	connID := fmt.Sprintf("%s:%s", c.instanceName, c.username)
	c.driver.mu.Lock()
	delete(c.driver.connections, connID)
	c.driver.mu.Unlock()

	return nil
}

// CreateStatement 创建一个语句对象
func (c *Connection) CreateStatement() (*Statement, error) {
	if c.closed {
		return nil, errors.ErrConnectionClosed
	}

	c.lastUsedAt = time.Now()

	return &Statement{
		conn:      c,
		resultSet: nil,
	}, nil
}

// PrepareStatement 创建一个预编译语句对象
func (c *Connection) PrepareStatement(sql string) (*PreparedStatement, error) {
	if c.closed {
		return nil, errors.ErrConnectionClosed
	}

	c.lastUsedAt = time.Now()

	return &PreparedStatement{
		conn:       c,
		sql:        sql,
		parameters: make([]interface{}, 0),
	}, nil
}

// ExecuteQuery 执行查询语句
func (s *Statement) ExecuteQuery(sql string) (*ResultSet, error) {
	if s.conn.closed {
		return nil, errors.ErrConnectionClosed
	}

	// 执行查询
	ctx := context.Background()
	result, err := s.conn.executeQuery(ctx, sql)
	if err != nil {
		return nil, err
	}

	// 创建结果集
	s.resultSet = &ResultSet{
		columns:    result.Columns,
		rows:       result.Rows,
		currentRow: -1,
		closed:     false,
	}

	return s.resultSet, nil
}

// ExecuteUpdate 执行更新语句
func (s *Statement) ExecuteUpdate(sql string) (int64, error) {
	if s.conn.closed {
		return 0, errors.ErrConnectionClosed
	}

	// 执行更新
	ctx := context.Background()
	result, err := s.conn.executeUpdate(ctx, sql)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected, nil
}

// Close 关闭语句
func (s *Statement) Close() error {
	if s.resultSet != nil {
		_ = s.resultSet.Close()
	}

	return nil
}

// Next 移动到下一行
func (rs *ResultSet) Next() bool {
	if rs.closed {
		return false
	}

	rs.currentRow++
	return rs.currentRow < len(rs.rows)
}

// GetString 获取字符串值
func (rs *ResultSet) GetString(columnIndex int) (string, error) {
	if rs.closed {
		return "", errors.ErrResultSetClosed
	}

	if rs.currentRow < 0 || rs.currentRow >= len(rs.rows) {
		return "", fmt.Errorf("invalid row index")
	}

	if columnIndex < 1 || columnIndex > len(rs.rows[rs.currentRow]) {
		return "", fmt.Errorf("invalid column index")
	}

	value := rs.rows[rs.currentRow][columnIndex-1]
	if value == nil {
		return "", nil
	}

	if str, ok := value.(string); ok {
		return str, nil
	}

	return fmt.Sprintf("%v", value), nil
}

// GetInt 获取整数值
func (rs *ResultSet) GetInt(columnIndex int) (int, error) {
	if rs.closed {
		return 0, errors.ErrResultSetClosed
	}

	if rs.currentRow < 0 || rs.currentRow >= len(rs.rows) {
		return 0, fmt.Errorf("invalid row index")
	}

	if columnIndex < 1 || columnIndex > len(rs.rows[rs.currentRow]) {
		return 0, fmt.Errorf("invalid column index")
	}

	value := rs.rows[rs.currentRow][columnIndex-1]
	if value == nil {
		return 0, nil
	}

	if i, ok := value.(int); ok {
		return i, nil
	}

	return 0, fmt.Errorf("value is not an integer")
}

// GetMetaData 获取结果集元数据
func (rs *ResultSet) GetMetaData() *ResultSetMetaData {
	return &ResultSetMetaData{
		columns: rs.columns,
	}
}

// Close 关闭结果集
func (rs *ResultSet) Close() error {
	rs.closed = true
	return nil
}

// SetString 设置字符串参数
func (ps *PreparedStatement) SetString(index int, value string) error {
	if index < 1 {
		return fmt.Errorf("invalid parameter index")
	}

	// 扩展参数切片
	for len(ps.parameters) < index {
		ps.parameters = append(ps.parameters, nil)
	}

	ps.parameters[index-1] = value
	return nil
}

// SetInt 设置整数参数
func (ps *PreparedStatement) SetInt(index int, value int) error {
	if index < 1 {
		return fmt.Errorf("invalid parameter index")
	}

	// 扩展参数切片
	for len(ps.parameters) < index {
		ps.parameters = append(ps.parameters, nil)
	}

	ps.parameters[index-1] = value
	return nil
}

// ExecuteQuery 执行预编译查询
func (ps *PreparedStatement) ExecuteQuery() (*ResultSet, error) {
	if ps.conn.closed {
		return nil, errors.ErrConnectionClosed
	}

	// 替换参数
	sql := ps.sql
	for i, param := range ps.parameters {
		placeholder := fmt.Sprintf("?%d", i+1)
		if param == nil {
			sql = strings.Replace(sql, placeholder, "NULL", 1)
		} else {
			sql = strings.Replace(sql, placeholder, fmt.Sprintf("'%v'", param), 1)
		}
	}

	// 执行查询
	ctx := context.Background()
	result, err := ps.conn.executeQuery(ctx, sql)
	if err != nil {
		return nil, err
	}

	// 创建结果集
	return &ResultSet{
		columns:    result.Columns,
		rows:       result.Rows,
		currentRow: -1,
		closed:     false,
	}, nil
}

// ExecuteUpdate 执行预编译更新
func (ps *PreparedStatement) ExecuteUpdate() (int64, error) {
	if ps.conn.closed {
		return 0, errors.ErrConnectionClosed
	}

	// 替换参数
	sql := ps.sql
	for i, param := range ps.parameters {
		placeholder := fmt.Sprintf("?%d", i+1)
		if param == nil {
			sql = strings.Replace(sql, placeholder, "NULL", 1)
		} else {
			sql = strings.Replace(sql, placeholder, fmt.Sprintf("'%v'", param), 1)
		}
	}

	// 执行更新
	ctx := context.Background()
	result, err := ps.conn.executeUpdate(ctx, sql)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected, nil
}

// Close 关闭预编译语句
func (ps *PreparedStatement) Close() error {
	return nil
}

// GetColumnCount 获取列数
func (md *ResultSetMetaData) GetColumnCount() int {
	return len(md.columns)
}

// GetColumnName 获取列名
func (md *ResultSetMetaData) GetColumnName(columnIndex int) string {
	if columnIndex < 1 || columnIndex > len(md.columns) {
		return ""
	}

	return md.columns[columnIndex-1].Name
}

// GetColumnType 获取列类型
func (md *ResultSetMetaData) GetColumnType(columnIndex int) string {
	if columnIndex < 1 || columnIndex > len(md.columns) {
		return ""
	}

	return md.columns[columnIndex-1].Type
}

// GetDatabaseMetaData 获取数据库元数据
func (c *Connection) GetDatabaseMetaData() *DatabaseMetaData {
	return &DatabaseMetaData{
		conn: c,
	}
}

// getURL 获取数据库 URL
func (md *DatabaseMetaData) getURL() string {
	return fmt.Sprintf("jdbc:mystisql://%s/%s", md.conn.driver.serverURL, md.conn.instanceName)
}

// getUserName 获取用户名
func (md *DatabaseMetaData) getUserName() string {
	return md.conn.username
}

// executeQuery 执行查询
func (c *Connection) executeQuery(ctx context.Context, sql string) (*types.QueryResult, error) {
	// 构建请求 URL
	url := fmt.Sprintf("%s/api/v1/instances/%s/query", c.driver.serverURL, c.instanceName)

	// 构建请求体
	reqBody := fmt.Sprintf(`{"sql": "%s"}`, sql)

	// 发送请求
	resp, err := c.driver.httpClient.Post(url, "application/json", strings.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	// 解析响应
	var result types.QueryResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// executeUpdate 执行更新
func (c *Connection) executeUpdate(ctx context.Context, sql string) (*types.ExecResult, error) {
	// 构建请求 URL
	url := fmt.Sprintf("%s/api/v1/instances/%s/exec", c.driver.serverURL, c.instanceName)

	// 构建请求体
	reqBody := fmt.Sprintf(`{"sql": "%s"}`, sql)

	// 发送请求
	resp, err := c.driver.httpClient.Post(url, "application/json", strings.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	// 解析响应
	var result types.ExecResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
