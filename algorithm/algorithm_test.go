package algorithm_test

import (
	"math"
	"testing"

	"github.com/robert-malhotra/go-topology-suite/algorithm"
	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/stretchr/testify/assert"
)

func TestOrientationIndex(t *testing.T) {
	t.Run("CounterClockwise", func(t *testing.T) {
		p1 := geom.NewCoordinate(0, 0)
		p2 := geom.NewCoordinate(10, 0)
		p3 := geom.NewCoordinate(5, 5)

		orientation := algorithm.OrientationIndex(p1, p2, p3)
		assert.Equal(t, algorithm.CounterClockwise, orientation, "Expected CounterClockwise")
	})

	t.Run("Clockwise", func(t *testing.T) {
		p1 := geom.NewCoordinate(0, 0)
		p2 := geom.NewCoordinate(10, 0)
		p3 := geom.NewCoordinate(5, -5)

		orientation := algorithm.OrientationIndex(p1, p2, p3)
		assert.Equal(t, algorithm.Clockwise, orientation, "Expected Clockwise")
	})

	t.Run("Collinear", func(t *testing.T) {
		p1 := geom.NewCoordinate(0, 0)
		p2 := geom.NewCoordinate(10, 0)
		p3 := geom.NewCoordinate(5, 0)

		orientation := algorithm.OrientationIndex(p1, p2, p3)
		assert.Equal(t, algorithm.Collinear, orientation, "Expected Collinear")
	})
}

func TestAngle(t *testing.T) {
	t.Run("Angle", func(t *testing.T) {
		p1 := geom.NewCoordinate(0, 0)
		p2 := geom.NewCoordinate(1, 0) // Angle should be 0
		angle := algorithm.Angle(p1, p2)
		assert.InDelta(t, 0.0, angle, 0.001, "Expected angle ~0")

		p3 := geom.NewCoordinate(0, 1) // Angle should be Pi/2
		angle = algorithm.Angle(p1, p3)
		assert.InDelta(t, math.Pi/2, angle, 0.001, "Expected angle ~Pi/2")
	})

	t.Run("ToDegrees", func(t *testing.T) {
		deg := algorithm.ToDegrees(math.Pi)
		assert.InDelta(t, 180.0, deg, 0.001, "Expected 180 degrees")
	})

	t.Run("ToRadians", func(t *testing.T) {
		rad := algorithm.ToRadians(180)
		assert.InDelta(t, math.Pi, rad, 0.001, "Expected Pi radians")
	})
}

func TestArea(t *testing.T) {
	t.Run("PolygonArea", func(t *testing.T) {
		shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		poly := geom.NewPolygon(shell, nil)
		area := algorithm.Area(poly)
		assert.Equal(t, 100.0, area)
	})

	t.Run("RingArea", func(t *testing.T) {
		coords := geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		area := algorithm.RingArea(coords)
		assert.Equal(t, 100.0, area)
	})

	t.Run("SignedArea", func(t *testing.T) {
		// Counter-clockwise should be positive
		ccwCoords := geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		ccwArea := algorithm.SignedArea(ccwCoords)
		assert.Greater(t, ccwArea, 0.0, "Expected positive signed area for CCW")

		// Clockwise should be negative
		cwCoords := geom.NewCoordinateSequenceXY(0, 0, 0, 10, 10, 10, 10, 0, 0, 0)
		cwArea := algorithm.SignedArea(cwCoords)
		assert.Less(t, cwArea, 0.0, "Expected negative signed area for CW")
	})

	t.Run("LineLength", func(t *testing.T) {
		coords := geom.NewCoordinateSequenceXY(0, 0, 3, 4)
		length := algorithm.LineLength(coords)
		assert.Equal(t, 5.0, length)
	})
}

