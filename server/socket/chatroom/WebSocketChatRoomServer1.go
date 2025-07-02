package chatroom

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // 允许跨域
}

// 所有客户端列表
var clients = make(map[*Client]bool)

// 全局广播通道
var broadcast = make(chan Message)

var mu sync.Mutex

type WebSocketChatRoomServer1Model struct{}

func (m *WebSocketChatRoomServer1Model) StartServer() error {
	log.Printf("启动websocket聊天服务器")

	// 启动 HTTP 服务等逻辑
	startServer()

	return nil
}

func startServer() {
	initDB()

	fs := http.FileServer(http.Dir("resource/server/socket/chatRoom"))
	http.Handle("/", fs)

	http.HandleFunc("/register", handleRegister)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/chat", handleWebSocket)

	go handleBroadcast()

	fmt.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}

// 处理 WebSocket 请求
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Missing username", http.StatusBadRequest)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WS upgrade error:", err)
		return
	}

	client := &Client{conn: ws, username: username}
	mu.Lock()
	clients[client] = true
	mu.Unlock()

	defer func() {
		mu.Lock()
		delete(clients, client)
		mu.Unlock()
		ws.Close()
	}()

	for {
		var msg Message
		if err := ws.ReadJSON(&msg); err != nil {
			break
		}
		msg.Username = username
		broadcast <- msg
	}
}

func handleBroadcast() {
	for {
		msg := <-broadcast
		mu.Lock()
		for client := range clients {
			client.conn.WriteJSON(msg)
		}
		mu.Unlock()
	}
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "无效参数", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", user.Username, user.Password)
	if err != nil {
		http.Error(w, "用户名已存在", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "无效参数", http.StatusBadRequest)
		return
	}

	var count int
	row := db.QueryRow("SELECT COUNT(*) FROM users WHERE username=? AND password=?", user.Username, user.Password)
	row.Scan(&count)

	if count == 1 {
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "用户名或密码错误", http.StatusUnauthorized)
	}
}
