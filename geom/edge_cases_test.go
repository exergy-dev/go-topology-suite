package geom_test

import (
	"testing"

	"github.com/go-topology-suite/gts/geom"
)

// TestEnvelope_EdgeCases tests edge cases for envelope operations
func TestEnvelope_EdgeCases(t *testing.T) {
	t.Run("IsNull", func(t *testing.T) {
		env := geom.NewEnvelopeEmpty()
		if !env.IsNull() {
			t.Error("Empty envelope should be null")
		}

		env2 := geom.NewEnvelope(0, 0, 10, 10)
		if env2.IsNull() {
			t.Error("Non-empty envelope should not be null")
		}
	})

	t.Run("Width and Height", func(t *testing.T) {
		env := geom.NewEnvelope(0, 5, 10, 15)
		if env.Width() != 10 {
			t.Errorf("Expected width 10, got %f", env.Width())
		}
		if env.Height() != 10 {
			t.Errorf("Expected height 10, got %f", env.Height())
		}
	})

	t.Run("Centre", func(t *testing.T) {
		env := geom.NewEnvelope(0, 0, 10, 20)
		center := env.Centre()
		if center.X != 5 || center.Y != 10 {
			t.Errorf("Expected center (5, 10), got (%f, %f)", center.X, center.Y)
		}
	})

	t.Run("ExpandToInclude", func(t *testing.T) {
		env1 := geom.NewEnvelope(0, 0, 10, 10)
		env2 := geom.NewEnvelope(5, 5, 15, 15)

		env1.ExpandToInclude(env2)

		if env1.MinX != 0 || env1.MaxX != 15 || env1.MinY != 0 || env1.MaxY != 15 {
			t.Errorf("Expected (0,0,15,15), got (%f,%f,%f,%f)",
				env1.MinX, env1.MinY, env1.MaxX, env1.MaxY)
		}
	})

	t.Run("Clone", func(t *testing.T) {
		env := geom.NewEnvelope(0, 0, 10, 10)
		clone := env.Clone()

		clone.MinX = 100

		if env.MinX == 100 {
			t.Error("Clone should be independent of original")
		}
	})

	t.Run("Distance", func(t *testing.T) {
		env1 := geom.NewEnvelope(0, 0, 10, 10)
		env2 := geom.NewEnvelope(20, 20, 30, 30)
		dist := env1.Distance(env2)
		if dist == 0 {
			t.Error("Distance between non-intersecting envelopes should be > 0")
		}
	})
}

// TestPoint_EdgeCases tests edge cases for Point
func TestPoint_EdgeCases(t *testing.T) {
	t.Run("NumGeometries", func(t *testing.T) {
		p := geom.NewPoint(0, 0)
		if p.NumGeometries() != 1 {
			t.Errorf("Point should have 1 geometry, got %d", p.NumGeometries())
		}
	})

	t.Run("GeometryN", func(t *testing.T) {
		p := geom.NewPoint(0, 0)
		if p.GeometryN(0) != p {
			t.Error("GeometryN(0) should return the point itself")
		}
		if p.GeometryN(1) != nil {
			t.Error("GeometryN(1) should return nil")
		}
	})

	t.Run("Boundary", func(t *testing.T) {
		p := geom.NewPoint(0, 0)
		boundary := p.Boundary()
		if !boundary.IsEmpty() {
			t.Error("Point boundary should be empty")
		}
	})

	t.Run("Dimension", func(t *testing.T) {
		p := geom.NewPoint(0, 0)
		if p.Dimension() != geom.DimensionPoint {
			t.Errorf("Expected dimension %d, got %d", geom.DimensionPoint, p.Dimension())
		}
	})

	t.Run("IsSimple", func(t *testing.T) {
		p := geom.NewPoint(0, 0)
		if !p.IsSimple() {
			t.Error("Point should always be simple")
		}
	})

	t.Run("IsValid", func(t *testing.T) {
		p := geom.NewPoint(0, 0)
		if !p.IsValid() {
			t.Error("Point should always be valid")
		}
	})

	t.Run("Distance", func(t *testing.T) {
		p1 := geom.NewPoint(0, 0)
		p2 := geom.NewPoint(3, 4)
		dist := p1.Distance(p2)
		if dist != 5.0 {
			t.Errorf("Expected distance 5, got %f", dist)
		}
	})

	t.Run("EqualsExact with tolerance", func(t *testing.T) {
		p1 := geom.NewPoint(1.0, 2.0)
		p2 := geom.NewPoint(1.001, 2.001)

		if p1.EqualsExact(p2, 0.0001) {
			t.Error("Points should not be equal with tight tolerance")
		}

		if !p1.EqualsExact(p2, 0.01) {
			t.Error("Points should be equal with loose tolerance")
		}
	})

	t.Run("SRID operations", func(t *testing.T) {
		p := geom.NewPoint(0, 0)
		p.SetSRID(4326)

		if p.SRID() != 4326 {
			t.Errorf("Expected SRID 4326, got %d", p.SRID())
		}

		clone := p.Clone()
		if clone.SRID() != 4326 {
			t.Error("Clone should preserve SRID")
		}
	})
}

