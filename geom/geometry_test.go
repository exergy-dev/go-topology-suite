package geom_test

import (
	"math"
	"testing"

	"github.com/go-topology-suite/gts/geom"
)

func TestCoordinate(t *testing.T) {
	t.Run("NewCoordinate", func(t *testing.T) {
		c := geom.NewCoordinate(1.5, 2.5)
		if c.X != 1.5 || c.Y != 2.5 {
			t.Errorf("Expected (1.5, 2.5), got (%v, %v)", c.X, c.Y)
		}
		if c.Z != nil || c.M != nil {
			t.Error("Expected Z and M to be nil")
		}
	})

	t.Run("NewCoordinateZ", func(t *testing.T) {
		c := geom.NewCoordinateZ(1.0, 2.0, 3.0)
		if c.X != 1.0 || c.Y != 2.0 || *c.Z != 3.0 {
			t.Errorf("Expected (1, 2, 3), got (%v, %v, %v)", c.X, c.Y, *c.Z)
		}
	})

	t.Run("Distance", func(t *testing.T) {
		c1 := geom.NewCoordinate(0, 0)
		c2 := geom.NewCoordinate(3, 4)
		dist := c1.Distance(c2)
		if dist != 5.0 {
			t.Errorf("Expected distance 5, got %v", dist)
		}
	})

	t.Run("Equals2D", func(t *testing.T) {
		c1 := geom.NewCoordinate(1.0, 2.0)
		c2 := geom.NewCoordinate(1.0, 2.0)
		c3 := geom.NewCoordinate(1.1, 2.0)

		if !c1.Equals2D(c2, 0.001) {
			t.Error("Expected c1 and c2 to be equal")
		}
		if c1.Equals2D(c3, 0.001) {
			t.Error("Expected c1 and c3 to be not equal")
		}
	})
}

func TestEnvelope(t *testing.T) {
	t.Run("NewEnvelope", func(t *testing.T) {
		env := geom.NewEnvelope(0, 0, 10, 10)
		if env.MinX != 0 || env.MinY != 0 || env.MaxX != 10 || env.MaxY != 10 {
			t.Errorf("Unexpected envelope bounds: %+v", env)
		}
	})

	t.Run("Contains", func(t *testing.T) {
		env := geom.NewEnvelope(0, 0, 10, 10)
		if !env.Contains(geom.NewCoordinate(5, 5)) {
			t.Error("Expected envelope to contain (5, 5)")
		}
		if env.Contains(geom.NewCoordinate(15, 5)) {
			t.Error("Expected envelope to not contain (15, 5)")
		}
	})

	t.Run("Intersects", func(t *testing.T) {
		env1 := geom.NewEnvelope(0, 0, 10, 10)
		env2 := geom.NewEnvelope(5, 5, 15, 15)
		env3 := geom.NewEnvelope(20, 20, 30, 30)

		if !env1.Intersects(env2) {
			t.Error("Expected env1 and env2 to intersect")
		}
		if env1.Intersects(env3) {
			t.Error("Expected env1 and env3 to not intersect")
		}
	})

	t.Run("Area", func(t *testing.T) {
		env := geom.NewEnvelope(0, 0, 10, 10)
		if env.Area() != 100 {
			t.Errorf("Expected area 100, got %v", env.Area())
		}
	})
}

func TestPoint(t *testing.T) {
	t.Run("Creation", func(t *testing.T) {
		p := geom.NewPoint(1.0, 2.0)
		if p.X() != 1.0 || p.Y() != 2.0 {
			t.Errorf("Expected (1, 2), got (%v, %v)", p.X(), p.Y())
		}
	})

	t.Run("GeometryType", func(t *testing.T) {
		p := geom.NewPoint(0, 0)
		if p.GeometryType() != "Point" {
			t.Errorf("Expected 'Point', got '%s'", p.GeometryType())
		}
	})

	t.Run("Empty", func(t *testing.T) {
		p := geom.NewPointEmpty()
		if !p.IsEmpty() {
			t.Error("Expected point to be empty")
		}
	})

	t.Run("String/WKT", func(t *testing.T) {
		p := geom.NewPoint(1, 2)
		wkt := p.String()
		if wkt != "POINT (1 2)" {
			t.Errorf("Expected 'POINT (1 2)', got '%s'", wkt)
		}
	})

	t.Run("Envelope", func(t *testing.T) {
		p := geom.NewPoint(5, 10)
		env := p.Envelope()
		if env.MinX != 5 || env.MinY != 10 || env.MaxX != 5 || env.MaxY != 10 {
			t.Errorf("Unexpected envelope: %+v", env)
		}
	})
}

