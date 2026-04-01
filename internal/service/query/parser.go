package query

import (
	"context"
	"fmt"
	"strings"
	"time"

	"MystiSql/pkg/types"
)

// SQLParser 定义 SQL 解析器接口
type SQLParser interface {
	// Parse 解析 SQL 语句
	// 返回解析结果，包含语句类型、目标表等信息
	Parse(sql string) (*SQLParseResult, error)

	// Validate 验证 SQL 语句是否合法
	Validate(sql string) error

	// SetMaxResultSize 设置最大结果集大小
	SetMaxResultSize(size int)

	// GetMaxResultSize 获取最大结果集大小
	GetMaxResultSize() int

	// SetQueryTimeout 设置查询超时时间
	SetQueryTimeout(timeout time.Duration)

	// GetQueryTimeout 获取查询超时时间
	GetQueryTimeout() time.Duration
}

// SQLStatementType 定义 SQL 语句类型
type SQLStatementType string

const (
	// StatementTypeSelect SELECT 语句
	StatementTypeSelect SQLStatementType = "SELECT"
	// StatementTypeInsert INSERT 语句
	StatementTypeInsert SQLStatementType = "INSERT"
	// StatementTypeUpdate UPDATE 语句
	StatementTypeUpdate SQLStatementType = "UPDATE"
	// StatementTypeDelete DELETE 语句
	StatementTypeDelete SQLStatementType = "DELETE"
	// StatementTypeCreate CREATE 语句
	StatementTypeCreate SQLStatementType = "CREATE"
	// StatementTypeDrop DROP 语句
	StatementTypeDrop SQLStatementType = "DROP"
	// StatementTypeAlter ALTER 语句
	StatementTypeAlter SQLStatementType = "ALTER"
	// StatementTypeOther 其他语句
	StatementTypeOther SQLStatementType = "OTHER"
)

// SQLParseResult 定义 SQL 解析结果
type SQLParseResult struct {
	// StatementType SQL 语句类型
	StatementType SQLStatementType

	// Tables 目标表列表
	Tables []string

	// IsReadOnly 是否为只读语句
	IsReadOnly bool

	// EstimatedSize 估计结果集大小
	EstimatedSize int

	// QueryTimeout 查询超时时间
	QueryTimeout time.Duration

	// MaxResultSize 最大结果集大小
	MaxResultSize int
}

// Parser 实现 SQLParser 接口
type Parser struct {
	// 最大结果集大小
	maxResultSize int

	// 查询超时时间
	queryTimeout time.Duration
}

// NewParser 创建一个新的 SQL 解析器
func NewParser() SQLParser {
	return &Parser{
		maxResultSize: 10000,            // 默认最大结果集大小为 10000
		queryTimeout:  30 * time.Second, // 默认查询超时时间为 30 秒
	}
}

// Parse 解析 SQL 语句
func (p *Parser) Parse(sql string) (*SQLParseResult, error) {
	// 去除 SQL 语句首尾空格
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return nil, fmt.Errorf("empty SQL statement")
	}

	// 解析语句类型
	statementType := p.parseStatementType(sql)

	// 解析目标表
	tables := p.parseTables(sql, statementType)

	// 判断是否为只读语句
	isReadOnly := p.isReadOnly(statementType)

	// 估计结果集大小
	estimatedSize := p.estimateResultSize(sql, statementType)

	return &SQLParseResult{
		StatementType: statementType,
		Tables:        tables,
		IsReadOnly:    isReadOnly,
		EstimatedSize: estimatedSize,
		QueryTimeout:  p.queryTimeout,
		MaxResultSize: p.maxResultSize,
	}, nil
}

// Validate 验证 SQL 语句是否合法
func (p *Parser) Validate(sql string) error {
	// 去除 SQL 语句首尾空格
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return fmt.Errorf("empty SQL statement")
	}

	// 检查是否包含危险操作
	dangerousKeywords := []string{
		"DROP DATABASE",
		"DROP TABLE",
		"TRUNCATE",
		"ALTER TABLE",
	}

	lowerSQL := strings.ToLower(sql)
	for _, keyword := range dangerousKeywords {
		if strings.Contains(lowerSQL, strings.ToLower(keyword)) {
			return fmt.Errorf("dangerous SQL operation detected: %s", keyword)
		}
	}

	// 检查是否包含注释
	if strings.Contains(lowerSQL, "--") || strings.Contains(lowerSQL, "/*") {
		// 简单检查，实际应用中可能需要更复杂的处理
	}

	return nil
}

