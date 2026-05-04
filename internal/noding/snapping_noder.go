package noding

import (
	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/index"
)

// SnappingNoder implements JTS's vertex-snap-and-node strategy
// (org.locationtech.jts.noding.snap.SnappingNoder). It is a robust
// alternative to snap-rounding for cases where the hot-pixel grid
// distorts geometry too aggressively: instead of rounding to a fixed
// grid, it snaps coordinates to each other within a configurable
// distance using a KdTree-backed point index. Coordinates that are
// within snapTolerance of an already-seen coordinate snap to that
// existing point; intersection points likewise snap to existing
// vertices and to each other.
//
// Behaviour relative to JTS:
//
//   - Vertex snapping uses the new index.KdTree's tolerance-based
//     dedup (which mirrors JTS SnappingPointIndex).
//   - Intra-segment intersections are computed by the underlying
//     MCIndexNoder run with OverlapTolerance = 2*snapTolerance, then
//     snapped through the same KdTree, matching JTS's
//     SnappingIntersectionAdder behaviour.
//
// The snap tolerance should be small relative to coordinate magnitude
// (JTS recommends 1e-12 of the input scale). With an appropriate
// tolerance the algorithm is very robust against near-coincident
// vertices and diagonal pairs.
//
// Note this is *not* snap-rounding (no fixed grid; no precision
// model). Use snaprounding.Noder for that strategy.
type SnappingNoder struct {
	SnapTolerance float64
}

// Node returns a noded copy of input. All vertices and intersection
// points in the output are snapped through a single KdTree.
func (n SnappingNoder) Node(input []*SegmentString) []*SegmentString {
	if len(input) == 0 {
		return nil
	}
	// snapTolerance must be > 0 for meaningful snap; on 0 fall back
	// to the underlying MCIndexNoder. (JTS does not guard against
	// zero, but its KdTree treats 0 as "exact dedup", which is fine.)
	snapIdx := index.NewKdTree[struct{}](n.SnapTolerance)

	// Phase 1: snap every vertex through the KdTree. Each Insert
	// returns the canonical node coordinate (which may be the snap
	// target rather than the input point).
	snappedInputs := make([]*SegmentString, len(input))
	for i, ss := range input {
		coords := make([]geom.XY, 0, len(ss.Coords))
		var prev geom.XY
		havePrev := false
		for _, p := range ss.Coords {
			node, _ := snapIdx.Insert(p, struct{}{})
			canon := node.Coordinate
			// CoordinateList-style: skip immediate duplicates so the
			// snap doesn't synthesise zero-length segments. (JTS
			// uses CoordinateList.add(pt, false).)
			if havePrev && canon == prev {
				continue
			}
			coords = append(coords, canon)
			prev = canon
			havePrev = true
		}
		snappedInputs[i] = &SegmentString{Coords: coords, Tag: ss.Tag}
	}

	// Phase 2: run the MCIndexNoder with overlap tolerance to surface
	// near-coincident segment pairs that snap-induced geometry could
	// otherwise miss. JTS uses 2*snapTolerance.
	noder := MCIndexNoder{OverlapTolerance: 2 * n.SnapTolerance}
	intermediate := noder.Node(snappedInputs)

	// Phase 3: snap every output coordinate (including intersection
	// points the noder synthesised) back through the same KdTree, so
	// near-duplicate intersection points coalesce.
	out := make([]*SegmentString, 0, len(intermediate))
	for _, ss := range intermediate {
		coords := make([]geom.XY, 0, len(ss.Coords))
		var prev geom.XY
		havePrev := false
		for _, p := range ss.Coords {
			node, _ := snapIdx.Insert(p, struct{}{})
			canon := node.Coordinate
			if havePrev && canon == prev {
				continue
			}
			coords = append(coords, canon)
			prev = canon
			havePrev = true
		}
		// Drop pieces that collapsed to a single point (or empty)
		// after re-snap — they no longer carry geometry.
		if len(coords) < 2 {
			continue
		}
		out = append(out, &SegmentString{Coords: coords, Tag: ss.Tag})
	}
	return out
}
