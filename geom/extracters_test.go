package geom

import "testing"

func TestPointsOf(t *testing.T) {
	gc := NewGeometryCollection(nil,
		NewPoint(nil, XY{1, 1}),
		NewLineString(nil, []XY{{0, 0}, {1, 0}}),
		NewMultiPoint(nil, []XY{{2, 2}, {3, 3}}),
	)
	pts := PointsOf(gc)
	if len(pts) != 3 {
		t.Fatalf("got %d points, want 3", len(pts))
	}
}

func TestLineStringsOf(t *testing.T) {
	mls := NewMultiLineString(nil,
		NewLineString(nil, []XY{{0, 0}, {1, 1}}),
		NewLineString(nil, []XY{{2, 2}, {3, 3}}),
	)
	gc := NewGeometryCollection(nil, mls, NewPoint(nil, XY{5, 5}))
	ls := LineStringsOf(gc)
	if len(ls) != 2 {
		t.Fatalf("got %d, want 2", len(ls))
	}
}

func TestPolygonsOf(t *testing.T) {
	p := NewPolygon(nil, []XY{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}})
	mp := NewMultiPolygon(nil, p, p)
	gc := NewGeometryCollection(nil, mp, p)
	got := PolygonsOf(gc)
	if len(got) != 3 {
		t.Fatalf("got %d, want 3", len(got))
	}
}

func TestExtracters_Empty(t *testing.T) {
	if pts := PointsOf(nil); pts != nil {
		t.Errorf("nil should give nil, got %v", pts)
	}
	gc := NewGeometryCollection(nil)
	if len(PolygonsOf(gc)) != 0 {
		t.Errorf("empty collection should give zero polygons")
	}
}
