package disk

import "testing"

func TestBinTree(t *testing.T) {
	cases := []struct {
		key   []byte
		value int64
	}{
		{[]byte("foo"), 1},
		{[]byte("bar"), 2},
		{[]byte("baz"), 3},
		{[]byte("qux"), 4},
	}

	rw, cleanup := tempFile(t)
	defer cleanup()
	f := NewFile(rw)

	root := NewBinaryTree(f)
	for _, c := range cases {
		root.Insert(c.key, c.value)
	}

	for _, c := range cases {
		node, err := root.Search(c.key)
		if err != nil {
			t.Errorf("got error %v, want nil", err)
		}

		if node == nil {
			t.Errorf("got <nil>, want %d", c.value)
			continue
		}

		value, err := node.Value()
		if err != nil {
			t.Errorf("got error %v, want nil", err)
		}

		if value != c.value {
			t.Errorf("got %d, want %d", value, c.value)
		}
	}
}
