package markov

import "io"

var _ Chain = &DiskChain{}

type DiskChain struct {
	r io.ReadSeeker
}

func NewDiskChain(r io.ReadSeeker) *DiskChain {
	return &DiskChain{
		r: r,
	}
}

func (c *DiskChain) Get(id int) (interface{}, error) {
	return nil, ErrNotFound
}

func (c *DiskChain) Links(id int) ([]Link, error) {
	return nil, ErrNotFound
}

func (c *DiskChain) Find(interface{}) (id int, err error) {
	return 0, ErrNotFound
}
