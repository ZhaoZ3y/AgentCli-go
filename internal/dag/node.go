package dag

import (
	"context"
	"fmt"
	"sync"
)

// NodeType 节点类型
type NodeType string

const (
	NodeTypeThink   NodeType = "think"   // 思考节点
	NodeTypeTool    NodeType = "tool"    // 工具节点
	NodeTypeDecision NodeType = "decision" // 决策节点
	NodeTypeEnd     NodeType = "end"     // 结束节点
)

// NodeStatus 节点状态
type NodeStatus string

const (
	NodeStatusPending   NodeStatus = "pending"   // 待处理
	NodeStatusRunning   NodeStatus = "running"   // 运行中
	NodeStatusCompleted NodeStatus = "completed" // 已完成
	NodeStatusFailed    NodeStatus = "failed"    // 失败
	NodeStatusSkipped   NodeStatus = "skipped"   // 跳过
)

// Node DAG节点
type Node struct {
	ID          string                 // 节点ID
	Type        NodeType               // 节点类型
	Name        string                 // 节点名称
	Description string                 // 节点描述
	Dependencies []string              // 依赖的节点ID列表
	Status      NodeStatus             // 节点状态
	Input       map[string]interface{} // 输入数据
	Output      map[string]interface{} // 输出数据
	Error       error                  // 错误信息
	Handler     NodeHandler            // 节点处理器
	mu          sync.RWMutex           // 互斥锁
}

// NodeHandler 节点处理器接口
type NodeHandler interface {
	Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)
}

// NewNode 创建新节点
func NewNode(id, name string, nodeType NodeType) *Node {
	return &Node{
		ID:           id,
		Type:         nodeType,
		Name:         name,
		Status:       NodeStatusPending,
		Dependencies: make([]string, 0),
		Input:        make(map[string]interface{}),
		Output:       make(map[string]interface{}),
	}
}

// SetHandler 设置节点处理器
func (n *Node) SetHandler(handler NodeHandler) {
	n.Handler = handler
}

// AddDependency 添加依赖
func (n *Node) AddDependency(nodeID string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Dependencies = append(n.Dependencies, nodeID)
}

// SetInput 设置输入
func (n *Node) SetInput(key string, value interface{}) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Input[key] = value
}

// GetOutput 获取输出
func (n *Node) GetOutput(key string) (interface{}, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	val, ok := n.Output[key]
	return val, ok
}

// Execute 执行节点
func (n *Node) Execute(ctx context.Context) error {
	n.mu.Lock()
	if n.Status != NodeStatusPending {
		n.mu.Unlock()
		return fmt.Errorf("节点 %s 状态不是待处理状态: %s", n.ID, n.Status)
	}
	n.Status = NodeStatusRunning
	
	// 复制input以便传递
	inputCopy := make(map[string]interface{})
	for k, v := range n.Input {
		inputCopy[k] = v
	}
	n.mu.Unlock()

	// 执行处理器
	if n.Handler != nil {
		output, err := n.Handler.Execute(ctx, inputCopy)
		n.mu.Lock()
		if err != nil {
			n.Status = NodeStatusFailed
			n.Error = err
			n.mu.Unlock()
			return fmt.Errorf("节点 %s 执行失败: %w", n.ID, err)
		}
		n.Output = output
		n.Status = NodeStatusCompleted
		n.mu.Unlock()
	} else {
		n.mu.Lock()
		n.Status = NodeStatusCompleted
		n.mu.Unlock()
	}

	return nil
}

// GetStatus 获取节点状态
func (n *Node) GetStatus() NodeStatus {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.Status
}

// IsCompleted 是否已完成
func (n *Node) IsCompleted() bool {
	return n.GetStatus() == NodeStatusCompleted
}

// IsFailed 是否失败
func (n *Node) IsFailed() bool {
	return n.GetStatus() == NodeStatusFailed
}

// CanExecute 是否可以执行
func (n *Node) CanExecute(nodes map[string]*Node) bool {
	n.mu.RLock()
	defer n.mu.RUnlock()

	// 检查所有依赖是否已完成
	for _, depID := range n.Dependencies {
		if depNode, ok := nodes[depID]; ok {
			if !depNode.IsCompleted() {
				return false
			}
		}
	}
	return n.Status == NodeStatusPending
}
