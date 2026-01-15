package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteCodeTool 写代码工具
type WriteCodeTool struct {
	maxLines           int
	supportedLanguages []string
}

// NewWriteCodeTool 创建写代码工具
func NewWriteCodeTool(maxLines int, supportedLanguages []string) *WriteCodeTool {
	return &WriteCodeTool{
		maxLines:           maxLines,
		supportedLanguages: supportedLanguages,
	}
}

func (t *WriteCodeTool) Name() string {
	return "write_code"
}

func (t *WriteCodeTool) Description() string {
	return "写入代码到文件。参数: filepath(文件路径), code(代码内容), language(编程语言)"
}

func (t *WriteCodeTool) GetParams() map[string]string {
	return map[string]string{
		"filepath": "要写入的文件路径",
		"code":     "要写入的代码内容",
		"language": "编程语言(可选，可从文件扩展名推断)",
	}
}

func (t *WriteCodeTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// 获取参数 - 支持filepath和file_path两种参数名
	filePath, ok := params["filepath"].(string)
	if !ok || filePath == "" {
		filePath, ok = params["file_path"].(string)
		if !ok || filePath == "" {
			return nil, fmt.Errorf("缺少文件路径参数")
		}
	}

	code, ok := params["code"].(string)
	if !ok || code == "" {
		return nil, fmt.Errorf("缺少代码内容参数")
	}

	// 获取语言参数 - 如果未提供，从文件扩展名推断
	language, ok := params["language"].(string)
	if !ok || language == "" {
		// 从文件扩展名推断语言
		ext := filepath.Ext(filePath)
		switch strings.ToLower(ext) {
		case ".py":
			language = "python"
		case ".go":
			language = "go"
		case ".js":
			language = "javascript"
		case ".ts":
			language = "typescript"
		case ".java":
			language = "java"
		case ".c":
			language = "c"
		case ".cpp", ".cc", ".cxx":
			language = "cpp"
		default:
			return nil, fmt.Errorf("无法推断编程语言，请指定language参数")
		}
	}

	// 验证编程语言
	if !t.isLanguageSupported(language) {
		return nil, fmt.Errorf("不支持的编程语言: %s", language)
	}

	// 验证代码行数
	lines := strings.Split(code, "\n")
	if len(lines) > t.maxLines {
		return nil, fmt.Errorf("代码行数超过限制: %d > %d", len(lines), t.maxLines)
	}

	// 创建目录
	dir := filepath.Dir(filePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("创建目录失败: %w", err)
		}
	}

	// 写入文件
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		return nil, fmt.Errorf("写入文件失败: %w", err)
	}

	return map[string]interface{}{
		"filepath": filePath,
		"lines":    len(lines),
		"bytes":    len(code),
	}, nil
}

func (t *WriteCodeTool) isLanguageSupported(lang string) bool {
	for _, supported := range t.supportedLanguages {
		if strings.EqualFold(supported, lang) {
			return true
		}
	}
	return false
}
