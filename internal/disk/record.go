package disk

import (
	"encoding/binary"
	"os"
)

const recordHeaderSize = 4

type Record struct {
	Offset int64
	List   *List
	file   *os.File
	buf    []byte
}

func NewRecord(file *os.File, value []byte, listElementSize uint16, listBucketLen uint16) (*Record, error) {
	size := recordHeaderSize + len(value)
	size += ListBucketSize(listElementSize, listBucketLen)

	buf := make([]byte, size)
	n := 0
	binary.BigEndian.PutUint16(buf[n:], uint16(len(value)))
	n += 2
	binary.BigEndian.PutUint16(buf[n:], listBucketLen)
	n += 2
	copy(buf[n:], value)
	n += len(value)

	r := &Record{
		Offset: -1,
		file:   file,
		buf:    buf,
	}

	err := r.Write()
	if err != nil {
		return nil, err
	}

	r.List, err = NewList(file, listElementSize, buf[n:])
	if err != nil {
		return nil, err
	}

	return r, nil
}

func ReadRecord(file *os.File, offset int64, listElementSize uint16) (*Record, error) {
	r := &Record{
		Offset: offset,
		file:   file,
	}

	r.buf = make([]byte, recordHeaderSize)
	_, err := file.ReadAt(r.buf, offset)
	if err != nil {
		return nil, err
	}

	valueLen := r.valueLength()
	size := int(valueLen) + ListBucketSize(listElementSize, r.listElements())
	rest := make([]byte, size)

	_, err = file.ReadAt(rest, offset+recordHeaderSize)
	if err != nil {
		return nil, err
	}

	r.buf = append(r.buf, rest...)

	listOffset := recordHeaderSize + int64(valueLen)
	r.List, err = NewList(file, listElementSize, r.buf[listOffset:])
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Record) Write() error {
	offset, err := Write(r.file, r.Offset, r.buf)
	if err != nil {
		return err
	}

	r.Offset = offset

	if r.List == nil {
		return nil
	}

	return r.List.Flush()
}

func (r *Record) valueLength() uint16 {
	return binary.BigEndian.Uint16(r.buf)
}

func (r *Record) listElements() uint16 {
	return binary.BigEndian.Uint16(r.buf[2:])
}

func (r *Record) Value() []byte {
	start := recordHeaderSize
	end := start + int(r.valueLength())
	return r.buf[start:end]
}
