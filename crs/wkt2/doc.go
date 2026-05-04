// Package wkt2 implements a v0.1 parser for OGC Well-Known Text 2
// (ISO 19162:2019) coordinate reference system definitions.
//
// The parser extracts only CRS identity: the top-level kind
// (Geographic / Projected / Unknown) and the outermost ID["EPSG", code]
// clause, if any. It deliberately does not build a structural CRS model:
// go-topology-suite has no projection engine in v0.1, so the parser exists solely to
// fill in *crs.CRS values for ad-hoc CRSes that the EPSG registry does
// not cover.
//
// Both `[` and `(` are accepted as opening brackets, matching the WKT2
// permissive grammar. Keywords are matched case-insensitively. The lexer
// tracks byte offsets so that errors carry diagnosable positions.
package wkt2
