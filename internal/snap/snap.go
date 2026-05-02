package snap

import (
	"math"

	"github.com/terra-geo/terra/geom"
)

// Rounder snaps vertices and segments to a regular grid whose cell side
// equals the configured tolerance.
//
// A Rounder is immutable and safe for concurrent use after construction.
type Rounder struct {
	tolerance float64
	// invTolerance caches 1/tolerance to turn the per-coordinate divide
	// in SnapVertex into a multiply. Snap rounding is on the hot path of
	// every overlay call, so the savings matter at scale.
	invTolerance float64
}

// New returns a Rounder snapping to a grid with the given side length.
// The tolerance must be a positive finite number; otherwise New panics.
//
// A reasonable default for unit-scale Cartesian inputs is 1e-9; for lon/lat
// inputs ~1e-7 (≈ 1 cm at the equator) is conservative.
func New(tolerance float64) *Rounder {
	if !(tolerance > 0) || math.IsInf(tolerance, 0) || math.IsNaN(tolerance) {
		panic("snap: tolerance must be a positive finite number")
	}
	return &Rounder{
		tolerance:    tolerance,
		invTolerance: 1.0 / tolerance,
	}
}

// Tolerance returns the grid cell side length.
func (r *Rounder) Tolerance() float64 { return r.tolerance }

// SnapVertex rounds a single vertex to the nearest grid point. Non-finite
// coordinates (NaN or ±Inf) are returned unchanged — snap rounding has no
// meaningful answer for them, and silently mapping NaN to 0 would mask the
// caller's upstream bug.
func (r *Rounder) SnapVertex(v geom.XY) geom.XY {
	return geom.XY{
		X: snapScalar(v.X, r.invTolerance, r.tolerance),
		Y: snapScalar(v.Y, r.invTolerance, r.tolerance),
	}
}

// snapScalar rounds x to the nearest integer multiple of tolerance.
//
// We use math.Round (round-half-away-from-zero) rather than the IEEE-754
// banker's rounding that float64→int conversion would give us, because
// banker's rounding makes the snap result depend on the parity of the grid
// cell index — which is a surprising behaviour for users specifying e.g.
// tolerance = 1e-3 and expecting "decimal rounding to three places".
func snapScalar(x, invTol, tol float64) float64 {
	if !isFinite(x) {
		return x
	}
	r := math.Round(x*invTol) * tol
	// Normalise IEEE-754 negative zero to positive zero so two vertices
	// at the same grid cell have identical bit patterns. Without this,
	// e.g. -0.2 rounds to -0 while 0.2 rounds to +0; downstream code
	// keyed on math.Float64bits sees them as distinct vertices.
	if r == 0 {
		return 0
	}
	return r
}

func isFinite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }

// SnapRing rounds every vertex of a ring and removes consecutive duplicates
// introduced by the snap. The closing vertex is preserved (the result is
// returned as a closed ring whose first and last vertices coincide), unless
// the ring collapses.
//
// Returns nil if the snapped ring has fewer than 4 distinct vertices
// (i.e., fewer than 3 distinct points plus a closing vertex) — that means
// the ring collapsed to a point, segment, or empty geometry under the
// tolerance.
//
// The input is not mutated.
func (r *Rounder) SnapRing(ring []geom.XY) []geom.XY {
	if len(ring) == 0 {
		return nil
	}
	out := make([]geom.XY, 0, len(ring))
	for _, v := range ring {
		s := r.SnapVertex(v)
		// Drop consecutive duplicates. We compare exactly because both
		// values are already snapped to the integer grid.
		if n := len(out); n > 0 && out[n-1].Equal(s) {
			continue
		}
		out = append(out, s)
	}
	// Ensure the ring is closed. After snapping it is possible that the
	// original closing vertex was dropped as a consecutive duplicate of
	// the penultimate vertex; conversely an open input becomes closed
	// here. Either way: if first != last, append first.
	if len(out) > 0 && !out[0].Equal(out[len(out)-1]) {
		out = append(out, out[0])
	}
	// A valid closed ring has at least 4 vertices (3 distinct + 1 closing).
	if len(out) < 4 {
		return nil
	}
	return out
}

// SnapPolygon snaps every ring of p. Returns nil if p is nil, empty, or if
// the outer ring collapses under the tolerance. Holes that collapse are
// silently dropped — a collapsed hole has zero area and contributes
// nothing topologically.
//
// The CRS of the input is preserved.
func (r *Rounder) SnapPolygon(p *geom.Polygon) *geom.Polygon {
	if p == nil || p.IsEmpty() {
		return nil
	}
	outer := r.SnapRing(p.ExteriorRing())
	if outer == nil {
		return nil
	}
	rings := [][]geom.XY{outer}
	for _, hole := range p.InteriorRings() {
		if snapped := r.SnapRing(hole); snapped != nil {
			rings = append(rings, snapped)
		}
	}
	return geom.NewPolygon(p.CRS(), rings...)
}
