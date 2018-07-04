package disk

import (
	"io"
	"sync"
)

type Writer struct {
	mu   sync.Mutex
	file io.WriteSeeker
}

func NewWriter(w io.WriteSeeker) *Writer {
	return &Writer{file: w}
}

func (w *Writer) Append(buf []byte) (int64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	off, err := w.file.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}

	_, err = w.file.Write(buf)
	return off, err
}

type Reader struct {
	mu   sync.Mutex
	file io.ReadSeeker
}

func NewReader(r io.ReadSeeker) *Reader {
	return &Reader{file: r}
}

func (r *Reader) Read(offset, length int64) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.file.Seek(offset, io.SeekStart)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, length)
	n, err := io.ReadFull(r.file, buf)

	return buf[:n], err
}
