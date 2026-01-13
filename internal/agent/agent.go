package agent

import (
	"agentcli/internal/config"
	"agentcli/internal/dag"
	"agentcli/internal/llm"
	"agentcli/internal/logger"
	"agentcli/internal/tools"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Agent 代理
type Agent struct {
	llmClient    *llm.Client
	toolRegistry *tools.ToolRegistry
	config       *config.Config
	logger       *logger.Logger
	memory       string // 定制化记忆
}

// NewAgent 创建代理
func NewAgent(cfg *config.Config, log *logger.Logger) *Agent {
	// 创建LLM客户端
	llmClient := llm.NewClient(
		cfg.API.OpenAIKey,
		cfg.API.BaseURL,
		cfg.API.Model,
		time.Duration(cfg.API.Timeout)*time.Second,
	)

	// 创建工具注册表
	toolRegistry := tools.NewToolRegistry()

	// 注册工具
	if contains(cfg.Tools.Enabled, "write_code") {
		toolRegistry.Register(tools.NewWriteCodeTool(
			cfg.Tools.WriteCode.MaxLines,
			cfg.Tools.WriteCode.SupportedLanguages,
		))
	}

	if contains(cfg.Tools.Enabled, "read_file") {
		toolRegistry.Register(tools.NewReadFileTool(
			cfg.Tools.ReadFile.MaxSizeMB,
			cfg.Tools.ReadFile.AllowedExtensions,
		))
	}

	if contains(cfg.Tools.Enabled, "recognize_image") {
		toolRegistry.Register(tools.NewRecognizeImageTool(
			cfg.Tools.RecognizeImage.MaxSizeMB,
			cfg.Tools.RecognizeImage.SupportedFormats,
			nil, // 图片识别API客户端可以后续实现
		))
	}

	if contains(cfg.Tools.Enabled, "execute_command") {
		toolRegistry.Register(tools.NewExecuteCommandTool(30 * time.Second))
	}

	return &Agent{
		llmClient:    llmClient,
		toolRegistry: toolRegistry,
		config:       cfg,
		logger:       log,
		memory:       "",
	}
}

// SetMemory 设置定制化记忆
func (a *Agent) SetMemory(mem string) {
	a.memory = mem
	if a.logger != nil {
		a.logger.Info("设置定制化记忆", map[string]interface{}{"memory": mem})
	}
}

// UpdateModel 更新模型
func (a *Agent) UpdateModel(model string) {
	a.llmClient.Model = model
	if a.logger != nil {
		a.logger.Info("更新模型", map[string]interface{}{"model": model})
	}
}

// ProcessRequest 处理用户请求
func (a *Agent) ProcessRequest(ctx context.Context, userInput string) (string, error) {
	fmt.Printf("\n🤔 开始深度思考用户意图...\n")

	// 第一步：分析用户意图
	intention, err := a.analyzeIntention(ctx, userInput)
	if err != nil {
		return "", fmt.Errorf("分析意图失败: %w", err)
	}

	fmt.Printf("📊 意图分析: %s\n", intention)

	// 第二步：使用DAG进行深度思考和规划
	result, err := a.executeWithDAG(ctx, userInput, intention)
	if err != nil {
		return "", fmt.Errorf("执行失败: %w", err)
	}

	return result, nil
}

// analyzeIntention 分析用户意图
func (a *Agent) analyzeIntention(ctx context.Context, userInput string) (string, error) {
	toolsList := a.getToolsDescription()

	prompt := fmt.Sprintf(`你是一个智能助手，请分析以下用户请求的意图，并确定需要使用哪些工具。

可用工具：
%s

用户请求：%s

请用一句话简洁地描述用户意图和需要执行的操作。`, toolsList, userInput)

	return a.llmClient.SimpleQuery(ctx, prompt)
}

// executeWithDAG 使用DAG执行任务
func (a *Agent) executeWithDAG(ctx context.Context, userInput, intention string) (string, error) {
	// 创建DAG
	d := dag.NewDAG(
		a.config.DAG.MaxDepth,
		a.config.DAG.ParallelNodes,
		time.Duration(a.config.DAG.Timeout)*time.Second,
		a.config.DAG.Verbose,
	)

	// 创建思考节点
	thinkNode := dag.NewNode("think", "深度思考", dag.NodeTypeThink)
	thinkNode.SetInput("user_input", userInput)
	thinkNode.SetInput("intention", intention)
	thinkNode.SetHandler(&ThinkHandler{agent: a})
	d.AddNode(thinkNode)

	// 创建决策节点
	decisionNode := dag.NewNode("decision", "决策执行", dag.NodeTypeDecision)
	decisionNode.AddDependency("think")
	decisionNode.SetHandler(&DecisionHandler{agent: a})
	d.AddNode(decisionNode)

	// 创建工具执行节点
	toolNode := dag.NewNode("tool", "工具执行", dag.NodeTypeTool)
	toolNode.AddDependency("decision")
	toolNode.SetHandler(&ToolHandler{agent: a})
	d.AddNode(toolNode)

	// 创建总结节点
	summaryNode := dag.NewNode("summary", "总结结果", dag.NodeTypeEnd)
	summaryNode.AddDependency("tool")
	summaryNode.SetHandler(&SummaryHandler{agent: a})
	d.AddNode(summaryNode)

	// 执行DAG
	fmt.Printf("\n🔄 开始执行DAG工作流...\n")
	if err := d.Execute(ctx); err != nil {
		return "", err
	}

	// 获取结果
	results := d.GetResults()
	if summary, ok := results["summary"]["result"].(string); ok {
		return summary, nil
	}

	return "执行完成，但未能获取结果", nil
}

// getToolsDescription 获取工具描述
func (a *Agent) getToolsDescription() string {
	toolsList := a.toolRegistry.List()
	var descriptions []string
	for _, tool := range toolsList {
		descriptions = append(descriptions, fmt.Sprintf("- %s: %s", tool.Name(), tool.Description()))
	}
	return strings.Join(descriptions, "\n")
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ThinkHandler 思考处理器
type ThinkHandler struct {
	agent *Agent
}

func (h *ThinkHandler) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	userInput := input["user_input"].(string)
	intention := input["intention"].(string)

	toolsList := h.agent.getToolsDescription()

	prompt := fmt.Sprintf(`基于用户请求和意图分析，请深度思考如何完成任务。

可用工具：
%s

用户请求：%s
意图分析：%s

请详细分析：
1. 需要执行哪些步骤
2. 需要使用哪些工具
3. 工具的执行顺序
4. 每个工具需要的参数

以JSON格式输出你的思考结果，格式如下：
{
  "steps": ["步骤1", "步骤2", ...],
  "tools_needed": ["tool1", "tool2", ...],
  "reasoning": "你的推理过程"
}`, toolsList, userInput, intention)

	response, err := h.agent.llmClient.SimpleQuery(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"thinking": response,
		"user_input": userInput,
	}, nil
}

// DecisionHandler 决策处理器
type DecisionHandler struct {
	agent *Agent
}

func (h *DecisionHandler) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	thinking := input["thinking"].(string)
	userInput := input["user_input"].(string)

	prompt := fmt.Sprintf(`基于以下思考结果，生成具体的工具调用计划。

思考结果：
%s

用户请求：%s

请以JSON数组格式输出需要调用的工具及其参数，格式如下：
[
  {
    "tool": "tool_name",
    "params": {
      "param1": "value1",
      "param2": "value2"
    }
  }
]

如果不需要使用工具，返回空数组 []`, thinking, userInput)

	response, err := h.agent.llmClient.SimpleQuery(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"plan": response,
		"user_input": userInput,
	}, nil
}

