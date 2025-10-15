// Code generated from Query.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // Query

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/antlr4-go/antlr/v4"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = strconv.Itoa
var _ = sync.Once{}

type Query struct {
	*antlr.BaseParser
}

var QueryParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func queryParserInit() {
	staticData := &QueryParserStaticData
	staticData.LiteralNames = []string{
		"", "'AND'", "'OR'", "", "'REP'", "'IN'", "'AS'", "'CBF'", "'SELECT'",
		"'FROM'", "'FILTER'", "'*'", "'EC'", "'SAME'", "'DISTINCT'", "'('",
		"')'", "'@'", "'/'", "", "", "'0'",
	}
	staticData.SymbolicNames = []string{
		"", "AND_OP", "OR_OP", "SIMPLE_OP", "REP", "IN", "AS", "CBF", "SELECT",
		"FROM", "FILTER", "WILDCARD", "EC", "CLAUSE_SAME", "CLAUSE_DISTINCT",
		"L_PAREN", "R_PAREN", "AT", "EC_SEP", "IDENT", "NUMBER1", "ZERO", "STRING",
		"WS",
	}
	staticData.RuleNames = []string{
		"policy", "ruleStmt", "repStmt", "ecStmt", "cbfStmt", "selectStmt",
		"clause", "filterExpr", "filterStmt", "expr", "filterKey", "filterValue",
		"number", "keyword", "ident", "identWC",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 23, 146, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15, 7, 15,
		1, 0, 4, 0, 34, 8, 0, 11, 0, 12, 0, 35, 1, 0, 3, 0, 39, 8, 0, 1, 0, 5,
		0, 42, 8, 0, 10, 0, 12, 0, 45, 9, 0, 1, 0, 5, 0, 48, 8, 0, 10, 0, 12, 0,
		51, 9, 0, 1, 0, 1, 0, 1, 1, 1, 1, 3, 1, 57, 8, 1, 1, 2, 1, 2, 1, 2, 1,
		2, 3, 2, 63, 8, 2, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 3, 3, 71, 8, 3,
		1, 4, 1, 4, 1, 4, 1, 5, 1, 5, 1, 5, 1, 5, 3, 5, 80, 8, 5, 1, 5, 3, 5, 83,
		8, 5, 1, 5, 1, 5, 1, 5, 1, 5, 3, 5, 89, 8, 5, 1, 6, 1, 6, 1, 7, 1, 7, 1,
		7, 1, 7, 1, 7, 1, 7, 3, 7, 99, 8, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7,
		5, 7, 107, 8, 7, 10, 7, 12, 7, 110, 9, 7, 1, 8, 1, 8, 1, 8, 1, 8, 1, 8,
		1, 9, 1, 9, 1, 9, 1, 9, 1, 9, 1, 9, 3, 9, 123, 8, 9, 1, 10, 1, 10, 3, 10,
		127, 8, 10, 1, 11, 1, 11, 1, 11, 3, 11, 132, 8, 11, 1, 12, 1, 12, 1, 13,
		1, 13, 1, 14, 1, 14, 3, 14, 140, 8, 14, 1, 15, 1, 15, 3, 15, 144, 8, 15,
		1, 15, 0, 1, 14, 16, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26,
		28, 30, 0, 3, 1, 0, 13, 14, 1, 0, 20, 21, 3, 0, 4, 6, 8, 10, 12, 12, 148,
		0, 33, 1, 0, 0, 0, 2, 56, 1, 0, 0, 0, 4, 58, 1, 0, 0, 0, 6, 64, 1, 0, 0,
		0, 8, 72, 1, 0, 0, 0, 10, 75, 1, 0, 0, 0, 12, 90, 1, 0, 0, 0, 14, 98, 1,
		0, 0, 0, 16, 111, 1, 0, 0, 0, 18, 122, 1, 0, 0, 0, 20, 126, 1, 0, 0, 0,
		22, 131, 1, 0, 0, 0, 24, 133, 1, 0, 0, 0, 26, 135, 1, 0, 0, 0, 28, 139,
		1, 0, 0, 0, 30, 143, 1, 0, 0, 0, 32, 34, 3, 2, 1, 0, 33, 32, 1, 0, 0, 0,
		34, 35, 1, 0, 0, 0, 35, 33, 1, 0, 0, 0, 35, 36, 1, 0, 0, 0, 36, 38, 1,
		0, 0, 0, 37, 39, 3, 8, 4, 0, 38, 37, 1, 0, 0, 0, 38, 39, 1, 0, 0, 0, 39,
		43, 1, 0, 0, 0, 40, 42, 3, 10, 5, 0, 41, 40, 1, 0, 0, 0, 42, 45, 1, 0,
		0, 0, 43, 41, 1, 0, 0, 0, 43, 44, 1, 0, 0, 0, 44, 49, 1, 0, 0, 0, 45, 43,
		1, 0, 0, 0, 46, 48, 3, 16, 8, 0, 47, 46, 1, 0, 0, 0, 48, 51, 1, 0, 0, 0,
		49, 47, 1, 0, 0, 0, 49, 50, 1, 0, 0, 0, 50, 52, 1, 0, 0, 0, 51, 49, 1,
		0, 0, 0, 52, 53, 5, 0, 0, 1, 53, 1, 1, 0, 0, 0, 54, 57, 3, 4, 2, 0, 55,
		57, 3, 6, 3, 0, 56, 54, 1, 0, 0, 0, 56, 55, 1, 0, 0, 0, 57, 3, 1, 0, 0,
		0, 58, 59, 5, 4, 0, 0, 59, 62, 5, 20, 0, 0, 60, 61, 5, 5, 0, 0, 61, 63,
		3, 28, 14, 0, 62, 60, 1, 0, 0, 0, 62, 63, 1, 0, 0, 0, 63, 5, 1, 0, 0, 0,
		64, 65, 5, 12, 0, 0, 65, 66, 5, 20, 0, 0, 66, 67, 5, 18, 0, 0, 67, 70,
		5, 20, 0, 0, 68, 69, 5, 5, 0, 0, 69, 71, 3, 28, 14, 0, 70, 68, 1, 0, 0,
		0, 70, 71, 1, 0, 0, 0, 71, 7, 1, 0, 0, 0, 72, 73, 5, 7, 0, 0, 73, 74, 5,
		20, 0, 0, 74, 9, 1, 0, 0, 0, 75, 76, 5, 8, 0, 0, 76, 82, 5, 20, 0, 0, 77,
		79, 5, 5, 0, 0, 78, 80, 3, 12, 6, 0, 79, 78, 1, 0, 0, 0, 79, 80, 1, 0,
		0, 0, 80, 81, 1, 0, 0, 0, 81, 83, 3, 28, 14, 0, 82, 77, 1, 0, 0, 0, 82,
		83, 1, 0, 0, 0, 83, 84, 1, 0, 0, 0, 84, 85, 5, 9, 0, 0, 85, 88, 3, 30,
		15, 0, 86, 87, 5, 6, 0, 0, 87, 89, 3, 28, 14, 0, 88, 86, 1, 0, 0, 0, 88,
		89, 1, 0, 0, 0, 89, 11, 1, 0, 0, 0, 90, 91, 7, 0, 0, 0, 91, 13, 1, 0, 0,
		0, 92, 93, 6, 7, -1, 0, 93, 94, 5, 15, 0, 0, 94, 95, 3, 14, 7, 0, 95, 96,
		5, 16, 0, 0, 96, 99, 1, 0, 0, 0, 97, 99, 3, 18, 9, 0, 98, 92, 1, 0, 0,
		0, 98, 97, 1, 0, 0, 0, 99, 108, 1, 0, 0, 0, 100, 101, 10, 4, 0, 0, 101,
		102, 5, 1, 0, 0, 102, 107, 3, 14, 7, 5, 103, 104, 10, 3, 0, 0, 104, 105,
		5, 2, 0, 0, 105, 107, 3, 14, 7, 4, 106, 100, 1, 0, 0, 0, 106, 103, 1, 0,
		0, 0, 107, 110, 1, 0, 0, 0, 108, 106, 1, 0, 0, 0, 108, 109, 1, 0, 0, 0,
		109, 15, 1, 0, 0, 0, 110, 108, 1, 0, 0, 0, 111, 112, 5, 10, 0, 0, 112,
		113, 3, 14, 7, 0, 113, 114, 5, 6, 0, 0, 114, 115, 3, 28, 14, 0, 115, 17,
		1, 0, 0, 0, 116, 117, 5, 17, 0, 0, 117, 123, 3, 28, 14, 0, 118, 119, 3,
		20, 10, 0, 119, 120, 5, 3, 0, 0, 120, 121, 3, 22, 11, 0, 121, 123, 1, 0,
		0, 0, 122, 116, 1, 0, 0, 0, 122, 118, 1, 0, 0, 0, 123, 19, 1, 0, 0, 0,
		124, 127, 3, 28, 14, 0, 125, 127, 5, 22, 0, 0, 126, 124, 1, 0, 0, 0, 126,
		125, 1, 0, 0, 0, 127, 21, 1, 0, 0, 0, 128, 132, 3, 28, 14, 0, 129, 132,
		3, 24, 12, 0, 130, 132, 5, 22, 0, 0, 131, 128, 1, 0, 0, 0, 131, 129, 1,
		0, 0, 0, 131, 130, 1, 0, 0, 0, 132, 23, 1, 0, 0, 0, 133, 134, 7, 1, 0,
		0, 134, 25, 1, 0, 0, 0, 135, 136, 7, 2, 0, 0, 136, 27, 1, 0, 0, 0, 137,
		140, 3, 26, 13, 0, 138, 140, 5, 19, 0, 0, 139, 137, 1, 0, 0, 0, 139, 138,
		1, 0, 0, 0, 140, 29, 1, 0, 0, 0, 141, 144, 3, 28, 14, 0, 142, 144, 5, 11,
		0, 0, 143, 141, 1, 0, 0, 0, 143, 142, 1, 0, 0, 0, 144, 31, 1, 0, 0, 0,
		18, 35, 38, 43, 49, 56, 62, 70, 79, 82, 88, 98, 106, 108, 122, 126, 131,
		139, 143,
	}
	deserializer := antlr.NewATNDeserializer(nil)
	staticData.atn = deserializer.Deserialize(staticData.serializedATN)
	atn := staticData.atn
	staticData.decisionToDFA = make([]*antlr.DFA, len(atn.DecisionToState))
	decisionToDFA := staticData.decisionToDFA
	for index, state := range atn.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(state, index)
	}
}

