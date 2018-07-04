package markov

import (
	"math"
	"testing"
	"unicode"
)

func TestFeed(t *testing.T) {
	b := NewBuilder(' ')
	runes := split(`“Well, Prince, so Genoa and Lucca are now just family estates of the Buonapartes. But I warn you, if you don’t tell me that this means war, if you still try to defend the infamies and horrors perpetrated by that Antichrist—I really believe he is Antichrist—I will have nothing more to do with you and you are no longer my friend, no longer my ‘faithful slave,’ as you call yourself! But how do you do? I see I have frightened you—sit down and tell me all the news.”`)

	b.Feed(runes)

	root := b.Root()
	if root.Value != ' ' {
		t.Errorf(`got root value %q, want " "`, root.Value)
	}

	// 88 words and 10 start with "A" or "a"
	expectedPA := float64(10 / 88)
	p := root.Probabilities()
	if math.Abs(p['a']-expectedPA) < 0.001 {
		t.Errorf("got %0.02f, want %0.02f", p['a'], expectedPA)
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

/*
func describe(n *Node) string {
	out := &strings.Builder{}
	fmt.Fprintln(out, string(n.Value.(rune)))
	for _, np := range n.Children {
		fmt.Fprintf(out, "    %v: %0.2f\n", string(np.Value.(rune)), np.Probability)
	}

	return out.String()
}
*/
