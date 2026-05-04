package snaprounding

import (
	"errors"
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/internal/noding"
)

// ss is a shorthand for building a SegmentString from raw XY pairs.
func ss(tag int, pts ...geom.XY) *noding.SegmentString {
	return &noding.SegmentString{Coords: pts, Tag: tag}
}

func xy(x, y float64) geom.XY { return geom.XY{X: x, Y: y} }

// vertexCount returns the number of (string, index) positions whose
// vertex equals p across the result set. Useful for asserting that an
// intersection is shared by both crossing segments.
func vertexCount(strs []*noding.SegmentString, p geom.XY) int {
	n := 0
	for _, s := range strs {
		for _, v := range s.Coords {
			if v == p {
				n++
			}
		}
	}
	return n
}

// TestNodeSimpleX verifies a clean X-intersection at a grid-aligned
// crossing produces a shared vertex on both segments.
func TestNodeSimpleX(t *testing.T) {
	tol := 1.0
	n := &Noder{Tolerance: tol}
	out, stats, err := n.Node([]*noding.SegmentString{
		ss(1, xy(0, 0), xy(10, 10)),
		ss(2, xy(0, 10), xy(10, 0)),
	})
	if err != nil {
		t.Fatalf("Node returned error: %v", err)
	}
	if !stats.Converged {
		t.Errorf("expected Converged=true, got %+v", stats)
	}
	mid := xy(5, 5)
	if vertexCount(out, mid) < 2 {
		t.Errorf("expected (5,5) shared by both segments; got %d occurrences\n  out=%v", vertexCount(out, mid), dump(out))
	}
}

// TestNodeNearMissSnapsTogether verifies two nearly-touching segments
// snap into a shared crossing under a coarse-enough grid.
func TestNodeNearMissSnapsTogether(t *testing.T) {
	// Two segments that are 0.4 apart at their nearest point — well
	// within a tolerance=1 grid. After snap rounding they should share
	// a hot pixel and be cross-noded.
	tol := 1.0
	n := &Noder{Tolerance: tol}
	out, stats, err := n.Node([]*noding.SegmentString{
		ss(1, xy(0, 0), xy(10, 0)),
		ss(2, xy(5, 0.4), xy(5, 10)),
	})
	if err != nil {
		t.Fatalf("Node returned error: %v", err)
	}
	if !stats.Converged {
		t.Errorf("expected convergence, got %+v", stats)
	}
	// After snap, the second segment's endpoint (5, 0.4) snaps to
	// (5, 0); the first segment passes through (5, 0). The hot pixel
	// at (5, 0) must therefore be a vertex of both strings.
	pix := xy(5, 0)
	if vertexCount(out, pix) < 2 {
		t.Errorf("expected hot pixel (5,0) shared; out=%v", dump(out))
	}
}

// TestNodeT verifies a T-intersection (one segment ending on another's
// interior) produces a shared vertex without splitting the dead-end
// segment.
func TestNodeT(t *testing.T) {
	tol := 1.0
	n := &Noder{Tolerance: tol}
	out, _, err := n.Node([]*noding.SegmentString{
		ss(1, xy(0, 0), xy(10, 0)),
		ss(2, xy(5, 0), xy(5, 5)),
	})
	if err != nil {
		t.Fatalf("Node returned error: %v", err)
	}
	pix := xy(5, 0)
	// The crossing point must appear in the horizontal segment as an
	// internal vertex (i.e. the original [0..10] string is split there).
	splitFound := false
	for _, s := range out {
		if s.Tag == 1 {
			for i, v := range s.Coords {
				if v == pix && i > 0 && i < len(s.Coords)-1 {
					splitFound = true
				}
			}
			// Multiple sub-strings also indicate a split.
		}
	}
	// Either an interior split or a multi-piece output for tag 1 satisfies T-noding.
	pieces := 0
	for _, s := range out {
		if s.Tag == 1 {
			pieces++
		}
	}
	if !splitFound && pieces < 2 {
		t.Errorf("expected horizontal segment to be split at (5,0); out=%v", dump(out))
	}
}

// TestNodeIdempotent verifies running the noder a second time on its
// own output produces an identical result (zero further splits).
func TestNodeIdempotent(t *testing.T) {
	tol := 1.0
	n := &Noder{Tolerance: tol}
	first, _, err := n.Node([]*noding.SegmentString{
		ss(1, xy(0, 0), xy(10, 10)),
		ss(2, xy(0, 10), xy(10, 0)),
	})
	if err != nil {
		t.Fatalf("first Node: %v", err)
	}
	second, stats, err := n.Node(first)
	if err != nil {
		t.Fatalf("second Node: %v", err)
	}
	if stats.Splits != 0 {
		t.Errorf("expected 0 splits on idempotent re-noding; got %d (stats=%+v)", stats.Splits, stats)
	}
	if !stats.Converged {
		t.Errorf("expected converged on second pass; got %+v", stats)
	}
	if len(second) != len(first) {
		t.Errorf("expected stable string count: first=%d second=%d", len(first), len(second))
	}
}

