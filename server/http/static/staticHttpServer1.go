package static

import (
	"flag"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func StartServer() {
	dirPtr := flag.String("dir", "resource/server/static/test1", "The root directory to serve files from.")
	portPtr := flag.Int("port", 8080, "The port to listen on.")

	flag.Parse()

	http.Handle("/file/", loggingHandler(http.StripPrefix("/file/", http.FileServer(http.Dir(*dirPtr)))))

	// 启动 HTTP 服务
	log.Printf("Starting server on :%d\n", *portPtr)

	addr := fmt.Sprintf("localhost:%d", *portPtr)
	err := http.ListenAndServe(addr, nil)
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

func loggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request path: %s", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func fileHandler(baseDir, urlPrefix string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 去掉 /file/ 前缀，获取相对路径
		relPath := r.URL.Path[len(urlPrefix):]
		fullPath := filepath.Join(baseDir, relPath)

		// 打印访问日志
		log.Printf("[%s] %s %s from %s", time.Now().Format(time.RFC3339), r.Method, r.URL.Path, r.RemoteAddr)

		// 判断文件是否存在
		if info, err := os.Stat(fullPath); err != nil || info.IsDir() {
			// 自定义 404 响应
			log.Printf("404 Not Found: %s", fullPath)
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `<html><body><h1>404 - 文件未找到</h1></body></html>`)
			return
		}

		// 如果存在，交给 http.FileServer 处理
		fs := http.FileServer(http.Dir(baseDir))
		http.StripPrefix(urlPrefix, fs).ServeHTTP(w, r)
	})
}

// responseWriterWrapper 包装 http.ResponseWriter 以捕获状态码
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

// loggingFileServerHandler 是一个自定义的处理器，它包装了 http.FileServer
// 并添加了日志记录和自定义 404 错误处理。
func loggingFileServerHandler(root http.Dir, prefix string) http.Handler {
	fileServer := http.FileServer(root)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Incoming request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

		// 检查路径是否包含 ..，防止路径遍历攻击
		if strings.Contains(r.URL.Path, "..") {
			log.Printf("Security alert: Path traversal attempt detected: %s", r.URL.Path)
			http.Error(w, "Access denied: Invalid path", http.StatusBadRequest)
			return
		}

		// 包装 ResponseWriter 以捕获状态码
		wrappedWriter := &responseWriterWrapper{ResponseWriter: w}

		// 核心：让 FileServer 执行其任务
		fileServer.ServeHTTP(wrappedWriter, r)

		// 检查 FileServer 返回的状态码
		if wrappedWriter.statusCode == http.StatusNotFound {
			log.Printf("File not found: %s. Returning custom 404 page.", r.URL.Path)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			// 使用 fmt.Sprintf 格式化 404 页面内容，插入请求路径
			fmt.Fprintf(w, string(customNotFoundPageContent), r.URL.Path)
		} else {
			log.Printf("Request processed: %s %s, Status: %d", r.Method, r.URL.Path, wrappedWriter.statusCode)
		}
	})
}
