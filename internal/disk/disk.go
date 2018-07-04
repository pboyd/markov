package disk

import "io"

const addressLength = 8

// Write writes the buffer to the file at the given offset.
//
// A negative offset starts from the end of the file. -1 is the very end, -2 is
// the byte before and so forth.
func Write(w io.WriteSeeker, offset int64, buf []byte) (int64, error) {
	var newOff int64
	var err error

	if offset >= 0 {
		if wa, ok := w.(io.WriterAt); ok {
			_, err = wa.WriteAt(buf, offset)
			return offset, err
		}

		newOff, err = w.Seek(offset, io.SeekStart)
	} else {
		newOff, err = w.Seek(^offset, io.SeekEnd)
	}
	if err != nil {
		return 0, err
	}

	_, err = w.Write(buf)
	return newOff, err
}

func Read(r io.ReadSeeker, offset, length int64) ([]byte, error) {
	buf := make([]byte, length)
	n, err := readBuf(r, offset, buf)
	return buf[:n], err
}

func readBuf(r io.ReadSeeker, offset int64, buf []byte) (int, error) {
	if ra, ok := r.(io.ReaderAt); ok {
		return ra.ReadAt(buf, offset)
	}

	_, err := r.Seek(offset, io.SeekStart)
	if err != nil {
		return 0, err
	}

	return io.ReadFull(r, buf)
}
