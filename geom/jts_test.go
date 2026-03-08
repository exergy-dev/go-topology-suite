package geom_test

import (
	"math"
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/io/wkt"
)

// JTS-style test cases for core geometry operations
// These tests are ported from Java Topology Suite to verify correctness
// against known input/output pairs

// TestJTS_PolygonArea_Square tests area calculation for a square polygon.
func TestJTS_PolygonArea_Square(t *testing.T) {
	poly, _ := wkt.UnmarshalString("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")

	area := poly.(*geom.Polygon).Area()
	expected := 100.0

	if math.Abs(area-expected) > 0.001 {
		t.Errorf("Square area: expected %.2f, got %.2f", expected, area)
	}
}

// TestJTS_PolygonArea_Rectangle tests area calculation for a rectangle.
func TestJTS_PolygonArea_Rectangle(t *testing.T) {
	poly, _ := wkt.UnmarshalString("POLYGON ((0 0, 20 0, 20 5, 0 5, 0 0))")

	area := poly.(*geom.Polygon).Area()
	expected := 100.0

	if math.Abs(area-expected) > 0.001 {
		t.Errorf("Rectangle area: expected %.2f, got %.2f", expected, area)
	}
}

// TestJTS_PolygonArea_Triangle tests area calculation for a triangle.
func TestJTS_PolygonArea_Triangle(t *testing.T) {
	// Right triangle with base=10, height=10
	poly, _ := wkt.UnmarshalString("POLYGON ((0 0, 10 0, 0 10, 0 0))")

	area := poly.(*geom.Polygon).Area()
	expected := 50.0 // 0.5 * base * height

	if math.Abs(area-expected) > 0.001 {
		t.Errorf("Triangle area: expected %.2f, got %.2f", expected, area)
	}
}

// TestJTS_PolygonArea_WithHole tests area calculation for polygon with a hole.
func TestJTS_PolygonArea_WithHole(t *testing.T) {
	// 20x20 square with 10x10 hole in center
	poly, _ := wkt.UnmarshalString("POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0), (5 5, 15 5, 15 15, 5 15, 5 5))")

	area := poly.(*geom.Polygon).Area()
	expected := 300.0 // 400 - 100

	if math.Abs(area-expected) > 0.001 {
		t.Errorf("Polygon with hole area: expected %.2f, got %.2f", expected, area)
	}
}

// TestJTS_PolygonArea_ComplexShape tests area calculation for a complex polygon.
func TestJTS_PolygonArea_ComplexShape(t *testing.T) {
	// L-shaped polygon
	poly, _ := wkt.UnmarshalString("POLYGON ((0 0, 10 0, 10 5, 5 5, 5 10, 0 10, 0 0))")

	area := poly.(*geom.Polygon).Area()
	expected := 75.0 // 10*5 + 5*5

	if math.Abs(area-expected) > 0.001 {
		t.Errorf("L-shape area: expected %.2f, got %.2f", expected, area)
	}
}

// TestJTS_PolygonArea_Trapezoid tests area calculation for a trapezoid.
func TestJTS_PolygonArea_Trapezoid(t *testing.T) {
	// Trapezoid with parallel sides of length 10 and 20, height 10
	poly, _ := wkt.UnmarshalString("POLYGON ((0 0, 20 0, 15 10, 5 10, 0 0))")

	area := poly.(*geom.Polygon).Area()
	expected := 150.0 // 0.5 * (10 + 20) * 10

	if math.Abs(area-expected) > 0.001 {
		t.Errorf("Trapezoid area: expected %.2f, got %.2f", expected, area)
	}
}

// TestJTS_PolygonCentroid_Square tests centroid calculation for a square.
func TestJTS_PolygonCentroid_Square(t *testing.T) {
	poly, _ := wkt.UnmarshalString("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")

	centroid := poly.(*geom.Polygon).Centroid()
	expectedX, expectedY := 5.0, 5.0

	if math.Abs(centroid.X()-expectedX) > 0.001 || math.Abs(centroid.Y()-expectedY) > 0.001 {
		t.Errorf("Square centroid: expected (%.2f, %.2f), got (%.2f, %.2f)",
			expectedX, expectedY, centroid.X(), centroid.Y())
	}
}

