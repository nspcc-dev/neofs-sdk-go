// Package policy provides facilities for creating policy from SQL-like language.
//   ANTLRv4 grammar is provided in `parser/Query.g4` and `parser/QueryLexer.g4`.
//
// Current limitations:
// 1. Filters must be defined before they are used.
//    This requirement may be relaxed in future.
// 2. Keywords are key-sensitive. This can be changed if necessary
//    https://github.com/antlr/antlr4/blob/master/doc/case-insensitive-lexing.md .
//
// Example query:
// REP 1 IN SPB
// REP 2 IN Americas
// CBF 4
// SELECT 1 IN City FROM SPBSSD AS SPB
// SELECT 2 IN SAME City FROM Americas AS Americas
// FILTER SSD EQ true AS IsSSD
// FILTER @IsSSD AND Country EQ "RU" AND City EQ "St.Petersburg" AS SPBSSD
// FILTER 'Continent' EQ 'North America' OR Continent EQ 'South America' AS Americas
package policy
