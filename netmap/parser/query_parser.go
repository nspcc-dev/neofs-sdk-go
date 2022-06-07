// Code generated from Query.g4 by ANTLR 4.9.2. DO NOT EDIT.

package parser // Query

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = reflect.Copy
var _ = strconv.Itoa

var parserATN = []uint16{
	3, 24715, 42794, 33075, 47597, 16764, 15335, 30598, 22884, 3, 23, 130,
	4, 2, 9, 2, 4, 3, 9, 3, 4, 4, 9, 4, 4, 5, 9, 5, 4, 6, 9, 6, 4, 7, 9, 7,
	4, 8, 9, 8, 4, 9, 9, 9, 4, 10, 9, 10, 4, 11, 9, 11, 4, 12, 9, 12, 4, 13,
	9, 13, 4, 14, 9, 14, 4, 15, 9, 15, 3, 2, 6, 2, 32, 10, 2, 13, 2, 14, 2,
	33, 3, 2, 5, 2, 37, 10, 2, 3, 2, 7, 2, 40, 10, 2, 12, 2, 14, 2, 43, 11,
	2, 3, 2, 7, 2, 46, 10, 2, 12, 2, 14, 2, 49, 11, 2, 3, 3, 3, 3, 3, 3, 3,
	3, 5, 3, 55, 10, 3, 3, 4, 3, 4, 3, 4, 3, 5, 3, 5, 3, 5, 3, 5, 5, 5, 64,
	10, 5, 3, 5, 5, 5, 67, 10, 5, 3, 5, 3, 5, 3, 5, 3, 5, 5, 5, 73, 10, 5,
	3, 6, 3, 6, 3, 7, 3, 7, 3, 7, 3, 7, 3, 7, 3, 7, 5, 7, 83, 10, 7, 3, 7,
	3, 7, 3, 7, 3, 7, 3, 7, 3, 7, 7, 7, 91, 10, 7, 12, 7, 14, 7, 94, 11, 7,
	3, 8, 3, 8, 3, 8, 3, 8, 3, 8, 3, 9, 3, 9, 3, 9, 3, 9, 3, 9, 3, 9, 5, 9,
	107, 10, 9, 3, 10, 3, 10, 5, 10, 111, 10, 10, 3, 11, 3, 11, 3, 11, 5, 11,
	116, 10, 11, 3, 12, 3, 12, 3, 13, 3, 13, 3, 14, 3, 14, 5, 14, 124, 10,
	14, 3, 15, 3, 15, 5, 15, 128, 10, 15, 3, 15, 2, 3, 12, 16, 2, 4, 6, 8,
	10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 2, 5, 3, 2, 14, 15, 3, 2, 20, 21,
	4, 2, 6, 8, 10, 12, 2, 132, 2, 31, 3, 2, 2, 2, 4, 50, 3, 2, 2, 2, 6, 56,
	3, 2, 2, 2, 8, 59, 3, 2, 2, 2, 10, 74, 3, 2, 2, 2, 12, 82, 3, 2, 2, 2,
	14, 95, 3, 2, 2, 2, 16, 106, 3, 2, 2, 2, 18, 110, 3, 2, 2, 2, 20, 115,
	3, 2, 2, 2, 22, 117, 3, 2, 2, 2, 24, 119, 3, 2, 2, 2, 26, 123, 3, 2, 2,
	2, 28, 127, 3, 2, 2, 2, 30, 32, 5, 4, 3, 2, 31, 30, 3, 2, 2, 2, 32, 33,
	3, 2, 2, 2, 33, 31, 3, 2, 2, 2, 33, 34, 3, 2, 2, 2, 34, 36, 3, 2, 2, 2,
	35, 37, 5, 6, 4, 2, 36, 35, 3, 2, 2, 2, 36, 37, 3, 2, 2, 2, 37, 41, 3,
	2, 2, 2, 38, 40, 5, 8, 5, 2, 39, 38, 3, 2, 2, 2, 40, 43, 3, 2, 2, 2, 41,
	39, 3, 2, 2, 2, 41, 42, 3, 2, 2, 2, 42, 47, 3, 2, 2, 2, 43, 41, 3, 2, 2,
	2, 44, 46, 5, 14, 8, 2, 45, 44, 3, 2, 2, 2, 46, 49, 3, 2, 2, 2, 47, 45,
	3, 2, 2, 2, 47, 48, 3, 2, 2, 2, 48, 3, 3, 2, 2, 2, 49, 47, 3, 2, 2, 2,
	50, 51, 7, 6, 2, 2, 51, 54, 7, 20, 2, 2, 52, 53, 7, 7, 2, 2, 53, 55, 5,
	26, 14, 2, 54, 52, 3, 2, 2, 2, 54, 55, 3, 2, 2, 2, 55, 5, 3, 2, 2, 2, 56,
	57, 7, 9, 2, 2, 57, 58, 7, 20, 2, 2, 58, 7, 3, 2, 2, 2, 59, 60, 7, 10,
	2, 2, 60, 66, 7, 20, 2, 2, 61, 63, 7, 7, 2, 2, 62, 64, 5, 10, 6, 2, 63,
	62, 3, 2, 2, 2, 63, 64, 3, 2, 2, 2, 64, 65, 3, 2, 2, 2, 65, 67, 5, 26,
	14, 2, 66, 61, 3, 2, 2, 2, 66, 67, 3, 2, 2, 2, 67, 68, 3, 2, 2, 2, 68,
	69, 7, 11, 2, 2, 69, 72, 5, 28, 15, 2, 70, 71, 7, 8, 2, 2, 71, 73, 5, 26,
	14, 2, 72, 70, 3, 2, 2, 2, 72, 73, 3, 2, 2, 2, 73, 9, 3, 2, 2, 2, 74, 75,
	9, 2, 2, 2, 75, 11, 3, 2, 2, 2, 76, 77, 8, 7, 1, 2, 77, 78, 7, 16, 2, 2,
	78, 79, 5, 12, 7, 2, 79, 80, 7, 17, 2, 2, 80, 83, 3, 2, 2, 2, 81, 83, 5,
	16, 9, 2, 82, 76, 3, 2, 2, 2, 82, 81, 3, 2, 2, 2, 83, 92, 3, 2, 2, 2, 84,
	85, 12, 6, 2, 2, 85, 86, 7, 3, 2, 2, 86, 91, 5, 12, 7, 7, 87, 88, 12, 5,
	2, 2, 88, 89, 7, 4, 2, 2, 89, 91, 5, 12, 7, 6, 90, 84, 3, 2, 2, 2, 90,
	87, 3, 2, 2, 2, 91, 94, 3, 2, 2, 2, 92, 90, 3, 2, 2, 2, 92, 93, 3, 2, 2,
	2, 93, 13, 3, 2, 2, 2, 94, 92, 3, 2, 2, 2, 95, 96, 7, 12, 2, 2, 96, 97,
	5, 12, 7, 2, 97, 98, 7, 8, 2, 2, 98, 99, 5, 26, 14, 2, 99, 15, 3, 2, 2,
	2, 100, 101, 7, 18, 2, 2, 101, 107, 5, 26, 14, 2, 102, 103, 5, 18, 10,
	2, 103, 104, 7, 5, 2, 2, 104, 105, 5, 20, 11, 2, 105, 107, 3, 2, 2, 2,
	106, 100, 3, 2, 2, 2, 106, 102, 3, 2, 2, 2, 107, 17, 3, 2, 2, 2, 108, 111,
	5, 26, 14, 2, 109, 111, 7, 22, 2, 2, 110, 108, 3, 2, 2, 2, 110, 109, 3,
	2, 2, 2, 111, 19, 3, 2, 2, 2, 112, 116, 5, 26, 14, 2, 113, 116, 5, 22,
	12, 2, 114, 116, 7, 22, 2, 2, 115, 112, 3, 2, 2, 2, 115, 113, 3, 2, 2,
	2, 115, 114, 3, 2, 2, 2, 116, 21, 3, 2, 2, 2, 117, 118, 9, 3, 2, 2, 118,
	23, 3, 2, 2, 2, 119, 120, 9, 4, 2, 2, 120, 25, 3, 2, 2, 2, 121, 124, 5,
	24, 13, 2, 122, 124, 7, 19, 2, 2, 123, 121, 3, 2, 2, 2, 123, 122, 3, 2,
	2, 2, 124, 27, 3, 2, 2, 2, 125, 128, 5, 26, 14, 2, 126, 128, 7, 13, 2,
	2, 127, 125, 3, 2, 2, 2, 127, 126, 3, 2, 2, 2, 128, 29, 3, 2, 2, 2, 18,
	33, 36, 41, 47, 54, 63, 66, 72, 82, 90, 92, 106, 110, 115, 123, 127,
}
var literalNames = []string{
	"", "'AND'", "'OR'", "", "'REP'", "'IN'", "'AS'", "'CBF'", "'SELECT'",
	"'FROM'", "'FILTER'", "'*'", "'SAME'", "'DISTINCT'", "'('", "')'", "'@'",
	"", "", "'0'",
}
var symbolicNames = []string{
	"", "AND_OP", "OR_OP", "SIMPLE_OP", "REP", "IN", "AS", "CBF", "SELECT",
	"FROM", "FILTER", "WILDCARD", "CLAUSE_SAME", "CLAUSE_DISTINCT", "L_PAREN",
	"R_PAREN", "AT", "IDENT", "NUMBER1", "ZERO", "STRING", "WS",
}

