package chatroom

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"sync"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // 允许跨域
}

// 所有客户端列表
var clients = make(map[*websocket.Conn]string)

// 全局广播通道
var broadcast = make(chan Message)

var mu sync.Mutex

var tokenMap = make(map[string]string) // token -> username

var usersConn = make(map[string]*websocket.Conn) // username -> conn

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
	http.HandleFunc("/logout", handleLogout)
	http.HandleFunc("/chat", handleWebSocket)

	go handleBroadcast()

	fmt.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}

// 处理 WebSocket 请求
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	username := tokenMap[token]
	if username == "" {
		http.Error(w, "未认证", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("升级失败:", err)
		return
	}
	defer conn.Close()

	mu.Lock()
	clients[conn] = username
	usersConn[username] = conn
	sendOnlineUsers() // 广播在线用户列表
	mu.Unlock()

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("读取失败:", err)
			break
		}
		msg.Username = username
		broadcast <- msg
	}

	mu.Lock()
	delete(clients, conn)
	delete(usersConn, username)
	sendOnlineUsers()
	mu.Unlock()
}

func handleBroadcast() {
	for {
		msg := <-broadcast
		mu.Lock()
		if msg.To != "" {
			// 私聊
			if toConn, ok := usersConn[msg.To]; ok {
				_ = toConn.WriteJSON(msg)
			}
			if selfConn, ok := usersConn[msg.Username]; ok {
				_ = selfConn.WriteJSON(msg) // 回显给自己
			}
		} else {
			// 群聊
			for conn := range clients {
				_ = conn.WriteJSON(msg)
			}
		}
		mu.Unlock()
	}
}

func sendOnlineUsers() {
	usernames := make([]string, 0)
	for name := range usersConn {
		usernames = append(usernames, name)
	}
	data, _ := json.Marshal(struct {
		Type  string   `json:"type"`
		Users []string `json:"users"`
	}{
		Type:  "online-users",
		Users: usernames,
	})
	for conn := range clients {
		conn.WriteMessage(websocket.TextMessage, data)
	}
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "无效参数", http.StatusBadRequest)
		return
	}

	// 生成 UUID
	user.ID = uuid.New().String()

	// 加密密码
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	err = insertUser(user.ID, user.Username, string(hash))
	if err != nil {
		http.Error(w, "用户名已存在", http.StatusBadRequest)
		return
	}

	token := uuid.New().String()
	tokenMap[token] = user.Username
	json.NewEncoder(w).Encode(Response{Success: true, Token: token})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "无效参数", http.StatusBadRequest)
		return
	}

	selectedUser, err := selectUser(user.Username)
	if err != nil {
		json.NewEncoder(w).Encode(Response{Success: false})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(selectedUser.Password), []byte(user.Password)) != nil {
		json.NewEncoder(w).Encode(Response{Success: false})
		return
	}

	token := uuid.New().String()
	tokenMap[token] = user.Username
	json.NewEncoder(w).Encode(Response{Success: true, Token: token})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	mu.Lock()
	delete(clients, usersConn[tokenMap[token]])
	delete(usersConn, tokenMap[token])
	delete(tokenMap, token)
	sendOnlineUsers()
	w.Write([]byte("logout success"))
}
