// markov-ngram builds a Markov chain from a text file by splitting it into
// N-grams (words, bigrams, trigrams, etc.)
//
// It is included as an example, not as a useful program.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	"github.com/pboyd/markov"
)

var (
	output string
	update bool
	onDisk bool
	n      int
)

func init() {
	flag.StringVar(&output, "chain", "", "path the the output chain file")
	flag.BoolVar(&update, "update", false, "update the output file instead of overwriting it")
	flag.BoolVar(&onDisk, "disk", false, "write the chain directly to disk")
	flag.IntVar(&n, "n", 1, "ngram size")
	flag.Parse()
}

func main() {
	if output == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	sources := flag.Args()
	if len(sources) == 0 {
		fmt.Fprintf(os.Stderr, "usage: %s [flags] source [source]...\n", os.Args[0])
		os.Exit(1)
	}

	ngrams := make([]<-chan interface{}, len(sources))

	for i, source := range sources {
		var err error
		ngrams[i], err = readFileByWord(source)
		if err != nil {
			fmt.Fprintf(os.Stderr, "file error (%s): %v\n", source, err)
			os.Exit(1)
		}

		if n > 1 {
			ngrams[i] = joinWords(ngrams[i], n)
		}
	}

	diskChain, err := openOutputFile(output, update)
	if err != nil {
		fmt.Fprintf(os.Stderr, "file error (%s): %v\n", output, err)
		os.Exit(1)
	}

	if onDisk {
		err := markov.Feed(diskChain, ngrams...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error building chain: %v\n", err)
			os.Exit(2)
		}
	} else {
		memoryChain := &markov.MemoryChain{}
		err := markov.Feed(memoryChain, ngrams...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error building chain: %v\n", err)
			os.Exit(2)
		}

		err = markov.Copy(diskChain, memoryChain)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error copying chain to disk: %v\n", err)
			os.Exit(2)
		}
	}
}

func openOutputFile(path string, update bool) (markov.WriteChain, error) {
	if update {
		exists, err := fileExists(path)
		if err != nil {
			return nil, err
		}

		if !exists {
			update = false
		}
	}

	fh, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	if update {
		return markov.OpenDiskChainWriter(fh)
	}

	return markov.NewDiskChainWriter(fh)
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	// Some other error, probably an invalid path.
	return false, err
}

func readFileByWord(path string) (<-chan interface{}, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(fh)

	words := make(chan interface{})

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

func joinWords(words <-chan interface{}, n int) <-chan interface{} {
	ngrams := make(chan interface{})

	go func() {
		defer close(ngrams)

		ngram := make([]string, 0, n)

		for word := range words {
			sword := word.(string)

			if len(ngram) < n {
				ngram = append(ngram, sword)

				if len(ngram) < n {
					continue
				}
			} else {
				copy(ngram[0:], ngram[1:])
				ngram[n-1] = sword
			}
			ngrams <- strings.Join(ngram, " ")
		}
	}()
	return ngrams
}
