package buffer

import (
	"errors"
	"math"
	"testing"

	"github.com/terra-geo/terra"
	"github.com/terra-geo/terra/geom"
)

func TestBufferPointRing(t *testing.T) {
	p := geom.NewPoint(nil, geom.XY{X: 0, Y: 0})
	const radius = 5.0
	const quad = 8

	g, err := Buffer(p, radius, WithQuadSegments(quad))
	if err != nil {
		t.Fatalf("Buffer: %v", err)
	}
	poly, ok := g.(*geom.Polygon)
	if !ok {
		t.Fatalf("expected *geom.Polygon, got %T", g)
	}
	ring := poly.ExteriorRing()
	wantVerts := 4*quad + 1
	if len(ring) != wantVerts {
		t.Fatalf("ring length = %d, want %d", len(ring), wantVerts)
	}
	if !ring[0].Equal(ring[len(ring)-1]) {
		t.Fatalf("ring not closed: %+v vs %+v", ring[0], ring[len(ring)-1])
	}
	for i, v := range ring {
		got := math.Hypot(v.X, v.Y)
		if math.Abs(got-radius) > 1e-6 {
			t.Errorf("vertex %d: distance %v, want %v", i, got, radius)
		}
	}
}

func TestBufferPointZeroDistance(t *testing.T) {
	p := geom.NewPoint(nil, geom.XY{X: 1, Y: 2})
	g, err := Buffer(p, 0)
	if err != nil {
		t.Fatalf("Buffer: %v", err)
	}
	if g != p {
		t.Fatalf("expected identity geometry, got %T", g)
	}
}

func TestBufferPointNegativeDistance(t *testing.T) {
	p := geom.NewPoint(nil, geom.XY{X: 0, Y: 0})
	_, err := Buffer(p, -1)
	if !errors.Is(err, terra.ErrInvalidGeometry) {
		t.Fatalf("err = %v, want ErrInvalidGeometry", err)
	}
}

func TestBufferLineFlatCapRectangle(t *testing.T) {
	ls := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}})
	g, err := Buffer(ls, 1, WithCapStyle(CapFlat))
	if err != nil {
		t.Fatalf("Buffer: %v", err)
	}
	poly, ok := g.(*geom.Polygon)
	if !ok {
		t.Fatalf("expected *geom.Polygon, got %T", g)
	}
	ring := poly.ExteriorRing()
	if len(ring) != 5 {
		t.Fatalf("rectangle ring vertices = %d, want 5; ring=%+v", len(ring), ring)
	}
	if !ring[0].Equal(ring[len(ring)-1]) {
		t.Fatalf("ring not closed")
	}
	// Validate the four corners are (0, +1), (10, +1), (10, -1), (0, -1) in
	// some order.
	want := map[geom.XY]bool{
		{X: 0, Y: 1}:  true,
		{X: 10, Y: 1}: true,
		{X: 10, Y: -1}: true,
		{X: 0, Y: -1}:  true,
	}
	for _, v := range ring[:4] {
		if !want[v] {
			t.Errorf("unexpected corner %+v", v)
		}
		delete(want, v)
	}
	if len(want) != 0 {
		t.Errorf("missing corners: %+v", want)
	}
}

func TestBufferLineRoundCap(t *testing.T) {
	const quad = 8
	ls := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}})
	g, err := Buffer(ls, 1, WithCapStyle(CapRound), WithQuadSegments(quad))
	if err != nil {
		t.Fatalf("Buffer: %v", err)
	}
	poly := g.(*geom.Polygon)
	ring := poly.ExteriorRing()
	// Two left corners + two right corners + two semicircles each with
	// 2*quad - 1 interior vertices + closure.
	// = 4 + 2*(2*quad - 1) + 1 = 4 + 4*quad - 2 + 1 = 4*quad + 3.
	wantVerts := 4*quad + 3
	if len(ring) != wantVerts {
		t.Fatalf("round-cap ring vertices = %d, want %d", len(ring), wantVerts)
	}
	// Sanity: every vertex within 1 of the line segment [0,0]-[10,0].
	for i, v := range ring {
		dx := v.X
		if dx < 0 {
			dx = -dx
		}
		// Distance from segment.
		var dist float64
		if v.X < 0 {
			dist = math.Hypot(v.X, v.Y)
		} else if v.X > 10 {
			dist = math.Hypot(v.X-10, v.Y)
		} else {
			dist = math.Abs(v.Y)
		}
		if math.Abs(dist-1) > 1e-9 {
			t.Errorf("vertex %d %+v: dist from line = %v, want 1", i, v, dist)
		}
	}
}

func TestBufferLineSquareCap(t *testing.T) {
	ls := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}})
	g, err := Buffer(ls, 1, WithCapStyle(CapSquare))
	if err != nil {
		t.Fatalf("Buffer: %v", err)
	}
	poly := g.(*geom.Polygon)
	ring := poly.ExteriorRing()
	// 2 left ends + 2 square cap forward extensions + 2 right ends + 2
	// square cap backward extensions + closure = 8 + 1 = 9.
	if len(ring) != 9 {
		t.Fatalf("square-cap ring vertices = %d, want 9; %+v", len(ring), ring)
	}
}

