package recycled

import (
	"github.com/valyala/fastjson"
)

// Parser adds a counter to a fastjson Parser
type Parser struct {
	fastjson.Parser
	n int
}

// Parse is a wrapper for Parser.Parse that also counts the number of calls.
func (pr *Parser) Parse(s string) (*fastjson.Value, error) {
	pr.n++
	return pr.Parser.Parse(s)
}

// ParseBytes is a wrapper for Parser.ParseBytes that also counts the number of calls.
func (pr *Parser) ParseBytes(b []byte) (*fastjson.Value, error) {
	pr.n++
	return pr.Parser.ParseBytes(b)
}
