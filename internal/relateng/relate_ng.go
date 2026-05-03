package relateng

import (
	"github.com/terra-geo/terra/geom"
)

// RelateNG is the driver that orchestrates a topology computation
// over a TopologyComputer. Port of
// org.locationtech.jts.operation.relateng.RelateNG.
//
// This Go port covers the point-locator path: P/P, P/L, P/A, L/A
// vertex/line-end interactions and area-vertex-on-target interactions.
// The edge-segment crossing pipeline (computeAtEdges in JTS) is not
// yet ported; callers in predicate/relateng.go must use the
// EdgesIntersected hint to decide whether to fall back to the legacy
// path.
type RelateNG struct {
	rule  BoundaryNodeRule
	geomA *Geometry
}

// NewRelateNG constructs a driver for geometry a with the given rule.
// rule may be nil, in which case the OGC SFS rule is used.
func NewRelateNG(a geom.Geometry, rule BoundaryNodeRule) *RelateNG {
	if rule == nil {
		rule = OGCSFSBoundaryRule
	}
	return &RelateNG{
		rule:  rule,
		geomA: NewGeometryRule(a, false, rule),
	}
}

// Evaluate computes the topological relationship A vs b against the
// supplied predicate. Returns the predicate's final boolean value.
func (r *RelateNG) Evaluate(b geom.Geometry, p TopologyPredicate) bool {
	geomB := NewGeometryRule(b, false, r.rule)

	// Envelope short-circuits.
	if !r.hasRequiredEnvelopeInteraction(geomB, p) {
		return false
	}

	dimA := r.geomA.DimensionReal()
	dimB := geomB.DimensionReal()

	p.InitDim(dimA, dimB)
	if p.IsKnown() {
		p.Finish()
		return p.Value()
	}
	p.InitEnv(r.geomA.Envelope(), geomB.Envelope())
	if p.IsKnown() {
		p.Finish()
		return p.Value()
	}

	tc := NewTopologyComputer(p, r.geomA, geomB)

	// Optimised P/P path.
	if dimA == DimP && dimB == DimP {
		r.computePP(geomB, tc)
		tc.Finish()
		return tc.Result()
	}

	// Test points against the (potentially indexed) target first.
	r.computeAtPoints(geomB, false, r.geomA, tc)
	if tc.IsResultKnown() {
		return tc.Result()
	}
	r.computeAtPoints(r.geomA, true, geomB, tc)
	if tc.IsResultKnown() {
		return tc.Result()
	}

	// Edge-segment intersection pass is not yet ported. The remaining
	// JTS step (computeAtEdges + EvaluateNodes) is documented as
	// future work in topology_computer.go.

	tc.Finish()
	return tc.Result()
}

// EvaluateMatrix runs the predicate to completion and returns the
// computed DE-9IM matrix.
func (r *RelateNG) EvaluateMatrix(b geom.Geometry) *IntersectionMatrix {
	pred := NewRelateMatrixPredicate()
	r.Evaluate(b, pred)
	return pred.Matrix()
}

// HasEdgeIntersection is a hint used by predicate/relateng.go to
// decide whether to trust this driver's result. Returns true when the
// inputs may have edge intersections that the current Go port can't
// detect (any case with two non-empty edge-bearing inputs whose
// envelopes interact).
func (r *RelateNG) HasEdgeIntersection(b geom.Geometry) bool {
	geomB := NewGeometryRule(b, false, r.rule)
	if !r.geomA.HasEdges() || !geomB.HasEdges() {
		return false
	}
	return r.geomA.Envelope().Intersects(geomB.Envelope())
}

func (r *RelateNG) hasRequiredEnvelopeInteraction(geomB *Geometry, p TopologyPredicate) bool {
	envA := r.geomA.Envelope()
	envB := geomB.Envelope()
	interacts := false
	if p.RequireCovers(true) {
		if !envelopeCovers(envA, envB) {
			return false
		}
		interacts = true
	} else if p.RequireCovers(false) {
		if !envelopeCovers(envB, envA) {
			return false
		}
		interacts = true
	}
	if !interacts && p.RequireInteraction() && !envA.Intersects(envB) {
		return false
	}
	return true
}