// QueryInit initializes any static state used to implement Query. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewQuery(). You can call this function if you wish to initialize the static state ahead
// of time.
func QueryInit() {
	staticData := &QueryParserStaticData
	staticData.once.Do(queryParserInit)
}

// NewQuery produces a new parser instance for the optional input antlr.TokenStream.
func NewQuery(input antlr.TokenStream) *Query {
	QueryInit()
	this := new(Query)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &QueryParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "Query.g4"

	return this
}

// Query tokens.
const (
	QueryEOF             = antlr.TokenEOF
	QueryAND_OP          = 1
	QueryOR_OP           = 2
	QuerySIMPLE_OP       = 3
	QueryREP             = 4
	QueryIN              = 5
	QueryAS              = 6
	QueryCBF             = 7
	QuerySELECT          = 8
	QueryFROM            = 9
	QueryFILTER          = 10
	QueryWILDCARD        = 11
	QueryEC              = 12
	QueryCLAUSE_SAME     = 13
	QueryCLAUSE_DISTINCT = 14
	QueryL_PAREN         = 15
	QueryR_PAREN         = 16
	QueryAT              = 17
	QueryEC_SEP          = 18
	QueryIDENT           = 19
	QueryNUMBER1         = 20
	QueryZERO            = 21
	QuerySTRING          = 22
	QueryWS              = 23
)

// Query rules.
const (
	QueryRULE_policy      = 0
	QueryRULE_ruleStmt    = 1
	QueryRULE_repStmt     = 2
	QueryRULE_ecStmt      = 3
	QueryRULE_cbfStmt     = 4
	QueryRULE_selectStmt  = 5
	QueryRULE_clause      = 6
	QueryRULE_filterExpr  = 7
	QueryRULE_filterStmt  = 8
	QueryRULE_expr        = 9
	QueryRULE_filterKey   = 10
	QueryRULE_filterValue = 11
	QueryRULE_number      = 12
	QueryRULE_keyword     = 13
	QueryRULE_ident       = 14
	QueryRULE_identWC     = 15
)

// IPolicyContext is an interface to support dynamic dispatch.
type IPolicyContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EOF() antlr.TerminalNode
	AllRuleStmt() []IRuleStmtContext
	RuleStmt(i int) IRuleStmtContext
	CbfStmt() ICbfStmtContext
	AllSelectStmt() []ISelectStmtContext
	SelectStmt(i int) ISelectStmtContext
	AllFilterStmt() []IFilterStmtContext
	FilterStmt(i int) IFilterStmtContext

	// IsPolicyContext differentiates from other interfaces.
	IsPolicyContext()
}

type PolicyContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPolicyContext() *PolicyContext {
	var p = new(PolicyContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_policy
	return p
}

func InitEmptyPolicyContext(p *PolicyContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_policy
}

func (*PolicyContext) IsPolicyContext() {}

func NewPolicyContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PolicyContext {
	var p = new(PolicyContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryRULE_policy

	return p
}

func (s *PolicyContext) GetParser() antlr.Parser { return s.parser }

func (s *PolicyContext) EOF() antlr.TerminalNode {
	return s.GetToken(QueryEOF, 0)
}

func (s *PolicyContext) AllRuleStmt() []IRuleStmtContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IRuleStmtContext); ok {
			len++
		}
	}

	tst := make([]IRuleStmtContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IRuleStmtContext); ok {
			tst[i] = t.(IRuleStmtContext)
			i++
		}
	}

	return tst
}

func (s *PolicyContext) RuleStmt(i int) IRuleStmtContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IRuleStmtContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IRuleStmtContext)
}

