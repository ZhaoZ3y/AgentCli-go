package tools

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RecognizeImageTool 图片识别工具
type RecognizeImageTool struct {
	maxSizeMB        int
	supportedFormats []string
	apiClient        ImageAPIClient
}

// ImageAPIClient 图片API客户端接口
type ImageAPIClient interface {
	RecognizeImage(ctx context.Context, imageData string) (string, error)
}

// NewRecognizeImageTool 创建图片识别工具
func NewRecognizeImageTool(maxSizeMB int, supportedFormats []string, apiClient ImageAPIClient) *RecognizeImageTool {
	return &RecognizeImageTool{
		maxSizeMB:        maxSizeMB,
		supportedFormats: supportedFormats,
		apiClient:        apiClient,
	}
}

func (t *RecognizeImageTool) Name() string {
	return "recognize_image"
}

func (t *RecognizeImageTool) Description() string {
	return "识别图片内容。参数: filepath(图片文件路径)"
}

func (t *RecognizeImageTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
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

	// 检查文件大小
	maxBytes := int64(t.maxSizeMB) * 1024 * 1024
	if info.Size() > maxBytes {
		return nil, fmt.Errorf("图片大小超过限制: %d MB > %d MB", info.Size()/(1024*1024), t.maxSizeMB)
	}

	// 检查图片格式
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(filePath)), ".")
	if !t.isFormatSupported(ext) {
		return nil, fmt.Errorf("不支持的图片格式: %s", ext)
	}

	// 读取图片
	imageData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取图片失败: %w", err)
	}

	// 编码为base64
	base64Data := base64.StdEncoding.EncodeToString(imageData)

	// 调用API识别图片
	if t.apiClient != nil {
		description, err := t.apiClient.RecognizeImage(ctx, base64Data)
		if err != nil {
			return nil, fmt.Errorf("图片识别失败: %w", err)
		}

		return map[string]interface{}{
			"filepath":    filePath,
			"size":        info.Size(),
			"format":      ext,
			"description": description,
		}, nil
	}

	return map[string]interface{}{
		"filepath": filePath,
		"size":     info.Size(),
		"format":   ext,
		"message":  "图片识别API未配置",
	}, nil
}

func (t *RecognizeImageTool) isFormatSupported(format string) bool {
	for _, supported := range t.supportedFormats {
		if strings.EqualFold(supported, format) {
			return true
		}
	}
	return false
}
