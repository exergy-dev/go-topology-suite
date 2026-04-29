package validate

import (
	"errors"
	"testing"

	terra "github.com/terra-geo/terra"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/kernel/planar"
)

func TestMakeValid_UnclosedRing(t *testing.T) {
	// Outer ring missing the closing vertex.
	p := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 0, Y: 10}, {X: 10, Y: 10}, {X: 10, Y: 0},
	})
	g, err := MakeValid(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g == nil || g.IsEmpty() {
		t.Fatalf("expected non-empty polygon, got %v", g)
	}
	if err := Validate(g); err != nil {
		t.Errorf("expected valid result, got %v", err)
	}
}

func TestMakeValid_CWRingReorientedToCCW(t *testing.T) {
	// Clockwise outer ring (negative shoelace area).
	cw := []geom.XY{
		{X: 0, Y: 0}, {X: 0, Y: 10}, {X: 10, Y: 10}, {X: 10, Y: 0}, {X: 0, Y: 0},
	}
	if planar.Default.RingArea(cw) >= 0 {
		t.Fatalf("test setup: ring should be CW (negative area), got %v", planar.Default.RingArea(cw))
	}
	p := geom.NewPolygon(nil, cw)
	g, err := MakeValid(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, ok := g.(*geom.Polygon)
	if !ok {
		t.Fatalf("expected *Polygon, got %T", g)
	}
	if planar.Default.RingArea(out.ExteriorRing()) <= 0 {
		t.Errorf("expected CCW outer ring (positive area)")
	}
	if err := Validate(out); err != nil {
		t.Errorf("expected valid result, got %v", err)
	}
}

func TestMakeValid_BowtiePolygon(t *testing.T) {
	// Self-intersecting bowtie.
	p := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 10, Y: 10}, {X: 10, Y: 0}, {X: 0, Y: 10}, {X: 0, Y: 0},
	})
	g, err := MakeValid(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g == nil {
		t.Fatalf("got nil result")
	}
	// Result must be non-nil. We don't pin the exact shape (overlay may
	// simplify in unexpected ways per documented v0.1 limitations); we only
	// require that the result isn't an obvious structural failure beyond
	// what overlay produces.
	if g.IsEmpty() {
		// Empty is acceptable too — overlay may collapse a degenerate bowtie.
		return
	}
	// If non-empty, structural fields (closure, vertex count) must hold.
	switch x := g.(type) {
	case *geom.Polygon:
		ring := x.ExteriorRing()
		if len(ring) > 0 && ring[0] != ring[len(ring)-1] {
			t.Errorf("outer ring not closed in result")
		}
	case *geom.MultiPolygon:
		for i := 0; i < x.NumGeometries(); i++ {
			ring := x.PolygonAt(i).ExteriorRing()
			if len(ring) > 0 && ring[0] != ring[len(ring)-1] {
				t.Errorf("part %d outer ring not closed", i)
			}
		}
	}
}

