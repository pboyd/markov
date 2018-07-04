package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
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
	}
	rand.Seed(int64(seed))

	b := markov.NewBuilder(' ')
	runes, err := readRunes(source)
	if err != nil {
		fmt.Printf("file error (%s): %v\n", source, err)
		os.Exit(1)
	}

	log.Printf("seed=%d", seed)

	b.Feed(runes)
	node := b.Root()

	generated := 0
	for {
		node = node.Next()
		r := node.Value.(rune)
		fmt.Print(string(r))

		if r == ' ' {
			generated++
			if generated == wordCount {
				break
			}
		}
	}

	fmt.Print("\n")
}

func readRunes(path string) (<-chan interface{}, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(fh)

	runes := make(chan interface{})

	go func() {
		defer fh.Close()
		defer close(runes)

		for {
			r, _, err := reader.ReadRune()
			if err != nil {
				if err != io.EOF {
					log.Printf("error reading file: %v", err)
				}
				break
			}

			if unicode.IsLetter(r) || r == ' ' {
				runes <- unicode.ToLower(r)
			}
		}
	}()

	return runes, nil
}
