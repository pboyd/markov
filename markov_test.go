package markov

import (
	"math"
	"testing"
	"unicode"
)

const testText = `“Well, Prince, so Genoa and Lucca are now just family estates of the Buonapartes. But I warn you, if you don’t tell me that this means war, if you still try to defend the infamies and horrors perpetrated by that Antichrist—I really believe he is Antichrist—I will have nothing more to do with you and you are no longer my friend, no longer my ‘faithful slave,’ as you call yourself! But how do you do? I see I have frightened you—sit down and tell me all the news.”`

func testReadWriteChain(t *testing.T, chain ReadWriteChain) {
	Feed(chain, split(testText))

	spaceID, err := chain.Find(' ')
	if err != nil {
		t.Fatalf("got error: %v", err)
	}

	actual, err := chain.Get(spaceID)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}

	if actual != ' ' {
		t.Errorf(`got %q, want " "`, string(actual.(rune)))
	}

	p, err := chain.Next(spaceID)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}

	aID, err := chain.Find('a')
	if err != nil {
		t.Fatalf("got error: %v", err)
	}

	for _, l := range p {
		if l.ID == aID {
			// 88 words in the paragraph. 10 of which start with "a".
			expectedPA := float64(10 / 88)
			if math.Abs(l.Probability-expectedPA) < 0.001 {
				t.Errorf("got %0.02f, want %0.02f", l.Probability, expectedPA)
			}

			break
		}
	}
}

func split(text string) <-chan Value {
	runes := make(chan Value)

	go func() {
		defer close(runes)

		// start at the beginning of a word
		runes <- ' '

		for _, r := range text {
			if unicode.IsLetter(r) || unicode.IsSpace(r) {
				runes <- unicode.ToLower(r)
			}
		}
	}()

	return runes
}
