# AgentCLI v2.0.0 - 智能终端助手

基于DAG（有向无环图）的智能终端助手，支持深度思考、工具调用、流式输出、历史记录管理等功能。

## ✨ 主要特性

### 🚀 核心功能
- **流式输出**: 所有对话（chat/interactive）都支持实时流式响应
- **历史记录**: 自动保存会话历史，支持加载和继续之前的对话
- **模型切换**: 交互式选择和切换多种AI模型
- **定制化记忆**: 通过/memory命令为Agent设置个性化角色和行为
- **完整日志**: 记录所有操作，包括用户输入、Agent输出、深度思考过程

### 🛠️ 工具支持
- **write_code**: 写入代码到文件
- **read_file**: 读取文件内容
- **recognize_image**: 识别图片内容
- **execute_command**: 执行系统命令

### 🧠 DAG深度思考引擎
- 意图分析
- 深度思考规划
- 工具调用决策
- 结果总结

## 📦 安装

```bash
# 克隆仓库
git clone <your-repo-url>
cd agentCli_by_go

# 编译
go build -o agentcli.exe

# 或使用go install
go install
```

## ⚙️ 配置

编辑 `configs/config.yaml`:

```yaml
api:
  openai_key: "your-api-key"
  base_url: "https://api.openai.com/v1"
  model: "gpt-4"
  timeout: 120

tools:
  enabled:
    - write_code
    - read_file
    - recognize_image
    - execute_command

dag:
  max_depth: 5
  parallel_nodes: 3
  timeout: 300
  verbose: true
```

## 🎯 使用方法

### 交互式模式（默认）

```bash
# 直接启动（默认进入交互式模式）
./agentcli

# 指定模型
./agentcli -m gpt-3.5-turbo

# 指定用户ID
./agentcli --user myname

# 设置定制化记忆
./agentcli --memory "你是一个Go语言专家"
```

**特点**:
- 默认启动即进入交互式模式
- 流式输出响应
- 自动保存对话历史
- 完整日志记录
- 支持深度思考和工具调用
- 智能分析用户意图（使用"thinking"显示思考过程）
- 自动分析项目代码文件
- 自动识别和分析图片
```

**交互式命令**:

| 命令 | 说明 | 示例 |
|------|------|------|
| `/new` | 开始新对话 | `/new` |
| `/model` | 切换模型 | `/model` |
| `/history` | 查看历史对话列表 | `/history` |
| `/load <id>` | 加载历史对话 | `/load default_1736765432` |
| `/memory <text>` | 设置Agent定制化记忆 | `/memory 你是一个Go语言专家` |
| `exit` 或 `quit` | 退出 | `quit` |

**示例会话**:

```
👤 你: 你好

🤖 Agent: 你好！有什么我可以帮助你的吗？

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
👤 你: /memory 你是一个专业的Python开发专家，擅长数据科学

✅ 已设置定制化记忆: 你是一个专业的Python开发专家，擅长数据科学

👤 你: 帮我写一个读取CSV文件的Python脚本

🤖 Agent: [流式输出Python代码和解释...]

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
👤 你: /model

📦 可用模型列表:
  [ ] 1. gpt-4
  [ ] 2. gpt-4-turbo
  [✓] 3. gpt-3.5-turbo
  [ ] 4. claude-3-opus
  [ ] 5. claude-3-sonnet
  [ ] 6. deepseek-chat
  [ ] 7. qwen-plus

当前模型: gpt-3.5-turbo
请输入模型编号或名称 (回车保持当前): 1

✅ 已切换到模型: gpt-4

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
👤 你: quit

✅ 对话已保存 (ID: myuser_1736765432)

👋 再见!
```

## 📝 历史记录管理

### 自动保存
- 所有对话自动保存到 `~/.agentcli/history/`
- 每个对话有唯一ID: `{userID}_{timestamp}`
- JSON格式存储，包含完整消息历史

### 加载历史
```bash
# 在interactive模式中
/history                    # 查看所有历史对话
/load default_1736765432    # 加载指定对话
```

### 历史文件结构
```json
{
  "id": "myuser_1736765432",
  "user_id": "myuser",
  "model": "gpt-4",
  "messages": [
    {
      "role": "user",
      "content": "你好",
      "timestamp": "2026-01-13T17:30:32+08:00"
    },
    {
      "role": "assistant",
      "content": "你好！...",
      "timestamp": "2026-01-13T17:30:35+08:00"
    }
  ],
  "created": "2026-01-13T17:30:32+08:00",
  "updated": "2026-01-13T17:35:42+08:00"
}
```