// ToolHandler 工具处理器
type ToolHandler struct {
	agent *Agent
}

func (h *ToolHandler) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	planStr := input["plan"].(string)
	
	// 提取JSON部分
	planStr = extractJSON(planStr)

	var toolCalls []struct {
		Tool   string                 `json:"tool"`
		Params map[string]interface{} `json:"params"`
	}

	if err := json.Unmarshal([]byte(planStr), &toolCalls); err != nil {
		// 如果无法解析，可能不需要调用工具
		return map[string]interface{}{
			"results": []string{},
		}, nil
	}

	var results []string
	for _, call := range toolCalls {
		tool, err := h.agent.toolRegistry.Get(call.Tool)
		if err != nil {
			results = append(results, fmt.Sprintf("❌ 工具 %s 不存在: %v", call.Tool, err))
			continue
		}

		fmt.Printf("⚙️  执行工具: %s\n", call.Tool)
		result, err := tool.Execute(ctx, call.Params)
		if err != nil {
			results = append(results, fmt.Sprintf("❌ 工具 %s 执行失败: %v", call.Tool, err))
		} else {
			resultJSON, _ := json.MarshalIndent(result, "", "  ")
			results = append(results, fmt.Sprintf("✅ 工具 %s 执行成功:\n%s", call.Tool, string(resultJSON)))
		}
	}

	return map[string]interface{}{
		"results": results,
		"user_input": input["user_input"],
	}, nil
}

// SummaryHandler 总结处理器
type SummaryHandler struct {
	agent *Agent
}

func (h *SummaryHandler) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	results := input["results"].([]string)
	userInput := input["user_input"].(string)

	resultsStr := strings.Join(results, "\n\n")

	if len(results) == 0 {
		// 如果没有工具调用，直接回答
		response, err := h.agent.llmClient.SimpleQuery(ctx, userInput)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"result": response,
		}, nil
	}

	prompt := fmt.Sprintf(`基于以下工具执行结果，为用户生成一个友好的总结回复。

用户请求：%s

工具执行结果：
%s

请用自然语言总结执行结果，告诉用户任务是否完成以及具体的结果。`, userInput, resultsStr)

	response, err := h.agent.llmClient.SimpleQuery(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"result": response,
	}, nil
}

// extractJSON 从文本中提取JSON部分
func extractJSON(text string) string {
	// 查找 [ 或 { 开头的部分
	start := strings.Index(text, "[")
	if start == -1 {
		start = strings.Index(text, "{")
	}
	if start == -1 {
		return text
	}

	// 查找对应的结束符
	end := strings.LastIndex(text, "]")
	if end == -1 {
		end = strings.LastIndex(text, "}")
	}
	if end == -1 || end <= start {
		return text
	}

	return text[start : end+1]
}