func envelopeCovers(outer, inner geom.Envelope) bool {
	if inner.IsEmpty() {
		return true
	}
	if outer.IsEmpty() {
		return false
	}
	return outer.MinX <= inner.MinX && outer.MaxX >= inner.MaxX &&
		outer.MinY <= inner.MinY && outer.MaxY >= inner.MaxY
}

func (r *RelateNG) computePP(geomB *Geometry, tc *TopologyComputer) {
	ptsA := uniquePoints(r.geomA.Geometry())
	ptsB := uniquePoints(geomB.Geometry())
	numBinA := 0
	for ptB := range ptsB {
		if _, ok := ptsA[ptB]; ok {
			numBinA++
			tc.AddPointOnPointInterior(ptB)
		} else {
			tc.AddPointOnPointExterior(false, ptB)
		}
		if tc.IsResultKnown() {
			return
		}
	}
	if numBinA < len(ptsA) {
		tc.AddPointOnPointExterior(true, geom.XY{})
	}
}

func (r *RelateNG) computeAtPoints(g *Geometry, isA bool, target *Geometry, tc *TopologyComputer) {
	if r.computePoints(g, isA, target, tc) {
		return
	}
	checkDisjoint := target.HasDimension(DimA) || tc.IsExteriorCheckRequired(isA)
	if !checkDisjoint {
		return
	}
	if r.computeLineEnds(g, isA, target, tc) {
		return
	}
	r.computeAreaVertex(g, isA, target, tc)
}

func (r *RelateNG) computePoints(g *Geometry, isA bool, target *Geometry, tc *TopologyComputer) bool {
	if !g.HasDimension(DimP) {
		return false
	}
	for _, pt := range effectivePoints(g.Geometry()) {
		locDimTarget := target.LocateWithDim(pt)
		locTarget := Location(locDimTarget)
		dimTarget := DimensionExt(locDimTarget, tc.GetDimension(!isA))
		tc.AddPointOnGeometry(isA, locTarget, dimTarget, pt)
		if tc.IsResultKnown() {
			return true
		}
	}
	return false
}

func (r *RelateNG) computeLineEnds(g *Geometry, isA bool, target *Geometry, tc *TopologyComputer) bool {
	if !g.HasDimension(DimL) {
		return false
	}
	hasExt := false
	walkLineStrings(g.Geometry(), func(line *geom.LineString) bool {
		if hasExt && envelopeDisjoint(line.Envelope(), target.Envelope()) {
			return true
		}
		e0 := line.PointAt(0)
		ext, stop := r.computeLineEnd(g, isA, e0, target, tc)
		hasExt = hasExt || ext
		if stop {
			return false
		}
		if !lineIsClosed(line) {
			e1 := line.PointAt(line.NumPoints() - 1)
			ext, stop = r.computeLineEnd(g, isA, e1, target, tc)
			hasExt = hasExt || ext
			if stop {
				return false
			}
		}
		return true
	})
	return tc.IsResultKnown()
}

func (r *RelateNG) computeLineEnd(g *Geometry, isA bool, pt geom.XY, target *Geometry, tc *TopologyComputer) (bool, bool) {
	locDimLineEnd := g.LocateLineEndWithDim(pt)
	dimLineEnd := DimensionExt(locDimLineEnd, tc.GetDimension(isA))
	if dimLineEnd != DimL {
		return false, false
	}
	locLineEnd := Location(locDimLineEnd)
	locDimTarget := target.LocateWithDim(pt)
	locTarget := Location(locDimTarget)
	dimTarget := DimensionExt(locDimTarget, tc.GetDimension(!isA))
	tc.AddLineEndOnGeometry(isA, locLineEnd, locTarget, dimTarget, pt)
	return locTarget == LocExterior, tc.IsResultKnown()
}

