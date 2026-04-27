package topology

import (
	"math"
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

func TestNodeLineSetsLabelsCrossingSegments(t *testing.T) {
	segments := NodeLineSets(
		[]*geom.LineString{mustLineStringXY(0, 0, 10, 0)},
		[]*geom.LineString{mustLineStringXY(5, -5, 5, 5)},
	)

	if len(segments) != 4 {
		t.Fatalf("expected 4 noded segments, got %d", len(segments))
	}
	for _, segment := range segments {
		if segment.InA() == segment.InB() {
			t.Fatalf("crossing segment should belong to exactly one input: %+v", segment)
		}
	}
}

func TestNodeLineSetsLabelsOverlappingSegments(t *testing.T) {
	segments := NodeLineSets(
		[]*geom.LineString{mustLineStringXY(0, 0, 10, 0)},
		[]*geom.LineString{mustLineStringXY(5, 0, 15, 0)},
	)

	if len(segments) != 3 {
		t.Fatalf("expected 3 noded segments, got %d", len(segments))
	}

	var both int
	for _, segment := range segments {
		if segment.InA() && segment.InB() {
			both++
		}
	}
	if both != 1 {
		t.Fatalf("expected exactly one overlapping segment, got %d", both)
	}
}

func TestNodeLineSetsWithPrecisionSnapsWithoutMutatingInputs(t *testing.T) {
	a := mustLineStringXY(0.04, 0.04, 10.04, 0.04)
	b := mustLineStringXY(5.04, -1.04, 5.04, 1.04)

	segments := NodeLineSetsWithPrecision(
		[]*geom.LineString{a},
		[]*geom.LineString{b},
		geom.NewFixedPrecision(1),
	)

	if !hasEndpoint(segments, geom.NewCoordinate(5, 0)) {
		t.Fatalf("expected snapped noding endpoint at (5,0), got %#v", segments)
	}
	if got := a.Coordinates()[0]; got.X != 0.04 || got.Y != 0.04 {
		t.Fatalf("precision noding mutated input coordinate: %v", got)
	}
}

func hasEndpoint(segments []NodedLineSegment, coord geom.Coordinate) bool {
	for _, segment := range segments {
		if segment.Start.Equals2D(coord, geom.DefaultEpsilon) ||
			segment.End.Equals2D(coord, geom.DefaultEpsilon) {
			return true
		}
	}
	return false
}

func mustLineStringXY(values ...float64) *geom.LineString {
	seq, err := geom.NewCoordinateSequenceXY(values...)
	if err != nil {
		panic(err)
	}
	return geom.NewLineString(seq)
}

func FuzzNodeLineSets(f *testing.F) {
	f.Add(0.0, 0.0, 10.0, 0.0, 5.0, -5.0, 5.0, 5.0)
	f.Add(0.0, 0.0, 10.0, 0.0, 5.0, 0.0, 15.0, 0.0)
	f.Add(-10.0, -10.0, -5.0, -5.0, 1.0, 1.0, 2.0, 2.0)

	f.Fuzz(func(t *testing.T, ax0, ay0, ax1, ay1, bx0, by0, bx1, by1 float64) {
		values := []float64{ax0, ay0, ax1, ay1, bx0, by0, bx1, by1}
		for _, value := range values {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				t.Skip()
			}
		}
		a := mustLineStringXY(
			clampFuzzCoord(ax0), clampFuzzCoord(ay0),
			clampFuzzCoord(ax1), clampFuzzCoord(ay1),
		)
		b := mustLineStringXY(
			clampFuzzCoord(bx0), clampFuzzCoord(by0),
			clampFuzzCoord(bx1), clampFuzzCoord(by1),
		)

		segments := NodeLineSets([]*geom.LineString{a}, []*geom.LineString{b})
		for _, segment := range segments {
			if segment.Start.Distance(segment.End) <= geom.DefaultEpsilon {
				t.Fatalf("noded line segment is degenerate: %+v", segment)
			}
			if !segment.InA() && !segment.InB() {
				t.Fatalf("noded line segment is unlabeled: %+v", segment)
			}
		}
	})
}

func clampFuzzCoord(value float64) float64 {
	const limit = 1e6
	if value > limit {
		return limit
	}
	if value < -limit {
		return -limit
	}
	return value
}

func BenchmarkNodeLineSetsGrid(b *testing.B) {
	const size = 25
	linesA := make([]*geom.LineString, 0, size)
	linesB := make([]*geom.LineString, 0, size)
	for i := 0; i < size; i++ {
		offset := float64(i)
		linesA = append(linesA, mustLineStringXY(0, offset, size-1, offset))
		linesB = append(linesB, mustLineStringXY(offset, 0, offset, size-1))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		segments := NodeLineSets(linesA, linesB)
		if len(segments) == 0 {
			b.Fatal("expected noded segments")
		}
	}
}
