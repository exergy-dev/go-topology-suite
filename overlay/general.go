package overlay

import (
	terra "github.com/terra-geo/terra"
	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/overlay/overlayng"
)

// tryOverlayNG runs the overlay-NG path. Returns ok=true when the
// result is usable. With A1 fully landed, both shells and polygons-with-
// holes are handled directly; the function only returns false on errors
// the caller should fall through (e.g. the disjoint-via-graph signal).
func tryOverlayNG(subj, clip *geom.Polygon, op overlayng.Op, c *crs.CRS) (geom.Geometry, bool) {
	first, rest, err := overlayng.Overlay(subj, clip, op)
	if err != nil {
		return nil, false
	}
	if first.IsEmpty() && len(rest) == 0 {
		return geom.NewEmptyPolygon(c, geom.LayoutXY), true
	}
	if len(rest) == 0 {
		return first, true
	}
	all := append([]*geom.Polygon{first}, rest...)
	return geom.NewMultiPolygon(c, all...), true
}

// IntersectionGeneral returns subject ∩ clipper for arbitrary simple
// polygons. v1.0 behaviour:
//
//   - If both inputs are shells (no holes), tries the overlay-NG path
//     which handles coincident edges and shared boundaries correctly.
//   - Falls back to the v0.1 Greiner-Hormann path when overlay-NG can't
//     handle the inputs (e.g. polygons with holes).
func IntersectionGeneral(subject, clipper geom.Geometry) (geom.Geometry, error) {
	subj, clip, err := unwrapPolygons(subject, clipper)
	if err != nil {
		return nil, err
	}
	if g, ok := tryOverlayNG(subj, clip, overlayng.OpIntersection, subject.CRS()); ok {
		return g, nil
	}
	rings, hadIx := runGreinerHormann(outerRing(subj), outerRing(clip), "intersection")
	if hadIx {
		return ringsToGeometry(subject.CRS(), rings)
	}
	switch {
	case ringContainsRing(outerRing(clip), outerRing(subj)):
		return geom.NewPolygon(subject.CRS(), outerRing(subj)), nil
	case ringContainsRing(outerRing(subj), outerRing(clip)):
		return geom.NewPolygon(subject.CRS(), outerRing(clip)), nil
	}
	return geom.NewEmptyPolygon(subject.CRS(), geom.LayoutXY), nil
}

// Union returns subject ∪ other for arbitrary simple polygons.
func Union(subject, other geom.Geometry) (geom.Geometry, error) {
	subj, oth, err := unwrapPolygons(subject, other)
	if err != nil {
		return nil, err
	}
	if g, ok := tryOverlayNG(subj, oth, overlayng.OpUnion, subject.CRS()); ok {
		return g, nil
	}
	rings, hadIx := runGreinerHormann(outerRing(subj), outerRing(oth), "union")
	if hadIx {
		return ringsToGeometry(subject.CRS(), rings)
	}
	switch {
	case ringContainsRing(outerRing(oth), outerRing(subj)):
		return geom.NewPolygon(subject.CRS(), outerRing(oth)), nil
	case ringContainsRing(outerRing(subj), outerRing(oth)):
		return geom.NewPolygon(subject.CRS(), outerRing(subj)), nil
	}
	return geom.NewMultiPolygon(subject.CRS(),
		geom.NewPolygon(subject.CRS(), outerRing(subj)),
		geom.NewPolygon(subject.CRS(), outerRing(oth)),
	), nil
}

