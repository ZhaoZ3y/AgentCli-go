package agent

import (
	"fmt"
	"strings"
)

func (a *Agent) resetContextLog() {
	if a == nil {
		return
	}
	a.contextMu.Lock()
	defer a.contextMu.Unlock()
	a.contextEntries = nil
}

func (a *Agent) appendContextEntry(kind, content string) {
	if a == nil {
		return
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return
	}
	a.contextMu.Lock()
	defer a.contextMu.Unlock()
	a.contextEntries = append(a.contextEntries, fmt.Sprintf("[%s] %s", kind, content))
}

func (a *Agent) ConsumeContextLog() string {
	if a == nil {
		return ""
	}
	a.contextMu.Lock()
	defer a.contextMu.Unlock()
	if len(a.contextEntries) == 0 {
		return ""
	}
	combined := strings.Join(a.contextEntries, "\n\n")
	a.contextEntries = nil
	return combined
}

func (a *Agent) recordToolCallContext(toolName string, params map[string]interface{}, result interface{}, err error) {
	if a == nil || toolName != "execute_command" {
		return
	}
	commandLine := formatExecuteCommand(params)
	if commandLine == "" {
		return
	}

	entry := commandLine
	if err != nil {
		entry = fmt.Sprintf("%s | error=%v", commandLine, err)
	} else if resultMap, ok := result.(map[string]interface{}); ok {
		if success, ok := resultMap["success"].(bool); ok {
			entry = fmt.Sprintf("%s | success=%t", commandLine, success)
		}
		if errMsg, ok := resultMap["error"].(string); ok && errMsg != "" {
			entry = fmt.Sprintf("%s | error=%s", commandLine, errMsg)
		}
	}

	a.appendContextEntry("execute_command", entry)
}

func formatExecuteCommand(params map[string]interface{}) string {
	if params == nil {
		return ""
	}
	command, _ := params["command"].(string)
	command = strings.TrimSpace(command)
	if command == "" {
		return ""
	}
	args := extractArgs(params["args"])
	if len(args) == 0 {
		return command
	}
	return command + " " + strings.Join(args, " ")
}

func extractArgs(raw interface{}) []string {
	if raw == nil {
		return nil
	}

	switch v := raw.(type) {
	case []string:
		return append([]string{}, v...)
	case []interface{}:
		args := make([]string, 0, len(v))
		for _, item := range v {
			switch itemTyped := item.(type) {
			case string:
				if itemTyped != "" {
					args = append(args, itemTyped)
				}
			default:
				args = append(args, fmt.Sprint(itemTyped))
			}
		}
		return args
	case string:
		if strings.TrimSpace(v) == "" {
			return nil
		}
		return []string{v}
	default:
		return []string{fmt.Sprint(v)}
	}
}
