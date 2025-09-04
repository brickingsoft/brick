package logs_test

import (
	"context"
	"testing"

	"github.com/brickingsoft/brick/pkg/mosses"
	"github.com/brickingsoft/brick/rpc/logs"
)

func TestInfo(t *testing.T) {
	log, logErr := mosses.New()
	if logErr != nil {
		t.Fatal(logErr)
	}
	defer log.Close()

	ctx := context.Background()
	ctx = logs.With(ctx, log)

	logs.Info(ctx, "hello")
}

// BenchmarkContext-20    	67085203	        17.73 ns/op	      16 B/op	       1 allocs/op
func BenchmarkContext(b *testing.B) {
	log, logErr := mosses.New()
	if logErr != nil {
		b.Fatal(logErr)
	}
	defer log.Close()

	ctx := context.Background()
	ctx = logs.With(ctx, log)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log = logs.Load(ctx)
	}
}
