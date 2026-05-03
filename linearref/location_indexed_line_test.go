package linearref

import (
	"math"
	"testing"

	"github.com/terra-geo/terra/geom"
)

func TestLocIndexedExtractPoint(t *testing.T) {
	idx := NewLocationIndexedLine(line100())
	got := idx.ExtractPoint(NewLinearLocation(0, 0.5))
	if got.X != 25 || got.Y != 0 {
		t.Fatalf("midpoint: %+v", got)
	}
}

func TestLocIndexedExtractLine(t *testing.T) {
	idx := NewLocationIndexedLine(line100())
	// from 25%-of-segment-0 to 50%-of-segment-1 = X in [25, 75].
	start := NewLinearLocation(0, 0.5)
	end := NewLinearLocation(1, 0.5)
	sub := idx.ExtractLine(start, end)
	ls, ok := sub.(*geom.LineString)
	if !ok {
		t.Fatalf("expected LineString, got %T", sub)
	}
	if ls.NumPoints() < 2 {
		t.Fatalf("subline has %d pts", ls.NumPoints())
	}
	// First and last points
	first := ls.PointAt(0)
	last := ls.PointAt(ls.NumPoints() - 1)
	if math.Abs(first.X-25) > 1e-9 {
		t.Fatalf("first %+v", first)
	}
	if math.Abs(last.X-75) > 1e-9 {
		t.Fatalf("last %+v", last)
	}
}

func TestLocIndexedExtractLineReversed(t *testing.T) {
	idx := NewLocationIndexedLine(line100())
	end := NewLinearLocation(0, 0.5)
	start := NewLinearLocation(1, 0.5)
	sub := idx.ExtractLine(start, end)
	ls := sub.(*geom.LineString)
	first := ls.PointAt(0)
	last := ls.PointAt(ls.NumPoints() - 1)
	if math.Abs(first.X-75) > 1e-9 || math.Abs(last.X-25) > 1e-9 {
		t.Fatalf("expected reverse: first %+v last %+v", first, last)
	}
}

func TestLocIndexedProjectExternal(t *testing.T) {
	// Y-aligned offset: project should land on the line at the foot.
	idx := NewLocationIndexedLine(line100())
	loc := idx.Project(geom.XY{X: 30, Y: 25})
	got := loc.GetCoordinate(line100())
	if math.Abs(got.X-30) > 1e-9 || got.Y != 0 {
		t.Fatalf("projected coord: %+v", got)
	}
}

func TestLocIndexedIndexOf(t *testing.T) {
	idx := NewLocationIndexedLine(line100())
	loc := idx.IndexOf(geom.XY{X: 70, Y: 0})
	got := loc.GetCoordinate(line100())
	if math.Abs(got.X-70) > 1e-9 {
		t.Fatalf("indexOf(70,0): got %+v", got)
	}
}

func TestLocIndexedIsValidIndex(t *testing.T) {
	idx := NewLocationIndexedLine(line100())
	if !idx.IsValidIndex(idx.StartIndex()) {
		t.Fatal("start should be valid")
	}
	if !idx.IsValidIndex(idx.EndIndex()) {
		t.Fatal("end should be valid")
	}
	bad := LinearLocation{ComponentIndex: 99}
	if idx.IsValidIndex(bad) {
		t.Fatal("out-of-range should not be valid")
	}
}
