package markov

import "os"

var _ Chain = &DiskChain{}

// DiskChain is a read-only Chain implementation for file-based chains.
type DiskChain struct {
	// DiskChain is wrapper around the read funcs of DiskChainWriter.
	w *DiskChainWriter
}

// ReadDiskChain reads a chain from a file.
func ReadDiskChain(fh *os.File) (*DiskChain, error) {
	w, err := OpenDiskChainWriter(fh)
	if err != nil {
		return nil, err
	}

	return &DiskChain{
		w: w,
	}, nil
}

// Get returns a value by it's ID.
func (c *DiskChain) Get(id int) (interface{}, error) {
	return c.w.Get(id)
}

// Links returns the items linked to the given item.
//
// Returns ErrNotFound if the ID doesn't exist.
func (c *DiskChain) Links(id int) ([]Link, error) {
	return c.w.Links(id)
}

// Find returns the ID for the given value.
//
// Returns ErrNotFound if the value doesn't exist.
func (c *DiskChain) Find(value interface{}) (id int, err error) {
	return c.w.Find(value)
}

// Next returns the id after the given id. Satifies the IterativeChain
// interface.
func (c *DiskChain) Next(id int) (int, error) {
	return c.w.Next(id)
}