var ruleNames = []string{
	"policy", "repStmt", "cbfStmt", "selectStmt", "clause", "filterExpr", "filterStmt",
	"expr", "filterKey", "filterValue", "number", "keyword", "ident", "identWC",
}

type Query struct {
	*antlr.BaseParser
}

// NewQuery produces a new parser instance for the optional input antlr.TokenStream.
//
// The *Query instance produced may be reused by calling the SetInputStream method.
// The initial parser configuration is expensive to construct, and the object is not thread-safe;
// however, if used within a Golang sync.Pool, the construction cost amortizes well and the
// objects can be used in a thread-safe manner.
func NewQuery(input antlr.TokenStream) *Query {
	this := new(Query)
	deserializer := antlr.NewATNDeserializer(nil)
	deserializedATN := deserializer.DeserializeFromUInt16(parserATN)
	decisionToDFA := make([]*antlr.DFA, len(deserializedATN.DecisionToState))
	for index, ds := range deserializedATN.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(ds, index)
	}
	this.BaseParser = antlr.NewBaseParser(input)

	this.Interpreter = antlr.NewParserATNSimulator(this, deserializedATN, decisionToDFA, antlr.NewPredictionContextCache())
	this.RuleNames = ruleNames
	this.LiteralNames = literalNames
	this.SymbolicNames = symbolicNames
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
	QueryCLAUSE_SAME     = 12
	QueryCLAUSE_DISTINCT = 13
	QueryL_PAREN         = 14
	QueryR_PAREN         = 15
	QueryAT              = 16
	QueryIDENT           = 17
	QueryNUMBER1         = 18
	QueryZERO            = 19
	QuerySTRING          = 20
	QueryWS              = 21
)

