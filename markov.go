package markov

import "errors"

var (
	ErrNotFound    error = errors.New("markov: not found")
	ErrBrokenChain error = errors.New("markov: broken chain")
)

// Chain is a read-only Markov chain.
type Chain interface {
	Get(id int) (interface{}, error)
	Links(id int) ([]Link, error)
	Find(interface{}) (id int, err error)
}

// WriteChain is a Markov chain that can only be written.
type WriteChain interface {
	Add(interface{}) (id int, err error)
	Relate(parent, child int, delta int) error
}

// ReadWriteChain is a chain that supports reading and writing.
type ReadWriteChain interface {
	Chain
	WriteChain
}

type Link struct {
	ID          int
	Probability float64
}
