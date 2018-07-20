package markov_test

import (
	"fmt"
	"math/rand"

	"github.com/pboyd/markov"
)

const text = `Listen, strange women lying in ponds distributing swords is no basis for a system of government. Supreme executive power derives from a mandate from the masses, not from some farcical aquatic ceremony.`

// This example generates a nonsense sentence from sample text.
func Example() {
	chain := markov.NewMemoryChain(0)

	// Feed each rune into the chain
	markov.Feed(chain, split(text))

	// Reset the random seed for consistent output.
	rand.Seed(0)

	// Walk the chain with randomly but with weighted probabilities.
	walker := markov.RandomWalker(chain, 0)
	for {
		r, _ := walker.Next()
		fmt.Printf(string(r.(rune)))

		// Stop after the first period.
		if r == '.' {
			break
		}
	}

	// Output: tera po me maswofove ndsting ng syinonds ssweny.
}

func split(text string) chan interface{} {
	runes := make(chan interface{})
	go func() {
		defer close(runes)

		// Start at the beginning of a word
		runes <- ' '

		for _, r := range text {
			runes <- r
		}
	}()
	return runes
}