// Query rules.
const (
	QueryRULE_policy      = 0
	QueryRULE_repStmt     = 1
	QueryRULE_cbfStmt     = 2
	QueryRULE_selectStmt  = 3
	QueryRULE_clause      = 4
	QueryRULE_filterExpr  = 5
	QueryRULE_filterStmt  = 6
	QueryRULE_expr        = 7
	QueryRULE_filterKey   = 8
	QueryRULE_filterValue = 9
	QueryRULE_number      = 10
	QueryRULE_keyword     = 11
	QueryRULE_ident       = 12
	QueryRULE_identWC     = 13
)

// IPolicyContext is an interface to support dynamic dispatch.
type IPolicyContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsPolicyContext differentiates from other interfaces.
	IsPolicyContext()
}

type PolicyContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPolicyContext() *PolicyContext {
	var p = new(PolicyContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = QueryRULE_policy
	return p
}

func (*PolicyContext) IsPolicyContext() {}

func NewPolicyContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PolicyContext {
	var p = new(PolicyContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryRULE_policy

	return p
}

func (s *PolicyContext) GetParser() antlr.Parser { return s.parser }

func (s *PolicyContext) AllRepStmt() []IRepStmtContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IRepStmtContext)(nil)).Elem())
	var tst = make([]IRepStmtContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IRepStmtContext)
		}
	}

	return tst
}

