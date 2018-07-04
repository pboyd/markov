package markov

import (
	"math"
	"testing"
)

func TestRandomWalker(t *testing.T) {
	chain := &MemoryChain{}
	Feed(chain, split(testText))

	const iterations = 100000

	counts := map[interface{}]int{}

	for i := 0; i < iterations; i++ {
		walker := RandomWalker(chain, 0)
		value, err := walker.Next()
		if err != nil {
			t.Fatalf("got error: %v", err)
		}
		counts[value]++
	}

	links, err := chain.Links(0)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}

	for _, link := range links {
		value, err := chain.Get(link.ID)
		if err != nil {
			t.Fatalf("got error: %v", err)
		}

		actual := counts[value]
		xp := int(link.Probability * float64(iterations))
		if !fuzzyEquals(actual, xp, 0.1) {
			t.Errorf("%d: got %d, want ~%d", link.ID, actual, xp)
		}
	}
}

func TestRandomWalkerWalk(t *testing.T) {
	chain := &MemoryChain{}
	Feed(chain, split(testText))

	walker := RandomWalker(chain, 0)
	for i := 0; i < 100; i++ {
		value, err := walker.Next()
		if err != nil {
			t.Fatalf("got error: %v", err)
		}

		if _, ok := value.(rune); !ok {
			t.Errorf("got %T, want rune", value)
		}
	}
}

func fuzzyEquals(a, b int, tolerance float64) bool {
	return math.Abs((float64(a)/float64(b))-1) < tolerance
}