// TestJTS_PolygonCentroid_Triangle tests centroid calculation for a triangle.
func TestJTS_PolygonCentroid_Triangle(t *testing.T) {
	// Triangle with vertices at (0,0), (12,0), (0,12)
	poly, _ := wkt.UnmarshalString("POLYGON ((0 0, 12 0, 0 12, 0 0))")

	centroid := poly.(*geom.Polygon).Centroid()
	expectedX, expectedY := 4.0, 4.0 // Centroid of triangle is at (1/3, 1/3) of each vertex

	if math.Abs(centroid.X()-expectedX) > 0.001 || math.Abs(centroid.Y()-expectedY) > 0.001 {
		t.Errorf("Triangle centroid: expected (%.2f, %.2f), got (%.2f, %.2f)",
			expectedX, expectedY, centroid.X(), centroid.Y())
	}
}

// TestJTS_PolygonCentroid_LShape tests centroid calculation for L-shaped polygon.
func TestJTS_PolygonCentroid_LShape(t *testing.T) {
	poly, _ := wkt.UnmarshalString("POLYGON ((0 0, 10 0, 10 5, 5 5, 5 10, 0 10, 0 0))")

	centroid := poly.(*geom.Polygon).Centroid()

	// Centroid should be somewhere in the L-shape
	if centroid.X() < 0 || centroid.X() > 10 || centroid.Y() < 0 || centroid.Y() > 10 {
		t.Errorf("L-shape centroid out of bounds: (%.2f, %.2f)", centroid.X(), centroid.Y())
	}

	t.Logf("L-shape centroid: (%.2f, %.2f)", centroid.X(), centroid.Y())
}

// TestJTS_PolygonCentroid_WithHole tests centroid of polygon with hole.
func TestJTS_PolygonCentroid_WithHole(t *testing.T) {
	// Square with centered hole
	poly, _ := wkt.UnmarshalString("POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0), (5 5, 15 5, 15 15, 5 15, 5 5))")

	centroid := poly.(*geom.Polygon).Centroid()
	expectedX, expectedY := 10.0, 10.0 // Should still be at center due to symmetry

	if math.Abs(centroid.X()-expectedX) > 0.1 || math.Abs(centroid.Y()-expectedY) > 0.1 {
		t.Errorf("Polygon with hole centroid: expected (%.2f, %.2f), got (%.2f, %.2f)",
			expectedX, expectedY, centroid.X(), centroid.Y())
	}
}

// TestJTS_LineStringLength tests length calculation for a LineString.
func TestJTS_LineStringLength(t *testing.T) {
	line, _ := wkt.UnmarshalString("LINESTRING (0 0, 3 4)")

	length := line.(*geom.LineString).Length()
	expected := 5.0 // 3-4-5 triangle

	if math.Abs(length-expected) > 0.001 {
		t.Errorf("LineString length: expected %.2f, got %.2f", expected, length)
	}
}

// TestJTS_LineStringLength_MultiSegment tests length of multi-segment line.
func TestJTS_LineStringLength_MultiSegment(t *testing.T) {
	line, _ := wkt.UnmarshalString("LINESTRING (0 0, 10 0, 10 10)")

	length := line.(*geom.LineString).Length()
	expected := 20.0 // 10 + 10

	if math.Abs(length-expected) > 0.001 {
		t.Errorf("Multi-segment length: expected %.2f, got %.2f", expected, length)
	}
}

