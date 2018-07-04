package markov

type Builder struct {
	initial interface{}
	nodes   map[interface{}]*Node
}

func NewBuilder(initial interface{}) *Builder {
	return &Builder{
		initial: initial,
		nodes:   make(map[interface{}]*Node, 1),
	}
}

func (b *Builder) Root() *Node {
	return b.nodes[b.initial]
}

func (b *Builder) Feed(values <-chan interface{}) {
	last := b.getNode(b.initial)

	for val := range values {
		child := b.getNode(val)
		last.Mark(child)
		last = child
	}
}

func (b *Builder) getNode(value interface{}) *Node {
	node, ok := b.nodes[value]
	if !ok {
		node = NewNode(value)

		b.nodes[value] = node
	}

	return node
}
