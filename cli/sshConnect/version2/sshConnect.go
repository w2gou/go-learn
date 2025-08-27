package version2

import (
	command2 "go-learn/cli/sshConnect/version2/command"
	"log"
	"os"
)

type GoSshConnectModel struct{}

func (m *GoSshConnectModel) StartServer() error {
	log.Printf("启动ssh连接器")

	// 启动 HTTP 服务等逻辑
	startServer()

	return nil
}

func startServer() {
	command()
}

func command() {
	os.Args = []string{"gossh", "list"}
	command2.Execute()
}