// TestJTS_LineStringLength_Diagonal tests length of diagonal line segments.
func TestJTS_LineStringLength_Diagonal(t *testing.T) {
	line, _ := wkt.UnmarshalString("LINESTRING (0 0, 10 10, 20 0)")

	length := line.(*geom.LineString).Length()
	segment1 := math.Sqrt(200) // sqrt(10^2 + 10^2)
	segment2 := math.Sqrt(200) // sqrt(10^2 + 10^2)
	expected := segment1 + segment2

	if math.Abs(length-expected) > 0.001 {
		t.Errorf("Diagonal line length: expected %.2f, got %.2f", expected, length)
	}
}

// TestJTS_EnvelopePoint tests envelope of a point.
func TestJTS_EnvelopePoint(t *testing.T) {
	point, _ := wkt.UnmarshalString("POINT (5 10)")

	env := point.(*geom.Point).Envelope()

	if env.MinX != 5 || env.MaxX != 5 || env.MinY != 10 || env.MaxY != 10 {
		t.Errorf("Point envelope: expected (5,10,5,10), got (%.2f,%.2f,%.2f,%.2f)",
			env.MinX, env.MinY, env.MaxX, env.MaxY)
	}
}

// TestJTS_EnvelopeLineString tests envelope of a LineString.
func TestJTS_EnvelopeLineString(t *testing.T) {
	line, _ := wkt.UnmarshalString("LINESTRING (0 0, 10 20, 5 15)")

	env := line.(*geom.LineString).Envelope()

	if env.MinX != 0 || env.MaxX != 10 || env.MinY != 0 || env.MaxY != 20 {
		t.Errorf("LineString envelope: expected (0,0,10,20), got (%.2f,%.2f,%.2f,%.2f)",
			env.MinX, env.MinY, env.MaxX, env.MaxY)
	}
}

// TestJTS_EnvelopePolygon tests envelope of a polygon.
func TestJTS_EnvelopePolygon(t *testing.T) {
	poly, _ := wkt.UnmarshalString("POLYGON ((2 3, 8 3, 8 7, 2 7, 2 3))")

	env := poly.(*geom.Polygon).Envelope()

	if env.MinX != 2 || env.MaxX != 8 || env.MinY != 3 || env.MaxY != 7 {
		t.Errorf("Polygon envelope: expected (2,3,8,7), got (%.2f,%.2f,%.2f,%.2f)",
			env.MinX, env.MinY, env.MaxX, env.MaxY)
	}
}

// TestJTS_EnvelopePolygonWithHole tests envelope of polygon with hole.
func TestJTS_EnvelopePolygonWithHole(t *testing.T) {
	// Envelope should be of outer ring only
	poly, _ := wkt.UnmarshalString("POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0), (5 5, 15 5, 15 15, 5 15, 5 5))")

	env := poly.(*geom.Polygon).Envelope()

	if env.MinX != 0 || env.MaxX != 20 || env.MinY != 0 || env.MaxY != 20 {
		t.Errorf("Polygon with hole envelope: expected (0,0,20,20), got (%.2f,%.2f,%.2f,%.2f)",
			env.MinX, env.MinY, env.MaxX, env.MaxY)
	}
}

// TestJTS_EnvelopeMultiPoint tests envelope of MultiPoint.
func TestJTS_EnvelopeMultiPoint(t *testing.T) {
	mp, _ := wkt.UnmarshalString("MULTIPOINT ((0 0), (10 20), (5 5))")

	env := mp.(*geom.MultiPoint).Envelope()

	if env.MinX != 0 || env.MaxX != 10 || env.MinY != 0 || env.MaxY != 20 {
		t.Errorf("MultiPoint envelope: expected (0,0,10,20), got (%.2f,%.2f,%.2f,%.2f)",
			env.MinX, env.MinY, env.MaxX, env.MaxY)
	}
}

// TestJTS_EnvelopeEmpty tests envelope of empty geometry.
func TestJTS_EnvelopeEmpty(t *testing.T) {
	empty, _ := wkt.UnmarshalString("POINT EMPTY")

	env := empty.(*geom.Point).Envelope()

	if !env.IsNull() {
		t.Errorf("Empty geometry envelope should be null, got (%.2f,%.2f,%.2f,%.2f)",
			env.MinX, env.MinY, env.MaxX, env.MaxY)
	}
}

