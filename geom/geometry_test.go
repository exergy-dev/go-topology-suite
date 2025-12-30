package geom_test

import (
	"testing"

	"github.com/go-topology-suite/gts/geom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCoordinate(t *testing.T) {
	t.Run("NewCoordinate", func(t *testing.T) {
		c := geom.NewCoordinate(1.5, 2.5)
		assert.Equal(t, 1.5, c.X)
		assert.Equal(t, 2.5, c.Y)
		assert.Nil(t, c.Z)
		assert.Nil(t, c.M)
	})

	t.Run("NewCoordinateZ", func(t *testing.T) {
		c := geom.NewCoordinateZ(1.0, 2.0, 3.0)
		assert.Equal(t, 1.0, c.X)
		assert.Equal(t, 2.0, c.Y)
		require.NotNil(t, c.Z)
		assert.Equal(t, 3.0, *c.Z)
	})

	t.Run("Distance", func(t *testing.T) {
		c1 := geom.NewCoordinate(0, 0)
		c2 := geom.NewCoordinate(3, 4)
		dist := c1.Distance(c2)
		assert.Equal(t, 5.0, dist)
	})

	t.Run("Equals2D", func(t *testing.T) {
		c1 := geom.NewCoordinate(1.0, 2.0)
		c2 := geom.NewCoordinate(1.0, 2.0)
		c3 := geom.NewCoordinate(1.1, 2.0)

		assert.True(t, c1.Equals2D(c2, 0.001), "Expected c1 and c2 to be equal")
		assert.False(t, c1.Equals2D(c3, 0.001), "Expected c1 and c3 to be not equal")
	})
}

func TestEnvelope(t *testing.T) {
	t.Run("NewEnvelope", func(t *testing.T) {
		env := geom.NewEnvelope(0, 0, 10, 10)
		assert.Equal(t, 0.0, env.MinX)
		assert.Equal(t, 0.0, env.MinY)
		assert.Equal(t, 10.0, env.MaxX)
		assert.Equal(t, 10.0, env.MaxY)
	})

	t.Run("Contains", func(t *testing.T) {
		env := geom.NewEnvelope(0, 0, 10, 10)
		assert.True(t, env.Contains(geom.NewCoordinate(5, 5)), "Expected envelope to contain (5, 5)")
		assert.False(t, env.Contains(geom.NewCoordinate(15, 5)), "Expected envelope to not contain (15, 5)")
	})

	t.Run("Intersects", func(t *testing.T) {
		env1 := geom.NewEnvelope(0, 0, 10, 10)
		env2 := geom.NewEnvelope(5, 5, 15, 15)
		env3 := geom.NewEnvelope(20, 20, 30, 30)

		assert.True(t, env1.Intersects(env2), "Expected env1 and env2 to intersect")
		assert.False(t, env1.Intersects(env3), "Expected env1 and env3 to not intersect")
	})

	t.Run("Area", func(t *testing.T) {
		env := geom.NewEnvelope(0, 0, 10, 10)
		assert.Equal(t, 100.0, env.Area())
	})
}

func TestPoint(t *testing.T) {
	t.Run("Creation", func(t *testing.T) {
		p := geom.NewPoint(1.0, 2.0)
		assert.Equal(t, 1.0, p.X())
		assert.Equal(t, 2.0, p.Y())
	})

	t.Run("GeometryType", func(t *testing.T) {
		p := geom.NewPoint(0, 0)
		assert.Equal(t, "Point", p.GeometryType())
	})

	t.Run("Empty", func(t *testing.T) {
		p := geom.NewPointEmpty()
		assert.True(t, p.IsEmpty(), "Expected point to be empty")
	})

	t.Run("String/WKT", func(t *testing.T) {
		p := geom.NewPoint(1, 2)
		wkt := p.String()
		assert.Equal(t, "POINT (1 2)", wkt)
	})

	t.Run("Envelope", func(t *testing.T) {
		p := geom.NewPoint(5, 10)
		env := p.Envelope()
		assert.Equal(t, 5.0, env.MinX)
		assert.Equal(t, 10.0, env.MinY)
		assert.Equal(t, 5.0, env.MaxX)
		assert.Equal(t, 10.0, env.MaxY)
	})
}

