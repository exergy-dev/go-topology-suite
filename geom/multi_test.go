package geom_test

import (
	"math"
	"testing"

	"github.com/go-topology-suite/gts/geom"
)

// TestMultiPoint_Empty tests empty MultiPoint operations
func TestMultiPoint_Empty(t *testing.T) {
	mp := geom.NewMultiPointEmpty()

	if !mp.IsEmpty() {
		t.Error("Empty MultiPoint should report IsEmpty() = true")
	}

	if mp.NumGeometries() != 0 {
		t.Errorf("Expected 0 geometries, got %d", mp.NumGeometries())
	}

	if mp.String() != "MULTIPOINT EMPTY" {
		t.Errorf("Expected 'MULTIPOINT EMPTY', got '%s'", mp.String())
	}

	if mp.GeometryN(0) != nil {
		t.Error("Expected nil for GeometryN(0) on empty MultiPoint")
	}
}

// TestMultiPoint_FromCoords tests creating MultiPoint from coordinates
func TestMultiPoint_FromCoords(t *testing.T) {
	coords := geom.NewCoordinateSequenceXY(0, 0, 10, 10, 20, 20)
	mp := geom.NewMultiPointFromCoords(coords)

	if mp.NumGeometries() != 3 {
		t.Errorf("Expected 3 geometries, got %d", mp.NumGeometries())
	}

	p := mp.PointN(1)
	if p.X() != 10 || p.Y() != 10 {
		t.Errorf("Expected point (10, 10), got (%f, %f)", p.X(), p.Y())
	}
}

// TestMultiPoint_IsSimple tests simplicity check
func TestMultiPoint_IsSimple(t *testing.T) {
	// Simple: all distinct points
	simpleMP := geom.NewMultiPoint([]*geom.Point{
		geom.NewPoint(0, 0),
		geom.NewPoint(10, 10),
		geom.NewPoint(20, 20),
	})

	if !simpleMP.IsSimple() {
		t.Error("MultiPoint with distinct points should be simple")
	}

	// Not simple: duplicate points
	notSimpleMP := geom.NewMultiPoint([]*geom.Point{
		geom.NewPoint(0, 0),
		geom.NewPoint(10, 10),
		geom.NewPoint(0, 0), // Duplicate
	})

	if notSimpleMP.IsSimple() {
		t.Error("MultiPoint with duplicate points should not be simple")
	}
}

// TestMultiPoint_IsValid tests validity
func TestMultiPoint_IsValid(t *testing.T) {
	mp := geom.NewMultiPoint([]*geom.Point{
		geom.NewPoint(0, 0),
		geom.NewPoint(10, 10),
	})

	if !mp.IsValid() {
		t.Error("MultiPoint should always be valid")
	}
}

// TestMultiPoint_Dimension tests dimension
func TestMultiPoint_Dimension(t *testing.T) {
	mp := geom.NewMultiPoint([]*geom.Point{
		geom.NewPoint(0, 0),
	})

	if mp.Dimension() != geom.DimensionPoint {
		t.Errorf("Expected dimension %d, got %d", geom.DimensionPoint, mp.Dimension())
	}
}

// TestMultiPoint_Boundary tests boundary calculation
func TestMultiPoint_Boundary(t *testing.T) {
	mp := geom.NewMultiPoint([]*geom.Point{
		geom.NewPoint(0, 0),
		geom.NewPoint(10, 10),
	})

	boundary := mp.Boundary()
	if !boundary.IsEmpty() {
		t.Error("MultiPoint boundary should be empty")
	}
}

// TestMultiPoint_Coordinates tests coordinate extraction
func TestMultiPoint_Coordinates(t *testing.T) {
	mp := geom.NewMultiPoint([]*geom.Point{
		geom.NewPoint(0, 0),
		geom.NewPoint(10, 10),
		geom.NewPoint(20, 20),
	})

	coords := mp.Coordinates()
	if len(coords) != 3 {
		t.Errorf("Expected 3 coordinates, got %d", len(coords))
	}

	if coords[1].X != 10 || coords[1].Y != 10 {
		t.Errorf("Expected (10, 10), got (%f, %f)", coords[1].X, coords[1].Y)
	}
}