// TestLineString_EdgeCases tests edge cases for LineString
func TestLineString_EdgeCases(t *testing.T) {
	t.Run("Empty LineString", func(t *testing.T) {
		ls := geom.NewLineStringEmpty()
		if !ls.IsEmpty() {
			t.Error("Empty LineString should report IsEmpty")
		}
		if ls.Length() != 0 {
			t.Error("Empty LineString length should be 0")
		}
		if ls.NumPoints() != 0 {
			t.Error("Empty LineString should have 0 points")
		}
	})

	t.Run("Single point LineString", func(t *testing.T) {
		coords := geom.NewCoordinateSequenceXY(0, 0)
		ls := geom.NewLineString(coords)
		if ls.IsValid() {
			t.Error("Single-point LineString should be invalid")
		}
	})

	t.Run("IsSimple", func(t *testing.T) {
		simple := geom.NewLineStringXY(0, 0, 10, 10, 20, 20)
		if !simple.IsSimple() {
			t.Error("Non-self-intersecting line should be simple")
		}
	})

	t.Run("Dimension", func(t *testing.T) {
		ls := geom.NewLineStringXY(0, 0, 10, 10)
		if ls.Dimension() != geom.DimensionLine {
			t.Errorf("Expected dimension %d, got %d", geom.DimensionLine, ls.Dimension())
		}
	})

	t.Run("Boundary", func(t *testing.T) {
		// Open linestring has boundary
		open := geom.NewLineStringXY(0, 0, 10, 10)
		boundary := open.Boundary()
		if boundary.IsEmpty() {
			t.Error("Open linestring should have boundary")
		}

		// Closed linestring has no boundary
		closed := geom.NewLineStringXY(0, 0, 10, 0, 10, 10, 0, 0)
		closedBoundary := closed.Boundary()
		if !closedBoundary.IsEmpty() {
			t.Error("Closed linestring should have empty boundary")
		}
	})

	t.Run("NumGeometries", func(t *testing.T) {
		ls := geom.NewLineStringXY(0, 0, 10, 10)
		if ls.NumGeometries() != 1 {
			t.Errorf("LineString should have 1 geometry, got %d", ls.NumGeometries())
		}
	})

	t.Run("GeometryN", func(t *testing.T) {
		ls := geom.NewLineStringXY(0, 0, 10, 10)
		if ls.GeometryN(0) != ls {
			t.Error("GeometryN(0) should return the linestring itself")
		}
		if ls.GeometryN(1) != nil {
			t.Error("GeometryN(1) should return nil")
		}
	})

	t.Run("PointN out of bounds", func(t *testing.T) {
		ls := geom.NewLineStringXY(0, 0, 10, 10)
		if ls.PointN(-1) != nil {
			t.Error("PointN(-1) should return nil")
		}
		if ls.PointN(100) != nil {
			t.Error("PointN(100) should return nil")
		}
	})

	t.Run("Normalize", func(t *testing.T) {
		ls := geom.NewLineStringXY(10, 10, 0, 0)
		ls.Normalize()
		// After normalization, should be in canonical form
		if ls.NumPoints() != 2 {
			t.Error("Normalize should not change number of points")
		}
	})
}

