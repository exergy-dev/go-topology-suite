package algorithm_test

import (
	"math"
	"testing"

	"github.com/go-topology-suite/gts/algorithm"
	"github.com/go-topology-suite/gts/geom"
)

func TestOrientationIndex(t *testing.T) {
	t.Run("CounterClockwise", func(t *testing.T) {
		p1 := geom.NewCoordinate(0, 0)
		p2 := geom.NewCoordinate(10, 0)
		p3 := geom.NewCoordinate(5, 5)

		orientation := algorithm.OrientationIndex(p1, p2, p3)
		if orientation != algorithm.CounterClockwise {
			t.Errorf("Expected CounterClockwise (1), got %d", orientation)
		}
	})

	t.Run("Clockwise", func(t *testing.T) {
		p1 := geom.NewCoordinate(0, 0)
		p2 := geom.NewCoordinate(10, 0)
		p3 := geom.NewCoordinate(5, -5)

		orientation := algorithm.OrientationIndex(p1, p2, p3)
		if orientation != algorithm.Clockwise {
			t.Errorf("Expected Clockwise (-1), got %d", orientation)
		}
	})

	t.Run("Collinear", func(t *testing.T) {
		p1 := geom.NewCoordinate(0, 0)
		p2 := geom.NewCoordinate(10, 0)
		p3 := geom.NewCoordinate(5, 0)

		orientation := algorithm.OrientationIndex(p1, p2, p3)
		if orientation != algorithm.Collinear {
			t.Errorf("Expected Collinear (0), got %d", orientation)
		}
	})
}

func TestAngle(t *testing.T) {
	t.Run("Angle", func(t *testing.T) {
		p1 := geom.NewCoordinate(0, 0)
		p2 := geom.NewCoordinate(1, 0) // Angle should be 0
		angle := algorithm.Angle(p1, p2)
		if math.Abs(angle) > 0.001 {
			t.Errorf("Expected angle ~0, got %v", angle)
		}

		p3 := geom.NewCoordinate(0, 1) // Angle should be Pi/2
		angle = algorithm.Angle(p1, p3)
		if math.Abs(angle-math.Pi/2) > 0.001 {
			t.Errorf("Expected angle ~Pi/2, got %v", angle)
		}
	})

	t.Run("ToDegrees", func(t *testing.T) {
		deg := algorithm.ToDegrees(math.Pi)
		if math.Abs(deg-180) > 0.001 {
			t.Errorf("Expected 180 degrees, got %v", deg)
		}
	})

	t.Run("ToRadians", func(t *testing.T) {
		rad := algorithm.ToRadians(180)
		if math.Abs(rad-math.Pi) > 0.001 {
			t.Errorf("Expected Pi radians, got %v", rad)
		}
	})
}

func TestArea(t *testing.T) {
	t.Run("PolygonArea", func(t *testing.T) {
		shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		poly := geom.NewPolygon(shell, nil)
		area := algorithm.Area(poly)
		if area != 100 {
			t.Errorf("Expected area 100, got %v", area)
		}
	})

	t.Run("RingArea", func(t *testing.T) {
		coords := geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		area := algorithm.RingArea(coords)
		if area != 100 {
			t.Errorf("Expected area 100, got %v", area)
		}
	})

	t.Run("SignedArea", func(t *testing.T) {
		// Counter-clockwise should be positive
		ccwCoords := geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		ccwArea := algorithm.SignedArea(ccwCoords)
		if ccwArea <= 0 {
			t.Errorf("Expected positive signed area for CCW, got %v", ccwArea)
		}

		// Clockwise should be negative
		cwCoords := geom.NewCoordinateSequenceXY(0, 0, 0, 10, 10, 10, 10, 0, 0, 0)
		cwArea := algorithm.SignedArea(cwCoords)
		if cwArea >= 0 {
			t.Errorf("Expected negative signed area for CW, got %v", cwArea)
		}
	})

	t.Run("LineLength", func(t *testing.T) {
		coords := geom.NewCoordinateSequenceXY(0, 0, 3, 4)
		length := algorithm.LineLength(coords)
		if length != 5 {
			t.Errorf("Expected length 5, got %v", length)
		}
	})
}

