package geom_test

import (
	"math"
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/stretchr/testify/assert"
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
	coords := mustCoordsXY(0, 0, 10, 10, 20, 20)
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

	normalized := mp.Normalized().(*geom.MultiPoint)

	// After normalization, points should be sorted
	// (exact order depends on Compare implementation)
	coords := normalized.Coordinates()
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
		mustLineStringXY(0, 0, 10, 0),         // Length 10
		mustLineStringXY(0, 0, 3, 4),          // Length 5
		mustLineStringXY(0, 0, 10, 10, 20, 0), // Length 10√2 + 10√2 ≈ 28.28
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
		mustLineStringXY(0, 0, 10, 0, 10, 10, 0, 0),
		mustLineStringXY(20, 20, 30, 20, 30, 30, 20, 20),
	})

	if !closed.IsClosed() {
		t.Error("MultiLineString with all closed lines should be closed")
	}

	// One open
	notClosed := geom.NewMultiLineString([]*geom.LineString{
		mustLineStringXY(0, 0, 10, 0, 10, 10, 0, 0), // Closed
		mustLineStringXY(20, 20, 30, 20),            // Open
	})

	if notClosed.IsClosed() {
		t.Error("MultiLineString with any open line should not be closed")
	}
}

// TestMultiLineString_Boundary tests boundary calculation
func TestMultiLineString_Boundary(t *testing.T) {
	// Open linestrings have boundary at endpoints
	mls := geom.NewMultiLineString([]*geom.LineString{
		mustLineStringXY(0, 0, 10, 0),
		mustLineStringXY(20, 0, 30, 0),
	})

	boundary := mls.Boundary()
	if boundary.IsEmpty() {
		t.Error("Boundary of open MultiLineString should not be empty")
	}

	// Closed linestrings have no boundary
	closedMLS := geom.NewMultiLineString([]*geom.LineString{
		mustLineStringXY(0, 0, 10, 0, 10, 10, 0, 0),
	})

	closedBoundary := closedMLS.Boundary()
	if !closedBoundary.IsEmpty() {
		t.Error("Boundary of closed MultiLineString should be empty")
	}
}

// TestMultiLineString_Dimension tests dimension
func TestMultiLineString_Dimension(t *testing.T) {
	mls := geom.NewMultiLineString([]*geom.LineString{
		mustLineStringXY(0, 0, 10, 0),
	})

	if mls.Dimension() != geom.DimensionLine {
		t.Errorf("Expected dimension %d, got %d", geom.DimensionLine, mls.Dimension())
	}
}

