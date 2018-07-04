package markov

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"

	"github.com/pboyd/markov/internal/disk"
)

var _ ReadWriteChain = &DiskChainWriter{}

const (
	diskHeader = "MKV\u0001"

	linkListItemSize       = 12
	linkListItemsPerBucket = 64
)

type DiskChainWriter struct {
	file  *os.File
	index map[interface{}]int64
}

func NewDiskChainWriter(w *os.File) (*DiskChainWriter, error) {
	_, err := w.Write([]byte(diskHeader))
	if err != nil {
		return nil, err
	}

	return &DiskChainWriter{
		file:  w,
		index: make(map[interface{}]int64),
	}, nil
}

func OpenDiskChainWriter(w *os.File) (*DiskChainWriter, error) {
	err := verifyDiskHeader(w)
	if err != nil {
		return nil, err
	}

	c := &DiskChainWriter{
		file:  w,
		index: make(map[interface{}]int64),
	}

	return c, c.buildIndex()
}

func verifyDiskHeader(w *os.File) error {
	actualHeader := make([]byte, len(diskHeader))
	_, err := w.ReadAt(actualHeader, 0)
	if err != nil {
		return err
	}

	if !bytes.Equal(actualHeader, []byte(diskHeader)) {
		return errors.New("markov: unrecognized file")
	}

	return nil
}

func (c *DiskChainWriter) Get(id int) (interface{}, error) {
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
			count += uint32(delta)
			c.updateLinkCount(value, count)
			newChild = false
			break
		}
	}

	if newChild {
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
