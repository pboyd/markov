package disk

import (
	"bytes"
	"testing"
)

func TestBlob(t *testing.T) {
	rw, cleanup := tempFile(t)
	defer cleanup()
	f := NewFile(rw)

	buf := make([]byte, 8)
	for i := range buf {
		buf[i] = 0xff
	}

	t.Run("Read/Write", func(t *testing.T) {
		for i := int64(0); i < 10; i++ {
			offset, err := f.WriteBlob(-1, buf)
			if err != nil {
				t.Fatalf("got error: %v", err)
			}

			if offset != i*10 {
				t.Errorf("got offset %d, want %d", offset, i*10)
			}

			rbuf, err := f.ReadBlob(offset)
			if err != nil {
				t.Fatalf("got read error: %v", err)
			}

			if !bytes.Equal(rbuf, buf) {
				t.Errorf("\ngot:  %v\nwant: %v", rbuf, buf)
			}
		}
	})

	t.Run("Overwrite", func(t *testing.T) {
		const offset = 10

		owbuf := make([]byte, 9)

		err := f.OverwriteBlob(offset, owbuf)
		if err == nil {
			t.Errorf("got nil error with wrong size buffer")
		}

		err = f.OverwriteBlob(offset, owbuf[:8])
		if err != nil {
			t.Fatalf("got write error: %v", err)
		}

		rbuf, err := f.ReadBlob(offset)
		if err != nil {
			t.Fatalf("got read error: %v", err)
		}

		if !bytes.Equal(rbuf, owbuf[:8]) {
			t.Errorf("\ngot:  %v\nwant: %v", rbuf, owbuf)
		}
	})
}