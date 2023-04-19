// Code generated from java-escape by ANTLR 4.11.1. DO NOT EDIT.

package parser // Query

import "github.com/antlr/antlr4/runtime/Go/antlr/v4"

// A complete Visitor for a parse tree produced by Query.
type QueryVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by Query#policy.
	VisitPolicy(ctx *PolicyContext) any

	// Visit a parse tree produced by Query#repStmt.
	VisitRepStmt(ctx *RepStmtContext) any

	// Visit a parse tree produced by Query#cbfStmt.
	VisitCbfStmt(ctx *CbfStmtContext) any

	// Visit a parse tree produced by Query#selectStmt.
	VisitSelectStmt(ctx *SelectStmtContext) any

	// Visit a parse tree produced by Query#clause.
	VisitClause(ctx *ClauseContext) any

	// Visit a parse tree produced by Query#filterExpr.
	VisitFilterExpr(ctx *FilterExprContext) any

	// Visit a parse tree produced by Query#filterStmt.
	VisitFilterStmt(ctx *FilterStmtContext) any

	// Visit a parse tree produced by Query#expr.
	VisitExpr(ctx *ExprContext) any

	// Visit a parse tree produced by Query#filterKey.
	VisitFilterKey(ctx *FilterKeyContext) any

	// Visit a parse tree produced by Query#filterValue.
	VisitFilterValue(ctx *FilterValueContext) any

	// Visit a parse tree produced by Query#number.
	VisitNumber(ctx *NumberContext) any

	// Visit a parse tree produced by Query#keyword.
	VisitKeyword(ctx *KeywordContext) any

	// Visit a parse tree produced by Query#ident.
	VisitIdent(ctx *IdentContext) any

	// Visit a parse tree produced by Query#identWC.
	VisitIdentWC(ctx *IdentWCContext) any
}