func (s *PolicyContext) CbfStmt() ICbfStmtContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICbfStmtContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICbfStmtContext)
}

func (s *PolicyContext) AllSelectStmt() []ISelectStmtContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ISelectStmtContext); ok {
			len++
		}
	}

	tst := make([]ISelectStmtContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ISelectStmtContext); ok {
			tst[i] = t.(ISelectStmtContext)
			i++
		}
	}

	return tst
}

func (s *PolicyContext) SelectStmt(i int) ISelectStmtContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISelectStmtContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISelectStmtContext)
}

func (s *PolicyContext) AllFilterStmt() []IFilterStmtContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IFilterStmtContext); ok {
			len++
		}
	}

	tst := make([]IFilterStmtContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IFilterStmtContext); ok {
			tst[i] = t.(IFilterStmtContext)
			i++
		}
	}

	return tst
}

func (s *PolicyContext) FilterStmt(i int) IFilterStmtContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFilterStmtContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFilterStmtContext)
}

func (s *PolicyContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PolicyContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PolicyContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.EnterPolicy(s)
	}
}

func (s *PolicyContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.ExitPolicy(s)
	}
}

func (s *PolicyContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryVisitor:
		return t.VisitPolicy(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *Query) Policy() (localctx IPolicyContext) {
	localctx = NewPolicyContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, QueryRULE_policy)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(33)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == QueryREP || _la == QueryEC {
		{
			p.SetState(32)
			p.RuleStmt()
		}

		p.SetState(35)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(38)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == QueryCBF {
		{
			p.SetState(37)
			p.CbfStmt()
		}

	}
	p.SetState(43)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == QuerySELECT {
		{
			p.SetState(40)
			p.SelectStmt()
		}

		p.SetState(45)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(49)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == QueryFILTER {
		{
			p.SetState(46)
			p.FilterStmt()
		}

		p.SetState(51)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(52)
		p.Match(QueryEOF)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IRuleStmtContext is an interface to support dynamic dispatch.
type IRuleStmtContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	RepStmt() IRepStmtContext
	EcStmt() IEcStmtContext

	// IsRuleStmtContext differentiates from other interfaces.
	IsRuleStmtContext()
}

type RuleStmtContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyRuleStmtContext() *RuleStmtContext {
	var p = new(RuleStmtContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_ruleStmt
	return p
}

func InitEmptyRuleStmtContext(p *RuleStmtContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_ruleStmt
}

func (*RuleStmtContext) IsRuleStmtContext() {}

func NewRuleStmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *RuleStmtContext {
	var p = new(RuleStmtContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryRULE_ruleStmt

	return p
}

func (s *RuleStmtContext) GetParser() antlr.Parser { return s.parser }

func (s *RuleStmtContext) RepStmt() IRepStmtContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IRepStmtContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IRepStmtContext)
}

func (s *RuleStmtContext) EcStmt() IEcStmtContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IEcStmtContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IEcStmtContext)
}

func (s *RuleStmtContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RuleStmtContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *RuleStmtContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.EnterRuleStmt(s)
	}
}

func (s *RuleStmtContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.ExitRuleStmt(s)
	}
}

func (s *RuleStmtContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryVisitor:
		return t.VisitRuleStmt(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *Query) RuleStmt() (localctx IRuleStmtContext) {
	localctx = NewRuleStmtContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, QueryRULE_ruleStmt)
	p.SetState(56)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case QueryREP:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(54)
			p.RepStmt()
		}

	case QueryEC:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(55)
			p.EcStmt()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IRepStmtContext is an interface to support dynamic dispatch.
type IRepStmtContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetCount returns the Count token.
	GetCount() antlr.Token

	// SetCount sets the Count token.
	SetCount(antlr.Token)

	// GetSelector returns the Selector rule contexts.
	GetSelector() IIdentContext

	// SetSelector sets the Selector rule contexts.
	SetSelector(IIdentContext)

	// Getter signatures
	REP() antlr.TerminalNode
	NUMBER1() antlr.TerminalNode
	IN() antlr.TerminalNode
	Ident() IIdentContext

	// IsRepStmtContext differentiates from other interfaces.
	IsRepStmtContext()
}

type RepStmtContext struct {
	antlr.BaseParserRuleContext
	parser   antlr.Parser
	Count    antlr.Token
	Selector IIdentContext
}

func NewEmptyRepStmtContext() *RepStmtContext {
	var p = new(RepStmtContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_repStmt
	return p
}

func InitEmptyRepStmtContext(p *RepStmtContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_repStmt
}

func (*RepStmtContext) IsRepStmtContext() {}

func NewRepStmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *RepStmtContext {
	var p = new(RepStmtContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryRULE_repStmt

	return p
}

func (s *RepStmtContext) GetParser() antlr.Parser { return s.parser }

func (s *RepStmtContext) GetCount() antlr.Token { return s.Count }

func (s *RepStmtContext) SetCount(v antlr.Token) { s.Count = v }

func (s *RepStmtContext) GetSelector() IIdentContext { return s.Selector }

func (s *RepStmtContext) SetSelector(v IIdentContext) { s.Selector = v }

func (s *RepStmtContext) REP() antlr.TerminalNode {
	return s.GetToken(QueryREP, 0)
}

func (s *RepStmtContext) NUMBER1() antlr.TerminalNode {
	return s.GetToken(QueryNUMBER1, 0)
}

func (s *RepStmtContext) IN() antlr.TerminalNode {
	return s.GetToken(QueryIN, 0)
}

func (s *RepStmtContext) Ident() IIdentContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentContext)
}

func (s *RepStmtContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RepStmtContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *RepStmtContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.EnterRepStmt(s)
	}
}

func (s *RepStmtContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.ExitRepStmt(s)
	}
}

