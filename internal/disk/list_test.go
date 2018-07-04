package disk

import (
	"encoding/binary"
	"testing"
)

func TestList(t *testing.T) {
	rw, cleanup := tempFile(t)
	defer cleanup()

	l, err := NewList(rw, 8, 16)
	if err != nil {
		t.Fatalf("NewList failed: %v", err)
	}

	t.Run("Write", func(t *testing.T) {
		inserts := 1024

		buf := make([]byte, l.ElementSize())
		for i := uint16(0); i < uint16(inserts); i++ {
			binary.BigEndian.PutUint64(buf, uint64(i+1))
			l.Append(buf)
		}

		if l.Len() != inserts {
			t.Errorf("got len %d, want %d", l.Len(), inserts)
		}

		for i := uint16(0); i < uint16(inserts); i++ {
			buf, err := l.Get(i)
			if err != nil {
				t.Fatalf("Get failed: %v", err)
			}

			actual := binary.BigEndian.Uint64(buf)
			if actual != uint64(i+1) {
				t.Errorf("%d: want %d, got %d", i, i+1, actual)
			}
		}

		_, err = l.Get(1025)
		if err != ErrOutOfBounds {
			t.Errorf("got error %v, want %v", err, ErrOutOfBounds)
		}

		binary.BigEndian.PutUint64(buf, uint64(1025))
		l.Append(buf)

		if l.Len() != inserts+1 {
			t.Errorf("got len %d, want %d", l.Len(), inserts+1)
		}

		_, err = l.Get(1026)
		if err != ErrOutOfBounds {
			t.Errorf("got error %v, want %v", err, ErrOutOfBounds)
		}

		err := l.Flush()
		if err != nil {
			t.Fatalf("Flush failed: %v", err)
		}
	})

	t.Run("Read", func(t *testing.T) {
		l2, err := ReadList(rw, l.offset)
		if err != nil {
			t.Fatalf("ReadList failed: %v", err)
		}

		if l2.ElementSize() != l.ElementSize() {
			t.Errorf("got size %d, want %d", l2.ElementSize(), l.ElementSize())
		}

		l2.Len()
		if l2.Len() != l.Len() {
			t.Errorf("got len %d, want %d", l2.Len(), l.Len())
		}

		length := uint16(l2.Len())

		for i := uint16(0); i < length; i++ {
			buf, err := l2.Get(i)
			if err != nil {
				t.Fatalf("Get failed: %v", err)
			}

			actual := binary.BigEndian.Uint64(buf)
			if actual != uint64(i+1) {
				t.Errorf("%d: want %d, got %d", i, i+1, actual)
			}
		}
	})

	t.Run("Write after read", func(t *testing.T) {
		l3, err := ReadList(rw, l.offset)
		if err != nil {
			t.Fatalf("ReadList failed: %v", err)
		}

		originalLen := l3.Len()
		inserts := 20

		buf := make([]byte, l3.ElementSize())
		for i := uint16(0); i < uint16(inserts); i++ {
			binary.BigEndian.PutUint64(buf, uint64(i+1))
			l3.Append(buf)
		}

		if l3.Len() != originalLen+inserts {
			t.Errorf("got len %d, want %d", l3.Len(), originalLen+inserts)
		}
	})
}
