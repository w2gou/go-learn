package chatroom

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // 允许跨域
}

// 客户端连接结构体
type Client struct {
	conn *websocket.Conn
	send chan []byte
}

// 所有客户端列表
var clients = make(map[*Client]bool)

// 全局广播通道
var broadcast = make(chan []byte)

type WebSocketChatRoomServer1Model struct{}

func (m *WebSocketChatRoomServer1Model) StartServer() error {
	log.Printf("启动websocket聊天服务器")

	// 启动 HTTP 服务等逻辑
	startServer()

	return nil
}

func startServer() {
	fs := http.FileServer(http.Dir("resource/server/socket/chatRoom"))
	http.Handle("/", fs)

	http.HandleFunc("/chat", handleConnections)

	go handleBroadcast()

	fmt.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}

// 处理 WebSocket 请求
func handleConnections(w http.ResponseWriter, r *http.Request) {
	// 升级 HTTP 为 WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)
		return
	}
	defer ws.Close()

	client := &Client{conn: ws, send: make(chan []byte)}
	clients[client] = true

	go handleMessagesFromClient(client)

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			delete(clients, client)
			break
		}
		broadcast <- msg
	}
}

// 客户端消息处理
func handleMessagesFromClient(client *Client) {
	for msg := range client.send {
		err := client.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			client.conn.Close()
			delete(clients, client)
			break
		}
	}
}

// 消息广播器
func handleBroadcast() {
	for {
		msg := <-broadcast
		for client := range clients {
			select {
			case client.send <- msg:
			default:
				close(client.send)
				delete(clients, client)
			}
		}
	}
}