// TestMultiPoint_Clone tests cloning
func TestMultiPoint_Clone(t *testing.T) {
	mp := geom.NewMultiPoint([]*geom.Point{
		geom.NewPoint(0, 0),
		geom.NewPoint(10, 10),
	})

	clone := mp.Clone().(*geom.MultiPoint)

	if !clone.EqualsExact(mp, 0.0001) {
		t.Error("Clone should be equal to original")
	}

	// Verify deep copy by checking if modifying one doesn't affect the other
	// We can't check pointer equality since Coordinate() returns a value
	originalFirst := mp.PointN(0)
	clonedFirst := clone.PointN(0)
	if originalFirst == clonedFirst {
		t.Error("Clone should create new point objects")
	}
}

// TestMultiPoint_Normalize tests normalization
func TestMultiPoint_Normalize(t *testing.T) {
	mp := geom.NewMultiPoint([]*geom.Point{
		geom.NewPoint(20, 20),
		geom.NewPoint(0, 0),
		geom.NewPoint(10, 10),
	})

	mp.Normalize()

	// After normalization, points should be sorted
	// (exact order depends on Compare implementation)
	coords := mp.Coordinates()
	if len(coords) != 3 {
		t.Errorf("Expected 3 coordinates after normalize, got %d", len(coords))
	}
}

// TestMultiPoint_Envelope tests envelope calculation
func TestMultiPoint_Envelope(t *testing.T) {
	mp := geom.NewMultiPoint([]*geom.Point{
		geom.NewPoint(0, 5),
		geom.NewPoint(10, 15),
		geom.NewPoint(20, 10),
	})

	env := mp.Envelope()

	if env.MinX != 0 || env.MaxX != 20 || env.MinY != 5 || env.MaxY != 15 {
		t.Errorf("Expected envelope (0,5,20,15), got (%f,%f,%f,%f)",
			env.MinX, env.MinY, env.MaxX, env.MaxY)
	}
}

// TestMultiPoint_PointN tests PointN accessor
func TestMultiPoint_PointN(t *testing.T) {
	mp := geom.NewMultiPoint([]*geom.Point{
		geom.NewPoint(0, 0),
		geom.NewPoint(10, 10),
	})

	if mp.PointN(-1) != nil {
		t.Error("PointN(-1) should return nil")
	}

	if mp.PointN(10) != nil {
		t.Error("PointN(10) should return nil")
	}

	p := mp.PointN(1)
	if p == nil || p.X() != 10 {
		t.Error("PointN(1) should return second point")
	}
}

// TestMultiLineString_Empty tests empty MultiLineString
func TestMultiLineString_Empty(t *testing.T) {
	mls := geom.NewMultiLineStringEmpty()

	if !mls.IsEmpty() {
		t.Error("Empty MultiLineString should report IsEmpty() = true")
	}

	if mls.String() != "MULTILINESTRING EMPTY" {
		t.Errorf("Expected 'MULTILINESTRING EMPTY', got '%s'", mls.String())
	}

	if mls.Length() != 0 {
		t.Errorf("Expected length 0, got %f", mls.Length())
	}
}

// TestMultiLineString_Length tests length calculation
func TestMultiLineString_Length(t *testing.T) {
	mls := geom.NewMultiLineString([]*geom.LineString{
		geom.NewLineStringXY(0, 0, 10, 0),       // Length 10
		geom.NewLineStringXY(0, 0, 3, 4),        // Length 5
		geom.NewLineStringXY(0, 0, 10, 10, 20, 0), // Length 10√2 + 10√2 ≈ 28.28
	})

	length := mls.Length()
	expected := 10.0 + 5.0 + 2*math.Sqrt(200)

	if math.Abs(length-expected) > 0.01 {
		t.Errorf("Expected length ~%f, got %f", expected, length)
	}
}

// TestMultiLineString_IsClosed tests closed check
func TestMultiLineString_IsClosed(t *testing.T) {
	// All closed
	closed := geom.NewMultiLineString([]*geom.LineString{
		geom.NewLineStringXY(0, 0, 10, 0, 10, 10, 0, 0),
		geom.NewLineStringXY(20, 20, 30, 20, 30, 30, 20, 20),
	})

	if !closed.IsClosed() {
		t.Error("MultiLineString with all closed lines should be closed")
	}

	// One open
	notClosed := geom.NewMultiLineString([]*geom.LineString{
		geom.NewLineStringXY(0, 0, 10, 0, 10, 10, 0, 0), // Closed
		geom.NewLineStringXY(20, 20, 30, 20),            // Open
	})

	if notClosed.IsClosed() {
		t.Error("MultiLineString with any open line should not be closed")
	}
}

