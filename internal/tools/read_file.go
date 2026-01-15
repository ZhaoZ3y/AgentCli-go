package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadFileTool 读取文件工具
type ReadFileTool struct {
	maxSizeMB         int
	allowedExtensions []string
}

// NewReadFileTool 创建读取文件工具
func NewReadFileTool(maxSizeMB int, allowedExtensions []string) *ReadFileTool {
	return &ReadFileTool{
		maxSizeMB:         maxSizeMB,
		allowedExtensions: allowedExtensions,
	}
}

func (t *ReadFileTool) Name() string {
	return "read_file"
}

func (t *ReadFileTool) Description() string {
	return "读取文件内容。参数: filepath(文件路径)"
}

func (t *ReadFileTool) GetParams() map[string]string {
	return map[string]string{
		"filepath": "要读取的文件路径",
	}
}

func (t *ReadFileTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// 获取参数
	filePath, ok := params["filepath"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("缺少文件路径参数")
	}

	// 检查文件是否存在
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("文件不存在: %s", filePath)
		}
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 检查是否是文件
	if info.IsDir() {
		return nil, fmt.Errorf("路径是目录，不是文件: %s", filePath)
	}

	// 检查文件大小
	maxBytes := int64(t.maxSizeMB) * 1024 * 1024
	if info.Size() > maxBytes {
		return nil, fmt.Errorf("文件大小超过限制: %d MB > %d MB", info.Size()/(1024*1024), t.maxSizeMB)
	}

	// 检查文件扩展名
	ext := filepath.Ext(filePath)
	if !t.isExtensionAllowed(ext) {
		return nil, fmt.Errorf("不支持的文件扩展名: %s", ext)
	}

	// 读取文件
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	return map[string]interface{}{
		"filepath": filePath,
		"content":  string(content),
		"size":     info.Size(),
		"lines":    strings.Count(string(content), "\n") + 1,
	}, nil
}

func (t *ReadFileTool) isExtensionAllowed(ext string) bool {
	for _, allowed := range t.allowedExtensions {
		if strings.EqualFold(allowed, ext) {
			return true
		}
	}
	return false
}