func (s *PolicyContext) RepStmt(i int) IRepStmtContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IRepStmtContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IRepStmtContext)
}

func (s *PolicyContext) CbfStmt() ICbfStmtContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ICbfStmtContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ICbfStmtContext)
}

func (s *PolicyContext) AllSelectStmt() []ISelectStmtContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*ISelectStmtContext)(nil)).Elem())
	var tst = make([]ISelectStmtContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(ISelectStmtContext)
		}
	}

	return tst
}

func (s *PolicyContext) SelectStmt(i int) ISelectStmtContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISelectStmtContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(ISelectStmtContext)
}

func (s *PolicyContext) AllFilterStmt() []IFilterStmtContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IFilterStmtContext)(nil)).Elem())
	var tst = make([]IFilterStmtContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IFilterStmtContext)
		}
	}

	return tst
}

func (s *PolicyContext) FilterStmt(i int) IFilterStmtContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IFilterStmtContext)(nil)).Elem(), i)

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

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	p.SetState(29)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == QueryREP {
		{
			p.SetState(28)
			p.RepStmt()
		}

		p.SetState(31)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(34)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == QueryCBF {
		{
			p.SetState(33)
			p.CbfStmt()
		}

	}
	p.SetState(39)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for _la == QuerySELECT {
		{
			p.SetState(36)
			p.SelectStmt()
		}

		p.SetState(41)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(45)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for _la == QueryFILTER {
		{
			p.SetState(42)
			p.FilterStmt()
		}

		p.SetState(47)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}

	return localctx
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

	// IsRepStmtContext differentiates from other interfaces.
	IsRepStmtContext()
}

type RepStmtContext struct {
	*antlr.BaseParserRuleContext
	parser   antlr.Parser
	Count    antlr.Token
	Selector IIdentContext
}

func NewEmptyRepStmtContext() *RepStmtContext {
	var p = new(RepStmtContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = QueryRULE_repStmt
	return p
}

func (*RepStmtContext) IsRepStmtContext() {}

func NewRepStmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *RepStmtContext {
	var p = new(RepStmtContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

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
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IIdentContext)(nil)).Elem(), 0)

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
	p.EnterRule(localctx, 2, QueryRULE_repStmt)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(48)
		p.Match(QueryREP)
	}
	{
		p.SetState(49)

		var _m = p.Match(QueryNUMBER1)

		localctx.(*RepStmtContext).Count = _m
	}
	p.SetState(52)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == QueryIN {
		{
			p.SetState(50)
			p.Match(QueryIN)
		}
		{
			p.SetState(51)

			var _x = p.Ident()

			localctx.(*RepStmtContext).Selector = _x
		}

	}

	return localctx
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

	// IsCbfStmtContext differentiates from other interfaces.
	IsCbfStmtContext()
}

type CbfStmtContext struct {
	*antlr.BaseParserRuleContext
	parser       antlr.Parser
	BackupFactor antlr.Token
}

func NewEmptyCbfStmtContext() *CbfStmtContext {
	var p = new(CbfStmtContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = QueryRULE_cbfStmt
	return p
}

func (*CbfStmtContext) IsCbfStmtContext() {}

func NewCbfStmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CbfStmtContext {
	var p = new(CbfStmtContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

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
	p.EnterRule(localctx, 4, QueryRULE_cbfStmt)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(54)
		p.Match(QueryCBF)
	}
	{
		p.SetState(55)

		var _m = p.Match(QueryNUMBER1)

		localctx.(*CbfStmtContext).BackupFactor = _m
	}

	return localctx
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

	// IsSelectStmtContext differentiates from other interfaces.
	IsSelectStmtContext()
}

type SelectStmtContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
	Count  antlr.Token
	Bucket IIdentContext
	Filter IIdentWCContext
	Name   IIdentContext
}