func TestLineString(t *testing.T) {
	coords := geom.NewCoordinateSequenceXY(0, 0, 10, 10, 20, 0)

	t.Run("Creation", func(t *testing.T) {
		ls := geom.NewLineString(coords)
		assert.Equal(t, 3, ls.NumPoints())
	})

	t.Run("GeometryType", func(t *testing.T) {
		ls := geom.NewLineString(coords)
		assert.Equal(t, "LineString", ls.GeometryType())
	})

	t.Run("Length", func(t *testing.T) {
		ls := geom.NewLineStringXY(0, 0, 3, 4)
		assert.Equal(t, 5.0, ls.Length())
	})

	t.Run("IsClosed", func(t *testing.T) {
		open := geom.NewLineStringXY(0, 0, 10, 0, 10, 10)
		closed := geom.NewLineStringXY(0, 0, 10, 0, 10, 10, 0, 0)

		assert.False(t, open.IsClosed(), "Expected open linestring to not be closed")
		assert.True(t, closed.IsClosed(), "Expected closed linestring to be closed")
	})

	t.Run("Envelope", func(t *testing.T) {
		ls := geom.NewLineString(coords)
		env := ls.Envelope()
		assert.Equal(t, 0.0, env.MinX)
		assert.Equal(t, 20.0, env.MaxX)
		assert.Equal(t, 0.0, env.MinY)
		assert.Equal(t, 10.0, env.MaxY)
	})
}

func TestPolygon(t *testing.T) {
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)

	t.Run("Creation", func(t *testing.T) {
		p := geom.NewPolygon(shell, nil)
		assert.False(t, p.IsEmpty(), "Expected polygon to not be empty")
	})

	t.Run("GeometryType", func(t *testing.T) {
		p := geom.NewPolygon(shell, nil)
		assert.Equal(t, "Polygon", p.GeometryType())
	})

	t.Run("Area", func(t *testing.T) {
		p := geom.NewPolygon(shell, nil)
		assert.Equal(t, 100.0, p.Area())
	})

	t.Run("WithHole", func(t *testing.T) {
		hole := geom.NewLinearRingXY(2, 2, 8, 2, 8, 8, 2, 8, 2, 2)
		p := geom.NewPolygon(shell, []*geom.LinearRing{hole})
		expected := 100.0 - 36.0 // 10x10 - 6x6
		assert.InDelta(t, expected, p.Area(), 0.001)
	})

	t.Run("ContainsPoint", func(t *testing.T) {
		p := geom.NewPolygon(shell, nil)
		assert.True(t, p.ContainsPoint(geom.NewCoordinate(5, 5)), "Expected polygon to contain (5, 5)")
		assert.False(t, p.ContainsPoint(geom.NewCoordinate(15, 5)), "Expected polygon to not contain (15, 5)")
	})

	t.Run("Centroid_Simple", func(t *testing.T) {
		p := geom.NewPolygon(shell, nil)
		centroid := p.Centroid()
		require.False(t, centroid.IsEmpty())
		assert.InDelta(t, 5.0, centroid.X(), 0.001)
		assert.InDelta(t, 5.0, centroid.Y(), 0.001)
	})

	t.Run("Centroid_WithSymmetricHole", func(t *testing.T) {
		// A symmetric hole centered at (5,5) should not change the centroid
		hole := geom.NewLinearRingXY(3, 3, 7, 3, 7, 7, 3, 7, 3, 3)
		p := geom.NewPolygon(shell, []*geom.LinearRing{hole})
		centroid := p.Centroid()
		require.False(t, centroid.IsEmpty())
		assert.InDelta(t, 5.0, centroid.X(), 0.001)
		assert.InDelta(t, 5.0, centroid.Y(), 0.001)
	})

	t.Run("Centroid_WithAsymmetricHole", func(t *testing.T) {
		// Shell: 20x10 rectangle from (0,0) to (20,10), centroid at (10, 5)
		// Hole: 4x4 square from (2,3) to (6,7), centroid at (4, 5)
		asymShell := geom.NewLinearRingXY(0, 0, 20, 0, 20, 10, 0, 10, 0, 0)
		hole := geom.NewLinearRingXY(2, 3, 6, 3, 6, 7, 2, 7, 2, 3)
		p := geom.NewPolygon(asymShell, []*geom.LinearRing{hole})

		// Shell area: 20*10 = 200, shell centroid: (10, 5)
		// Hole area: 4*4 = 16, hole centroid: (4, 5)
		// Weighted centroid X: (10*200 - 4*16) / (200-16) = (2000-64)/184 = 10.52...
		// Weighted centroid Y: (5*200 - 5*16) / (200-16) = (1000-80)/184 = 5.0
		centroid := p.Centroid()
		require.False(t, centroid.IsEmpty())
		expectedX := (10.0*200.0 - 4.0*16.0) / (200.0 - 16.0)
		assert.InDelta(t, expectedX, centroid.X(), 0.01)
		assert.InDelta(t, 5.0, centroid.Y(), 0.01)
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
		assert.Equal(t, 3, mp.NumGeometries(), "Expected 3 points")
	})

	t.Run("GeometryType", func(t *testing.T) {
		mp := geom.NewMultiPoint(points)
		assert.Equal(t, "MultiPoint", mp.GeometryType())
	})
}

