package markov

type Node struct {
	Value    interface{}
	children []*childNode
}

type childNode struct {
	*Node
	Count int
}

func NewNode(value interface{}) *Node {
	return &Node{
		Value:    value,
		children: []*childNode{},
	}
}

func (n *Node) findChild(value interface{}) *childNode {
	for _, child := range n.children {
		if child.Value == value {
			return child
		}
	}

	return nil
}

func (n *Node) Mark(child *Node) {
	cn := n.findChild(child.Value)
	if cn == nil {
		cn = &childNode{
			Node: child,
		}
		n.children = append(n.children, cn)
	}

	cn.Count++
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
