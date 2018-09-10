package markov

import "errors"

var (
	// ErrNotFound is returned by Find or Links when an item doesn't exist.
	ErrNotFound error = errors.New("markov: not found")

	// ErrBrokenChain is returned when the chain ends.
	ErrBrokenChain error = errors.New("markov: broken chain")
)

// Chain is a read-only Markov chain.
type Chain interface {
	// Get returns a value by it's ID. Returns nil if the ID doesn't exist.
	Get(id int) (interface{}, error)

	// Links returns the items linked to the given item.
	//
	// Returns ErrNotFound if the ID doesn't exist.
	Links(id int) ([]Link, error)

	// Find returns the ID for the given value.
	//
	// Returns ErrNotFound if the value doesn't exist.
	Find(value interface{}) (id int, err error)
}

// WriteChain is a Markov chain that can only be written.
type WriteChain interface {
	// Add conditionally inserts a new value to the chain.
	//
	// If the value exists it's ID is returned.
	Add(value interface{}) (id int, err error)

	// Relate increases the number of times child occurs after parent.
	Relate(parent, child int, delta int) error
}

// ReadWriteChain is a chain that supports reading and writing.
type ReadWriteChain interface {
	Chain
	WriteChain
}

// Link describes a child item.
type Link struct {
	ID          int
	Probability float64
}
