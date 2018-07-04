package disk

import (
	"encoding/binary"
	"errors"
	"io"
)

const listHeaderSize = 4

var ErrOutOfBounds = errors.New("list index out of bounds")

type List struct {
	file        io.ReadWriteSeeker
	offset      int64
	elementSize uint16
	bucketCap   uint16

	tailBucket       *listBucket
	tailBucketNumber uint16

	readBucket       *listBucket
	readBucketNumber uint16
}

func NewList(rws io.ReadWriteSeeker, elementSize, bucketCap uint16) (*List, error) {
	l := &List{
		file:        rws,
		elementSize: elementSize,
		bucketCap:   bucketCap,
		offset:      -1,
	}
	err := l.Flush()
	if err != nil {
		return nil, err
	}

	l.tailBucket, err = newListBucket(l.file, elementSize, bucketCap)
	if err != nil {
		return nil, err
	}

	l.readBucket = l.tailBucket

	return l, nil
}

func (l *List) Append(buf []byte) error {
	if l.tailBucket.Count == l.bucketCap {
		newTail, err := newListBucket(l.file, l.elementSize, l.bucketCap)
		if err != nil {
			return err
		}

		l.tailBucket.SetNext(newTail)
		err = l.tailBucket.Flush()
		if err != nil {
			return err
		}

		l.tailBucket = newTail
		l.tailBucketNumber++
	}

	l.tailBucket.Append(buf)
	return nil
}

func (l *List) Get(i uint16) ([]byte, error) {
	bucketNumber := i / l.bucketCap
	bucketIndex := i % l.bucketCap

	b, err := l.loadReadBucket(bucketNumber)
	if err != nil {
		return nil, err
	}

	if b == nil || bucketIndex > b.Count {
		return nil, ErrOutOfBounds
	}

	return b.Get(bucketIndex), nil
}

func (l *List) loadReadBucket(number uint16) (*listBucket, error) {
	if l.tailBucket != nil && number == l.tailBucketNumber {
		return l.tailBucket, nil
	}

	if number == l.readBucketNumber {
		return l.readBucket, nil
	}

	offset, err := l.bucketOffset(number)
	if err != nil {
		return nil, err
	}

	if offset < 0 {
		return nil, nil
	}

	bucket, err := readListBucket(l.file, offset, l.elementSize, l.bucketCap)
	if err != nil {
		return nil, err
	}

	l.readBucket = bucket
	l.readBucketNumber = number

	return l.readBucket, nil
}

func (l *List) bucketOffset(number uint16) (int64, error) {
	var offset int64 = l.offset + int64(listHeaderSize)
	var err error

	for i := uint16(0); i < number; i++ {
		// FIXME: It shouldn't allocate a new buffer on every iteration
		offset, err = readAddress(l.file, offset)
		if err != nil {
			return 0, err
		}

		// No more buckets
		if offset == 0 {
			return -1, nil
		}
	}

	return offset, nil
}

func (l *List) loadTailBucket() error {
	bucketOffset, number, err := l.findTailBucket()
	if err != nil {
		return err
	}

	l.tailBucket, err = readListBucket(l.file, bucketOffset, l.elementSize, l.bucketCap)
	l.tailBucketNumber = number
	return err
}

func (l *List) findTailBucket() (int64, uint16, error) {
	var offset int64 = l.offset + int64(listHeaderSize)
	var number uint16

	for {
		// FIXME: It shouldn't allocate a new buffer on every iteration
		next, err := readAddress(l.file, offset)
		if err != nil {
			return 0, 0, err
		}

		if next == 0 {
			break
		}

		offset = next
		number++
	}

	return offset, number, nil
}

func ReadList(rws io.ReadWriteSeeker, offset int64) (*List, error) {
	buf, err := Read(rws, offset, listHeaderSize)
	if err != nil {
		return nil, err
	}

	l := &List{
		file:        rws,
		elementSize: binary.BigEndian.Uint16(buf),
		bucketCap:   binary.BigEndian.Uint16(buf[2:]),
		offset:      offset,
	}

	l.readBucket, err = readListBucket(l.file, offset+listHeaderSize, l.elementSize, l.bucketCap)
	if err != nil {
		return nil, err
	}

	err = l.loadTailBucket()
	if err != nil {
		return nil, err
	}

	return l, nil
}

func (l *List) ElementSize() uint16 {
	return l.elementSize
}

// Length returns the number of elements in the list.
func (l *List) Len() int {
	length := int(l.tailBucketNumber) * int(l.bucketCap)
	length += int(l.tailBucket.Count)
	return length
}

func (l *List) Offset() int64 {
	return l.offset
}

func (l *List) Flush() error {
	offset, err := Write(l.file, l.offset, l.header())
	if err != nil {
		return err
	}

	l.offset = offset
	return l.tailBucket.Flush()
}

func (l *List) header() []byte {
	buf := make([]byte, listHeaderSize)
	binary.BigEndian.PutUint16(buf, l.elementSize)
	binary.BigEndian.PutUint16(buf[2:], l.bucketCap)
	return buf
}

type listBucket struct {
	file        io.ReadWriteSeeker
	offset      int64
	dirty       bool
	buf         []byte
	elementSize uint16
	Count       uint16
}

func newListBucket(rws io.ReadWriteSeeker, elementSize, cap uint16) (*listBucket, error) {
	b := &listBucket{
		file:        rws,
		offset:      -1,
		elementSize: elementSize,
		buf:         make([]byte, elementSize*cap+addressLength),
		dirty:       true,
	}
	return b, b.Flush()
}

func readListBucket(rws io.ReadWriteSeeker, offset int64, elementSize, cap uint16) (*listBucket, error) {
	b := &listBucket{
		file:        rws,
		offset:      offset,
		elementSize: elementSize,
	}

	size := elementSize*cap + addressLength

	var err error
	b.buf, err = Read(b.file, offset, int64(size))
	if err != nil {
		return nil, err
	}

	b.Count = b.countElements()

	return b, nil
}

func (b *listBucket) indexOffset(i uint16) uint16 {
	return addressLength + i*b.elementSize
}

func (b *listBucket) countElements() uint16 {
	// TODO: Maybe check the last element first? Only the last bucket will
	// partially filled.

	var count uint16

	for i := uint16(addressLength); i < uint16(len(b.buf)); i += b.elementSize {
		if isNull(b.buf[i : i+b.elementSize]) {
			return count
		}
		count++
	}
	return count
}

func isNull(buf []byte) bool {
	for i := 0; i < len(buf); i++ {
		if buf[i] != 0 {
			return false
		}
	}
	return true
}

func (b *listBucket) Append(buf []byte) {
	b.dirty = true
	copy(b.buf[b.indexOffset(b.Count):], buf[:b.elementSize])
	b.Count++
}

func (b *listBucket) Get(i uint16) []byte {
	o := b.indexOffset(i)
	return b.buf[o : o+b.elementSize]
}

func (b *listBucket) SetNext(next *listBucket) {
	binary.BigEndian.PutUint64(b.buf, uint64(next.offset))
}

func (b *listBucket) Flush() error {
	if b == nil || !b.dirty {
		return nil
	}

	offset, err := Write(b.file, b.offset, b.buf)
	if err != nil {
		return err
	}

	b.offset = offset
	return nil
}
