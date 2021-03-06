# Markov [![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://godoc.org/github.com/pboyd/markov)

`markov` is a Markov chain for Go. The package contains a in-memory chain and a
slower disk-based version, and functions to move between the memory and disk
versions.

## Example

This example generates a nonsense sentence from sample text.

```go
package main

import (
	"fmt"

	"github.com/pboyd/markov"
)

const text = `Listen, strange women lying in ponds distributing swords is no basis for a system of government. Supreme executive power derives from a mandate from the masses, not from some farcical aquatic ceremony.`

func main() {
	chain := markov.NewMemoryChain(0)

	// Feed each rune into the chain
	markov.Feed(chain, split(text))

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

        fmt.Print("\n")
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
```

For more in-depth examples see the `cmd/markov-ngram` and `cmd/markov-walk` programs.

# License

This package is released under the terms of the Apache 2.0 license. See LICENSE.TXT.
