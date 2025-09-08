package netmap

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/nspcc-dev/neofs-sdk-go/netmap/parser"
	protonetmap "github.com/nspcc-dev/neofs-sdk-go/proto/netmap"
)

// ECRule represents erasure coding rule for encoding of payload of container
// objects and their placement.
type ECRule struct {
	selector      string
	dataPartNum   uint32
	parityPartNum uint32
}

// NewECRules constructs new ECRule instance.
func NewECRule(dataPartNum, parityPartNum uint32) ECRule {
	return ECRule{
		dataPartNum:   dataPartNum,
		parityPartNum: parityPartNum,
	}
}

// verify checks whether x follows NeoFS API requirements.
func (x ECRule) verify() error {
	if x.dataPartNum == 0 {
		return errors.New("zero data part num")
	}

	if x.parityPartNum == 0 {
		return errors.New("zero parity part num")
	}

	if x.dataPartNum > maxTotalECParts || x.parityPartNum > maxTotalECParts ||
		x.dataPartNum+x.parityPartNum > maxTotalECParts {
		return fmt.Errorf("more than %d total parts", maxTotalECParts)
	}

	return nil
}

// fromProtoMessage validates m according to the NeoFS API protocol and restores
// x from it.
func (x *ECRule) fromProtoMessage(m *protonetmap.PlacementPolicy_ECRule) error {
	x.dataPartNum = m.DataPartNum
	x.parityPartNum = m.ParityPartNum
	x.selector = m.Selector

	return x.verify()
}

// protoMessage converts x into message to transmit using the NeoFS API
// protocol.
func (x ECRule) protoMessage() *protonetmap.PlacementPolicy_ECRule {
	return &protonetmap.PlacementPolicy_ECRule{
		DataPartNum:   x.dataPartNum,
		ParityPartNum: x.parityPartNum,
		Selector:      x.selector,
	}
}

// SetDataPartNum sets number of data parts payload is split into.
func (x *ECRule) SetDataPartNum(n uint32) {
	x.dataPartNum = n
}

// DataPartNum returns number of data parts payload is split into.
func (x ECRule) DataPartNum() uint32 {
	return x.dataPartNum
}

// SetParityPartNum sets number of parity parts payload is split into.
func (x *ECRule) SetParityPartNum(n uint32) {
	x.parityPartNum = n
}

// ParityPartNum returns number of parity parts payload is split into.
func (x ECRule) ParityPartNum() uint32 {
	return x.parityPartNum
}

// SetSelectorName sets name of the corresponding [Selector].
//
// Empty (default) value specifies selection from all possible nodes to store
// the object.
func (x *ECRule) SetSelectorName(s string) {
	x.selector = s
}

// SelectorName returns name of the corresponding [Selector].
func (x ECRule) SelectorName() string {
	return x.selector
}

// SetECRules sets list of EC rules.
func (p *PlacementPolicy) SetECRules(rules []ECRule) {
	p.ecRules = rules
}

// ECRules returns list of EC rules.
func (p PlacementPolicy) ECRules() []ECRule {
	return p.ecRules
}

// VisitEcStmt implements [parser.QueryVisitor] interface.
func (p *policyVisitor) VisitEcStmt(ctx *parser.EcStmtContext) any {
	dataPartNum, err := strconv.ParseUint(ctx.DataPartNum.GetText(), 10, 32)
	if err != nil {
		return p.reportError(errInvalidNumber)
	}
	parityPartNum, err := strconv.ParseUint(ctx.ParityPartNum.GetText(), 10, 32)
	if err != nil {
		return p.reportError(errInvalidNumber)
	}

	var res protonetmap.PlacementPolicy_ECRule
	if sel := ctx.GetSelector(); sel != nil {
		res.Selector = sel.GetText()
	}

	res.DataPartNum = uint32(dataPartNum)
	res.ParityPartNum = uint32(parityPartNum)

	return &res
}
