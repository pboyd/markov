package disk

import (
	"encoding/binary"
	"io"
)

const addressLength = 8

// File reads and writes data from a file (or, at least, a io.ReadWriteSeeker).
type File struct {
	rw io.ReadWriteSeeker
}

func NewFile(rws io.ReadWriteSeeker) *File {
	return &File{rw: rws}
}

func (f *File) Append(buf []byte) (int64, error) {
	return f.Write(-1, buf)
}

// Write writes the buffer to the file at the given offset.
//
// A negative offset starts from the end of the file. -1 is the very end, -2 is
// the byte before and so forth.
func (f *File) Write(offset int64, buf []byte) (int64, error) {
	var newOff int64
	var err error

	if offset >= 0 {
		newOff, err = f.rw.Seek(offset, io.SeekStart)
	} else {
		newOff, err = f.rw.Seek(^offset, io.SeekEnd)
	}
	if err != nil {
		return 0, err
	}

	_, err = f.rw.Write(buf)
	return newOff, err
}

func (f *File) Read(offset, length int64) ([]byte, error) {
	_, err := f.rw.Seek(offset, io.SeekStart)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, length)
	n, err := io.ReadFull(f.rw, buf)

	return buf[:n], err
}

func (f *File) readNext(length int64) ([]byte, error) {
	buf := make([]byte, length)
	n, err := io.ReadFull(f.rw, buf)

	return buf[:n], err
}

func (f *File) writeAddress(offset, address int64) (int64, error) {
	buf := make([]byte, addressLength)
	binary.BigEndian.PutUint64(buf, uint64(address))
	return f.Write(offset, buf)
}

func (f *File) readAddress(offset int64) (int64, error) {
	buf, err := f.Read(offset, addressLength)
	if err != nil {
		return 0, err
	}

	return int64(binary.BigEndian.Uint64(buf)), nil
}
