// Package policy provides facilities for creating policy from SQL-like language.
//   ANTLRv4 grammar is provided in `parser/Query.g4` and `parser/QueryLexer.g4`.
//
// Current limitations:
// 1. Filters must be defined before they are used.
// 	This requirement may be relaxed in future.
//
// Example query:
// REP 1 in SPB
// REP 2 in Americas
// CBF 4
// SELECT 1 Node IN City FROM SPBSSD AS SPB
// SELECT 2 Node IN SAME City FROM Americas AS Americas
// FILTER SSD EQ true AS IsSSD
// FILTER @IsSSD AND Country eq "RU" AND City eq "St.Petersburg" AS SPBSSD
// FILTER 'Continent' == 'North America' OR Continent == 'South America' AS Americas
package policy
