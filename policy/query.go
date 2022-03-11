package policy

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/policy/parser"
)

var (
	// ErrInvalidNumber is returned when a value of SELECT is 0.
	ErrInvalidNumber = errors.New("policy: expected positive integer")
	// ErrUnknownClause is returned when a statement(clause) in a query is unknown.
	ErrUnknownClause = errors.New("policy: unknown clause")
	// ErrUnknownOp is returned when an operation in a query is unknown.
	ErrUnknownOp = errors.New("policy: unknown operation")
	// ErrUnknownFilter is returned when a value of FROM in a query is unknown.
	ErrUnknownFilter = errors.New("policy: filter not found")
	// ErrUnknownSelector is returned when a value of IN is unknown.
	ErrUnknownSelector = errors.New("policy: selector not found")
	// ErrSyntaxError is returned for errors found by ANTLR parser.
	ErrSyntaxError = errors.New("policy: syntax error")
)

type policyVisitor struct {
	errors []error
	parser.BaseQueryVisitor
	antlr.DefaultErrorListener
}

// Parse parses s into a placement policy.
func Parse(s string) (*netmap.PlacementPolicy, error) {
	return parse(s)
}

func newPolicyVisitor() *policyVisitor {
	return &policyVisitor{}
}

func parse(s string) (*netmap.PlacementPolicy, error) {
	input := antlr.NewInputStream(s)
	lexer := parser.NewQueryLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, 0)

	p := parser.NewQuery(stream)
	p.BuildParseTrees = true

	v := newPolicyVisitor()
	p.RemoveErrorListeners()
	p.AddErrorListener(v)
	pl := p.Policy().Accept(v)

	if len(v.errors) != 0 {
		return nil, v.errors[0]
	}
	if err := validatePolicy(pl.(*netmap.PlacementPolicy)); err != nil {
		return nil, err
	}
	return pl.(*netmap.PlacementPolicy), nil
}

func (p *policyVisitor) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	p.reportError(fmt.Errorf("%w: line %d:%d %s", ErrSyntaxError, line, column, msg))
}

func (p *policyVisitor) reportError(err error) interface{} {
	p.errors = append(p.errors, err)
	return nil
}

// VisitPolicy implements parser.QueryVisitor interface.
func (p *policyVisitor) VisitPolicy(ctx *parser.PolicyContext) interface{} {
	if len(p.errors) != 0 {
		return nil
	}

	pl := new(netmap.PlacementPolicy)

	repStmts := ctx.AllRepStmt()
	rs := make([]netmap.Replica, 0, len(repStmts))
	for _, r := range repStmts {
		res, ok := r.Accept(p).(*netmap.Replica)
		if !ok {
			return nil
		}

		rs = append(rs, *res)
	}
	pl.SetReplicas(rs...)

	if cbfStmt := ctx.CbfStmt(); cbfStmt != nil {
		cbf, ok := cbfStmt.(*parser.CbfStmtContext).Accept(p).(uint32)
		if !ok {
			return nil
		}
		pl.SetContainerBackupFactor(cbf)
	}

	selStmts := ctx.AllSelectStmt()
	ss := make([]netmap.Selector, 0, len(selStmts))
	for _, s := range selStmts {
		res, ok := s.Accept(p).(*netmap.Selector)
		if !ok {
			return nil
		}

		ss = append(ss, *res)
	}
	pl.SetSelectors(ss...)

	filtStmts := ctx.AllFilterStmt()
	fs := make([]netmap.Filter, 0, len(filtStmts))
	for _, f := range filtStmts {
		fs = append(fs, *f.Accept(p).(*netmap.Filter))
	}
	pl.SetFilters(fs...)

	return pl
}

func (p *policyVisitor) VisitCbfStmt(ctx *parser.CbfStmtContext) interface{} {
	cbf, err := strconv.ParseUint(ctx.GetBackupFactor().GetText(), 10, 32)
	if err != nil {
		return p.reportError(ErrInvalidNumber)
	}

	return uint32(cbf)
}

// VisitRepStmt implements parser.QueryVisitor interface.
func (p *policyVisitor) VisitRepStmt(ctx *parser.RepStmtContext) interface{} {
	num, err := strconv.ParseUint(ctx.GetCount().GetText(), 10, 32)
	if err != nil {
		return p.reportError(ErrInvalidNumber)
	}

	rs := new(netmap.Replica)
	rs.SetCount(uint32(num))

	if sel := ctx.GetSelector(); sel != nil {
		rs.SetSelector(sel.GetText())
	}

	return rs
}

