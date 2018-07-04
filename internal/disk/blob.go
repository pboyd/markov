package disk

import (
	"encoding/binary"
	"errors"
)

const (
	maxBlobSize    = 1 << 16
	blobSizeLength = 2
)

// WriteBlob writes a variable length buffer to the given offset.
func (f *File) WriteBlob(offset int64, buf []byte) (int64, error) {
	if len(buf) > maxBlobSize {
		return 0, errors.New("blob exceeds max size")
	}

	size := make([]byte, blobSizeLength)
	binary.BigEndian.PutUint16(size, uint16(len(buf)))

	f.mu.Lock()
	defer f.mu.Unlock()

	startOffset, err := f.writeLock(offset, size)
	if err != nil {
		return 0, err
	}

	_, err = f.writeLock(startOffset+int64(len(size)), buf)
	return startOffset, err
}

// OverwriteBlob modifies an existing blob. The new buffer must be the same
// size as the original.
func (f *File) OverwriteBlob(offset int64, buf []byte) error {
	if offset < 0 {
		return errors.New("OverwriteBlob with negative offset")
	}

	size, err := f.ReadBlobSize(offset)
	if err != nil {
		return err
	}

	if uint16(len(buf)) != size {
		return errors.New("OverwriteBlob size mismatch")
	}

	_, err = f.WriteBlob(offset, buf)
	return err
}

func (f *File) ReadBlobSize(offset int64) (uint16, error) {
	sizeBuf, err := f.Read(offset, blobSizeLength)
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint16(sizeBuf), nil
}

func (f *File) ReadBlob(offset int64) ([]byte, error) {
	size, err := f.ReadBlobSize(offset)
	if err != nil {
		return nil, err
	}

	return f.Read(offset+blobSizeLength, int64(size))
}
