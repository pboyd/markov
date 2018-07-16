package markov

import "math"

func Copy(dest WriteChain, src Chain) error {
	destValueToID := make(map[interface{}]int)
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

		destValueToID[value] = destID

		srcID, err := src.Find(value)
		if err != nil {
			return err
		}

		srcIDtoDestID[srcID] = destID
	}

	walker = IterativeWalker(src)
	for {
		value, err := walker.Next()
		if err != nil {
			if err == ErrBrokenChain {
				break
			}

			return err
		}

		srcID, err := src.Find(value)
		if err != nil {
			return err
		}

		links, err := src.Links(srcID)
		if err != nil {
			return err
		}

		destID := destValueToID[value]

		for _, link := range links {
			var count int

			if link.Probability < 1 {
				// Relate takes a count, but Links returns a
				// percentage, so take the percentage of a big
				// number.
				count = int(link.Probability * float64(math.MaxUint32))
			} else {
				count = math.MaxUint32
			}
			err = dest.Relate(destID, srcIDtoDestID[link.ID], count)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