func TestBufferLineNegativeDistance(t *testing.T) {
	ls := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}})
	_, err := Buffer(ls, -1)
	if !errors.Is(err, terra.ErrInvalidGeometry) {
		t.Fatalf("err = %v, want ErrInvalidGeometry", err)
	}
}

func TestBufferGeometryCollectionRejected(t *testing.T) {
	gc := geom.NewGeometryCollection(nil, geom.NewPoint(nil, geom.XY{}))
	_, err := Buffer(gc, 1)
	if err == nil {
		t.Fatal("expected error for collection input")
	}
}

func TestBufferMultiPoint(t *testing.T) {
	mp := geom.NewMultiPoint(nil, []geom.XY{{X: 0, Y: 0}, {X: 100, Y: 100}})
	g, err := Buffer(mp, 1, WithQuadSegments(4))
	if err != nil {
		t.Fatalf("Buffer: %v", err)
	}
	mpoly, ok := g.(*geom.MultiPolygon)
	if !ok {
		t.Fatalf("expected *geom.MultiPolygon, got %T", g)
	}
	if mpoly.NumGeometries() != 2 {
		t.Errorf("got %d members, want 2", mpoly.NumGeometries())
	}
}

func TestBufferMultiLineString(t *testing.T) {
	ls1 := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}})
	ls2 := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 100}, {X: 10, Y: 100}})
	mls := geom.NewMultiLineString(nil, ls1, ls2)
	g, err := Buffer(mls, 1)
	if err != nil {
		t.Fatalf("Buffer: %v", err)
	}
	mpoly, ok := g.(*geom.MultiPolygon)
	if !ok {
		t.Fatalf("expected *geom.MultiPolygon, got %T", g)
	}
	if mpoly.NumGeometries() != 2 {
		t.Errorf("got %d members, want 2", mpoly.NumGeometries())
	}
}

func TestBufferAreaPositive(t *testing.T) {
	cases := []struct {
		name string
		g    geom.Geometry
		dist float64
		opts []Option
	}{
		{"point round", geom.NewPoint(nil, geom.XY{X: 5, Y: 5}), 3, nil},
		{"line flat", geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}}), 2, []Option{WithCapStyle(CapFlat)}},
		{"line round", geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}}), 1, []Option{WithCapStyle(CapRound), WithJoinStyle(JoinRound)}},
		{"line mitre", geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}}), 1, []Option{WithJoinStyle(JoinMitre), WithCapStyle(CapFlat)}},
		{"line bevel", geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}}), 1, []Option{WithJoinStyle(JoinBevel), WithCapStyle(CapFlat)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g, err := Buffer(tc.g, tc.dist, tc.opts...)
			if err != nil {
				t.Fatalf("Buffer: %v", err)
			}
			poly, ok := g.(*geom.Polygon)
			if !ok {
				t.Fatalf("expected polygon, got %T", g)
			}
			a := math.Abs(shoelace(poly.ExteriorRing()))
			if a <= 0 {
				t.Errorf("area = %v, want > 0", a)
			}
		})
	}
}

func TestBufferEmptyPoint(t *testing.T) {
	p := geom.NewEmptyPoint(nil, geom.LayoutXY)
	g, err := Buffer(p, 5)
	if err != nil {
		t.Fatalf("Buffer: %v", err)
	}
	if !g.IsEmpty() {
		t.Errorf("expected empty result, got non-empty %T", g)
	}
}

func TestBufferNilGeometry(t *testing.T) {
	_, err := Buffer(nil, 1)
	if !errors.Is(err, terra.ErrInvalidGeometry) {
		t.Fatalf("err = %v, want ErrInvalidGeometry", err)
	}
}

func TestBufferNaNDistance(t *testing.T) {
	p := geom.NewPoint(nil, geom.XY{})
	_, err := Buffer(p, math.NaN())
	if !errors.Is(err, terra.ErrInvalidGeometry) {
		t.Fatalf("err = %v, want ErrInvalidGeometry", err)
	}
}

func TestBufferMitreLimitFallsBackToBevel(t *testing.T) {
	// Near-180° turn: a→b→c almost reverses direction. Mitre extension
	// would be huge; with a small mitre limit it should fall back to bevel.
	ls := geom.NewLineString(nil, []geom.XY{
		{X: 0, Y: 0},
		{X: 10, Y: 0},
		{X: 0, Y: 0.01}, // turn back almost 180°
	})
	g, err := Buffer(ls, 1, WithJoinStyle(JoinMitre), WithMitreLimit(1.5), WithCapStyle(CapFlat))
	if err != nil {
		t.Fatalf("Buffer: %v", err)
	}
	poly := g.(*geom.Polygon)
	ring := poly.ExteriorRing()
	// Verify no vertex is wildly far from the line (sanity that bevel
	// fallback engaged).
	for _, v := range ring {
		if math.Hypot(v.X, v.Y) > 100 {
			t.Errorf("vertex %+v far from origin — mitre fallback did not engage", v)
		}
	}
}

// --- helpers ---

func shoelace(ring []geom.XY) float64 {
	if len(ring) < 3 {
		return 0
	}
	var sum float64
	for i := 0; i < len(ring)-1; i++ {
		sum += ring[i].X*ring[i+1].Y - ring[i+1].X*ring[i].Y
	}
	return sum / 2
}