// Difference returns subject \ other for arbitrary simple polygons.
func Difference(subject, other geom.Geometry) (geom.Geometry, error) {
	subj, oth, err := unwrapPolygons(subject, other)
	if err != nil {
		return nil, err
	}
	if g, ok := tryOverlayNG(subj, oth, overlayng.OpDifference, subject.CRS()); ok {
		return g, nil
	}
	rings, hadIx := runGreinerHormann(outerRing(subj), outerRing(oth), "difference")
	if hadIx {
		return ringsToGeometry(subject.CRS(), rings)
	}
	switch {
	case ringContainsRing(outerRing(oth), outerRing(subj)):
		// Subject inside other → empty.
		return geom.NewEmptyPolygon(subject.CRS(), geom.LayoutXY), nil
	case ringContainsRing(outerRing(subj), outerRing(oth)):
		// Other inside subject → polygon with subject as outer, other as hole.
		return geom.NewPolygon(subject.CRS(), outerRing(subj), reverseRing(outerRing(oth))), nil
	}
	// Disjoint.
	return geom.NewPolygon(subject.CRS(), outerRing(subj)), nil
}

// SymmetricDifference returns (a \ b) ∪ (b \ a). For polygons without
// shared boundary this is the union of both differences.
func SymmetricDifference(a, b geom.Geometry) (geom.Geometry, error) {
	d1, err := Difference(a, b)
	if err != nil {
		return nil, err
	}
	d2, err := Difference(b, a)
	if err != nil {
		return nil, err
	}
	if d1.IsEmpty() {
		return d2, nil
	}
	if d2.IsEmpty() {
		return d1, nil
	}
	return collectAsMultiPolygon(a.CRS(), d1, d2), nil
}

// unwrapPolygons normalises operands to (*geom.Polygon, *geom.Polygon)
// after CRS-equal and non-empty checks.
func unwrapPolygons(a, b geom.Geometry) (*geom.Polygon, *geom.Polygon, error) {
	if !crs.Equal(a.CRS(), b.CRS()) {
		return nil, nil, terra.ErrCRSMismatch
	}
	if a.IsEmpty() || b.IsEmpty() {
		return nil, nil, nil
	}
	pa, ok := a.(*geom.Polygon)
	if !ok {
		return nil, nil, terra.ErrUnsupportedKernel
	}
	pb, ok := b.(*geom.Polygon)
	if !ok {
		return nil, nil, terra.ErrUnsupportedKernel
	}
	return pa, pb, nil
}

// ringsToGeometry converts a slice of result rings into a Polygon (one
// ring) or MultiPolygon (multiple disjoint rings). v0.1 does not detect
// holes inside the result; every output ring is treated as outer.
func ringsToGeometry(c *crs.CRS, rings [][]geom.XY) (geom.Geometry, error) {
	switch len(rings) {
	case 0:
		return geom.NewEmptyPolygon(c, geom.LayoutXY), nil
	case 1:
		return geom.NewPolygon(c, rings[0]), nil
	default:
		polys := make([]*geom.Polygon, 0, len(rings))
		for _, r := range rings {
			polys = append(polys, geom.NewPolygon(c, r))
		}
		return geom.NewMultiPolygon(c, polys...), nil
	}
}

// ringContainsRing reports whether outer fully contains inner (every
// vertex of inner lies inside outer's ring). Used in no-intersection
// fallbacks; assumes outer is simple and non-self-intersecting.
func ringContainsRing(outer, inner []geom.XY) bool {
	for _, p := range inner {
		if !pointInRingXY(p, outer) {
			return false
		}
	}
	return true
}

func reverseRing(r []geom.XY) []geom.XY {
	out := make([]geom.XY, len(r))
	for i := range r {
		out[i] = r[len(r)-1-i]
	}
	return out
}

func collectAsMultiPolygon(c *crs.CRS, geoms ...geom.Geometry) geom.Geometry {
	var polys []*geom.Polygon
	for _, g := range geoms {
		switch v := g.(type) {
		case *geom.Polygon:
			if !v.IsEmpty() {
				polys = append(polys, v)
			}
		case *geom.MultiPolygon:
			for i := 0; i < v.NumGeometries(); i++ {
				polys = append(polys, v.PolygonAt(i))
			}
		}
	}
	if len(polys) == 0 {
		return geom.NewEmptyPolygon(c, geom.LayoutXY)
	}
	if len(polys) == 1 {
		return polys[0]
	}
	return geom.NewMultiPolygon(c, polys...)
}