// TestPolygon_EdgeCases tests edge cases for Polygon
func TestPolygon_EdgeCases(t *testing.T) {
	t.Run("Empty Polygon", func(t *testing.T) {
		poly := geom.NewPolygonEmpty()
		if !poly.IsEmpty() {
			t.Error("Empty polygon should report IsEmpty")
		}
		if poly.Area() != 0 {
			t.Error("Empty polygon area should be 0")
		}
		if poly.Perimeter() != 0 {
			t.Error("Empty polygon perimeter should be 0")
		}
	})

	t.Run("NumGeometries", func(t *testing.T) {
		poly := geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
		if poly.NumGeometries() != 1 {
			t.Errorf("Polygon should have 1 geometry, got %d", poly.NumGeometries())
		}
	})

	t.Run("GeometryN", func(t *testing.T) {
		poly := geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
		if poly.GeometryN(0) != poly {
			t.Error("GeometryN(0) should return the polygon itself")
		}
		if poly.GeometryN(1) != nil {
			t.Error("GeometryN(1) should return nil")
		}
	})

	t.Run("Dimension", func(t *testing.T) {
		poly := geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
		if poly.Dimension() != geom.DimensionArea {
			t.Errorf("Expected dimension %d, got %d", geom.DimensionArea, poly.Dimension())
		}
	})

	t.Run("IsSimple", func(t *testing.T) {
		poly := geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
		if !poly.IsSimple() {
			t.Error("Valid polygon should be simple")
		}
	})

	t.Run("Boundary", func(t *testing.T) {
		poly := geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
		boundary := poly.Boundary()
		if boundary.IsEmpty() {
			t.Error("Polygon boundary should not be empty")
		}
	})

	t.Run("ExteriorRing", func(t *testing.T) {
		ring := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		poly := geom.NewPolygon(ring, nil)
		exterior := poly.ExteriorRing()
		if exterior == nil {
			t.Error("ExteriorRing should not be nil")
		}
	})

	t.Run("NumInteriorRings", func(t *testing.T) {
		shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		hole := geom.NewLinearRingXY(2, 2, 8, 2, 8, 8, 2, 8, 2, 2)
		poly := geom.NewPolygon(shell, []*geom.LinearRing{hole})

		if poly.NumInteriorRings() != 1 {
			t.Errorf("Expected 1 hole, got %d", poly.NumInteriorRings())
		}
	})

	t.Run("InteriorRingN", func(t *testing.T) {
		shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		hole := geom.NewLinearRingXY(2, 2, 8, 2, 8, 8, 2, 8, 2, 2)
		poly := geom.NewPolygon(shell, []*geom.LinearRing{hole})

		if poly.InteriorRingN(-1) != nil {
			t.Error("InteriorRingN(-1) should return nil")
		}
		if poly.InteriorRingN(100) != nil {
			t.Error("InteriorRingN(100) should return nil")
		}

		interior := poly.InteriorRingN(0)
		if interior == nil {
			t.Error("InteriorRingN(0) should return the hole")
		}
	})

	t.Run("Normalize", func(t *testing.T) {
		poly := geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
		poly.Normalize()
		// After normalization, should be in canonical form
		if poly.IsEmpty() {
			t.Error("Normalize should not make polygon empty")
		}
	})
}

// TestLinearRing_EdgeCases tests edge cases for LinearRing
func TestLinearRing_EdgeCases(t *testing.T) {
	t.Run("Auto-close behavior", func(t *testing.T) {
		coords := geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10, 0, 10)
		lr := geom.NewLinearRing(coords)
		if !lr.IsClosed() {
			t.Error("LinearRing should auto-close")
		}
	})

	t.Run("Empty LinearRing", func(t *testing.T) {
		lr := geom.NewLinearRingEmpty()
		if !lr.IsEmpty() {
			t.Error("Empty LinearRing should report IsEmpty")
		}
	})

	t.Run("Too few points", func(t *testing.T) {
		coords := geom.NewCoordinateSequenceXY(0, 0, 10, 0)
		lr := geom.NewLinearRing(coords)
		if lr.IsValid() {
			t.Error("LinearRing with < 4 points should be invalid")
		}
	})

	t.Run("SignedArea", func(t *testing.T) {
		// CCW ring has positive signed area
		ccwCoords := geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		if geom.SignedArea(ccwCoords) <= 0 {
			t.Error("CCW ring should have positive signed area")
		}

		// CW ring has negative signed area
		cwCoords := geom.NewCoordinateSequenceXY(0, 0, 0, 10, 10, 10, 10, 0, 0, 0)
		if geom.SignedArea(cwCoords) >= 0 {
			t.Error("CW ring should have negative signed area")
		}
	})

	t.Run("Reverse", func(t *testing.T) {
		lr := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		wasIsCCW := lr.IsCCW()

		reversed := lr.Reverse()

		if reversed.IsCCW() == wasIsCCW {
			t.Error("Reverse should change orientation")
		}
	})
}

// TestCoordinate_EdgeCases tests edge cases for Coordinate
func TestCoordinate_EdgeCases(t *testing.T) {
	t.Run("Clone", func(t *testing.T) {
		c := geom.NewCoordinate(1, 2)
		clone := c.Clone()

		clone.X = 999

		if c.X == 999 {
			t.Error("Clone should be independent of original")
		}
	})

	t.Run("3D coordinate", func(t *testing.T) {
		c := geom.NewCoordinateZ(1, 2, 3)
		if c.Z == nil || *c.Z != 3 {
			t.Error("3D coordinate should have Z value")
		}

		clone := c.Clone()
		if clone.Z == nil || *clone.Z != 3 {
			t.Error("Cloned 3D coordinate should preserve Z")
		}
	})

	t.Run("Distance to self", func(t *testing.T) {
		c := geom.NewCoordinate(5, 5)
		if c.Distance(c) != 0 {
			t.Error("Distance to self should be 0")
		}
	})

	t.Run("Equals2D with same coordinates", func(t *testing.T) {
		c1 := geom.NewCoordinate(1, 2)
		c2 := geom.NewCoordinate(1, 2)

		if !c1.Equals2D(c2, 0.0001) {
			t.Error("Identical coordinates should be equal")
		}
	})
}

