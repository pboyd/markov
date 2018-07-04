package disk

type ListItem struct {
	f      *File
	Offset int64
}

func newListItem(f *File, next int64, value []byte) (ListItem, error) {
	item := ListItem{f: f}

	var err error
	item.Offset, err = f.writeAddress(-1, next)
	if err != nil {
		return item, err
	}

	_, err = f.WriteBlob(-1, value)
	return item, err
}

func NewList(f *File, rootValue []byte) (ListItem, error) {
	return newListItem(f, -1, rootValue)
}

func (i ListItem) SetValue(value []byte) error {
	return i.f.OverwriteBlob(i.Offset+addressLength, value)
}

func (i ListItem) Value() ([]byte, error) {
	return i.f.ReadBlob(i.Offset + addressLength)
}

func (i ListItem) Next() (ListItem, error) {
	offset, err := i.f.readAddress(i.Offset)
	if err != nil {
		return ListItem{}, err
	}

	return ListItem{
		f:      i.f,
		Offset: offset,
	}, nil
}

// InsertAfter insert a new item into the list after the current item.
func (i ListItem) InsertAfter(value []byte) (ListItem, error) {
	oldNext, err := i.Next()
	if err != nil {
		return ListItem{}, err
	}

	newNext, err := newListItem(i.f, oldNext.Offset, value)
	if err != nil {
		return newNext, err
	}

	_, err = i.f.writeAddress(i.Offset, newNext.Offset)
	return newNext, err
}
