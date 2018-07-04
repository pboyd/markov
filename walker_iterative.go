package markov

var _ Walker = &iterativeWalker{}

type iterativeWalker struct {
	chain Chain
	ids   []int
	next  int
}

func IterativeWalker(chain Chain) Walker {
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
