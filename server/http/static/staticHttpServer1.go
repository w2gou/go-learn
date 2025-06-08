package static

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// 定义一个变量来存储 404 页面的内容
var customNotFoundPageContent []byte

type HttpServerModel1 struct{}

func (m *HttpServerModel1) StartServer() error {
	log.Printf("启动静态http服务器")

	// 启动 HTTP 服务等逻辑
	startServer()

	return nil
}

func startServer() {
	// 1. 定义命令行参数
	dirPtr := flag.String("dir", "resource/server/static/test1", "The root directory to serve files from.")
	portPtr := flag.Int("port", 8080, "The port to listen on.")

	// 2. 关键：在访问任何 flag 值之前，必须调用 flag.Parse()
	flag.Parse()

	// 3. 解引用指针获取实际的值
	rootDir := *dirPtr
	port := *portPtr
	notFoundPagePath := rootDir + "/404.html"

	// 读取 404 页面内容
	log.Printf("Attempting to load custom 404 page from: %s", notFoundPagePath)
	notFoundPageContent, err := os.ReadFile(notFoundPagePath)
	if err != nil {
		log.Fatalf("Error reading custom 404 page from '%s': %v. Please ensure the file exists.", notFoundPagePath, err)
	}
	customNotFoundPageContent = notFoundPageContent
	log.Println("Custom 404 page loaded successfully.")

	// 注册文件服务器处理器
	http.Handle("/file/", loggingHandler(http.StripPrefix("/file/", loggingFileServerHandler(http.Dir(rootDir)))))

	// 构建监听地址
	addr := fmt.Sprintf(":%d", port)

	// 调试日志：确认监听的地址和目录
	log.Printf("Starting custom file server on %s", addr)

	//addr := fmt.Sprintf("localhost:%d", port)
	err = http.ListenAndServe(addr, nil)
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

// Header 方法返回包装器自己的 Header map
// FileServer 会通过这个方法设置其 Content-Type 等头
func (rw *responseWriterWrapper) Header() http.Header {
	return rw.headers // 返回我们自己的 Header map
}

// WriteHeader 捕获状态码，并标记头已写入。
// 它不会立即写入到底层的 ResponseWriter。
func (rw *responseWriterWrapper) WriteHeader(code int) {
	if rw.headerWritten {
		return // 防止多次调用 WriteHeader
	}
	rw.statusCode = code
	rw.headerWritten = true
	// 实际的底层 WriteHeader 将在我们决定刷新响应时才调用。
}

// Write 将响应体写入到我们的内部缓冲区。
// 如果 WriteHeader 尚未调用，它会隐式地将状态设置为 200 OK。
func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	if !rw.headerWritten {
		// 如果在 WriteHeader 之前调用 Write，隐式地将状态设置为 200 OK。
		rw.WriteHeader(http.StatusOK)
	}
	// 将数据写入内部缓冲区
	return rw.body.Write(b)
}

// responseWriterWrapper 包装 http.ResponseWriter 以捕获状态码
type responseWriterWrapper struct {
	http.ResponseWriter              // 嵌入原始的 ResponseWriter
	statusCode          int          // 捕获的状态码
	body                bytes.Buffer // 缓冲响应体
	headerWritten       bool         // 标记 WriteHeader 是否已被调用
	headers             http.Header  // 捕获原始 ResponseWriter 的响应头
}

// loggingFileServerHandler 是一个自定义的处理器，它包装了 http.FileServer
// 并添加了日志记录和自定义 404 错误处理。
func loggingFileServerHandler(root http.Dir) http.Handler {
	fileServer := http.FileServer(root)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Incoming request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

		// 检查路径是否包含 ..，防止路径遍历攻击
		if strings.Contains(r.URL.Path, "..") {
			log.Printf("Security alert: Path traversal attempt detected: %s", r.URL.Path)
			//http.Error(w, "Access denied: Invalid path", http.StatusBadRequest)
			fmt.Fprintf(w, string(customNotFoundPageContent), r.URL.Path)
			return
		}

		// 包装 ResponseWriter 以捕获状态码和响应体
		// 初始状态码设置为 200 OK，以防 FileServer 没有显式调用 WriteHeader
		wrappedWriter := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,     // 默认状态码
			headers:        make(http.Header), // 初始化 Header map
		}

		// 核心：让 FileServer 执行其任务
		fileServer.ServeHTTP(wrappedWriter, r)

		// FileServer.ServeHTTP 返回后，确保状态码已设置
		if !wrappedWriter.headerWritten {
			wrappedWriter.WriteHeader(http.StatusOK) // 如果 FileServer 没有调用 WriteHeader，则默认为 OK
		}

		// 检查 FileServer 返回的状态码
		if wrappedWriter.statusCode == http.StatusNotFound || wrappedWriter.statusCode == 0 {
			log.Printf("File not found: %s. Returning custom 404 page.", r.URL.Path)
			// 清除 FileServer 可能设置的任何头（例如 Content-Type: text/plain）
			// 并设置我们自己的 HTML Content-Type
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			// 使用 fmt.Sprintf 格式化 404 页面内容，插入请求路径
			fmt.Fprintf(w, string(customNotFoundPageContent), r.URL.Path)
		} else {
			// 如果不是 404，将 FileServer 捕获的头和体写回原始 ResponseWriter
			// 复制所有 FileServer 可能设置的头
			for k, v := range wrappedWriter.Header() {
				for _, vv := range v {
					w.Header().Add(k, vv)
				}
			}
			w.WriteHeader(wrappedWriter.statusCode) // 写入捕获的状态码
			// 将 FileServer 写入的缓冲区内容写入原始 ResponseWriter
			if wrappedWriter.body.Len() > 0 {
				_, err := w.Write(wrappedWriter.body.Bytes())
				if err != nil {
					log.Printf("Error writing captured body: %v", err)
				}
			}
			log.Printf("Request processed: %s %s, Status: %d", r.Method, r.URL.Path, wrappedWriter.statusCode)
		}
	})
}
