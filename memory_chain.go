package markov

import "math/rand"

type MemoryChain struct {
	nodes map[interface{}]*MemoryNode
}

func NewChain() *MemoryChain {
	return &MemoryChain{
		nodes: map[interface{}]*MemoryNode{},
	}
}

func (c *MemoryChain) Get(value interface{}) *MemoryNode {
	node, ok := c.nodes[value]
	if !ok {
		node = &MemoryNode{
			Value:    value,
			chain:    c,
			children: []*memoryChildNode{},
		}
		c.nodes[value] = node
	}
	return node
}

type MemoryNode struct {
	Value    interface{}
	chain    *MemoryChain
	children []*memoryChildNode
}

type memoryChildNode struct {
	*MemoryNode
	Count int
}

func (n *MemoryNode) findChild(value interface{}) *memoryChildNode {
	for _, child := range n.children {
		if child.Value == value {
			return child
		}
	}

	return nil
}

func (n *MemoryNode) Mark(value interface{}) *MemoryNode {
	child := n.chain.Get(value)

	cn := n.findChild(child.Value)
	if cn == nil {
		cn = &memoryChildNode{
			MemoryNode: child,
		}
		n.children = append(n.children, cn)
	}

	cn.Count++

	return child
}

func (n *MemoryNode) sum() int {
	total := 0
	for _, c := range n.children {
		total += c.Count
	}

	return total
}

func (n *MemoryNode) Probabilities() map[interface{}]float64 {
	total := float64(n.sum())
	p := make(map[interface{}]float64, len(n.children))

	for _, child := range n.children {
		p[child.Value] = float64(child.Count) / total
	}

	return p
}

func (n *MemoryNode) Next() *MemoryNode {
	if len(n.children) == 0 {
		// This happens if the chain ends
		return nil
	}

	index := rand.Intn(n.sum())
	passed := 0

	for _, child := range n.children {
		passed += child.Count
		if passed > index {
			return child.MemoryNode
		}
	}

	panic("Next() failed")
}
