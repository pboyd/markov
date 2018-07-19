package markov

import "sync"

// Feed reads values from the channels and writes them to the WriteChain.
//
// Blocks until all the channels have been closed.
//
// If the WriteChain returns an error Feed returns it immediately, leaving
// unread values on the channels.
func Feed(wc WriteChain, channels ...<-chan interface{}) error {
	var wg sync.WaitGroup
	wg.Add(len(channels))

	cancel := make(chan struct{})
	errChannel := make(chan error, len(channels))

	for _, ch := range channels {
		go func(values <-chan interface{}) {
			defer wg.Done()
			errChannel <- feedOne(cancel, wc, values)
		}(ch)
	}

	wg.Wait()

	close(errChannel)
	for err := range errChannel {
		return err
	}

	return nil
}

func feedOne(cancel chan struct{}, wc WriteChain, values <-chan interface{}) error {
	var next int
	var err error

	last, err := wc.Add(<-values)
	if err != nil {
		return err
	}

	for {
		select {
		case <-cancel:
			return nil
		case val, ok := <-values:
			if !ok {
				return nil
			}

			next, err = wc.Add(val)
			if err != nil {
				return err
			}

			err = wc.Relate(last, next, 1)
			if err != nil {
				return err
			}

			last = next
		}
	}
}