// SetMaxResultSize 设置最大结果集大小
func (p *Parser) SetMaxResultSize(size int) {
	if size > 0 {
		p.maxResultSize = size
	}
}

// GetMaxResultSize 获取最大结果集大小
func (p *Parser) GetMaxResultSize() int {
	return p.maxResultSize
}

// SetQueryTimeout 设置查询超时时间
func (p *Parser) SetQueryTimeout(timeout time.Duration) {
	if timeout > 0 {
		p.queryTimeout = timeout
	}
}

// GetQueryTimeout 获取查询超时时间
func (p *Parser) GetQueryTimeout() time.Duration {
	return p.queryTimeout
}

// parseStatementType 解析 SQL 语句类型
func (p *Parser) parseStatementType(sql string) SQLStatementType {
	// 转换为大写以便比较
	upperSQL := strings.ToUpper(sql)

	// 检查语句类型
	switch {
	case strings.HasPrefix(upperSQL, "SELECT"):
		return StatementTypeSelect
	case strings.HasPrefix(upperSQL, "INSERT"):
		return StatementTypeInsert
	case strings.HasPrefix(upperSQL, "UPDATE"):
		return StatementTypeUpdate
	case strings.HasPrefix(upperSQL, "DELETE"):
		return StatementTypeDelete
	case strings.HasPrefix(upperSQL, "CREATE"):
		return StatementTypeCreate
	case strings.HasPrefix(upperSQL, "DROP"):
		return StatementTypeDrop
	case strings.HasPrefix(upperSQL, "ALTER"):
		return StatementTypeAlter
	default:
		return StatementTypeOther
	}
}