// VisitSelectStmt implements parser.QueryVisitor interface.
func (p *policyVisitor) VisitSelectStmt(ctx *parser.SelectStmtContext) interface{} {
	res, err := strconv.ParseUint(ctx.GetCount().GetText(), 10, 32)
	if err != nil {
		return p.reportError(ErrInvalidNumber)
	}

	s := new(netmap.Selector)
	s.SetCount(uint32(res))

	if clStmt := ctx.Clause(); clStmt != nil {
		s.SetClause(clauseFromString(clStmt.GetText()))
	}

	if bStmt := ctx.GetBucket(); bStmt != nil {
		s.SetAttribute(ctx.GetBucket().GetText())
	}

	s.SetFilter(ctx.GetFilter().GetText()) // either ident or wildcard

	if ctx.AS() != nil {
		s.SetName(ctx.GetName().GetText())
	}
	return s
}

// VisitFilterStmt implements parser.QueryVisitor interface.
func (p *policyVisitor) VisitFilterStmt(ctx *parser.FilterStmtContext) interface{} {
	f := p.VisitFilterExpr(ctx.GetExpr().(*parser.FilterExprContext)).(*netmap.Filter)
	f.SetName(ctx.GetName().GetText())
	return f
}

func (p *policyVisitor) VisitFilterExpr(ctx *parser.FilterExprContext) interface{} {
	if eCtx := ctx.Expr(); eCtx != nil {
		return eCtx.Accept(p)
	}

	if inner := ctx.GetInner(); inner != nil {
		return inner.Accept(p)
	}

	f := new(netmap.Filter)
	op := operationFromString(ctx.GetOp().GetText())
	f.SetOperation(op)

	f1 := *ctx.GetF1().Accept(p).(*netmap.Filter)
	f2 := *ctx.GetF2().Accept(p).(*netmap.Filter)

	// Consider f1=(.. AND ..) AND f2. This can be merged because our AND operation
	// is of arbitrary arity. ANTLR generates left-associative parse-tree by default.
	if f1.Operation() == op {
		f.SetInnerFilters(append(f1.InnerFilters(), f2)...)
		return f
	}

	f.SetInnerFilters(f1, f2)
	return f
}

// VisitFilterKey implements parser.QueryVisitor interface.
func (p *policyVisitor) VisitFilterKey(ctx *parser.FilterKeyContext) interface{} {
	if id := ctx.Ident(); id != nil {
		return id.GetText()
	}

	str := ctx.STRING().GetText()
	return str[1 : len(str)-1]
}

func (p *policyVisitor) VisitFilterValue(ctx *parser.FilterValueContext) interface{} {
	if id := ctx.Ident(); id != nil {
		return id.GetText()
	}

	if num := ctx.Number(); num != nil {
		return num.GetText()
	}

	str := ctx.STRING().GetText()
	return str[1 : len(str)-1]
}

// VisitExpr implements parser.QueryVisitor interface.
func (p *policyVisitor) VisitExpr(ctx *parser.ExprContext) interface{} {
	f := new(netmap.Filter)
	if flt := ctx.GetFilter(); flt != nil {
		f.SetName(flt.GetText())
		return f
	}

	key := ctx.GetKey().Accept(p)
	opStr := ctx.SIMPLE_OP().GetText()
	value := ctx.GetValue().Accept(p)

	f.SetKey(key.(string))
	f.SetOperation(operationFromString(opStr))
	f.SetValue(value.(string))
	return f
}

// validatePolicy checks high-level constraints such as filter link in SELECT
// being actually defined in FILTER section.
func validatePolicy(p *netmap.PlacementPolicy) error {
	seenFilters := map[string]bool{}
	for _, f := range p.Filters() {
		seenFilters[f.Name()] = true
	}

	seenSelectors := map[string]bool{}
	for _, s := range p.Selectors() {
		if flt := s.Filter(); flt != netmap.MainFilterName && !seenFilters[flt] {
			return fmt.Errorf("%w: '%s'", ErrUnknownFilter, flt)
		}
		seenSelectors[s.Name()] = true
	}

	for _, r := range p.Replicas() {
		if sel := r.Selector(); sel != "" && !seenSelectors[sel] {
			return fmt.Errorf("%w: '%s'", ErrUnknownSelector, sel)
		}
	}

	return nil
}

func clauseFromString(s string) netmap.Clause {
	switch strings.ToUpper(s) {
	case "SAME":
		return netmap.ClauseSame
	case "DISTINCT":
		return netmap.ClauseDistinct
	default:
		// Such errors should be handled by ANTLR code thus this panic.
		panic(fmt.Errorf("BUG: invalid clause: %s", s))
	}
}

func operationFromString(op string) netmap.Operation {
	switch strings.ToUpper(op) {
	case "AND":
		return netmap.OpAND
	case "OR":
		return netmap.OpOR
	case "EQ":
		return netmap.OpEQ
	case "NE":
		return netmap.OpNE
	case "GE":
		return netmap.OpGE
	case "GT":
		return netmap.OpGT
	case "LE":
		return netmap.OpLE
	case "LT":
		return netmap.OpLT
	default:
		// Such errors should be handled by ANTLR code thus this panic.
		panic(fmt.Errorf("BUG: invalid operation: %s", op))
	}
}
