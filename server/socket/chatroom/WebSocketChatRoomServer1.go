package chatroom

import "log"

type WebSocketChatRoomServer1Model struct{}

func (m *WebSocketChatRoomServer1Model) StartServer() error {
	log.Printf("启动静态http服务器")

	// 启动 HTTP 服务等逻辑
	startServer()

	return nil
}

func startServer() {

}
