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

func BenchmarkMemoryChainIterator(b *testing.B) {
	chain := &MemoryChain{}
	Feed(chain, split(testText))
	b.ResetTimer()
	benchmarkIterativeWalker(b, chain)
}

func BenchmarkDiskChainIterator(b *testing.B) {
	f, cleanup := tempFile(b)
	defer cleanup()

	chain, err := NewDiskChainWriter(f)
	if err != nil {
		b.Fatalf("error: %v", err)
	}

	Feed(chain, split(testText))
	b.ResetTimer()

	benchmarkIterativeWalker(b, chain)
}

func benchmarkIterativeWalker(b *testing.B, chain Chain) {
	for i := 0; i < b.N; i++ {
		walker := IterativeWalker(chain)

		for {
			_, err := walker.Next()
			if err != nil {
				if err == ErrBrokenChain {
					break
				}
				b.Fatalf("got error: %v", err)
			}

		}
	}
}
