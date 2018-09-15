package markov

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"os"
	"sync"

	"github.com/pboyd/markov/internal/disk"
)

var _ ReadWriteChain = &DiskChainWriter{}

const (
	diskHeader = "MKV\u0001"

	linkListItemSize       = 12
	linkListItemsPerBucket = 128
)

// DiskChainWriter is a ReadWriteChain implementation for file-based chains.
//
// The chain operates on-disk as much as possible, which makes it much slower
// than an in-memory chain. Building a disk chain by feeding values is
// particularly inefficient (but possible, and sometimes necessary). Building a
// MemoryChain and copying to a DiskChainWriter is likely faster.
//
// Values can be strings, runes or any builtin numeric type.
type DiskChainWriter struct {
	file           *os.File
	fileWriteMutex sync.Mutex

	index      map[interface{}]int64
	indexMutex sync.RWMutex
}

// NewDiskChainWriter creates a new DiskChainWriter. File must be writable. Any
// existing data in the file will be lost.
func NewDiskChainWriter(file *os.File) (*DiskChainWriter, error) {
	err := file.Truncate(0)
	if err != nil {
		return nil, err
	}

	_, err = file.WriteAt([]byte(diskHeader), 0)
	if err != nil {
		return nil, err
	}

	return &DiskChainWriter{
		file:  file,
		index: make(map[interface{}]int64),
	}, nil
}

// OpenDiskChainWriter reads an existing disk chain. If file is a read/write
// handle the disk chain can be updated.
func OpenDiskChainWriter(file *os.File) (*DiskChainWriter, error) {
	err := verifyDiskHeader(file)
	if err != nil {
		return nil, err
	}

	c := &DiskChainWriter{
		file:  file,
		index: make(map[interface{}]int64),
	}

	return c, c.buildIndex()
}

func verifyDiskHeader(file *os.File) error {
	actualHeader := make([]byte, len(diskHeader))
	_, err := file.ReadAt(actualHeader, 0)
	if err != nil {
		return err
	}

	if !bytes.Equal(actualHeader, []byte(diskHeader)) {
		return errors.New("markov: unrecognized file")
	}

	return nil
}

// Get returns a value by it's ID. Returns nil if the ID doesn't exist.
func (c *DiskChainWriter) Get(id int) (interface{}, error) {
	if id == 0 {
		id = len(diskHeader)
	}

	record, err := disk.ReadRecord(c.file, int64(id), linkListItemSize)
	if err != nil {
		return nil, err
	}

	return unmarshalValue(record.Value())
}

// Links returns the items linked to the given item.
//
// Returns ErrNotFound if the ID doesn't exist.
func (c *DiskChainWriter) Links(id int) ([]Link, error) {
	if id == 0 {
		id = len(diskHeader)
	}

	list, err := c.linkList(int64(id))
	if err != nil {
		return nil, err
	}

	total := list.Len()

	links := make([]Link, total)
	counts := make([]uint32, total)
	sum := 0

	for i := 0; i < total; i++ {
		value, err := list.Get(uint16(i))
		if err != nil {
			return nil, err
		}

		id, count := c.unpackLinkValue(value)
		sum += int(count)
		counts[i] = count
		links[i] = Link{ID: id}
	}

	for i := range links {
		links[i].Probability = float64(counts[i]) / float64(sum)
	}

	return links, nil
}

// Find returns the ID for the given value.
//
// Returns ErrNotFound if the value doesn't exist.
func (c *DiskChainWriter) Find(value interface{}) (int, error) {
	c.indexMutex.RLock()
	defer c.indexMutex.RUnlock()

	id, ok := c.index[value]
	if !ok {
		return 0, ErrNotFound
	}

	return int(id), nil
}

// Add conditionally inserts a new value to the chain.
//
// If the value exists it's ID is returned.
func (c *DiskChainWriter) Add(value interface{}) (int, error) {
	return c.add(value, linkListItemsPerBucket)
}

func (c *DiskChainWriter) add(value interface{}, bucketSize uint16) (int, error) {
	existing, err := c.Find(value)
	if err == nil {
		return existing, nil
	}

	valueBuf, err := marshalValue(value)
	if err != nil {
		return 0, err
	}

	c.fileWriteMutex.Lock()
	defer c.fileWriteMutex.Unlock()

	record, err := disk.NewRecord(c.file, valueBuf, linkListItemSize, bucketSize)
	if err != nil {
		return 0, err
	}

	c.indexMutex.Lock()
	defer c.indexMutex.Unlock()

	c.index[value] = record.Offset

	return int(record.Offset), err
}

