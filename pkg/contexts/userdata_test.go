package contexts_test

import (
	"context"
	"testing"

	"github.com/brickingsoft/brick/pkg/contexts"
)

// BenchmarkWrapUserdataContext-20    	1000000000	         0.5448 ns/op	       0 b/op	       0 allocs/op
func BenchmarkWrapUserdataContext(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	uc := contexts.WrapUserdataContext(ctx)
	ctx = uc
	b.ReportAllocs()
	b.ResetTimer()
	ko := float64(0)
	for i := 0; i < b.N; i++ {
		if _, ok := uc.(contexts.UserdataContext); !ok {
			ko++
			b.ReportMetric(ko, "ko")
		}
	}
}
