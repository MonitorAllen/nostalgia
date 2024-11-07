package util

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"path/filepath"
)

// DownloadFile 从 URL 下载文件并保存到本地
func DownloadFile(client *resty.Client, url, dst string) error {
	// 执行 HTTP GET 请求
	resp, err := client.R().
		SetOutput(dst). // 直接将响应内容保存到目标文件
		Get(url)

	if err != nil {
		return fmt.Errorf("failed to download file from %s: %v", url, err)
	}

	// 检查请求状态码
	if resp.StatusCode() != 200 {
		return fmt.Errorf("error downloading %s: status code %d", url, resp.StatusCode())
	}

	return nil
}

// DownloadFiles 批量下载文件并保存到指定目录
func DownloadFiles(urls []string, dstDir string) error {
	// 创建 Resty 客户端
	client := resty.New()

	// 遍历 URL 列表
	for _, url := range urls {
		// 提取文件名
		filename := filepath.Base(url)

		// 拼接目标路径
		dst := filepath.Join(dstDir, filename)

		// 下载并保存文件
		err := DownloadFile(client, url, dst)
		if err != nil {
			return fmt.Errorf("error downloading file: %v\n", err)
		}
	}
	return nil
}