// Relate increases the number of times child occurs after parent.
func (c *DiskChainWriter) Relate(parent, child int, delta int) error {
	record, err := disk.ReadRecord(c.file, int64(parent), linkListItemSize)
	if err != nil {
		return err
	}

	err = c.relateToRecord(record, child, delta)
	if err != nil {
		return err
	}

	return record.Write()
}

func (c *DiskChainWriter) relateToRecord(record *disk.Record, child, delta int) error {
	c.fileWriteMutex.Lock()
	defer c.fileWriteMutex.Unlock()

	newChild := true

	// Check for an existing entry
	for i := 0; i < record.List.Len(); i++ {
		value, err := record.List.Get(uint16(i))
		if err != nil {
			return err
		}

		id, count := c.unpackLinkValue(value)
		if id == child {
			if int(count)+delta > math.MaxUint32 {
				return errors.New("uint32 overflow")
			}

			count += uint32(delta)
			c.updateLinkCount(value, count)
			newChild = false
			break
		}
	}

	if newChild {
		if delta > math.MaxUint32 {
			return errors.New("uint32 overflow")
		}

		err := record.List.Append(c.packLinkValue(child, uint32(delta)))
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *DiskChainWriter) linkList(id int64) (*disk.List, error) {
	record, err := disk.ReadRecord(c.file, id, linkListItemSize)
	if err != nil {
		return nil, err
	}

	return record.List, nil
}

func (c *DiskChainWriter) unpackLinkValue(value []byte) (id int, count uint32) {
	id = int(binary.BigEndian.Uint64(value))
	count = binary.BigEndian.Uint32(value[8:])
	return
}

func (c *DiskChainWriter) packLinkValue(id int, count uint32) []byte {
	buf := make([]byte, 12)
	binary.BigEndian.PutUint64(buf, uint64(id))
	binary.BigEndian.PutUint32(buf[8:], count)
	return buf
}

func (c *DiskChainWriter) updateLinkCount(buf []byte, count uint32) {
	binary.BigEndian.PutUint32(buf[8:], count)
}

func (c *DiskChainWriter) buildIndex() error {
	c.indexMutex.Lock()
	defer c.indexMutex.Unlock()

	rr := disk.NewRecordReader(c.file, int64(len(diskHeader)), linkListItemSize)
	for {
		record, err := rr.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		value, err := unmarshalValue(record.Value())
		if err != nil {
			// FIXME? Is this really fatal?
			return err
		}

		c.index[value] = record.Offset
	}
}

// Next returns the id after the given id. Satisfies the IterativeChain
// interface.
func (c *DiskChainWriter) Next(id int) (int, error) {
	if id == 0 {
		id = len(diskHeader)
	}

	rr := disk.NewRecordReader(c.file, int64(id), linkListItemSize)
	next, err := rr.Next()
	if err == io.EOF {
		return 0, ErrBrokenChain
	}

	if err != nil {
		return 0, err
	}

	return int(next), nil
}

// CopyFrom satisfies the CopyFrom interface. It's faster than the generic Copy
// algorithm implemented by Copy.
func (c *DiskChainWriter) CopyFrom(src Chain) error {
	valueToDestID := make(map[interface{}]int)
	srcIDtoDestID := make(map[int]int)

	walker := IterativeWalker(src)
	for {
		value, err := walker.Next()
		if err != nil {
			if err == ErrBrokenChain {
				break
			}

			return err
		}

		srcID, err := src.Find(value)
		if err != nil {
			return err
		}

		links, err := linkCounts(src, srcID)
		if err != nil {
			return err
		}

		destID, err := c.add(value, uint16(len(links)))
		if err != nil {
			return err
		}

		valueToDestID[value] = destID
		srcIDtoDestID[srcID] = destID
	}

	for value, destID := range valueToDestID {
		srcID, err := src.Find(value)
		if err != nil {
			return err
		}

		linkCounts, err := linkCounts(src, srcID)
		if err != nil {
			return err
		}

		record, err := disk.ReadRecord(c.file, int64(destID), linkListItemSize)
		if err != nil {
			return err
		}

		for _, link := range linkCounts {
			err = c.relateToRecord(record, srcIDtoDestID[link.ID], link.Count)
			if err != nil {
				return err
			}
		}

		err = record.Write()
		if err != nil {
			return err
		}
	}

	return nil
}

// Random pseudo-randomly picks a value and returns it. Satisfies the
// RandomChain interface.
func (c *DiskChainWriter) Random() (interface{}, error) {
	c.indexMutex.RLock()
	defer c.indexMutex.RUnlock()

	for v := range c.index {
		return v, nil
	}

	return nil, nil
}
