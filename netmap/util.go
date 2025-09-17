package netmap

import (
	"strconv"

	"github.com/antlr4-go/antlr/v4"
)

func (p *policyVisitor) parseUint32Token(dst *uint32, token antlr.Token) bool {
	n, err := strconv.ParseUint(token.GetText(), 10, 32)
	if err != nil {
		p.reportError(errInvalidNumber)
		return false
	}

	*dst = uint32(n)

	return true
}