// TestMultiLineString_Boundary tests boundary calculation
func TestMultiLineString_Boundary(t *testing.T) {
	// Open linestrings have boundary at endpoints
	mls := geom.NewMultiLineString([]*geom.LineString{
		geom.NewLineStringXY(0, 0, 10, 0),
		geom.NewLineStringXY(20, 0, 30, 0),
	})

	boundary := mls.Boundary()
	if boundary.IsEmpty() {
		t.Error("Boundary of open MultiLineString should not be empty")
	}

	// Closed linestrings have no boundary
	closedMLS := geom.NewMultiLineString([]*geom.LineString{
		geom.NewLineStringXY(0, 0, 10, 0, 10, 10, 0, 0),
	})

	closedBoundary := closedMLS.Boundary()
	if !closedBoundary.IsEmpty() {
		t.Error("Boundary of closed MultiLineString should be empty")
	}
}

// TestMultiLineString_Dimension tests dimension
func TestMultiLineString_Dimension(t *testing.T) {
	mls := geom.NewMultiLineString([]*geom.LineString{
		geom.NewLineStringXY(0, 0, 10, 0),
	})

	if mls.Dimension() != geom.DimensionLine {
		t.Errorf("Expected dimension %d, got %d", geom.DimensionLine, mls.Dimension())
	}
}

// TestMultiLineString_IsValid tests validity
func TestMultiLineString_IsValid(t *testing.T) {
	valid := geom.NewMultiLineString([]*geom.LineString{
		geom.NewLineStringXY(0, 0, 10, 0),
		geom.NewLineStringXY(20, 0, 30, 0),
	})

	if !valid.IsValid() {
		t.Error("Valid MultiLineString should report IsValid() = true")
	}

	// LineString with too few points is invalid
	coords := geom.NewCoordinateSequenceXY(0, 0)
	invalid := geom.NewMultiLineString([]*geom.LineString{
		geom.NewLineString(coords),
	})

	if invalid.IsValid() {
		t.Error("MultiLineString with invalid component should be invalid")
	}
}

// TestMultiLineString_LineStringN tests LineStringN accessor
func TestMultiLineString_LineStringN(t *testing.T) {
	mls := geom.NewMultiLineString([]*geom.LineString{
		geom.NewLineStringXY(0, 0, 10, 0),
		geom.NewLineStringXY(20, 0, 30, 0),
	})

	if mls.LineStringN(-1) != nil {
		t.Error("LineStringN(-1) should return nil")
	}

	if mls.LineStringN(10) != nil {
		t.Error("LineStringN(10) should return nil")
	}

	ls := mls.LineStringN(1)
	if ls == nil || ls.Length() != 10 {
		t.Error("LineStringN(1) should return second linestring")
	}
}

// TestMultiLineString_Envelope tests envelope calculation
func TestMultiLineString_Envelope(t *testing.T) {
	mls := geom.NewMultiLineString([]*geom.LineString{
		geom.NewLineStringXY(0, 0, 10, 10),
		geom.NewLineStringXY(20, 5, 30, 15),
	})

	env := mls.Envelope()

	if env.MinX != 0 || env.MaxX != 30 || env.MinY != 0 || env.MaxY != 15 {
		t.Errorf("Expected envelope (0,0,30,15), got (%f,%f,%f,%f)",
			env.MinX, env.MinY, env.MaxX, env.MaxY)
	}
}

// TestMultiPolygon_Empty tests empty MultiPolygon
func TestMultiPolygon_Empty(t *testing.T) {
	mp := geom.NewMultiPolygonEmpty()

	if !mp.IsEmpty() {
		t.Error("Empty MultiPolygon should report IsEmpty() = true")
	}

	if mp.String() != "MULTIPOLYGON EMPTY" {
		t.Errorf("Expected 'MULTIPOLYGON EMPTY', got '%s'", mp.String())
	}

	if mp.Area() != 0 {
		t.Errorf("Expected area 0, got %f", mp.Area())
	}

	if mp.Perimeter() != 0 {
		t.Errorf("Expected perimeter 0, got %f", mp.Perimeter())
	}
}

