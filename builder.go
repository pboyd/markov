package markov

type Builder struct {
	initial interface{}
	Chain   *MemoryChain
}

func NewBuilder(initial interface{}) *Builder {
	return &Builder{
		initial: initial,
		Chain:   NewChain(),
	}
}

func (b *Builder) Root() *MemoryNode {
	return b.Chain.Get(b.initial)
}

func (b *Builder) Feed(values <-chan interface{}) {
	last := b.Chain.Get(b.initial)

	for val := range values {
		last = last.Mark(val)
	}
}
