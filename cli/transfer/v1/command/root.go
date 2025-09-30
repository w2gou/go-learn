package transferV1Command

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var (
	// 版本信息
	version   string
	gitCommit string
	buildTime string

	// 配置文件路径
	cfgFile string
)

// rootCmd 根命令
var rootCmd = &cobra.Command{
	Use:   "go-transfer",
	Short: "二维码驱动的局域网临时文件分享CLI工具",
	Long: `Go-Transfer 是一个高性能的文件分享工具，支持：

• 🚀 高并发文件服务器 - 基于 Goroutine 的并发处理
• 📱 二维码访问 - 终端生成二维码，手机扫描即可访问  
• 🌐 自动网络发现 - 智能获取局域网IP地址
• ⚙️  灵活配置 - 支持配置文件和命令行参数
• 📊 性能监控 - 实时监控连接数和传输状态
• 🛡️  安全可控 - 临时服务器，支持访问控制

使用 Go 语言的高并发特性，为每个请求分配独立的 Goroutine，
提供企业级的稳定性和性能。`,
	Version: getVersionString(),
}

// Execute 执行根命令
func Execute() error {
	return rootCmd.Execute()
}

// SetVersionInfo 设置版本信息
func SetVersionInfo(ver, commit, buildT string) {
	version = ver
	gitCommit = commit
	buildTime = buildT
	rootCmd.Version = getVersionString()
}

// getVersionString 获取版本字符串
func getVersionString() string {
	if version == "" {
		version = "dev"
	}
	if gitCommit == "" {
		gitCommit = "unknown"
	}
	if buildTime == "" {
		buildTime = "unknown"
	}
	return fmt.Sprintf("%s (commit: %s, built: %s)", version, gitCommit, buildTime)
}

func init() {
	cobra.OnInitialize(initConfig)

	// 全局标志
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"配置文件路径 (默认搜索: $HOME/.go-transfer.yaml)")

	// 绑定环境变量
	viper.SetEnvPrefix("GOTRANSFER")
	viper.AutomaticEnv()
}

// initConfig 初始化配置
func initConfig() {
	if cfgFile != "" {
		// 使用指定的配置文件
		viper.SetConfigFile(cfgFile)
	} else {
		// 查找主目录
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// 搜索配置文件
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.AddConfigPath("./configs")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".go-transfer")
	}

	// 读取配置文件
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "使用配置文件:", viper.ConfigFileUsed())
	}
}
