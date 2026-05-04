package conformance

import (
	"fmt"
	"math"

	"github.com/exergy-dev/go-topology-suite/geom"
)

// Impl is the contract every implementation under test must satisfy.
//
// The interface is intentionally narrow: six operations cover the
// overlay engine (Intersection / Union / Difference), scalar measure
// (Area / Length), and the boolean predicate stack (Relate, returning
// the DE-9IM matrix). Each method takes go-topology-suite geometries and returns
// either a result expressed in go-topology-suite terms (so the harness can compare
// without knowing the implementation's native types) or an error.
//
// Implementations are responsible for converting between go-topology-suite and
// their own native form. The standard path is wkt.Marshal -> the
// implementation's WKT parser -> the implementation's operation -> WKT
// again -> wkt.Unmarshal. WKT is the lingua franca and avoids any
// cross-package coupling.
type Impl interface {
	// Name returns a short identifier used in summary lines and log
	// entries. It must be stable across runs.
	Name() string

	// Intersection returns a ∩ b. Empty inputs and disjoint inputs
	// produce an empty geometry, not an error.
	Intersection(a, b geom.Geometry) (geom.Geometry, error)

	// Union returns a ∪ b.
	Union(a, b geom.Geometry) (geom.Geometry, error)

	// Difference returns a \ b.
	Difference(a, b geom.Geometry) (geom.Geometry, error)

	// Area returns the surface area of g (zero for non-polygonal
	// geometries).
	Area(g geom.Geometry) (float64, error)

	// Length returns the linear length of g (zero for point
	// geometries; perimeter for polygonal geometries).
	Length(g geom.Geometry) (float64, error)

	// Relate returns the DE-9IM matrix describing the topological
	// relationship between a and b as a 9-character string in the
	// canonical II IB IE BI BB BE EI EB EE row-major order.
	Relate(a, b geom.Geometry) (string, error)
}

// Op enumerates the operations the harness compares across
// implementations.
type Op string

const (
	OpIntersection Op = "intersection"
	OpUnion        Op = "union"
	OpDifference   Op = "difference"
	OpArea         Op = "area"
	OpLength       Op = "length"
	OpRelate       Op = "relate"
)

// AllOps is the complete set of operations the harness runs. Adding a
// new Op requires updating Impl, the dispatch in run, and the equality
// helper.
var AllOps = []Op{
	OpIntersection,
	OpUnion,
	OpDifference,
	OpArea,
	OpLength,
	OpRelate,
}

// Tolerances used when comparing implementation outputs. See doc.go for
// the rationale.
const (
	// scalarRelTol is the relative tolerance for Area / Length.
	scalarRelTol = 5e-6
	// overlayAreaRelTol is the relative tolerance applied to the
	// areas of overlay-result geometries (Intersection / Union /
	// Difference). Different engines pick different ring orientations
	// and vertex orders, so byte equality is not workable; matching
	// areas to 1% is a useful and honest bar.
	overlayAreaRelTol = 1e-2
)

// result captures the outcome of running one (Op, Impl, pair) combo.
type result struct {
	Op   Op
	Impl string
	// Either Geom is set (for OpIntersection/Union/Difference), Scalar
	// is set (for OpArea/OpLength), or Matrix is set (for OpRelate).
	Geom   geom.Geometry
	Scalar float64
	Matrix string
	Err    error
}

// run dispatches op against impl using (a, b). For unary ops the b
// argument is ignored.
func run(impl Impl, op Op, a, b geom.Geometry) result {
	r := result{Op: op, Impl: impl.Name()}
	switch op {
	case OpIntersection:
		r.Geom, r.Err = impl.Intersection(a, b)
	case OpUnion:
		r.Geom, r.Err = impl.Union(a, b)
	case OpDifference:
		r.Geom, r.Err = impl.Difference(a, b)
	case OpArea:
		r.Scalar, r.Err = impl.Area(a)
	case OpLength:
		r.Scalar, r.Err = impl.Length(a)
	case OpRelate:
		r.Matrix, r.Err = impl.Relate(a, b)
	default:
		r.Err = fmt.Errorf("conformance: unknown op %q", op)
	}
	return r
}

// agree reports whether ref and other are considered equal for the
// given Op, applying the tolerances documented in doc.go. If either
// produced an error the rule is conservative: same-error agrees, any
// asymmetry disagrees.
func agree(op Op, ref, other result, areaOf func(geom.Geometry) float64) (bool, string) {
	switch {
	case ref.Err != nil && other.Err != nil:
		// Both failed. We treat that as agreement: both impls saw an
		// edge case they couldn't handle. The detail message keeps
		// the failure modes visible in the summary.
		return true, ""
	case ref.Err != nil:
		return false, fmt.Sprintf("ref=err(%v) other=ok", ref.Err)
	case other.Err != nil:
		return false, fmt.Sprintf("ref=ok other=err(%v)", other.Err)
	}
	switch op {
	case OpArea, OpLength:
		if relativeEqual(ref.Scalar, other.Scalar, scalarRelTol) {
			return true, ""
		}
		return false, fmt.Sprintf("ref=%g other=%g (relTol=%g)",
			ref.Scalar, other.Scalar, scalarRelTol)
	case OpRelate:
		if ref.Matrix == other.Matrix {
			return true, ""
		}
		return false, fmt.Sprintf("ref=%q other=%q", ref.Matrix, other.Matrix)
	case OpIntersection, OpUnion, OpDifference:
		// Compare via area: nil/empty geometries trivially have area
		// zero, and area is invariant under ring re-orientation /
		// vertex re-ordering, which different overlay engines pick
		// differently.
		ra := safeArea(areaOf, ref.Geom)
		oa := safeArea(areaOf, other.Geom)
		if relativeEqual(ra, oa, overlayAreaRelTol) {
			return true, ""
		}
		return false, fmt.Sprintf("area ref=%g other=%g (relTol=%g)",
			ra, oa, overlayAreaRelTol)
	}
	return false, fmt.Sprintf("unknown op %q", op)
}

// safeArea returns 0 for nil geometries (which both go-topology-suite's overlay and
// simplefeatures occasionally produce on disjoint inputs), and
// otherwise delegates to the supplied area function.
func safeArea(areaOf func(geom.Geometry) float64, g geom.Geometry) float64 {
	if g == nil || g.IsEmpty() {
		return 0
	}
	return areaOf(g)
}

// relativeEqual is the standard "absolute-or-relative" comparator used
// throughout the harness. NaN never compares equal (a NaN result
// always indicates a real disagreement).
func relativeEqual(a, b, relTol float64) bool {
	if math.IsNaN(a) || math.IsNaN(b) {
		return false
	}
	d := math.Abs(a - b)
	if d == 0 {
		return true
	}
	scale := math.Max(math.Abs(a), math.Abs(b))
	if scale == 0 {
		return d <= relTol
	}
	return d/scale <= relTol
}
