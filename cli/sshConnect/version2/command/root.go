package command

import (
	"fmt"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gossh",
	Short: "Go SSH & SFTP CLI 工具",
	Long:  `一个用 Go 实现的简易 SSH & SFTP 工具，支持连接管理与交互式连接`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("执行出错:", err)
	}
}
