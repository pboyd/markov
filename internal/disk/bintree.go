package disk

import (
	"encoding/binary"
	"hash/fnv"
)

const (
	bstLength      = 32
	bstKeyOffset   = 0
	bstLeftOffset  = 8
	bstRightOffset = 16
	bstValueOffset = 24
)

type BinaryTreeNode struct {
	f      *File
	Offset int64
}

func NewBinaryTree(f *File) *BinaryTreeNode {
	return &BinaryTreeNode{
		f:      f,
		Offset: -1,
	}
}

func ReadBinaryTree(f *File, offset int64) *BinaryTreeNode {
	return &BinaryTreeNode{
		f:      f,
		Offset: offset,
	}
}

func (root *BinaryTreeNode) Search(key []byte) (*BinaryTreeNode, error) {
	return root.searchByHash(root.hashKey(key))
}

func (root *BinaryTreeNode) searchByHash(keyHash uint64) (*BinaryTreeNode, error) {
	if root.Offset < 0 {
		return nil, nil
	}

	rootKey, err := root.key()
	if err != nil {
		return nil, err
	}

	if rootKey == keyHash {
		return root, nil
	}

	child := &BinaryTreeNode{f: root.f}
	switch {
	case keyHash > rootKey:
		child.Offset, err = root.right()
	case keyHash < rootKey:
		child.Offset, err = root.left()
	}
	if err != nil {
		return nil, err
	}

	return child.searchByHash(keyHash)
}

func (root *BinaryTreeNode) Insert(key []byte, value int64) (*BinaryTreeNode, error) {
	keyHash := root.hashKey(key)
	return root.insertHash(keyHash, value)
}

func (root *BinaryTreeNode) insertHash(keyHash uint64, value int64) (*BinaryTreeNode, error) {
	if root.Offset < 0 {
		return root, root.create(keyHash, value)
	}

	rootKey, err := root.key()
	if err != nil {
		return nil, err
	}

	switch {
	case keyHash == rootKey:
		return root, root.SetValue(value)
	case keyHash > rootKey:
		return root.insertChild(bstRightOffset, keyHash, value)
	default:
		return root.insertChild(bstLeftOffset, keyHash, value)
	}
}

func (root *BinaryTreeNode) insertChild(refOffset int64, keyHash uint64, value int64) (*BinaryTreeNode, error) {
	address, err := root.f.readAddress(root.Offset + refOffset)
	if err != nil {
		return nil, err
	}

	node := &BinaryTreeNode{
		f:      root.f,
		Offset: address,
	}

	child, err := node.insertHash(keyHash, value)
	if err != nil {
		return nil, err
	}

	if address < 0 {
		_, err = root.f.writeAddress(root.Offset+refOffset, node.Offset)
	}

	return child, err
}

func (root *BinaryTreeNode) create(keyHash uint64, value int64) error {
	buf := make([]byte, bstLength)
	binary.BigEndian.PutUint64(buf[bstKeyOffset:], keyHash)
	binary.BigEndian.PutUint64(buf[bstLeftOffset:], ^uint64(0))
	binary.BigEndian.PutUint64(buf[bstRightOffset:], ^uint64(0))
	binary.BigEndian.PutUint64(buf[bstValueOffset:], uint64(value))

	var err error
	root.Offset, err = root.f.Write(-1, buf)
	return err
}

func (root *BinaryTreeNode) SetValue(value int64) error {
	_, err := root.f.writeAddress(root.Offset+bstValueOffset, value)
	return err
}

func (root *BinaryTreeNode) Value() (int64, error) {
	return root.f.readAddress(root.Offset + bstValueOffset)
}

func (root *BinaryTreeNode) key() (uint64, error) {
	buf, err := root.f.Read(root.Offset+bstKeyOffset, 8)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(buf), nil
}

func (root *BinaryTreeNode) left() (int64, error) {
	return root.f.readAddress(root.Offset + bstLeftOffset)
}

func (root *BinaryTreeNode) right() (int64, error) {
	return root.f.readAddress(root.Offset + bstRightOffset)
}

func (BinaryTreeNode) hashKey(buf []byte) uint64 {
	h := fnv.New64a()
	h.Write(buf)
	return h.Sum64()
}
