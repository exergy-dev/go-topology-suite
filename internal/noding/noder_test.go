package noding

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terra-geo/terra/geom"
)

func xy(x, y float64) geom.XY { return geom.XY{X: x, Y: y} }

// approxEqual returns true if two XYs match within a small tolerance —
// SegmentIntersection can leave sub-ULP round-off at the intersection
// point, so exact equality is the wrong test.
func approxEqual(a, b geom.XY) bool {
	const eps = 1e-9
	return math.Abs(a.X-b.X) < eps && math.Abs(a.Y-b.Y) < eps
}

// stringHas returns true if ss has the given vertex sequence (in order),
// up to floating-point round-off.
func stringHas(ss *SegmentString, want ...geom.XY) bool {
	if len(ss.Coords) != len(want) {
		return false
	}
	for i := range want {
		if !approxEqual(ss.Coords[i], want[i]) {
			return false
		}
	}
	return true
}

func TestSimpleNoder_TwoCrossingSegments(t *testing.T) {
	// Horizontal (0,0)-(2,0) crosses vertical (1,-1)-(1,1) at (1,0).
	ssA := &SegmentString{Coords: []geom.XY{xy(0, 0), xy(2, 0)}, Tag: 1}
	ssB := &SegmentString{Coords: []geom.XY{xy(1, -1), xy(1, 1)}, Tag: 2}

	out := SimpleNoder{}.Node([]*SegmentString{ssA, ssB})
	require.Equal(t, 4, len(out), "expected 4 output strings")

	// Check by tag the two halves of each input.
	tagged := map[int][]*SegmentString{}
	for _, s := range out {
		tagged[s.Tag] = append(tagged[s.Tag], s)
	}
	require.Equal(t, 2, len(tagged[1]), "expected 2 strings per tag, got %v", tagged)
	require.Equal(t, 2, len(tagged[2]), "expected 2 strings per tag, got %v", tagged)

	// Tag 1 should be (0,0)-(1,0) and (1,0)-(2,0).
	wantA := [][]geom.XY{
		{xy(0, 0), xy(1, 0)},
		{xy(1, 0), xy(2, 0)},
	}
	for i, s := range tagged[1] {
		assert.True(t, stringHas(s, wantA[i]...), "tag-1 piece %d: got %v want %v", i, s.Coords, wantA[i])
	}
	wantB := [][]geom.XY{
		{xy(1, -1), xy(1, 0)},
		{xy(1, 0), xy(1, 1)},
	}
	for i, s := range tagged[2] {
		assert.True(t, stringHas(s, wantB[i]...), "tag-2 piece %d: got %v want %v", i, s.Coords, wantB[i])
	}
}

func TestSimpleNoder_ParallelNonOverlapping(t *testing.T) {
	ssA := &SegmentString{Coords: []geom.XY{xy(0, 0), xy(2, 0)}, Tag: 1}
	ssB := &SegmentString{Coords: []geom.XY{xy(0, 1), xy(2, 1)}, Tag: 2}

	out := SimpleNoder{}.Node([]*SegmentString{ssA, ssB})
	require.Equal(t, 2, len(out), "expected 2 output strings")
	for _, s := range out {
		assert.Equal(t, 2, len(s.Coords), "expected unchanged 2-vertex string, got %v", s.Coords)
	}
}

func TestSimpleNoder_SharedEndpoint(t *testing.T) {
	// Two segments meeting at (1,0): no interior intersection, no split.
	ssA := &SegmentString{Coords: []geom.XY{xy(0, 0), xy(1, 0)}, Tag: 1}
	ssB := &SegmentString{Coords: []geom.XY{xy(1, 0), xy(2, 1)}, Tag: 2}

	out := SimpleNoder{}.Node([]*SegmentString{ssA, ssB})
	require.Equal(t, 2, len(out), "expected 2 output strings")
	for _, s := range out {
		assert.Equal(t, 2, len(s.Coords), "expected unchanged 2-vertex string, got %v", s.Coords)
	}
}

