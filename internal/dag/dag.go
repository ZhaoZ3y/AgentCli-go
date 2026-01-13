package dag

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DAG 有向无环图
type DAG struct {
	nodes       map[string]*Node
	maxDepth    int
	parallelNum int
	timeout     time.Duration
	verbose     bool
	mu          sync.RWMutex
}

// NewDAG 创建新的DAG
func NewDAG(maxDepth, parallelNum int, timeout time.Duration, verbose bool) *DAG {
	return &DAG{
		nodes:       make(map[string]*Node),
		maxDepth:    maxDepth,
		parallelNum: parallelNum,
		timeout:     timeout,
		verbose:     verbose,
	}
}

// AddNode 添加节点
func (d *DAG) AddNode(node *Node) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.nodes[node.ID]; exists {
		return fmt.Errorf("节点 %s 已存在", node.ID)
	}

	d.nodes[node.ID] = node
	return nil
}

// GetNode 获取节点
func (d *DAG) GetNode(id string) (*Node, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	node, ok := d.nodes[id]
	return node, ok
}

// Validate 验证DAG
func (d *DAG) Validate() error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// 检查依赖是否存在
	for _, node := range d.nodes {
		for _, depID := range node.Dependencies {
			if _, exists := d.nodes[depID]; !exists {
				return fmt.Errorf("节点 %s 依赖的节点 %s 不存在", node.ID, depID)
			}
		}
	}

	// 检查是否有循环依赖
	if err := d.detectCycle(); err != nil {
		return err
	}

	return nil
}

// detectCycle 检测循环依赖
func (d *DAG) detectCycle() error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for nodeID := range d.nodes {
		if !visited[nodeID] {
			if d.detectCycleUtil(nodeID, visited, recStack) {
				return fmt.Errorf("检测到循环依赖")
			}
		}
	}

	return nil
}

func (d *DAG) detectCycleUtil(nodeID string, visited, recStack map[string]bool) bool {
	visited[nodeID] = true
	recStack[nodeID] = true

	node := d.nodes[nodeID]
	for _, depID := range node.Dependencies {
		if !visited[depID] {
			if d.detectCycleUtil(depID, visited, recStack) {
				return true
			}
		} else if recStack[depID] {
			return true
		}
	}

	recStack[nodeID] = false
	return false
}

// Execute 执行DAG
func (d *DAG) Execute(ctx context.Context) error {
	// 验证DAG
	if err := d.Validate(); err != nil {
		return fmt.Errorf("DAG验证失败: %w", err)
	}

	// 创建超时上下文
	execCtx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	// 执行节点
	return d.executeNodes(execCtx)
}

// executeNodes 执行节点
func (d *DAG) executeNodes(ctx context.Context) error {
	d.mu.RLock()
	totalNodes := len(d.nodes)
	d.mu.RUnlock()

	completed := 0
	errChan := make(chan error, totalNodes)
	semaphore := make(chan struct{}, d.parallelNum)

	for completed < totalNodes {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 获取可执行节点
		executableNodes := d.getExecutableNodes()
		if len(executableNodes) == 0 {
			// 检查是否有失败的节点
			if d.hasFailedNodes() {
				return fmt.Errorf("存在失败的节点")
			}
			// 等待一段时间后重试
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// 并行执行可执行节点
		var wg sync.WaitGroup
		for _, node := range executableNodes {
			wg.Add(1)
			go func(n *Node) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				if d.verbose {
					fmt.Printf("[DAG] 执行节点: %s (%s)\n", n.Name, n.ID)
				}

				// 在执行前，将依赖节点的输出作为输入
				d.prepareDependencyOutputs(n)

				if err := n.Execute(ctx); err != nil {
					errChan <- err
				} else {
					if d.verbose {
						fmt.Printf("[DAG] 节点完成: %s (%s)\n", n.Name, n.ID)
					}
				}
			}(node)
		}

		wg.Wait()

		// 检查错误
		select {
		case err := <-errChan:
			return err
		default:
		}

		// 更新完成计数
		completed = d.getCompletedCount()
	}

	return nil
}

// getExecutableNodes 获取可执行节点
func (d *DAG) getExecutableNodes() []*Node {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var executable []*Node
	for _, node := range d.nodes {
		if node.CanExecute(d.nodes) {
			executable = append(executable, node)
		}
	}
	return executable
}

// hasFailedNodes 是否有失败的节点
func (d *DAG) hasFailedNodes() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	for _, node := range d.nodes {
		if node.IsFailed() {
			return true
		}
	}
	return false
}

// getCompletedCount 获取已完成节点数量
func (d *DAG) getCompletedCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	count := 0
	for _, node := range d.nodes {
		if node.IsCompleted() || node.IsFailed() {
			count++
		}
	}
	return count
}

// GetResults 获取所有节点结果
func (d *DAG) GetResults() map[string]map[string]interface{} {
	d.mu.RLock()
	defer d.mu.RUnlock()

	results := make(map[string]map[string]interface{})
	for id, node := range d.nodes {
		results[id] = node.Output
	}
	return results
}

// prepareDependencyOutputs 准备依赖节点的输出作为当前节点的输入
func (d *DAG) prepareDependencyOutputs(node *Node) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// 遍历所有依赖节点
	for _, depID := range node.Dependencies {
		if depNode, ok := d.nodes[depID]; ok {
			// 将依赖节点的输出合并到当前节点的输入
			for key, value := range depNode.Output {
				node.SetInput(key, value)
			}
		}
	}
}
