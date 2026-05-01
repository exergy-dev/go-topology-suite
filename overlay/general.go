package overlay

import (
	terra "github.com/terra-geo/terra"
	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/overlay/overlayng"
)

// tryOverlayNG runs the overlay-NG path on polygonal inputs (single
// polygon or multipolygon, supplied as polygon slices). Returns ok=true
// when the result is usable. With A1 + Item 8 fully landed, polygons
// with holes and multi-polygons are handled directly; the function
// only returns false on errors the caller should fall through.
func tryOverlayNG(subj, clip []*geom.Polygon, op overlayng.Op, c *crs.CRS) (geom.Geometry, bool) {
	first, rest, err := overlayng.OverlayPolygonal(subj, clip, op)
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

func requireSameCRS(a, b geom.Geometry) error {
	if !crs.Equal(a.CRS(), b.CRS()) {
		return terra.ErrCRSMismatch
	}
	return nil
}

// IntersectionGeneral returns subject ∩ clipper for arbitrary polygons
// or multipolygons. Falls back to the v0.1 Greiner-Hormann path on
// inputs the overlay-NG path can't handle (currently only single-polygon
// inputs go through GH; multi-polygon inputs always use overlay-NG).
func IntersectionGeneral(subject, clipper geom.Geometry) (geom.Geometry, error) {
	if err := requireSameCRS(subject, clipper); err != nil {
		return nil, err
	}
	if subject.IsEmpty() || clipper.IsEmpty() {
		return emptyOfDim(subject.CRS(), minDim(subject, clipper)), nil
	}
	if !isPolygonal(subject) || !isPolygonal(clipper) {
		return intersectionNonPolygonal(subject, clipper)
	}
	subj, clip, err := unwrapPolygonal(subject, clipper)
	if err != nil {
		return nil, err
	}
	if subj == nil || clip == nil {
		return emptyOfDim(subject.CRS(), minDim(subject, clipper)), nil
	}
	if g, ok := tryOverlayNG(subj, clip, overlayng.OpIntersection, subject.CRS()); ok {
		return g, nil
	}
	// Greiner-Hormann fallback only handles single-polygon inputs.
	if len(subj) != 1 || len(clip) != 1 {
		return nil, terra.ErrUnsupportedKernel
	}
	sp, cp := subj[0], clip[0]
	rings, hadIx := runGreinerHormann(outerRing(sp), outerRing(cp), string(opIntersection))
	if hadIx {
		return ringsToGeometry(subject.CRS(), rings)
	}
	switch {
	case ringContainsRing(outerRing(cp), outerRing(sp)):
		return geom.NewPolygon(subject.CRS(), outerRing(sp)), nil
	case ringContainsRing(outerRing(sp), outerRing(cp)):
		return geom.NewPolygon(subject.CRS(), outerRing(cp)), nil
	}
	return geom.NewEmptyPolygon(subject.CRS(), geom.LayoutXY), nil
}

// Union returns subject ∪ other for arbitrary polygons or multipolygons.
func Union(subject, other geom.Geometry) (geom.Geometry, error) {
	if err := requireSameCRS(subject, other); err != nil {
		return nil, err
	}
	if subject.IsEmpty() && other.IsEmpty() {
		return emptyOfDim(subject.CRS(), maxDim(subject, other)), nil
	}
	if subject.IsEmpty() {
		return other, nil
	}
	if other.IsEmpty() {
		return subject, nil
	}
	if !isPolygonal(subject) || !isPolygonal(other) {
		return unionNonPolygonal(subject, other)
	}
	subj, oth, err := unwrapPolygonal(subject, other)
	if err != nil {
		return nil, err
	}
	if subj == nil || oth == nil {
		// One side empty: result equals the other side.
		return nonEmptyOf(subj, oth, subject.CRS()), nil
	}
	if g, ok := tryOverlayNG(subj, oth, overlayng.OpUnion, subject.CRS()); ok {
		return g, nil
	}
	if len(subj) != 1 || len(oth) != 1 {
		return nil, terra.ErrUnsupportedKernel
	}
	sp, op := subj[0], oth[0]
	rings, hadIx := runGreinerHormann(outerRing(sp), outerRing(op), string(opUnion))
	if hadIx {
		return ringsToGeometry(subject.CRS(), rings)
	}
	switch {
	case ringContainsRing(outerRing(op), outerRing(sp)):
		return geom.NewPolygon(subject.CRS(), outerRing(op)), nil
	case ringContainsRing(outerRing(sp), outerRing(op)):
		return geom.NewPolygon(subject.CRS(), outerRing(sp)), nil
	}
	return geom.NewMultiPolygon(subject.CRS(),
		geom.NewPolygon(subject.CRS(), outerRing(sp)),
		geom.NewPolygon(subject.CRS(), outerRing(op)),
	), nil
}

// Difference returns subject \ other for arbitrary polygons or
// multipolygons.
func Difference(subject, other geom.Geometry) (geom.Geometry, error) {
	if err := requireSameCRS(subject, other); err != nil {
		return nil, err
	}
	if subject.IsEmpty() {
		return emptyOfDim(subject.CRS(), dimensionOf(subject)), nil
	}
	if other.IsEmpty() {
		return subject, nil
	}
	if !isPolygonal(subject) || !isPolygonal(other) {
		return differenceNonPolygonal(subject, other)
	}
	subj, oth, err := unwrapPolygonal(subject, other)
	if err != nil {
		return nil, err
	}
	if subj == nil {
		return emptyOfDim(subject.CRS(), dimensionOf(subject)), nil
	}
	if oth == nil {
		// Nothing to subtract.
		return polygonsToGeometry(subject.CRS(), subj), nil
	}
	if g, ok := tryOverlayNG(subj, oth, overlayng.OpDifference, subject.CRS()); ok {
		return g, nil
	}
	if len(subj) != 1 || len(oth) != 1 {
		return nil, terra.ErrUnsupportedKernel
	}
	sp, op := subj[0], oth[0]
	rings, hadIx := runGreinerHormann(outerRing(sp), outerRing(op), string(opDifference))
	if hadIx {
		return ringsToGeometry(subject.CRS(), rings)
	}
	switch {
	case ringContainsRing(outerRing(op), outerRing(sp)):
		return geom.NewEmptyPolygon(subject.CRS(), geom.LayoutXY), nil
	case ringContainsRing(outerRing(sp), outerRing(op)):
		return geom.NewPolygon(subject.CRS(), outerRing(sp), reverseRing(outerRing(op))), nil
	}
	return geom.NewPolygon(subject.CRS(), outerRing(sp)), nil
}

// SymmetricDifference returns (a \ b) ∪ (b \ a). For polygons without
// shared boundary this is the union of both differences.
func SymmetricDifference(a, b geom.Geometry) (geom.Geometry, error) {
	if err := requireSameCRS(a, b); err != nil {
		return nil, err
	}
	if a.IsEmpty() && b.IsEmpty() {
		return emptyOfDim(a.CRS(), maxDim(a, b)), nil
	}
	if a.IsEmpty() {
		return b, nil
	}
	if b.IsEmpty() {
		return a, nil
	}
	if !isPolygonal(a) || !isPolygonal(b) {
		return symDifferenceNonPolygonal(a, b)
	}
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

// dimensionOf returns the topological dimension of g (0=point, 1=line,
// 2=areal). For GeometryCollection the maximum member dimension is used,
// matching JTS overlay-NG empty-result-type rules.
func dimensionOf(g geom.Geometry) int {
	switch v := g.(type) {
	case *geom.Point, *geom.MultiPoint:
		return 0
	case *geom.LineString, *geom.MultiLineString:
		return 1
	case *geom.Polygon, *geom.MultiPolygon:
		return 2
	case *geom.GeometryCollection:
		max := 0
		for i := 0; i < v.NumGeometries(); i++ {
			if d := dimensionOf(v.GeometryAt(i)); d > max {
				max = d
			}
		}
		return max
	}
	return 0
}

func minDim(a, b geom.Geometry) int {
	da, db := dimensionOf(a), dimensionOf(b)
	if da < db {
		return da
	}
	return db
}

func maxDim(a, b geom.Geometry) int {
	da, db := dimensionOf(a), dimensionOf(b)
	if da > db {
		return da
	}
	return db
}

// emptyOfDim returns the canonical empty geometry of the given dimension.
// JTS overlay-NG returns `POINT EMPTY` / `LINESTRING EMPTY` / `POLYGON
// EMPTY` (not the multi variants) for empty overlay results.
func emptyOfDim(c *crs.CRS, dim int) geom.Geometry {
	switch dim {
	case 0:
		return geom.NewEmptyPoint(c, geom.LayoutXY)
	case 1:
		return geom.NewLineStringFlat(geom.LayoutXY, c, nil)
	default:
		return geom.NewEmptyPolygon(c, geom.LayoutXY)
	}
}

// unwrapPolygonal normalises operands to ([]*geom.Polygon, []*geom.Polygon)
// after CRS-equal checks. Empty inputs return nil slices (caller must
// handle). Both *geom.Polygon and *geom.MultiPolygon are accepted; any
// other geometry type returns ErrUnsupportedKernel.
func unwrapPolygonal(a, b geom.Geometry) ([]*geom.Polygon, []*geom.Polygon, error) {
	if err := requireSameCRS(a, b); err != nil {
		return nil, nil, err
	}
	pa, err := polygonsOf(a)
	if err != nil {
		return nil, nil, err
	}
	pb, err := polygonsOf(b)
	if err != nil {
		return nil, nil, err
	}
	return pa, pb, nil
}

// polygonsOf returns the constituent polygons of a Polygon or
// MultiPolygon input. Returns nil (no error) for an empty input.
func polygonsOf(g geom.Geometry) ([]*geom.Polygon, error) {
	if g.IsEmpty() {
		return nil, nil
	}
	switch v := g.(type) {
	case *geom.Polygon:
		return []*geom.Polygon{v}, nil
	case *geom.MultiPolygon:
		out := make([]*geom.Polygon, 0, v.NumGeometries())
		for i := 0; i < v.NumGeometries(); i++ {
			p := v.PolygonAt(i)
			if !p.IsEmpty() {
				out = append(out, p)
			}
		}
		return out, nil
	}
	return nil, terra.ErrUnsupportedKernel
}

// polygonsToGeometry returns a single polygon, multipolygon, or empty
// based on the slice contents. Used to box results from the overlay
// fallbacks where the operation effectively returns "the input
// unchanged" or "a subset of the input".
func polygonsToGeometry(c *crs.CRS, polys []*geom.Polygon) geom.Geometry {
	switch len(polys) {
	case 0:
		return geom.NewEmptyPolygon(c, geom.LayoutXY)
	case 1:
		return polys[0]
	}
	return geom.NewMultiPolygon(c, polys...)
}

// nonEmptyOf returns whichever of subj/oth is non-nil, packed as a
// geometry. Used for the union short-circuits when one side is empty.
func nonEmptyOf(subj, oth []*geom.Polygon, c *crs.CRS) geom.Geometry {
	if subj == nil && oth == nil {
		return geom.NewEmptyPolygon(c, geom.LayoutXY)
	}
	if subj == nil {
		return polygonsToGeometry(c, oth)
	}
	return polygonsToGeometry(c, subj)
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
