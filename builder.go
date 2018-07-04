package markov

type Builder struct {
	initial interface{}
	nodes   map[interface{}]*builderNode
}

func NewBuilder(initial interface{}) *Builder {
	return &Builder{
		initial: initial,
		nodes:   make(map[interface{}]*builderNode, 1),
	}
}

func (b *Builder) Build() []*Node {
	nodes := make(map[interface{}]*Node, len(b.nodes))

	for value, _ := range b.nodes {
		nodes[value] = &Node{
			Value: value,
		}
	}

	nodeSlice := make([]*Node, 0, len(b.nodes))

	for value, bn := range b.nodes {
		node := nodes[value]
		node.Children = make([]NodeProbability, 0, len(bn.counts))

		for nextValue, probability := range bn.Probabilities() {
			node.Children = append(node.Children, NodeProbability{
				Node:        nodes[nextValue],
				Probability: probability,
			})
		}

		nodeSlice = append(nodeSlice, node)
	}

	return nodeSlice
}

func (b *Builder) Feed(values <-chan interface{}) {
	last := b.getNode(b.initial)

	for val := range values {
		last.counts[val]++
		last = b.getNode(val)
	}
}

func (b *Builder) getNode(value interface{}) *builderNode {
	node, ok := b.nodes[value]
	if !ok {
		node = &builderNode{
			counts: make(map[interface{}]int, 1),
		}

		b.nodes[value] = node
	}

	return node
}

type builderNode struct {
	counts map[interface{}]int
}

func (bn *builderNode) Probabilities() map[interface{}]float64 {
	sum := 0
	for _, c := range bn.counts {
		sum += c
	}

	values := make(map[interface{}]float64, len(bn.counts))

	for value, count := range bn.counts {
		values[value] = float64(count) / float64(sum)
	}

	return values
}
