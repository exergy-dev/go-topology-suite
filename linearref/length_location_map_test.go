package linearref

import (
	"math"
	"testing"

	"github.com/terra-geo/terra/geom"
)

func TestGetLocationMidpoint(t *testing.T) {
	ls := line100()
	loc := GetLocation(ls, 50)
	// length 50 falls exactly at the inner vertex (segment 1, frac 0).
	if loc.ComponentIndex != 0 || loc.SegmentIndex != 1 || loc.SegmentFraction != 0 {
		t.Fatalf("midpoint location: got %+v", loc)
	}
	got := loc.GetCoordinate(ls)
	if got.X != 50 || got.Y != 0 {
		t.Fatalf("midpoint coord: got %+v", got)
	}
}

func TestGetLocationFraction(t *testing.T) {
	ls := line100()
	// Length 25 -> midpoint of first segment.
	loc := GetLocation(ls, 25)
	got := loc.GetCoordinate(ls)
	if got.X != 25 || got.Y != 0 {
		t.Fatalf("quarter point: got %+v", got)
	}
}

func TestGetLocationOutOfRange(t *testing.T) {
	ls := line100()
	loc := GetLocation(ls, 1000)
	if loc.GetCoordinate(ls).X != 100 {
		t.Fatalf("over-range -> end")
	}
	loc = GetLocation(ls, -10)
	got := loc.GetCoordinate(ls)
	if math.Abs(got.X-90) > 1e-9 {
		t.Fatalf("negative measured from end: got %+v", got)
	}
}

func TestGetLengthRoundTrip(t *testing.T) {
	ls := line100()
	for _, want := range []float64{0, 12.5, 25, 50, 75, 99.9} {
		loc := GetLocation(ls, want)
		got := GetLength(ls, loc)
		if math.Abs(got-want) > 1e-9 {
			t.Fatalf("round-trip %v: got %v", want, got)
		}
	}
}

func TestGetLocationOnMulti(t *testing.T) {
	a := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 50, Y: 0}})
	b := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 100}, {X: 50, Y: 100}})
	mls := geom.NewMultiLineString(nil, a, b)
	// total length = 100; index 75 is 25 into the second component.
	loc := GetLocation(mls, 75)
	if loc.ComponentIndex != 1 {
		t.Fatalf("expected component 1, got %+v", loc)
	}
	got := loc.GetCoordinate(mls)
	if got.X != 25 || got.Y != 100 {
		t.Fatalf("coord: got %+v", got)
	}
}
