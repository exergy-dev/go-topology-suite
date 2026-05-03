package linearref

import (
	"math"
	"testing"

	"github.com/terra-geo/terra/geom"
)

func TestLengthIndexedExtractPoint(t *testing.T) {
	li := NewLengthIndexedLine(line100())
	if got := li.ExtractPoint(50); got.X != 50 || got.Y != 0 {
		t.Fatalf("midpoint: %+v", got)
	}
	if got := li.ExtractPoint(0); got.X != 0 || got.Y != 0 {
		t.Fatalf("start: %+v", got)
	}
	if got := li.ExtractPoint(100); got.X != 100 || got.Y != 0 {
		t.Fatalf("end: %+v", got)
	}
}

func TestLengthIndexedExtractPointOutOfRange(t *testing.T) {
	li := NewLengthIndexedLine(line100())
	if got := li.ExtractPoint(1000); got.X != 100 {
		t.Fatalf("over-range: %+v", got)
	}
	// Negative index counts from the end.
	if got := li.ExtractPoint(-25); math.Abs(got.X-75) > 1e-9 {
		t.Fatalf("-25 from end: %+v", got)
	}
}

func TestLengthIndexedExtractLine(t *testing.T) {
	li := NewLengthIndexedLine(line100())
	sub := li.ExtractLine(25, 75).(*geom.LineString)
	if math.Abs(sub.PointAt(0).X-25) > 1e-9 {
		t.Fatalf("first pt: %+v", sub.PointAt(0))
	}
	last := sub.PointAt(sub.NumPoints() - 1)
	if math.Abs(last.X-75) > 1e-9 {
		t.Fatalf("last pt: %+v", last)
	}
}

func TestLengthIndexedRoundTrip(t *testing.T) {
	li := NewLengthIndexedLine(line100())
	for _, want := range []float64{0, 12.5, 25, 50, 75, 99} {
		p := li.ExtractPoint(want)
		got := li.IndexOf(p)
		if math.Abs(got-want) > 1e-9 {
			t.Fatalf("round-trip %v: got %v", want, got)
		}
	}
}

func TestLengthIndexedProjectExternal(t *testing.T) {
	li := NewLengthIndexedLine(line100())
	got := li.Project(geom.XY{X: 30, Y: 25})
	if math.Abs(got-30) > 1e-9 {
		t.Fatalf("project (30,25): got %v", got)
	}
}

func TestLengthIndexedClampIndex(t *testing.T) {
	li := NewLengthIndexedLine(line100())
	if got := li.ClampIndex(-10); math.Abs(got-90) > 1e-9 {
		t.Fatalf("negative clamp: %v", got)
	}
	if got := li.ClampIndex(1000); got != 100 {
		t.Fatalf("over clamp: %v", got)
	}
	if got := li.ClampIndex(50); got != 50 {
		t.Fatalf("in-range: %v", got)
	}
}

func TestLengthIndexedIsValidIndex(t *testing.T) {
	li := NewLengthIndexedLine(line100())
	if !li.IsValidIndex(0) || !li.IsValidIndex(100) {
		t.Fatal("endpoints should be valid")
	}
	if li.IsValidIndex(-1) || li.IsValidIndex(101) {
		t.Fatal("out-of-range should not be valid")
	}
}

func TestLengthIndexedStartEnd(t *testing.T) {
	li := NewLengthIndexedLine(line100())
	if li.StartIndex() != 0 {
		t.Fatal("start")
	}
	if li.EndIndex() != 100 {
		t.Fatalf("end: %v", li.EndIndex())
	}
}
