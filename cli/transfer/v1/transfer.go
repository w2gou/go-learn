package transferV1

import (
	command2 "go-learn/cli/sshConnect/version2/command"
	"log"
)

type GoTransferModel struct{}

func (m *GoTransferModel) StartServer() error {
	log.Printf("启动文件分享工具")

	// 启动 HTTP 服务等逻辑
	startServer()

	return nil
}

func startServer() {
	command()
}

func command() {
	//os.Args = []string{"gossh", "list"}
	command2.Execute()
}
