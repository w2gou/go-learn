package command

import (
	"fmt"
	"github.com/spf13/cobra"
	"go-learn/cli/sshConnect/version2/config"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "查看所有连接",
	Run: func(cmd *cobra.Command, args []string) {
		conns, err := config.LoadConnections()
		if err != nil {
			fmt.Println("❌ 加载配置失败:", err)
			return
		}
		if len(conns) == 0 {
			fmt.Println("暂无保存的连接")
			return
		}
		fmt.Println("已保存的连接：")
		for _, c := range conns {
			fmt.Printf("- %s (%s@%s:%d)\n", c.Name, c.User, c.Host, c.Port)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