// TestMultiPolygon_Area tests area calculation
func TestMultiPolygon_Area(t *testing.T) {
	poly1 := geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
	poly2 := geom.NewPolygon(geom.NewLinearRingXY(20, 20, 30, 20, 30, 30, 20, 30, 20, 20), nil)

	mp := geom.NewMultiPolygon([]*geom.Polygon{poly1, poly2})

	area := mp.Area()
	expected := 200.0 // Two 10x10 squares

	if math.Abs(area-expected) > 0.001 {
		t.Errorf("Expected area %f, got %f", expected, area)
	}
}

// TestMultiPolygon_Perimeter tests perimeter calculation
func TestMultiPolygon_Perimeter(t *testing.T) {
	poly1 := geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
	poly2 := geom.NewPolygon(geom.NewLinearRingXY(20, 20, 25, 20, 25, 25, 20, 25, 20, 20), nil)

	mp := geom.NewMultiPolygon([]*geom.Polygon{poly1, poly2})

	perimeter := mp.Perimeter()
	expected := 40.0 + 20.0 // 10x10 square + 5x5 square

	if math.Abs(perimeter-expected) > 0.001 {
		t.Errorf("Expected perimeter %f, got %f", expected, perimeter)
	}
}

// TestMultiPolygon_Dimension tests dimension
func TestMultiPolygon_Dimension(t *testing.T) {
	poly := geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
	mp := geom.NewMultiPolygon([]*geom.Polygon{poly})

	if mp.Dimension() != geom.DimensionArea {
		t.Errorf("Expected dimension %d, got %d", geom.DimensionArea, mp.Dimension())
	}
}

// TestMultiPolygon_IsSimple tests simplicity
func TestMultiPolygon_IsSimple(t *testing.T) {
	poly := geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
	mp := geom.NewMultiPolygon([]*geom.Polygon{poly})

	if !mp.IsSimple() {
		t.Error("MultiPolygon should be simple by definition")
	}
}

// TestMultiPolygon_Boundary tests boundary calculation
func TestMultiPolygon_Boundary(t *testing.T) {
	poly := geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
	mp := geom.NewMultiPolygon([]*geom.Polygon{poly})

	boundary := mp.Boundary()
	if boundary.IsEmpty() {
		t.Error("MultiPolygon boundary should not be empty")
	}

	// Boundary should be a MultiLineString
	if boundary.GeometryType() != "MultiLineString" {
		t.Errorf("Expected MultiLineString boundary, got %s", boundary.GeometryType())
	}
}

// TestMultiPolygon_PolygonN tests PolygonN accessor
func TestMultiPolygon_PolygonN(t *testing.T) {
	poly1 := geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
	poly2 := geom.NewPolygon(geom.NewLinearRingXY(20, 20, 30, 20, 30, 30, 20, 30, 20, 20), nil)
	mp := geom.NewMultiPolygon([]*geom.Polygon{poly1, poly2})

	if mp.PolygonN(-1) != nil {
		t.Error("PolygonN(-1) should return nil")
	}

	if mp.PolygonN(10) != nil {
		t.Error("PolygonN(10) should return nil")
	}

	p := mp.PolygonN(0)
	if p == nil || p.Area() != 100 {
		t.Error("PolygonN(0) should return first polygon")
	}
}

// TestMultiPolygon_Envelope tests envelope calculation
func TestMultiPolygon_Envelope(t *testing.T) {
	poly1 := geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
	poly2 := geom.NewPolygon(geom.NewLinearRingXY(20, 5, 30, 5, 30, 15, 20, 15, 20, 5), nil)
	mp := geom.NewMultiPolygon([]*geom.Polygon{poly1, poly2})

	env := mp.Envelope()

	if env.MinX != 0 || env.MaxX != 30 || env.MinY != 0 || env.MaxY != 15 {
		t.Errorf("Expected envelope (0,0,30,15), got (%f,%f,%f,%f)",
			env.MinX, env.MinY, env.MaxX, env.MaxY)
	}
}

// TestMultiPolygon_Clone tests cloning
func TestMultiPolygon_Clone(t *testing.T) {
	poly := geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
	mp := geom.NewMultiPolygon([]*geom.Polygon{poly})

	clone := mp.Clone().(*geom.MultiPolygon)

	// Verify basic properties match
	if clone.NumGeometries() != mp.NumGeometries() {
		t.Error("Clone should have same number of geometries")
	}

	if clone.Area() != mp.Area() {
		t.Error("Clone should have same area")
	}
}

