package buffer

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/measure"
)

// TestPolygonize_ConvexSquare: a 10×10 CCW square offset outward by
// d=1 yields a 12×12 square (area 144) of buffer interior.
func TestPolygonize_ConvexSquare(t *testing.T) {
	// Outer ring CCW: (0,0)-(10,0)-(10,10)-(0,10)
	// Outward offset for d=1: (-1,-1)-(11,-1)-(11,11)-(-1,11)
	// Walked in same direction as original (CCW): buffer-interior is on
	// LEFT, depthDelta = +1.
	segs := []offsetSegment{
		{p0: geom.XY{X: -1, Y: -1}, p1: geom.XY{X: 11, Y: -1}, depthDelta: 1},
		{p0: geom.XY{X: 11, Y: -1}, p1: geom.XY{X: 11, Y: 11}, depthDelta: 1},
		{p0: geom.XY{X: 11, Y: 11}, p1: geom.XY{X: -1, Y: 11}, depthDelta: 1},
		{p0: geom.XY{X: -1, Y: 11}, p1: geom.XY{X: -1, Y: -1}, depthDelta: 1},
	}
	got, err := polygonizeBuffer(nil, segs, 0)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.False(t, got.IsEmpty(), "result should be non-empty")
	assert.InDelta(t, 144.0, measure.Area(got), 1e-9, "buffer area")
	if p, ok := got.(*geom.Polygon); ok {
		assert.Equal(t, 1, p.NumRings(), "single ring expected")
	}
}

// TestPolygonize_SquareWithHole: the buffer-interior region of a 20×20
// CCW outer with a 6×6 CW hole at the centre, offset both rings
// outward by d=1, should yield a polygon-with-hole: outer 22×22
// (area 484) minus shrunk hole 4×4 (area 16) = 468.
func TestPolygonize_SquareWithHole(t *testing.T) {
	// Outer offset CCW: (-1,-1)-(21,-1)-(21,21)-(-1,21)
	// Hole offset CW (the SHRUNK hole): walking CW. Original hole CW
	// vertices (with hole interior on RIGHT of CW direction): the
	// shrunk hole is INSIDE the original hole, also CW. For positive
	// buffer, the offset is on the side AWAY from polygon interior =
	// INSIDE hole. Walked CW: polygon interior on LEFT (= buffer
	// interior side). depthDelta = +1.
	segs := []offsetSegment{
		// Outer offset (CCW, depthDelta=+1).
		{p0: geom.XY{X: -1, Y: -1}, p1: geom.XY{X: 21, Y: -1}, depthDelta: 1},
		{p0: geom.XY{X: 21, Y: -1}, p1: geom.XY{X: 21, Y: 21}, depthDelta: 1},
		{p0: geom.XY{X: 21, Y: 21}, p1: geom.XY{X: -1, Y: 21}, depthDelta: 1},
		{p0: geom.XY{X: -1, Y: 21}, p1: geom.XY{X: -1, Y: -1}, depthDelta: 1},
		// Hole offset (CW, depthDelta=+1) — the SHRUNK hole at (8,8)-(12,8)-(12,12)-(8,12).
		// Original hole CW: (7,7)-(7,13)-(13,13)-(13,7). Inward into hole at d=1:
		// (8,8)-(8,12)-(12,12)-(12,8) walked CW.
		{p0: geom.XY{X: 8, Y: 8}, p1: geom.XY{X: 8, Y: 12}, depthDelta: 1},
		{p0: geom.XY{X: 8, Y: 12}, p1: geom.XY{X: 12, Y: 12}, depthDelta: 1},
		{p0: geom.XY{X: 12, Y: 12}, p1: geom.XY{X: 12, Y: 8}, depthDelta: 1},
		{p0: geom.XY{X: 12, Y: 8}, p1: geom.XY{X: 8, Y: 8}, depthDelta: 1},
	}
	got, err := polygonizeBuffer(nil, segs, 0)
	require.NoError(t, err)
	require.NotNil(t, got)
	expected := 22.0*22 - 4.0*4 // 484 - 16 = 468
	assert.InDelta(t, expected, measure.Area(got), 1e-9, "buffer area with hole")
}

