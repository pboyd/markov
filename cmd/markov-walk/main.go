// markov-walk reads a Markov chain from disk and outputs items proportionally.
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"

	"github.com/pboyd/markov"
)

var (
	source    string
	count     int
	seed      int
	delimeter string
	start     string
)

func init() {
	flag.StringVar(&source, "chain", "", "path to the chain file")
	flag.StringVar(&delimeter, "d", " ", "delimeter to insert between items")
	flag.IntVar(&count, "count", 100, "number of items to generate")
	flag.IntVar(&seed, "seed", 0, "random seed")
	flag.StringVar(&start, "start", "", "term to start with (currently only supports strings)")
	flag.Parse()
}

func main() {
	if seed == 0 {
		seed = os.Getpid()
		fmt.Fprintf(os.Stderr, "-seed=%d\n", seed)
	}
	rand.Seed(int64(seed))

	delimeter, err := strconv.Unquote("\"" + delimeter + "\"")
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid delimeter: %v", err)
		os.Exit(1)
	}

	fh, err := os.OpenFile(source, os.O_RDWR, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "file error (%s): %v\n", source, err)
		os.Exit(1)
	}

	chain, err := markov.ReadDiskChain(fh)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to read chain: %v\n", err)
		os.Exit(1)
	}

	startID := 0
	if start != "" {
		startID, err = chain.Find(start)
		if err != nil {
			fmt.Fprint(os.Stderr, "start item does not exist\n")
			os.Exit(1)
		}

		fmt.Print(start, delimeter)
	}

	walker := markov.RandomWalker(chain, startID)

	for generated := 0; generated < count; generated++ {
		word, err := walker.Next()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error generating item: %v\n", err)
			os.Exit(2)
		}

		fmt.Print(word, delimeter)
	}

	fmt.Print("\n")
}
