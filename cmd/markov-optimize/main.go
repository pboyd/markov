// markov-optimize optimizes a chain file for reading.
//
// Links between items are stored in one or more buckets. The buckets
// typically contain unused space where new links are added. markov-optimize
// copies each entry's links into a single, correctly-sized, bucket. This
// reduces the file size and I/O operations on subsequent reads.
//
// This comes at the expense of future writes. A new link added to an
// "optimized" chain will go into new bucket, and therefore be slower.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pboyd/markov"
)

var (
	input  string
	output string
)

func init() {
	flag.StringVar(&input, "in", "", "path the the input chain file")
	flag.StringVar(&output, "out", "", "path the the output chain file")
	flag.Parse()
}

func main() {
	if output == "" || input == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	inFile, err := os.Open(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "file error (%s): %v\n", input, err)
		os.Exit(2)
	}
	defer inFile.Close()

	inChain, err := markov.ReadDiskChain(inFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading input %s: %v\n", input, err)
		os.Exit(2)
	}

	outFile, err := os.Create(output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "file error (%s): %v\n", output, err)
		os.Exit(2)
	}
	defer outFile.Close()

	outChain, err := markov.NewDiskChainWriter(outFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating output %s: %v\n", output, err)
		os.Exit(2)
	}

	err = markov.Copy(outChain, inChain)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error copying chain: %v\n", err)
		os.Exit(2)
	}
}
