package contexts_test

import (
	"context"
	"testing"

	"github.com/brickingsoft/brick/pkg/contexts"
)

func TestMatch(t *testing.T) {
	c := context.Background()
	c1 := context.WithValue(c, "1", 1)
	c2 := context.WithValue(c1, "2", 2)
	c3 := context.WithValue(c, "3", 3)

	t.Log(contexts.Match(c1, c))
	t.Log(contexts.Match(c2, c))
	t.Log(contexts.Match(c2, c1))
	t.Log(contexts.Match(c3, c))
	t.Log(contexts.Match(c3, c1))
	t.Log(contexts.Match(c3, c2))
}

// BenchmarkMatch-20    	16645996	        70.65 ns/op	       0 b/op	       0 allocs/op
func BenchmarkMatch(b *testing.B) {
	c := context.Background()
	c1 := context.WithValue(c, "1", 1)
	c2 := context.WithValue(c1, "2", 2)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		contexts.Match(c2, c)
	}
}
