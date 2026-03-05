package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd 版本命令
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Long:  `显示 MystiSql 的版本信息，包括版本号、Git 提交和构建日期。`,
	Run: func(cmd *cobra.Command, args []string) {
		if verbose {
			// 详细模式：输出 JSON 格式
			versionInfo := GetFullVersion()
			jsonData, err := json.MarshalIndent(versionInfo, "", "  ")
			if err != nil {
				fmt.Printf("版本: %s\n", GetVersion())
				return
			}
			fmt.Println(string(jsonData))
		} else {
			// 简单模式：只输出版本号
			fmt.Printf("MystiSql %s\n", GetVersion())
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
