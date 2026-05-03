package triangulate

import (
	"github.com/terra-geo/terra/geom"
)

// TriangulatePolygon computes a triangulation of a single polygon
// (with optional holes) using the ear-clipping algorithm.
//
// The triangulation is non-overlapping, covers the whole polygon, and
// uses only the polygon's existing vertices — no Steiner points are
// introduced. Triangle quality is not optimised; for a higher-quality
// (Delaunay) triangulation, use the constrained Delaunay path. Every
// returned triangle is fully contained within the polygon, so unlike a
// raw Delaunay triangulation of the vertex set, this respects holes
// and concave boundaries.
//
// Holes are joined to the shell by the simplified
// PolygonHoleJoiner (see polygon_hole_joiner.go), producing a single
// self-touching ring that the ear-clipper can process. The choice of
// bridge segments is heuristic; pathological inputs may produce
// unevenly-shaped triangles but the result will still tile the polygon.
//
// Returns nil for empty or invalid polygons.
//
// Port of org.locationtech.jts.triangulate.polygon.PolygonTriangulator.
func TriangulatePolygon(p *geom.Polygon) []Triangle {
	if p == nil || p.IsEmpty() {
		return nil
	}
	shell := joinPolygonHoles(p)
	if len(shell) < 4 {
		return nil
	}
	return earClipTriangulate(shell)
}

// TriangulatePolygons triangulates every polygon in the given geometry
// (handling MultiPolygon and GeometryCollection inputs by recursion)
// and returns the concatenated triangle list. Non-polygonal components
// are ignored.
func TriangulatePolygons(g geom.Geometry) []Triangle {
	if g == nil || g.IsEmpty() {
		return nil
	}
	switch v := g.(type) {
	case *geom.Polygon:
		return TriangulatePolygon(v)
	case *geom.MultiPolygon:
		var out []Triangle
		for i := 0; i < v.NumGeometries(); i++ {
			out = append(out, TriangulatePolygon(v.PolygonAt(i))...)
		}
		return out
	case *geom.GeometryCollection:
		var out []Triangle
		for i := 0; i < v.NumGeometries(); i++ {
			out = append(out, TriangulatePolygons(v.GeometryAt(i))...)
		}
		return out
	}
	return nil
}
