package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"sync"
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

	letters, lengths, err := readFile(source)
	if err != nil {
		fmt.Printf("file error (%s): %v\n", source, err)
		os.Exit(1)
	}

	log.Printf("seed=%d", seed)

	var letterNode, lengthNode *markov.Node

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		b := markov.NewBuilder(' ')
		b.Feed(letters)
		letterNode = b.Root()
		wg.Done()
	}()

	go func() {
		b := markov.NewBuilder(0)
		b.Feed(lengths)
		lengthNode = b.Root()
		wg.Done()
	}()

	wg.Wait()

	for generated := 0; generated < wordCount; generated++ {
		lengthNode = lengthNode.Next()
		length := lengthNode.Value.(int)

		for i := 0; i < length; i++ {
			letterNode = letterNode.Next()
			r := letterNode.Value.(rune)
			fmt.Print(string(r))
		}

		fmt.Print(" ")
	}

	fmt.Print("\n")
}

func readFile(path string) (<-chan interface{}, <-chan interface{}, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}

	reader := bufio.NewReader(fh)

	letters := make(chan interface{})
	lengths := make(chan interface{})

	go func() {
		defer func() {
			fh.Close()
			close(letters)
			close(lengths)
		}()

		wordLength := 0

		for {
			r, _, err := reader.ReadRune()
			if err != nil {
				if err != io.EOF {
					log.Printf("error reading file: %v", err)
				}
				break
			}

			if unicode.IsSpace(r) {
				if wordLength > 0 {
					lengths <- wordLength
					wordLength = 0
				}
			} else {
				wordLength++
			}

			if unicode.IsLetter(r) {
				letters <- unicode.ToLower(r)
			}
		}
	}()

	return letters, lengths, nil
}
