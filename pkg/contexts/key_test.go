package contexts_test

import (
	"testing"

	"github.com/brickingsoft/brick/pkg/contexts"
)

// BenchmarkKey-20    	1000000000	         0.1119 ns/op	       0 b/op	       0 allocs/op
func BenchmarkKey(b *testing.B) {
	k1 := contexts.Key{Name: "sss"}
	k2 := k1
	b.ReportAllocs()
	b.ResetTimer()
	ko := float64(0)
	for i := 0; i < b.N; i++ {
		if k1 == k2 {
			continue
		} else {
			ko++
			b.ReportMetric(ko, "ok")
		}
	}
}
