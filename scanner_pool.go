package recycled

import (
	"sync"
)

// ScannerPool may be used for pooling Scanners for structurally dissimilar JSON.
type ScannerPool struct {
	pool sync.Pool
	maxReuse int
}

// NewScannerPool enables JSON Scanner pooling for semi-structured JSON
//
// MaxReuse prevents a scanner from being returned to the pool after MaxReuse
// number of uses. This prevents scanner reuse from causing unbounded memory
// growth for structurally dissimilar JSON. 1,000 is probably a good number.
func NewScannerPool(maxReuse int) *ScannerPool {
	return &ScannerPool{
		sync.Pool{},
		maxReuse,
	}
}

// Get returns a Scanner from spr.
//
// The Scanner must be Put to spr after use.
func (spr *ScannerPool) Get() *Scanner {
	v := spr.pool.Get()
	if v == nil {
		return &Scanner{}
	}
	return v.(*Scanner)
}


// Put returns sr to spr.
//
// sr and objects recursively returned from sr cannot be used after sr
// is put into spr.
func (spr *ScannerPool) Put(sr *Scanner) {
	if sr.n > spr.maxReuse {
		return
	}
	spr.pool.Put(sr)
}
