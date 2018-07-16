package markov

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"os"

	"github.com/pboyd/markov/internal/disk"
)

var _ ReadWriteChain = &DiskChainWriter{}

const (
	diskHeader = "MKV\u0001"

	linkListItemSize       = 12
	linkListItemsPerBucket = 128
)

// DiskChainWriter is a ReadWriteChain stored in a file.
type DiskChainWriter struct {
	file  *os.File
	index map[interface{}]int64
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

func (c *DiskChainWriter) Find(value interface{}) (int, error) {
	id, ok := c.index[value]
	if !ok {
		return 0, ErrNotFound
	}

	return int(id), nil
}

func (c *DiskChainWriter) Add(value interface{}) (int, error) {
	existing, err := c.Find(value)
	if err == nil {
		return existing, nil
	}

	valueBuf, err := marshalValue(value)
	if err != nil {
		return 0, err
	}

	record, err := disk.NewRecord(c.file, valueBuf, linkListItemSize, linkListItemsPerBucket)
	if err != nil {
		return 0, err
	}

	c.index[value] = record.Offset

	return int(record.Offset), err
}

func (c *DiskChainWriter) Relate(parent, child int, delta int) error {
	record, err := disk.ReadRecord(c.file, int64(parent), linkListItemSize)
	if err != nil {
		return err
	}

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

		record.List.Append(c.packLinkValue(child, uint32(delta)))
	}

	return record.Write()
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

func (c *DiskChainWriter) Next(id int) (int, error) {
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
