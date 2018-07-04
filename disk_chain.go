package markov

import "os"

var _ Chain = &DiskChain{}

type DiskChain struct {
	r     *os.File
	index map[interface{}]int64
}

func NewDiskChain(r *os.File) (*DiskChain, error) {
	return &DiskChain{
		r: r,
	}, nil
}

func (c *DiskChain) Get(id int) (interface{}, error) {
	return nil, ErrNotFound
}

func (c *DiskChain) Links(id int) ([]Link, error) {
	return nil, ErrNotFound
}

func (c *DiskChain) Find(value interface{}) (id int, err error) {
	return 0, ErrNotFound
}
