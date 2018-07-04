package markov

import "io"

var _ WriteChain = &DiskChainWriter{}

type DiskChainWriter struct {
	w io.WriteSeeker
}

func NewDiskChainWriter(w io.WriteSeeker) *DiskChainWriter {
	return &DiskChainWriter{
		w: w,
	}
}

func (c *DiskChainWriter) Add(interface{}) (id int, err error) {
	return 0, nil
}

func (c *DiskChainWriter) Relate(parent, child int, delta int) error {
	return nil
}
