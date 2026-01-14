package agent

import (
	"agentcli/internal/llm"
	"context"
	"encoding/json"
	"fmt"
)

// ProcessRequestStream 处理用户请求（流式输出）
func (a *Agent) ProcessRequestStream(ctx context.Context, userInput string, onChunk func(string) error) (string, error) {
	// 记录开始处理
	if a.logger != nil {
		a.logger.ThinkingProcess("开始处理", "用户输入: "+userInput)
	}
	
	// 第一步：分析用户意图（带思考过程显示）
	intention, err := a.analyzeIntentionWithContext(ctx, userInput)
	if err != nil {
		if a.logger != nil {
			a.logger.Error("分析意图失败", err, nil)
		}
		return "", fmt.Errorf("分析意图失败: %w", err)
	}

	if a.logger != nil {
		a.logger.ThinkingProcess("意图分析", intention)
	}

	// 第二步：使用DAG进行深度思考和规划
	result, err := a.executeWithDAGStream(ctx, userInput, intention, onChunk)
	if err != nil {
		if a.logger != nil {
			a.logger.Error("执行失败", err, nil)
		}
		return "", fmt.Errorf("执行失败: %w", err)
	}

	if a.logger != nil {
		a.logger.ThinkingProcess("完成处理", "输出长度: "+fmt.Sprintf("%d", len(result)))
	}

	return result, nil
}

// executeWithDAGStream 使用DAG执行任务（流式输出）
func (a *Agent) executeWithDAGStream(ctx context.Context, userInput, intention string, onChunk func(string) error) (string, error) {
	// 简化版本：直接调用LLM流式输出，不使用复杂的DAG
	toolsList := a.getToolsDescription()
	
	// 构建提示词，包含定制化记忆
	systemPrompt := "你是一个智能助手。"
	if a.memory != "" {
		systemPrompt = a.memory
		if a.logger != nil {
			a.logger.ThinkingProcess("应用定制化记忆", a.memory)
		}
	}
	
	prompt := fmt.Sprintf(`%s

可用工具：
%s

前置分析与操作：
%s

用户请求：%s

请根据用户需求和前置分析结果（可能已经读取了文件），如果任务已完成请直接回答。
如果需要使用工具，请在回答的最后以JSON数组格式输出工具调用计划（不要使用Markdown代码块），格式如下：
[{"tool": "tool_name", "params": {"param1": "value1"}}]
`, systemPrompt, toolsList, intention, userInput)

	messages := []llm.Message{
		{Role: "user", Content: prompt},
	}

	if a.logger != nil {
		fmt.Printf("\n🤖 Agent: ")
		a.logger.ThinkingProcess("发送LLM请求", "模型: "+a.llmClient.Model)
	}

	response, err := a.llmClient.ChatStream(ctx, messages, onChunk)
	if err != nil {
		return "", err
	}

	// 尝试解析并执行工具
	jsonStr := extractJSON(response)
	if jsonStr != "" && jsonStr != response {
		var toolCalls []struct {
			Tool   string                 `json:"tool"`
			Params map[string]interface{} `json:"params"`
		}

		if err := json.Unmarshal([]byte(jsonStr), &toolCalls); err == nil && len(toolCalls) > 0 {
			onChunk("\n\n") // 换行
			for _, call := range toolCalls {
				tool, err := a.toolRegistry.Get(call.Tool)
				if err != nil {
					msg := fmt.Sprintf("❌ 工具 %s 不存在\n", call.Tool)
					onChunk(msg)
					continue
				}

				if a.logger != nil {
					a.logger.ThinkingProcess("执行工具", fmt.Sprintf("%s: %v", call.Tool, call.Params))
				} else {
					onChunk(fmt.Sprintf("⚙️ 执行工具: %s...\n", call.Tool))
				}

				result, err := tool.Execute(ctx, call.Params)
				if err != nil {
					msg := fmt.Sprintf("❌ 执行失败: %v\n", err)
					onChunk(msg)
				} else {
					resultJSON, _ := json.MarshalIndent(result, "", "  ")
					msg := fmt.Sprintf("✅ 执行成功:\n%s\n", string(resultJSON))
					onChunk(msg)
				}
			}
		}
	}

	return response, nil
}
