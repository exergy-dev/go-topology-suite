package geom

import "testing"

func TestEdit_Point(t *testing.T) {
	p := NewPoint(nil, XY{1, 2})
	out := Edit(p, func(xy XY) XY { return XY{xy.X + 1, xy.Y * 2} })
	got, ok := out.(*Point)
	if !ok {
		t.Fatalf("expected *Point, got %T", out)
	}
	if got.XY() != (XY{2, 4}) {
		t.Errorf("got %v, want (2,4)", got.XY())
	}
}

func TestEdit_LineString(t *testing.T) {
	ls := NewLineString(nil, []XY{{0, 0}, {1, 1}, {2, 0}})
	out := Edit(ls, func(xy XY) XY { return XY{xy.X * 10, xy.Y * 10} }).(*LineString)
	if out.NumPoints() != 3 {
		t.Fatalf("npoints %d", out.NumPoints())
	}
	if out.PointAt(1) != (XY{10, 10}) {
		t.Errorf("got %v", out.PointAt(1))
	}
}

func TestEdit_Polygon(t *testing.T) {
	shell := []XY{{0, 0}, {10, 0}, {10, 10}, {0, 10}, {0, 0}}
	hole := []XY{{2, 2}, {4, 2}, {4, 4}, {2, 4}, {2, 2}}
	p := NewPolygon(nil, shell, hole)
	out := Edit(p, func(xy XY) XY { return XY{xy.X + 100, xy.Y + 100} }).(*Polygon)
	if out.NumRings() != 2 {
		t.Fatalf("rings %d", out.NumRings())
	}
	if out.Ring(0)[0] != (XY{100, 100}) {
		t.Errorf("shell[0] = %v", out.Ring(0)[0])
	}
	if out.Ring(1)[0] != (XY{102, 102}) {
		t.Errorf("hole[0] = %v", out.Ring(1)[0])
	}
}

func TestEdit_GeometryCollection(t *testing.T) {
	gc := NewGeometryCollection(nil,
		NewPoint(nil, XY{1, 1}),
		NewLineString(nil, []XY{{0, 0}, {2, 2}}))
	out := Edit(gc, func(xy XY) XY { return XY{-xy.X, -xy.Y} }).(*GeometryCollection)
	if out.NumGeometries() != 2 {
		t.Fatalf("ngeoms %d", out.NumGeometries())
	}
	pt := out.GeometryAt(0).(*Point)
	if pt.XY() != (XY{-1, -1}) {
		t.Errorf("pt = %v", pt.XY())
	}
}
