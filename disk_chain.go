package markov

import "os"

var _ Chain = &DiskChain{}

type DiskChain struct {
	w *DiskChainWriter
}

func ReadDiskChain(r *os.File) (*DiskChain, error) {
	w, err := OpenDiskChainWriter(r)
	if err != nil {
		return nil, err
	}

	return &DiskChain{
		w: w,
	}, nil
}

func (c *DiskChain) Get(id int) (interface{}, error) {
	return c.w.Get(id)
}

func (c *DiskChain) Links(id int) ([]Link, error) {
	return c.w.Links(id)
}

func (c *DiskChain) Find(value interface{}) (id int, err error) {
	return c.w.Find(value)
}
