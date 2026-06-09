package util

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-resty/resty/v2"
	"golang.org/x/net/html"
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

func ListFiles(dirPath string) ([]string, error) {
	// 读取目录中的所有文件和子目录
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var fileNames []string
	for _, entry := range entries {
		if !entry.IsDir() { // 判断是否为文件
			fileNames = append(fileNames, entry.Name())
		}
	}

	return fileNames, nil
}

func ExtractFileNames(content string) []string {
	var fileNames []string
	seen := make(map[string]struct{})
	tokenizer := html.NewTokenizer(strings.NewReader(content))
	resourceAttrs := map[string]struct{}{
		"href":     {},
		"poster":   {},
		"src":      {},
		"data-src": {},
	}

	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			if tokenizer.Err() == io.EOF {
				break
			}
			return fileNames
		}

		token := tokenizer.Token()
		for _, attr := range token.Attr {
			if _, ok := resourceAttrs[strings.ToLower(attr.Key)]; !ok {
				continue
			}

			fileName := extractResourceFileName(attr.Val)
			if fileName == "" {
				continue
			}
			if _, ok := seen[fileName]; ok {
				continue
			}

			fileNames = append(fileNames, fileName)
			seen[fileName] = struct{}{}
		}
	}

	return fileNames
}

func extractResourceFileName(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	path := parsed.EscapedPath()
	if path == "" {
		path = raw
	}
	normalizedPath := "/" + strings.TrimPrefix(path, "/")
	if !strings.Contains(normalizedPath, "/resources/") {
		return ""
	}

	fileName, err := url.PathUnescape(filepath.Base(normalizedPath))
	if err != nil {
		return ""
	}
	if fileName == "." || fileName == "/" {
		return ""
	}

	return fileName
}

func dirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // 不存在
		}
		return false, err // 其他错误
	}
	return info.IsDir(), nil // 存在且是目录
}
