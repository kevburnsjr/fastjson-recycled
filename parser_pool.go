package recycled

import (
	"sync"
)

// ParserPool may be used for pooling Parsers for structurally dissimilar JSON.
type ParserPool struct {
	pool sync.Pool
	maxReuse int
}

// NewParserPool enables JSON Parser pooling for semi-structured JSON
//
// MaxReuse prevents a parser from being returned to the pool after MaxReuse
// number of uses. This prevents parser reuse from causing unbounded memory
// growth for structurally dissimilar JSON. 1,000 is probably a good number.
func NewParserPool(maxReuse int) *ParserPool {
	return &ParserPool{
		sync.Pool{},
		maxReuse,
	}
}

// Get returns a Parser from ppr.
//
// The Parser must be Put to ppr after use.
func (ppr *ParserPool) Get() *Parser {
	v := ppr.pool.Get()
	if v == nil {
		return &Parser{}
	}
	return v.(*Parser)
}

// Put returns pr to ppr.
//
// pr and objects recursively returned from pr cannot be used after pr
// is put into ppr.
func (ppr *ParserPool) Put(pr *Parser) {
	if pr.n > ppr.maxReuse {
		return
	}
	ppr.pool.Put(pr)
}