func TestLineString(t *testing.T) {
	coords := geom.NewCoordinateSequenceXY(0, 0, 10, 10, 20, 0)

	t.Run("Creation", func(t *testing.T) {
		ls := geom.NewLineString(coords)
		if ls.NumPoints() != 3 {
			t.Errorf("Expected 3 points, got %d", ls.NumPoints())
		}
	})

	t.Run("GeometryType", func(t *testing.T) {
		ls := geom.NewLineString(coords)
		if ls.GeometryType() != "LineString" {
			t.Errorf("Expected 'LineString', got '%s'", ls.GeometryType())
		}
	})

	t.Run("Length", func(t *testing.T) {
		ls := geom.NewLineStringXY(0, 0, 3, 4)
		if ls.Length() != 5.0 {
			t.Errorf("Expected length 5, got %v", ls.Length())
		}
	})

	t.Run("IsClosed", func(t *testing.T) {
		open := geom.NewLineStringXY(0, 0, 10, 0, 10, 10)
		closed := geom.NewLineStringXY(0, 0, 10, 0, 10, 10, 0, 0)

		if open.IsClosed() {
			t.Error("Expected open linestring to not be closed")
		}
		if !closed.IsClosed() {
			t.Error("Expected closed linestring to be closed")
		}
	})

	t.Run("Envelope", func(t *testing.T) {
		ls := geom.NewLineString(coords)
		env := ls.Envelope()
		if env.MinX != 0 || env.MaxX != 20 || env.MinY != 0 || env.MaxY != 10 {
			t.Errorf("Unexpected envelope: %+v", env)
		}
	})
}

func TestPolygon(t *testing.T) {
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)

	t.Run("Creation", func(t *testing.T) {
		p := geom.NewPolygon(shell, nil)
		if p.IsEmpty() {
			t.Error("Expected polygon to not be empty")
		}
	})

	t.Run("GeometryType", func(t *testing.T) {
		p := geom.NewPolygon(shell, nil)
		if p.GeometryType() != "Polygon" {
			t.Errorf("Expected 'Polygon', got '%s'", p.GeometryType())
		}
	})

	t.Run("Area", func(t *testing.T) {
		p := geom.NewPolygon(shell, nil)
		area := p.Area()
		if area != 100 {
			t.Errorf("Expected area 100, got %v", area)
		}
	})

	t.Run("WithHole", func(t *testing.T) {
		hole := geom.NewLinearRingXY(2, 2, 8, 2, 8, 8, 2, 8, 2, 2)
		p := geom.NewPolygon(shell, []*geom.LinearRing{hole})
		area := p.Area()
		expected := 100.0 - 36.0 // 10x10 - 6x6
		if math.Abs(area-expected) > 0.001 {
			t.Errorf("Expected area %v, got %v", expected, area)
		}
	})

	t.Run("ContainsPoint", func(t *testing.T) {
		p := geom.NewPolygon(shell, nil)
		if !p.ContainsPoint(geom.NewCoordinate(5, 5)) {
			t.Error("Expected polygon to contain (5, 5)")
		}
		if p.ContainsPoint(geom.NewCoordinate(15, 5)) {
			t.Error("Expected polygon to not contain (15, 5)")
		}
	})
}

func TestMultiPoint(t *testing.T) {
	points := []*geom.Point{
		geom.NewPoint(0, 0),
		geom.NewPoint(1, 1),
		geom.NewPoint(2, 2),
	}

	t.Run("Creation", func(t *testing.T) {
		mp := geom.NewMultiPoint(points)
		if mp.NumGeometries() != 3 {
			t.Errorf("Expected 3 points, got %d", mp.NumGeometries())
		}
	})

	t.Run("GeometryType", func(t *testing.T) {
		mp := geom.NewMultiPoint(points)
		if mp.GeometryType() != "MultiPoint" {
			t.Errorf("Expected 'MultiPoint', got '%s'", mp.GeometryType())
		}
	})
}

