package disk

import (
	"encoding/binary"
	"errors"
	"os"
)

var ErrOutOfBounds = errors.New("list index out of bounds")

type List struct {
	file        *os.File
	elementSize uint16
	bucketCap   uint16

	headBucket *listBucket

	readBucket       *listBucket
	readBucketNumber uint16

	tailBucket       *listBucket
	tailBucketNumber uint16
}

func NewList(f *os.File, elementSize uint16, buf []byte) (*List, error) {
	l := &List{
		file:        f,
		elementSize: elementSize,
		bucketCap:   (uint16(len(buf)) - offsetLength) / elementSize,
	}

	l.headBucket = newListBucketFromBuf(-1, elementSize, buf)
	l.readBucket = l.headBucket

	err := l.loadTailBucket()
	if err != nil {
		return nil, err
	}

	return l, nil
}

func ListBucketSize(elementSize, cap uint16) int {
	return int(elementSize*cap) + offsetLength
}

func (l *List) Append(buf []byte) error {
	if l.tailBucket.Count == l.bucketCap {
		newTail, err := newListBucket(l.file, l.elementSize, l.bucketCap)
		if err != nil {
			return err
		}

		l.tailBucket.SetNext(newTail.offset)
		err = l.tailBucket.Flush(l.file)
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
	if number == 0 {
		return l.headBucket, nil
	}

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
	offset := l.headBucket.Next()
	var err error

	buf := make([]byte, offsetLength)

	for i := uint16(1); i < number; i++ {
		_, err = l.file.ReadAt(buf, offset+sectionHeaderLength)
		if err != nil {
			return 0, err
		}

		offset = int64(binary.BigEndian.Uint64(buf))

		// No more buckets
		if offset == 0 {
			return -1, nil
		}
	}

	return offset, nil
}

func (l *List) loadTailBucket() error {
	secondOffset := l.headBucket.Next()
	if secondOffset == 0 {
		l.tailBucketNumber = 0
		l.tailBucket = l.headBucket
		return nil
	}

	bucketOffset, number, err := l.findTailBucket(secondOffset)
	if err != nil {
		return err
	}

	l.tailBucket, err = readListBucket(l.file, bucketOffset, l.elementSize, l.bucketCap)
	// The search starts at the second bucket, so add 1.
	l.tailBucketNumber = number + 1

	return err
}

func (l *List) findTailBucket(offset int64) (int64, uint16, error) {
	var number uint16
	var err error

	buf := make([]byte, offsetLength)

	for {
		_, err = l.file.ReadAt(buf, offset+sectionHeaderLength)
		if err != nil {
			return 0, 0, err
		}

		next := int64(binary.BigEndian.Uint64(buf))
		if next == 0 {
			break
		}

		offset = next
		number++
	}

	return offset, number, nil
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

func (l *List) Flush() error {
	// Never write the head bucket
	if l.tailBucketNumber == 0 {
		return nil
	}

	return l.tailBucket.Flush(l.file)
}

// listBucket contains a subset of the list's elements. The list can contain
// any number of buckets. The offset of the next address is stored in the first
// 8 bytes of the bucket.
type listBucket struct {
	offset      int64
	buf         []byte
	elementSize uint16
	managed     bool
	Count       uint16
}

func newListBucket(f *os.File, elementSize, cap uint16) (*listBucket, error) {
	b := &listBucket{
		offset:      -1,
		elementSize: elementSize,
		buf:         make([]byte, sectionHeaderLength+ListBucketSize(elementSize, cap)),
		managed:     true,
	}

	putSectionHeader(b.buf, listBucketSection, uint32(len(b.buf))-sectionHeaderLength)

	return b, b.Flush(f)
}

func newListBucketFromBuf(offset int64, elementSize uint16, buf []byte) *listBucket {
	b := &listBucket{
		offset:      offset,
		elementSize: elementSize,
		buf:         buf,
		managed:     false,
	}
	b.Count = b.countElements()

	return b
}

func readListBucket(f *os.File, offset int64, elementSize, cap uint16) (*listBucket, error) {
	b := &listBucket{
		offset:      offset,
		elementSize: elementSize,
		managed:     true,
	}

	size := elementSize*cap + offsetLength + recordHeaderLength

	b.buf = make([]byte, size)
	_, err := f.ReadAt(b.buf, offset)
	if err != nil {
		return nil, err
	}

	st, _ := sectionHeader(b.buf)
	if st != listBucketSection {
		return nil, sectionTypeError(st)
	}

	b.Count = b.countElements()

	return b, nil
}

func (b *listBucket) indexOffset(i uint16) (offset uint16) {
	offset = offsetLength + i*b.elementSize
	if b.managed {
		offset += sectionHeaderLength
	}
	return
}

func (b *listBucket) countElements() uint16 {
	// TODO: Maybe check the last element first? Only the last bucket will
	// partially filled.

	var count uint16

	for i := b.indexOffset(0); i < uint16(len(b.buf)); i += b.elementSize {
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
	copy(b.buf[b.indexOffset(b.Count):], buf[:b.elementSize])
	b.Count++
}

func (b *listBucket) Get(i uint16) []byte {
	o := b.indexOffset(i)
	return b.buf[o : o+b.elementSize]
}

func (b *listBucket) Next() int64 {
	start := 0
	if b.managed {
		start += sectionHeaderLength
	}
	return int64(binary.BigEndian.Uint64(b.buf[start:]))
}

func (b *listBucket) SetNext(offset int64) {
	start := 0
	if b.managed {
		start += sectionHeaderLength
	}
	binary.BigEndian.PutUint64(b.buf[start:], uint64(offset))
}

func (b *listBucket) Flush(f *os.File) error {
	if !b.managed {
		return nil
	}

	offset, err := writeAt(f, b.offset, b.buf)
	if err != nil {
		return err
	}

	b.offset = offset
	return nil
}
