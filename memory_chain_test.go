package markov

import "testing"

func TestMemoryChain(t *testing.T) {
	testReadWriteChain(t, &MemoryChain{})
}
