package markov

import (
	"math"
	"testing"
)

func TestMarshalValue(t *testing.T) {
	cases := []interface{}{
		'a', 'â', '∈', '⚄', '💾',
		"a", "aa", "some words", "aâ∈⚄💾'",
		uint64(0), uint64(1), uint64(1 << 32), uint64(1<<64 - 1),
		int64(0), int64(1), int64(-1), int64(1 << 32), int64(1<<63 - 1), int64(^(1<<63 - 1) + 1),
		uint(0), uint(1), uint(1 << 32), uint(1<<64 - 1),
		0, 1, -1, 1 << 32, 1<<63 - 1, ^(1<<63 - 1) + 1,
		uint32(0), uint32(1), uint32(1<<32 - 1),
		int32(0), int32(1), int32(-1), int32(1<<31 - 1), int32(^(1<<31 - 1) + 1),
		float32(0), float32(1), float32(-1), float32(math.MaxFloat32), float32(math.SmallestNonzeroFloat32),
		0.0, 1.0, -1.0, math.MaxFloat64, math.SmallestNonzeroFloat64,
	}

	for _, v1 := range cases {
		buf, err := marshalValue(v1)
		if err != nil {
			t.Errorf("%v: got MarshalValue error %v", v1, err)
			continue
		}

		v2, err := unmarshalValue(buf)
		if err != nil {
			t.Errorf("%v: got UnmarshalValue error %v", 11, err)
			continue
		}

		if v2 != v1 {
			t.Errorf("got %v, want %v", v2, v1)
		}
	}
}
