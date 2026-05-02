// Package snaprounding implements a Goodrich-Guibas-style snap-rounding
// noder. Given a set of input segment strings and a precision tolerance,
// it produces a topologically consistent noded output: every segment-
// segment intersection (after rounding to the precision grid) is a
// shared vertex, and no segment passes through a hot pixel without
// having that pixel's centre as one of its vertices.
//
// The implementation iterates a noding/hot-pixel-insertion fixpoint
// until no segment requires further splitting at a hot pixel. Each
// iteration may add new hot pixels (created by re-noding produces fresh
// intersection points), so convergence is monotone — the hot-pixel set
// only grows. MaxIter is a safety belt for pathological inputs; if the
// fixpoint has not converged by then, Node returns the best-effort
// result with Stats.Converged == false and ErrNotConverged.
//
// This package is internal to terra: callers are overlay/overlayng (for
// SR overlays) and buffer (for offset-curve cleanup). It reuses
// internal/snap.HotPixelSet for the R-tree-backed pixel index and
// internal/noding for the underlying segment intersection primitive.
package snaprounding

import (
	"errors"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/internal/noding"
	"github.com/terra-geo/terra/internal/snap"
)

// Noder is a snap-rounding noder. The zero value is invalid (Tolerance
// must be set); call with a positive precision-grid spacing.
type Noder struct {
	// Tolerance is the precision-grid cell side. Must be > 0; a value
	// of 0 (or negative) is rejected by Node — callers that want plain
	// noding without snap rounding should call internal/noding directly.
	Tolerance float64

	// MaxIter caps the noding/insertion fixpoint iterations. Defaults
	// to 5 when zero. Set higher only when convergence stalls on
	// pathological inputs (very fine grids relative to coordinate
	// magnitude).
	MaxIter int

	// MergeNearCollinear opts in to a post-noding pass that merges
	// pairs of segments lying within tolerance/2 perpendicular
	// distance of each other onto a shared hot pixel chain. The pass
	// is opt-in because it can shift result areas at the tolerance
	// level — fine for overlay-NG (where snap rounding is the
	// expected outcome) but disruptive for buffer's offset-curve
	// polygonization, which relies on the noder preserving the
	// offset-curve geometry intact at tolerance much smaller than
	// the buffer distance.
	//
	// Currently the standard fixpoint loop already merges
	// near-collinear segments via hot-pixel insertion when their
	// vertices fall within the same grid cell; setting this flag
	// enables a stricter perpendicular-distance test that catches
	// segment pairs whose shared collinearity emerges only after
	// rounding. The flag is reserved for future expansion: today it
	// is recognised but does not change the noder output. Callers
	// that need stricter behaviour should still set the flag so the
	// upgrade ships transparently.
	MergeNearCollinear bool
}

// Stats reports per-Node telemetry. Useful both for tests (asserting
// convergence) and for surfacing diagnostic detail to JTS test logs.
type Stats struct {
	// Iterations is the number of fixpoint passes actually run.
	Iterations int
	// HotPixels is the size of the hot-pixel set at the final iteration.
	HotPixels int
	// Splits is the cumulative count of hot-pixel insertions made
	// across all iterations.
	Splits int
	// Converged is true iff the noder reached a fixpoint within MaxIter
	// iterations (or one more for the final guard pass).
	Converged bool
}

// ErrNotConverged is returned when the snap-rounding fixpoint has not
// stabilised within MaxIter+1 iterations. Callers should treat the
// returned segments as best-effort and decide whether to fall back to a
// non-snap-rounded result.
var ErrNotConverged = errors.New("snaprounding: fixpoint did not converge")

const defaultMaxIter = 5

// Node runs the snap-rounding fixpoint on input and returns the noded
// segment strings. Tags are preserved through every transformation.
//
// The returned slice is freshly allocated; the input slice and its
// SegmentString contents are not mutated.
func (n *Noder) Node(input []*noding.SegmentString) ([]*noding.SegmentString, Stats, error) {
	if n.Tolerance <= 0 {
		return nil, Stats{}, errors.New("snaprounding: Tolerance must be > 0")
	}
	if len(input) == 0 {
		return nil, Stats{Converged: true}, nil
	}
	maxIter := n.MaxIter
	if maxIter <= 0 {
		maxIter = defaultMaxIter
	}

	rd := snap.New(n.Tolerance)
	// First noding pass: realise every segment-segment intersection
	// before any rounding happens. The output is freshly allocated so
	// subsequent in-place vertex snapping is safe.
	noded := adaptiveNode(input)

	stats := Stats{}
	for iter := 0; iter < maxIter; iter++ {
		stats.Iterations++

		snapAndDedupe(noded, rd)

		hp := buildHotPixelSet(noded, n.Tolerance)
		stats.HotPixels = hp.Len()

		next, inserted := insertHotPixelSplits(noded, hp)
		stats.Splits += inserted

		if inserted == 0 {
			stats.Converged = true
			return finalise(noded), stats, nil
		}

		// Re-node so cross-segment intersections at the new vertices
		// are realised as shared endpoints.
		noded = adaptiveNode(next)
	}

	// Guard pass: snap, build pixel set, see if anything still wants to
	// split. If not, we converged on the boundary; if so, surface the
	// non-convergence to the caller.
	snapAndDedupe(noded, rd)
	hp := buildHotPixelSet(noded, n.Tolerance)
	stats.HotPixels = hp.Len()
	_, finalIns := insertHotPixelSplits(noded, hp)
	if finalIns == 0 {
		stats.Converged = true
		return finalise(noded), stats, nil
	}
	return finalise(noded), stats, ErrNotConverged
}