func TestSimpleNoder_SelfCrossing(t *testing.T) {
	// Figure-eight-like string: (0,0) -> (2,2) -> (2,0) -> (0,2).
	// The first edge (0,0)-(2,2) crosses the third edge (2,0)-(0,2)
	// at (1,1).
	ss := &SegmentString{
		Coords: []geom.XY{xy(0, 0), xy(2, 2), xy(2, 0), xy(0, 2)},
		Tag:    7,
	}
	out := SimpleNoder{}.Node([]*SegmentString{ss})

	// Expect the original 3 edges to produce: edge0 split at (1,1) ->
	// 2 pieces, edge1 unchanged -> 1 piece, edge2 split at (1,1) ->
	// 2 pieces. But pieces between consecutive vertices are
	// concatenated unless interrupted, so the actual output is:
	//   (0,0)-(1,1)
	//   (1,1)-(2,2)-(2,0)-(1,1)   // original v1 (2,2) and v2 (2,0) are not breaks
	//   (1,1)-(0,2)
	// Total: 3 strings.
	require.Equal(t, 3, len(out), "expected 3 noded substrings, got %d: %+v", len(out), dumpCoords(out))
	for _, s := range out {
		assert.Equal(t, 7, s.Tag, "tag preserved")
	}

	// Verify each output piece starts and ends at one of {original
	// endpoint, intersection point}.
	wantNodes := []geom.XY{xy(0, 0), xy(1, 1), xy(0, 2)}
	for _, s := range out {
		first := s.Coords[0]
		last := s.Coords[len(s.Coords)-1]
		assert.True(t, nodeIn(first, wantNodes) && nodeIn(last, wantNodes),
			"piece endpoints %v -> %v not in node set %v", first, last, wantNodes)
	}
}

func nodeIn(p geom.XY, set []geom.XY) bool {
	for _, q := range set {
		if approxEqual(p, q) {
			return true
		}
	}
	return false
}

func dumpCoords(out []*SegmentString) [][]geom.XY {
	r := make([][]geom.XY, len(out))
	for i, s := range out {
		r[i] = s.Coords
	}
	return r
}

// TestSimpleNoder_CoincidentSegments documents the collinear-overlap
// limitation: SegmentIntersection returns ok=false for parallel/
// collinear inputs, so the noder cannot detect overlap. The two input
// strings pass through unchanged. This test pins that behaviour so the
// limitation is visible in CI; when overlay adds explicit overlap
// handling, this test should be updated.
func TestSimpleNoder_CoincidentSegments(t *testing.T) {
	ssA := &SegmentString{Coords: []geom.XY{xy(0, 0), xy(2, 0)}, Tag: 1}
	ssB := &SegmentString{Coords: []geom.XY{xy(0, 0), xy(2, 0)}, Tag: 2}

	out := SimpleNoder{}.Node([]*SegmentString{ssA, ssB})
	require.Equal(t, 2, len(out),
		"expected 2 unchanged strings (collinear-overlap is a known limitation)")
	for _, s := range out {
		assert.True(t, stringHas(s, xy(0, 0), xy(2, 0)),
			"expected unchanged collinear string, got %v", s.Coords)
	}
}

func TestSimpleNoder_EmptyInput(t *testing.T) {
	assert.Nil(t, (SimpleNoder{}).Node(nil), "nil input should give nil output")
	assert.Nil(t, (SimpleNoder{}).Node([]*SegmentString{}), "empty input should give nil output")
}

func TestSimpleNoder_TagsPreserved(t *testing.T) {
	// Three strings, each crossing the others at distinct points.
	ssA := &SegmentString{Coords: []geom.XY{xy(0, 0), xy(4, 0)}, Tag: 10}
	ssB := &SegmentString{Coords: []geom.XY{xy(1, -1), xy(1, 1)}, Tag: 20}
	ssC := &SegmentString{Coords: []geom.XY{xy(3, -1), xy(3, 1)}, Tag: 30}

	out := SimpleNoder{}.Node([]*SegmentString{ssA, ssB, ssC})

	tags := map[int]int{}
	for _, s := range out {
		tags[s.Tag]++
	}
	// A is split at x=1 and x=3 -> 3 pieces.
	// B is split once at (1,0) -> 2 pieces.
	// C is split once at (3,0) -> 2 pieces.
	assert.Equal(t, 3, tags[10], "piece counts by tag = %v, want {10:3, 20:2, 30:2}", tags)
	assert.Equal(t, 2, tags[20], "piece counts by tag = %v, want {10:3, 20:2, 30:2}", tags)
	assert.Equal(t, 2, tags[30], "piece counts by tag = %v, want {10:3, 20:2, 30:2}", tags)
}

func TestSimpleNoder_RingClosed(t *testing.T) {
	// Closed square ring — first vertex equals last. No self-
	// intersections, so should round-trip unchanged as a single
	// 5-vertex string.
	ring := &SegmentString{
		Coords: []geom.XY{
			xy(0, 0), xy(1, 0), xy(1, 1), xy(0, 1), xy(0, 0),
		},
		Tag: 1,
	}
	out := SimpleNoder{}.Node([]*SegmentString{ring})
	require.Equal(t, 1, len(out), "expected 1 string")
	assert.Equal(t, 5, len(out[0].Coords), "expected 5 vertices, got %v", out[0].Coords)
}
