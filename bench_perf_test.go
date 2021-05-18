// +build !race

// The behavior of sync.Pool is not deterministic under race mode

package recycled

import (
	"fmt"
	"testing"
	"strings"

	"github.com/valyala/fastjson"
)

func BenchmarkParserPool(b *testing.B) {
	for _, n := range []int{0, 10, 100, 1000, 10000, 100000, 1e6} {
		b.Run(fmt.Sprintf("maxreuse_%d", n), func(b *testing.B) {
			benchmarkParserPool(b, n)
		})
	}
}

func benchmarkParserPool(b *testing.B, maxReuse int) {
	b.ReportAllocs()
	ppr := NewParserPool(maxReuse)
	var v *fastjson.Value
	benchPool(b, maxReuse, func(json []byte) {
		pr := ppr.Get()
		v, _ = pr.ParseBytes(json)
		ppr.Put(pr)
	})
	_ = v
}

func BenchmarkScannerPool(b *testing.B) {
	for _, n := range []int{0, 10, 100, 1000, 10000, 100000, 1e6} {
		b.Run(fmt.Sprintf("maxreuse_%d", n), func(b *testing.B) {
			benchmarkScannerPool(b, n)
		})
	}
}

func benchmarkScannerPool(b *testing.B, maxReuse int) {
	b.ReportAllocs()
	spr := NewScannerPool(maxReuse)
	var v *fastjson.Value
	benchPool(b, maxReuse, func(json []byte) {
		sr := spr.Get()
		sr.InitBytes(json)
		for sr.Next() {
			v = sr.Value()
		}
		spr.Put(sr)
	})
	_ = v
}

func benchPool(b *testing.B, maxReuse int, fn func([]byte)) {
	b.ReportAllocs()
	var json = []byte(`{"test":"` + strings.Repeat("*", 100000) + `"}`)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			fn(json)
		}
	})
	b.SetBytes(int64(len(json)))
}