func TestGeometryCollection(t *testing.T) {
	geoms := []geom.Geometry{
		geom.NewPoint(0, 0),
		geom.NewLineStringXY(0, 0, 10, 10),
	}

	t.Run("Creation", func(t *testing.T) {
		gc := geom.NewGeometryCollection(geoms)
		assert.Equal(t, 2, gc.NumGeometries(), "Expected 2 geometries")
	})

	t.Run("GeometryType", func(t *testing.T) {
		gc := geom.NewGeometryCollection(geoms)
		assert.Equal(t, "GeometryCollection", gc.GeometryType())
	})
}

func TestGeometryFactory(t *testing.T) {
	factory := geom.NewGeometryFactoryDefault()

	t.Run("CreatePoint", func(t *testing.T) {
		p := factory.CreatePoint(1, 2)
		assert.Equal(t, 1.0, p.X())
		assert.Equal(t, 2.0, p.Y())
	})

	t.Run("CreateLineString", func(t *testing.T) {
		ls := factory.CreateLineStringXY(0, 0, 10, 10)
		assert.Equal(t, 2, ls.NumPoints())
	})

	t.Run("WithSRID", func(t *testing.T) {
		factoryWithSRID := geom.NewGeometryFactoryWithSRID(4326)
		p := factoryWithSRID.CreatePoint(1, 2)
		assert.Equal(t, 4326, p.SRID())
	})
}

func TestPrecisionModel(t *testing.T) {
	t.Run("FloatingPrecision", func(t *testing.T) {
		pm := geom.NewFloatingPrecision()
		val := pm.MakePreciseValue(1.23456789)
		assert.Equal(t, 1.23456789, val)
	})

	t.Run("FixedPrecision", func(t *testing.T) {
		pm := geom.NewFixedPrecision(1000) // 3 decimal places
		val := pm.MakePreciseValue(1.23456789)
		expected := 1.235 // Rounded
		assert.InDelta(t, expected, val, 0.0001)
	})

	t.Run("SinglePrecision", func(t *testing.T) {
		pm := geom.NewFloatingSinglePrecision()
		val := pm.MakePreciseValue(1.23456789)
		// Single precision may differ from double precision
		assert.NotEqual(t, 0.0, val, "Single precision should return a value")
	})
}

func TestLinearRing(t *testing.T) {
	t.Run("AutoClose", func(t *testing.T) {
		// Ring without explicit closure
		coords := geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10, 0, 10)
		lr := geom.NewLinearRing(coords)
		// Should be auto-closed
		assert.True(t, lr.IsClosed(), "Expected ring to be closed")
	})

	t.Run("IsCCW", func(t *testing.T) {
		// Counter-clockwise ring
		ccwCoords := geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		ccwRing := geom.NewLinearRing(ccwCoords)
		assert.True(t, ccwRing.IsCCW(), "Expected ring to be counter-clockwise")

		// Clockwise ring
		cwCoords := geom.NewCoordinateSequenceXY(0, 0, 0, 10, 10, 10, 10, 0, 0, 0)
		cwRing := geom.NewLinearRing(cwCoords)
		assert.True(t, cwRing.IsCW(), "Expected ring to be clockwise")
	})

	t.Run("Area", func(t *testing.T) {
		coords := geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		lr := geom.NewLinearRing(coords)
		area := lr.Area()
		assert.Equal(t, 100.0, area)
	})
}