// TestPolygonize_RayCastBasics: the rayCastDepth helper must respect
// the half-open winding-number convention so a vertex at the ray's
// y-level is counted on at most one of its incident edges.
func TestPolygonize_RayCastBasics(t *testing.T) {
	cases := []struct {
		name   string
		p      geom.XY
		segs   []offsetSegment
		expect int
	}{
		{
			name: "inside CCW square, depth=+1",
			p:    geom.XY{X: 0, Y: 0},
			segs: []offsetSegment{
				{p0: geom.XY{X: -1, Y: -1}, p1: geom.XY{X: 1, Y: -1}, depthDelta: 1},
				{p0: geom.XY{X: 1, Y: -1}, p1: geom.XY{X: 1, Y: 1}, depthDelta: 1},
				{p0: geom.XY{X: 1, Y: 1}, p1: geom.XY{X: -1, Y: 1}, depthDelta: 1},
				{p0: geom.XY{X: -1, Y: 1}, p1: geom.XY{X: -1, Y: -1}, depthDelta: 1},
			},
			expect: 1,
		},
		{
			name: "outside CCW square, depth=0",
			p:    geom.XY{X: 5, Y: 0},
			segs: []offsetSegment{
				{p0: geom.XY{X: -1, Y: -1}, p1: geom.XY{X: 1, Y: -1}, depthDelta: 1},
				{p0: geom.XY{X: 1, Y: -1}, p1: geom.XY{X: 1, Y: 1}, depthDelta: 1},
				{p0: geom.XY{X: 1, Y: 1}, p1: geom.XY{X: -1, Y: 1}, depthDelta: 1},
				{p0: geom.XY{X: -1, Y: 1}, p1: geom.XY{X: -1, Y: -1}, depthDelta: 1},
			},
			expect: 0,
		},
		{
			name: "inside CW square (= reversed CCW), depth=-1",
			p:    geom.XY{X: 0, Y: 0},
			segs: []offsetSegment{
				{p0: geom.XY{X: -1, Y: -1}, p1: geom.XY{X: -1, Y: 1}, depthDelta: 1},
				{p0: geom.XY{X: -1, Y: 1}, p1: geom.XY{X: 1, Y: 1}, depthDelta: 1},
				{p0: geom.XY{X: 1, Y: 1}, p1: geom.XY{X: 1, Y: -1}, depthDelta: 1},
				{p0: geom.XY{X: 1, Y: -1}, p1: geom.XY{X: -1, Y: -1}, depthDelta: 1},
			},
			expect: -1,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := rayCastDepth(c.p, c.segs)
			assert.Equal(t, c.expect, got)
		})
	}
}

// TestPolygonize_PolygonRoundTrip: emit offset segments from a real
// CCW polygon and verify the polygonizer produces the expected dilated
// shape. 10×10 square at distance +1 → 12×12 square (area 144).
func TestPolygonize_PolygonRoundTrip(t *testing.T) {
	poly := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0},
	})
	cfg := config{join: JoinMitre, mitreLimit: 5, quadSegments: 4}
	segs := emitPolygonOffsetSegments(poly, 1.0, cfg)
	require.NotEmpty(t, segs)
	got, err := polygonizeBuffer(nil, segs, 0)
	require.NoError(t, err)
	assert.InDelta(t, 144.0, measure.Area(got), 1e-9, "12x12 square area")
}

// TestPolygonize_PolygonWithHoleRoundTrip: 20×20 outer with a 6×6
// hole at the centre, distance +1 → 22×22 - 4×4 = 468.
func TestPolygonize_PolygonWithHoleRoundTrip(t *testing.T) {
	poly := geom.NewPolygon(nil,
		[]geom.XY{{X: 0, Y: 0}, {X: 20, Y: 0}, {X: 20, Y: 20}, {X: 0, Y: 20}, {X: 0, Y: 0}},
		[]geom.XY{{X: 7, Y: 7}, {X: 7, Y: 13}, {X: 13, Y: 13}, {X: 13, Y: 7}, {X: 7, Y: 7}},
	)
	cfg := config{join: JoinMitre, mitreLimit: 5, quadSegments: 4}
	segs := emitPolygonOffsetSegments(poly, 1.0, cfg)
	require.NotEmpty(t, segs)
	got, err := polygonizeBuffer(nil, segs, 0)
	require.NoError(t, err)
	expected := 22.0*22 - 4.0*4
	assert.InDelta(t, expected, measure.Area(got), 1e-9, "polygon-with-hole dilation")
}

// TestPolygonize_SelfIntersectingOffset: a deliberately self-intersecting
// "bowtie" offset curve (which arises from a concave reflex corner)
// should resolve to two disjoint regions where depth >= 1, even though
// the un-noded ring crosses itself.
//
// The bowtie has vertices (0,0)-(10,10)-(10,0)-(0,10)-(0,0), which is
// a CCW + CW intertwined. The DCEL build with snap-rounding will node
// the crossing at (5,5) and produce two triangular faces; one has
// depth +1 (the kept buffer face), the other -1 (cancelled).
func TestPolygonize_SelfIntersectingOffset(t *testing.T) {
	segs := []offsetSegment{
		{p0: geom.XY{X: 0, Y: 0}, p1: geom.XY{X: 10, Y: 10}, depthDelta: 1},
		{p0: geom.XY{X: 10, Y: 10}, p1: geom.XY{X: 10, Y: 0}, depthDelta: 1},
		{p0: geom.XY{X: 10, Y: 0}, p1: geom.XY{X: 0, Y: 10}, depthDelta: 1},
		{p0: geom.XY{X: 0, Y: 10}, p1: geom.XY{X: 0, Y: 0}, depthDelta: 1},
	}
	got, err := polygonizeBuffer(nil, segs, 0)
	require.NoError(t, err)
	require.NotNil(t, got)
	// The kept region is one triangle of the bowtie. Specifically: by
	// symmetry the depth of each triangle is determined by orientation
	// — the LOWER triangle (vertices (0,0),(10,0),(5,5)) has depth +1,
	// the UPPER triangle (5,5),(10,10),(0,10) has depth -1.
	// Total kept area = 25 (one triangle, base 10, height 5).
	assert.InDelta(t, 25.0, math.Abs(measure.Area(got)), 1e-9, "single triangle kept")
}
