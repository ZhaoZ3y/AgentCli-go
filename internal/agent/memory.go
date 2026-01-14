package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// MemoryStore 记忆存储
type MemoryStore struct {
	UserID    string    `json:"user_id"`
	Memory    string    `json:"memory"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SaveMemoryToFile 保存记忆到文件
func SaveMemoryToFile(userID, memory string) error {
	// 创建memory目录
	memoryDir := "memory"
	if err := os.MkdirAll(memoryDir, 0755); err != nil {
		return fmt.Errorf("创建memory目录失败: %w", err)
	}

	// 构建文件路径
	filePath := filepath.Join(memoryDir, fmt.Sprintf("%s.json", userID))

	// 创建记忆存储对象
	store := MemoryStore{
		UserID:    userID,
		Memory:    memory,
		UpdatedAt: time.Now(),
	}

	// 序列化为JSON
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化记忆失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入记忆文件失败: %w", err)
	}

	return nil
}

// LoadMemoryFromFile 从文件加载记忆
func LoadMemoryFromFile(userID string) (string, error) {
	// 构建文件路径
	filePath := filepath.Join("memory", fmt.Sprintf("%s.json", userID))

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", nil // 文件不存在，返回空字符串
	}

	// 读取文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("读取记忆文件失败: %w", err)
	}

	// 反序列化
	var store MemoryStore
	if err := json.Unmarshal(data, &store); err != nil {
		return "", fmt.Errorf("解析记忆文件失败: %w", err)
	}

	return store.Memory, nil
}
