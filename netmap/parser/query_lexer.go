// Code generated from QueryLexer.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"sync"
	"unicode"
)

// Suppress unused import error
var _ = fmt.Printf
var _ = sync.Once{}
var _ = unicode.IsLetter

type QueryLexer struct {
	*antlr.BaseLexer
	channelNames []string
	modeNames    []string
	// TODO: EOF string
}

var QueryLexerLexerStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	ChannelNames           []string
	ModeNames              []string
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func querylexerLexerInit() {
	staticData := &QueryLexerLexerStaticData
	staticData.ChannelNames = []string{
		"DEFAULT_TOKEN_CHANNEL", "HIDDEN",
	}
	staticData.ModeNames = []string{
		"DEFAULT_MODE",
	}
	staticData.LiteralNames = []string{
		"", "'AND'", "'OR'", "", "'REP'", "'IN'", "'AS'", "'CBF'", "'SELECT'",
		"'FROM'", "'FILTER'", "'*'", "'EC'", "'MINPARTS'", "'SAME'", "'DISTINCT'",
		"'('", "')'", "'@'", "'/'", "", "", "'0'",
	}
	staticData.SymbolicNames = []string{
		"", "AND_OP", "OR_OP", "SIMPLE_OP", "REP", "IN", "AS", "CBF", "SELECT",
		"FROM", "FILTER", "WILDCARD", "EC", "MINPARTS", "CLAUSE_SAME", "CLAUSE_DISTINCT",
		"L_PAREN", "R_PAREN", "AT", "EC_SEP", "IDENT", "NUMBER1", "ZERO", "STRING",
		"WS",
	}
	staticData.RuleNames = []string{
		"AND_OP", "OR_OP", "SIMPLE_OP", "REP", "IN", "AS", "CBF", "SELECT",
		"FROM", "FILTER", "WILDCARD", "EC", "MINPARTS", "CLAUSE_SAME", "CLAUSE_DISTINCT",
		"L_PAREN", "R_PAREN", "AT", "EC_SEP", "IDENT", "Digit", "Nondigit",
		"NUMBER1", "ZERO", "STRING", "ESC", "UNICODE", "HEX", "SAFECODEPOINTSINGLE",
		"SAFECODEPOINTDOUBLE", "WS",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 0, 24, 218, 6, -1, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2,
		4, 7, 4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2,
		10, 7, 10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15,
		7, 15, 2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19, 2, 20, 7,
		20, 2, 21, 7, 21, 2, 22, 7, 22, 2, 23, 7, 23, 2, 24, 7, 24, 2, 25, 7, 25,
		2, 26, 7, 26, 2, 27, 7, 27, 2, 28, 7, 28, 2, 29, 7, 29, 2, 30, 7, 30, 1,
		0, 1, 0, 1, 0, 1, 0, 1, 1, 1, 1, 1, 1, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1,
		2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 3, 2, 83, 8, 2, 1, 3, 1, 3, 1, 3,
		1, 3, 1, 4, 1, 4, 1, 4, 1, 5, 1, 5, 1, 5, 1, 6, 1, 6, 1, 6, 1, 6, 1, 7,
		1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 8, 1, 8, 1, 8, 1, 8, 1, 8, 1, 9,
		1, 9, 1, 9, 1, 9, 1, 9, 1, 9, 1, 9, 1, 10, 1, 10, 1, 11, 1, 11, 1, 11,
		1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 13, 1,
		13, 1, 13, 1, 13, 1, 13, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14,
		1, 14, 1, 14, 1, 15, 1, 15, 1, 16, 1, 16, 1, 17, 1, 17, 1, 18, 1, 18, 1,
		19, 1, 19, 1, 19, 5, 19, 157, 8, 19, 10, 19, 12, 19, 160, 9, 19, 1, 20,
		1, 20, 1, 21, 1, 21, 1, 22, 1, 22, 5, 22, 168, 8, 22, 10, 22, 12, 22, 171,
		9, 22, 1, 23, 1, 23, 1, 24, 1, 24, 1, 24, 5, 24, 178, 8, 24, 10, 24, 12,
		24, 181, 9, 24, 1, 24, 1, 24, 1, 24, 1, 24, 5, 24, 187, 8, 24, 10, 24,
		12, 24, 190, 9, 24, 1, 24, 3, 24, 193, 8, 24, 1, 25, 1, 25, 1, 25, 3, 25,
		198, 8, 25, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 27, 1, 27, 1,
		28, 1, 28, 1, 29, 1, 29, 1, 30, 4, 30, 213, 8, 30, 11, 30, 12, 30, 214,
		1, 30, 1, 30, 0, 0, 31, 1, 1, 3, 2, 5, 3, 7, 4, 9, 5, 11, 6, 13, 7, 15,
		8, 17, 9, 19, 10, 21, 11, 23, 12, 25, 13, 27, 14, 29, 15, 31, 16, 33, 17,
		35, 18, 37, 19, 39, 20, 41, 0, 43, 0, 45, 21, 47, 22, 49, 23, 51, 0, 53,
		0, 55, 0, 57, 0, 59, 0, 61, 24, 1, 0, 8, 1, 0, 48, 57, 3, 0, 65, 90, 95,
		95, 97, 122, 1, 0, 49, 57, 9, 0, 34, 34, 39, 39, 47, 47, 92, 92, 98, 98,
		102, 102, 110, 110, 114, 114, 116, 116, 3, 0, 48, 57, 65, 70, 97, 102,
		3, 0, 0, 31, 39, 39, 92, 92, 3, 0, 0, 31, 34, 34, 92, 92, 3, 0, 9, 10,
		13, 13, 32, 32, 225, 0, 1, 1, 0, 0, 0, 0, 3, 1, 0, 0, 0, 0, 5, 1, 0, 0,
		0, 0, 7, 1, 0, 0, 0, 0, 9, 1, 0, 0, 0, 0, 11, 1, 0, 0, 0, 0, 13, 1, 0,
		0, 0, 0, 15, 1, 0, 0, 0, 0, 17, 1, 0, 0, 0, 0, 19, 1, 0, 0, 0, 0, 21, 1,
		0, 0, 0, 0, 23, 1, 0, 0, 0, 0, 25, 1, 0, 0, 0, 0, 27, 1, 0, 0, 0, 0, 29,
		1, 0, 0, 0, 0, 31, 1, 0, 0, 0, 0, 33, 1, 0, 0, 0, 0, 35, 1, 0, 0, 0, 0,
		37, 1, 0, 0, 0, 0, 39, 1, 0, 0, 0, 0, 45, 1, 0, 0, 0, 0, 47, 1, 0, 0, 0,
		0, 49, 1, 0, 0, 0, 0, 61, 1, 0, 0, 0, 1, 63, 1, 0, 0, 0, 3, 67, 1, 0, 0,
		0, 5, 82, 1, 0, 0, 0, 7, 84, 1, 0, 0, 0, 9, 88, 1, 0, 0, 0, 11, 91, 1,
		0, 0, 0, 13, 94, 1, 0, 0, 0, 15, 98, 1, 0, 0, 0, 17, 105, 1, 0, 0, 0, 19,
		110, 1, 0, 0, 0, 21, 117, 1, 0, 0, 0, 23, 119, 1, 0, 0, 0, 25, 122, 1,
		0, 0, 0, 27, 131, 1, 0, 0, 0, 29, 136, 1, 0, 0, 0, 31, 145, 1, 0, 0, 0,
		33, 147, 1, 0, 0, 0, 35, 149, 1, 0, 0, 0, 37, 151, 1, 0, 0, 0, 39, 153,
		1, 0, 0, 0, 41, 161, 1, 0, 0, 0, 43, 163, 1, 0, 0, 0, 45, 165, 1, 0, 0,
		0, 47, 172, 1, 0, 0, 0, 49, 192, 1, 0, 0, 0, 51, 194, 1, 0, 0, 0, 53, 199,
		1, 0, 0, 0, 55, 205, 1, 0, 0, 0, 57, 207, 1, 0, 0, 0, 59, 209, 1, 0, 0,
		0, 61, 212, 1, 0, 0, 0, 63, 64, 5, 65, 0, 0, 64, 65, 5, 78, 0, 0, 65, 66,
		5, 68, 0, 0, 66, 2, 1, 0, 0, 0, 67, 68, 5, 79, 0, 0, 68, 69, 5, 82, 0,
		0, 69, 4, 1, 0, 0, 0, 70, 71, 5, 69, 0, 0, 71, 83, 5, 81, 0, 0, 72, 73,
		5, 78, 0, 0, 73, 83, 5, 69, 0, 0, 74, 75, 5, 71, 0, 0, 75, 83, 5, 69, 0,
		0, 76, 77, 5, 71, 0, 0, 77, 83, 5, 84, 0, 0, 78, 79, 5, 76, 0, 0, 79, 83,
		5, 84, 0, 0, 80, 81, 5, 76, 0, 0, 81, 83, 5, 69, 0, 0, 82, 70, 1, 0, 0,
		0, 82, 72, 1, 0, 0, 0, 82, 74, 1, 0, 0, 0, 82, 76, 1, 0, 0, 0, 82, 78,
		1, 0, 0, 0, 82, 80, 1, 0, 0, 0, 83, 6, 1, 0, 0, 0, 84, 85, 5, 82, 0, 0,
		85, 86, 5, 69, 0, 0, 86, 87, 5, 80, 0, 0, 87, 8, 1, 0, 0, 0, 88, 89, 5,
		73, 0, 0, 89, 90, 5, 78, 0, 0, 90, 10, 1, 0, 0, 0, 91, 92, 5, 65, 0, 0,
		92, 93, 5, 83, 0, 0, 93, 12, 1, 0, 0, 0, 94, 95, 5, 67, 0, 0, 95, 96, 5,
		66, 0, 0, 96, 97, 5, 70, 0, 0, 97, 14, 1, 0, 0, 0, 98, 99, 5, 83, 0, 0,
		99, 100, 5, 69, 0, 0, 100, 101, 5, 76, 0, 0, 101, 102, 5, 69, 0, 0, 102,
		103, 5, 67, 0, 0, 103, 104, 5, 84, 0, 0, 104, 16, 1, 0, 0, 0, 105, 106,
		5, 70, 0, 0, 106, 107, 5, 82, 0, 0, 107, 108, 5, 79, 0, 0, 108, 109, 5,
		77, 0, 0, 109, 18, 1, 0, 0, 0, 110, 111, 5, 70, 0, 0, 111, 112, 5, 73,
		0, 0, 112, 113, 5, 76, 0, 0, 113, 114, 5, 84, 0, 0, 114, 115, 5, 69, 0,
		0, 115, 116, 5, 82, 0, 0, 116, 20, 1, 0, 0, 0, 117, 118, 5, 42, 0, 0, 118,
		22, 1, 0, 0, 0, 119, 120, 5, 69, 0, 0, 120, 121, 5, 67, 0, 0, 121, 24,
		1, 0, 0, 0, 122, 123, 5, 77, 0, 0, 123, 124, 5, 73, 0, 0, 124, 125, 5,
		78, 0, 0, 125, 126, 5, 80, 0, 0, 126, 127, 5, 65, 0, 0, 127, 128, 5, 82,
		0, 0, 128, 129, 5, 84, 0, 0, 129, 130, 5, 83, 0, 0, 130, 26, 1, 0, 0, 0,
		131, 132, 5, 83, 0, 0, 132, 133, 5, 65, 0, 0, 133, 134, 5, 77, 0, 0, 134,
		135, 5, 69, 0, 0, 135, 28, 1, 0, 0, 0, 136, 137, 5, 68, 0, 0, 137, 138,
		5, 73, 0, 0, 138, 139, 5, 83, 0, 0, 139, 140, 5, 84, 0, 0, 140, 141, 5,
		73, 0, 0, 141, 142, 5, 78, 0, 0, 142, 143, 5, 67, 0, 0, 143, 144, 5, 84,
		0, 0, 144, 30, 1, 0, 0, 0, 145, 146, 5, 40, 0, 0, 146, 32, 1, 0, 0, 0,
		147, 148, 5, 41, 0, 0, 148, 34, 1, 0, 0, 0, 149, 150, 5, 64, 0, 0, 150,
		36, 1, 0, 0, 0, 151, 152, 5, 47, 0, 0, 152, 38, 1, 0, 0, 0, 153, 158, 3,
		43, 21, 0, 154, 157, 3, 41, 20, 0, 155, 157, 3, 43, 21, 0, 156, 154, 1,
		0, 0, 0, 156, 155, 1, 0, 0, 0, 157, 160, 1, 0, 0, 0, 158, 156, 1, 0, 0,
		0, 158, 159, 1, 0, 0, 0, 159, 40, 1, 0, 0, 0, 160, 158, 1, 0, 0, 0, 161,
		162, 7, 0, 0, 0, 162, 42, 1, 0, 0, 0, 163, 164, 7, 1, 0, 0, 164, 44, 1,
		0, 0, 0, 165, 169, 7, 2, 0, 0, 166, 168, 3, 41, 20, 0, 167, 166, 1, 0,
		0, 0, 168, 171, 1, 0, 0, 0, 169, 167, 1, 0, 0, 0, 169, 170, 1, 0, 0, 0,
		170, 46, 1, 0, 0, 0, 171, 169, 1, 0, 0, 0, 172, 173, 5, 48, 0, 0, 173,
		48, 1, 0, 0, 0, 174, 179, 5, 34, 0, 0, 175, 178, 3, 51, 25, 0, 176, 178,
		3, 59, 29, 0, 177, 175, 1, 0, 0, 0, 177, 176, 1, 0, 0, 0, 178, 181, 1,
		0, 0, 0, 179, 177, 1, 0, 0, 0, 179, 180, 1, 0, 0, 0, 180, 182, 1, 0, 0,
		0, 181, 179, 1, 0, 0, 0, 182, 193, 5, 34, 0, 0, 183, 188, 5, 39, 0, 0,
		184, 187, 3, 51, 25, 0, 185, 187, 3, 57, 28, 0, 186, 184, 1, 0, 0, 0, 186,
		185, 1, 0, 0, 0, 187, 190, 1, 0, 0, 0, 188, 186, 1, 0, 0, 0, 188, 189,
		1, 0, 0, 0, 189, 191, 1, 0, 0, 0, 190, 188, 1, 0, 0, 0, 191, 193, 5, 39,
		0, 0, 192, 174, 1, 0, 0, 0, 192, 183, 1, 0, 0, 0, 193, 50, 1, 0, 0, 0,
		194, 197, 5, 92, 0, 0, 195, 198, 7, 3, 0, 0, 196, 198, 3, 53, 26, 0, 197,
		195, 1, 0, 0, 0, 197, 196, 1, 0, 0, 0, 198, 52, 1, 0, 0, 0, 199, 200, 5,
		117, 0, 0, 200, 201, 3, 55, 27, 0, 201, 202, 3, 55, 27, 0, 202, 203, 3,
		55, 27, 0, 203, 204, 3, 55, 27, 0, 204, 54, 1, 0, 0, 0, 205, 206, 7, 4,
		0, 0, 206, 56, 1, 0, 0, 0, 207, 208, 8, 5, 0, 0, 208, 58, 1, 0, 0, 0, 209,
		210, 8, 6, 0, 0, 210, 60, 1, 0, 0, 0, 211, 213, 7, 7, 0, 0, 212, 211, 1,
		0, 0, 0, 213, 214, 1, 0, 0, 0, 214, 212, 1, 0, 0, 0, 214, 215, 1, 0, 0,
		0, 215, 216, 1, 0, 0, 0, 216, 217, 6, 30, 0, 0, 217, 62, 1, 0, 0, 0, 12,
		0, 82, 156, 158, 169, 177, 179, 186, 188, 192, 197, 214, 1, 6, 0, 0,
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

// QueryLexerInit initializes any static state used to implement QueryLexer. By default the
// static state used to implement the lexer is lazily initialized during the first call to
// NewQueryLexer(). You can call this function if you wish to initialize the static state ahead
// of time.
func QueryLexerInit() {
	staticData := &QueryLexerLexerStaticData
	staticData.once.Do(querylexerLexerInit)
}

// NewQueryLexer produces a new lexer instance for the optional input antlr.CharStream.
func NewQueryLexer(input antlr.CharStream) *QueryLexer {
	QueryLexerInit()
	l := new(QueryLexer)
	l.BaseLexer = antlr.NewBaseLexer(input)
	staticData := &QueryLexerLexerStaticData
	l.Interpreter = antlr.NewLexerATNSimulator(l, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	l.channelNames = staticData.ChannelNames
	l.modeNames = staticData.ModeNames
	l.RuleNames = staticData.RuleNames
	l.LiteralNames = staticData.LiteralNames
	l.SymbolicNames = staticData.SymbolicNames
	l.GrammarFileName = "QueryLexer.g4"
	// TODO: l.EOF = antlr.TokenEOF

	return l
}

// QueryLexer tokens.
const (
	QueryLexerAND_OP          = 1
	QueryLexerOR_OP           = 2
	QueryLexerSIMPLE_OP       = 3
	QueryLexerREP             = 4
	QueryLexerIN              = 5
	QueryLexerAS              = 6
	QueryLexerCBF             = 7
	QueryLexerSELECT          = 8
	QueryLexerFROM            = 9
	QueryLexerFILTER          = 10
	QueryLexerWILDCARD        = 11
	QueryLexerEC              = 12
	QueryLexerMINPARTS        = 13
	QueryLexerCLAUSE_SAME     = 14
	QueryLexerCLAUSE_DISTINCT = 15
	QueryLexerL_PAREN         = 16
	QueryLexerR_PAREN         = 17
	QueryLexerAT              = 18
	QueryLexerEC_SEP          = 19
	QueryLexerIDENT           = 20
	QueryLexerNUMBER1         = 21
	QueryLexerZERO            = 22
	QueryLexerSTRING          = 23
	QueryLexerWS              = 24
)
