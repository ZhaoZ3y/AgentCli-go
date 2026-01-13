package agent

import (
	"agentcli/internal/llm"
	"context"
	"fmt"
)

// ProcessRequestStream 处理用户请求（流式输出）
func (a *Agent) ProcessRequestStream(ctx context.Context, userInput string, onChunk func(string) error) (string, error) {
	// 记录开始处理
	if a.logger != nil {
		a.logger.ThinkingProcess("开始处理", "用户输入: "+userInput)
	}
	
	// 第一步：分析用户意图（静默模式）
	intention, err := a.analyzeIntention(ctx, userInput)
	if err != nil {
		if a.logger != nil {
			a.logger.Error("分析意图失败", err, nil)
		}
		return "", fmt.Errorf("分析意图失败: %w", err)
	}

	if a.logger != nil {
		a.logger.ThinkingProcess("意图分析", intention)
	}

	// 第二步：使用DAG进行深度思考和规划（静默模式）
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

用户请求：%s

请根据用户需求，如果需要使用工具请说明，否则直接回答问题。`, systemPrompt, toolsList, userInput)

	messages := []llm.Message{
		{Role: "user", Content: prompt},
	}

	if a.logger != nil {
		a.logger.ThinkingProcess("发送LLM请求", "模型: "+a.llmClient.Model)
	}

	return a.llmClient.ChatStream(ctx, messages, onChunk)
}