// TestCoordinateSequence_EdgeCases tests edge cases for CoordinateSequence
func TestCoordinateSequence_EdgeCases(t *testing.T) {
	t.Run("Clone", func(t *testing.T) {
		seq := geom.NewCoordinateSequenceXY(0, 0, 10, 10)
		clone := seq.Clone()

		clone[0].X = 999

		if seq[0].X == 999 {
			t.Error("Clone should be independent of original")
		}
	})

	t.Run("First and Last", func(t *testing.T) {
		seq := geom.NewCoordinateSequenceXY(0, 0, 10, 10, 20, 20)

		first := seq.First()
		if first.X != 0 || first.Y != 0 {
			t.Error("First should return first coordinate")
		}

		last := seq.Last()
		if last.X != 20 || last.Y != 20 {
			t.Error("Last should return last coordinate")
		}
	})

	t.Run("IsClosed", func(t *testing.T) {
		closed := geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10, 0, 0)
		if !closed.IsClosed(0.0001) {
			t.Error("Closed sequence should report IsClosed")
		}

		open := geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10)
		if open.IsClosed(0.0001) {
			t.Error("Open sequence should not report IsClosed")
		}
	})
}

// TestMultiGeometry_EmptyComponents tests multi-geometries with empty components
func TestMultiGeometry_EmptyComponents(t *testing.T) {
	t.Run("MultiPoint with empty point", func(t *testing.T) {
		mp := geom.NewMultiPoint([]*geom.Point{
			geom.NewPoint(0, 0),
			geom.NewPointEmpty(),
			geom.NewPoint(10, 10),
		})

		if mp.NumGeometries() != 3 {
			t.Errorf("Expected 3 geometries, got %d", mp.NumGeometries())
		}
	})

	t.Run("MultiLineString with empty line", func(t *testing.T) {
		mls := geom.NewMultiLineString([]*geom.LineString{
			geom.NewLineStringXY(0, 0, 10, 0),
			geom.NewLineStringEmpty(),
		})

		if mls.NumGeometries() != 2 {
			t.Errorf("Expected 2 geometries, got %d", mls.NumGeometries())
		}
	})

	t.Run("MultiPolygon with empty polygon", func(t *testing.T) {
		mp := geom.NewMultiPolygon([]*geom.Polygon{
			geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
			geom.NewPolygonEmpty(),
		})

		if mp.NumGeometries() != 2 {
			t.Errorf("Expected 2 geometries, got %d", mp.NumGeometries())
		}
	})
}

// TestPrecisionModel_EdgeCases tests edge cases for precision models
func TestPrecisionModel_EdgeCases(t *testing.T) {
	t.Run("Fixed precision with 0 scale", func(t *testing.T) {
		pm := geom.NewFixedPrecision(1)
		val := pm.MakePreciseValue(123.456)
		if val != 123 {
			t.Errorf("Expected 123, got %f", val)
		}
	})

	t.Run("Fixed precision with large scale", func(t *testing.T) {
		pm := geom.NewFixedPrecision(100000)
		val := pm.MakePreciseValue(1.234567)
		// Should round to 5 decimal places
		expected := 1.23457
		if val < expected-0.000001 || val > expected+0.000001 {
			t.Errorf("Expected ~%f, got %f", expected, val)
		}
	})

	t.Run("MakePrecise coordinate", func(t *testing.T) {
		pm := geom.NewFixedPrecision(10)
		coord := geom.NewCoordinate(1.23456, 2.34567)
		pm.MakePrecise(&coord)

		if coord.X < 1.1 || coord.X > 1.3 {
			t.Errorf("Expected X ~1.2, got %f", coord.X)
		}
	})

	t.Run("MakePrecise 3D coordinate", func(t *testing.T) {
		pm := geom.NewFixedPrecision(10)
		coord := geom.NewCoordinateZ(1.23456, 2.34567, 3.45678)
		pm.MakePrecise(&coord)

		if coord.Z != nil && (*coord.Z < 3.3 || *coord.Z > 3.5) {
			t.Errorf("Expected Z ~3.4, got %f", *coord.Z)
		}
	})
}
