package cli

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"MystiSql/internal/connection"
	"MystiSql/internal/service/query"
	"MystiSql/pkg/types"

	"github.com/spf13/cobra"
)

var (
	instancesFormat string
)

// instancesCmd 实例命令
var instancesCmd = &cobra.Command{
	Use:     "instances",
	Aliases: []string{"instance", "inst"},
	Short:   "管理数据库实例",
	Long:    `管理数据库实例，包括列出、查看实例信息等操作。`,
}

// instancesListCmd 列出实例命令
var instancesListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "列出所有数据库实例",
	Long:    `列出所有已注册的数据库实例。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 获取实例列表
		instances, err := GetRegistry().ListInstances()
		if err != nil {
			return fmt.Errorf("获取实例列表失败: %w", err)
		}

		if len(instances) == 0 {
			fmt.Println("没有找到任何实例")
			return nil
		}

		// 根据格式输出
		switch instancesFormat {
		case "json":
			return outputInstancesJSON(instances)
		case "csv":
			return outputInstancesCSV(instances)
		default:
			return outputInstancesTable(instances)
		}
	},
}

// instancesGetCmd 获取单个实例命令
var instancesGetCmd = &cobra.Command{
	Use:   "get <instance-name>",
	Short: "获取单个实例的详细信息",
	Long:  `获取指定名称的数据库实例的详细信息。`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		instanceName := args[0]

		// 获取实例
		instance, err := GetRegistry().GetInstance(instanceName)
		if err != nil {
			return fmt.Errorf("获取实例失败: %w", err)
		}

		// 根据格式输出
		switch instancesFormat {
		case "json":
			return outputInstanceJSON(instance)
		case "csv":
			return outputInstanceCSV(instance)
		default:
			return outputInstanceDetail(instance)
		}
	},
}

// instancesHealthCmd 检查实例健康状态命令
var instancesHealthCmd = &cobra.Command{
	Use:   "health <instance-name>",
	Short: "检查实例健康状态",
	Long:  `检查指定名称的数据库实例的健康状态。`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		instanceName := args[0]

		// 创建 query engine
		engine := query.NewEngine(GetRegistry(), connection.GetRegistry())

		// 创建带超时的上下文
		ctx, cancel := context.WithTimeout(GetContext(), 10*time.Second)
		defer cancel()

		// 检查健康状态
		status, err := engine.GetInstanceHealth(ctx, instanceName)
		if err != nil {
			return fmt.Errorf("检查健康状态失败: %w", err)
		}

		// 根据格式输出
		switch instancesFormat {
		case "json":
			return outputHealthStatusJSON(instanceName, status)
		case "csv":
			return outputHealthStatusCSV(instanceName, status)
		default:
			return outputHealthStatusTable(instanceName, status)
		}
	},
}

// instancesPoolCmd 查看连接池统计信息命令
var instancesPoolCmd = &cobra.Command{
	Use:   "pool <instance-name>",
	Short: "查看连接池统计信息",
	Long:  `查看指定名称的数据库实例的连接池统计信息。`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		instanceName := args[0]

		// 创建 query engine
		engine := query.NewEngine(GetRegistry(), connection.GetRegistry())

		// 获取连接池统计信息
		stats, err := engine.GetPoolStats(cmd.Context(), instanceName)
		if err != nil {
			return fmt.Errorf("获取连接池统计信息失败: %w", err)
		}

		// 根据格式输出
		switch instancesFormat {
		case "json":
			return outputPoolStatsJSON(instanceName, stats)
		case "csv":
			return outputPoolStatsCSV(instanceName, stats)
		default:
			return outputPoolStatsTable(instanceName, stats)
		}
	},
}

func init() {
	// 添加子命令
	instancesCmd.AddCommand(instancesListCmd)
	instancesCmd.AddCommand(instancesGetCmd)
	instancesCmd.AddCommand(instancesHealthCmd)
	instancesCmd.AddCommand(instancesPoolCmd)

	// 添加全局标志
	instancesCmd.PersistentFlags().StringVarP(&instancesFormat, "format", "f", "table", "输出格式 (table, json, csv)")

	// 添加到根命令
	rootCmd.AddCommand(instancesCmd)
}

// outputInstancesTable 以表格格式输出实例列表
func outputInstancesTable(instances []*types.DatabaseInstance) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	// 打印表头
	if _, err := fmt.Fprintln(w, "NAME\tTYPE\tHOST\tPORT\tDATABASE\tSTATUS"); err != nil {
		return fmt.Errorf("写入表头失败: %w", err)
	}

	// 打印数据
	for _, inst := range instances {
		database := inst.Database
		if database == "" {
			database = "-"
		}
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n",
			inst.Name,
			inst.Type,
			inst.Host,
			inst.Port,
			database,
			inst.Status,
		); err != nil {
			return fmt.Errorf("写入数据失败: %w", err)
		}
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("刷新输出失败: %w", err)
	}

	return nil
}

// outputInstancesJSON 以 JSON 格式输出实例列表
func outputInstancesJSON(instances []*types.DatabaseInstance) error {
	jsonData, err := json.MarshalIndent(instances, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 JSON 失败: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

// outputInstancesCSV 以 CSV 格式输出实例列表
func outputInstancesCSV(instances []*types.DatabaseInstance) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// 写入表头
	if err := writer.Write([]string{"name", "type", "host", "port", "database", "status"}); err != nil {
		return fmt.Errorf("写入 CSV 表头失败: %w", err)
	}

	// 写入数据
	for _, inst := range instances {
		database := inst.Database
		if database == "" {
			database = ""
		}
		record := []string{
			inst.Name,
			string(inst.Type),
			inst.Host,
			fmt.Sprintf("%d", inst.Port),
			database,
			string(inst.Status),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("写入 CSV 数据失败: %w", err)
		}
	}

	return nil
}

// outputInstanceJSON 以 JSON 格式输出单个实例
func outputInstanceJSON(instance *types.DatabaseInstance) error {
	jsonData, err := json.MarshalIndent(instance, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 JSON 失败: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

// outputInstanceCSV 以 CSV 格式输出单个实例
func outputInstanceCSV(instance *types.DatabaseInstance) error {
	return outputInstancesCSV([]*types.DatabaseInstance{instance})
}

// outputInstanceDetail 以详细信息格式输出单个实例
func outputInstanceDetail(instance *types.DatabaseInstance) error {
	fmt.Printf("实例名称: %s\n", instance.Name)
	fmt.Printf("数据库类型: %s\n", instance.Type)
	fmt.Printf("主机地址: %s\n", instance.Host)
	fmt.Printf("端口号: %d\n", instance.Port)
	if instance.Database != "" {
		fmt.Printf("数据库名: %s\n", instance.Database)
	}
	if instance.Username != "" {
		fmt.Printf("用户名: %s\n", instance.Username)
	}
	fmt.Printf("状态: %s\n", instance.Status)
	fmt.Printf("创建时间: %s\n", instance.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("更新时间: %s\n", instance.UpdatedAt.Format("2006-01-02 15:04:05"))

	if len(instance.Labels) > 0 {
		fmt.Println("\n标签:")
		for k, v := range instance.Labels {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	if len(instance.Annotations) > 0 {
		fmt.Println("\n注解:")
		for k, v := range instance.Annotations {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	return nil
}

// outputHealthStatusTable 以表格格式输出健康状态
func outputHealthStatusTable(instanceName string, status types.InstanceStatus) error {
	fmt.Printf("实例名称: %s\n", instanceName)
	fmt.Printf("健康状态: %s\n", status)
	fmt.Printf("检查时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	return nil
}

// outputHealthStatusJSON 以 JSON 格式输出健康状态
func outputHealthStatusJSON(instanceName string, status types.InstanceStatus) error {
	data := map[string]interface{}{
		"instance":  instanceName,
		"status":    status,
		"timestamp": time.Now(),
	}
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 JSON 失败: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

// outputHealthStatusCSV 以 CSV 格式输出健康状态
func outputHealthStatusCSV(instanceName string, status types.InstanceStatus) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// 写入表头
	if err := writer.Write([]string{"instance", "status", "timestamp"}); err != nil {
		return fmt.Errorf("写入 CSV 表头失败: %w", err)
	}

	// 写入数据
	record := []string{
		instanceName,
		string(status),
		time.Now().Format("2006-01-02 15:04:05"),
	}
	if err := writer.Write(record); err != nil {
		return fmt.Errorf("写入 CSV 数据失败: %w", err)
	}

	return nil
}

// outputPoolStatsTable 以表格格式输出连接池统计信息
func outputPoolStatsTable(instanceName string, stats *connection.PoolStats) error {
	fmt.Printf("实例名称: %s\n", instanceName)
	fmt.Printf("总连接数: %d\n", stats.TotalConnections)
	fmt.Printf("空闲连接数: %d\n", stats.IdleConnections)
	fmt.Printf("活跃连接数: %d\n", stats.ActiveConnections)
	fmt.Printf("最大连接数: %d\n", stats.MaxConnections)
	fmt.Printf("最小连接数: %d\n", stats.MinConnections)
	fmt.Printf("获取连接次数: %d\n", stats.AcquireCount)
	fmt.Printf("获取连接失败次数: %d\n", stats.AcquireFailed)
	fmt.Printf("释放连接次数: %d\n", stats.ReleaseCount)
	fmt.Printf("统计时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	return nil
}

// outputPoolStatsJSON 以 JSON 格式输出连接池统计信息
func outputPoolStatsJSON(instanceName string, stats *connection.PoolStats) error {
	data := map[string]interface{}{
		"instance":  instanceName,
		"stats":     stats,
		"timestamp": time.Now(),
	}
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 JSON 失败: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

// outputPoolStatsCSV 以 CSV 格式输出连接池统计信息
func outputPoolStatsCSV(instanceName string, stats *connection.PoolStats) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// 写入表头
	if err := writer.Write([]string{
		"instance",
		"total_connections",
		"idle_connections",
		"active_connections",
		"max_connections",
		"min_connections",
		"acquire_count",
		"acquire_failed",
		"release_count",
		"timestamp",
	}); err != nil {
		return fmt.Errorf("写入 CSV 表头失败: %w", err)
	}

	// 写入数据
	record := []string{
		instanceName,
		fmt.Sprintf("%d", stats.TotalConnections),
		fmt.Sprintf("%d", stats.IdleConnections),
		fmt.Sprintf("%d", stats.ActiveConnections),
		fmt.Sprintf("%d", stats.MaxConnections),
		fmt.Sprintf("%d", stats.MinConnections),
		fmt.Sprintf("%d", stats.AcquireCount),
		fmt.Sprintf("%d", stats.AcquireFailed),
		fmt.Sprintf("%d", stats.ReleaseCount),
		time.Now().Format("2006-01-02 15:04:05"),
	}
	if err := writer.Write(record); err != nil {
		return fmt.Errorf("写入 CSV 数据失败: %w", err)
	}

	return nil
}
