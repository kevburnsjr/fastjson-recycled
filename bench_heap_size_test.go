// +build !race

// The behavior of sync.Pool is not deterministic under race mode

package recycled

import (
	"fmt"
	// "log"
	"math"
	"testing"
	"runtime"
	"strings"

	"github.com/valyala/fastjson"
	"github.com/valyala/fastrand"
)

func BenchmarkHeapSizeParserPool(b *testing.B) {
	for _, n := range []int{0, 10, 100, 1000, 10000, 100000, 1e6} {
		b.Run(fmt.Sprintf("maxreuse_%d", n), func(b *testing.B) {
			benchmarkHeapSizeParserPool(b, n)
		})
	}
}

func benchmarkHeapSizeParserPool(b *testing.B, maxReuse int) {
	ppr := NewParserPool(maxReuse)
	var v *fastjson.Value
	benchmarkHeapSize(b, maxReuse, func(json []byte) {
		pr := ppr.Get()
		v, _ = pr.ParseBytes(json)
		ppr.Put(pr)
	})
	_ = v
}


func BenchmarkHeapSizeScannerPool(b *testing.B) {
	for _, n := range []int{1e8} {
		b.Run(fmt.Sprintf("maxreuse_%d", n), func(b *testing.B) {
			benchmarkHeapSizeScannerPool(b, n)
		})
	}
}

func benchmarkHeapSizeScannerPool(b *testing.B, maxReuse int) {
	spr := NewScannerPool(maxReuse)
	var v *fastjson.Value
	benchmarkHeapSize(b, maxReuse, func(json []byte) {
		sr := spr.Get()
		sr.InitBytes(json)
		for sr.Next() {
			v = sr.Value()
		}
		spr.Put(sr)
	})
	_ = v
}


func benchmarkHeapSize(b *testing.B, maxReuse int, fn func([]byte)) {
	b.ReportAllocs()
	var total int
	var size int
	var rng fastrand.RNG
	var before runtime.MemStats
	var after runtime.MemStats
	var json = make([][]byte, 100)
	for i := 0; i < len(json); i++ {
		json[i] = []byte(`{"test":"` + strings.Repeat("*", 100 + int(math.Floor(math.Pow(10, 6-math.Log2(float64(i+1)))))) + `"}`)
	}
	var i int
	runtime.GC()
	runtime.GC()
	runtime.ReadMemStats(&before)
	for i = 0; i < b.N; i++ {
		var j = int(rng.Uint32n(uint32(len(json))))
		size += len(json[j])
		total++
		fn(json[j])
	}
	runtime.ReadMemStats(&after)
	b.ReportMetric(float64(after.HeapInuse - before.HeapInuse), "heap-sz")
	b.SetBytes(int64(size / total))
}
