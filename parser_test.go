// +build !race

// The behavior of sync.Pool is not deterministic under race mode

package recycled

import (
	"fmt"
	"testing"
	"sync"

	"github.com/valyala/fastjson"
)

func TestParserPool(t *testing.T) {
	var news int
	ppr := &ParserPool{
		sync.Pool{New: func() interface{} { news++; return new(Parser) }},
		100,
	}
	var v *fastjson.Value
	var v2 *fastjson.Value
	for i := 333; i > 0; i-- {
		var json = fmt.Sprintf(`{"%d":"test"}`, i)
		pr := ppr.Get()
		v, _ = pr.Parse(json)
		v2, _ = pr.ParseBytes([]byte(json))
		ppr.Put(pr)
	}
	if news != 7 {
		t.Fatalf("Expected exactly 7 calls to Put (not %d)", news)
	}
	ppr = NewParserPool(10)
	if ppr.maxReuse != 10 {
		t.Fatalf("Expected maxReuse to be 10 (not %d)", ppr.maxReuse)
	}
	pr := ppr.Get()
	_ = pr
	_ = v
	_ = v2
}