// TestMultiLineString_IsValid tests validity
func TestMultiLineString_IsValid(t *testing.T) {
	valid := geom.NewMultiLineString([]*geom.LineString{
		mustLineStringXY(0, 0, 10, 0),
		mustLineStringXY(20, 0, 30, 0),
	})

	if !valid.IsValid() {
		t.Error("Valid MultiLineString should report IsValid() = true")
	}

	// LineString with too few points is invalid
	coords := mustCoordsXY(0, 0)
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
		mustLineStringXY(0, 0, 10, 0),
		mustLineStringXY(20, 0, 30, 0),
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
		mustLineStringXY(0, 0, 10, 10),
		mustLineStringXY(20, 5, 30, 15),
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
	poly1 := geom.NewPolygon(mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
	poly2 := geom.NewPolygon(mustLinearRingXY(20, 20, 30, 20, 30, 30, 20, 30, 20, 20), nil)

	mp := geom.NewMultiPolygon([]*geom.Polygon{poly1, poly2})

	area := mp.Area()
	expected := 200.0 // Two 10x10 squares

	if math.Abs(area-expected) > 0.001 {
		t.Errorf("Expected area %f, got %f", expected, area)
	}
}

// TestMultiPolygon_Perimeter tests perimeter calculation
func TestMultiPolygon_Perimeter(t *testing.T) {
	poly1 := geom.NewPolygon(mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
	poly2 := geom.NewPolygon(mustLinearRingXY(20, 20, 25, 20, 25, 25, 20, 25, 20, 20), nil)

	mp := geom.NewMultiPolygon([]*geom.Polygon{poly1, poly2})

	perimeter := mp.Perimeter()
	expected := 40.0 + 20.0 // 10x10 square + 5x5 square

	if math.Abs(perimeter-expected) > 0.001 {
		t.Errorf("Expected perimeter %f, got %f", expected, perimeter)
	}
}

// TestMultiPolygon_Dimension tests dimension
func TestMultiPolygon_Dimension(t *testing.T) {
	poly := geom.NewPolygon(mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
	mp := geom.NewMultiPolygon([]*geom.Polygon{poly})

	if mp.Dimension() != geom.DimensionArea {
		t.Errorf("Expected dimension %d, got %d", geom.DimensionArea, mp.Dimension())
	}
}

// TestMultiPolygon_IsSimple tests simplicity
func TestMultiPolygon_IsSimple(t *testing.T) {
	poly := geom.NewPolygon(mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
	mp := geom.NewMultiPolygon([]*geom.Polygon{poly})

	if !mp.IsSimple() {
		t.Error("MultiPolygon should be simple by definition")
	}
}

// TestMultiPolygon_Boundary tests boundary calculation
func TestMultiPolygon_Boundary(t *testing.T) {
	poly := geom.NewPolygon(mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
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
	poly1 := geom.NewPolygon(mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
	poly2 := geom.NewPolygon(mustLinearRingXY(20, 20, 30, 20, 30, 30, 20, 30, 20, 20), nil)
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
	poly1 := geom.NewPolygon(mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
	poly2 := geom.NewPolygon(mustLinearRingXY(20, 5, 30, 5, 30, 15, 20, 15, 20, 5), nil)
	mp := geom.NewMultiPolygon([]*geom.Polygon{poly1, poly2})

	env := mp.Envelope()

	if env.MinX != 0 || env.MaxX != 30 || env.MinY != 0 || env.MaxY != 15 {
		t.Errorf("Expected envelope (0,0,30,15), got (%f,%f,%f,%f)",
			env.MinX, env.MinY, env.MaxX, env.MaxY)
	}
}

// TestMultiPolygon_Clone tests cloning
func TestMultiPolygon_Clone(t *testing.T) {
	poly := geom.NewPolygon(mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
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
	hole := mustLinearRingXY(2, 2, 8, 2, 8, 8, 2, 8, 2, 2)
	poly := geom.NewPolygon(
		mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0),
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
		mustLineStringXY(0, 0, 10, 0),
		mustLineStringXY(20, 0, 30, 0),
	})

	coords := mls.Coordinates()
	if len(coords) != 4 {
		t.Errorf("Expected 4 coordinates, got %d", len(coords))
	}
}

// TestMultiPolygon_Coordinates tests coordinate extraction
func TestMultiPolygon_Coordinates(t *testing.T) {
	poly := geom.NewPolygon(mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
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
		mustLineStringXY(0, 0, 10, 0),
		mustLineStringXY(20, 0, 30, 0),
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

// TestMultiLineString_IsSimple_InterLinestring tests that IsSimple detects
// inter-linestring crossing.
func TestMultiLineString_IsSimple_InterLinestring(t *testing.T) {
	t.Run("NonCrossing_IsSimple", func(t *testing.T) {
		// Two parallel lines - should be simple
		mls := geom.NewMultiLineString([]*geom.LineString{
			mustLineStringXY(0, 0, 10, 0),
			mustLineStringXY(0, 5, 10, 5),
		})
		assert.True(t, mls.IsSimple(), "Parallel lines should be simple")
	})

	t.Run("Crossing_NotSimple", func(t *testing.T) {
		// Two crossing lines (X shape) - should not be simple
		mls := geom.NewMultiLineString([]*geom.LineString{
			mustLineStringXY(0, 0, 10, 10),
			mustLineStringXY(0, 10, 10, 0),
		})
		assert.False(t, mls.IsSimple(), "Crossing lines should not be simple")
	})

	t.Run("TouchingAtEndpoint_IsSimple", func(t *testing.T) {
		// Two lines that touch only at an endpoint - should be simple
		mls := geom.NewMultiLineString([]*geom.LineString{
			mustLineStringXY(0, 0, 10, 0),
			mustLineStringXY(10, 0, 20, 0),
		})
		assert.True(t, mls.IsSimple(), "Lines touching at endpoints should be simple")
	})

	t.Run("EndpointTouchingInterior_NotSimple", func(t *testing.T) {
		mls := geom.NewMultiLineString([]*geom.LineString{
			mustLineStringXY(0, 0, 10, 0),
			mustLineStringXY(5, 0, 5, 5),
		})
		assert.False(t, mls.IsSimple(), "Endpoint touching another line interior should not be simple")
	})

	t.Run("CollinearOverlap_NotSimple", func(t *testing.T) {
		mls := geom.NewMultiLineString([]*geom.LineString{
			mustLineStringXY(0, 0, 10, 0),
			mustLineStringXY(5, 0, 15, 0),
		})
		assert.False(t, mls.IsSimple(), "Collinear overlapping lines should not be simple")
	})

	t.Run("NonIntersecting_IsSimple", func(t *testing.T) {
		// Two completely separate lines
		mls := geom.NewMultiLineString([]*geom.LineString{
			mustLineStringXY(0, 0, 10, 0),
			mustLineStringXY(100, 100, 110, 100),
		})
		assert.True(t, mls.IsSimple(), "Non-intersecting lines should be simple")
	})

	t.Run("SelfIntersectingLine_NotSimple", func(t *testing.T) {
		// One line that self-intersects (figure 8)
		mls := geom.NewMultiLineString([]*geom.LineString{
			mustLineStringXY(0, 0, 10, 10, 10, 0, 0, 10),
		})
		assert.False(t, mls.IsSimple(), "Self-intersecting line should not be simple")
	})
}

// TestMultiPolygon_IsValid_Overlap tests that IsValid detects overlapping polygons.
func TestMultiPolygon_IsValid_Overlap(t *testing.T) {
	t.Run("NonOverlapping_IsValid", func(t *testing.T) {
		// Two separate squares
		poly1 := geom.NewPolygon(mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
		poly2 := geom.NewPolygon(mustLinearRingXY(20, 0, 30, 0, 30, 10, 20, 10, 20, 0), nil)
		mp := geom.NewMultiPolygon([]*geom.Polygon{poly1, poly2})
		assert.True(t, mp.IsValid(), "Non-overlapping polygons should be valid")
	})

	t.Run("Overlapping_NotValid", func(t *testing.T) {
		// Two overlapping squares (second square overlaps first)
		poly1 := geom.NewPolygon(mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
		poly2 := geom.NewPolygon(mustLinearRingXY(5, 0, 15, 0, 15, 10, 5, 10, 5, 0), nil)
		mp := geom.NewMultiPolygon([]*geom.Polygon{poly1, poly2})
		assert.False(t, mp.IsValid(), "Overlapping polygons should not be valid")
	})

	t.Run("CornerOverlap_NoCentroidContainment_NotValid", func(t *testing.T) {
		// Overlap only at a corner region; centroids are outside each other.
		poly1 := geom.NewPolygon(mustLinearRingXY(0, 0, 4, 0, 4, 4, 0, 4, 0, 0), nil)
		poly2 := geom.NewPolygon(mustLinearRingXY(3, 3, 7, 3, 7, 7, 3, 7, 3, 3), nil)
		mp := geom.NewMultiPolygon([]*geom.Polygon{poly1, poly2})
		assert.False(t, mp.IsValid(), "Corner-overlapping polygons should not be valid")
	})

	t.Run("TouchingAtEdge_NotValid", func(t *testing.T) {
		// Two squares that share an edge (adjacent) - boundary overlap is invalid.
		poly1 := geom.NewPolygon(mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
		poly2 := geom.NewPolygon(mustLinearRingXY(10, 0, 20, 0, 20, 10, 10, 10, 10, 0), nil)
		mp := geom.NewMultiPolygon([]*geom.Polygon{poly1, poly2})
		assert.False(t, mp.IsValid(), "Edge-adjacent polygons should not be valid")
	})

	t.Run("PartialBoundaryOverlap_NotValid", func(t *testing.T) {
		poly1 := geom.NewPolygon(mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
		poly2 := geom.NewPolygon(mustLinearRingXY(10, 2, 20, 2, 20, 8, 10, 8, 10, 2), nil)
		mp := geom.NewMultiPolygon([]*geom.Polygon{poly1, poly2})
		assert.False(t, mp.IsValid(), "MultiPolygon polygons sharing a boundary segment should be invalid")
	})

	t.Run("HoleBoundaryOverlap_NotValid", func(t *testing.T) {
		outer := geom.NewPolygon(
			mustLinearRingXY(0, 0, 12, 0, 12, 12, 0, 12, 0, 0),
			[]*geom.LinearRing{mustLinearRingXY(4, 4, 8, 4, 8, 8, 4, 8, 4, 4)},
		)
		island := geom.NewPolygon(mustLinearRingXY(6, 5, 8, 5, 8, 7, 6, 7, 6, 5), nil)
		mp := geom.NewMultiPolygon([]*geom.Polygon{outer, island})
		assert.False(t, mp.IsValid(), "MultiPolygon polygon sharing a hole boundary segment should be invalid")
	})

	t.Run("ContainedPolygon_NotValid", func(t *testing.T) {
		// Small polygon inside larger polygon - should not be valid
		outerPoly := geom.NewPolygon(mustLinearRingXY(0, 0, 100, 0, 100, 100, 0, 100, 0, 0), nil)
		innerPoly := geom.NewPolygon(mustLinearRingXY(10, 10, 20, 10, 20, 20, 10, 20, 10, 10), nil)
		mp := geom.NewMultiPolygon([]*geom.Polygon{outerPoly, innerPoly})
		assert.False(t, mp.IsValid(), "Contained polygon should not be valid")
	})

	t.Run("InvalidComponentPolygon_NotValid", func(t *testing.T) {
		// One polygon is invalid (self-intersecting bowtie shell)
		validPoly := geom.NewPolygon(mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
		// This ring self-intersects (figure-8 / bowtie shape)
		invalidPoly := geom.NewPolygon(geom.NewLinearRing(geom.CoordinateSequence{
			geom.NewCoordinate(20, 0),
			geom.NewCoordinate(30, 10),
			geom.NewCoordinate(30, 0),
			geom.NewCoordinate(20, 10),
			geom.NewCoordinate(20, 0),
		}), nil)
		mp := geom.NewMultiPolygon([]*geom.Polygon{validPoly, invalidPoly})
		assert.False(t, mp.IsValid(), "MultiPolygon with invalid component should not be valid")
	})
}

// TestMultiPoint_StringZM tests that MultiPoint.String() preserves Z/M dimensions.
func TestMultiPoint_StringZM(t *testing.T) {
	t.Run("2D", func(t *testing.T) {
		mp := geom.NewMultiPoint([]*geom.Point{
			geom.NewPoint(1, 2),
			geom.NewPoint(3, 4),
		})
		got := mp.String()
		assert.Equal(t, "MULTIPOINT ((1 2), (3 4))", got)
	})

	t.Run("Z", func(t *testing.T) {
		mp := geom.NewMultiPointFromCoords(geom.CoordinateSequence{
			geom.NewCoordinateZ(1, 2, 3),
			geom.NewCoordinateZ(4, 5, 6),
		})
		got := mp.String()
		assert.Equal(t, "MULTIPOINT Z ((1 2 3), (4 5 6))", got)
	})

	t.Run("M", func(t *testing.T) {
		mp := geom.NewMultiPointFromCoords(geom.CoordinateSequence{
			geom.NewCoordinateM(1, 2, 10),
			geom.NewCoordinateM(3, 4, 20),
		})
		got := mp.String()
		assert.Equal(t, "MULTIPOINT M ((1 2 10), (3 4 20))", got)
	})

	t.Run("ZM", func(t *testing.T) {
		mp := geom.NewMultiPointFromCoords(geom.CoordinateSequence{
			geom.NewCoordinateZM(1, 2, 3, 10),
			geom.NewCoordinateZM(4, 5, 6, 20),
		})
		got := mp.String()
		assert.Equal(t, "MULTIPOINT ZM ((1 2 3 10), (4 5 6 20))", got)
	})
}

// TestMultiLineString_StringZM tests that MultiLineString.String() preserves Z/M dimensions.
func TestMultiLineString_StringZM(t *testing.T) {
	t.Run("2D", func(t *testing.T) {
		mls := geom.NewMultiLineString([]*geom.LineString{
			geom.NewLineString(geom.CoordinateSequence{
				geom.NewCoordinate(0, 0),
				geom.NewCoordinate(1, 1),
			}),
		})
		got := mls.String()
		assert.Equal(t, "MULTILINESTRING ((0 0, 1 1))", got)
	})

	t.Run("ZM", func(t *testing.T) {
		mls := geom.NewMultiLineString([]*geom.LineString{
			geom.NewLineString(geom.CoordinateSequence{
				geom.NewCoordinateZM(0, 0, 1, 10),
				geom.NewCoordinateZM(1, 1, 2, 20),
			}),
		})
		got := mls.String()
		assert.Contains(t, got, "MULTILINESTRING ZM ")
		assert.Contains(t, got, "0 0 1 10")
		assert.Contains(t, got, "1 1 2 20")
	})
}

// TestMultiPolygon_StringZM tests that MultiPolygon.String() preserves Z/M dimensions.
func TestMultiPolygon_StringZM(t *testing.T) {
	t.Run("2D", func(t *testing.T) {
		poly := geom.NewPolygon(mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
		mp := geom.NewMultiPolygon([]*geom.Polygon{poly})
		got := mp.String()
		assert.Contains(t, got, "MULTIPOLYGON (")
		assert.NotContains(t, got, "MULTIPOLYGON Z")
		assert.NotContains(t, got, "MULTIPOLYGON M")
	})

	t.Run("M", func(t *testing.T) {
		ring := geom.NewLinearRing(geom.CoordinateSequence{
			geom.NewCoordinateM(0, 0, 1),
			geom.NewCoordinateM(10, 0, 2),
			geom.NewCoordinateM(10, 10, 3),
			geom.NewCoordinateM(0, 10, 4),
			geom.NewCoordinateM(0, 0, 1),
		})
		poly := geom.NewPolygon(ring, nil)
		mp := geom.NewMultiPolygon([]*geom.Polygon{poly})
		got := mp.String()
		assert.Contains(t, got, "MULTIPOLYGON M ")
		assert.Contains(t, got, "0 0 1")
	})
}
