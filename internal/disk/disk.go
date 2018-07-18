package disk

import (
	"fmt"
	"io"
	"os"
)

const (
	offsetLength        = 8
	sectionHeaderLength = 4
	recordHeaderLength  = 4
)

type sectionType uint8

const (
	recordSection sectionType = 1 << iota
	listBucketSection
)

type sectionTypeError sectionType

func (err sectionTypeError) Error() string {
	return fmt.Sprintf("unexpected section type %d", uint8(err))
}

// putSectionHeader writes a 4 byte header to buf.
//
// The low four bits of t become the 4 high bits of the first byte in the
// buffer. The low 28 bits of len take the remaining space. The high four bits
// of t and len are ignored.
func putSectionHeader(buf []byte, t sectionType, len uint32) {
	buf[0] = byte(t<<4) | byte(len>>20)
	buf[1] = byte(len >> 12)
	buf[2] = byte(len >> 4)
	buf[3] = byte(len)
}

func sectionHeader(buf []byte) (t sectionType, len uint32) {
	t = sectionType(buf[0] >> 4)
	len = uint32(buf[3]) | uint32(buf[2])<<4 | uint32(buf[1])<<12 | uint32(buf[0]&0xf)<<20
	return
}

// writeAt writes the buffer to the file at the given offset.
//
// A negative offset starts from the end of the file. -1 is the very end, -2 is
// the byte before and so forth.
func writeAt(w *os.File, offset int64, buf []byte) (int64, error) {
	var newOff int64
	var err error

	if offset >= 0 {
		_, err = w.WriteAt(buf, offset)
		return offset, err
	}

	newOff, err = w.Seek(^offset, io.SeekEnd)
	if err != nil {
		return 0, err
	}

	_, err = w.Write(buf)
	return newOff, err
}