func (s *RepStmtContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryVisitor:
		return t.VisitRepStmt(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *Query) RepStmt() (localctx IRepStmtContext) {
	localctx = NewRepStmtContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, QueryRULE_repStmt)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(58)
		p.Match(QueryREP)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(59)

		var _m = p.Match(QueryNUMBER1)

		localctx.(*RepStmtContext).Count = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(62)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == QueryIN {
		{
			p.SetState(60)
			p.Match(QueryIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(61)

			var _x = p.Ident()

			localctx.(*RepStmtContext).Selector = _x
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IEcStmtContext is an interface to support dynamic dispatch.
type IEcStmtContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetDataPartNum returns the DataPartNum token.
	GetDataPartNum() antlr.Token

	// GetParityPartNum returns the ParityPartNum token.
	GetParityPartNum() antlr.Token

	// SetDataPartNum sets the DataPartNum token.
	SetDataPartNum(antlr.Token)

	// SetParityPartNum sets the ParityPartNum token.
	SetParityPartNum(antlr.Token)

	// GetSelector returns the Selector rule contexts.
	GetSelector() IIdentContext

	// SetSelector sets the Selector rule contexts.
	SetSelector(IIdentContext)

	// Getter signatures
	EC() antlr.TerminalNode
	EC_SEP() antlr.TerminalNode
	AllNUMBER1() []antlr.TerminalNode
	NUMBER1(i int) antlr.TerminalNode
	IN() antlr.TerminalNode
	Ident() IIdentContext

	// IsEcStmtContext differentiates from other interfaces.
	IsEcStmtContext()
}

type EcStmtContext struct {
	antlr.BaseParserRuleContext
	parser        antlr.Parser
	DataPartNum   antlr.Token
	ParityPartNum antlr.Token
	Selector      IIdentContext
}

func NewEmptyEcStmtContext() *EcStmtContext {
	var p = new(EcStmtContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_ecStmt
	return p
}

func InitEmptyEcStmtContext(p *EcStmtContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_ecStmt
}

func (*EcStmtContext) IsEcStmtContext() {}

func NewEcStmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *EcStmtContext {
	var p = new(EcStmtContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryRULE_ecStmt

	return p
}

func (s *EcStmtContext) GetParser() antlr.Parser { return s.parser }

func (s *EcStmtContext) GetDataPartNum() antlr.Token { return s.DataPartNum }

func (s *EcStmtContext) GetParityPartNum() antlr.Token { return s.ParityPartNum }

func (s *EcStmtContext) SetDataPartNum(v antlr.Token) { s.DataPartNum = v }

func (s *EcStmtContext) SetParityPartNum(v antlr.Token) { s.ParityPartNum = v }

func (s *EcStmtContext) GetSelector() IIdentContext { return s.Selector }

func (s *EcStmtContext) SetSelector(v IIdentContext) { s.Selector = v }

func (s *EcStmtContext) EC() antlr.TerminalNode {
	return s.GetToken(QueryEC, 0)
}

func (s *EcStmtContext) EC_SEP() antlr.TerminalNode {
	return s.GetToken(QueryEC_SEP, 0)
}

func (s *EcStmtContext) AllNUMBER1() []antlr.TerminalNode {
	return s.GetTokens(QueryNUMBER1)
}

func (s *EcStmtContext) NUMBER1(i int) antlr.TerminalNode {
	return s.GetToken(QueryNUMBER1, i)
}

func (s *EcStmtContext) IN() antlr.TerminalNode {
	return s.GetToken(QueryIN, 0)
}

func (s *EcStmtContext) Ident() IIdentContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentContext)
}

func (s *EcStmtContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *EcStmtContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *EcStmtContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.EnterEcStmt(s)
	}
}

func (s *EcStmtContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.ExitEcStmt(s)
	}
}

func (s *EcStmtContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryVisitor:
		return t.VisitEcStmt(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *Query) EcStmt() (localctx IEcStmtContext) {
	localctx = NewEcStmtContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, QueryRULE_ecStmt)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(64)
		p.Match(QueryEC)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(65)

		var _m = p.Match(QueryNUMBER1)

		localctx.(*EcStmtContext).DataPartNum = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(66)
		p.Match(QueryEC_SEP)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(67)

		var _m = p.Match(QueryNUMBER1)

		localctx.(*EcStmtContext).ParityPartNum = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(70)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == QueryIN {
		{
			p.SetState(68)
			p.Match(QueryIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(69)

			var _x = p.Ident()

			localctx.(*EcStmtContext).Selector = _x
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ICbfStmtContext is an interface to support dynamic dispatch.
type ICbfStmtContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetBackupFactor returns the BackupFactor token.
	GetBackupFactor() antlr.Token

	// SetBackupFactor sets the BackupFactor token.
	SetBackupFactor(antlr.Token)

	// Getter signatures
	CBF() antlr.TerminalNode
	NUMBER1() antlr.TerminalNode

	// IsCbfStmtContext differentiates from other interfaces.
	IsCbfStmtContext()
}

type CbfStmtContext struct {
	antlr.BaseParserRuleContext
	parser       antlr.Parser
	BackupFactor antlr.Token
}

func NewEmptyCbfStmtContext() *CbfStmtContext {
	var p = new(CbfStmtContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_cbfStmt
	return p
}

func InitEmptyCbfStmtContext(p *CbfStmtContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_cbfStmt
}

func (*CbfStmtContext) IsCbfStmtContext() {}

func NewCbfStmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CbfStmtContext {
	var p = new(CbfStmtContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryRULE_cbfStmt

	return p
}

func (s *CbfStmtContext) GetParser() antlr.Parser { return s.parser }

func (s *CbfStmtContext) GetBackupFactor() antlr.Token { return s.BackupFactor }

func (s *CbfStmtContext) SetBackupFactor(v antlr.Token) { s.BackupFactor = v }

func (s *CbfStmtContext) CBF() antlr.TerminalNode {
	return s.GetToken(QueryCBF, 0)
}

func (s *CbfStmtContext) NUMBER1() antlr.TerminalNode {
	return s.GetToken(QueryNUMBER1, 0)
}

func (s *CbfStmtContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CbfStmtContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CbfStmtContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.EnterCbfStmt(s)
	}
}

func (s *CbfStmtContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.ExitCbfStmt(s)
	}
}

func (s *CbfStmtContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryVisitor:
		return t.VisitCbfStmt(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *Query) CbfStmt() (localctx ICbfStmtContext) {
	localctx = NewCbfStmtContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, QueryRULE_cbfStmt)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(72)
		p.Match(QueryCBF)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(73)

		var _m = p.Match(QueryNUMBER1)

		localctx.(*CbfStmtContext).BackupFactor = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISelectStmtContext is an interface to support dynamic dispatch.
type ISelectStmtContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetCount returns the Count token.
	GetCount() antlr.Token

	// SetCount sets the Count token.
	SetCount(antlr.Token)

	// GetBucket returns the Bucket rule contexts.
	GetBucket() IIdentContext

	// GetFilter returns the Filter rule contexts.
	GetFilter() IIdentWCContext

	// GetName returns the Name rule contexts.
	GetName() IIdentContext

	// SetBucket sets the Bucket rule contexts.
	SetBucket(IIdentContext)

	// SetFilter sets the Filter rule contexts.
	SetFilter(IIdentWCContext)

	// SetName sets the Name rule contexts.
	SetName(IIdentContext)

	// Getter signatures
	SELECT() antlr.TerminalNode
	FROM() antlr.TerminalNode
	NUMBER1() antlr.TerminalNode
	IdentWC() IIdentWCContext
	IN() antlr.TerminalNode
	AS() antlr.TerminalNode
	AllIdent() []IIdentContext
	Ident(i int) IIdentContext
	Clause() IClauseContext

	// IsSelectStmtContext differentiates from other interfaces.
	IsSelectStmtContext()
}

type SelectStmtContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
	Count  antlr.Token
	Bucket IIdentContext
	Filter IIdentWCContext
	Name   IIdentContext
}

func NewEmptySelectStmtContext() *SelectStmtContext {
	var p = new(SelectStmtContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_selectStmt
	return p
}

func InitEmptySelectStmtContext(p *SelectStmtContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_selectStmt
}

func (*SelectStmtContext) IsSelectStmtContext() {}

func NewSelectStmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SelectStmtContext {
	var p = new(SelectStmtContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryRULE_selectStmt

	return p
}

func (s *SelectStmtContext) GetParser() antlr.Parser { return s.parser }

func (s *SelectStmtContext) GetCount() antlr.Token { return s.Count }

func (s *SelectStmtContext) SetCount(v antlr.Token) { s.Count = v }

func (s *SelectStmtContext) GetBucket() IIdentContext { return s.Bucket }

func (s *SelectStmtContext) GetFilter() IIdentWCContext { return s.Filter }

func (s *SelectStmtContext) GetName() IIdentContext { return s.Name }

func (s *SelectStmtContext) SetBucket(v IIdentContext) { s.Bucket = v }

func (s *SelectStmtContext) SetFilter(v IIdentWCContext) { s.Filter = v }

func (s *SelectStmtContext) SetName(v IIdentContext) { s.Name = v }

func (s *SelectStmtContext) SELECT() antlr.TerminalNode {
	return s.GetToken(QuerySELECT, 0)
}

func (s *SelectStmtContext) FROM() antlr.TerminalNode {
	return s.GetToken(QueryFROM, 0)
}

func (s *SelectStmtContext) NUMBER1() antlr.TerminalNode {
	return s.GetToken(QueryNUMBER1, 0)
}

func (s *SelectStmtContext) IdentWC() IIdentWCContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentWCContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentWCContext)
}

func (s *SelectStmtContext) IN() antlr.TerminalNode {
	return s.GetToken(QueryIN, 0)
}

func (s *SelectStmtContext) AS() antlr.TerminalNode {
	return s.GetToken(QueryAS, 0)
}

func (s *SelectStmtContext) AllIdent() []IIdentContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IIdentContext); ok {
			len++
		}
	}

	tst := make([]IIdentContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IIdentContext); ok {
			tst[i] = t.(IIdentContext)
			i++
		}
	}

	return tst
}

