package markov

import (
	"math/rand"
	"testing"
)

func BenchmarkBuildUnique(b *testing.B) {
	Feed(NewMemoryChain(b.N), normalDistGenerator(b.N, b.N*2))
}

func BenchmarkBuildDuplicate(b *testing.B) {
	Feed(NewMemoryChain(b.N), normalDistGenerator(b.N, 10))
}

func BenchmarkBuildUniqueDisk(b *testing.B) {
	f, cleanup := tempFile(b)
	defer cleanup()

	chain, err := NewDiskChainWriter(f)
	if err != nil {
		b.Fatalf("error: %v", err)
	}
	b.ResetTimer()

	Feed(chain, normalDistGenerator(b.N, b.N*2))
}

func BenchmarkBuildDuplicateDisk(b *testing.B) {
	f, cleanup := tempFile(b)
	defer cleanup()

	chain, err := NewDiskChainWriter(f)
	if err != nil {
		b.Fatalf("error: %v", err)
	}
	b.ResetTimer()

	Feed(chain, normalDistGenerator(b.N, 10))
}

func BenchmarkRandomWalk(b *testing.B) {
	chain := NewMemoryChain(b.N)
	Feed(chain, normalDistGenerator(b.N, b.N/4))
	walker := RandomWalker(chain, 0)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		walker.Next()
	}
}

func normalDistGenerator(count, stddev int) <-chan interface{} {
	numbers := make(chan interface{})

	go func() {
		defer close(numbers)

		for i := 0; i < count; i++ {
			numbers <- int(rand.NormFloat64() * float64(stddev))
		}
		numbers <- 0
	}()

	return numbers
}
