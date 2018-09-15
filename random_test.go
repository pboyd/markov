package markov

import "testing"

func TestRandom(t *testing.T) {
	chain := &MemoryChain{}
	testWriteChain(t, chain)

	value, err := Random(chain)
	if err != nil {
		t.Fatalf("got error %v, want nil", err)
	}

	id, err := chain.Find(value)
	if err != nil {
		t.Fatalf("got error %v, want nil", err)
	}

	if id < 0 {
		t.Errorf("got id %d, want >= 0", id)
	}
}

func BenchmarkMemoryRandom(b *testing.B) {
	chain := NewMemoryChain(b.N)
	Feed(chain, normalDistGenerator(10000, 100))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Random(chain)
	}
}

func BenchmarkDiskRandom(b *testing.B) {
	f, cleanup := tempFile(b)
	defer cleanup()

	chain, err := NewDiskChainWriter(f)
	if err != nil {
		b.Fatalf("error: %v", err)
	}
	Feed(chain, normalDistGenerator(10000, 100))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Random(chain)
	}
}
