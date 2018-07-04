package markov

import (
	"encoding/binary"
	"io"

	"github.com/pboyd/markov/internal/disk"
)

var _ ReadWriteChain = &DiskChainWriter{}

var diskHeader = []byte{'M', 'K', 'V', 1}

type DiskChainWriter struct {
	ReadWriteChain
	file *disk.File
}

func NewDiskChainWriter(w io.ReadWriteSeeker) (*DiskChainWriter, error) {
	_, err := w.Write(diskHeader)
	if err != nil {
		return nil, err
	}

	file := disk.NewFile(w)

	return &DiskChainWriter{
		ReadWriteChain: &MemoryChain{},
		file:           file,
	}, nil
}

/*
func (c *DiskChainWriter) Get(id int) (interface{}, error) {
	return nil, ErrNotFound
}
*/

func (c *DiskChainWriter) Links(id int) ([]Link, error) {
	// FIXME: id == 0 is a special case

	id = translateDiskID(id)

	// I have no way to know the size that this should be..
	links := []Link{}
	counts := []uint32{}
	sum := 0

	item := disk.ReadList(c.file, int64(id))
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

/*
func (c *DiskChainWriter) Find(interface{}) (id int, err error) {
	return 0, ErrNotFound
}
*/

func (c *DiskChainWriter) Add(value interface{}) (int, error) {
	id, err := c.ReadWriteChain.Add(value)
	if err != nil {
		return 0, err
	}

	_, err = disk.NewList(c.file, c.packLinkValue(-1, 0))

	return id, err
}

func translateDiskID(id int) int {
	// This is temporary. Disk ID will eventually be the ID.
	return id*22 + len(diskHeader)
}

func (c *DiskChainWriter) Relate(parent, child int, delta int) error {
	parent = translateDiskID(parent)
	child = translateDiskID(child)

	item := disk.ReadList(c.file, int64(parent))

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
