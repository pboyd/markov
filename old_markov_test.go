package markov

/*
func TestNext(t *testing.T) {
	const iterations = 100000

	root := testModel()

	counts := map[rune]int{}

	for i := 0; i < iterations; i++ {
		next := root.Next()
		counts[next.Value.(rune)]++
	}

	for v, p := range root.Probabilities() {
		actual := counts[v.(rune)]
		xp := int(p * float64(iterations))
		if !fuzzyEquals(actual, xp, 0.1) {
			t.Errorf("%q: got %d, want ~%d", v, actual, xp)
		}
	}
}

func fuzzyEquals(a, b int, tolerance float64) bool {
	return math.Abs((float64(a)/float64(b))-1) < tolerance
}
*/