// snapAndDedupe rounds every vertex of every string to the grid and
// drops consecutive-duplicate vertices that result. Operates in place
// because the strings were freshly allocated by adaptiveNode.
func snapAndDedupe(strs []*noding.SegmentString, rd *snap.Rounder) {
	for _, s := range strs {
		for i, v := range s.Coords {
			s.Coords[i] = rd.SnapVertex(v)
		}
		s.Coords = dedupeConsecutive(s.Coords)
	}
}

// buildHotPixelSet returns a hot-pixel set populated with every vertex
// of every string. Inputs must already be grid-snapped.
func buildHotPixelSet(strs []*noding.SegmentString, tolerance float64) *snap.HotPixelSet {
	hp := snap.NewHotPixelSet(tolerance)
	for _, s := range strs {
		for _, v := range s.Coords {
			hp.Add(v)
		}
	}
	return hp
}

// insertHotPixelSplits walks every string and inserts hot-pixel centres
// into segments whose interior passes through a hot pixel. Returns the
// updated string set and the cumulative insertion count. Strings with
// fewer than two vertices pass through unchanged.
func insertHotPixelSplits(strs []*noding.SegmentString, hp *snap.HotPixelSet) ([]*noding.SegmentString, int) {
	out := make([]*noding.SegmentString, 0, len(strs))
	totalIns := 0
	for _, s := range strs {
		if len(s.Coords) < 2 {
			out = append(out, s)
			continue
		}
		newCoords, ins := insertSplitsInto(s.Coords, hp)
		totalIns += ins
		out = append(out, &noding.SegmentString{Coords: newCoords, Tag: s.Tag})
	}
	return out, totalIns
}

// insertSplitsInto inserts hot-pixel centres between consecutive
// vertices of pts when a segment's interior passes through a hot pixel.
// Returns the new vertex chain and the number of insertions performed.
//
// Insertions are skipped when the candidate centre coincides with an
// existing endpoint of the segment under consideration, OR when it
// duplicates the previous emitted vertex (a hot pixel already
// represented as the start of the next segment).
func insertSplitsInto(pts []geom.XY, hp *snap.HotPixelSet) ([]geom.XY, int) {
	if len(pts) < 2 {
		return pts, 0
	}
	out := make([]geom.XY, 0, len(pts))
	out = append(out, pts[0])
	inserted := 0
	for i := 0; i+1 < len(pts); i++ {
		a, b := pts[i], pts[i+1]
		splits := hp.SegmentSplitsAt(a, b)
		for _, sp := range splits {
			if sp == a || sp == b {
				continue
			}
			if n := len(out); n > 0 && out[n-1] == sp {
				continue
			}
			out = append(out, sp)
			inserted++
		}
		if n := len(out); n == 0 || out[n-1] != b {
			out = append(out, b)
		}
	}
	return out, inserted
}

// finalise drops any string whose Coords collapsed below two vertices
// (which happens when an entire string was absorbed into a single grid
// cell). The remainder is returned as-is.
func finalise(strs []*noding.SegmentString) []*noding.SegmentString {
	out := make([]*noding.SegmentString, 0, len(strs))
	for _, s := range strs {
		if len(s.Coords) >= 2 {
			out = append(out, s)
		}
	}
	return out
}

// dedupeConsecutive removes runs of equal consecutive vertices from pts.
// Returns a slice that aliases the input — safe because callers
// immediately reassign s.Coords to the result.
func dedupeConsecutive(pts []geom.XY) []geom.XY {
	if len(pts) <= 1 {
		return pts
	}
	out := pts[:1]
	for i := 1; i < len(pts); i++ {
		if pts[i] != pts[i-1] {
			out = append(out, pts[i])
		}
	}
	return out
}

// adaptiveNode picks SimpleNoder for small inputs and IndexedNoder once
// the segment count crosses the empirical threshold (matched to
// overlay/overlayng's tuning).
func adaptiveNode(strs []*noding.SegmentString) []*noding.SegmentString {
	const indexThreshold = 64
	total := 0
	for _, s := range strs {
		total += s.NumSegments()
	}
	if total < indexThreshold {
		return noding.SimpleNoder{}.Node(strs)
	}
	return noding.IndexedNoder{}.Node(strs)
}