func TestGeometryCollection(t *testing.T) {
	geoms := []geom.Geometry{
		geom.NewPoint(0, 0),
		geom.NewLineStringXY(0, 0, 10, 10),
	}

	t.Run("Creation", func(t *testing.T) {
		gc := geom.NewGeometryCollection(geoms)
		if gc.NumGeometries() != 2 {
			t.Errorf("Expected 2 geometries, got %d", gc.NumGeometries())
		}
	})

	t.Run("GeometryType", func(t *testing.T) {
		gc := geom.NewGeometryCollection(geoms)
		if gc.GeometryType() != "GeometryCollection" {
			t.Errorf("Expected 'GeometryCollection', got '%s'", gc.GeometryType())
		}
	})
}

func TestGeometryFactory(t *testing.T) {
	factory := geom.NewGeometryFactoryDefault()

	t.Run("CreatePoint", func(t *testing.T) {
		p := factory.CreatePoint(1, 2)
		if p.X() != 1 || p.Y() != 2 {
			t.Errorf("Expected (1, 2), got (%v, %v)", p.X(), p.Y())
		}
	})

	t.Run("CreateLineString", func(t *testing.T) {
		ls := factory.CreateLineStringXY(0, 0, 10, 10)
		if ls.NumPoints() != 2 {
			t.Errorf("Expected 2 points, got %d", ls.NumPoints())
		}
	})

	t.Run("WithSRID", func(t *testing.T) {
		factoryWithSRID := geom.NewGeometryFactoryWithSRID(4326)
		p := factoryWithSRID.CreatePoint(1, 2)
		if p.SRID() != 4326 {
			t.Errorf("Expected SRID 4326, got %d", p.SRID())
		}
	})
}

func TestPrecisionModel(t *testing.T) {
	t.Run("FloatingPrecision", func(t *testing.T) {
		pm := geom.NewFloatingPrecision()
		val := pm.MakePreciseValue(1.23456789)
		if val != 1.23456789 {
			t.Errorf("Expected 1.23456789, got %v", val)
		}
	})

	t.Run("FixedPrecision", func(t *testing.T) {
		pm := geom.NewFixedPrecision(1000) // 3 decimal places
		val := pm.MakePreciseValue(1.23456789)
		expected := 1.235 // Rounded
		if math.Abs(val-expected) > 0.0001 {
			t.Errorf("Expected %v, got %v", expected, val)
		}
	})

	t.Run("SinglePrecision", func(t *testing.T) {
		pm := geom.NewFloatingSinglePrecision()
		val := pm.MakePreciseValue(1.23456789)
		if val == 1.23456789 {
			t.Log("Single precision may differ from double precision")
		}
	})
}

func TestLinearRing(t *testing.T) {
	t.Run("AutoClose", func(t *testing.T) {
		// Ring without explicit closure
		coords := geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10, 0, 10)
		lr := geom.NewLinearRing(coords)
		// Should be auto-closed
		if !lr.IsClosed() {
			t.Error("Expected ring to be closed")
		}
	})

	t.Run("IsCCW", func(t *testing.T) {
		// Counter-clockwise ring
		ccwCoords := geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		ccwRing := geom.NewLinearRing(ccwCoords)
		if !ccwRing.IsCCW() {
			t.Error("Expected ring to be counter-clockwise")
		}

		// Clockwise ring
		cwCoords := geom.NewCoordinateSequenceXY(0, 0, 0, 10, 10, 10, 10, 0, 0, 0)
		cwRing := geom.NewLinearRing(cwCoords)
		if !cwRing.IsCW() {
			t.Error("Expected ring to be clockwise")
		}
	})

	t.Run("Area", func(t *testing.T) {
		coords := geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		lr := geom.NewLinearRing(coords)
		area := lr.Area()
		if area != 100 {
			t.Errorf("Expected area 100, got %v", area)
		}
	})
}
