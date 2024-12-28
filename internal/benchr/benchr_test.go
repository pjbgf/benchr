package benchr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseBenchmarkData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		data string
		want map[string]map[string]float64
	}{
		{
			data: `
BenchmarkFoo-16    	       1	1332537337 ns/op	116206032 B/op	 1084268 allocs/op
BenchmarkBar-16     	  113340	     10377 ns/op	    2784 B/op	      42 allocs/op
`,
			want: map[string]map[string]float64{
				"BenchmarkFoo-16": {"allocs/op": 1084268, "ns/op": 1.332537337e+09},
				"BenchmarkBar-16": {"allocs/op": 42, "ns/op": 10377},
			},
		},
	}

	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			t.Parallel()

			got := parseBenchmarkData(tc.data)

			assert.Equal(t, tc.want, got)
		})
	}
}
