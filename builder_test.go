package markov

import (
	"testing"
	"unicode"
)

func TestFeedErrors(t *testing.T) {
	f, cleanup := tempFile(t)

	chain, err := NewDiskChainWriter(f)
	if err != nil {
		t.Fatalf("NewDiskChainWriter failed: %v", err)
	}

	err = Feed(chain, splitAbort(cleanup, testText))
	if err == nil {
		t.Errorf("got nil, want an error")
	}
}

func splitAbort(abort func(), text string) <-chan interface{} {
	runes := make(chan interface{})

	go func() {
		defer close(runes)

		for i, r := range text {
			if i >= 10 {
				abort()
			}

			if unicode.IsLetter(r) || unicode.IsSpace(r) {
				runes <- unicode.ToLower(r)
			}
		}
	}()

	return runes
}