func (r *RelateNG) computeAreaVertex(g *Geometry, isA bool, target *Geometry, tc *TopologyComputer) bool {
	if !g.HasDimension(DimA) {
		return false
	}
	if target.Dimension() < DimL {
		return false
	}
	hasExt := false
	walkPolygons(g.Geometry(), func(poly *geom.Polygon) bool {
		if hasExt && envelopeDisjoint(poly.Envelope(), target.Envelope()) {
			return true
		}
		ringPts := append([][]geom.XY{poly.ExteriorRing()}, poly.InteriorRings()...)
		for _, ring := range ringPts {
			if len(ring) == 0 {
				continue
			}
			pt := ring[0]
			locArea := g.LocateAreaVertex(pt)
			locDimTarget := target.LocateWithDim(pt)
			locTarget := Location(locDimTarget)
			dimTarget := DimensionExt(locDimTarget, tc.GetDimension(!isA))
			tc.AddAreaVertex(isA, locArea, locTarget, dimTarget, pt)
			if locTarget == LocExterior {
				hasExt = true
			}
			if tc.IsResultKnown() {
				return false
			}
		}
		return true
	})
	return tc.IsResultKnown()
}

// uniquePoints extracts the unique XY coordinates of all Point /
// MultiPoint members in g.
func uniquePoints(g geom.Geometry) map[geom.XY]struct{} {
	out := make(map[geom.XY]struct{})
	collectUniquePoints(g, out)
	return out
}

func collectUniquePoints(g geom.Geometry, out map[geom.XY]struct{}) {
	if g == nil || g.IsEmpty() {
		return
	}
	switch v := g.(type) {
	case *geom.Point:
		out[v.XY()] = struct{}{}
	case *geom.MultiPoint:
		for i := 0; i < v.NumGeometries(); i++ {
			out[v.PointAt(i)] = struct{}{}
		}
	case *geom.GeometryCollection:
		for i := 0; i < v.NumGeometries(); i++ {
			collectUniquePoints(v.GeometryAt(i), out)
		}
	}
}

// effectivePoints returns the XY coords of every Point/MultiPoint
// member, deduplicated, in arbitrary order. Mirrors JTS
// RelateGeometry.getEffectivePoints (zero-length lines included
// elsewhere via DimensionReal logic).
func effectivePoints(g geom.Geometry) []geom.XY {
	pts := uniquePoints(g)
	out := make([]geom.XY, 0, len(pts))
	for p := range pts {
		out = append(out, p)
	}
	return out
}

// walkLineStrings invokes fn on every non-empty LineString member.
// Returning false stops the walk early.
func walkLineStrings(g geom.Geometry, fn func(*geom.LineString) bool) {
	if g == nil || g.IsEmpty() {
		return
	}
	switch v := g.(type) {
	case *geom.LineString:
		if !v.IsEmpty() {
			fn(v)
		}
	case *geom.LinearRing:
		if !v.IsEmpty() {
			fn(v.AsLineString())
		}
	case *geom.MultiLineString:
		for i := 0; i < v.NumGeometries(); i++ {
			ls := v.LineStringAt(i)
			if !ls.IsEmpty() && !fn(ls) {
				return
			}
		}
	case *geom.GeometryCollection:
		for i := 0; i < v.NumGeometries(); i++ {
			walkLineStrings(v.GeometryAt(i), fn)
		}
	}
}

// walkPolygons invokes fn on every non-empty Polygon member.
func walkPolygons(g geom.Geometry, fn func(*geom.Polygon) bool) {
	if g == nil || g.IsEmpty() {
		return
	}
	switch v := g.(type) {
	case *geom.Polygon:
		if !v.IsEmpty() {
			fn(v)
		}
	case *geom.MultiPolygon:
		for i := 0; i < v.NumGeometries(); i++ {
			poly := v.PolygonAt(i)
			if !poly.IsEmpty() && !fn(poly) {
				return
			}
		}
	case *geom.GeometryCollection:
		for i := 0; i < v.NumGeometries(); i++ {
			walkPolygons(v.GeometryAt(i), fn)
		}
	}
}

func lineIsClosed(ls *geom.LineString) bool {
	n := ls.NumPoints()
	if n < 2 {
		return false
	}
	return ls.PointAt(0) == ls.PointAt(n-1)
}

func envelopeDisjoint(a, b geom.Envelope) bool {
	if a.IsEmpty() || b.IsEmpty() {
		return true
	}
	return !a.Intersects(b)
}
