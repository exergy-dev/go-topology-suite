// Package wkt encodes and decodes geometries in OGC Well-Known Text format.
//
// The encoder produces canonical WKT (OGC SFS 1.2.1 + ISO 13249-3): types
// are uppercase, no leading whitespace, decimal points use '.', empty
// geometries are written as "TYPE EMPTY". XY/XYZ/XYM/XYZM layouts are
// distinguished by the explicit "Z", "M", or "ZM" keyword between the type
// and the coordinate list.
//
// The decoder accepts canonical WKT plus several common dialects: optional
// SRID prefix ("SRID=4326;POINT(1 2)") is parsed and attached as an EPSG
// CRS; whitespace and case are flexible.
package wkt
