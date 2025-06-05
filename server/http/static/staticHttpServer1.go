package static

import (
	"log"
	"mime"
	"net/http"
	"path/filepath"
)

func StartServer() {
	//http.HandleFunc("/count", counter)

	// 设置静态文件目录
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("public"))))

	// 启动 HTTP 服务
	log.Println("Starting server on :8080")

	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		log.Fatal("Error starting server: ", err)
	}
}

func getMimeType(file string) string {
	mimeType := mime.TypeByExtension(filepath.Ext(file))
	if mimeType == "" {
		mimeType = "application/octet-stream" // 默认类型
	}
	return mimeType
}

func counter(w http.ResponseWriter, r *http.Request) {
}
