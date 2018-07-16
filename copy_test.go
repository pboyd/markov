package markov

import "testing"

func TestCopy(t *testing.T) {
	src := &MemoryChain{}
	testWriteChain(t, src)

	dest := &MemoryChain{}
	err := Copy(dest, src)
	if err != nil {
		t.Fatalf("Copy failed with error: %v", err)
	}

	testReadChain(t, dest)
}

func BenchmarkDiskCopy(b *testing.B) {
	src := NewMemoryChain(b.N)
	err := Feed(src, normalDistGenerator(b.N, b.N*2))
	if err != nil {
		b.Fatalf("got error: %v", err)
	}

	f, cleanup := tempFile(b)
	defer cleanup()

	dest, err := NewDiskChainWriter(f)
	if err != nil {
		b.Fatalf("error: %v", err)
	}

	b.ResetTimer()

	err = Copy(dest, src)
	if err != nil {
		b.Fatalf("Copy failed with error: %v", err)
	}
}
