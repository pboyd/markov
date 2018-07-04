package markov

/*
func BenchmarkBuildUnique(b *testing.B) {
	numbers := normalDistGenerator(b.N, b.N*2)
	NewBuilder(0).Feed(numbers)
}

func BenchmarkBuildDuplicate(b *testing.B) {
	numbers := normalDistGenerator(b.N, 10)
	NewBuilder(0).Feed(numbers)
}

func BenchmarkNext(b *testing.B) {
	builder := NewBuilder(0)
	builder.Feed(normalDistGenerator(b.N, b.N/4))
	node := builder.Root()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		node = node.Next()
	}
}

func normalDistGenerator(count, stddev int) <-chan interface{} {
	numbers := make(chan interface{})

	go func() {
		defer close(numbers)

		for i := 0; i < count; i++ {
			numbers <- int(rand.NormFloat64() * float64(stddev))
		}
		numbers <- 0
	}()

	return numbers
}
*/
