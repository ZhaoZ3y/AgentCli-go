package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"agentcli/internal/llm"
)

// Message 消息
type Message struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Conversation 对话
type Conversation struct {
	ID       string    `json:"id"`
	UserID   string    `json:"user_id"`
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
}

// Manager 历史记录管理器
type Manager struct {
	historyDir string
}

// NewManager 创建历史记录管理器
func NewManager(historyDir string) *Manager {
	return &Manager{
		historyDir: historyDir,
	}
}

// Init 初始化历史记录目录
func (m *Manager) Init() error {
	return os.MkdirAll(m.historyDir, 0755)
}

// SaveConversation 保存对话
func (m *Manager) SaveConversation(conv *Conversation) error {
	conv.Updated = time.Now()
	
	filename := filepath.Join(m.historyDir, fmt.Sprintf("%s.json", conv.ID))
	data, err := json.MarshalIndent(conv, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化对话失败: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("保存对话失败: %w", err)
	}

	return nil
}

// LoadConversation 加载对话
func (m *Manager) LoadConversation(id string) (*Conversation, error) {
	filename := filepath.Join(m.historyDir, fmt.Sprintf("%s.json", id))
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("对话不存在: %s", id)
		}
		return nil, fmt.Errorf("读取对话失败: %w", err)
	}

	var conv Conversation
	if err := json.Unmarshal(data, &conv); err != nil {
		return nil, fmt.Errorf("解析对话失败: %w", err)
	}

	return &conv, nil
}

// ListConversations 列出所有对话
func (m *Manager) ListConversations(userID string) ([]*Conversation, error) {
	files, err := os.ReadDir(m.historyDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Conversation{}, nil
		}
		return nil, fmt.Errorf("读取历史目录失败: %w", err)
	}

	var conversations []*Conversation
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		id := file.Name()[:len(file.Name())-5] // 移除 .json
		conv, err := m.LoadConversation(id)
		if err != nil {
			continue
		}

		if userID == "" || conv.UserID == userID {
			conversations = append(conversations, conv)
		}
	}

	return conversations, nil
}

// DeleteConversation 删除对话
func (m *Manager) DeleteConversation(id string) error {
	filename := filepath.Join(m.historyDir, fmt.Sprintf("%s.json", id))
	if err := os.Remove(filename); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("对话不存在: %s", id)
		}
		return fmt.Errorf("删除对话失败: %w", err)
	}
	return nil
}

// NewConversation 创建新对话
func NewConversation(userID, model string) *Conversation {
	now := time.Now()
	return &Conversation{
		ID:       fmt.Sprintf("%s_%d", userID, now.Unix()),
		UserID:   userID,
		Model:    model,
		Messages: []Message{},
		Created:  now,
		Updated:  now,
	}
}

// AddMessage 添加消息到对话
func (c *Conversation) AddMessage(role, content string) {
	c.Messages = append(c.Messages, Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
}

// GetRecentMessages 获取最近N条消息
func (c *Conversation) GetRecentMessages(n int) []Message {
	if n <= 0 || n >= len(c.Messages) {
		return c.Messages
	}
	return c.Messages[len(c.Messages)-n:]
}

// ToLLMMessages 转换消息为LLM格式
func (c *Conversation) ToLLMMessages() []llm.Message {
	messages := make([]llm.Message, 0, len(c.Messages))
	for _, msg := range c.Messages {
		messages = append(messages, llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	return messages
}

// History 历史记录包装器，用于Agent
type History struct {
	conversation *Conversation
}

// NewHistory 创建历史记录包装器
func NewHistory(conv *Conversation) *History {
	return &History{
		conversation: conv,
	}
}

// GetMessages 获取消息列表（转换为LLM消息格式）
func (h *History) GetMessages() []interface{} {
	messages := make([]interface{}, 0, len(h.conversation.Messages))
	for _, msg := range h.conversation.Messages {
		messages = append(messages, map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}
	return messages
}

// AddMessage 添加消息
func (h *History) AddMessage(role, content string) {
	h.conversation.AddMessage(role, content)
}

// Clear 清空历史记录
func (h *History) Clear() {
	h.conversation.Messages = []Message{}
}

// GetConversation 获取对话对象
func (h *History) GetConversation() *Conversation {
	return h.conversation
}
