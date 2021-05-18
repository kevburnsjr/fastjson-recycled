package recycled

import (
	"github.com/valyala/fastjson"
)

// Scanner adds a counter to a Scanner for use with ScannerPool
type Scanner struct {
	fastjson.Scanner
	n int
}

// Init is a wrapper for Scanner.Init that also counts the number of calls.
func (sr *Scanner) Init(s string) {
	sr.n++
	sr.Scanner.Init(s)
}

// InitBytes is a wrapper for Scanner.InitBytes that also counts the number of calls.
func (sr *Scanner) InitBytes(b []byte) {
	sr.n++
	sr.Scanner.InitBytes(b)
}
