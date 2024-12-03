package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

// 定义远程 TXT 文件 URL
const remoteTxtURL = "https://xz.gqru.com/xiazai.txt"

// 最大并发数
const maxConcurrentDownloads = 2

func main() {
	// 从远程获取下载列表
	downloadList, err := fetchDownloadList(remoteTxtURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取下载列表失败: %v\n", err)
		return
	}

	if len(downloadList) == 0 {
		fmt.Fprintf(os.Stderr, "下载列表为空，程序退出\n")
		return
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrentDownloads) // 控制最大并发数

	for i, url := range downloadList {
		wg.Add(1)
		go func(i int, url string) {
			defer wg.Done()
			semaphore <- struct{}{} // 占用一个并发槽
			defer func() { <-semaphore }() // 释放一个并发槽

			// 静默处理下载任务，不输出日志
			_ = downloadFile(url, fmt.Sprintf("file_%d", i+1))
		}(i, url)
	}

	wg.Wait()
}

// 从远程 TXT 文件获取下载地址
func fetchDownloadList(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("无法访问远程文件: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("远程文件访问失败，状态码: %d", resp.StatusCode)
	}

	var downloadList []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			downloadList = append(downloadList, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取远程文件内容失败: %v", err)
	}

	return downloadList, nil
}

// 下载文件
func downloadFile(url, fileName string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err // 静默处理错误
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	// 创建目标文件
	out, err := os.Create(fileName)
	if err != nil {
		return err // 静默处理错误
	}
	defer out.Close()

	// 将响应体写入文件
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err // 静默处理错误
	}

	return nil
}
