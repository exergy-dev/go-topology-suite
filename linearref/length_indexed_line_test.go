package linearref

import (
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/stretchr/testify/assert"
)

func TestLengthIndexedExtractPoint(t *testing.T) {
	li := NewLengthIndexedLine(line100())
	got := li.ExtractPoint(50)
	assert.Equalf(t, 50.0, got.X, "midpoint: %+v", got)
	assert.Equalf(t, 0.0, got.Y, "midpoint: %+v", got)
	got = li.ExtractPoint(0)
	assert.Equalf(t, 0.0, got.X, "start: %+v", got)
	assert.Equalf(t, 0.0, got.Y, "start: %+v", got)
	got = li.ExtractPoint(100)
	assert.Equalf(t, 100.0, got.X, "end: %+v", got)
	assert.Equalf(t, 0.0, got.Y, "end: %+v", got)
}

func TestLengthIndexedExtractPointOutOfRange(t *testing.T) {
	li := NewLengthIndexedLine(line100())
	got := li.ExtractPoint(1000)
	assert.Equalf(t, 100.0, got.X, "over-range: %+v", got)
	// Negative index counts from the end.
	got = li.ExtractPoint(-25)
	assert.InDeltaf(t, 75.0, got.X, 1e-9, "-25 from end: %+v", got)
}

func TestLengthIndexedExtractLine(t *testing.T) {
	li := NewLengthIndexedLine(line100())
	sub := li.ExtractLine(25, 75).(*geom.LineString)
	first := sub.PointAt(0)
	assert.InDeltaf(t, 25.0, first.X, 1e-9, "first pt: %+v", first)
	last := sub.PointAt(sub.NumPoints() - 1)
	assert.InDeltaf(t, 75.0, last.X, 1e-9, "last pt: %+v", last)
}

func TestLengthIndexedRoundTrip(t *testing.T) {
	li := NewLengthIndexedLine(line100())
	for _, want := range []float64{0, 12.5, 25, 50, 75, 99} {
		p := li.ExtractPoint(want)
		got := li.IndexOf(p)
		assert.InDeltaf(t, want, got, 1e-9, "round-trip %v: got %v", want, got)
	}
}

func TestLengthIndexedProjectExternal(t *testing.T) {
	li := NewLengthIndexedLine(line100())
	got := li.Project(geom.XY{X: 30, Y: 25})
	assert.InDeltaf(t, 30.0, got, 1e-9, "project (30,25): got %v", got)
}

func TestLengthIndexedClampIndex(t *testing.T) {
	li := NewLengthIndexedLine(line100())
	got := li.ClampIndex(-10)
	assert.InDeltaf(t, 90.0, got, 1e-9, "negative clamp: %v", got)
	got = li.ClampIndex(1000)
	assert.Equalf(t, 100.0, got, "over clamp: %v", got)
	got = li.ClampIndex(50)
	assert.Equalf(t, 50.0, got, "in-range: %v", got)
}

func TestLengthIndexedIsValidIndex(t *testing.T) {
	li := NewLengthIndexedLine(line100())
	assert.True(t, li.IsValidIndex(0), "endpoints should be valid")
	assert.True(t, li.IsValidIndex(100), "endpoints should be valid")
	assert.False(t, li.IsValidIndex(-1), "out-of-range should not be valid")
	assert.False(t, li.IsValidIndex(101), "out-of-range should not be valid")
}

func TestLengthIndexedStartEnd(t *testing.T) {
	li := NewLengthIndexedLine(line100())
	assert.Equal(t, 0.0, li.StartIndex(), "start")
	assert.Equalf(t, 100.0, li.EndIndex(), "end: %v", li.EndIndex())
}
