package markov

import "math/rand"

var _ Walker = &randomWalker{}

type randomWalker struct {
	chain Chain
	last  int
}

func RandomWalker(chain Chain, startID int) Walker {
	return &randomWalker{
		chain: chain,
		last:  startID,
	}
}

func (w *randomWalker) Next() (Value, error) {
	links, err := w.chain.Links(w.last)
	if err != nil {
		return 0, err
	}

	if len(links) == 0 {
		return 0, ErrBrokenChain
	}

	index := rand.Float64()
	var passed float64

	for _, link := range links {
		passed += link.Probability
		if passed > index {
			w.last = link.ID
			return w.chain.Get(link.ID)
		}
	}

	panic("Next() failed")
}
