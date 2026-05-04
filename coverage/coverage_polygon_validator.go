// Port of org.locationtech.jts.coverage.CoveragePolygonValidator.
//
// Per-polygon coverage validation: checks that a single target
// polygon forms a valid coverage with a set of (ostensibly adjacent)
// neighbours. This complements CoverageValidator (which validates the
// entire set as a whole) for cases where:
//   - the caller wants per-target validation results without paying
//     the O(n²) cost of validating every pair in the whole coverage,
//   - or the caller is incrementally adding polygons to a coverage
//     and only needs to verify the new arrival.
//
// The set of validity rules is exactly the same as CoverageValidator's
// per-pair rules:
//
//	1. The target's interior must not intersect any neighbour's interior.
//	2. Where target boundaries meet neighbour boundaries, the
//	   intersection points must be vertex-aligned in both polygons.
//	3. (Optional) gaps narrower than gapWidth are flagged.
//
// Empty inputs and nil neighbours are skipped silently. Self-disjoint
// neighbours (those whose envelopes don't intersect the target) are
// cheaply filtered out.

package coverage

import (
	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/predicate"
)

// ValidationError reports a per-polygon coverage violation. The
// NeighborIndex is the index into the neighbours slice passed to
// ValidatePolygon; -1 indicates a self-only error (target's own
// boundary not vertex-clean).
type ValidationError struct {
	NeighborIndex int
	Kind          CoverageErrorKind
	Edge          [2]geom.XY
}

// ValidatePolygon checks that target forms a valid coverage cell with
// respect to the given neighbours. Returns nil when valid.
//
// gapWidth controls narrow-gap detection: pass 0 to disable.
//
// Port of org.locationtech.jts.coverage.CoveragePolygonValidator.
// JTS returns a Geometry of offending linework; this port returns a
// flat slice of structured errors, mirroring our existing
// CoverageError style.
func ValidatePolygon(target *geom.Polygon, neighbors []*geom.Polygon, gapWidth float64) []ValidationError {
	if target == nil || target.IsEmpty() {
		return nil
	}
	var errs []ValidationError

	tEnv := target.Envelope()
	probeEnv := tEnv
	if gapWidth > 0 {
		probeEnv.MinX -= gapWidth
		probeEnv.MinY -= gapWidth
		probeEnv.MaxX += gapWidth
		probeEnv.MaxY += gapWidth
	}

	for i, nb := range neighbors {
		if nb == nil || nb.IsEmpty() {
			continue
		}
		nbEnv := nb.Envelope()
		if !probeEnv.Intersects(nbEnv) {
			continue
		}
		// Rule 1: interior overlap.
		if overlaps, err := predicate.Overlaps(target, nb); err == nil && overlaps {
			errs = append(errs, ValidationError{
				NeighborIndex: i,
				Kind:          CoverageErrorOverlap,
			})
			continue
		}
		// Rule 2: any boundary segment of either polygon that has a
		// vertex of the other lying strictly mid-segment is a
		// non-vertex-aligned intersection.
		if mm := mismatchedEdges(target, nb); len(mm) > 0 {
			for _, e := range mm {
				errs = append(errs, ValidationError{
					NeighborIndex: i,
					Kind:          CoverageErrorMismatchedEdge,
					Edge:          e,
				})
			}
			continue
		}
		// Rule 3 (optional): narrow gap detection. If the polygons
		// are within gapWidth but share no exact edge, flag a gap.
		if gapWidth > 0 && !sharesAnyEdge(target, nb) {
			if dist := approxMinDistance(target, nb); dist > 0 && dist <= gapWidth {
				errs = append(errs, ValidationError{
					NeighborIndex: i,
					Kind:          CoverageErrorGap,
				})
			}
		}
	}
	return errs
}

// IsPolygonValid is a convenience wrapper around ValidatePolygon.
func IsPolygonValid(target *geom.Polygon, neighbors []*geom.Polygon, gapWidth float64) bool {
	return len(ValidatePolygon(target, neighbors, gapWidth)) == 0
}
