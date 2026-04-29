package geom

import (
	"testing"

	"github.com/terra-geo/terra/crs"
)

func TestPointConstruction(t *testing.T) {
	p := NewPoint(crs.WGS84, XY{-75.16, 39.95})
	if p.IsEmpty() {
		t.Fatalf("point should not be empty")
	}
	if p.Type() != PointType {
		t.Errorf("Type = %v", p.Type())
	}
	if p.Layout() != LayoutXY {
		t.Errorf("Layout = %v", p.Layout())
	}
	if p.NumGeometries() != 1 {
		t.Errorf("NumGeometries = %d", p.NumGeometries())
	}
	got := p.XY()
	if got.X != -75.16 || got.Y != 39.95 {
		t.Errorf("XY() = %+v", got)
	}
	env := p.Envelope()
	if env.MinX != -75.16 || env.MaxX != -75.16 {
		t.Errorf("envelope wrong: %+v", env)
	}
}

func TestEmptyPoint(t *testing.T) {
	p := NewEmptyPoint(crs.WGS84, LayoutXY)
	if !p.IsEmpty() {
		t.Errorf("empty point should be empty")
	}
	if !p.Envelope().IsEmpty() {
		t.Errorf("empty point envelope should be empty")
	}
}

func TestLineStringConstruction(t *testing.T) {
	ls := NewLineString(crs.WGS84, []XY{{0, 0}, {1, 1}, {2, 2}})
	if ls.NumPoints() != 3 {
		t.Fatalf("NumPoints = %d, want 3", ls.NumPoints())
	}
	if got := ls.PointAt(1); got.X != 1 || got.Y != 1 {
		t.Errorf("PointAt(1) = %+v", got)
	}
	count := 0
	for p := range ls.CoordsXY() {
		_ = p
		count++
	}
	if count != 3 {
		t.Errorf("CoordsXY iterator yielded %d, want 3", count)
	}
}

func TestPolygonConstruction(t *testing.T) {
	outer := []XY{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}
	hole := []XY{{2, 2}, {2, 4}, {4, 4}, {4, 2}, {2, 2}}
	p := NewPolygon(crs.WGS84, outer, hole)
	if p.NumRings() != 2 {
		t.Fatalf("NumRings = %d, want 2", p.NumRings())
	}
	got := p.ExteriorRing()
	if len(got) != 5 {
		t.Errorf("exterior len = %d, want 5", len(got))
	}
	holes := p.InteriorRings()
	if len(holes) != 1 || len(holes[0]) != 5 {
		t.Errorf("holes wrong: %+v", holes)
	}
	env := p.Envelope()
	if env.MinX != 0 || env.MaxX != 10 || env.MinY != 0 || env.MaxY != 10 {
		t.Errorf("polygon envelope = %+v", env)
	}
}

func TestMultiLineStringEnvelopeUnion(t *testing.T) {
	a := NewLineString(crs.WGS84, []XY{{0, 0}, {1, 1}})
	b := NewLineString(crs.WGS84, []XY{{5, 5}, {6, 6}})
	m := NewMultiLineString(crs.WGS84, a, b)
	env := m.Envelope()
	if env.MinX != 0 || env.MaxX != 6 || env.MinY != 0 || env.MaxY != 6 {
		t.Errorf("union envelope = %+v", env)
	}
	if m.NumGeometries() != 2 {
		t.Errorf("NumGeometries = %d", m.NumGeometries())
	}
}

func TestGeometryInterfaceSatisfaction(t *testing.T) {
	// Compile-time checks that every concrete type implements Geometry.
	var _ Geometry = (*Point)(nil)
	var _ Geometry = (*LineString)(nil)
	var _ Geometry = (*Polygon)(nil)
	var _ Geometry = (*MultiPoint)(nil)
	var _ Geometry = (*MultiLineString)(nil)
	var _ Geometry = (*MultiPolygon)(nil)
	var _ Geometry = (*GeometryCollection)(nil)
}
