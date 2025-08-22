package version2

import (
	"log"
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
}
