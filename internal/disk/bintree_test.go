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
		n, err := root.Insert(c.key, c.value)
		if err != nil {
			t.Fatalf("got error %v, want nil", err)
		}

		actual := n.Value()
		if actual != c.value {
			t.Errorf("got %d, want %d", actual, c.value)
		}
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

		value := node.Value()
		if value != c.value {
			t.Errorf("got %d, want %d", value, c.value)
		}
	}
}