func (s *SelectStmtContext) Ident(i int) IIdentContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentContext)
}

func (s *SelectStmtContext) Clause() IClauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IClauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IClauseContext)
}

func (s *SelectStmtContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SelectStmtContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SelectStmtContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.EnterSelectStmt(s)
	}
}

func (s *SelectStmtContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.ExitSelectStmt(s)
	}
}

func (s *SelectStmtContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryVisitor:
		return t.VisitSelectStmt(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *Query) SelectStmt() (localctx ISelectStmtContext) {
	localctx = NewSelectStmtContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, QueryRULE_selectStmt)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(75)
		p.Match(QuerySELECT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(76)

		var _m = p.Match(QueryNUMBER1)

		localctx.(*SelectStmtContext).Count = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(82)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == QueryIN {
		{
			p.SetState(77)
			p.Match(QueryIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(79)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == QueryCLAUSE_SAME || _la == QueryCLAUSE_DISTINCT {
			{
				p.SetState(78)
				p.Clause()
			}

		}
		{
			p.SetState(81)

			var _x = p.Ident()

			localctx.(*SelectStmtContext).Bucket = _x
		}

	}
	{
		p.SetState(84)
		p.Match(QueryFROM)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(85)

		var _x = p.IdentWC()

		localctx.(*SelectStmtContext).Filter = _x
	}
	p.SetState(88)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == QueryAS {
		{
			p.SetState(86)
			p.Match(QueryAS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(87)

			var _x = p.Ident()

			localctx.(*SelectStmtContext).Name = _x
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IClauseContext is an interface to support dynamic dispatch.
type IClauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	CLAUSE_SAME() antlr.TerminalNode
	CLAUSE_DISTINCT() antlr.TerminalNode

	// IsClauseContext differentiates from other interfaces.
	IsClauseContext()
}

type ClauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyClauseContext() *ClauseContext {
	var p = new(ClauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_clause
	return p
}

func InitEmptyClauseContext(p *ClauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_clause
}

func (*ClauseContext) IsClauseContext() {}

func NewClauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ClauseContext {
	var p = new(ClauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryRULE_clause

	return p
}

func (s *ClauseContext) GetParser() antlr.Parser { return s.parser }

func (s *ClauseContext) CLAUSE_SAME() antlr.TerminalNode {
	return s.GetToken(QueryCLAUSE_SAME, 0)
}

func (s *ClauseContext) CLAUSE_DISTINCT() antlr.TerminalNode {
	return s.GetToken(QueryCLAUSE_DISTINCT, 0)
}

func (s *ClauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ClauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ClauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.EnterClause(s)
	}
}

func (s *ClauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.ExitClause(s)
	}
}

func (s *ClauseContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryVisitor:
		return t.VisitClause(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *Query) Clause() (localctx IClauseContext) {
	localctx = NewClauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, QueryRULE_clause)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(90)
		_la = p.GetTokenStream().LA(1)

		if !(_la == QueryCLAUSE_SAME || _la == QueryCLAUSE_DISTINCT) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IFilterExprContext is an interface to support dynamic dispatch.
type IFilterExprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetOp returns the Op token.
	GetOp() antlr.Token

	// SetOp sets the Op token.
	SetOp(antlr.Token)

	// GetF1 returns the F1 rule contexts.
	GetF1() IFilterExprContext

	// GetInner returns the Inner rule contexts.
	GetInner() IFilterExprContext

	// GetF2 returns the F2 rule contexts.
	GetF2() IFilterExprContext

	// SetF1 sets the F1 rule contexts.
	SetF1(IFilterExprContext)

	// SetInner sets the Inner rule contexts.
	SetInner(IFilterExprContext)

	// SetF2 sets the F2 rule contexts.
	SetF2(IFilterExprContext)

	// Getter signatures
	L_PAREN() antlr.TerminalNode
	R_PAREN() antlr.TerminalNode
	AllFilterExpr() []IFilterExprContext
	FilterExpr(i int) IFilterExprContext
	Expr() IExprContext
	AND_OP() antlr.TerminalNode
	OR_OP() antlr.TerminalNode

	// IsFilterExprContext differentiates from other interfaces.
	IsFilterExprContext()
}

type FilterExprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
	F1     IFilterExprContext
	Inner  IFilterExprContext
	Op     antlr.Token
	F2     IFilterExprContext
}

func NewEmptyFilterExprContext() *FilterExprContext {
	var p = new(FilterExprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_filterExpr
	return p
}

func InitEmptyFilterExprContext(p *FilterExprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_filterExpr
}

func (*FilterExprContext) IsFilterExprContext() {}

func NewFilterExprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FilterExprContext {
	var p = new(FilterExprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryRULE_filterExpr

	return p
}

func (s *FilterExprContext) GetParser() antlr.Parser { return s.parser }

func (s *FilterExprContext) GetOp() antlr.Token { return s.Op }

func (s *FilterExprContext) SetOp(v antlr.Token) { s.Op = v }

func (s *FilterExprContext) GetF1() IFilterExprContext { return s.F1 }

func (s *FilterExprContext) GetInner() IFilterExprContext { return s.Inner }

func (s *FilterExprContext) GetF2() IFilterExprContext { return s.F2 }

func (s *FilterExprContext) SetF1(v IFilterExprContext) { s.F1 = v }

func (s *FilterExprContext) SetInner(v IFilterExprContext) { s.Inner = v }

func (s *FilterExprContext) SetF2(v IFilterExprContext) { s.F2 = v }

func (s *FilterExprContext) L_PAREN() antlr.TerminalNode {
	return s.GetToken(QueryL_PAREN, 0)
}

func (s *FilterExprContext) R_PAREN() antlr.TerminalNode {
	return s.GetToken(QueryR_PAREN, 0)
}

func (s *FilterExprContext) AllFilterExpr() []IFilterExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IFilterExprContext); ok {
			len++
		}
	}

	tst := make([]IFilterExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IFilterExprContext); ok {
			tst[i] = t.(IFilterExprContext)
			i++
		}
	}

	return tst
}

func (s *FilterExprContext) FilterExpr(i int) IFilterExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFilterExprContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFilterExprContext)
}

func (s *FilterExprContext) Expr() IExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *FilterExprContext) AND_OP() antlr.TerminalNode {
	return s.GetToken(QueryAND_OP, 0)
}

