package markov

import "sync"

func Feed(wc WriteChain, chans ...<-chan interface{}) {
	var wg sync.WaitGroup
	wg.Add(len(chans))

	for _, ch := range chans {
		go func(values <-chan interface{}) {
			defer wg.Done()
			feedOne(wc, values)
		}(ch)
	}

	wg.Wait()
}

func feedOne(wc WriteChain, values <-chan interface{}) {
	var next int
	var err error

	last, err := wc.Add(<-values)
	if err != nil {
		// FIXME
		panic(err)
	}

	for val := range values {
		next, err = wc.Add(val)
		if err != nil {
			// FIXME
			panic(err)
		}

		err = wc.Relate(last, next, 1)
		if err != nil {
			// FIXME
			panic(err)
		}

		last = next
	}
}