func NewEmptySelectStmtContext() *SelectStmtContext {
	var p = new(SelectStmtContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = QueryRULE_selectStmt
	return p
}

func (*SelectStmtContext) IsSelectStmtContext() {}

func NewSelectStmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SelectStmtContext {
	var p = new(SelectStmtContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

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
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IIdentWCContext)(nil)).Elem(), 0)

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
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IIdentContext)(nil)).Elem())
	var tst = make([]IIdentContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IIdentContext)
		}
	}

	return tst
}

func (s *SelectStmtContext) Ident(i int) IIdentContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IIdentContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IIdentContext)
}

func (s *SelectStmtContext) Clause() IClauseContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IClauseContext)(nil)).Elem(), 0)

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
	p.EnterRule(localctx, 6, QueryRULE_selectStmt)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(57)
		p.Match(QuerySELECT)
	}
	{
		p.SetState(58)

		var _m = p.Match(QueryNUMBER1)

		localctx.(*SelectStmtContext).Count = _m
	}
	p.SetState(64)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == QueryIN {
		{
			p.SetState(59)
			p.Match(QueryIN)
		}
		p.SetState(61)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)

		if _la == QueryCLAUSE_SAME || _la == QueryCLAUSE_DISTINCT {
			{
				p.SetState(60)
				p.Clause()
			}

		}
		{
			p.SetState(63)

			var _x = p.Ident()

			localctx.(*SelectStmtContext).Bucket = _x
		}

	}
	{
		p.SetState(66)
		p.Match(QueryFROM)
	}
	{
		p.SetState(67)

		var _x = p.IdentWC()

		localctx.(*SelectStmtContext).Filter = _x
	}
	p.SetState(70)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == QueryAS {
		{
			p.SetState(68)
			p.Match(QueryAS)
		}
		{
			p.SetState(69)

			var _x = p.Ident()

			localctx.(*SelectStmtContext).Name = _x
		}

	}

	return localctx
}

// IClauseContext is an interface to support dynamic dispatch.
type IClauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsClauseContext differentiates from other interfaces.
	IsClauseContext()
}

type ClauseContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyClauseContext() *ClauseContext {
	var p = new(ClauseContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = QueryRULE_clause
	return p
}

func (*ClauseContext) IsClauseContext() {}

func NewClauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ClauseContext {
	var p = new(ClauseContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

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
	p.EnterRule(localctx, 8, QueryRULE_clause)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(72)
		_la = p.GetTokenStream().LA(1)

		if !(_la == QueryCLAUSE_SAME || _la == QueryCLAUSE_DISTINCT) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

	return localctx
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

	// IsFilterExprContext differentiates from other interfaces.
	IsFilterExprContext()
}

type FilterExprContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
	F1     IFilterExprContext
	Inner  IFilterExprContext
	Op     antlr.Token
	F2     IFilterExprContext
}

func NewEmptyFilterExprContext() *FilterExprContext {
	var p = new(FilterExprContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = QueryRULE_filterExpr
	return p
}

func (*FilterExprContext) IsFilterExprContext() {}

func NewFilterExprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FilterExprContext {
	var p = new(FilterExprContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

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
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IFilterExprContext)(nil)).Elem())
	var tst = make([]IFilterExprContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IFilterExprContext)
		}
	}

	return tst
}

func (s *FilterExprContext) FilterExpr(i int) IFilterExprContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IFilterExprContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IFilterExprContext)
}

