package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// StreamResponse 流式响应
type StreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role      string     `json:"role,omitempty"`
			Content   string     `json:"content,omitempty"`
			ToolCalls []ToolCall `json:"tool_calls,omitempty"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason,omitempty"`
	} `json:"choices"`
}

// ChatStream 发送流式聊天请求
func (c *Client) ChatStream(ctx context.Context, messages []Message, onChunk func(content string) error) (string, error) {
	return c.ChatStreamWithTools(ctx, messages, nil, "", onChunk)
}

// ChatStreamWithTools 发送带工具的流式聊天请求
func (c *Client) ChatStreamWithTools(ctx context.Context, messages []Message, tools []Tool, toolChoice string, onChunk func(content string) error) (string, error) {
	// 构建请求
	reqBody := map[string]interface{}{
		"model":    c.Model,
		"messages": messages,
		"stream":   true,
	}
	
	if len(tools) > 0 {
		reqBody["tools"] = tools
		if toolChoice != "" {
			reqBody["tool_choice"] = toolChoice
		}
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	// 构建URL
	baseURL := strings.TrimRight(c.baseURL, "/")
	url := fmt.Sprintf("%s/chat/completions", baseURL)

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Accept", "text/event-stream")

	// 发送请求
	// 流式请求可能持续很长时间，创建一个没有超时的客户端副本
	streamClient := *c.client
	streamClient.Timeout = 0
	resp, err := streamClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API请求失败 (status %d): %s", resp.StatusCode, string(body))
	}

	// 读取流式响应
	var fullContent strings.Builder
	reader := bufio.NewReader(resp.Body)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("读取流失败: %w", err)
		}

		// 跳过空行
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// SSE格式: data: {...}
		if bytes.HasPrefix(line, []byte("data: ")) {
			data := bytes.TrimPrefix(line, []byte("data: "))
			
			// 检查结束标记
			if bytes.Equal(data, []byte("[DONE]")) {
				break
			}

			// 解析JSON
			var streamResp StreamResponse
			if err := json.Unmarshal(data, &streamResp); err != nil {
				continue // 跳过无法解析的行
			}

			// 提取内容
			if len(streamResp.Choices) > 0 {
				content := streamResp.Choices[0].Delta.Content
				if content != "" {
					fullContent.WriteString(content)
					// 调用回调函数
					if onChunk != nil {
						if err := onChunk(content); err != nil {
							return "", err
						}
					}
				}
			}
		}
	}

	return fullContent.String(), nil
}
