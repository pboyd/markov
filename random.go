package markov

import "math/rand"

// RandomChain is a chain that can return a random value.
type RandomChain interface {
	Random() (interface{}, error)
}

// Random pseudo-randomly picks a value from the chain.
//
// If the chain implements RandomChain, it's Random function will be used.
func Random(chain Chain) (interface{}, error) {
	if rc, ok := chain.(RandomChain); ok {
		return rc.Random()
	}

	limit, err := chainLen(chain)
	if err != nil {
		return nil, err
	}

	iw := IterativeWalker(chain)

	n := rand.Intn(limit)
	for i := 0; i < n-1; i++ {
		_, err := iw.Next()
		if err != nil {
			return nil, err
		}
	}

	return iw.Next()
}

func chainLen(chain Chain) (int, error) {
	l := 0
	iw := IterativeWalker(chain)

	for {
		_, err := iw.Next()
		if err == ErrBrokenChain {
			break
		}
		if err != nil {
			return 0, err
		}

		l++
	}

	return l, nil
}
