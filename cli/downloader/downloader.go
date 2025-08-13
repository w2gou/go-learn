package downloader

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
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
	url := "https://mirrors.tuna.tsinghua.edu.cn/aosp-monthly/aosp-latest.tar"
	//url := "https://mirrors.ustc.edu.cn/aosp-monthly/aosp-latest.tar"
	//url := "https://mirrors.ustc.edu.cn/centos-cloud/centos/7/images/CentOS-7-x86_64-Azure-1907.vhd"
	//url := "https://mirrors.ustc.edu.cn/ubuntu-releases/25.04/ubuntu-25.04-desktop-amd64.iso"
	//singleThreadDownload(url)
	err := multiThreadDownload(url, "F:\\zdaobao\\aosp-latest.tar", 4)
	if err != nil {
		log.Fatalf("失败: %v", err)
		return
	}
}

func singleThreadDownload(url string) {
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

func multiThreadDownload(url string, outputPath string, threadCount int) error {
	// 获取文件大小
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("创建请求失败: %v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36 Edg/138.0.0.0")
	//req.Header.Set("Cookie", "addr=218.88.54.148")

	client := &http.Client{}
	resp, err := client.Do(req)
	//resp, err := http.Head(url)
	if err != nil {
		return fmt.Errorf("获取文件头失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	sizeStr := resp.Header.Get("Content-Length")
	if sizeStr == "" {
		return fmt.Errorf("服务器未返回文件大小，可能不支持 Range 请求")
	}

	totalSize, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return fmt.Errorf("无法解析文件大小: %v", err)
	}

	log.Printf("文件大小：%.2f MB\n", float64(totalSize)/1024/1024)

	// 创建目标文件
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer outFile.Close()

	// 预分配文件空间
	if err := outFile.Truncate(totalSize); err != nil {
		return fmt.Errorf("预分配文件空间失败: %v", err)
	}

	// 创建 WaitGroup 控制并发
	var wg sync.WaitGroup
	blockSize := totalSize / int64(threadCount)

	for i := 0; i < threadCount; i++ {
		start := int64(i) * blockSize
		end := start + blockSize - 1
		if i == threadCount-1 {
			end = totalSize - 1 // 最后一块确保完整
		}

		wg.Add(1)
		go func(start, end int64, idx int) {
			defer wg.Done()

			// 每个线程启动时等待 i 秒
			time.Sleep(time.Duration(i) * time.Second)

			log.Printf("[线程 %d] 开始下载", idx)

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				log.Printf("[线程 %d] 创建请求失败: %v", idx, err)
				return
			}
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36 Edg/138.0.0.0")
			//req.Header.Set("Cookie", "addr=218.88.54.142")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Printf("[线程 %d] 请求失败: %v", idx, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
				log.Printf("[线程 %d] 非预期响应: %d", idx, resp.StatusCode)
				return
			}

			// 定位写入
			buf := make([]byte, 32*1024)
			file, err := os.OpenFile(outputPath, os.O_WRONLY, 0644)
			if err != nil {
				log.Printf("[线程 %d] 打开文件失败: %v", idx, err)
				return
			}
			defer file.Close()

			file.Seek(start, 0)
			_, err = io.CopyBuffer(file, resp.Body, buf)
			if err != nil {
				log.Printf("[线程 %d] 写入失败: %v", idx, err)
			} else {
				log.Printf("[线程 %d] 下载完成 %d - %d", idx, start, end)
			}
		}(start, end, i)
	}

	wg.Wait()
	log.Println("多线程下载完成")

	return nil
}
