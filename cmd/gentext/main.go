package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"unicode"

	"github.com/pboyd/markov"
)

var (
	source    string
	wordCount int
	seed      int
)

func init() {
	flag.StringVar(&source, "source", "", "path to the input text")
	flag.IntVar(&wordCount, "words", 100, "number of words to generate")
	flag.IntVar(&seed, "seed", 0, "random seed")
	flag.Parse()
}

func main() {
	if seed == 0 {
		seed = os.Getpid()
		fmt.Printf("seed=%d\n", seed)
	}
	rand.Seed(int64(seed))

	words, err := readFile(source)
	if err != nil {
		fmt.Fprintf(os.Stderr, "file error (%s): %v\n", source, err)
		os.Exit(1)
	}

	chain := &markov.MemoryChain{}
	markov.Feed(chain, words)

	l, err := chain.Len()
	if err != nil {
		fmt.Fprintf(os.Stderr, "len error: %v\n", err)
		os.Exit(2)
	}

	walker := markov.NewRandomWalker(chain, rand.Intn(l))

	for generated := 0; generated < wordCount; generated++ {
		word, err := walker.Next()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error generating word: %v\n", err)
			os.Exit(2)
		}

		fmt.Print(word.(string))
		fmt.Print(" ")
	}

	fmt.Print("\n")
}

func readFile(path string) (<-chan markov.Value, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(fh)

	words := make(chan markov.Value)

	go func() {
		defer func() {
			fh.Close()
			close(words)
		}()

		var word strings.Builder

		for {
			r, _, err := reader.ReadRune()
			if err != nil {
				if err != io.EOF {
					fmt.Fprintf(os.Stderr, "error reading file: %v", err)
				}
				break
			}

			if unicode.IsLetter(r) || strings.ContainsRune("'’", r) {
				word.WriteRune(r)
			} else {
				if word.Len() > 0 {
					words <- word.String()
					word.Reset()
				}
			}
		}
	}()

	return words, nil
}
