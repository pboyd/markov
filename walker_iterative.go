package markov

var _ Walker = &iterativeWalker{}
var _ Walker = &iterativeChainWalker{}

// IterativeChain is a chain that can be iterated efficiently.
type IterativeChain interface {
	Chain

	// Next returns the ID of the entry which follows the input ID.
	//
	// If the ID is at the end of the chain, ErrBrokenChain will be
	// returned.
	Next(id int) (int, error)
}

type iterativeWalker struct {
	chain Chain
	ids   []int
	next  int
}

// IterativeWalker returns a Walker that traverses every item in the Chain.
//
// If the chain implements IterativeChain, it will be used.
func IterativeWalker(chain Chain) Walker {
	if iw, ok := chain.(IterativeChain); ok {
		return &iterativeChainWalker{
			last:  -1,
			chain: iw,
		}
	}

	return &iterativeWalker{
		chain: chain,
	}
}

func (w *iterativeWalker) Next() (interface{}, error) {
	if w.ids == nil {
		var err error
		w.ids, err = w.allIDs()
		if err != nil {
			return nil, err
		}
	}

	if w.next >= len(w.ids) {
		return nil, ErrBrokenChain
	}

	v, err := w.chain.Get(w.ids[w.next])
	w.next++
	return v, err
}

func (w *iterativeWalker) allIDs() ([]int, error) {
	idMap := map[int]struct{}{}
	err := w.subtreeIDs(0, idMap)
	if err != nil {
		return nil, err
	}

	ids := make([]int, 0, len(idMap))
	for id := range idMap {
		ids = append(ids, id)
	}

	return ids, nil
}

func (w *iterativeWalker) subtreeIDs(root int, ids map[int]struct{}) error {
	links, err := w.chain.Links(root)
	if err != nil {
		return err
	}

	for _, link := range links {
		if _, ok := ids[link.ID]; ok {
			continue
		}

		ids[link.ID] = struct{}{}
		err = w.subtreeIDs(link.ID, ids)
		if err != nil {
			return err
		}
	}

	return nil
}

type iterativeChainWalker struct {
	chain IterativeChain
	last  int
}

func (w *iterativeChainWalker) Next() (interface{}, error) {
	id := 0
	if w.last >= 0 {
		var err error
		id, err = w.chain.Next(w.last)
		if err != nil {
			return nil, err
		}
	}

	w.last = id

	return w.chain.Get(id)
}
