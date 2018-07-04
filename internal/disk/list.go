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

func ReadList(f *File, offset int64) *ListItem {
	return &ListItem{f, offset}
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

func (i ListItem) Next() (*ListItem, error) {
	nextOffset, err := i.f.readAddress(i.Offset)
	if err != nil {
		return nil, err
	}

	if nextOffset < 0 {
		return nil, nil
	}

	return &ListItem{
		f:      i.f,
		Offset: nextOffset,
	}, nil
}

// InsertAfter insert a new item into the list after the current item.
func (i ListItem) InsertAfter(value []byte) (ListItem, error) {
	oldNext, err := i.Next()
	if err != nil {
		return ListItem{}, err
	}

	var nextOffset int64 = -1
	if oldNext != nil {
		nextOffset = oldNext.Offset
	}

	newNext, err := newListItem(i.f, nextOffset, value)
	if err != nil {
		return newNext, err
	}

	_, err = i.f.writeAddress(i.Offset, newNext.Offset)
	return newNext, err
}