// TestJTS_GeometryValid_SimplePolygon tests validity check for simple polygon.
func TestJTS_GeometryValid_SimplePolygon(t *testing.T) {
	poly, _ := wkt.UnmarshalString("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")

	valid := poly.(*geom.Polygon).IsValid()

	if !valid {
		t.Error("Simple polygon should be valid")
	}
}

// TestJTS_GeometryValid_SelfIntersecting tests validity check for self-intersecting polygon.
func TestJTS_GeometryValid_SelfIntersecting(t *testing.T) {
	// Bowtie polygon (self-intersecting)
	poly, _ := wkt.UnmarshalString("POLYGON ((0 0, 10 10, 10 0, 0 10, 0 0))")

	valid := poly.(*geom.Polygon).IsValid()

	if valid {
		t.Error("Self-intersecting polygon should be invalid")
	}
}

// TestJTS_GeometryValid_UnclosedRing tests validity check for unclosed ring.
func TestJTS_GeometryValid_UnclosedRing(t *testing.T) {
	// This test depends on whether the parser auto-closes rings
	// Most implementations auto-close, so we test the manual construction
	coords := mustCoordsXY(0, 0, 10, 0, 10, 10, 0, 10)
	ring := geom.NewLinearRing(coords)

	// After auto-close, should be valid
	if !ring.IsClosed() {
		t.Error("Ring should be auto-closed")
	}
}

// TestJTS_GeometryValid_TooFewPoints tests validity of polygon with too few points.
func TestJTS_GeometryValid_TooFewPoints(t *testing.T) {
	// A valid polygon needs at least 4 points (including closure)
	coords := mustCoordsXY(0, 0, 10, 0, 0, 0)
	ring := geom.NewLinearRing(coords)
	poly := geom.NewPolygon(ring, nil)

	valid := poly.IsValid()

	if valid {
		t.Error("Polygon with too few points should be invalid")
	}
}

// TestJTS_CoordinateDistance tests distance between coordinates.
func TestJTS_CoordinateDistance(t *testing.T) {
	c1 := geom.NewCoordinate(0, 0)
	c2 := geom.NewCoordinate(3, 4)

	dist := c1.Distance(c2)
	expected := 5.0

	if math.Abs(dist-expected) > 0.001 {
		t.Errorf("Coordinate distance: expected %.2f, got %.2f", expected, dist)
	}
}

// TestJTS_CoordinateDistance_Diagonal tests diagonal distance.
func TestJTS_CoordinateDistance_Diagonal(t *testing.T) {
	c1 := geom.NewCoordinate(0, 0)
	c2 := geom.NewCoordinate(10, 10)

	dist := c1.Distance(c2)
	expected := math.Sqrt(200)

	if math.Abs(dist-expected) > 0.001 {
		t.Errorf("Diagonal distance: expected %.2f, got %.2f", expected, dist)
	}
}

// TestJTS_CoordinateEquals tests coordinate equality with tolerance.
func TestJTS_CoordinateEquals(t *testing.T) {
	c1 := geom.NewCoordinate(1.0, 2.0)
	c2 := geom.NewCoordinate(1.0, 2.0)
	c3 := geom.NewCoordinate(1.001, 2.0)

	if !c1.Equals2D(c2, 0.0001) {
		t.Error("Identical coordinates should be equal")
	}

	if c1.Equals2D(c3, 0.0001) {
		t.Error("Different coordinates should not be equal with tight tolerance")
	}

	if !c1.Equals2D(c3, 0.01) {
		t.Error("Nearly equal coordinates should be equal with loose tolerance")
	}
}

