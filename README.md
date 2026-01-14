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

## 📊 日志系统

### 日志位置
日志保存在: `~/.agentcli/logs/{date}/{sessionID}.log`

### 日志内容
- **USER_INPUT**: 用户输入记录
- **AGENT_OUTPUT**: Agent响应记录
- **THINKING**: 深度思考过程
- **TOOL_CALL**: 工具调用详情
- **INFO**: 一般信息
- **ERROR**: 错误信息

### 示例日志
```
[2026-01-13 17:30:32.123] [INFO] 会话开始 | Data: map[session_id:myuser_1736765432 timestamp:2026-01-13T17:30:32+08:00]
[2026-01-13 17:30:35.456] [USER_INPUT] 写一个Python Hello World程序
[2026-01-13 17:30:36.789] [THINKING] 开始处理 | Data: map[content:用户输入: 写一个Python Hello World程序]
[2026-01-13 17:30:38.012] [THINKING] 意图分析 | Data: map[content:用户想要创建一个Python程序...]
[2026-01-13 17:30:42.345] [AGENT_OUTPUT] 已经为您创建了一个简单的Python Hello World程序...
[2026-01-13 17:35:42.678] [INFO] 会话结束 | Data: map[session_id:myuser_1736765432 timestamp:2026-01-13T17:35:42+08:00]
```

## 🎨 定制化Agent

使用`/memory`命令设置Agent的角色和行为：

```bash
# Python专家
/memory 你是一个专业的Python开发专家，擅长Web开发和性能优化

# Go语言专家
/memory 你是一个资深的Go语言工程师，精通并发编程和微服务架构

# 数据科学家
/memory 你是一个数据科学专家，精通机器学习和数据分析

# 前端开发者
/memory 你是一个前端开发专家，精通React、Vue和现代化前端工具链

# DevOps工程师
/memory 你是一个DevOps工程师，擅长CI/CD、容器化和云原生技术
```

## 🔧 高级功能

### 1. 多用户支持
```bash
./agentcli --user alice interactive
./agentcli --user bob chat "问题"
```

### 2. 会话管理
```bash
# 指定会话ID
./agentcli --session my-session-001 interactive
```

### 3. 配置文件
```bash
# 使用自定义配置
./agentcli --config /path/to/config.yaml chat "问题"
```

## 📋 完整命令参考

### 全局标志
- `-c, --config`: 配置文件路径
- `-u, --user`: 用户ID
- `-s, --session`: 会话ID
- `-m, --model`: 指定模型
- `--memory`: 设置Agent定制化记忆
- `-h, --help`: 帮助信息

### 命令
- `version`: 显示版本信息
- `help`: 帮助信息

注意：默认启动即进入交互式模式，无需额外命令。

## 🎯 使用场景

### 1. 代码开发助手
```bash
👤: 写一个Go语言的HTTP服务器，支持CORS
👤: 优化这段代码的性能
👤: 添加单元测试
```

### 2. 系统管理
```bash
👤: 查看当前目录下的所有Go文件
👤: 统计代码行数
👤: 清理临时文件
```

### 3. 学习辅导
```bash
/memory 你是一个耐心的编程教师，善于用简单的语言解释复杂概念
👤: 解释什么是闭包
👤: 举个例子说明
```

### 4. 数据分析
```bash
/memory 你是一个数据分析专家
👤: 读取data.csv并分析数据分布
👤: 生成可视化图表
```

## 🔍 故障排查

### 1. API连接失败
- 检查配置文件中的`api_key`和`base_url`
- 确认网络连接正常
- 查看日志文件获取详细错误信息

### 2. 历史记录加载失败
- 确认`~/.agentcli/history/`目录存在
- 检查JSON文件格式是否正确
- 使用`/history`命令查看可用的对话ID

### 3. 工具调用失败
- 确认工具在配置文件中已启用
- 检查文件路径和权限
- 查看日志中的`TOOL_CALL`记录

## 📄 许可证

[添加您的许可证信息]

## 🤝 贡献

欢迎提交Issue和Pull Request！

## 📧 联系方式

[添加您的联系方式]

## 🔄 更新日志

### v2.0.0 (2026-01-13)
- ✅ 统一所有模式为流式输出
- ✅ 添加历史记录持久化
- ✅ 实现模型切换功能
- ✅ 添加/memory定制化命令
- ✅ 集成完整日志系统
- ✅ 移除simple模式
- ✅ 优化用户体验

### v1.0.0
- 初始版本
- 基础DAG思考引擎
- 工具调用支持
