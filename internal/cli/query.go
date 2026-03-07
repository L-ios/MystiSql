package cli

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"MystiSql/internal/service/query"
	"MystiSql/pkg/types"

	"github.com/spf13/cobra"
)

var (
	queryFormat  string
	queryTimeout time.Duration
)

// queryCmd 查询命令
var queryCmd = &cobra.Command{
	Use:   "query <instance-name> <sql>",
	Short: "在指定数据库实例上执行 SQL 查询",
	Long: `在指定的数据库实例上执行 SQL 查询语句。

支持 SELECT、INSERT、UPDATE、DELETE 等各类 SQL 语句。
输出格式支持表格、JSON、CSV 三种格式。`,
	Example: `  # 执行查询
  mystisql query my-mysql "SELECT * FROM users LIMIT 10"
  
  # 以 JSON 格式输出
  mystisql query my-mysql "SELECT * FROM users LIMIT 10" --format json
  
  # 执行 INSERT 语句
  mystisql query my-mysql "INSERT INTO users (name, email) VALUES ('Alice', 'alice@example.com')"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		instanceName := args[0]
		sqlQuery := args[1]

		GetSugar().Debugf("在实例 %s 上执行查询: %s", instanceName, sqlQuery)

		// 创建 query engine
		engine := query.NewEngine(GetRegistry())

		// 创建带超时的上下文
		ctx, cancel := context.WithTimeout(GetContext(), queryTimeout)
		defer cancel()

		// 判断是查询还是执行语句
		upperQuery := strings.ToUpper(strings.TrimSpace(sqlQuery))
		isQuery := strings.HasPrefix(upperQuery, "SELECT") ||
			strings.HasPrefix(upperQuery, "SHOW") ||
			strings.HasPrefix(upperQuery, "DESC") ||
			strings.HasPrefix(upperQuery, "EXPLAIN")

		var result interface{}
		if isQuery {
			// 执行查询
			queryResult, err := engine.ExecuteQuery(ctx, instanceName, sqlQuery)
			if err != nil {
				return fmt.Errorf("执行查询失败: %w", err)
			}
			result = queryResult
			GetSugar().Debugf("查询完成，返回 %d 行，耗时 %v", len(queryResult.Rows), queryResult.ExecutionTime)
		} else {
			// 执行非查询语句
			execResult, err := engine.ExecuteExec(ctx, instanceName, sqlQuery)
			if err != nil {
				return fmt.Errorf("执行语句失败: %w", err)
			}
			result = execResult
			GetSugar().Debugf("执行完成，影响 %d 行，耗时 %v", execResult.RowsAffected, execResult.ExecutionTime)
		}

		// 根据格式输出结果
		switch queryFormat {
		case "json":
			return outputQueryResultJSON(result, isQuery)
		case "csv":
			if isQuery {
				return outputQueryResultCSV(result.(*types.QueryResult))
			}
			return outputExecResultCSV(result.(*types.ExecResult))
		default:
			if isQuery {
				return outputQueryResultTable(result.(*types.QueryResult))
			}
			return outputExecResultTable(result.(*types.ExecResult))
		}
	},
}

func init() {
	// 添加标志
	queryCmd.Flags().StringVarP(&queryFormat, "format", "f", "table", "输出格式 (table, json, csv)")
	queryCmd.Flags().DurationVarP(&queryTimeout, "timeout", "t", 30*time.Second, "查询超时时间")

	// 添加到根命令
	rootCmd.AddCommand(queryCmd)
}

// outputQueryResultTable 以表格格式输出查询结果
func outputQueryResultTable(result *types.QueryResult) error {
	if len(result.Columns) == 0 {
		fmt.Println("查询成功，但没有返回数据")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	// 打印表头
	headers := make([]string, len(result.Columns))
	for i, col := range result.Columns {
		headers[i] = col.Name
	}
	if _, err := fmt.Fprintln(w, strings.Join(headers, "\t")); err != nil {
		return fmt.Errorf("写入表头失败: %w", err)
	}

	// 打印数据
	for _, row := range result.Rows {
		values := make([]string, len(row))
		for i, val := range row {
			if val == nil {
				values[i] = "NULL"
			} else {
				values[i] = fmt.Sprintf("%v", val)
			}
		}
		if _, err := fmt.Fprintln(w, strings.Join(values, "\t")); err != nil {
			return fmt.Errorf("写入数据失败: %w", err)
		}
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("刷新输出失败: %w", err)
	}

	// 打印统计信息
	fmt.Printf("\n%d 行数据，执行时间: %v\n", result.RowCount, result.ExecutionTime)

	return nil
}

// outputQueryResultJSON 以 JSON 格式输出查询结果
func outputQueryResultJSON(result interface{}, isQuery bool) error {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 JSON 失败: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

// outputQueryResultCSV 以 CSV 格式输出查询结果
func outputQueryResultCSV(result *types.QueryResult) error {
	if len(result.Columns) == 0 {
		return nil
	}

	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// 写入表头
	headers := make([]string, len(result.Columns))
	for i, col := range result.Columns {
		headers[i] = col.Name
	}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("写入 CSV 表头失败: %w", err)
	}

	// 写入数据
	for _, row := range result.Rows {
		values := make([]string, len(row))
		for i, val := range row {
			if val == nil {
				values[i] = ""
			} else {
				values[i] = fmt.Sprintf("%v", val)
			}
		}
		if err := writer.Write(values); err != nil {
			return fmt.Errorf("写入 CSV 数据失败: %w", err)
		}
	}

	return nil
}

// outputExecResultTable 以表格格式输出执行结果
func outputExecResultTable(result *types.ExecResult) error {
	fmt.Printf("受影响行数: %d\n", result.RowsAffected)
	if result.LastInsertID > 0 {
		fmt.Printf("最后插入ID: %d\n", result.LastInsertID)
	}
	fmt.Printf("执行时间: %v\n", result.ExecutionTime)
	return nil
}

// outputExecResultCSV 以 CSV 格式输出执行结果
func outputExecResultCSV(result *types.ExecResult) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// 写入表头
	if err := writer.Write([]string{"affected_rows", "last_insert_id", "execution_time"}); err != nil {
		return fmt.Errorf("写入 CSV 表头失败: %w", err)
	}

	// 写入数据
	record := []string{
		fmt.Sprintf("%d", result.RowsAffected),
		fmt.Sprintf("%d", result.LastInsertID),
		result.ExecutionTime.String(),
	}
	if err := writer.Write(record); err != nil {
		return fmt.Errorf("写入 CSV 数据失败: %w", err)
	}

	return nil
}