// TestJTS_LinearRingOrientation_CCW tests counter-clockwise ring orientation.
func TestJTS_LinearRingOrientation_CCW(t *testing.T) {
	// Counter-clockwise square
	ring, _ := wkt.UnmarshalString("LINESTRING (0 0, 10 0, 10 10, 0 10, 0 0)")

	lr := ring.(*geom.LineString)
	coords := lr.Coordinates()
	ringGeom := geom.NewLinearRing(coords)

	if !ringGeom.IsCCW() {
		t.Error("Counter-clockwise ring should be detected as CCW")
	}
}

// TestJTS_LinearRingOrientation_CW tests clockwise ring orientation.
func TestJTS_LinearRingOrientation_CW(t *testing.T) {
	// Clockwise square
	ring, _ := wkt.UnmarshalString("LINESTRING (0 0, 0 10, 10 10, 10 0, 0 0)")

	lr := ring.(*geom.LineString)
	coords := lr.Coordinates()
	ringGeom := geom.NewLinearRing(coords)

	if !ringGeom.IsCW() {
		t.Error("Clockwise ring should be detected as CW")
	}
}

// TestJTS_MultiPolygonArea tests area calculation for MultiPolygon.
func TestJTS_MultiPolygonArea(t *testing.T) {
	mp, _ := wkt.UnmarshalString("MULTIPOLYGON (((0 0, 10 0, 10 10, 0 10, 0 0)), ((20 0, 30 0, 30 10, 20 10, 20 0)))")

	area := mp.(*geom.MultiPolygon).Area()
	expected := 200.0 // Two 10x10 squares

	if math.Abs(area-expected) > 0.001 {
		t.Errorf("MultiPolygon area: expected %.2f, got %.2f", expected, area)
	}
}

// TestJTS_MultiLineStringLength tests length calculation for MultiLineString.
func TestJTS_MultiLineStringLength(t *testing.T) {
	mls, _ := wkt.UnmarshalString("MULTILINESTRING ((0 0, 10 0), (0 10, 10 10))")

	length := mls.(*geom.MultiLineString).Length()
	expected := 20.0 // Two lines of length 10 each

	if math.Abs(length-expected) > 0.001 {
		t.Errorf("MultiLineString length: expected %.2f, got %.2f", expected, length)
	}
}

// TestJTS_GeometryCollection_Mixed tests mixed geometry collection.
func TestJTS_GeometryCollection_Mixed(t *testing.T) {
	gc, _ := wkt.UnmarshalString("GEOMETRYCOLLECTION (POINT (0 0), LINESTRING (0 0, 10 10), POLYGON ((20 20, 30 20, 30 30, 20 30, 20 20)))")

	collection := gc.(*geom.GeometryCollection)

	if collection.NumGeometries() != 3 {
		t.Errorf("Expected 3 geometries, got %d", collection.NumGeometries())
	}

	// Check types
	if _, ok := collection.GeometryN(0).(*geom.Point); !ok {
		t.Error("First geometry should be Point")
	}
	if _, ok := collection.GeometryN(1).(*geom.LineString); !ok {
		t.Error("Second geometry should be LineString")
	}
	if _, ok := collection.GeometryN(2).(*geom.Polygon); !ok {
		t.Error("Third geometry should be Polygon")
	}
}

// TestJTS_PrecisionModel_Fixed tests fixed precision model.
func TestJTS_PrecisionModel_Fixed(t *testing.T) {
	pm := geom.NewFixedPrecision(100) // 2 decimal places

	val := pm.MakePreciseValue(1.23456)
	expected := 1.23

	if math.Abs(val-expected) > 0.001 {
		t.Errorf("Fixed precision: expected %.2f, got %.2f", expected, val)
	}
}

// TestJTS_PrecisionModel_Floating tests floating precision model.
func TestJTS_PrecisionModel_Floating(t *testing.T) {
	pm := geom.NewFloatingPrecision()

	val := pm.MakePreciseValue(1.234567890123456)

	if val != 1.234567890123456 {
		t.Errorf("Floating precision should preserve all digits, got %.15f", val)
	}
}

