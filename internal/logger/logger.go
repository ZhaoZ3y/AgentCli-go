package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Logger 日志记录器
type Logger struct {
	sessionID string
	logFile   *os.File
	mu        sync.Mutex
}

// NewLogger 创建新的日志记录器
func NewLogger(sessionID string) (*Logger, error) {
	// 创建日志目录（当前目录下）
	today := time.Now().Format("2006-01-02")
	logDir := filepath.Join("logs", today)
	
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 创建日志文件
	logPath := filepath.Join(logDir, fmt.Sprintf("%s.log", sessionID))
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("创建日志文件失败: %w", err)
	}

	logger := &Logger{
		sessionID: sessionID,
		logFile:   file,
	}

	logger.Info("会话开始", map[string]interface{}{
		"session_id": sessionID,
		"timestamp":  time.Now().Format(time.RFC3339),
	})

	return logger, nil
}

// Info 记录信息日志
func (l *Logger) Info(message string, data map[string]interface{}) {
	l.log("INFO", message, data)
}

// Debug 记录调试日志
func (l *Logger) Debug(message string, data map[string]interface{}) {
	l.log("DEBUG", message, data)
}

// Error 记录错误日志
func (l *Logger) Error(message string, err error, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}
	if err != nil {
		data["error"] = err.Error()
	}
	l.log("ERROR", message, data)
}

// UserInput 记录用户输入
func (l *Logger) UserInput(input string) {
	l.log("USER_INPUT", input, nil)
}

// AgentOutput 记录Agent输出
func (l *Logger) AgentOutput(output string) {
	l.log("AGENT_OUTPUT", output, nil)
}

// ThinkingProcess 记录思考过程
func (l *Logger) ThinkingProcess(stage string, content string) {
	l.log("THINKING", stage, map[string]interface{}{
		"content": content,
	})
}

// ToolCall 记录工具调用
func (l *Logger) ToolCall(toolName string, params map[string]interface{}, result interface{}, err error) {
	data := map[string]interface{}{
		"tool":   toolName,
		"params": params,
		"result": result,
	}
	if err != nil {
		data["error"] = err.Error()
	}
	l.log("TOOL_CALL", toolName, data)
}

// log 内部日志记录方法
func (l *Logger) log(level, message string, data map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	logLine := fmt.Sprintf("[%s] [%s] %s", timestamp, level, message)

	if data != nil && len(data) > 0 {
		logLine += fmt.Sprintf(" | Data: %+v", data)
	}

	logLine += "\n"

	if l.logFile != nil {
		l.logFile.WriteString(logLine)
		l.logFile.Sync()
	}
}

// Close 关闭日志记录器
func (l *Logger) Close() error {
	l.Info("会话结束", map[string]interface{}{
		"session_id": l.sessionID,
		"timestamp":  time.Now().Format(time.RFC3339),
	})

	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}
