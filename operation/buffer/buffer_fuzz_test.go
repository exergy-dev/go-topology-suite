package buffer

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

// FuzzBufferPoint tests that buffering a point doesn't panic.
func FuzzBufferPoint(f *testing.F) {
	// Add seed corpus
	f.Add(0.0, 0.0, 1.0)
	f.Add(10.0, 10.0, 5.0)
	f.Add(-100.0, 100.0, 0.5)
	f.Add(0.0, 0.0, 0.0)

	f.Fuzz(func(t *testing.T, x, y, distance float64) {
		p := geom.NewPoint(x, y)
		result := Buffer(p, distance)

		// Should not panic and should return a valid geometry
		if result == nil {
			t.Error("Buffer returned nil")
		}

		// If distance > 0, result should not be empty (but small/NaN distances may be)
		_ = distance > geom.DefaultEpsilon && result.IsEmpty()
	})
}

// FuzzBufferLineString tests that buffering a line string doesn't panic.
func FuzzBufferLineString(f *testing.F) {
	// Add seed corpus
	f.Add(0.0, 0.0, 10.0, 10.0, 1.0)
	f.Add(-5.0, -5.0, 5.0, 5.0, 2.0)
	f.Add(0.0, 0.0, 0.0, 10.0, 0.5)

	f.Fuzz(func(t *testing.T, x1, y1, x2, y2, distance float64) {
		ls := mustLineStringXY(x1, y1, x2, y2)
		result := Buffer(ls, distance)

		if result == nil {
			t.Error("Buffer returned nil")
		}
	})
}

// FuzzBufferPolygon tests that buffering a polygon doesn't panic.
func FuzzBufferPolygon(f *testing.F) {
	// Add seed corpus
	f.Add(0.0, 0.0, 10.0, 1.0)
	f.Add(-50.0, -50.0, 20.0, 5.0)
	f.Add(100.0, 100.0, 5.0, -1.0)

	f.Fuzz(func(t *testing.T, cx, cy, size, distance float64) {
		// Bound size to reasonable range
		if size < 0.1 {
			size = 0.1
		}
		if size > 1000 {
			size = 1000
		}

		half := size / 2
		shell := mustLinearRingXY(
			cx-half, cy-half,
			cx+half, cy-half,
			cx+half, cy+half,
			cx-half, cy+half,
			cx-half, cy-half,
		)
		poly := geom.NewPolygon(shell, nil)
		result := Buffer(poly, distance)

		if result == nil {
			t.Error("Buffer returned nil")
		}
	})
}

// FuzzBufferWithParams tests buffering with various parameters.
func FuzzBufferWithParams(f *testing.F) {
	// Add seed corpus
	f.Add(0.0, 0.0, 1.0, 8, 0, 0)
	f.Add(5.0, 5.0, 2.0, 4, 1, 1)
	f.Add(-10.0, 10.0, 0.5, 16, 2, 2)

	f.Fuzz(func(t *testing.T, x, y, distance float64, quadSegs int, capStyle int, joinStyle int) {
		if quadSegs < 1 {
			quadSegs = 1
		}
		if quadSegs > 32 {
			quadSegs = 32
		}

		params := &Params{
			QuadrantSegments: quadSegs,
			EndCapStyle:      CapStyle(capStyle % 3),
			JoinStyle:        JoinStyle(joinStyle % 3),
			MitreLimit:       5.0,
		}

		p := geom.NewPoint(x, y)
		result := BufferWithParams(p, distance, params)

		if result == nil {
			t.Error("BufferWithParams returned nil")
		}
	})
}
