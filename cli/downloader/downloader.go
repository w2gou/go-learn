package downloader

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type GoDownLoaderModel struct{}

func (m *GoDownLoaderModel) StartServer() error {
	log.Printf("启动下载器")

	// 启动 HTTP 服务等逻辑
	startServer()

	return nil
}

func startServer() {
	singleThreadDownload("")
}

func singleThreadDownload(url string) {
	url = "https://mirrors.ustc.edu.cn/ubuntu-releases/25.04/ubuntu-25.04-desktop-amd64.iso"

	log.Println("文件开始下载")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("创建请求失败: %v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36 Edg/138.0.0.0")
	req.Header.Set("Cookie", "addr=218.88.54.97")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("请求失败:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("请求失败，状态码: %d", resp.StatusCode)
	}

	//outFile, err := os.Create("F:\\downloaded_file.iso")
	//if err != nil {
	//	log.Fatalf("无法创建文件: %v", err)
	//}
	//defer outFile.Close()

	// 获取 Content-Length
	total := resp.ContentLength
	if total <= 0 {
		log.Fatalf("无法获取文件大小")
	}
	fmt.Printf("文件大小：%.2f MB\n", float64(total)/1024/1024)

	writer := &ProgressWriter{
		Total: total,
		Start: time.Now(),
	}

	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		log.Fatalf("写入文件失败: %v", err)
	}

	log.Println("文件下载完成")

	elapsed := time.Since(writer.Start).Seconds()
	avgSpeed := float64(writer.Written) / 1024 / 1024 / elapsed

	fmt.Printf("\n测速完成，总耗时：%.2f 秒，平均速度：%.2f MB/s\n", elapsed, avgSpeed)

	//err = os.Remove("F:\\downloaded_file.iso")
	//if err != nil {
	//	log.Fatalf("删除文件失败: %v", err)
	//}
	//log.Println("文件删除完成")
}

type ProgressWriter struct {
	Total       int64
	Written     int64
	LastPrinted time.Time
	Start       time.Time
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.Written += int64(n)

	now := time.Now()
	if now.Sub(pw.LastPrinted) >= 1*time.Second {
		percent := float64(pw.Written) / float64(pw.Total) * 100
		elapsed := now.Sub(pw.Start).Seconds()
		speed := float64(pw.Written) / 1024 / 1024 / elapsed // MB/s
		fmt.Printf("下载进度：%.1f%% | 速度：%.2f MB/s\r", percent, speed)
		pw.LastPrinted = now
	}

	return n, nil
}
