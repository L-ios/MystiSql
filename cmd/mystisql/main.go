package main

import (
	"fmt"
	"os"

	cli "MystiSql/internal/cli"
)

var (
	version = "0.1.0"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	cli.SetVersion(version, commit, date)

	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}
