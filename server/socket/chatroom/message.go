package chatroom

import "github.com/gorilla/websocket"

// 客户端连接结构体
type Client struct {
	conn     *websocket.Conn
	username string
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Message struct {
	Username string `json:"username"`
	Content  string `json:"content"`
	To       string `json:"to,omitempty"`
}

type Response struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
}
