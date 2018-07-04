package disk

import (
	"encoding/binary"
	"errors"
	"io"
)

const (
	maxBlobSize      = 1 << 16
	BlobHeaderLength = 2
)

// WriteBlob writes a variable length buffer to the given offset.
func WriteBlob(ws io.WriteSeeker, offset int64, buf []byte) (int64, error) {
	if len(buf) > maxBlobSize {
		return 0, errors.New("blob exceeds max size")
	}

	size := make([]byte, BlobHeaderLength)
	binary.BigEndian.PutUint16(size, uint16(len(buf)))

	startOffset, err := Write(ws, offset, size)
	if err != nil {
		return 0, err
	}

	_, err = ws.Write(buf)
	return startOffset, err
}

func ReadBlobSize(rs io.ReadSeeker, offset int64) (uint16, error) {
	sizeBuf, err := Read(rs, offset, BlobHeaderLength)
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint16(sizeBuf), nil
}

func ReadBlob(rs io.ReadSeeker, offset int64) ([]byte, error) {
	size, err := ReadBlobSize(rs, offset)
	if err != nil {
		return nil, err
	}

	return Read(rs, offset+2, int64(size))
}
