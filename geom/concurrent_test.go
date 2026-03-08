package geom_test

import (
	"sync"
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

// TestConcurrentEnvelopeAccess verifies that concurrent Envelope() calls
// on a shared geometry do not race. Run with -race to verify.
func TestConcurrentEnvelopeAccess(t *testing.T) {
	// Point
	t.Run("Point", func(t *testing.T) {
		p := geom.NewPoint(5, 10)
		concurrentEnvelope(t, p, 100)
	})

	// LineString
	t.Run("LineString", func(t *testing.T) {
		ls := geom.NewLineString(geom.CoordinateSequence{
			geom.NewCoordinate(0, 0),
			geom.NewCoordinate(10, 10),
			geom.NewCoordinate(20, 0),
		})
		concurrentEnvelope(t, ls, 100)
	})

	// Polygon
	t.Run("Polygon", func(t *testing.T) {
		shell := geom.NewLinearRing(geom.CoordinateSequence{
			geom.NewCoordinate(0, 0),
			geom.NewCoordinate(10, 0),
			geom.NewCoordinate(10, 10),
			geom.NewCoordinate(0, 10),
			geom.NewCoordinate(0, 0),
		})
		poly := geom.NewPolygon(shell, nil)
		concurrentEnvelope(t, poly, 100)
	})

	// MultiPoint
	t.Run("MultiPoint", func(t *testing.T) {
		mp := geom.NewMultiPoint([]*geom.Point{
			geom.NewPoint(0, 0),
			geom.NewPoint(10, 10),
			geom.NewPoint(20, 20),
		})
		concurrentEnvelope(t, mp, 100)
	})

	// MultiLineString
	t.Run("MultiLineString", func(t *testing.T) {
		mls := geom.NewMultiLineString([]*geom.LineString{
			geom.NewLineString(geom.CoordinateSequence{
				geom.NewCoordinate(0, 0),
				geom.NewCoordinate(10, 10),
			}),
			geom.NewLineString(geom.CoordinateSequence{
				geom.NewCoordinate(5, 5),
				geom.NewCoordinate(15, 15),
			}),
		})
		concurrentEnvelope(t, mls, 100)
	})

	// MultiPolygon
	t.Run("MultiPolygon", func(t *testing.T) {
		shell := geom.NewLinearRing(geom.CoordinateSequence{
			geom.NewCoordinate(0, 0),
			geom.NewCoordinate(10, 0),
			geom.NewCoordinate(10, 10),
			geom.NewCoordinate(0, 10),
			geom.NewCoordinate(0, 0),
		})
		mpoly := geom.NewMultiPolygon([]*geom.Polygon{
			geom.NewPolygon(shell, nil),
		})
		concurrentEnvelope(t, mpoly, 100)
	})

	// GeometryCollection
	t.Run("GeometryCollection", func(t *testing.T) {
		gc := geom.NewGeometryCollection([]geom.Geometry{
			geom.NewPoint(0, 0),
			geom.NewPoint(100, 100),
		})
		concurrentEnvelope(t, gc, 100)
	})
}

func concurrentEnvelope(t *testing.T, g geom.Geometry, goroutines int) {
	t.Helper()
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			env := g.Envelope()
			if env == nil {
				t.Error("Envelope returned nil")
			}
		}()
	}
	wg.Wait()
}

// TestConcurrentReadMethods verifies that various read methods can be
// called concurrently on a shared geometry.
func TestConcurrentReadMethods(t *testing.T) {
	shell := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(0, 0),
	})
	poly := geom.NewPolygon(shell, nil)

	var wg sync.WaitGroup
	n := 100
	wg.Add(n * 5)
	for i := 0; i < n; i++ {
		go func() { defer wg.Done(); _ = poly.Envelope() }()
		go func() { defer wg.Done(); _ = poly.IsEmpty() }()
		go func() { defer wg.Done(); _ = poly.Coordinates() }()
		go func() { defer wg.Done(); _ = poly.String() }()
		go func() { defer wg.Done(); _ = poly.Area() }()
	}
	wg.Wait()
}
