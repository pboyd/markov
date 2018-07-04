package disk

import (
	"bytes"
	"io"
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
		}
	})

	t.Run("Read", func(t *testing.T) {
		rr := NewRecordReader(file, 0, listElementSize)
		found := 0

		for {
			record, err := rr.Read()
			if err == io.EOF {
				break
			}

			if err != nil {
				t.Fatalf("got read error: %v", err)
			}

			found++

			value := record.Value()
			if !bytes.Equal(value, buf) {
				t.Errorf("\ngot:  %v\nwant: %v", value, buf)
			}

			listLen := record.List.Len()
			if listLen != listBucketLength*2 {
				t.Errorf("got %d list items, want %d", listLen, listBucketLength*2)
			}
		}

		if found != inserts {
			t.Errorf("got %d records, want %d", found, inserts)
		}
	})
}
