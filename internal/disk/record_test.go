package disk

import (
	"bytes"
	"testing"
)

func TestRecord(t *testing.T) {
	file, cleanup := tempFile(t)
	defer cleanup()

	buf := make([]byte, 8)
	for i := range buf {
		buf[i] = byte(i)
	}

	const (
		listElementSize  = 8
		listBucketLength = 16
		inserts          = 10
	)

	offsets := make([]int64, inserts)

	t.Run("Read/Write", func(t *testing.T) {
		for i := int64(0); i < inserts; i++ {
			recordIn, err := NewRecord(file, buf, listElementSize, listBucketLength)
			if err != nil {
				t.Fatalf("got error: %v", err)
			}

			for i := 0; i < listBucketLength*2; i++ {
				recordIn.List.Append(buf)
			}

			err = recordIn.Write()
			if err != nil {
				t.Fatalf("got error: %v", err)
			}

			if recordIn.Offset < 0 {
				t.Errorf("got offset %d, want >= 0", recordIn.Offset)
			}

			offsets[i] = recordIn.Offset
		}
	})

	t.Run("Read", func(t *testing.T) {
		for _, offset := range offsets {
			recordOut, err := ReadRecord(file, offset, listElementSize)
			if err != nil {
				t.Fatalf("got read error: %v", err)
			}

			value := recordOut.Value()
			if !bytes.Equal(value, buf) {
				t.Errorf("\ngot:  %v\nwant: %v", value, buf)
			}

			listLen := recordOut.List.Len()
			if listLen != listBucketLength*2 {
				t.Errorf("got %d list items, want %d", listLen, listBucketLength*2)
			}
		}
	})
}