func TestCentroid(t *testing.T) {
	t.Run("PointCentroid", func(t *testing.T) {
		p := geom.NewPoint(5, 10)
		centroid := algorithm.Centroid(p)
		if centroid.X != 5 || centroid.Y != 10 {
			t.Errorf("Expected (5, 10), got (%v, %v)", centroid.X, centroid.Y)
		}
	})

	t.Run("LineCentroid", func(t *testing.T) {
		coords := geom.NewCoordinateSequenceXY(0, 0, 10, 0)
		centroid := algorithm.LineCentroid(coords)
		if centroid.X != 5 || centroid.Y != 0 {
			t.Errorf("Expected (5, 0), got (%v, %v)", centroid.X, centroid.Y)
		}
	})

	t.Run("PolygonCentroid", func(t *testing.T) {
		shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		poly := geom.NewPolygon(shell, nil)
		centroid := algorithm.Centroid(poly)
		if math.Abs(centroid.X-5) > 0.001 || math.Abs(centroid.Y-5) > 0.001 {
			t.Errorf("Expected (5, 5), got (%v, %v)", centroid.X, centroid.Y)
		}
	})
}

func TestDistance(t *testing.T) {
	t.Run("PointToPoint", func(t *testing.T) {
		p1 := geom.NewCoordinate(0, 0)
		p2 := geom.NewCoordinate(3, 4)
		dist := algorithm.DistancePointToPoint(p1, p2)
		if dist != 5 {
			t.Errorf("Expected distance 5, got %v", dist)
		}
	})

	t.Run("PointToSegment", func(t *testing.T) {
		p := geom.NewCoordinate(5, 5)
		a := geom.NewCoordinate(0, 0)
		b := geom.NewCoordinate(10, 0)
		dist := algorithm.DistancePointToSegment(p, a, b)
		if dist != 5 {
			t.Errorf("Expected distance 5, got %v", dist)
		}
	})

	t.Run("PointToSegmentAtEndpoint", func(t *testing.T) {
		p := geom.NewCoordinate(-5, 0)
		a := geom.NewCoordinate(0, 0)
		b := geom.NewCoordinate(10, 0)
		dist := algorithm.DistancePointToSegment(p, a, b)
		if dist != 5 {
			t.Errorf("Expected distance 5, got %v", dist)
		}
	})

	t.Run("GeometryDistance", func(t *testing.T) {
		pt := geom.NewPoint(5, 5)
		ls := geom.NewLineStringXY(0, 0, 10, 0)
		dist := algorithm.Distance(pt, ls)
		if dist != 5 {
			t.Errorf("Expected distance 5, got %v", dist)
		}
	})
}

func TestPointLocation(t *testing.T) {
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly := geom.NewPolygon(shell, nil)

	t.Run("Inside", func(t *testing.T) {
		loc := algorithm.PointLocation(geom.NewCoordinate(5, 5), poly)
		if loc != geom.LocationInterior {
			t.Errorf("Expected Interior, got %v", loc)
		}
	})

	t.Run("Outside", func(t *testing.T) {
		loc := algorithm.PointLocation(geom.NewCoordinate(15, 5), poly)
		if loc != geom.LocationExterior {
			t.Errorf("Expected Exterior, got %v", loc)
		}
	})

	t.Run("OnBoundary", func(t *testing.T) {
		loc := algorithm.PointLocation(geom.NewCoordinate(0, 5), poly)
		if loc != geom.LocationBoundary {
			t.Errorf("Expected Boundary, got %v", loc)
		}
	})
}

func TestIsPointInRing(t *testing.T) {
	ring := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)

	t.Run("Inside", func(t *testing.T) {
		if !algorithm.IsPointInRing(geom.NewCoordinate(5, 5), ring) {
			t.Error("Expected point to be inside ring")
		}
	})

	t.Run("Outside", func(t *testing.T) {
		if algorithm.IsPointInRing(geom.NewCoordinate(15, 5), ring) {
			t.Error("Expected point to be outside ring")
		}
	})
}

func TestSegmentsIntersect(t *testing.T) {
	t.Run("Intersecting", func(t *testing.T) {
		a1 := geom.NewCoordinate(0, 0)
		a2 := geom.NewCoordinate(10, 10)
		b1 := geom.NewCoordinate(0, 10)
		b2 := geom.NewCoordinate(10, 0)

		if !algorithm.SegmentsIntersect(a1, a2, b1, b2) {
			t.Error("Expected segments to intersect")
		}
	})

	t.Run("NotIntersecting", func(t *testing.T) {
		a1 := geom.NewCoordinate(0, 0)
		a2 := geom.NewCoordinate(5, 0)
		b1 := geom.NewCoordinate(10, 0)
		b2 := geom.NewCoordinate(15, 0)

		if algorithm.SegmentsIntersect(a1, a2, b1, b2) {
			t.Error("Expected segments to not intersect")
		}
	})
}
