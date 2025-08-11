package endpoints_test

import (
	"strings"
	"sync"
	"testing"
)

func do(s string) string {
	return strings.TrimSpace(s)
}

type fn func(string) string

var (
	values = map[string]fn{
		"a": do, "b": do, "c": do, "d": do, "e": do, "f": do, "g": do,
	}
)

// BenchmarkMap-20    	18563149	        65.09 ns/op	       0 B/op	       0 allocs/op
func BenchmarkMap(b *testing.B) {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, k := range keys {
			v := values[k]
			v(k)
		}
	}
}

type KeyValue struct {
	key string
	val fn
}

// BenchmarkSlice-20    	 9161850	       126.2 ns/op	       0 B/op	       0 allocs/op
func BenchmarkSlice(b *testing.B) {
	keys := make([]string, 0, len(values))
	kvs := make([]KeyValue, 0, len(values))
	for k, v := range values {
		keys = append(keys, k)
		kvs = append(kvs, KeyValue{k, v})
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, k := range keys {
			for _, kv := range kvs {
				if kv.key == k {
					kv.val(kv.key)
				}
			}
		}
	}
}

// BenchmarkOnce-20    	42192163	        28.08 ns/op	       0 B/op	       0 allocs/op
func BenchmarkOnce(b *testing.B) {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	b.ReportAllocs()
	b.ResetTimer()
	var (
		af   fn = nil
		bf   fn = nil
		cf   fn = nil
		df   fn = nil
		ef   fn = nil
		ff   fn = nil
		gf   fn = nil
		once sync.Once
	)

	for i := 0; i < b.N; i++ {
		once.Do(func() {
			af = values["a"]
			bf = values["b"]
			cf = values["c"]
			df = values["d"]
			ef = values["e"]
			ff = values["f"]
			gf = values["g"]
		})
		for _, k := range keys {
			switch k {
			case "a":
				af(k)
				break
			case "b":
				bf(k)
				break
			case "c":
				cf(k)
				break
			case "d":
				df(k)
				break
			case "e":
				ef(k)
				break
			case "f":
				ff(k)
				break
			case "g":
				gf(k)
				break
			default:
				break
			}
		}
	}
}