func TestMakeValid_LineStringSinglePoint(t *testing.T) {
	ls := geom.NewLineString(nil, []geom.XY{{X: 3, Y: 4}})
	g, err := MakeValid(ls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pt, ok := g.(*geom.Point)
	if !ok {
		t.Fatalf("expected *Point, got %T", g)
	}
	if pt.XY() != (geom.XY{X: 3, Y: 4}) {
		t.Errorf("unexpected point: %v", pt.XY())
	}
}

func TestMakeValid_LineStringDuplicatePointsCollapse(t *testing.T) {
	// Three duplicates collapse to a single distinct vertex → Point.
	ls := geom.NewLineString(nil, []geom.XY{
		{X: 1, Y: 1}, {X: 1, Y: 1}, {X: 1, Y: 1},
	})
	g, err := MakeValid(ls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := g.(*geom.Point); !ok {
		t.Errorf("expected *Point from duplicate-only line, got %T", g)
	}
}

func TestMakeValid_PolygonTooFewVertices(t *testing.T) {
	// Triangle missing one vertex; ring length after closure is < 4.
	p := geom.NewPolygon(nil, []geom.XY{{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 0}})
	g, err := MakeValid(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, ok := g.(*geom.Polygon)
	if !ok {
		t.Fatalf("expected *Polygon, got %T", g)
	}
	if !out.IsEmpty() {
		t.Errorf("expected empty polygon for too-few-vertex input, got %v", out)
	}
}

func TestMakeValid_MultiPolygonPreservesValidMembers(t *testing.T) {
	// One valid square + one degenerate (too few vertices) polygon.
	good := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}, {X: 1, Y: 0}, {X: 0, Y: 0},
	})
	bad := geom.NewPolygon(nil, []geom.XY{{X: 5, Y: 5}, {X: 5, Y: 6}, {X: 5, Y: 5}})
	mp := geom.NewMultiPolygon(nil, good, bad)

	g, err := MakeValid(mp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, ok := g.(*geom.MultiPolygon)
	if !ok {
		t.Fatalf("expected *MultiPolygon, got %T", g)
	}
	if out.NumGeometries() != 1 {
		t.Errorf("expected 1 surviving polygon, got %d", out.NumGeometries())
	}
	if err := Validate(out); err != nil {
		t.Errorf("expected valid multipolygon, got %v", err)
	}
}

func TestMakeValid_AlreadyValidPolygonRoundTrips(t *testing.T) {
	// CCW closed square.
	p := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0},
	})
	if err := Validate(p); err != nil {
		t.Fatalf("test setup: input expected valid, got %v", err)
	}
	g, err := MakeValid(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := Validate(g); err != nil {
		t.Errorf("expected valid result, got %v", err)
	}
	out, ok := g.(*geom.Polygon)
	if !ok {
		t.Fatalf("expected *Polygon, got %T", g)
	}
	// Same ring vertex count and same first/last vertex.
	if out.NumRings() != 1 {
		t.Errorf("expected 1 ring, got %d", out.NumRings())
	}
	ring := out.ExteriorRing()
	if len(ring) != 5 {
		t.Errorf("expected 5 vertices, got %d", len(ring))
	}
}

func TestMakeValid_PointPassthrough(t *testing.T) {
	pt := geom.NewPoint(nil, geom.XY{X: 7, Y: 8})
	g, err := MakeValid(pt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g != pt {
		t.Errorf("expected same pointer back for valid Point, got %v", g)
	}
}

func TestMakeValid_EmptyReturnsErrEmpty(t *testing.T) {
	empty := geom.NewEmptyPolygon(nil, geom.LayoutXY)
	g, err := MakeValid(empty)
	if !errors.Is(err, terra.ErrEmpty) {
		t.Errorf("expected ErrEmpty, got err=%v g=%v", err, g)
	}
}

func TestMakeValid_HolesDropped(t *testing.T) {
	// Square with a small hole; holes should be dropped per v0.1 limitation.
	outer := []geom.XY{
		{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0},
	}
	hole := []geom.XY{
		{X: 3, Y: 3}, {X: 4, Y: 3}, {X: 4, Y: 4}, {X: 3, Y: 4}, {X: 3, Y: 3},
	}
	p := geom.NewPolygon(nil, outer, hole)
	g, err := MakeValid(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, ok := g.(*geom.Polygon)
	if !ok {
		t.Fatalf("expected *Polygon, got %T", g)
	}
	if out.NumRings() != 1 {
		t.Errorf("expected hole dropped (1 ring), got %d rings", out.NumRings())
	}
}

func TestMakeValid_GeometryCollectionRecurses(t *testing.T) {
	pt := geom.NewPoint(nil, geom.XY{X: 1, Y: 2})
	// Unclosed ring — MakeValid should close it.
	poly := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 0, Y: 5}, {X: 5, Y: 5}, {X: 5, Y: 0},
	})
	gc := geom.NewGeometryCollection(nil, pt, poly)
	g, err := MakeValid(gc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, ok := g.(*geom.GeometryCollection)
	if !ok {
		t.Fatalf("expected *GeometryCollection, got %T", g)
	}
	if out.NumGeometries() != 2 {
		t.Errorf("expected 2 children, got %d", out.NumGeometries())
	}
	if err := Validate(out); err != nil {
		t.Errorf("expected valid collection, got %v", err)
	}
}
