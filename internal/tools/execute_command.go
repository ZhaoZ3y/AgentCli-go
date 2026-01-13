package tools

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"
)

// ExecuteCommandTool 执行命令工具
type ExecuteCommandTool struct {
	timeout time.Duration
}

// NewExecuteCommandTool 创建执行命令工具
func NewExecuteCommandTool(timeout time.Duration) *ExecuteCommandTool {
	return &ExecuteCommandTool{
		timeout: timeout,
	}
}

func (t *ExecuteCommandTool) Name() string {
	return "execute_command"
}

func (t *ExecuteCommandTool) Description() string {
	return "执行系统命令。参数: command(命令), args(参数列表,可选)"
}

func (t *ExecuteCommandTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// 获取参数
	command, ok := params["command"].(string)
	if !ok || command == "" {
		return nil, fmt.Errorf("缺少命令参数")
	}

	// 创建超时上下文
	cmdCtx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	// 根据操作系统选择shell
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(cmdCtx, "cmd", "/c", command)
	} else {
		cmd = exec.CommandContext(cmdCtx, "sh", "-c", command)
	}

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 检查是否超时
		if cmdCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("命令执行超时")
		}
		return map[string]interface{}{
			"command": command,
			"output":  string(output),
			"error":   err.Error(),
			"success": false,
		}, nil
	}

	return map[string]interface{}{
		"command": command,
		"output":  string(output),
		"success": true,
	}, nil
}