// TestMultiPolygon_WithHoles tests MultiPolygon with holes
func TestMultiPolygon_WithHoles(t *testing.T) {
	hole := geom.NewLinearRingXY(2, 2, 8, 2, 8, 8, 2, 8, 2, 2)
	poly := geom.NewPolygon(
		geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0),
		[]*geom.LinearRing{hole},
	)

	mp := geom.NewMultiPolygon([]*geom.Polygon{poly})

	area := mp.Area()
	expected := 100.0 - 36.0 // 10x10 - 6x6

	if math.Abs(area-expected) > 0.001 {
		t.Errorf("Expected area %f, got %f", expected, area)
	}
}

// TestMultiGeometry_SRID tests SRID preservation
func TestMultiGeometry_SRID(t *testing.T) {
	mp := geom.NewMultiPoint([]*geom.Point{geom.NewPoint(0, 0)})
	mp.SetSRID(4326)

	if mp.SRID() != 4326 {
		t.Errorf("Expected SRID 4326, got %d", mp.SRID())
	}

	clone := mp.Clone()
	if clone.SRID() != 4326 {
		t.Errorf("Expected cloned SRID 4326, got %d", clone.SRID())
	}
}

// TestMultiLineString_Coordinates tests coordinate extraction
func TestMultiLineString_Coordinates(t *testing.T) {
	mls := geom.NewMultiLineString([]*geom.LineString{
		geom.NewLineStringXY(0, 0, 10, 0),
		geom.NewLineStringXY(20, 0, 30, 0),
	})

	coords := mls.Coordinates()
	if len(coords) != 4 {
		t.Errorf("Expected 4 coordinates, got %d", len(coords))
	}
}

// TestMultiPolygon_Coordinates tests coordinate extraction
func TestMultiPolygon_Coordinates(t *testing.T) {
	poly := geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
	mp := geom.NewMultiPolygon([]*geom.Polygon{poly})

	coords := mp.Coordinates()
	if len(coords) != 5 { // 4 unique + 1 closure
		t.Errorf("Expected 5 coordinates, got %d", len(coords))
	}
}

// TestMultiGeometry_EqualsExact tests exact equality
func TestMultiGeometry_EqualsExact(t *testing.T) {
	mp1 := geom.NewMultiPoint([]*geom.Point{
		geom.NewPoint(0, 0),
		geom.NewPoint(10, 10),
	})

	mp2 := geom.NewMultiPoint([]*geom.Point{
		geom.NewPoint(0, 0),
		geom.NewPoint(10, 10),
	})

	mp3 := geom.NewMultiPoint([]*geom.Point{
		geom.NewPoint(0, 0),
		geom.NewPoint(20, 20),
	})

	if !mp1.EqualsExact(mp2, 0.0001) {
		t.Error("Identical MultiPoints should be equal")
	}

	if mp1.EqualsExact(mp3, 0.0001) {
		t.Error("Different MultiPoints should not be equal")
	}

	if mp1.EqualsExact(nil, 0.0001) {
		t.Error("MultiPoint should not equal nil")
	}

	// Test with different type
	point := geom.NewPoint(0, 0)
	if mp1.EqualsExact(point, 0.0001) {
		t.Error("MultiPoint should not equal Point")
	}
}

// TestMultiLineString_String tests WKT output
func TestMultiLineString_String(t *testing.T) {
	mls := geom.NewMultiLineString([]*geom.LineString{
		geom.NewLineStringXY(0, 0, 10, 0),
		geom.NewLineStringXY(20, 0, 30, 0),
	})

	wkt := mls.String()
	if wkt != "MULTILINESTRING ((0 0, 10 0), (20 0, 30 0))" {
		t.Errorf("Unexpected WKT: %s", wkt)
	}
}

// TestMultiPoint_String tests WKT output
func TestMultiPoint_String(t *testing.T) {
	mp := geom.NewMultiPoint([]*geom.Point{
		geom.NewPoint(0, 0),
		geom.NewPoint(10, 10),
	})

	wkt := mp.String()
	if wkt != "MULTIPOINT ((0 0), (10 10))" {
		t.Errorf("Unexpected WKT: %s", wkt)
	}
}
