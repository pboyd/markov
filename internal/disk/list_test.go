package disk

import (
	"bytes"
	"testing"
)

func TestList(t *testing.T) {
	rw, cleanup := tempFile(t)
	defer cleanup()

	f := NewFile(rw)

	valueA := make([]byte, 8)
	for i := range valueA {
		valueA[i] = 0xff
	}

	valueB := make([]byte, 8)
	for i := range valueB {
		valueB[i] = byte(i)
	}

	root, err := NewList(f, valueA)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}

	t.Run("value", func(t *testing.T) {
		actual, err := root.Value()
		if err != nil {
			t.Fatalf("got error: %v", err)
		}

		if !bytes.Equal(actual, valueA) {
			t.Fatalf("\ngot:  %v\nwant: %v", actual, valueA)
		}

		root.SetValue(valueB)
		actual, _ = root.Value()
		if !bytes.Equal(actual, valueB) {
			t.Errorf("\ngot:  %v\nwant: %v", actual, valueB)
		}
	})

	t.Run("InsertAfter", func(t *testing.T) {
		item, err := root.InsertAfter(valueA)
		if err != nil {
			t.Fatalf("got error: %v", err)
		}

		actual, _ := item.Value()
		if !bytes.Equal(actual, valueA) {
			t.Errorf("\ngot:  %v\nwant: %v", actual, valueA)
		}

		next, _ := root.Next()
		if next != item {
			t.Errorf("got item %d, want %d", next.Offset, item.Offset)
		}
	})
}
