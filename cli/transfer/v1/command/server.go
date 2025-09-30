package transferV1Command

//import (
//	"context"
//	"fmt"
//	"github.com/spf13/cobra"
//	"github.com/spf13/viper"
//	"log"
//	"os"
//	"os/signal"
//	"path/filepath"
//	"syscall"
//	"time"
//)
//
//// serveCmd 表示 serve 命令
//var serveCmd = &cobra.Command{
//	Use:   "serve [path]",
//	Short: "启动文件服务器",
//	Long: `启动高性能的HTTP文件服务器，支持：
//
//• 并发处理 - 每个请求独立的 Goroutine
//• 自动发现 - 智能获取局域网IP地址
//• 二维码生成 - 终端显示访问二维码
//• 实时监控 - 连接数和传输状态监控
//• 优雅关闭 - 支持信号处理和优雅退出
//
//示例:
//  go-transfer serve ./myfile.txt          # 分享单个文件
//  go-transfer serve ./myfolder            # 分享文件夹
//  go-transfer serve --port 9000 ./files  # 自定义端口
//  go-transfer serve --auth               # 启用基础认证`,
//	Args: cobra.MaximumNArgs(1),
//	RunE: runServe,
//}
//
//func init() {
//	rootCmd.AddCommand(serveCmd)
//
//	// 服务器配置标志
//	serveCmd.Flags().StringP("port", "p", "8080", "服务器监听端口")
//	serveCmd.Flags().StringP("host", "H", "", "绑定主机地址 (默认: 自动检测)")
//	serveCmd.Flags().BoolP("auth", "a", false, "启用基础HTTP认证")
//	serveCmd.Flags().String("username", "admin", "认证用户名")
//	serveCmd.Flags().String("password", "", "认证密码 (为空则随机生成)")
//
//	// 并发控制标志
//	serveCmd.Flags().Int("max-connections", 100, "最大并发连接数")
//	serveCmd.Flags().Duration("timeout", 30*time.Second, "请求超时时间")
//	serveCmd.Flags().Int64("max-file-size", 1024*1024*1024, "最大文件大小 (字节)")
//
//	// 二维码配置标志
//	serveCmd.Flags().Bool("no-qr", false, "禁用二维码显示")
//	serveCmd.Flags().String("qr-size", "medium", "二维码大小 (small|medium|large)")
//
//	// 日志配置标志
//	serveCmd.Flags().Bool("verbose", false, "详细日志输出")
//	serveCmd.Flags().String("log-format", "text", "日志格式 (text|json)")
//
//	// 绑定配置
//	viper.BindPFlag("server.port", serveCmd.Flags().Lookup("port"))
//	viper.BindPFlag("server.host", serveCmd.Flags().Lookup("host"))
//	viper.BindPFlag("server.auth", serveCmd.Flags().Lookup("auth"))
//	viper.BindPFlag("server.username", serveCmd.Flags().Lookup("username"))
//	viper.BindPFlag("server.password", serveCmd.Flags().Lookup("password"))
//	viper.BindPFlag("server.max_connections", serveCmd.Flags().Lookup("max-connections"))
//	viper.BindPFlag("server.timeout", serveCmd.Flags().Lookup("timeout"))
//	viper.BindPFlag("server.max_file_size", serveCmd.Flags().Lookup("max-file-size"))
//	viper.BindPFlag("qr.disabled", serveCmd.Flags().Lookup("no-qr"))
//	viper.BindPFlag("qr.size", serveCmd.Flags().Lookup("qr-size"))
//	viper.BindPFlag("logging.verbose", serveCmd.Flags().Lookup("verbose"))
//	viper.BindPFlag("logging.format", serveCmd.Flags().Lookup("log-format"))
//}
//
//// runServe 执行服务命令
//func runServe(cmd *cobra.Command, args []string) error {
//	// 解析路径参数
//	var sharePath string
//	if len(args) > 0 {
//		sharePath = args[0]
//	} else {
//		// 如果没有提供路径，使用当前目录
//		var err error
//		sharePath, err = os.Getwd()
//		if err != nil {
//			return fmt.Errorf("无法获取当前目录: %v", err)
//		}
//	}
//
//	// 获取绝对路径
//	absPath, err := filepath.Abs(sharePath)
//	if err != nil {
//		return fmt.Errorf("无法获取绝对路径: %v", err)
//	}
//
//	// 检查路径是否存在
//	if _, err := os.Stat(absPath); os.IsNotExist(err) {
//		return fmt.Errorf("指定的路径不存在: %s", absPath)
//	}
//
//	// 创建配置
//	cfg, err := config.New(absPath)
//	if err != nil {
//		return fmt.Errorf("创建配置失败: %v", err)
//	}
//
//	// 获取本机IP地址
//	localIP, err := network.GetLocalIP()
//	if err != nil {
//		return fmt.Errorf("无法获取本机IP地址: %v", err)
//	}
//
//	// 如果没有指定host，使用检测到的IP
//	if cfg.Server.Host == "" {
//		cfg.Server.Host = localIP
//	}
//
//	// 构建服务器URL
//	serverURL := fmt.Sprintf("http://%s:%s", cfg.Server.Host, cfg.Server.Port)
//
//	// 显示二维码（如果启用）
//	if !cfg.QR.Disabled {
//		fmt.Println("正在生成二维码...")
//		if err := qr.Display(serverURL, cfg.QR.Size); err != nil {
//			log.Printf("二维码生成失败: %v", err)
//		}
//	}
//
//	// 创建并启动服务器
//	srv, err := server.New(cfg)
//	if err != nil {
//		return fmt.Errorf("创建服务器失败: %v", err)
//	}
//
//	// 启动服务器（非阻塞）
//	go func() {
//		fmt.Printf("\n🚀 文件服务器已启动\n")
//		fmt.Printf("📂 分享路径: %s\n", cfg.SharePath)
//		fmt.Printf("🌐 访问地址: %s\n", serverURL)
//		fmt.Printf("⚡ 最大并发: %d 连接\n", cfg.Server.MaxConnections)
//		if cfg.Server.Auth {
//			fmt.Printf("🔐 认证信息: %s / %s\n", cfg.Server.Username, cfg.Server.Password)
//		}
//		fmt.Printf("📊 使用 Ctrl+C 停止服务器\n\n")
//
//		if err := srv.Start(); err != nil {
//			log.Printf("服务器启动失败: %v", err)
//		}
//	}()
//
//	// 等待中断信号
//	return waitForShutdown(srv)
//}
//
//// waitForShutdown 等待停止信号并优雅关闭服务器
//func waitForShutdown(srv *server.Server) error {
//	quit := make(chan os.Signal, 1)
//	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
//
//	<-quit
//	fmt.Println("\n📊 正在关闭服务器...")
//
//	// 设置关闭超时
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//
//	// 优雅关闭服务器
//	if err := srv.Shutdown(ctx); err != nil {
//		return fmt.Errorf("服务器关闭失败: %v", err)
//	}
//
//	fmt.Println("✅ 服务器已安全关闭")
//	return nil
//}
