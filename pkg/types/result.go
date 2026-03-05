package types

import "time"

// ColumnInfo 定义查询结果中的列信息
type ColumnInfo struct {
	Name string `json:"name"` // 列名
	Type string `json:"type"` // 数据类型
}

// Row 定义查询结果中的一行数据
type Row []interface{}

// QueryResult 定义查询结果
type QueryResult struct {
	Columns       []ColumnInfo  `json:"columns"`       // 列信息
	Rows          []Row         `json:"rows"`          // 行数据
	RowCount      int           `json:"rowCount"`      // 行数
	ExecutionTime time.Duration `json:"executionTime"` // 执行时间
}

// NewQueryResult 创建一个新的查询结果
func NewQueryResult(columns []ColumnInfo, rows []Row, execTime time.Duration) *QueryResult {
	return &QueryResult{
		Columns:       columns,
		Rows:          rows,
		RowCount:      len(rows),
		ExecutionTime: execTime,
	}
}

// GetColumnNames 获取所有列名
func (r *QueryResult) GetColumnNames() []string {
	names := make([]string, len(r.Columns))
	for i, col := range r.Columns {
		names[i] = col.Name
	}
	return names
}

// GetRowByIndex 根据索引获取行
func (r *QueryResult) GetRowByIndex(index int) (Row, bool) {
	if index < 0 || index >= len(r.Rows) {
		return nil, false
	}
	return r.Rows[index], true
}

// GetValue 根据列名和行索引获取值
func (r *QueryResult) GetValue(rowIndex int, columnName string) (interface{}, bool) {
	row, ok := r.GetRowByIndex(rowIndex)
	if !ok {
		return nil, false
	}

	// 查找列索引
	for i, col := range r.Columns {
		if col.Name == columnName {
			if i < len(row) {
				return row[i], true
			}
			return nil, false
		}
	}

	return nil, false
}

// ExecResult 定义执行结果（用于 INSERT/UPDATE/DELETE）
type ExecResult struct {
	RowsAffected  int64         `json:"rowsAffected"`  // 受影响的行数
	LastInsertID  int64         `json:"lastInsertId"`  // 最后插入的 ID
	ExecutionTime time.Duration `json:"executionTime"` // 执行时间
}

// NewExecResult 创建一个新的执行结果
func NewExecResult(rowsAffected, lastInsertID int64, execTime time.Duration) *ExecResult {
	return &ExecResult{
		RowsAffected:  rowsAffected,
		LastInsertID:  lastInsertID,
		ExecutionTime: execTime,
	}
}
