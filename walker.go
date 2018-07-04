package markov

// Walker walks through the chain.
type Walker interface {
	// Next returns the next value in the chain.
	//
	// If the chain has no further nodes, Next() returns ErrBrokenChain.
	Next() (Value, error)
}
