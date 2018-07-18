// markov-feeder builds a Markov chain by reading lines from STDIN.
//
// As an example:
//
//	< /path/to/a/text/file tr ' ' '\n' | sed 's/[^[:alpha:]]//g' | markov-feeder -chain out.mkv

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/pboyd/markov"
)

var (
	output string
	update bool
	onDisk bool
)

func init() {
	flag.StringVar(&output, "chain", "", "path the the output chain file")
	flag.BoolVar(&update, "update", false, "update the output file instead of overwriting it")
	flag.BoolVar(&onDisk, "disk", false, "write the chain directly to disk")
	flag.Parse()
}

func main() {
	if output == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	entries := readLines(os.Stdin)

	diskChain, err := openOutputFile(output, update)
	if err != nil {
		fmt.Fprintf(os.Stderr, "file error (%s): %v\n", output, err)
		os.Exit(1)
	}

	if onDisk {
		err := markov.Feed(diskChain, entries)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error building chain: %v\n", err)
			os.Exit(2)
		}
	} else {
		memoryChain := &markov.MemoryChain{}
		err := markov.Feed(memoryChain, entries)
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

func readLines(r io.Reader) <-chan interface{} {
	lines := make(chan interface{})

	go func() {
		defer close(lines)
		br := bufio.NewReader(r)

		for {
			line, err := br.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					fmt.Fprintf(os.Stderr, "read error: %v", err)
				}
				break
			}

			lines <- line[:len(line)-1]
		}
	}()

	return lines
}
