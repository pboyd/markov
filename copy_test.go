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
