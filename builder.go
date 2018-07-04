package markov

type Builder struct {
	initial interface{}
	Chain   *Chain
}

func NewBuilder(initial interface{}) *Builder {
	return &Builder{
		initial: initial,
		Chain:   NewChain(),
	}
}

func (b *Builder) Root() *Node {
	return b.Chain.GetNode(b.initial)
}

func (b *Builder) Feed(values <-chan interface{}) {
	last := b.Chain.GetNode(b.initial)

	for val := range values {
		last = last.Mark(val)
	}
}