// TestNodeNoTolerance verifies the API rejects Tolerance <= 0.
func TestNodeNoTolerance(t *testing.T) {
	n := &Noder{Tolerance: 0}
	_, _, err := n.Node([]*noding.SegmentString{ss(1, xy(0, 0), xy(1, 1))})
	if err == nil {
		t.Fatal("expected error for Tolerance=0")
	}
}

// TestNodeEmpty returns an empty result with Converged=true.
func TestNodeEmpty(t *testing.T) {
	n := &Noder{Tolerance: 1.0}
	out, stats, err := n.Node(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 0 || !stats.Converged {
		t.Errorf("expected empty result and converged; got out=%d stats=%+v", len(out), stats)
	}
}

// TestNodePreservesTags verifies output strings carry their input Tag.
func TestNodePreservesTags(t *testing.T) {
	n := &Noder{Tolerance: 1.0}
	out, _, err := n.Node([]*noding.SegmentString{
		ss(7, xy(0, 0), xy(10, 0)),
		ss(11, xy(5, -5), xy(5, 5)),
	})
	if err != nil {
		t.Fatalf("Node: %v", err)
	}
	saw7, saw11 := false, false
	for _, s := range out {
		switch s.Tag {
		case 7:
			saw7 = true
		case 11:
			saw11 = true
		default:
			t.Errorf("unexpected tag in output: %d", s.Tag)
		}
	}
	if !saw7 || !saw11 {
		t.Errorf("missing tag in output: saw7=%v saw11=%v", saw7, saw11)
	}
}

// TestNodeSliverPrecision verifies a sliver-precision input (an
// intersection landing 0.05 from a vertex on a segment with tolerance 1)
// is correctly resolved as a shared vertex.
func TestNodeSliverPrecision(t *testing.T) {
	n := &Noder{Tolerance: 1.0}
	out, stats, err := n.Node([]*noding.SegmentString{
		ss(1, xy(0, 0), xy(10, 0)),
		ss(2, xy(5.05, -3), xy(5.05, 3)),
	})
	if err != nil {
		t.Fatalf("Node: %v", err)
	}
	if !stats.Converged {
		t.Errorf("expected convergence on sliver case; got %+v", stats)
	}
	// The vertical segment's (5.05, *) endpoints round to (5, *), so
	// after snap the crossing is at (5, 0) — shared by both strings.
	if vertexCount(out, xy(5, 0)) < 2 {
		t.Errorf("expected (5,0) shared after snap; out=%v", dump(out))
	}
}

// TestNodeMaxIterRespected verifies that MaxIter > 0 is honoured.
func TestNodeMaxIterRespected(t *testing.T) {
	n := &Noder{Tolerance: 1.0, MaxIter: 1}
	_, stats, err := n.Node([]*noding.SegmentString{
		ss(1, xy(0, 0), xy(10, 10)),
		ss(2, xy(0, 10), xy(10, 0)),
	})
	// Simple X converges in one iteration so this should still succeed.
	if err != nil && !errors.Is(err, ErrNotConverged) {
		t.Fatalf("Node: %v", err)
	}
	if stats.Iterations > 1 {
		t.Errorf("expected ≤1 iteration with MaxIter=1; got %d", stats.Iterations)
	}
}

// dump returns a compact representation of a noded result for test
// failure messages.
func dump(strs []*noding.SegmentString) string {
	out := "["
	for i, s := range strs {
		if i > 0 {
			out += ", "
		}
		out += "tag" + itoa(s.Tag) + ":"
		for j, v := range s.Coords {
			if j > 0 {
				out += "->"
			}
			out += "(" + ftoa(v.X) + "," + ftoa(v.Y) + ")"
		}
	}
	return out + "]"
}

// TestNodeSeedIntersections runs the simple X-cross with the JTS-style
// pre-noding intersection seed enabled and asserts the same shared
// vertex appears on both crossings, confirming the seeded hot-pixel
// path produces the same topology as the bare fix-point loop.
func TestNodeSeedIntersections(t *testing.T) {
	tol := 1.0
	n := &Noder{Tolerance: tol, SeedIntersections: true}
	out, stats, err := n.Node([]*noding.SegmentString{
		ss(1, xy(0, 0), xy(10, 10)),
		ss(2, xy(0, 10), xy(10, 0)),
	})
	if err != nil {
		t.Fatalf("Node returned error: %v", err)
	}
	if !stats.Converged {
		t.Errorf("expected Converged=true, got %+v", stats)
	}
	mid := xy(5, 5)
	if vertexCount(out, mid) < 2 {
		t.Errorf("expected (5,5) shared by both segments, got %d occurrences", vertexCount(out, mid))
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func ftoa(f float64) string {
	// Tests use integer or near-integer values; truncate to int for
	// readability without pulling in fmt.
	i := int(f)
	if float64(i) == f {
		return itoa(i)
	}
	// Fallback for non-integer: print with one decimal.
	whole := int(f)
	frac := int((f - float64(whole)) * 100)
	if frac < 0 {
		frac = -frac
	}
	return itoa(whole) + "." + itoa(frac)
}
