package markov

type Node struct {
	Value    interface{}
	Children []NodeProbability
}

type NodeProbability struct {
	*Node
	Probability float64
}
