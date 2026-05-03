package geom

import (
	"testing"
)

func TestOctagonalEnvelope_NullByDefault(t *testing.T) {
	oe := NewOctagonalEnvelope()
	if !oe.IsNull() {
		t.Fatalf("expected null envelope")
	}
	if oe.IntersectsXY(XY{X: 0, Y: 0}) {
		t.Fatalf("null envelope should not intersect any point")
	}
}

func TestOctagonalEnvelope_ExpandSinglePoint(t *testing.T) {
	oe := NewOctagonalEnvelope()
	oe.ExpandToInclude(3, 4)
	if oe.IsNull() {
		t.Fatalf("expected non-null after expand")
	}
	if oe.MinX() != 3 || oe.MaxX() != 3 || oe.MinY() != 4 || oe.MaxY() != 4 {
		t.Fatalf("axis bounds wrong: %v..%v / %v..%v",
			oe.MinX(), oe.MaxX(), oe.MinY(), oe.MaxY())
	}
	if oe.MinA() != 7 || oe.MaxA() != 7 || oe.MinB() != -1 || oe.MaxB() != -1 {
		t.Fatalf("diag bounds wrong: A=%v..%v B=%v..%v",
			oe.MinA(), oe.MaxA(), oe.MinB(), oe.MaxB())
	}
}

func TestOctagonalEnvelope_FromGeometryAndIntersects(t *testing.T) {
	g := NewPolygon(nil, []XY{{0, 0}, {4, 0}, {4, 4}, {0, 4}, {0, 0}})
	oe := NewOctagonalEnvelopeFromGeometry(g)
	if oe.IsNull() {
		t.Fatalf("expected non-null")
	}
	// Inside the square — must be inside the octagon too.
	for _, p := range []XY{{2, 2}, {0, 0}, {4, 4}, {1, 3}} {
		if !oe.IntersectsXY(p) {
			t.Fatalf("expected octagon to contain %+v", p)
		}
	}
	// Far away should not.
	if oe.IntersectsXY(XY{X: -10, Y: -10}) {
		t.Fatalf("octagon should not contain far point")
	}
}

func TestOctagonalEnvelope_ContainsAndIntersects(t *testing.T) {
	g := NewPolygon(nil, []XY{{0, 0}, {10, 0}, {10, 10}, {0, 10}, {0, 0}})
	oe := NewOctagonalEnvelopeFromGeometry(g)
	inner := NewOctagonalEnvelope()
	inner.ExpandToInclude(2, 2)
	inner.ExpandToInclude(5, 5)
	if !oe.Contains(inner) {
		t.Fatalf("oe should contain inner")
	}
	if !oe.Intersects(inner) {
		t.Fatalf("oe should intersect inner")
	}
	disjoint := NewOctagonalEnvelope()
	disjoint.ExpandToInclude(100, 100)
	if oe.Intersects(disjoint) {
		t.Fatalf("disjoint envelopes should not intersect")
	}
}

func TestOctagonalEnvelope_ToGeometry_Polygon(t *testing.T) {
	g := NewPolygon(nil, []XY{{0, 0}, {10, 0}, {10, 10}, {0, 10}, {0, 0}})
	oe := NewOctagonalEnvelopeFromGeometry(g)
	got := oe.ToGeometry(nil)
	poly, ok := got.(*Polygon)
	if !ok {
		t.Fatalf("expected Polygon, got %T", got)
	}
	if poly.NumRings() == 0 {
		t.Fatalf("expected non-empty polygon")
	}
	// All input vertices should be contained in the octagon.
	for _, p := range []XY{{0, 0}, {10, 0}, {10, 10}, {0, 10}} {
		if !oe.IntersectsXY(p) {
			t.Fatalf("octagon should contain %+v", p)
		}
	}
}

func TestOctagonalEnvelope_ToGeometry_Point(t *testing.T) {
	oe := NewOctagonalEnvelope()
	oe.ExpandToInclude(5, 5)
	g := oe.ToGeometry(nil)
	if _, ok := g.(*Point); !ok {
		t.Fatalf("expected Point, got %T", g)
	}
}

func TestOctagonalEnvelope_ToGeometry_NullEmpty(t *testing.T) {
	oe := NewOctagonalEnvelope()
	g := oe.ToGeometry(nil)
	if !g.IsEmpty() {
		t.Fatalf("expected empty geometry for null envelope")
	}
}

func TestOctagonalEnvelope_ExpandBy(t *testing.T) {
	oe := NewOctagonalEnvelope()
	oe.ExpandToInclude(0, 0)
	oe.ExpandBy(1)
	if oe.IsNull() {
		t.Fatalf("non-null expected after positive expandBy")
	}
	if !oe.IntersectsXY(XY{X: 0.5, Y: 0.5}) {
		t.Fatalf("expanded envelope should contain (0.5,0.5)")
	}
}
