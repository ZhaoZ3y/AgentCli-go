package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client LLM客户端
type Client struct {
	apiKey  string
	baseURL string
	Model   string // 改为公开字段，允许外部修改
	timeout time.Duration
	client  *http.Client
}

// Message 消息结构
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Model      string    `json:"model"`
	Messages   []Message `json:"messages"`
	Tools      []Tool    `json:"tools,omitempty"`
	ToolChoice string    `json:"tool_choice,omitempty"`
}

// Tool 工具定义
type Tool struct {
	Type     string       `json:"type"`
	Function FunctionDef  `json:"function"`
}

// FunctionDef 函数定义
type FunctionDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolCall 工具调用
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall 函数调用
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ChatMessage 扩展的消息结构（包含工具调用）
type ChatMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Index   int         `json:"index"`
		Message ChatMessage `json:"message"`
		Finish  string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// NewClient 创建LLM客户端
func NewClient(apiKey, baseURL, model string, timeout time.Duration) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		Model:   model,
		timeout: timeout,
		client:  &http.Client{Timeout: timeout},
	}
}

// Chat 发送聊天请求（带工具支持）
func (c *Client) Chat(ctx context.Context, messages []Message, tools []Tool, toolChoice string) (*ChatResponse, error) {
	// 构建请求
	reqBody := ChatRequest{
		Model:      c.Model,
		Messages:   messages,
		Tools:      tools,
		ToolChoice: toolChoice,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 构建URL，确保正确处理斜杠
	baseURL := strings.TrimRight(c.baseURL, "/")
	url := fmt.Sprintf("%s/chat/completions", baseURL)

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API请求失败 (status %d): %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w\n响应内容: %s", err, string(body))
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("响应中没有消息")
	}

	return &chatResp, nil
}

// SimpleQuery 简单查询
func (c *Client) SimpleQuery(ctx context.Context, prompt string) (string, error) {
	messages := []Message{
		{Role: "user", Content: prompt},
	}
	resp, err := c.Chat(ctx, messages, nil, "")
	if err != nil {
		return "", err
	}
	
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("响应中没有消息")
	}
	
	return resp.Choices[0].Message.Content, nil
}