func (s *FilterExprContext) Expr() IExprContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExprContext)(nil)).Elem(), 0)

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
	_startState := 10
	p.EnterRecursionRule(localctx, 10, QueryRULE_filterExpr, _p)

	defer func() {
		p.UnrollRecursionContexts(_parentctx)
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(80)
	p.GetErrorHandler().Sync(p)

	switch p.GetTokenStream().LA(1) {
	case QueryL_PAREN:
		{
			p.SetState(75)
			p.Match(QueryL_PAREN)
		}
		{
			p.SetState(76)

			var _x = p.filterExpr(0)

			localctx.(*FilterExprContext).Inner = _x
		}
		{
			p.SetState(77)
			p.Match(QueryR_PAREN)
		}

	case QueryREP, QueryIN, QueryAS, QuerySELECT, QueryFROM, QueryFILTER, QueryAT, QueryIDENT, QuerySTRING:
		{
			p.SetState(79)
			p.Expr()
		}

	default:
		panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(90)
	p.GetErrorHandler().Sync(p)
	_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 10, p.GetParserRuleContext())

	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(88)
			p.GetErrorHandler().Sync(p)
			switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 9, p.GetParserRuleContext()) {
			case 1:
				localctx = NewFilterExprContext(p, _parentctx, _parentState)
				localctx.(*FilterExprContext).F1 = _prevctx
				p.PushNewRecursionContext(localctx, _startState, QueryRULE_filterExpr)
				p.SetState(82)

				if !(p.Precpred(p.GetParserRuleContext(), 4)) {
					panic(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 4)", ""))
				}
				{
					p.SetState(83)

					var _m = p.Match(QueryAND_OP)

					localctx.(*FilterExprContext).Op = _m
				}
				{
					p.SetState(84)

					var _x = p.filterExpr(5)

					localctx.(*FilterExprContext).F2 = _x
				}

			case 2:
				localctx = NewFilterExprContext(p, _parentctx, _parentState)
				localctx.(*FilterExprContext).F1 = _prevctx
				p.PushNewRecursionContext(localctx, _startState, QueryRULE_filterExpr)
				p.SetState(85)

				if !(p.Precpred(p.GetParserRuleContext(), 3)) {
					panic(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 3)", ""))
				}
				{
					p.SetState(86)

					var _m = p.Match(QueryOR_OP)

					localctx.(*FilterExprContext).Op = _m
				}
				{
					p.SetState(87)

					var _x = p.filterExpr(4)

					localctx.(*FilterExprContext).F2 = _x
				}

			}

		}
		p.SetState(92)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 10, p.GetParserRuleContext())
	}

	return localctx
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

	// IsFilterStmtContext differentiates from other interfaces.
	IsFilterStmtContext()
}

type FilterStmtContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
	Expr   IFilterExprContext
	Name   IIdentContext
}

func NewEmptyFilterStmtContext() *FilterStmtContext {
	var p = new(FilterStmtContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = QueryRULE_filterStmt
	return p
}

func (*FilterStmtContext) IsFilterStmtContext() {}

func NewFilterStmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FilterStmtContext {
	var p = new(FilterStmtContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

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
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IFilterExprContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IFilterExprContext)
}

func (s *FilterStmtContext) Ident() IIdentContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IIdentContext)(nil)).Elem(), 0)

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
	p.EnterRule(localctx, 12, QueryRULE_filterStmt)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(93)
		p.Match(QueryFILTER)
	}
	{
		p.SetState(94)

		var _x = p.filterExpr(0)

		localctx.(*FilterStmtContext).Expr = _x
	}
	{
		p.SetState(95)
		p.Match(QueryAS)
	}
	{
		p.SetState(96)

		var _x = p.Ident()

		localctx.(*FilterStmtContext).Name = _x
	}

	return localctx
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

	// IsExprContext differentiates from other interfaces.
	IsExprContext()
}

type ExprContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
	Filter IIdentContext
	Key    IFilterKeyContext
	Value  IFilterValueContext
}

