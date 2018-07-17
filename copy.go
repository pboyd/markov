package markov

import "math"

type CopyFrom interface {
	CopyFrom(src Chain) error
}

// Copy copies src to dest.
//
// If dest implements CopyFrom it will be used.
func Copy(dest WriteChain, src Chain) error {
	if cf, ok := dest.(CopyFrom); ok {
		return cf.CopyFrom(src)
	}

	valueToDestID := make(map[interface{}]int)
	srcIDtoDestID := make(map[int]int)

	walker := IterativeWalker(src)
	for {
		value, err := walker.Next()
		if err != nil {
			if err == ErrBrokenChain {
				break
			}

			return err
		}

		destID, err := dest.Add(value)
		if err != nil {
			return err
		}

		valueToDestID[value] = destID

		srcID, err := src.Find(value)
		if err != nil {
			return err
		}

		srcIDtoDestID[srcID] = destID
	}

	for value, destID := range valueToDestID {
		srcID, err := src.Find(value)
		if err != nil {
			return err
		}

		linkCounts, err := linkCounts(src, srcID)
		if err != nil {
			return err
		}

		for _, link := range linkCounts {
			err = dest.Relate(destID, srcIDtoDestID[link.ID], link.Count)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type linkCountChain interface {
	linkCounts(id int) (linkCountSlice, error)
}

func linkCounts(c Chain, id int) (linkCountSlice, error) {
	if lcc, ok := c.(linkCountChain); ok {
		return lcc.linkCounts(id)
	}

	links, err := c.Links(id)
	if err != nil {
		return nil, err
	}

	lcs := make(linkCountSlice, len(links))

	for i, link := range links {
		var count int

		if link.Probability < 1 {
			// We need a count, but Links returns a percentage, so
			// take the percentage of a big number.
			count = int(link.Probability * float64(math.MaxUint32))
		} else {
			count = math.MaxUint32
		}

		lcs[i] = linkCount{
			ID:    link.ID,
			Count: count,
		}
	}

	return lcs, nil
}
