parser grammar Query;

options {
    tokenVocab = QueryLexer;
}

policy: repStmt+ cbfStmt? selectStmt* filterStmt* EOF;

repStmt:
    REP Count = NUMBER1     // number of object replicas
    (IN Selector = ident)?; // optional selector name

cbfStmt: CBF BackupFactor = NUMBER1; // container backup factor

selectStmt:
    SELECT Count = NUMBER1       // number of nodes to select without container backup factor *)
    (IN clause? Bucket = ident)? // bucket name
    FROM Filter = identWC        // filter reference or whole netmap
    (AS Name = ident)?           // optional selector name
    ;

clause: CLAUSE_SAME | CLAUSE_DISTINCT; // nodes from distinct buckets

filterExpr:
    F1 = filterExpr Op = AND_OP F2 = filterExpr
    | F1 = filterExpr Op = OR_OP F2 = filterExpr
    | '(' Inner = filterExpr ')'
    | expr
    ;

filterStmt:
    FILTER Expr = filterExpr
    AS Name = ident // obligatory filter name
    ;

expr:
    AT Filter = ident                               // reference to named filter
    | Key = filterKey SIMPLE_OP Value = filterValue // attribute comparison
    ;

filterKey : ident | STRING;
filterValue : ident | number | STRING;
number : ZERO | NUMBER1;
keyword : REP | IN | AS | SELECT | FROM | FILTER;
ident : keyword | IDENT;
identWC : ident | WILDCARD;