// parseTables 解析目标表
func (p *Parser) parseTables(sql string, statementType SQLStatementType) []string {
	// 简化实现，实际应用中可能需要更复杂的解析
	tables := make([]string, 0)

	// 转换为小写以便处理
	lowerSQL := strings.ToLower(sql)

	switch statementType {
	case StatementTypeSelect:
		// 查找 FROM 子句
		fromIndex := strings.Index(lowerSQL, " from ")
		if fromIndex != -1 {
			// 提取 FROM 子句后的表名
			fromClause := lowerSQL[fromIndex+6:]
			// 查找 WHERE、JOIN、GROUP BY 等后续子句
			endIndex := len(fromClause)
			for _, keyword := range []string{" where ", " join ", " group by ", " order by ", " limit "} {
				if idx := strings.Index(fromClause, keyword); idx != -1 && idx < endIndex {
					endIndex = idx
				}
			}
			tableClause := strings.TrimSpace(fromClause[:endIndex])
			// 简单处理，实际应用中可能需要处理别名、多表等情况
			tableParts := strings.Split(tableClause, ",")
			for _, part := range tableParts {
				part = strings.TrimSpace(part)
				if part != "" {
					// 去除别名
					if idx := strings.Index(part, " as "); idx != -1 {
						part = part[:idx]
					} else if idx := strings.LastIndex(part, " "); idx != -1 {
						// 处理没有 AS 的别名
						part = part[:idx]
					}
					tables = append(tables, strings.TrimSpace(part))
				}
			}
		}

	case StatementTypeInsert:
		// 查找 INTO 子句
		intoIndex := strings.Index(lowerSQL, " into ")
		if intoIndex != -1 {
			// 提取 INTO 子句后的表名
			intoClause := lowerSQL[intoIndex+6:]
			// 查找 VALUES 子句
			valuesIndex := strings.Index(intoClause, " values ")
			if valuesIndex != -1 {
				tableClause := strings.TrimSpace(intoClause[:valuesIndex])
				// 去除括号
				tableClause = strings.Trim(tableClause, "()")
				tables = append(tables, tableClause)
			}
		}

	case StatementTypeUpdate:
		updateClause := strings.TrimPrefix(lowerSQL, "update")
		updateClause = strings.TrimSpace(updateClause)
		setIndex := strings.Index(updateClause, " set ")
		if setIndex != -1 {
			tableClause := strings.TrimSpace(updateClause[:setIndex])
			tables = append(tables, tableClause)
		}

	case StatementTypeDelete:
		// 查找 FROM 子句
		fromIndex := strings.Index(lowerSQL, " from ")
		if fromIndex != -1 {
			// 提取 FROM 子句后的表名
			fromClause := lowerSQL[fromIndex+6:]
			// 查找 WHERE 子句
			whereIndex := strings.Index(fromClause, " where ")
			if whereIndex != -1 {
				tableClause := strings.TrimSpace(fromClause[:whereIndex])
				tables = append(tables, tableClause)
			} else {
				tableClause := strings.TrimSpace(fromClause)
				tables = append(tables, tableClause)
			}
		}

	case StatementTypeCreate:
		// 查找 CREATE TABLE 子句
		if strings.Contains(lowerSQL, "create table ") {
			tableIndex := strings.Index(lowerSQL, "create table ")
			tableClause := lowerSQL[tableIndex+13:]
			// 查找左括号
			leftParenIndex := strings.Index(tableClause, "(")
			if leftParenIndex != -1 {
				tableName := strings.TrimSpace(tableClause[:leftParenIndex])
				tables = append(tables, tableName)
			}
		}

	case StatementTypeDrop:
		// 查找 DROP TABLE 子句
		if strings.Contains(lowerSQL, "drop table ") {
			tableIndex := strings.Index(lowerSQL, "drop table ")
			tableClause := lowerSQL[tableIndex+11:]
			// 查找分号
			semicolonIndex := strings.Index(tableClause, ";")
			if semicolonIndex != -1 {
				tableName := strings.TrimSpace(tableClause[:semicolonIndex])
				tables = append(tables, tableName)
			} else {
				tableName := strings.TrimSpace(tableClause)
				tables = append(tables, tableName)
			}
		}

	case StatementTypeAlter:
		// 查找 ALTER TABLE 子句
		if strings.Contains(lowerSQL, "alter table ") {
			tableIndex := strings.Index(lowerSQL, "alter table ")
			tableClause := lowerSQL[tableIndex+12:]
			// 查找下一个关键字
			keywords := []string{" add ", " drop ", " modify ", " rename ", " change "}
			endIndex := len(tableClause)
			for _, keyword := range keywords {
				if idx := strings.Index(tableClause, keyword); idx != -1 && idx < endIndex {
					endIndex = idx
				}
			}
			tableName := strings.TrimSpace(tableClause[:endIndex])
			tables = append(tables, tableName)
		}
	}

	return tables
}

// isReadOnly 判断是否为只读语句
func (p *Parser) isReadOnly(statementType SQLStatementType) bool {
	switch statementType {
	case StatementTypeSelect:
		return true
	default:
		return false
	}
}

// estimateResultSize 估计结果集大小
func (p *Parser) estimateResultSize(sql string, statementType SQLStatementType) int {
	// 简化实现，实际应用中可能需要更复杂的估计
	if statementType != StatementTypeSelect {
		return 0
	}

	// 检查是否包含 LIMIT 子句
	lowerSQL := strings.ToLower(sql)
	limitIndex := strings.Index(lowerSQL, " limit ")
	if limitIndex != -1 {
		// 提取 LIMIT 子句的值
		limitClause := lowerSQL[limitIndex+7:]
		// 查找分号
		semicolonIndex := strings.Index(limitClause, ";")
		if semicolonIndex != -1 {
			limitClause = limitClause[:semicolonIndex]
		}
		// 简单处理，实际应用中可能需要解析数字
		return 100 // 默认估计值
	}

	// 检查是否包含 WHERE 子句
	if strings.Contains(lowerSQL, " where ") {
		return 500 // 有 WHERE 子句，估计结果集较小
	}

	// 没有 LIMIT 和 WHERE 子句，估计结果集较大
	return 1000
}

// WithTimeout 为上下文添加查询超时
func WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}

// WithResultSizeLimit 为查询结果添加大小限制
func WithResultSizeLimit(result *types.QueryResult, maxSize int) *types.QueryResult {
	if result == nil {
		return nil
	}

	if maxSize > 0 && len(result.Rows) > maxSize {
		result.Rows = result.Rows[:maxSize]
		result.Truncated = true
	}

	return result
}
