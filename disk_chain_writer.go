package markov

import (
	"encoding/binary"
	"io"

	"github.com/pboyd/markov/internal/disk"
)

var _ ReadWriteChain = &DiskChainWriter{}

const (
	diskHeader = "MKV\u0001"

	linkListItemSize       = 12
	linkListItemsPerBucket = 8
)

type DiskChainWriter struct {
	file  io.ReadWriteSeeker
	index map[interface{}]int64
}

func NewDiskChainWriter(w io.ReadWriteSeeker) (*DiskChainWriter, error) {
	_, err := w.Write([]byte(diskHeader))
	if err != nil {
		return nil, err
	}

	return &DiskChainWriter{
		file:  w,
		index: make(map[interface{}]int64),
	}, nil
}

func (c *DiskChainWriter) Get(id int) (interface{}, error) {
	valueBuf, err := disk.ReadBlob(c.file, int64(id))
	if err != nil {
		return nil, err
	}

	return unmarshalValue(valueBuf)
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

	id, err := disk.WriteBlob(c.file, -1, valueBuf)
	if err != nil {
		return 0, err
	}

	c.index[value] = id

	_, err = disk.NewList(c.file, linkListItemSize, linkListItemsPerBucket)

	return int(id), err
}

func (c *DiskChainWriter) Relate(parent, child int, delta int) error {
	list, err := c.linkList(int64(parent))
	if err != nil {
		return err
	}

	newChild := true

	// Check for an existing entry
	for i := 0; i < list.Len(); i++ {
		value, err := list.Get(uint16(i))
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
		list.Append(c.packLinkValue(child, uint32(delta)))
	}

	return list.Flush()
}

func (c *DiskChainWriter) linkList(id int64) (*disk.List, error) {
	skip, err := disk.ReadBlobSize(c.file, id)
	if err != nil {
		return nil, err
	}

	linkOffset := id + disk.BlobHeaderLength + int64(skip)
	return disk.ReadList(c.file, linkOffset)
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
