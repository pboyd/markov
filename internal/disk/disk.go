package disk

import (
	"encoding/binary"
	"io"
)

const addressLength = 8

// Write writes the buffer to the file at the given offset.
//
// A negative offset starts from the end of the file. -1 is the very end, -2 is
// the byte before and so forth.
func Write(ws io.WriteSeeker, offset int64, buf []byte) (int64, error) {
	var newOff int64
	var err error

	if offset >= 0 {
		newOff, err = ws.Seek(offset, io.SeekStart)
	} else {
		newOff, err = ws.Seek(^offset, io.SeekEnd)
	}
	if err != nil {
		return 0, err
	}

	_, err = ws.Write(buf)
	return newOff, err
}

func Read(rs io.ReadSeeker, offset, length int64) ([]byte, error) {
	_, err := rs.Seek(offset, io.SeekStart)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, length)
	n, err := io.ReadFull(rs, buf)

	return buf[:n], err
}

func readNext(r io.Reader, length int64) ([]byte, error) {
	buf := make([]byte, length)
	n, err := io.ReadFull(r, buf)

	return buf[:n], err
}

func readAddress(rs io.ReadSeeker, offset int64) (int64, error) {
	buf, err := Read(rs, offset, addressLength)
	if err != nil {
		return 0, err
	}

	return int64(binary.BigEndian.Uint64(buf)), nil
}
