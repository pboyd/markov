package markov

import "testing"

func TestIterativeWalker(t *testing.T) {
	chain := &MemoryChain{}
	Feed(chain, split(testText))

	count := 0
	maxID := 0
	walker := IterativeWalker(chain)

	for {
		value, err := walker.Next()
		if err != nil {
			if err == ErrBrokenChain {
				break
			}
			t.Fatalf("got error: %v", err)
		}

		id, err := chain.Find(value)
		if err != nil {
			t.Fatalf("got error: %v", err)
		}

		if id > maxID {
			maxID = id
		}
		count++
	}

	// MemoryChain ID's array indices, so maxID and count are related.
	if count != maxID+1 {
		t.Errorf("got %d, want %d", count, maxID+1)
	}
}
