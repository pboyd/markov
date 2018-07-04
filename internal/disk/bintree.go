package disk

import (
	"encoding/binary"
	"errors"
	"hash/fnv"
)

const (
	bstLength      = 32
	bstKeyOffset   = 0
	bstLeftOffset  = 8
	bstRightOffset = 16
	bstValueOffset = 24
)

var ErrDuplicate = errors.New("duplicate")

type BinaryTreeNode struct {
	f      *File
	buf    []byte
	Offset int64
}

func NewBinaryTree(f *File) *BinaryTreeNode {
	return &BinaryTreeNode{
		f:      f,
		buf:    make([]byte, bstLength),
		Offset: -1,
	}
}

func ReadBinaryTree(f *File, offset int64) (*BinaryTreeNode, error) {
	node := &BinaryTreeNode{
		f:      f,
		Offset: offset,
	}

	var err error
	node.buf, err = f.Read(offset, bstLength)
	return node, err
}

func (root *BinaryTreeNode) Search(key []byte) (*BinaryTreeNode, error) {
	return root.searchByHash(root.hashKey(key))
}

func (root *BinaryTreeNode) searchByHash(keyHash uint64) (*BinaryTreeNode, error) {
	if root.Offset < 0 {
		return nil, nil
	}

	rootKey := root.key()
	if rootKey == keyHash {
		return root, nil
	}

	var child *BinaryTreeNode
	var err error
	switch {
	case keyHash > rootKey:
		child, err = root.right()
	case keyHash < rootKey:
		child, err = root.left()
	}
	if err != nil {
		return nil, err
	}

	return child.searchByHash(keyHash)
}

// Insert inserts a new node on the tree and returns it.
//
// If the node already exists the existing node is returned along with
// `ErrDuplicate`.
func (root *BinaryTreeNode) Insert(key []byte, value int64) (*BinaryTreeNode, error) {
	keyHash := root.hashKey(key)
	return root.insertHash(keyHash, value)
}

func (root *BinaryTreeNode) insertHash(keyHash uint64, value int64) (*BinaryTreeNode, error) {
	if root.Offset < 0 {
		return root, root.create(keyHash, value)
	}

	rootKey := root.key()
	switch {
	case keyHash == rootKey:
		return root, ErrDuplicate
	case keyHash > rootKey:
		return root.insertChild(bstRightOffset, keyHash, value)
	default:
		return root.insertChild(bstLeftOffset, keyHash, value)
	}
}

func (root *BinaryTreeNode) insertChild(which int64, keyHash uint64, value int64) (*BinaryTreeNode, error) {
	// FIXME: What's the difference between "node" and "child"?
	node, err := root.child(which)
	if err != nil {
		return nil, err
	}

	isNew := node.Offset < 0

	child, err := node.insertHash(keyHash, value)
	if !isNew {
		return child, err
	}

	binary.BigEndian.PutUint64(root.buf[which:], uint64(node.Offset))
	return child, root.Write()
}

func (root *BinaryTreeNode) create(keyHash uint64, value int64) error {
	binary.BigEndian.PutUint64(root.buf[bstKeyOffset:], keyHash)
	binary.BigEndian.PutUint64(root.buf[bstLeftOffset:], ^uint64(0))
	binary.BigEndian.PutUint64(root.buf[bstRightOffset:], ^uint64(0))
	binary.BigEndian.PutUint64(root.buf[bstValueOffset:], uint64(value))

	return root.Write()
}

func (root *BinaryTreeNode) SetValue(value int64) {
	binary.BigEndian.PutUint64(root.buf[bstValueOffset:], uint64(value))
}

func (root *BinaryTreeNode) Value() int64 {
	return int64(binary.BigEndian.Uint64(root.buf[bstValueOffset:]))
}

func (root *BinaryTreeNode) key() uint64 {
	return binary.BigEndian.Uint64(root.buf[bstKeyOffset:])
}

func (root *BinaryTreeNode) left() (*BinaryTreeNode, error) {
	return root.child(bstLeftOffset)
}

func (root *BinaryTreeNode) right() (*BinaryTreeNode, error) {
	return root.child(bstRightOffset)
}

func (root *BinaryTreeNode) child(which int64) (*BinaryTreeNode, error) {
	offset := int64(binary.BigEndian.Uint64(root.buf[which:]))
	if offset < 0 {
		return NewBinaryTree(root.f), nil
	}
	return ReadBinaryTree(root.f, int64(offset))
}

func (root *BinaryTreeNode) Write() error {
	var err error
	root.Offset, err = root.f.Write(root.Offset, root.buf)
	return err
}

func (BinaryTreeNode) hashKey(buf []byte) uint64 {
	h := fnv.New64a()
	h.Write(buf)
	return h.Sum64()
}
