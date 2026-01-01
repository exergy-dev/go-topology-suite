package buffer

import (
	"math"
	"testing"
	"testing/quick"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

// getArea returns the area of a geometry.
func getArea(g geom.Geometry) float64 {
	if g == nil || g.IsEmpty() {
		return 0
	}
	switch v := g.(type) {
	case *geom.Polygon:
		return v.Area()
	case *geom.MultiPolygon:
		return v.Area()
	default:
		return 0
	}
}

// generatePolygon creates a simple polygon from random values.
func generatePolygon(cx, cy, size float64) *geom.Polygon {
	if size < 1 {
		size = 1
	}
	if size > 1000 {
		size = 1000
	}
	// Create a square centered at (cx, cy)
	half := size / 2
	shell := geom.NewLinearRingXY(
		cx-half, cy-half,
		cx+half, cy-half,
		cx+half, cy+half,
		cx-half, cy+half,
		cx-half, cy-half,
	)
	return geom.NewPolygon(shell, nil)
}

// normalizeCoord bounds a coordinate to a reasonable range.
func normalizeCoord(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	// Bound to reasonable range
	if v > 10000 {
		return 10000
	}
	if v < -10000 {
		return -10000
	}
	return v
}

// TestBufferZeroDistancePreservesGeometry tests that buffering with distance 0
// returns a geometry equivalent to the original.
func TestBufferZeroDistancePreservesGeometry(t *testing.T) {
	f := func(cx, cy, size float64) bool {
		cx = normalizeCoord(cx)
		cy = normalizeCoord(cy)
		size = math.Abs(normalizeCoord(size))
		if size < 1 {
			size = 1
		}

		poly := generatePolygon(cx, cy, size)
		result := Buffer(poly, 0)

		// Area should be unchanged
		originalArea := poly.Area()
		resultArea := getArea(result)

		return math.Abs(originalArea-resultArea) < geom.DefaultEpsilon
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// TestBufferPositiveIncreasesArea tests that buffering a polygon with positive
// distance increases or maintains the area.
func TestBufferPositiveIncreasesArea(t *testing.T) {
	f := func(cx, cy, size, dist float64) bool {
		cx = normalizeCoord(cx)
		cy = normalizeCoord(cy)
		size = math.Abs(normalizeCoord(size))
		dist = math.Abs(normalizeCoord(dist))

		if size < 1 {
			size = 1
		}
		if size > 100 {
			size = 100
		}
		if dist < 0.1 {
			dist = 0.1
		}
		if dist > 10 {
			dist = 10
		}

		poly := generatePolygon(cx, cy, size)
		result := Buffer(poly, dist)

		originalArea := poly.Area()
		resultArea := getArea(result)

		// Buffered area should be greater than or equal to original
		return resultArea >= originalArea-geom.DefaultEpsilon
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// TestBufferNegativeDecreasesArea tests that buffering a polygon with negative
// distance decreases the area.
func TestBufferNegativeDecreasesArea(t *testing.T) {
	f := func(cx, cy float64) bool {
		cx = normalizeCoord(cx)
		cy = normalizeCoord(cy)

		// Use fixed size and small negative buffer to ensure polygon doesn't collapse
		size := 20.0
		dist := -1.0

		poly := generatePolygon(cx, cy, size)
		result := Buffer(poly, dist)

		originalArea := poly.Area()
		resultArea := getArea(result)

		// Eroded area should be less than or equal to original
		return resultArea <= originalArea+geom.DefaultEpsilon
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// TestBufferPointCreatesCircle tests that buffering a point creates a
// polygon with area approximately pi*r^2.
func TestBufferPointCreatesCircle(t *testing.T) {
	f := func(x, y, r float64) bool {
		x = normalizeCoord(x)
		y = normalizeCoord(y)
		r = math.Abs(normalizeCoord(r))

		if r < 0.1 {
			r = 0.1
		}
		if r > 100 {
			r = 100
		}

		p := geom.NewPoint(x, y)
		result := Buffer(p, r)

		expectedArea := math.Pi * r * r
		actualArea := getArea(result)

		// 2% tolerance (JTS-compatible)
		tolerance := expectedArea * 0.02
		return math.Abs(actualArea-expectedArea) < tolerance
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// TestBufferLineAreaApproximation tests that buffering a line segment
// creates a polygon with area approximately 2*r*length + pi*r^2 (for round caps).
func TestBufferLineAreaApproximation(t *testing.T) {
	f := func(x1, y1, x2, y2, r float64) bool {
		x1 = normalizeCoord(x1)
		y1 = normalizeCoord(y1)
		x2 = normalizeCoord(x2)
		y2 = normalizeCoord(y2)
		r = math.Abs(normalizeCoord(r))

		if r < 0.1 {
			r = 0.1
		}
		if r > 10 {
			r = 10
		}

		// Ensure line has some length
		dx := x2 - x1
		dy := y2 - y1
		length := math.Sqrt(dx*dx + dy*dy)
		if length < 1 {
			x2 = x1 + 10
			y2 = y1
			length = 10
		}
		if length > 100 {
			return true // Skip very long lines
		}

		ls := geom.NewLineStringXY(x1, y1, x2, y2)
		result := Buffer(ls, r)

		// Expected area: rectangle (2*r*length) + two semicircles (pi*r^2)
		expectedArea := 2*r*length + math.Pi*r*r
		actualArea := getArea(result)

		// JTS-compatible tolerance: 2% for default quality (8 quadrant segments)
		tolerance := expectedArea * 0.02
		return math.Abs(actualArea-expectedArea) < tolerance
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 50}); err != nil {
		t.Error(err)
	}
}

// TestBufferSymmetry tests that buffering produces symmetric results.
func TestBufferSymmetry(t *testing.T) {
	// Buffer a square and check that the result is also approximately square
	poly := generatePolygon(0, 0, 10)
	result := Buffer(poly, 2)

	env := result.Envelope()
	width := env.MaxX - env.MinX
	height := env.MaxY - env.MinY

	// Width and height should be approximately equal for a buffered square
	ratio := width / height
	if math.Abs(ratio-1) > 0.01 {
		t.Errorf("Buffered square not symmetric: width=%f, height=%f, ratio=%f", width, height, ratio)
	}
}
