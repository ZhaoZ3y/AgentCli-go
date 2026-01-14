package cmd

import (
	"agentcli/internal/agent"
	"agentcli/internal/config"
	"agentcli/internal/history"
	"agentcli/internal/logger"
	"bufio"
	"context"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	configFile   string
	chatModel    string
	sessionID    string
	cfg          *config.Config
	historyMgr   *history.Manager
	log          *logger.Logger
	userID       string
	memory       string // Agent定制化记忆
)

// rootCmd 根命令
var rootCmd = &cobra.Command{
	Use:   "agentcli",
	Short: "智能终端Agent - 基于DAG的深度思考助手",
	Long: `AgentCLI 是一个智能终端助手，使用DAG（有向无环图）进行深度思考，
支持多种工具调用，包括：
  - 写代码 (write_code)
  - 读取文件 (read_file)
  - 识别图片 (recognize_image)
  - 执行命令 (execute_command)

通过API Key连接大语言模型，智能理解用户意图并自动调用相应工具完成任务。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 默认启动交互式模式
		return runInteractive()
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// 加载配置
		var err error
		cfg, err = config.Load(configFile)
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		// 获取用户ID
		if userID == "" {
			currentUser, err := user.Current()
			if err == nil {
				userID = currentUser.Username
				// 处理 Windows 下的 DOMAIN\User 格式
				if idx := strings.LastIndex(userID, "\\"); idx >= 0 {
					userID = userID[idx+1:]
				}
			} else {
				userID = "default"
			}
		}

		// 初始化历史记录管理器（当前目录下）
		historyDir := "history"
		historyMgr = history.NewManager(historyDir)
		if err := historyMgr.Init(); err != nil {
			return fmt.Errorf("初始化历史记录失败: %w", err)
		}

		// 初始化日志记录器
		if sessionID == "" {
			sessionID = fmt.Sprintf("%s_%d", userID, time.Now().Unix())
		}
		log, err = logger.NewLogger(sessionID)
		if err != nil {
			return fmt.Errorf("初始化日志失败: %w", err)
		}

		// 加载持久化的memory（如果命令行没有指定）
		if memory == "" {
			loadedMemory, err := agent.LoadMemoryFromFile(userID)
			if err == nil && loadedMemory != "" {
				memory = loadedMemory
				fmt.Printf("📝 已加载定制化记忆: %s\n", memory)
			}
		}

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		// 关闭日志记录器
		if log != nil {
			log.Close()
		}
		return nil
	},
}

// Execute 执行命令
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "配置文件路径 (默认: ./configs/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&userID, "user", "u", "", "用户ID（用于历史记录）")
	rootCmd.PersistentFlags().StringVarP(&sessionID, "session", "s", "", "会话ID")
	rootCmd.PersistentFlags().StringVarP(&chatModel, "model", "m", "", "指定使用的模型")
	rootCmd.PersistentFlags().StringVarP(&memory, "memory", "", "", "Agent定制化记忆")
	
	// 添加子命令
	rootCmd.AddCommand(versionCmd)
}

// runInteractive 运行交互式模式
func runInteractive() error {
	model := cfg.API.Model
	if chatModel != "" {
		model = chatModel
	}
	
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("🤖 AgentCLI - 交互式模式\n")
	fmt.Printf("📦 模型: %s\n", model)
	fmt.Printf("👤 用户: %s\n", userID)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("提示:\n")
	fmt.Printf("  - 输入 'exit' 或 'quit' 退出\n")
	fmt.Printf("  - 输入 '/new' 开始新对话\n")
	fmt.Printf("  - 输入 '/model' 切换模型\n")
	fmt.Printf("  - 输入 '/history' 查看历史对话\n")
	fmt.Printf("  - 输入 '/load <id>' 加载历史对话\n")
	fmt.Printf("  - 输入 '/memory <text>' 设置Agent定制化记忆\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	
	// 创建新对话
	conv := history.NewConversation(userID, model)
	
	// 创建Agent
	a := agent.NewAgent(cfg, log)
	
	// 应用命令行指定的记忆
	if memory != "" {
		a.SetMemory(memory)
	}
	
	// 创建读取器
	reader := bufio.NewReader(os.Stdin)
	ctx := context.Background()
	
	for {
		fmt.Print("👤 你: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Error("读取输入失败", err, nil)
			return fmt.Errorf("读取输入失败: %w", err)
		}
		
		input = strings.TrimSpace(input)
		
		// 检查退出命令
		if input == "exit" || input == "quit" {
			// 保存对话
			if len(conv.Messages) > 0 {
				if err := historyMgr.SaveConversation(conv); err != nil {
					log.Error("保存对话失败", err, nil)
					fmt.Printf("⚠️  保存对话失败: %v\n", err)
				} else {
					fmt.Printf("✅ 对话已保存 (ID: %s)\n", conv.ID)
				}
			}
			fmt.Println("\n👋 再见!")
			break
		}
		
		if input == "" {
			continue
		}
		
		// 处理特殊命令
		if strings.HasPrefix(input, "/") {
			if handleCommand(input, &model, conv, historyMgr, a, log) {
				continue
			}
		}
		
		// 记录用户输入
		log.UserInput(input)
		conv.AddMessage("user", input)
		
		// 流式输出处理请求
		var fullResponse string
		response, err := a.ProcessRequestStream(ctx, input, func(chunk string) error {
			fmt.Print(chunk)
			fullResponse += chunk
			return nil
		})
		
		if err != nil {
			log.Error("处理请求失败", err, nil)
			fmt.Printf("\n❌ 错误: %v\n\n", err)
			continue
		}
		
		// 记录Agent输出
		log.AgentOutput(response)
		conv.AddMessage("assistant", response)
		
		fmt.Println("\n\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	}
	
	return nil
}

// interactiveCmd 交互式命令（流式输出）
var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "进入交互式对话模式（流式输出）",
	Long:  "进入交互式模式，可以持续与Agent对话，支持流式输出、历史记录、模型切换等",
	Aliases: []string{"i", "repl"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInteractive()
	},
}

// versionCmd 版本命令
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("AgentCLI v2.0.0")
		fmt.Println("基于DAG的智能终端助手 - 流式输出版本")
	},
}

// handleCommand 处理特殊命令
func handleCommand(input string, model *string, conv *history.Conversation, historyMgr *history.Manager, a *agent.Agent, log *logger.Logger) bool {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return false
	}

	cmd := parts[0]

	switch cmd {
	case "/new":
		// 保存当前对话
		if len(conv.Messages) > 0 {
			if err := historyMgr.SaveConversation(conv); err != nil {
				log.Error("保存对话失败", err, nil)
				fmt.Printf("⚠️  保存对话失败: %v\n", err)
			} else {
				fmt.Printf("✅ 对话已保存 (ID: %s)\n", conv.ID)
			}
		}
		// 创建新对话
		*conv = *history.NewConversation(conv.UserID, *model)
		fmt.Println("🆕 开始新对话")
		log.Info("开始新对话", map[string]interface{}{"conversation_id": conv.ID})
		return true

	case "/model":
		availableModels := []string{
			"gpt-4",
			"gpt-5.2",
			"o4-mini",
			"o3",
			"o3-pro",
			"sora_image",
			"sora-2-pro",
			"claude-opus-4-5-20251101-thinking",
			"claude-sonnet-4-5-20250929",
			"claude-sonnet-4-5-20250929-thinking",
			"gemini-3-pro-preview-thinking",
			"gemini-3-pro-preview",
			"gemini-3-pro-all",
			"gemini-3-pro-image-preview",
			"qwen-plus",
		}
	
		fmt.Println("\n📦 可用模型列表:")
		for i, m := range availableModels {
			marker := " "
			if m == *model {
				marker = "✓"
			}
			fmt.Printf("  [%s] %d. %s\n", marker, i+1, m)
		}
		fmt.Printf("\n当前模型: %s\n", *model)
		fmt.Print("请输入模型编号或名称 (回车保持当前): ")
	
		reader := bufio.NewReader(os.Stdin)
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)
	
		if choice == "" {
			fmt.Println("保持当前模型")
			return true
		}
	
		var selectedModel string
	
		// 1) 先尝试按“编号”解析（支持 >9）
		if idx, err := strconv.Atoi(choice); err == nil {
			idx-- // 变成 0-based
			if idx >= 0 && idx < len(availableModels) {
				selectedModel = availableModels[idx]
			} else {
				fmt.Printf("❌ 无效编号: %d (范围: 1-%d)\n", idx+1, len(availableModels))
				return true
			}
		} else {
			// 2) 再按“名称”匹配（可选：也可以做不区分大小写）
			selectedModel = choice
		}
	
		// 可选：验证名称是否在列表中，避免输入不存在的模型
		found := false
		for _, m := range availableModels {
			if m == selectedModel {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("❌ 未知模型名称: %s\n", selectedModel)
			return true
		}
	
		*model = selectedModel
		conv.Model = selectedModel
		cfg.API.Model = selectedModel
		a.UpdateModel(selectedModel)
		fmt.Printf("✅ 已切换到模型: %s\n", selectedModel)
		log.Info("切换模型", map[string]interface{}{"model": selectedModel})
		return true

	case "/history":
		conversations, err := historyMgr.ListConversations(conv.UserID)
		if err != nil {
			log.Error("获取历史记录失败", err, nil)
			fmt.Printf("❌ 获取历史记录失败: %v\n", err)
			return true
		}
		if len(conversations) == 0 {
			fmt.Println("📭 没有历史对话记录")
			return true
		}
		fmt.Println("\n📜 历史对话:")
		for i, c := range conversations {
			fmt.Printf("  %d. ID: %s | 模型: %s | 消息数: %d | 更新: %s\n",
				i+1, c.ID, c.Model, len(c.Messages), c.Updated.Format("2006-01-02 15:04"))
		}
		fmt.Println()
		return true

	case "/load":
		if len(parts) < 2 {
			fmt.Println("用法: /load <对话ID>")
			return true
		}
		convID := parts[1]
		loadedConv, err := historyMgr.LoadConversation(convID)
		if err != nil {
			log.Error("加载对话失败", err, map[string]interface{}{"conversation_id": convID})
			fmt.Printf("❌ 加载对话失败: %v\n", err)
			return true
		}
		
		// 保存当前对话
		if len(conv.Messages) > 0 {
			historyMgr.SaveConversation(conv)
		}
		
		*conv = *loadedConv
		*model = conv.Model
		cfg.API.Model = conv.Model
		a.UpdateModel(conv.Model)
		
		fmt.Printf("✅ 已加载对话 (ID: %s, 消息数: %d)\n", conv.ID, len(conv.Messages))
		log.Info("加载历史对话", map[string]interface{}{
			"conversation_id": conv.ID,
			"message_count": len(conv.Messages),
		})
		
		// 显示最近几条消息
		recent := conv.GetRecentMessages(6)
		if len(recent) > 0 {
			fmt.Println("\n📝 最近的对话记录:")
			for _, msg := range recent {
				role := "👤"
				if msg.Role == "assistant" {
					role = "🤖"
				}
				content := msg.Content
				if len(content) > 100 {
					content = content[:100] + "..."
				}
				fmt.Printf("  %s: %s\n", role, content)
			}
			fmt.Println()
		}
		return true

	case "/memory":
		if len(parts) < 2 {
			if memory == "" {
				fmt.Println("📝 当前没有设置定制化记忆")
			} else {
				fmt.Printf("📝 当前定制化记忆: %s\n", memory)
			}
			fmt.Println("用法: /memory <定制化文本>")
			fmt.Println("例如: /memory 你是一个专业的Go语言开发专家，擅长性能优化")
			return true
		}
		
		memory = strings.Join(parts[1:], " ")
		a.SetMemory(memory)
		
		// 保存memory到文件
		if err := agent.SaveMemoryToFile(userID, memory); err != nil {
			log.Error("保存记忆失败", err, nil)
			fmt.Printf("⚠️  保存记忆失败: %v\n", err)
		} else {
			fmt.Printf("✅ 已设置并保存定制化记忆: %s\n", memory)
			log.Info("设置定制化记忆", map[string]interface{}{"memory": memory})
		}
		return true

	default:
		return false
	}
}
