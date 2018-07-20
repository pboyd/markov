package markov

import (
	"testing"
	"unicode"
)

func TestFeed(t *testing.T) {
	b := NewBuilder()
	runes := split(`“Well, Prince, so Genoa and Lucca are now just family estates of the Buonapartes. But I warn you, if you don’t tell me that this means war, if you still try to defend the infamies and horrors perpetrated by that Antichrist—I really believe he is Antichrist—I will have nothing more to do with you and you are no longer my friend, no longer my ‘faithful slave,’ as you call yourself! But how do you do? I see I have frightened you—sit down and tell me all the news.”`)

	b.Feed(runes)

	nodes := b.Build()

	if len(nodes) != 24 {
		t.Errorf("got %d nodes, want 24", len(nodes))
	}
}

func split(text string) <-chan interface{} {
	runes := make(chan interface{})

	go func() {
		defer close(runes)
		for _, r := range text {
			if unicode.IsLetter(r) || unicode.IsSpace(r) {
				runes <- unicode.ToLower(r)
			}
		}
	}()

	return runes
}