func NewEmptyExprContext() *ExprContext {
	var p = new(ExprContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = QueryRULE_expr
	return p
}

func (*ExprContext) IsExprContext() {}

func NewExprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExprContext {
	var p = new(ExprContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

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
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IIdentContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IIdentContext)
}

func (s *ExprContext) SIMPLE_OP() antlr.TerminalNode {
	return s.GetToken(QuerySIMPLE_OP, 0)
}

func (s *ExprContext) FilterKey() IFilterKeyContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IFilterKeyContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IFilterKeyContext)
}

func (s *ExprContext) FilterValue() IFilterValueContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IFilterValueContext)(nil)).Elem(), 0)

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
	p.EnterRule(localctx, 14, QueryRULE_expr)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.SetState(104)
	p.GetErrorHandler().Sync(p)

	switch p.GetTokenStream().LA(1) {
	case QueryAT:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(98)
			p.Match(QueryAT)
		}
		{
			p.SetState(99)

			var _x = p.Ident()

			localctx.(*ExprContext).Filter = _x
		}

	case QueryREP, QueryIN, QueryAS, QuerySELECT, QueryFROM, QueryFILTER, QueryIDENT, QuerySTRING:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(100)

			var _x = p.FilterKey()

			localctx.(*ExprContext).Key = _x
		}
		{
			p.SetState(101)
			p.Match(QuerySIMPLE_OP)
		}
		{
			p.SetState(102)

			var _x = p.FilterValue()

			localctx.(*ExprContext).Value = _x
		}

	default:
		panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
	}

	return localctx
}

// IFilterKeyContext is an interface to support dynamic dispatch.
type IFilterKeyContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsFilterKeyContext differentiates from other interfaces.
	IsFilterKeyContext()
}

type FilterKeyContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFilterKeyContext() *FilterKeyContext {
	var p = new(FilterKeyContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = QueryRULE_filterKey
	return p
}

func (*FilterKeyContext) IsFilterKeyContext() {}

func NewFilterKeyContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FilterKeyContext {
	var p = new(FilterKeyContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryRULE_filterKey

	return p
}

func (s *FilterKeyContext) GetParser() antlr.Parser { return s.parser }

func (s *FilterKeyContext) Ident() IIdentContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IIdentContext)(nil)).Elem(), 0)

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
	p.EnterRule(localctx, 16, QueryRULE_filterKey)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.SetState(108)
	p.GetErrorHandler().Sync(p)

	switch p.GetTokenStream().LA(1) {
	case QueryREP, QueryIN, QueryAS, QuerySELECT, QueryFROM, QueryFILTER, QueryIDENT:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(106)
			p.Ident()
		}

	case QuerySTRING:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(107)
			p.Match(QuerySTRING)
		}

	default:
		panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
	}

	return localctx
}

// IFilterValueContext is an interface to support dynamic dispatch.
type IFilterValueContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsFilterValueContext differentiates from other interfaces.
	IsFilterValueContext()
}

type FilterValueContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFilterValueContext() *FilterValueContext {
	var p = new(FilterValueContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = QueryRULE_filterValue
	return p
}

func (*FilterValueContext) IsFilterValueContext() {}

func NewFilterValueContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FilterValueContext {
	var p = new(FilterValueContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryRULE_filterValue

	return p
}

func (s *FilterValueContext) GetParser() antlr.Parser { return s.parser }

func (s *FilterValueContext) Ident() IIdentContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IIdentContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IIdentContext)
}

func (s *FilterValueContext) Number() INumberContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*INumberContext)(nil)).Elem(), 0)

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
	p.EnterRule(localctx, 18, QueryRULE_filterValue)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.SetState(113)
	p.GetErrorHandler().Sync(p)

	switch p.GetTokenStream().LA(1) {
	case QueryREP, QueryIN, QueryAS, QuerySELECT, QueryFROM, QueryFILTER, QueryIDENT:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(110)
			p.Ident()
		}

	case QueryNUMBER1, QueryZERO:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(111)
			p.Number()
		}

	case QuerySTRING:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(112)
			p.Match(QuerySTRING)
		}

	default:
		panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
	}

	return localctx
}

