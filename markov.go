package markov

import "errors"

var (
	ErrNotFound    error = errors.New("markov: not found")
	ErrBrokenChain error = errors.New("markov: broken chain")
)

// Chain is a read-only Markov chain.
type Chain interface {
	Get(id int) (Value, error)
	Next(id int) ([]Link, error)
	Find(Value) (id int, err error)
	Len() (int, error)
}

// WriteChain is a Markov chain that can only be written.
type WriteChain interface {
	Add(Value) (id int, err error)
	Relate(parent, child int, delta int) error
}

// ReadWriteChain is a chain that supports reading and writing.
type ReadWriteChain interface {
	Chain
	WriteChain
}

type Value interface{}

type Link struct {
	ID          int
	Probability float64
}
