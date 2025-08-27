package command

import (
	"fmt"
	"github.com/spf13/cobra"
	"go-learn/cli/sshConnect/version2/config"
)

var (
	name, host, user, password, keyPath string
	port                                int
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "新增一个SSH连接配置",
	Run: func(cmd *cobra.Command, args []string) {
		conn := config.Connection{
			Name:     name,
			Host:     host,
			Port:     port,
			User:     user,
			Password: password,
			KeyPath:  keyPath,
		}
		if err := config.AddConnection(conn); err != nil {
			fmt.Println("❌ 保存失败:", err)
			return
		}
		fmt.Println("✅ 连接保存成功:", conn.Name)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVar(&name, "name", "", "连接名称")
	addCmd.Flags().StringVar(&host, "host", "", "服务器地址")
	addCmd.Flags().IntVar(&port, "port", 22, "端口 (默认: 22)")
	addCmd.Flags().StringVar(&user, "user", "root", "用户名")
	addCmd.Flags().StringVar(&password, "password", "", "密码")
	addCmd.Flags().StringVar(&keyPath, "key", "~/.ssh/id_rsa", "私钥路径 (默认: ~/.ssh/id_rsa)")
}
