package agent

import (
	"agentcli/internal/llm"
	"context"
	"encoding/json"
	"fmt"
)

// convertToolsToOpenAIFormat 将工具转换为OpenAI函数调用格式
func (a *Agent) convertToolsToOpenAIFormat() []llm.Tool {
	tools := make([]llm.Tool, 0)

	for _, tool := range a.toolRegistry.List() {
		// 构建参数schema
		properties := make(map[string]interface{})
		required := make([]string, 0)

		for paramName, paramDesc := range tool.GetParams() {
			properties[paramName] = map[string]interface{}{
				"type":        "string",
				"description": paramDesc,
			}
			required = append(required, paramName)
		}

		tools = append(tools, llm.Tool{
			Type: "function",
			Function: llm.FunctionDef{
				Name:        tool.Name(),
				Description: tool.Description(),
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": properties,
					"required":   required,
				},
			},
		})
	}

	return tools
}

// ProcessRequestStream 处理用户请求（流式输出，带对话历史）
func (a *Agent) ProcessRequestStream(ctx context.Context, userInput string, conversationHistory []llm.Message, onChunk func(string) error) (string, error) {
	a.resetContextLog()
	// 记录开始处理
	if a.logger != nil {
		a.logger.ThinkingProcess("开始处理", "用户输入: "+userInput)
	}

	// 第一步：分析用户意图（带思考过程显示和对话历史）
	intention, err := a.analyzeIntentionWithContext(ctx, userInput, conversationHistory)
	if err != nil {
		if a.logger != nil {
			a.logger.Error("分析意图失败", err, nil)
		}
		return "", fmt.Errorf("分析意图失败: %w", err)
	}

	if a.logger != nil {
		a.logger.ThinkingProcess("意图分析", intention)
	}

	// 第二步：使用DAG进行深度思考和规划（带对话历史）
	result, err := a.executeWithDAGStream(ctx, userInput, intention, conversationHistory, onChunk)
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

// executeWithDAGStream 使用DAG执行任务（流式输出，带对话历史）
func (a *Agent) executeWithDAGStream(ctx context.Context, userInput, intention string, conversationHistory []llm.Message, onChunk func(string) error) (string, error) {
	// 构建系统提示词，包含定制化记忆
	systemPrompt := "你是一个智能助手。\n当前系统：" + a.osHint() + "。请仅给出匹配该系统的命令与操作。\n" + a.toolUsagePolicy()
	if a.memory != "" {
		systemPrompt = a.memory + "\n当前系统：" + a.osHint() + "。请仅给出匹配该系统的命令与操作。\n" + a.toolUsagePolicy()
		if a.logger != nil {
			a.logger.ThinkingProcess("应用定制化记忆", a.memory)
		}
	}

	systemPrompt += "\n\n你可以使用提供的工具来完成任务。当需要使用工具时，系统会自动调用它们。"

	// 构建消息列表：系统提示 + 对话历史 + 当前任务
	messages := []llm.Message{
		{Role: "system", Content: systemPrompt},
	}

	// 添加对话历史
	messages = append(messages, conversationHistory...)

	// 添加当前任务
	messages = append(messages, llm.Message{
		Role:    "user",
		Content: fmt.Sprintf("前置分析：%s\n\n用户请求：%s", intention, userInput),
	})

	// 转换工具为OpenAI格式
	tools := a.convertToolsToOpenAIFormat()

	if a.logger != nil {
		a.logger.ThinkingProcess("准备工具", fmt.Sprintf("可用工具数量: %d", len(tools)))
	}

	// 执行函数调用循环
	maxIterations := 10
	for i := 0; i < maxIterations; i++ {
		if a.logger != nil {
			a.logger.ThinkingProcess("LLM调用", fmt.Sprintf("迭代 %d/%d", i+1, maxIterations))
		}

		// 调用LLM（带工具）
		response, err := a.llmClient.Chat(ctx, messages, tools, "auto")
		if err != nil {
			return "", fmt.Errorf("LLM调用失败: %w", err)
		}

		// 检查是否有工具调用
		if len(response.Choices) == 0 {
			return "", fmt.Errorf("LLM返回空响应")
		}

		choice := response.Choices[0]

		// 如果没有工具调用，说明LLM给出了最终答案
		if len(choice.Message.ToolCalls) == 0 {
			// 流式输出最终答案
			if a.logger != nil {
				fmt.Printf("\n🤖 Agent: ")
			}

			// 直接输出内容（因为已经从Chat获取了完整响应）
			if choice.Message.Content != "" {
				if err := onChunk(choice.Message.Content); err != nil {
					return "", err
				}
			}

			return choice.Message.Content, nil
		}

		// 有工具调用，执行工具
		if a.logger != nil {
			a.logger.ThinkingProcess("工具调用", fmt.Sprintf("需要执行 %d 个工具", len(choice.Message.ToolCalls)))
		}

		// 将助手的消息（包含工具调用）添加到历史
		messages = append(messages, llm.Message{
			Role:      "assistant",
			Content:   choice.Message.Content,
			ToolCalls: choice.Message.ToolCalls,
		})

		// 执行每个工具调用
		for _, toolCall := range choice.Message.ToolCalls {
			if toolCall.Type != "function" {
				continue
			}

			funcName := toolCall.Function.Name
			funcArgs := toolCall.Function.Arguments

			if a.logger != nil {
				onChunk(fmt.Sprintf("\n⚙️ 执行工具: %s\n", funcName))
				a.logger.ThinkingProcess("执行工具", fmt.Sprintf("%s(%s)", funcName, funcArgs))
			} else {
				onChunk(fmt.Sprintf("\n⚙️ 执行工具: %s\n", funcName))
			}

			// 解析参数
			var params map[string]interface{}
			if err := json.Unmarshal([]byte(funcArgs), &params); err != nil {
				errMsg := fmt.Sprintf("参数解析失败: %v", err)
				onChunk(fmt.Sprintf("❌ %s\n", errMsg))

				// 将错误结果添加到消息历史
				messages = append(messages, llm.Message{
					Role:       "tool",
					Content:    errMsg,
					ToolCallID: toolCall.ID,
				})
				continue
			}

			// 获取并执行工具
			tool, err := a.toolRegistry.Get(funcName)
			if err != nil {
				errMsg := fmt.Sprintf("工具不存在: %v", err)
				onChunk(fmt.Sprintf("❌ %s\n", errMsg))

				messages = append(messages, llm.Message{
					Role:       "tool",
					Content:    errMsg,
					ToolCallID: toolCall.ID,
				})
				continue
			}

			// 执行工具
			result, err := tool.Execute(ctx, params)
			a.recordToolCallContext(funcName, params, result, err)
			if err != nil {
				errMsg := fmt.Sprintf("执行失败: %v", err)
				onChunk(fmt.Sprintf("❌ %s\n", errMsg))

				messages = append(messages, llm.Message{
					Role:       "tool",
					Content:    errMsg,
					ToolCallID: toolCall.ID,
				})
				continue
			}

			// 格式化结果
			resultJSON, _ := json.Marshal(result)
			resultStr := string(resultJSON)

			onChunk(fmt.Sprintf("✅ 执行成功\n"))

			if a.logger != nil {
				a.logger.ThinkingProcess("工具结果", resultStr)
			}

			// 将工具结果添加到消息历史
			messages = append(messages, llm.Message{
				Role:       "tool",
				Content:    resultStr,
				ToolCallID: toolCall.ID,
			})
		}

		onChunk("\n")
	}

	return "", fmt.Errorf("达到最大迭代次数 (%d)，任务未完成", maxIterations)
}