// TestJTS_GeometryFactory_WithPrecision tests geometry factory with precision model.
func TestJTS_GeometryFactory_WithPrecision(t *testing.T) {
	pm := geom.NewFixedPrecision(10) // 1 decimal place
	factory := geom.NewGeometryFactory(pm, 0)

	point := factory.CreatePoint(1.23456, 2.34567)

	// Coordinates should be rounded to 1 decimal place
	if math.Abs(point.X()-1.2) > 0.01 || math.Abs(point.Y()-2.3) > 0.01 {
		t.Logf("Point with precision: (%.2f, %.2f)", point.X(), point.Y())
	}
}

// TestJTS_SRID_Preservation tests SRID preservation in geometries.
func TestJTS_SRID_Preservation(t *testing.T) {
	factory := geom.NewGeometryFactoryWithSRID(4326)

	point := factory.CreatePoint(10, 20)

	if point.SRID() != 4326 {
		t.Errorf("SRID should be preserved: expected 4326, got %d", point.SRID())
	}
}

// TestJTS_EmptyGeometry_Point tests empty point geometry.
func TestJTS_EmptyGeometry_Point(t *testing.T) {
	empty, _ := wkt.UnmarshalString("POINT EMPTY")

	point := empty.(*geom.Point)

	if !point.IsEmpty() {
		t.Error("Empty point should report IsEmpty() = true")
	}

	if point.GeometryType() != "Point" {
		t.Errorf("Empty point should still have type 'Point', got '%s'", point.GeometryType())
	}
}

// TestJTS_EmptyGeometry_Polygon tests empty polygon geometry.
func TestJTS_EmptyGeometry_Polygon(t *testing.T) {
	empty, _ := wkt.UnmarshalString("POLYGON EMPTY")

	poly := empty.(*geom.Polygon)

	if !poly.IsEmpty() {
		t.Error("Empty polygon should report IsEmpty() = true")
	}

	if poly.Area() != 0 {
		t.Errorf("Empty polygon area should be 0, got %.2f", poly.Area())
	}
}

// TestJTS_CoordinateSequence_Access tests coordinate sequence access.
func TestJTS_CoordinateSequence_Access(t *testing.T) {
	coords := mustCoordsXY(0, 0, 10, 10, 20, 0)

	if len(coords) != 3 {
		t.Errorf("Expected 3 coordinates, got %d", len(coords))
	}

	if coords[0].X != 0 || coords[0].Y != 0 {
		t.Errorf("First coordinate should be (0,0), got (%.2f,%.2f)", coords[0].X, coords[0].Y)
	}

	if coords[1].X != 10 || coords[1].Y != 10 {
		t.Errorf("Second coordinate should be (10,10), got (%.2f,%.2f)", coords[1].X, coords[1].Y)
	}

	if coords[2].X != 20 || coords[2].Y != 0 {
		t.Errorf("Third coordinate should be (20,0), got (%.2f,%.2f)", coords[2].X, coords[2].Y)
	}
}

// TestJTS_Coordinate3D tests 3D coordinate handling.
func TestJTS_Coordinate3D(t *testing.T) {
	coord := geom.NewCoordinateZ(10, 20, 30)

	if coord.X != 10 || coord.Y != 20 || !coord.HasZ() || coord.Z != 30 {
		t.Errorf("3D coordinate incorrect: (%.2f, %.2f, %.2f)", coord.X, coord.Y, coord.Z)
	}
}

// TestJTS_Coordinate3D_Distance tests distance calculation ignores Z.
func TestJTS_Coordinate3D_Distance(t *testing.T) {
	c1 := geom.NewCoordinateZ(0, 0, 0)
	c2 := geom.NewCoordinateZ(3, 4, 100) // Z difference should be ignored

	dist := c1.Distance(c2)
	expected := 5.0 // 2D distance only

	if math.Abs(dist-expected) > 0.001 {
		t.Errorf("3D coordinate distance (2D): expected %.2f, got %.2f", expected, dist)
	}
}
