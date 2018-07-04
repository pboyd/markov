package markov

import (
	"encoding/binary"
	"io"

	"github.com/pboyd/markov/internal/disk"
)

var _ ReadWriteChain = &DiskChainWriter{}

const (
	diskHeader = "MKV\u0001"

	// Put the index right after the header
	indexOffset = int64(len(diskHeader))
)

type DiskChainWriter struct {
	file *disk.File

	indexCreated bool
}

func NewDiskChainWriter(w io.ReadWriteSeeker) (*DiskChainWriter, error) {
	_, err := w.Write([]byte(diskHeader))
	if err != nil {
		return nil, err
	}

	file := disk.NewFile(w)

	return &DiskChainWriter{
		file: file,
	}, nil
}

func (c *DiskChainWriter) Get(id int) (interface{}, error) {
	valueBuf, err := c.file.ReadBlob(int64(id))
	if err != nil {
		return nil, err
	}

	return UnmarshalValue(valueBuf)
}

func (c *DiskChainWriter) Links(id int) ([]Link, error) {
	// FIXME: id == 0 is a special case

	// I have no way to know the size that this should be..
	links := []Link{}
	counts := []uint32{}
	sum := 0

	item, err := c.linkList(int64(id))
	if err != nil {
		return nil, err
	}

	for item != nil {
		value, err := item.Value()
		if err != nil {
			return nil, err
		}

		id, count := c.unpackLinkValue(value)
		sum += int(count)
		counts = append(counts, count)
		links = append(links, Link{
			ID: id,
		})

		item, err = item.Next()
		if err != nil {
			return nil, err
		}
	}

	for i := range links {
		links[i].Probability = float64(counts[i]) / float64(sum)
	}

	return links, nil
}

func (c *DiskChainWriter) Find(value interface{}) (int, error) {
	valueBuf, err := MarshalValue(value)
	if err != nil {
		return 0, err
	}

	return c.findByRawKey(valueBuf)
}

func (c *DiskChainWriter) findByRawKey(value []byte) (int, error) {
	root := disk.ReadBinaryTree(c.file, indexOffset)
	node, err := root.Search(value)
	if err != nil {
		return 0, err
	}

	if node == nil {
		return 0, ErrNotFound
	}

	id, err := node.Value()
	if id == 0 {
		panic(node.Offset)
	}
	return int(id), err
}

func (c *DiskChainWriter) Add(value interface{}) (int, error) {
	valueBuf, err := MarshalValue(value)
	if err != nil {
		return 0, err
	}

	existing, err := c.findByRawKey(valueBuf)
	if err == nil {
		return existing, nil
	}

	indexNode, err := c.createIndexEntry(valueBuf)
	if err != nil {
		return 0, err
	}

	id, err := c.file.WriteBlob(-1, valueBuf)
	if err != nil {
		return 0, err
	}

	err = indexNode.SetValue(id)
	if err != nil {
		return 0, err
	}

	_, err = disk.NewList(c.file, c.packLinkValue(-1, 0))

	return int(id), err
}

func (c *DiskChainWriter) Relate(parent, child int, delta int) error {
	item, err := c.linkList(int64(parent))
	if err != nil {
		return err
	}

	for {
		value, err := item.Value()
		if err != nil {
			return err
		}

		id, count := c.unpackLinkValue(value)

		// FIXME: What about uint32 overflows?
		if id < 0 {
			// It's a new list with a place holder value.
			return item.SetValue(c.packLinkValue(child, uint32(delta)))
		}

		if id == child {
			count += uint32(delta)
			return item.SetValue(c.packLinkValue(id, count))
		}

		next, err := item.Next()
		if err != nil {
			return err
		}

		if next == nil {
			break
		}

		item = next
	}

	item.InsertAfter(c.packLinkValue(child, uint32(delta)))

	return nil
}

func (c *DiskChainWriter) linkList(id int64) (*disk.ListItem, error) {
	skip, err := c.file.ReadBlobSize(id)
	if err != nil {
		return nil, err
	}

	linkOffset := id + disk.BlobHeaderLength + int64(skip)
	return disk.ReadList(c.file, linkOffset), nil
}

func (c *DiskChainWriter) unpackLinkValue(value []byte) (id int, count uint32) {
	id = int(binary.BigEndian.Uint64(value))
	count = binary.BigEndian.Uint32(value[8:])
	return
}

func (c *DiskChainWriter) packLinkValue(id int, count uint32) []byte {
	buf := make([]byte, 12)
	binary.BigEndian.PutUint64(buf, uint64(id))
	binary.BigEndian.PutUint32(buf[8:], uint32(count))
	return buf
}

func (c *DiskChainWriter) createIndexEntry(key []byte) (*disk.BinaryTreeNode, error) {
	if !c.indexCreated {
		root := disk.NewBinaryTree(c.file)
		c.indexCreated = true
		_, err := root.Insert(key, 0)

		if root.Offset != indexOffset {
			panic("index at wrong location")
		}

		return root, err
	}

	bst := disk.ReadBinaryTree(c.file, indexOffset)
	return bst.Insert(key, 0)
}
