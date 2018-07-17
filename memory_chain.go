package markov

import "sync"

var _ Chain = &MemoryChain{}

type MemoryChain struct {
	mu         sync.RWMutex
	valueIndex map[interface{}]int
	values     []interface{}
	links      []linkCountSlice
}

func NewMemoryChain(cap int) *MemoryChain {
	return &MemoryChain{
		valueIndex: make(map[interface{}]int, cap),
		values:     make([]interface{}, 0, cap),
		links:      make([]linkCountSlice, 0, cap),
	}
}

func (c *MemoryChain) Get(id int) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if id >= len(c.values) {
		return nil, nil
	}

	return c.values[id], nil
}

func (c *MemoryChain) Links(id int) ([]Link, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if id >= len(c.links) {
		return nil, ErrNotFound
	}

	return c.links[id].LinkSlice(), nil
}

func (c *MemoryChain) Find(value interface{}) (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	id, ok := c.valueIndex[value]

	if !ok {
		return 0, ErrNotFound
	}

	return id, nil
}

func (c *MemoryChain) Add(value interface{}) (int, error) {
	existing, err := c.Find(value)
	if err == nil {
		return existing, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.values = append(c.values, value)
	c.links = append(c.links, make(linkCountSlice, 0, 1))

	if c.valueIndex == nil {
		c.valueIndex = map[interface{}]int{}
	}

	id := len(c.values) - 1
	c.valueIndex[value] = id

	return id, nil
}

func (c *MemoryChain) Relate(parent, child int, delta int) error {
	childIndex := c.links[parent].Find(child)
	if childIndex < 0 {
		c.links[parent] = append(c.links[parent], linkCount{ID: child})
		childIndex = len(c.links[parent]) - 1
	}

	c.links[parent][childIndex].Count += delta

	return nil
}

func (c *MemoryChain) Next(last int) (int, error) {
	c.mu.RLock()
	length := len(c.values)
	c.mu.RUnlock()

	if last >= length-1 {
		return 0, ErrBrokenChain
	}
	return last + 1, nil
}

func (c *MemoryChain) linkCounts(id int) (linkCountSlice, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.links[id], nil
}

type linkCount struct {
	ID    int
	Count int
}

func (l *linkCount) Link(total int) Link {
	return Link{
		ID:          l.ID,
		Probability: float64(l.Count) / float64(total),
	}
}

type linkCountSlice []linkCount

func (ls linkCountSlice) sum() int {
	total := 0
	for _, l := range ls {
		total += l.Count
	}
	return total
}

func (ls linkCountSlice) LinkSlice() []Link {
	total := ls.sum()

	links := make([]Link, len(ls))
	for i, l := range ls {
		links[i] = l.Link(total)
	}

	return links
}

func (ls linkCountSlice) Find(id int) int {
	for i, l := range ls {
		if l.ID == id {
			return i
		}
	}
	return -1
}
