package markov

import "math/rand"

type Chain struct {
	nodes map[interface{}]*Node
}

func NewChain() *Chain {
	return &Chain{
		nodes: map[interface{}]*Node{},
	}
}

func (c *Chain) GetNode(value interface{}) *Node {
	node, ok := c.nodes[value]
	if !ok {
		node = &Node{
			Value:    value,
			chain:    c,
			children: []*childNode{},
		}
		c.nodes[value] = node
	}
	return node
}

type Node struct {
	Value    interface{}
	chain    *Chain
	children []*childNode
}

type childNode struct {
	*Node
	Count int
}

func (n *Node) findChild(value interface{}) *childNode {
	for _, child := range n.children {
		if child.Value == value {
			return child
		}
	}

	return nil
}

func (n *Node) Mark(value interface{}) *Node {
	child := n.chain.GetNode(value)

	cn := n.findChild(child.Value)
	if cn == nil {
		cn = &childNode{
			Node: child,
		}
		n.children = append(n.children, cn)
	}

	cn.Count++

	return child
}

func (n *Node) sum() int {
	total := 0
	for _, c := range n.children {
		total += c.Count
	}

	return total
}

func (n *Node) Probabilities() map[interface{}]float64 {
	total := float64(n.sum())
	p := make(map[interface{}]float64, len(n.children))

	for _, child := range n.children {
		p[child.Value] = float64(child.Count) / total
	}

	return p
}

func (n *Node) Next() *Node {
	index := rand.Intn(n.sum())
	passed := 0

	for _, child := range n.children {
		passed += child.Count
		if passed > index {
			return child.Node
		}
	}

	panic("Next() failed")
}