// INumberContext is an interface to support dynamic dispatch.
type INumberContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsNumberContext differentiates from other interfaces.
	IsNumberContext()
}

type NumberContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyNumberContext() *NumberContext {
	var p = new(NumberContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = QueryRULE_number
	return p
}

func (*NumberContext) IsNumberContext() {}

func NewNumberContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *NumberContext {
	var p = new(NumberContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

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
	p.EnterRule(localctx, 20, QueryRULE_number)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(115)
		_la = p.GetTokenStream().LA(1)

		if !(_la == QueryNUMBER1 || _la == QueryZERO) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

	return localctx
}

// IKeywordContext is an interface to support dynamic dispatch.
type IKeywordContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsKeywordContext differentiates from other interfaces.
	IsKeywordContext()
}

type KeywordContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyKeywordContext() *KeywordContext {
	var p = new(KeywordContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = QueryRULE_keyword
	return p
}

func (*KeywordContext) IsKeywordContext() {}

func NewKeywordContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *KeywordContext {
	var p = new(KeywordContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

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
	p.EnterRule(localctx, 22, QueryRULE_keyword)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(117)
		_la = p.GetTokenStream().LA(1)

		if !(((_la)&-(0x1f+1)) == 0 && ((1<<uint(_la))&((1<<QueryREP)|(1<<QueryIN)|(1<<QueryAS)|(1<<QuerySELECT)|(1<<QueryFROM)|(1<<QueryFILTER))) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

	return localctx
}

// IIdentContext is an interface to support dynamic dispatch.
type IIdentContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsIdentContext differentiates from other interfaces.
	IsIdentContext()
}

type IdentContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyIdentContext() *IdentContext {
	var p = new(IdentContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = QueryRULE_ident
	return p
}

func (*IdentContext) IsIdentContext() {}

func NewIdentContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *IdentContext {
	var p = new(IdentContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryRULE_ident

	return p
}

func (s *IdentContext) GetParser() antlr.Parser { return s.parser }

func (s *IdentContext) Keyword() IKeywordContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IKeywordContext)(nil)).Elem(), 0)

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
	p.EnterRule(localctx, 24, QueryRULE_ident)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.SetState(121)
	p.GetErrorHandler().Sync(p)

	switch p.GetTokenStream().LA(1) {
	case QueryREP, QueryIN, QueryAS, QuerySELECT, QueryFROM, QueryFILTER:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(119)
			p.Keyword()
		}

	case QueryIDENT:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(120)
			p.Match(QueryIDENT)
		}

	default:
		panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
	}

	return localctx
}

// IIdentWCContext is an interface to support dynamic dispatch.
type IIdentWCContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsIdentWCContext differentiates from other interfaces.
	IsIdentWCContext()
}

type IdentWCContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyIdentWCContext() *IdentWCContext {
	var p = new(IdentWCContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = QueryRULE_identWC
	return p
}

func (*IdentWCContext) IsIdentWCContext() {}

func NewIdentWCContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *IdentWCContext {
	var p = new(IdentWCContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryRULE_identWC

	return p
}

func (s *IdentWCContext) GetParser() antlr.Parser { return s.parser }

func (s *IdentWCContext) Ident() IIdentContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IIdentContext)(nil)).Elem(), 0)

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
	p.EnterRule(localctx, 26, QueryRULE_identWC)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.SetState(125)
	p.GetErrorHandler().Sync(p)

	switch p.GetTokenStream().LA(1) {
	case QueryREP, QueryIN, QueryAS, QuerySELECT, QueryFROM, QueryFILTER, QueryIDENT:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(123)
			p.Ident()
		}

	case QueryWILDCARD:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(124)
			p.Match(QueryWILDCARD)
		}

	default:
		panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
	}

	return localctx
}

func (p *Query) Sempred(localctx antlr.RuleContext, ruleIndex, predIndex int) bool {
	switch ruleIndex {
	case 5:
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