func TestCentroid(t *testing.T) {
	t.Run("PointCentroid", func(t *testing.T) {
		p := geom.NewPoint(5, 10)
		centroid := algorithm.Centroid(p)
		assert.Equal(t, 5.0, centroid.X)
		assert.Equal(t, 10.0, centroid.Y)
	})

	t.Run("LineCentroid", func(t *testing.T) {
		coords := geom.NewCoordinateSequenceXY(0, 0, 10, 0)
		centroid := algorithm.LineCentroid(coords)
		assert.Equal(t, 5.0, centroid.X)
		assert.Equal(t, 0.0, centroid.Y)
	})

	t.Run("PolygonCentroid", func(t *testing.T) {
		shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		poly := geom.NewPolygon(shell, nil)
		centroid := algorithm.Centroid(poly)
		assert.InDelta(t, 5.0, centroid.X, 0.001)
		assert.InDelta(t, 5.0, centroid.Y, 0.001)
	})
}

func TestDistance(t *testing.T) {
	t.Run("PointToPoint", func(t *testing.T) {
		p1 := geom.NewCoordinate(0, 0)
		p2 := geom.NewCoordinate(3, 4)
		dist := algorithm.DistancePointToPoint(p1, p2)
		assert.Equal(t, 5.0, dist)
	})

	t.Run("PointToSegment", func(t *testing.T) {
		p := geom.NewCoordinate(5, 5)
		a := geom.NewCoordinate(0, 0)
		b := geom.NewCoordinate(10, 0)
		dist := algorithm.DistancePointToSegment(p, a, b)
		assert.Equal(t, 5.0, dist)
	})

	t.Run("PointToSegmentAtEndpoint", func(t *testing.T) {
		p := geom.NewCoordinate(-5, 0)
		a := geom.NewCoordinate(0, 0)
		b := geom.NewCoordinate(10, 0)
		dist := algorithm.DistancePointToSegment(p, a, b)
		assert.Equal(t, 5.0, dist)
	})

	t.Run("GeometryDistance", func(t *testing.T) {
		pt := geom.NewPoint(5, 5)
		ls := geom.NewLineStringXY(0, 0, 10, 0)
		dist := algorithm.Distance(pt, ls)
		assert.Equal(t, 5.0, dist)
	})
}

func TestPointLocation(t *testing.T) {
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly := geom.NewPolygon(shell, nil)

	t.Run("Inside", func(t *testing.T) {
		loc := algorithm.PointLocation(geom.NewCoordinate(5, 5), poly)
		assert.Equal(t, geom.LocationInterior, loc, "Expected Interior")
	})

	t.Run("Outside", func(t *testing.T) {
		loc := algorithm.PointLocation(geom.NewCoordinate(15, 5), poly)
		assert.Equal(t, geom.LocationExterior, loc, "Expected Exterior")
	})

	t.Run("OnBoundary", func(t *testing.T) {
		loc := algorithm.PointLocation(geom.NewCoordinate(0, 5), poly)
		assert.Equal(t, geom.LocationBoundary, loc, "Expected Boundary")
	})
}

func TestIsPointInRing(t *testing.T) {
	ring := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)

	t.Run("Inside", func(t *testing.T) {
		assert.True(t, algorithm.IsPointInRing(geom.NewCoordinate(5, 5), ring), "Expected point to be inside ring")
	})

	t.Run("Outside", func(t *testing.T) {
		assert.False(t, algorithm.IsPointInRing(geom.NewCoordinate(15, 5), ring), "Expected point to be outside ring")
	})
}

func TestSegmentsIntersect(t *testing.T) {
	t.Run("Intersecting", func(t *testing.T) {
		a1 := geom.NewCoordinate(0, 0)
		a2 := geom.NewCoordinate(10, 10)
		b1 := geom.NewCoordinate(0, 10)
		b2 := geom.NewCoordinate(10, 0)

		assert.True(t, algorithm.SegmentsIntersect(a1, a2, b1, b2), "Expected segments to intersect")
	})

	t.Run("NotIntersecting", func(t *testing.T) {
		a1 := geom.NewCoordinate(0, 0)
		a2 := geom.NewCoordinate(5, 0)
		b1 := geom.NewCoordinate(10, 0)
		b2 := geom.NewCoordinate(15, 0)

		assert.False(t, algorithm.SegmentsIntersect(a1, a2, b1, b2), "Expected segments to not intersect")
	})
}