func (s *FilterExprContext) OR_OP() antlr.TerminalNode {
	return s.GetToken(QueryOR_OP, 0)
}

func (s *FilterExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FilterExprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *FilterExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.EnterFilterExpr(s)
	}
}

func (s *FilterExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.ExitFilterExpr(s)
	}
}

func (s *FilterExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryVisitor:
		return t.VisitFilterExpr(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *Query) FilterExpr() (localctx IFilterExprContext) {
	return p.filterExpr(0)
}

func (p *Query) filterExpr(_p int) (localctx IFilterExprContext) {
	var _parentctx antlr.ParserRuleContext = p.GetParserRuleContext()

	_parentState := p.GetState()
	localctx = NewFilterExprContext(p, p.GetParserRuleContext(), _parentState)
	var _prevctx IFilterExprContext = localctx
	var _ antlr.ParserRuleContext = _prevctx // TODO: To prevent unused variable warning.
	_startState := 14
	p.EnterRecursionRule(localctx, 14, QueryRULE_filterExpr, _p)
	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(98)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case QueryL_PAREN:
		{
			p.SetState(93)
			p.Match(QueryL_PAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(94)

			var _x = p.filterExpr(0)

			localctx.(*FilterExprContext).Inner = _x
		}
		{
			p.SetState(95)
			p.Match(QueryR_PAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case QueryREP, QueryIN, QueryAS, QuerySELECT, QueryFROM, QueryFILTER, QueryEC, QueryAT, QueryIDENT, QuerySTRING:
		{
			p.SetState(97)
			p.Expr()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(108)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 12, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(106)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 11, p.GetParserRuleContext()) {
			case 1:
				localctx = NewFilterExprContext(p, _parentctx, _parentState)
				localctx.(*FilterExprContext).F1 = _prevctx
				p.PushNewRecursionContext(localctx, _startState, QueryRULE_filterExpr)
				p.SetState(100)

				if !(p.Precpred(p.GetParserRuleContext(), 4)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 4)", ""))
					goto errorExit
				}
				{
					p.SetState(101)

					var _m = p.Match(QueryAND_OP)

					localctx.(*FilterExprContext).Op = _m
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(102)

					var _x = p.filterExpr(5)

					localctx.(*FilterExprContext).F2 = _x
				}

			case 2:
				localctx = NewFilterExprContext(p, _parentctx, _parentState)
				localctx.(*FilterExprContext).F1 = _prevctx
				p.PushNewRecursionContext(localctx, _startState, QueryRULE_filterExpr)
				p.SetState(103)

				if !(p.Precpred(p.GetParserRuleContext(), 3)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 3)", ""))
					goto errorExit
				}
				{
					p.SetState(104)

					var _m = p.Match(QueryOR_OP)

					localctx.(*FilterExprContext).Op = _m
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(105)

					var _x = p.filterExpr(4)

					localctx.(*FilterExprContext).F2 = _x
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(110)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 12, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.UnrollRecursionContexts(_parentctx)
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IFilterStmtContext is an interface to support dynamic dispatch.
type IFilterStmtContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetExpr returns the Expr rule contexts.
	GetExpr() IFilterExprContext

	// GetName returns the Name rule contexts.
	GetName() IIdentContext

	// SetExpr sets the Expr rule contexts.
	SetExpr(IFilterExprContext)

	// SetName sets the Name rule contexts.
	SetName(IIdentContext)

	// Getter signatures
	FILTER() antlr.TerminalNode
	AS() antlr.TerminalNode
	FilterExpr() IFilterExprContext
	Ident() IIdentContext

	// IsFilterStmtContext differentiates from other interfaces.
	IsFilterStmtContext()
}

type FilterStmtContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
	Expr   IFilterExprContext
	Name   IIdentContext
}

func NewEmptyFilterStmtContext() *FilterStmtContext {
	var p = new(FilterStmtContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_filterStmt
	return p
}

func InitEmptyFilterStmtContext(p *FilterStmtContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_filterStmt
}

func (*FilterStmtContext) IsFilterStmtContext() {}

func NewFilterStmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FilterStmtContext {
	var p = new(FilterStmtContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryRULE_filterStmt

	return p
}

func (s *FilterStmtContext) GetParser() antlr.Parser { return s.parser }

func (s *FilterStmtContext) GetExpr() IFilterExprContext { return s.Expr }

func (s *FilterStmtContext) GetName() IIdentContext { return s.Name }

func (s *FilterStmtContext) SetExpr(v IFilterExprContext) { s.Expr = v }

func (s *FilterStmtContext) SetName(v IIdentContext) { s.Name = v }

func (s *FilterStmtContext) FILTER() antlr.TerminalNode {
	return s.GetToken(QueryFILTER, 0)
}

func (s *FilterStmtContext) AS() antlr.TerminalNode {
	return s.GetToken(QueryAS, 0)
}

func (s *FilterStmtContext) FilterExpr() IFilterExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFilterExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFilterExprContext)
}

func (s *FilterStmtContext) Ident() IIdentContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentContext)
}

func (s *FilterStmtContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FilterStmtContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *FilterStmtContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.EnterFilterStmt(s)
	}
}

func (s *FilterStmtContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.ExitFilterStmt(s)
	}
}

