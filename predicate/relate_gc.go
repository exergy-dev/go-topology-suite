package predicate

import (
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/kernel"
)

// unaryUnionFn is registered by the overlay package via SetUnaryUnion
// at init. relate uses it for GC operands so that the GC's areal
// members are unioned (mod-2 boundary collapse) before classifying
// against the other operand. Without this, per-member relate over a
// GC of adjacent or overlapping polygons reports shared edges as
// boundary instead of interior.
var unaryUnionFn func(g geom.Geometry) (geom.Geometry, error)

// SetUnaryUnion is wired by the overlay package's init().
func SetUnaryUnion(fn func(g geom.Geometry) (geom.Geometry, error)) {
	unaryUnionFn = fn
}

// dispatchGCPair routes relate(GC, *) — and relate(*, GC) — by
// simplifying the GC operand via UnaryUnion before falling back to
// computeMatrix. Returns (matrix, true) when a dispatch was made.
//
// We only act on *GeometryCollection operands; MultiPolygon members
// are by spec non-overlapping so the existing relateMulti path
// handles them correctly.
func dispatchGCPair(a, b geom.Geometry, k kernel.Kernel) (matrix, bool) {
	if unaryUnionFn == nil {
		return matrix{}, false
	}
	_, aIsGC := a.(*geom.GeometryCollection)
	_, bIsGC := b.(*geom.GeometryCollection)
	if !aIsGC && !bIsGC {
		return matrix{}, false
	}
	aSimplified := a
	bSimplified := b
	if aIsGC {
		s, err := unaryUnionFn(a)
		if err != nil || s == nil {
			return matrix{}, false
		}
		aSimplified = s
	}
	if bIsGC {
		s, err := unaryUnionFn(b)
		if err != nil || s == nil {
			return matrix{}, false
		}
		bSimplified = s
	}
	if !changedAfterUnaryUnion(a, aSimplified) && !changedAfterUnaryUnion(b, bSimplified) {
		// UnaryUnion didn't simplify anything — avoid an infinite loop
		// (computeMatrix would re-enter dispatchGCPair).
		return matrix{}, false
	}
	return computeMatrix(aSimplified, bSimplified, k), true
}

// changedAfterUnaryUnion reports whether the simplification actually
// reduced the GC structure (so we won't re-enter dispatchGCPair).
func changedAfterUnaryUnion(orig, simplified geom.Geometry) bool {
	if orig == simplified {
		return false
	}
	_, origIsGC := orig.(*geom.GeometryCollection)
	_, simpIsGC := simplified.(*geom.GeometryCollection)
	if origIsGC && !simpIsGC {
		return true
	}
	if origIsGC && simpIsGC {
		// Both still GCs — only meaningfully changed if simplification
		// reduced the member count.
		og := orig.(*geom.GeometryCollection)
		sg := simplified.(*geom.GeometryCollection)
		return sg.NumGeometries() < og.NumGeometries()
	}
	return false
}
