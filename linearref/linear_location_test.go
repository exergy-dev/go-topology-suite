package linearref

import (
	"math"
	"testing"

	"github.com/terra-geo/terra/geom"
)

// line100 returns a horizontal LineString of total length 100 (two
// 50-unit segments).
func line100() *geom.LineString {
	return geom.NewLineString(nil, []geom.XY{
		{X: 0, Y: 0},
		{X: 50, Y: 0},
		{X: 100, Y: 0},
	})
}

func TestLinearLocationNormalize(t *testing.T) {
	loc := NewLinearLocation(0, 1.0)
	if loc.SegmentIndex != 1 || loc.SegmentFraction != 0 {
		t.Fatalf("fraction=1 should advance segment: got %+v", loc)
	}
	loc = NewLinearLocation(2, -0.5)
	if loc.SegmentFraction != 0 {
		t.Fatalf("negative fraction should clamp to 0: got %+v", loc)
	}
	loc = NewLinearLocation(2, 1.5)
	if loc.SegmentFraction != 0 || loc.SegmentIndex != 3 {
		t.Fatalf("over-1 fraction should clamp+advance: got %+v", loc)
	}
}

func TestLinearLocationCompare(t *testing.T) {
	a := NewLinearLocation(1, 0.25)
	b := NewLinearLocation(1, 0.75)
	if a.Compare(b) != -1 {
		t.Fatalf("expected a<b")
	}
	if b.Compare(a) != 1 {
		t.Fatalf("expected b>a")
	}
	if a.Compare(a) != 0 {
		t.Fatalf("expected a==a")
	}
}

func TestLinearLocationGetCoordinate(t *testing.T) {
	ls := line100()
	loc := NewLinearLocation(0, 0.5)
	got := loc.GetCoordinate(ls)
	if got.X != 25 || got.Y != 0 {
		t.Fatalf("midpoint of first segment: got %+v", got)
	}
	end := EndLocation(ls)
	got = end.GetCoordinate(ls)
	if got.X != 100 || got.Y != 0 {
		t.Fatalf("end coord: got %+v", got)
	}
}

func TestLinearLocationClampOutOfRange(t *testing.T) {
	ls := line100()
	loc := LinearLocation{ComponentIndex: 5, SegmentIndex: 0, SegmentFraction: 0}
	loc.Clamp(ls)
	end := EndLocation(ls)
	if loc != end {
		t.Fatalf("clamp past-end -> end: got %+v want %+v", loc, end)
	}
}

func TestLinearLocationIsEndpoint(t *testing.T) {
	ls := line100()
	if !EndLocation(ls).IsEndpoint(ls) {
		t.Fatal("end location should be endpoint")
	}
	if NewLinearLocation(0, 0.5).IsEndpoint(ls) {
		t.Fatal("midpoint should not be endpoint")
	}
}

func TestLinearLocationToLowest(t *testing.T) {
	ls := line100()
	end := EndLocation(ls)
	low := end.ToLowest(ls)
	// total segments = 2, so lowest = (segIndex=1, frac=1.0)
	if low.SegmentIndex != 1 || low.SegmentFraction != 1 {
		t.Fatalf("toLowest: got %+v", low)
	}
}

// numComponents/multi sanity.
func TestLinearLocationOnMultiLine(t *testing.T) {
	a := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}})
	b := geom.NewLineString(nil, []geom.XY{{X: 100, Y: 0}, {X: 110, Y: 0}})
	mls := geom.NewMultiLineString(nil, a, b)
	loc := NewLinearLocationFull(1, 0, 0.5)
	got := loc.GetCoordinate(mls)
	if math.Abs(got.X-105) > 1e-9 || got.Y != 0 {
		t.Fatalf("midpoint of second component: got %+v", got)
	}
}
