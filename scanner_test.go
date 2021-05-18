// +build !race

// The behavior of sync.Pool is not deterministic under race mode

package recycled

import (
	"fmt"
	"testing"
	"sync"

	"github.com/valyala/fastjson"
)

func TestScannerPool(t *testing.T) {
	var news int
	spr := &ScannerPool{
		sync.Pool{New: func() interface{} { news++; return new(Scanner) }},
		100,
	}
	var v *fastjson.Value
	var v2 *fastjson.Value
	for i := 333; i > 0; i-- {
		var json = fmt.Sprintf(`{"%d":"test"}`, i)
		sr := spr.Get()
		sr.Init(json)
		v = sr.Value()
		sr.InitBytes([]byte(json))
		v2 = sr.Value()
		spr.Put(sr)
	}
	if news != 7 {
		t.Fatalf("Expected exactly 7 calls to Put (not %d)", news)
	}
	spr = NewScannerPool(10)
	if spr.maxReuse != 10 {
		t.Fatalf("Expected maxReuse to be 10 (not %d)", spr.maxReuse)
	}
	sr := spr.Get()
	_ = sr
	_ = v
	_ = v2
}
