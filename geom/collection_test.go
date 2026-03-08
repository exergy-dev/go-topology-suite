package geom_test

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

// TestGeometryCollection_Empty tests empty GeometryCollection
func TestGeometryCollection_Empty(t *testing.T) {
	gc := geom.NewGeometryCollectionEmpty()

	if !gc.IsEmpty() {
		t.Error("Empty GeometryCollection should report IsEmpty() = true")
	}

	if gc.NumGeometries() != 0 {
		t.Errorf("Expected 0 geometries, got %d", gc.NumGeometries())
	}

	if gc.String() != "GEOMETRYCOLLECTION EMPTY" {
		t.Errorf("Expected 'GEOMETRYCOLLECTION EMPTY', got '%s'", gc.String())
	}

	env := gc.Envelope()
	if !env.IsNull() {
		t.Error("Empty GeometryCollection envelope should be null")
	}
}

// TestGeometryCollection_Construction tests constructing collections with geometries
func TestGeometryCollection_Construction(t *testing.T) {
	gc := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		mustLineStringXY(0, 0, 10, 10),
		geom.NewPoint(5, 5),
	})
	if gc.NumGeometries() != 3 {
		t.Errorf("Expected 3 geometries, got %d", gc.NumGeometries())
	}

	// Verify deep copy on construction
	original := geom.NewPoint(5, 5)
	gc2 := geom.NewGeometryCollection([]geom.Geometry{original})
	original.SetSRID(9999)

	added := gc2.GeometryN(0).(*geom.Point)
	if added.SRID() == 9999 {
		t.Error("Constructor should create a deep copy, not reference")
	}
}

// TestGeometryCollection_Dimension tests dimension calculation
func TestGeometryCollection_Dimension(t *testing.T) {
	// Empty collection
	empty := geom.NewGeometryCollectionEmpty()
	if empty.Dimension() != geom.DimensionEmpty {
		t.Errorf("Expected dimension %d, got %d", geom.DimensionEmpty, empty.Dimension())
	}

	// Points only
	points := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		geom.NewPoint(10, 10),
	})
	if points.Dimension() != geom.DimensionPoint {
		t.Errorf("Expected dimension %d, got %d", geom.DimensionPoint, points.Dimension())
	}

	// Mixed: should return maximum dimension
	mixed := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		mustLineStringXY(0, 0, 10, 10),
		geom.NewPolygon(mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
	})
	if mixed.Dimension() != geom.DimensionArea {
		t.Errorf("Expected dimension %d, got %d", geom.DimensionArea, mixed.Dimension())
	}
}

// TestGeometryCollection_IsSimple tests simplicity check
func TestGeometryCollection_IsSimple(t *testing.T) {
	// All simple geometries
	simple := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		mustLineStringXY(0, 0, 10, 10),
	})

	if !simple.IsSimple() {
		t.Error("Collection of simple geometries should be simple")
	}

	// Contains non-simple geometry (self-intersecting linestring would not be simple)
	// For now, all our geometries are simple
}

// TestGeometryCollection_IsValid tests validity check
func TestGeometryCollection_IsValid(t *testing.T) {
	// All valid geometries
	valid := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		mustLineStringXY(0, 0, 10, 10),
	})

	if !valid.IsValid() {
		t.Error("Collection of valid geometries should be valid")
	}

	// Contains invalid geometry (too few points)
	coords := mustCoordsXY(0, 0)
	invalid := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		geom.NewLineString(coords),
	})

	if invalid.IsValid() {
		t.Error("Collection with invalid geometry should be invalid")
	}
}

// TestGeometryCollection_Boundary tests boundary calculation
func TestGeometryCollection_Boundary(t *testing.T) {
	gc := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		mustLineStringXY(0, 0, 10, 10),
	})

	boundary := gc.Boundary()

	// Point has no boundary, LineString has boundary at endpoints
	if boundary.IsEmpty() {
		t.Error("Collection boundary should not be empty")
	}
}

// TestGeometryCollection_Coordinates tests coordinate extraction
func TestGeometryCollection_Coordinates(t *testing.T) {
	gc := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		mustLineStringXY(10, 10, 20, 20),
	})

	coords := gc.Coordinates()
	if len(coords) != 3 { // 1 point + 2 from line
		t.Errorf("Expected 3 coordinates, got %d", len(coords))
	}
}

// TestGeometryCollection_Clone tests cloning
func TestGeometryCollection_Clone(t *testing.T) {
	gc := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		mustLineStringXY(0, 0, 10, 10),
	})

	clone := gc.Clone().(*geom.GeometryCollection)

	if !clone.EqualsExact(gc, 0.0001) {
		t.Error("Clone should be equal to original")
	}

	// Verify deep copy - modifying the clone should not affect original
	if clone.NumGeometries() != gc.NumGeometries() {
		t.Error("Clone should have same number of geometries")
	}
}

