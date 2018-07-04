package markov

/*
func (n *MemoryNode) Next() *MemoryNode {
	if len(n.children) == 0 {
		// This happens if the chain ends
		return nil
	}

	index := rand.Intn(n.sum())
	passed := 0

	for _, child := range n.children {
		passed += child.Count
		if passed > index {
			return child.MemoryNode
		}
	}

	panic("Next() failed")
}
*/