func (s *FilterStmtContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryVisitor:
		return t.VisitFilterStmt(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *Query) FilterStmt() (localctx IFilterStmtContext) {
	localctx = NewFilterStmtContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, QueryRULE_filterStmt)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(111)
		p.Match(QueryFILTER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(112)

		var _x = p.filterExpr(0)

		localctx.(*FilterStmtContext).Expr = _x
	}
	{
		p.SetState(113)
		p.Match(QueryAS)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(114)

		var _x = p.Ident()

		localctx.(*FilterStmtContext).Name = _x
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IExprContext is an interface to support dynamic dispatch.
type IExprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetFilter returns the Filter rule contexts.
	GetFilter() IIdentContext

	// GetKey returns the Key rule contexts.
	GetKey() IFilterKeyContext

	// GetValue returns the Value rule contexts.
	GetValue() IFilterValueContext

	// SetFilter sets the Filter rule contexts.
	SetFilter(IIdentContext)

	// SetKey sets the Key rule contexts.
	SetKey(IFilterKeyContext)

	// SetValue sets the Value rule contexts.
	SetValue(IFilterValueContext)

	// Getter signatures
	AT() antlr.TerminalNode
	Ident() IIdentContext
	SIMPLE_OP() antlr.TerminalNode
	FilterKey() IFilterKeyContext
	FilterValue() IFilterValueContext

	// IsExprContext differentiates from other interfaces.
	IsExprContext()
}

type ExprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
	Filter IIdentContext
	Key    IFilterKeyContext
	Value  IFilterValueContext
}

func NewEmptyExprContext() *ExprContext {
	var p = new(ExprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_expr
	return p
}

func InitEmptyExprContext(p *ExprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_expr
}

func (*ExprContext) IsExprContext() {}

func NewExprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExprContext {
	var p = new(ExprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryRULE_expr

	return p
}

func (s *ExprContext) GetParser() antlr.Parser { return s.parser }

func (s *ExprContext) GetFilter() IIdentContext { return s.Filter }

func (s *ExprContext) GetKey() IFilterKeyContext { return s.Key }

func (s *ExprContext) GetValue() IFilterValueContext { return s.Value }

func (s *ExprContext) SetFilter(v IIdentContext) { s.Filter = v }

func (s *ExprContext) SetKey(v IFilterKeyContext) { s.Key = v }

func (s *ExprContext) SetValue(v IFilterValueContext) { s.Value = v }

func (s *ExprContext) AT() antlr.TerminalNode {
	return s.GetToken(QueryAT, 0)
}

func (s *ExprContext) Ident() IIdentContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentContext)
}

func (s *ExprContext) SIMPLE_OP() antlr.TerminalNode {
	return s.GetToken(QuerySIMPLE_OP, 0)
}

func (s *ExprContext) FilterKey() IFilterKeyContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFilterKeyContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFilterKeyContext)
}

func (s *ExprContext) FilterValue() IFilterValueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFilterValueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFilterValueContext)
}

func (s *ExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.EnterExpr(s)
	}
}

func (s *ExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.ExitExpr(s)
	}
}

func (s *ExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryVisitor:
		return t.VisitExpr(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *Query) Expr() (localctx IExprContext) {
	localctx = NewExprContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, QueryRULE_expr)
	p.SetState(122)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case QueryAT:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(116)
			p.Match(QueryAT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(117)

			var _x = p.Ident()

			localctx.(*ExprContext).Filter = _x
		}

	case QueryREP, QueryIN, QueryAS, QuerySELECT, QueryFROM, QueryFILTER, QueryEC, QueryIDENT, QuerySTRING:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(118)

			var _x = p.FilterKey()

			localctx.(*ExprContext).Key = _x
		}
		{
			p.SetState(119)
			p.Match(QuerySIMPLE_OP)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(120)

			var _x = p.FilterValue()

			localctx.(*ExprContext).Value = _x
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IFilterKeyContext is an interface to support dynamic dispatch.
type IFilterKeyContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Ident() IIdentContext
	STRING() antlr.TerminalNode

	// IsFilterKeyContext differentiates from other interfaces.
	IsFilterKeyContext()
}

type FilterKeyContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFilterKeyContext() *FilterKeyContext {
	var p = new(FilterKeyContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_filterKey
	return p
}

func InitEmptyFilterKeyContext(p *FilterKeyContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_filterKey
}

func (*FilterKeyContext) IsFilterKeyContext() {}

func NewFilterKeyContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FilterKeyContext {
	var p = new(FilterKeyContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryRULE_filterKey

	return p
}

func (s *FilterKeyContext) GetParser() antlr.Parser { return s.parser }

func (s *FilterKeyContext) Ident() IIdentContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentContext)
}

func (s *FilterKeyContext) STRING() antlr.TerminalNode {
	return s.GetToken(QuerySTRING, 0)
}

func (s *FilterKeyContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FilterKeyContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *FilterKeyContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.EnterFilterKey(s)
	}
}

func (s *FilterKeyContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.ExitFilterKey(s)
	}
}

func (s *FilterKeyContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryVisitor:
		return t.VisitFilterKey(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *Query) FilterKey() (localctx IFilterKeyContext) {
	localctx = NewFilterKeyContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, QueryRULE_filterKey)
	p.SetState(126)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case QueryREP, QueryIN, QueryAS, QuerySELECT, QueryFROM, QueryFILTER, QueryEC, QueryIDENT:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(124)
			p.Ident()
		}

	case QuerySTRING:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(125)
			p.Match(QuerySTRING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IFilterValueContext is an interface to support dynamic dispatch.
type IFilterValueContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Ident() IIdentContext
	Number() INumberContext
	STRING() antlr.TerminalNode

	// IsFilterValueContext differentiates from other interfaces.
	IsFilterValueContext()
}

type FilterValueContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFilterValueContext() *FilterValueContext {
	var p = new(FilterValueContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_filterValue
	return p
}

func InitEmptyFilterValueContext(p *FilterValueContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_filterValue
}

func (*FilterValueContext) IsFilterValueContext() {}

func NewFilterValueContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FilterValueContext {
	var p = new(FilterValueContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryRULE_filterValue

	return p
}

func (s *FilterValueContext) GetParser() antlr.Parser { return s.parser }

func (s *FilterValueContext) Ident() IIdentContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentContext)
}

func (s *FilterValueContext) Number() INumberContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(INumberContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(INumberContext)
}

func (s *FilterValueContext) STRING() antlr.TerminalNode {
	return s.GetToken(QuerySTRING, 0)
}

func (s *FilterValueContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FilterValueContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *FilterValueContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.EnterFilterValue(s)
	}
}

func (s *FilterValueContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.ExitFilterValue(s)
	}
}

func (s *FilterValueContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryVisitor:
		return t.VisitFilterValue(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *Query) FilterValue() (localctx IFilterValueContext) {
	localctx = NewFilterValueContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 22, QueryRULE_filterValue)
	p.SetState(131)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case QueryREP, QueryIN, QueryAS, QuerySELECT, QueryFROM, QueryFILTER, QueryEC, QueryIDENT:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(128)
			p.Ident()
		}

	case QueryNUMBER1, QueryZERO:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(129)
			p.Number()
		}

	case QuerySTRING:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(130)
			p.Match(QuerySTRING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// INumberContext is an interface to support dynamic dispatch.
type INumberContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	ZERO() antlr.TerminalNode
	NUMBER1() antlr.TerminalNode

	// IsNumberContext differentiates from other interfaces.
	IsNumberContext()
}

type NumberContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyNumberContext() *NumberContext {
	var p = new(NumberContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_number
	return p
}

func InitEmptyNumberContext(p *NumberContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_number
}

func (*NumberContext) IsNumberContext() {}

func NewNumberContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *NumberContext {
	var p = new(NumberContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryRULE_number

	return p
}

func (s *NumberContext) GetParser() antlr.Parser { return s.parser }

func (s *NumberContext) ZERO() antlr.TerminalNode {
	return s.GetToken(QueryZERO, 0)
}

func (s *NumberContext) NUMBER1() antlr.TerminalNode {
	return s.GetToken(QueryNUMBER1, 0)
}

func (s *NumberContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NumberContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *NumberContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.EnterNumber(s)
	}
}

func (s *NumberContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.ExitNumber(s)
	}
}

func (s *NumberContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryVisitor:
		return t.VisitNumber(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *Query) Number() (localctx INumberContext) {
	localctx = NewNumberContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 24, QueryRULE_number)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(133)
		_la = p.GetTokenStream().LA(1)

		if !(_la == QueryNUMBER1 || _la == QueryZERO) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IKeywordContext is an interface to support dynamic dispatch.
type IKeywordContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	REP() antlr.TerminalNode
	IN() antlr.TerminalNode
	AS() antlr.TerminalNode
	SELECT() antlr.TerminalNode
	FROM() antlr.TerminalNode
	FILTER() antlr.TerminalNode
	EC() antlr.TerminalNode

	// IsKeywordContext differentiates from other interfaces.
	IsKeywordContext()
}

type KeywordContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyKeywordContext() *KeywordContext {
	var p = new(KeywordContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_keyword
	return p
}

func InitEmptyKeywordContext(p *KeywordContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_keyword
}

func (*KeywordContext) IsKeywordContext() {}

func NewKeywordContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *KeywordContext {
	var p = new(KeywordContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryRULE_keyword

	return p
}

func (s *KeywordContext) GetParser() antlr.Parser { return s.parser }

func (s *KeywordContext) REP() antlr.TerminalNode {
	return s.GetToken(QueryREP, 0)
}

func (s *KeywordContext) IN() antlr.TerminalNode {
	return s.GetToken(QueryIN, 0)
}

func (s *KeywordContext) AS() antlr.TerminalNode {
	return s.GetToken(QueryAS, 0)
}

func (s *KeywordContext) SELECT() antlr.TerminalNode {
	return s.GetToken(QuerySELECT, 0)
}

func (s *KeywordContext) FROM() antlr.TerminalNode {
	return s.GetToken(QueryFROM, 0)
}

func (s *KeywordContext) FILTER() antlr.TerminalNode {
	return s.GetToken(QueryFILTER, 0)
}

func (s *KeywordContext) EC() antlr.TerminalNode {
	return s.GetToken(QueryEC, 0)
}

func (s *KeywordContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *KeywordContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *KeywordContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.EnterKeyword(s)
	}
}

func (s *KeywordContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.ExitKeyword(s)
	}
}

func (s *KeywordContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryVisitor:
		return t.VisitKeyword(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *Query) Keyword() (localctx IKeywordContext) {
	localctx = NewKeywordContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 26, QueryRULE_keyword)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(135)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&6000) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IIdentContext is an interface to support dynamic dispatch.
type IIdentContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Keyword() IKeywordContext
	IDENT() antlr.TerminalNode

	// IsIdentContext differentiates from other interfaces.
	IsIdentContext()
}

type IdentContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyIdentContext() *IdentContext {
	var p = new(IdentContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_ident
	return p
}

func InitEmptyIdentContext(p *IdentContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_ident
}

func (*IdentContext) IsIdentContext() {}

func NewIdentContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *IdentContext {
	var p = new(IdentContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryRULE_ident

	return p
}

func (s *IdentContext) GetParser() antlr.Parser { return s.parser }

func (s *IdentContext) Keyword() IKeywordContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IKeywordContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IKeywordContext)
}

func (s *IdentContext) IDENT() antlr.TerminalNode {
	return s.GetToken(QueryIDENT, 0)
}

func (s *IdentContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *IdentContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *IdentContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.EnterIdent(s)
	}
}

func (s *IdentContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.ExitIdent(s)
	}
}

func (s *IdentContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryVisitor:
		return t.VisitIdent(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *Query) Ident() (localctx IIdentContext) {
	localctx = NewIdentContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 28, QueryRULE_ident)
	p.SetState(139)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case QueryREP, QueryIN, QueryAS, QuerySELECT, QueryFROM, QueryFILTER, QueryEC:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(137)
			p.Keyword()
		}

	case QueryIDENT:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(138)
			p.Match(QueryIDENT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IIdentWCContext is an interface to support dynamic dispatch.
type IIdentWCContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Ident() IIdentContext
	WILDCARD() antlr.TerminalNode

	// IsIdentWCContext differentiates from other interfaces.
	IsIdentWCContext()
}

type IdentWCContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyIdentWCContext() *IdentWCContext {
	var p = new(IdentWCContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_identWC
	return p
}

func InitEmptyIdentWCContext(p *IdentWCContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryRULE_identWC
}

func (*IdentWCContext) IsIdentWCContext() {}

func NewIdentWCContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *IdentWCContext {
	var p = new(IdentWCContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryRULE_identWC

	return p
}

func (s *IdentWCContext) GetParser() antlr.Parser { return s.parser }

func (s *IdentWCContext) Ident() IIdentContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentContext)
}

func (s *IdentWCContext) WILDCARD() antlr.TerminalNode {
	return s.GetToken(QueryWILDCARD, 0)
}

func (s *IdentWCContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *IdentWCContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *IdentWCContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.EnterIdentWC(s)
	}
}

func (s *IdentWCContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryListener); ok {
		listenerT.ExitIdentWC(s)
	}
}

func (s *IdentWCContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryVisitor:
		return t.VisitIdentWC(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *Query) IdentWC() (localctx IIdentWCContext) {
	localctx = NewIdentWCContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 30, QueryRULE_identWC)
	p.SetState(143)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case QueryREP, QueryIN, QueryAS, QuerySELECT, QueryFROM, QueryFILTER, QueryEC, QueryIDENT:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(141)
			p.Ident()
		}

	case QueryWILDCARD:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(142)
			p.Match(QueryWILDCARD)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

func (p *Query) Sempred(localctx antlr.RuleContext, ruleIndex, predIndex int) bool {
	switch ruleIndex {
	case 7:
		var t *FilterExprContext = nil
		if localctx != nil {
			t = localctx.(*FilterExprContext)
		}
		return p.FilterExpr_Sempred(t, predIndex)

	default:
		panic("No predicate with index: " + fmt.Sprint(ruleIndex))
	}
}

func (p *Query) FilterExpr_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 0:
		return p.Precpred(p.GetParserRuleContext(), 4)

	case 1:
		return p.Precpred(p.GetParserRuleContext(), 3)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
