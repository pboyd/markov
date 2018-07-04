package disk

import (
	"io"
	"sync"
)

// File reads and writes data from a file (or, at least, a io.ReadWriteSeeker).
type File struct {
	mu sync.Mutex
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
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.writeLock(offset, buf)
}

func (f *File) writeLock(offset int64, buf []byte) (int64, error) {
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
	f.mu.Lock()
	defer f.mu.Unlock()

	_, err := f.rw.Seek(offset, io.SeekStart)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, length)
	n, err := io.ReadFull(f.rw, buf)

	return buf[:n], err
}
