// Code generated from java-escape by ANTLR 4.11.1. DO NOT EDIT.

package parser // Query

import "github.com/antlr/antlr4/runtime/Go/antlr/v4"

type BaseQueryVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseQueryVisitor) VisitPolicy(ctx *PolicyContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseQueryVisitor) VisitRepStmt(ctx *RepStmtContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseQueryVisitor) VisitCbfStmt(ctx *CbfStmtContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseQueryVisitor) VisitSelectStmt(ctx *SelectStmtContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseQueryVisitor) VisitClause(ctx *ClauseContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseQueryVisitor) VisitFilterExpr(ctx *FilterExprContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseQueryVisitor) VisitFilterStmt(ctx *FilterStmtContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseQueryVisitor) VisitExpr(ctx *ExprContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseQueryVisitor) VisitFilterKey(ctx *FilterKeyContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseQueryVisitor) VisitFilterValue(ctx *FilterValueContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseQueryVisitor) VisitNumber(ctx *NumberContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseQueryVisitor) VisitKeyword(ctx *KeywordContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseQueryVisitor) VisitIdent(ctx *IdentContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseQueryVisitor) VisitIdentWC(ctx *IdentWCContext) any {
	return v.VisitChildren(ctx)
}
