package command

import (
	"fmt"
	"github.com/spf13/cobra"
	"go-learn/cli/sshConnect/version2/config"
	"go-learn/cli/sshConnect/version2/internal/sshclient"
)

var connectCmd = &cobra.Command{
	Use:   "connect [name]",
	Short: "连接到服务器 (交互式Shell)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		conns, err := config.LoadConnections()
		if err != nil {
			fmt.Println("❌ 读取配置失败:", err)
			return
		}
		var target *config.Connection
		for _, c := range conns {
			if c.Name == name {
				target = &c
				break
			}
		}
		if target == nil {
			fmt.Println("❌ 未找到连接:", name)
			return
		}
		fmt.Printf("🔗 正在连接 %s@%s:%d ...\n", target.User, target.Host, target.Port)
		if err := sshclient.InteractiveShell(*target); err != nil {
			fmt.Println("❌ SSH 连接失败:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)
}