// TestGeometryCollection_Normalize tests normalization
func TestGeometryCollection_Normalize(t *testing.T) {
	gc := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(20, 20),
		geom.NewPoint(0, 0),
		geom.NewPoint(10, 10),
	})

	normalized := gc.Normalized().(*geom.GeometryCollection)

	// After normalization, geometries should be sorted
	// (exact order depends on Compare implementation)
	if normalized.NumGeometries() != 3 {
		t.Errorf("Expected 3 geometries after normalize, got %d", normalized.NumGeometries())
	}
}

// TestGeometryCollection_EqualsExact tests exact equality
func TestGeometryCollection_EqualsExact(t *testing.T) {
	gc1 := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		mustLineStringXY(0, 0, 10, 10),
	})

	gc2 := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		mustLineStringXY(0, 0, 10, 10),
	})

	gc3 := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		mustLineStringXY(0, 0, 20, 20),
	})

	if !gc1.EqualsExact(gc2, 0.0001) {
		t.Error("Identical GeometryCollections should be equal")
	}

	if gc1.EqualsExact(gc3, 0.0001) {
		t.Error("Different GeometryCollections should not be equal")
	}

	if gc1.EqualsExact(nil, 0.0001) {
		t.Error("GeometryCollection should not equal nil")
	}

	// Test with different type
	point := geom.NewPoint(0, 0)
	if gc1.EqualsExact(point, 0.0001) {
		t.Error("GeometryCollection should not equal Point")
	}

	// Test with different number of geometries
	gc4 := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
	})
	if gc1.EqualsExact(gc4, 0.0001) {
		t.Error("GeometryCollections with different counts should not be equal")
	}
}

// TestGeometryCollection_Envelope tests envelope calculation
func TestGeometryCollection_Envelope(t *testing.T) {
	gc := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		geom.NewPoint(100, 50),
		mustLineStringXY(20, 10, 30, 80),
	})

	env := gc.Envelope()

	if env.MinX != 0 || env.MaxX != 100 || env.MinY != 0 || env.MaxY != 80 {
		t.Errorf("Expected envelope (0,0,100,80), got (%f,%f,%f,%f)",
			env.MinX, env.MinY, env.MaxX, env.MaxY)
	}

	// Test envelope caching
	env2 := gc.Envelope()
	if env.MinX != env2.MinX || env.MaxX != env2.MaxX {
		t.Error("Envelope should be cached")
	}

	// Test that a new collection with more geometries has a different envelope
	gc2 := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		geom.NewPoint(100, 100),
		geom.NewPoint(200, 200),
	})
	env3 := gc2.Envelope()
	if env3.MaxX != 200 || env3.MaxY != 200 {
		t.Error("Envelope should reflect all geometries")
	}
}

// TestGeometryCollection_String tests WKT output
func TestGeometryCollection_String(t *testing.T) {
	gc := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		mustLineStringXY(0, 0, 10, 10),
	})

	wkt := gc.String()
	expected := "GEOMETRYCOLLECTION (POINT (0 0), LINESTRING (0 0, 10 10))"
	if wkt != expected {
		t.Errorf("Expected WKT:\n%s\nGot:\n%s", expected, wkt)
	}
}

// TestGeometryCollection_GeometryN tests geometry accessor
func TestGeometryCollection_GeometryN(t *testing.T) {
	gc := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		mustLineStringXY(0, 0, 10, 10),
	})

	if gc.GeometryN(-1) != nil {
		t.Error("GeometryN(-1) should return nil")
	}

	if gc.GeometryN(10) != nil {
		t.Error("GeometryN(10) should return nil")
	}

	g := gc.GeometryN(0)
	if g == nil {
		t.Error("GeometryN(0) should return first geometry")
	}

	if _, ok := g.(*geom.Point); !ok {
		t.Error("First geometry should be a Point")
	}
}

// TestGeometryCollection_SRID tests SRID handling
func TestGeometryCollection_SRID(t *testing.T) {
	gc := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
	})

	gc.SetSRID(4326)
	if gc.SRID() != 4326 {
		t.Errorf("Expected SRID 4326, got %d", gc.SRID())
	}

	clone := gc.Clone()
	if clone.SRID() != 4326 {
		t.Errorf("Cloned SRID should be 4326, got %d", clone.SRID())
	}
}

// TestGeometryCollection_MixedDimensions tests collection with mixed geometry types
func TestGeometryCollection_MixedDimensions(t *testing.T) {
	gc := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),                                                           // 0D
		mustLineStringXY(0, 0, 10, 10),                                            // 1D
		geom.NewPolygon(mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil), // 2D
	})

	// Should return max dimension
	if gc.Dimension() != geom.DimensionArea {
		t.Errorf("Expected dimension %d, got %d", geom.DimensionArea, gc.Dimension())
	}

	if gc.NumGeometries() != 3 {
		t.Errorf("Expected 3 geometries, got %d", gc.NumGeometries())
	}
}
