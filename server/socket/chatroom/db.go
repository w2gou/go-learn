package chatroom

import (
	"database/sql"
)

var db *sql.DB

func initDB() {
	var err error
	dsn := "root:admin123@tcp(127.0.0.1:3306)/go_socket_chat_room?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic("数据库连接失败: " + err.Error())
	}

	if err = db.Ping(); err != nil {
		panic("无法连接数据库: " + err.Error())
	}
}

func selectUser(username string) (User, error) {
	var user User
	err := db.QueryRow("SELECT id, password FROM users WHERE username = ?", username).Scan(&user.ID, &user.Password)
	user.Username = username
	return user, err
}

func insertUser(id string, username string, password string) error {
	_, err := db.Exec("INSERT INTO users (id, username, password) VALUES (?, ?, ?)",
		id, username, password)
	return err
}
