package transferV1Server

//import (
//	"context"
//	"crypto/subtle"
//	"encoding/json"
//	"fmt"
//	"net/http"
//	"os"
//	"path/filepath"
//	"runtime"
//	"strconv"
//	"strings"
//	"sync"
//	"sync/atomic"
//	"time"
//)
//
//// Server 高性能文件服务器
//type Server struct {
//	config     *config.Config
//	httpServer *http.Server
//	logger     *zap.Logger
//
//	// Goroutine池 - 新的并发控制架构
//	workerPool *WorkerPool
//	poolConfig *PoolConfig
//
//	// 传统并发控制 (已废弃，保留用于兼容)
//	semaphore     chan struct{} // 限制并发连接数
//	activeConns   int64         // 当前活跃连接数
//	totalRequests int64         // 总请求数
//	totalBytes    int64         // 总传输字节数
//
//	// 统计信息
//	stats     *ServerStats
//	statsLock sync.RWMutex
//
//	// 关闭控制
//	shutdown chan struct{}
//	done     chan struct{}
//}
//
//// ServerStats 服务器统计信息
//type ServerStats struct {
//	StartTime      time.Time        `json:"start_time"`
//	ActiveConns    int64            `json:"active_connections"`
//	TotalRequests  int64            `json:"total_requests"`
//	TotalBytes     int64            `json:"total_bytes"`
//	RequestsPerSec float64          `json:"requests_per_sec"`
//	BytesPerSec    float64          `json:"bytes_per_sec"`
//	TopFiles       map[string]int64 `json:"top_files"`
//	ClientIPs      map[string]int64 `json:"client_ips"`
//	Goroutines     int              `json:"goroutines"`
//	MemoryUsage    runtime.MemStats `json:"memory_usage"`
//
//	// Goroutine池统计信息
//	PoolStats *PoolStats `json:"pool_stats,omitempty"`
//}
//
//// New 创建新的服务器实例
//func New(cfg *config.Config) (*Server, error) {
//	// 创建日志记录器
//	logger, err := createLogger(cfg)
//	if err != nil {
//		return nil, fmt.Errorf("创建日志记录器失败: %v", err)
//	}
//
//	// 配置Goroutine池参数
//	poolConfig := &PoolConfig{
//		PoolSize:  cfg.Server.PoolSize,
//		QueueSize: cfg.Server.QueueSize,
//		SemSize:   cfg.Server.SemaphoreSize,
//	}
//
//	// 确保池配置合理（保险检查）
//	if poolConfig.PoolSize < 1 {
//		poolConfig.PoolSize = 4
//	}
//	if poolConfig.QueueSize < poolConfig.PoolSize {
//		poolConfig.QueueSize = poolConfig.PoolSize * 10
//	}
//	if poolConfig.SemSize < poolConfig.PoolSize {
//		poolConfig.SemSize = poolConfig.PoolSize * 2
//	}
//
//	// 创建Goroutine池
//	workerPool := NewWorkerPool(poolConfig, logger)
//
//	// 创建服务器实例
//	srv := &Server{
//		config:     cfg,
//		logger:     logger,
//		workerPool: workerPool,
//		poolConfig: poolConfig,
//		semaphore:  make(chan struct{}, cfg.Server.MaxConnections), // 保留用于兼容
//		stats: &ServerStats{
//			StartTime: time.Now(),
//			TopFiles:  make(map[string]int64),
//			ClientIPs: make(map[string]int64),
//		},
//		shutdown: make(chan struct{}),
//		done:     make(chan struct{}),
//	}
//
//	// 设置HTTP路由
//	if err := srv.setupRoutes(); err != nil {
//		return nil, fmt.Errorf("设置路由失败: %v", err)
//	}
//
//	// 启动统计协程
//	go srv.statsWorker()
//
//	logger.Info("服务器创建完成",
//		zap.Int("pool_size", poolConfig.PoolSize),
//		zap.Int("queue_size", poolConfig.QueueSize),
//		zap.Int("semaphore_size", poolConfig.SemSize))
//
//	return srv, nil
//}
//
//// setupRoutes 设置HTTP路由
//func (s *Server) setupRoutes() error {
//	r := mux.NewRouter()
//
//	// 添加中间件
//	r.Use(s.loggingMiddleware)
//
//	// 根据配置选择并发控制方式
//	if s.config.Server.UseWorkerPool {
//		r.Use(s.workerPoolMiddleware) // 使用Goroutine池中间件
//	} else {
//		r.Use(s.concurrencyMiddleware) // 使用传统并发控制
//	}
//
//	r.Use(s.authMiddleware)
//	r.Use(s.statsMiddleware)
//
//	// API路由
//	api := r.PathPrefix("/api").Subrouter()
//	api.HandleFunc("/stats", s.handleStats).Methods("GET")
//	api.HandleFunc("/health", s.handleHealth).Methods("GET")
//
//	// 检查分享路径类型
//	info, err := os.Stat(s.config.SharePath)
//	if err != nil {
//		return fmt.Errorf("无法访问分享路径: %v", err)
//	}
//
//	if info.IsDir() {
//		// 文件夹服务器
//		fs := &fileServer{
//			root:   s.config.SharePath,
//			server: s,
//		}
//		r.PathPrefix("/").Handler(fs)
//	} else {
//		// 单文件服务器
//		r.HandleFunc("/", s.handleSingleFile).Methods("GET")
//		r.HandleFunc("/download", s.handleSingleFile).Methods("GET")
//	}
//
//	// 创建HTTP服务器
//	s.httpServer = &http.Server{
//		Addr:         ":" + s.config.Server.Port,
//		Handler:      r,
//		ReadTimeout:  s.config.Server.Timeout,
//		WriteTimeout: s.config.Server.Timeout,
//		IdleTimeout:  60 * time.Second,
//	}
//
//	return nil
//}
//
//// workerPoolMiddleware Goroutine池中间件 - 新的并发控制架构
//func (s *Server) workerPoolMiddleware(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		// 使用WorkerPool提交任务
//		if !s.workerPool.SubmitJob(w, r, next) {
//			// 任务提交失败（队列满）
//			s.logger.Warn("任务队列已满，拒绝请求",
//				zap.String("path", r.URL.Path),
//				zap.String("client_ip", getClientIP(r)))
//			http.Error(w, "服务器繁忙，任务队列已满，请稍后重试", http.StatusServiceUnavailable)
//		}
//	})
//}
//
//// concurrencyMiddleware 传统并发控制中间件（已废弃，保留用于兼容）
//func (s *Server) concurrencyMiddleware(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		// 获取信号量
//		select {
//		case s.semaphore <- struct{}{}:
//			atomic.AddInt64(&s.activeConns, 1)
//			defer func() {
//				<-s.semaphore
//				atomic.AddInt64(&s.activeConns, -1)
//			}()
//			next.ServeHTTP(w, r)
//		default:
//			// 服务器繁忙
//			http.Error(w, "服务器繁忙，请稍后重试", http.StatusServiceUnavailable)
//		}
//	})
//}
//
//// authMiddleware 认证中间件
//func (s *Server) authMiddleware(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		if !s.config.Server.Auth {
//			next.ServeHTTP(w, r)
//			return
//		}
//
//		username, password, ok := r.BasicAuth()
//		if !ok {
//			w.Header().Set("WWW-Authenticate", `Basic realm="Go-Transfer"`)
//			http.Error(w, "需要认证", http.StatusUnauthorized)
//			return
//		}
//
//		// 使用恒定时间比较防止时序攻击
//		usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(s.config.Server.Username)) == 1
//		passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(s.config.Server.Password)) == 1
//
//		if !usernameMatch || !passwordMatch {
//			w.Header().Set("WWW-Authenticate", `Basic realm="Go-Transfer"`)
//			http.Error(w, "认证失败", http.StatusUnauthorized)
//			return
//		}
//
//		next.ServeHTTP(w, r)
//	})
//}
//
//// statsMiddleware 统计中间件
//func (s *Server) statsMiddleware(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		start := time.Now()
//
//		// 记录请求
//		atomic.AddInt64(&s.totalRequests, 1)
//
//		// 获取客户端IP
//		clientIP := getClientIP(r)
//
//		// 包装ResponseWriter以捕获字节数
//		wrappedWriter := &responseWriter{
//			ResponseWriter: w,
//			statusCode:     200,
//		}
//
//		next.ServeHTTP(wrappedWriter, r)
//
//		// 更新统计信息
//		duration := time.Since(start)
//		bytes := int64(wrappedWriter.bytesWritten)
//		atomic.AddInt64(&s.totalBytes, bytes)
//
//		// 更新详细统计
//		s.updateDetailedStats(r.URL.Path, clientIP, bytes)
//
//		// 记录请求日志
//		s.logger.Info("HTTP请求",
//			zap.String("method", r.Method),
//			zap.String("path", r.URL.Path),
//			zap.String("client_ip", clientIP),
//			zap.Int("status", wrappedWriter.statusCode),
//			zap.Int64("bytes", bytes),
//			zap.Duration("duration", duration),
//		)
//	})
//}
//
//// loggingMiddleware 日志中间件
//func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		if s.config.Logging.Verbose {
//			s.logger.Debug("处理请求",
//				zap.String("method", r.Method),
//				zap.String("url", r.URL.String()),
//				zap.String("user_agent", r.UserAgent()),
//				zap.String("remote_addr", r.RemoteAddr),
//			)
//		}
//		next.ServeHTTP(w, r)
//	})
//}
//
//// handleSingleFile 处理单文件下载
//func (s *Server) handleSingleFile(w http.ResponseWriter, r *http.Request) {
//	info, err := os.Stat(s.config.SharePath)
//	if err != nil {
//		http.Error(w, "文件不存在", http.StatusNotFound)
//		return
//	}
//
//	// 设置响应头
//	filename := filepath.Base(s.config.SharePath)
//	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
//	w.Header().Set("Content-Type", "application/octet-stream")
//	w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
//
//	// 支持断点续传
//	http.ServeFile(w, r, s.config.SharePath)
//}
//
//// handleStats 处理统计信息请求
//func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
//	s.statsLock.RLock()
//	stats := *s.stats
//	s.statsLock.RUnlock()
//
//	// 更新实时数据
//	stats.ActiveConns = atomic.LoadInt64(&s.activeConns)
//	stats.TotalRequests = atomic.LoadInt64(&s.totalRequests)
//	stats.TotalBytes = atomic.LoadInt64(&s.totalBytes)
//	stats.Goroutines = runtime.NumGoroutine()
//
//	// 计算速率
//	elapsed := time.Since(stats.StartTime).Seconds()
//	if elapsed > 0 {
//		stats.RequestsPerSec = float64(stats.TotalRequests) / elapsed
//		stats.BytesPerSec = float64(stats.TotalBytes) / elapsed
//	}
//
//	// 获取内存统计
//	runtime.ReadMemStats(&stats.MemoryUsage)
//
//	// 获取WorkerPool统计信息
//	if s.workerPool != nil {
//		stats.PoolStats = s.workerPool.GetStats()
//	}
//
//	w.Header().Set("Content-Type", "application/json")
//	json.NewEncoder(w).Encode(stats)
//}
//
//// handleHealth 处理健康检查
//func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
//	poolHealthy := s.workerPool != nil && s.workerPool.IsHealthy()
//	status := "healthy"
//	httpStatus := http.StatusOK
//
//	if !poolHealthy {
//		status = "unhealthy"
//		httpStatus = http.StatusServiceUnavailable
//	}
//
//	health := map[string]interface{}{
//		"status":       status,
//		"timestamp":    time.Now(),
//		"uptime":       time.Since(s.stats.StartTime).String(),
//		"goroutines":   runtime.NumGoroutine(),
//		"pool_healthy": poolHealthy,
//	}
//
//	if s.workerPool != nil {
//		poolStats := s.workerPool.GetStats()
//		health["pool_stats"] = map[string]interface{}{
//			"queued_jobs":     poolStats.QueuedJobs,
//			"total_jobs":      poolStats.TotalJobs,
//			"rejected_jobs":   poolStats.RejectedJobs,
//			"semaphore_count": poolStats.SemaphoreCount,
//		}
//	}
//
//	w.Header().Set("Content-Type", "application/json")
//	w.WriteHeader(httpStatus)
//	json.NewEncoder(w).Encode(health)
//}
//
//// Start 启动服务器
//func (s *Server) Start() error {
//	// 如果启用了WorkerPool，则启动它
//	if s.config.Server.UseWorkerPool && s.workerPool != nil {
//		if err := s.workerPool.Start(); err != nil {
//			return fmt.Errorf("启动WorkerPool失败: %v", err)
//		}
//	}
//
//	logFields := []zap.Field{
//		zap.String("address", s.httpServer.Addr),
//		zap.String("share_path", s.config.SharePath),
//		zap.Int("max_connections", s.config.Server.MaxConnections),
//		zap.Bool("use_worker_pool", s.config.Server.UseWorkerPool),
//	}
//
//	if s.config.Server.UseWorkerPool {
//		logFields = append(logFields,
//			zap.Int("worker_pool_size", s.poolConfig.PoolSize),
//			zap.Int("job_queue_size", s.poolConfig.QueueSize),
//			zap.Int("semaphore_size", s.poolConfig.SemSize),
//		)
//	}
//
//	s.logger.Info("启动HTTP服务器", logFields...)
//
//	if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
//		return fmt.Errorf("服务器启动失败: %v", err)
//	}
//	return nil
//}
//
//// Shutdown 优雅关闭服务器
//func (s *Server) Shutdown(ctx context.Context) error {
//	s.logger.Info("开始关闭服务器")
//
//	close(s.shutdown)
//
//	// 先关闭HTTP服务器
//	if err := s.httpServer.Shutdown(ctx); err != nil {
//		s.logger.Error("HTTP服务器关闭失败", zap.Error(err))
//	}
//
//	// 关闭WorkerPool
//	if s.workerPool != nil {
//		s.logger.Info("关闭WorkerPool...")
//		s.workerPool.Stop()
//	}
//
//	// 等待统计协程结束
//	<-s.done
//
//	s.logger.Info("服务器已完全关闭")
//	return nil
//}
//
//// statsWorker 统计信息工作协程
//func (s *Server) statsWorker() {
//	ticker := time.NewTicker(5 * time.Second)
//	defer ticker.Stop()
//	defer close(s.done)
//
//	for {
//		select {
//		case <-ticker.C:
//			// 定期打印统计信息
//			if s.config.Logging.Verbose {
//				activeConns := atomic.LoadInt64(&s.activeConns)
//				totalRequests := atomic.LoadInt64(&s.totalRequests)
//				totalBytes := atomic.LoadInt64(&s.totalBytes)
//
//				logFields := []zap.Field{
//					zap.Int64("active_connections", activeConns),
//					zap.Int64("total_requests", totalRequests),
//					zap.String("total_bytes", formatBytes(totalBytes)),
//					zap.Int("goroutines", runtime.NumGoroutine()),
//				}
//
//				// 添加WorkerPool统计信息
//				if s.workerPool != nil {
//					poolStats := s.workerPool.GetStats()
//					logFields = append(logFields,
//						zap.Int64("queued_jobs", poolStats.QueuedJobs),
//						zap.Int64("pool_total_jobs", poolStats.TotalJobs),
//						zap.Int64("rejected_jobs", poolStats.RejectedJobs),
//						zap.Int32("semaphore_count", poolStats.SemaphoreCount),
//						zap.Bool("pool_healthy", s.workerPool.IsHealthy()),
//					)
//				}
//
//				s.logger.Info("服务器统计", logFields...)
//			}
//		case <-s.shutdown:
//			return
//		}
//	}
//}
//
//// updateDetailedStats 更新详细统计信息
//func (s *Server) updateDetailedStats(path, clientIP string, bytes int64) {
//	s.statsLock.Lock()
//	defer s.statsLock.Unlock()
//
//	// 更新文件访问统计
//	s.stats.TopFiles[path]++
//
//	// 更新客户端IP统计
//	s.stats.ClientIPs[clientIP]++
//}
//
//// responseWriter 包装ResponseWriter以捕获写入的字节数
//type responseWriter struct {
//	http.ResponseWriter
//	statusCode   int
//	bytesWritten int
//}
//
//func (rw *responseWriter) WriteHeader(code int) {
//	rw.statusCode = code
//	rw.ResponseWriter.WriteHeader(code)
//}
//
//func (rw *responseWriter) Write(b []byte) (int, error) {
//	n, err := rw.ResponseWriter.Write(b)
//	rw.bytesWritten += n
//	return n, err
//}
//
//// fileServer 自定义文件服务器
//type fileServer struct {
//	root   string
//	server *Server
//}
//
//func (fs *fileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//	// 清理路径防止目录遍历攻击
//	path := filepath.Clean(r.URL.Path)
//	if strings.Contains(path, "..") {
//		http.Error(w, "禁止访问", http.StatusForbidden)
//		return
//	}
//
//	fullPath := filepath.Join(fs.root, path)
//
//	// 检查文件是否存在
//	info, err := os.Stat(fullPath)
//	if os.IsNotExist(err) {
//		http.Error(w, "文件不存在", http.StatusNotFound)
//		return
//	}
//	if err != nil {
//		http.Error(w, "服务器错误", http.StatusInternalServerError)
//		return
//	}
//
//	// 如果是目录，显示文件列表
//	if info.IsDir() {
//		fs.serveDirectory(w, r, fullPath)
//		return
//	}
//
//	// 服务文件
//	http.ServeFile(w, r, fullPath)
//}
//
//func (fs *fileServer) serveDirectory(w http.ResponseWriter, r *http.Request, dirPath string) {
//	// 读取目录内容
//	entries, err := os.ReadDir(dirPath)
//	if err != nil {
//		http.Error(w, "无法读取目录", http.StatusInternalServerError)
//		return
//	}
//
//	// 生成HTML页面
//	html := generateDirectoryHTML(r.URL.Path, entries)
//
//	w.Header().Set("Content-Type", "text/html; charset=utf-8")
//	w.Write([]byte(html))
//}
//
//// createLogger 创建日志记录器
//func createLogger(cfg *config.Config) (*zap.Logger, error) {
//	var zapConfig zap.Config
//
//	if cfg.Logging.Format == "json" {
//		zapConfig = zap.NewProductionConfig()
//	} else {
//		zapConfig = zap.NewDevelopmentConfig()
//		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
//	}
//
//	if cfg.Logging.Verbose {
//		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
//	} else {
//		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
//	}
//
//	return zapConfig.Build()
//}
//
//// getClientIP 获取客户端IP地址
//func getClientIP(r *http.Request) string {
//	// 检查 X-Forwarded-For 头
//	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
//		ips := strings.Split(xff, ",")
//		return strings.TrimSpace(ips[0])
//	}
//
//	// 检查 X-Real-IP 头
//	if xri := r.Header.Get("X-Real-IP"); xri != "" {
//		return xri
//	}
//
//	// 从 RemoteAddr 提取IP
//	ip := r.RemoteAddr
//	if colon := strings.LastIndex(ip, ":"); colon != -1 {
//		ip = ip[:colon]
//	}
//	return ip
//}
//
//// formatBytes 格式化字节数
//func formatBytes(bytes int64) string {
//	const unit = 1024
//	if bytes < unit {
//		return fmt.Sprintf("%d B", bytes)
//	}
//	div, exp := int64(unit), 0
//	for n := bytes / unit; n >= unit; n /= unit {
//		div *= unit
//		exp++
//	}
//	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
//}
//
//// generateDirectoryHTML 生成目录列表HTML
//func generateDirectoryHTML(path string, entries []os.DirEntry) string {
//	var html strings.Builder
//
//	html.WriteString(`<!DOCTYPE html>
//<html lang="zh-CN">
//<head>
//    <meta charset="utf-8">
//    <meta name="viewport" content="width=device-width, initial-scale=1">
//    <title>Go-Transfer - 文件列表</title>
//    <style>
//        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
//        .container { max-width: 800px; margin: 0 auto; background: white; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
//        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 20px; border-radius: 8px 8px 0 0; }
//        .header h1 { margin: 0; font-size: 24px; }
//        .header p { margin: 5px 0 0 0; opacity: 0.9; }
//        .file-list { padding: 0; }
//        .file-item { display: flex; align-items: center; padding: 12px 20px; border-bottom: 1px solid #eee; text-decoration: none; color: #333; }
//        .file-item:hover { background: #f8f9fa; }
//        .file-icon { width: 20px; height: 20px; margin-right: 12px; }
//        .file-info { flex: 1; }
//        .file-name { font-weight: 500; }
//        .file-size { font-size: 12px; color: #666; margin-top: 2px; }
//        .footer { padding: 20px; text-align: center; color: #666; font-size: 12px; border-top: 1px solid #eee; }
//    </style>
//</head>
//<body>
//    <div class="container">
//        <div class="header">
//            <h1>📁 ` + path + `</h1>
//            <p>Go-Transfer 文件服务器</p>
//        </div>
//        <div class="file-list">`)
//
//	// 添加返回上级目录链接
//	if path != "/" {
//		parentPath := filepath.Dir(path)
//		if parentPath == "." {
//			parentPath = "/"
//		}
//		html.WriteString(fmt.Sprintf(`
//            <a href="%s" class="file-item">
//                <div class="file-icon">📁</div>
//                <div class="file-info">
//                    <div class="file-name">..</div>
//                    <div class="file-size">返回上级目录</div>
//                </div>
//            </a>`, parentPath))
//	}
//
//	// 添加文件和目录
//	for _, entry := range entries {
//		info, err := entry.Info()
//		if err != nil {
//			continue
//		}
//
//		icon := "📄"
//		sizeStr := formatBytes(info.Size())
//		if entry.IsDir() {
//			icon = "📁"
//			sizeStr = "文件夹"
//		}
//
//		entryPath := filepath.Join(path, entry.Name())
//		if !strings.HasPrefix(entryPath, "/") {
//			entryPath = "/" + entryPath
//		}
//
//		html.WriteString(fmt.Sprintf(`
//            <a href="%s" class="file-item">
//                <div class="file-icon">%s</div>
//                <div class="file-info">
//                    <div class="file-name">%s</div>
//                    <div class="file-size">%s | %s</div>
//                </div>
//            </a>`, entryPath, icon, entry.Name(), sizeStr, info.ModTime().Format("2006-01-02 15:04:05")))
//	}
//
//	html.WriteString(`
//        </div>
//        <div class="footer">
//            Powered by Go-Transfer | 高性能文件分享工具
//        </div>
//    </div>
//</body>
//</html>`)
//
//	return html.String()
//}
