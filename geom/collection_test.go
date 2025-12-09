package geom_test

import (
	"testing"

	"github.com/go-topology-suite/gts/geom"
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

// TestGeometryCollection_Add tests adding geometries
func TestGeometryCollection_Add(t *testing.T) {
	gc := geom.NewGeometryCollectionEmpty()

	gc.Add(geom.NewPoint(0, 0))
	if gc.NumGeometries() != 1 {
		t.Errorf("Expected 1 geometry after add, got %d", gc.NumGeometries())
	}

	gc.Add(geom.NewLineStringXY(0, 0, 10, 10))
	if gc.NumGeometries() != 2 {
		t.Errorf("Expected 2 geometries after second add, got %d", gc.NumGeometries())
	}

	// Verify deep copy on add
	original := geom.NewPoint(5, 5)
	gc.Add(original)
	original.SetSRID(9999)

	added := gc.GeometryN(2).(*geom.Point)
	if added.SRID() == 9999 {
		t.Error("Add should create a deep copy, not reference")
	}
}

// TestGeometryCollection_Remove tests removing geometries
func TestGeometryCollection_Remove(t *testing.T) {
	gc := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		geom.NewPoint(10, 10),
		geom.NewPoint(20, 20),
	})

	gc.Remove(1) // Remove middle element
	if gc.NumGeometries() != 2 {
		t.Errorf("Expected 2 geometries after remove, got %d", gc.NumGeometries())
	}

	// Verify correct element was removed
	if gc.GeometryN(0).(*geom.Point).X() != 0 {
		t.Error("First element should still be (0,0)")
	}
	if gc.GeometryN(1).(*geom.Point).X() != 20 {
		t.Error("Second element should now be (20,20)")
	}

	// Test remove invalid index
	gc.Remove(-1)
	if gc.NumGeometries() != 2 {
		t.Error("Remove with invalid index should not change collection")
	}

	gc.Remove(100)
	if gc.NumGeometries() != 2 {
		t.Error("Remove with out-of-bounds index should not change collection")
	}
}

// TestGeometryCollection_Filter tests filtering geometries
func TestGeometryCollection_Filter(t *testing.T) {
	gc := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		geom.NewLineStringXY(0, 0, 10, 10),
		geom.NewPoint(10, 10),
		geom.NewLineStringXY(20, 20, 30, 30),
	})

	// Filter only points
	points := gc.Filter(func(g geom.Geometry) bool {
		_, ok := g.(*geom.Point)
		return ok
	})

	if points.NumGeometries() != 2 {
		t.Errorf("Expected 2 points, got %d", points.NumGeometries())
	}

	for i := 0; i < points.NumGeometries(); i++ {
		if _, ok := points.GeometryN(i).(*geom.Point); !ok {
			t.Error("Filtered collection should only contain points")
		}
	}

	// Filter only lines
	lines := gc.Filter(func(g geom.Geometry) bool {
		_, ok := g.(*geom.LineString)
		return ok
	})

	if lines.NumGeometries() != 2 {
		t.Errorf("Expected 2 lines, got %d", lines.NumGeometries())
	}
}

// TestGeometryCollection_Map tests mapping over geometries
func TestGeometryCollection_Map(t *testing.T) {
	gc := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		geom.NewPoint(10, 10),
		geom.NewPoint(20, 20),
	})

	// Map to create envelopes
	envelopes := gc.Map(func(g geom.Geometry) geom.Geometry {
		env := g.Envelope()
		// Convert envelope to a point at its center (for testing)
		centerX := (env.MinX + env.MaxX) / 2
		centerY := (env.MinY + env.MaxY) / 2
		return geom.NewPoint(centerX, centerY)
	})

	if envelopes.NumGeometries() != 3 {
		t.Errorf("Expected 3 mapped geometries, got %d", envelopes.NumGeometries())
	}

	// Verify first point
	p := envelopes.GeometryN(0).(*geom.Point)
	if p.X() != 0 || p.Y() != 0 {
		t.Errorf("Expected mapped point (0,0), got (%f,%f)", p.X(), p.Y())
	}
}

// TestGeometryCollection_ForEach tests iteration
func TestGeometryCollection_ForEach(t *testing.T) {
	gc := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		geom.NewPoint(10, 10),
		geom.NewPoint(20, 20),
	})

	count := 0
	gc.ForEach(func(g geom.Geometry) {
		count++
		if _, ok := g.(*geom.Point); !ok {
			t.Error("ForEach should iterate over all geometries")
		}
	})

	if count != 3 {
		t.Errorf("Expected ForEach to visit 3 geometries, visited %d", count)
	}
}

// TestGeometryCollection_Points tests extracting points
func TestGeometryCollection_Points(t *testing.T) {
	gc := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		geom.NewLineStringXY(0, 0, 10, 10),
		geom.NewPoint(10, 10),
		geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
	})

	points := gc.Points()
	if len(points) != 2 {
		t.Errorf("Expected 2 points, got %d", len(points))
	}
}

// TestGeometryCollection_LineStrings tests extracting linestrings
func TestGeometryCollection_LineStrings(t *testing.T) {
	gc := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		geom.NewLineStringXY(0, 0, 10, 10),
		geom.NewLineStringXY(20, 20, 30, 30),
		geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
	})

	lines := gc.LineStrings()
	if len(lines) != 2 {
		t.Errorf("Expected 2 linestrings, got %d", len(lines))
	}
}

// TestGeometryCollection_Polygons tests extracting polygons
func TestGeometryCollection_Polygons(t *testing.T) {
	gc := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		geom.NewLineStringXY(0, 0, 10, 10),
		geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
		geom.NewPolygon(geom.NewLinearRingXY(20, 20, 30, 20, 30, 30, 20, 30, 20, 20), nil),
	})

	polygons := gc.Polygons()
	if len(polygons) != 2 {
		t.Errorf("Expected 2 polygons, got %d", len(polygons))
	}
}

// TestGeometryCollection_Flatten tests flattening nested collections
func TestGeometryCollection_Flatten(t *testing.T) {
	// Create nested collection
	inner := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		geom.NewPoint(10, 10),
	})

	outer := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(20, 20),
		inner,
		geom.NewLineStringXY(0, 0, 10, 10),
	})

	flattened := outer.Flatten()

	// Should have 4 geometries: 3 points + 1 line
	if flattened.NumGeometries() != 4 {
		t.Errorf("Expected 4 flattened geometries, got %d", flattened.NumGeometries())
	}

	// Verify no nested collections
	for i := 0; i < flattened.NumGeometries(); i++ {
		if _, ok := flattened.GeometryN(i).(*geom.GeometryCollection); ok {
			t.Error("Flattened collection should not contain GeometryCollections")
		}
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
		geom.NewLineStringXY(0, 0, 10, 10),
		geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
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
		geom.NewLineStringXY(0, 0, 10, 10),
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
		geom.NewLineStringXY(0, 0, 10, 10),
	})

	if !valid.IsValid() {
		t.Error("Collection of valid geometries should be valid")
	}

	// Contains invalid geometry (too few points)
	coords := geom.NewCoordinateSequenceXY(0, 0)
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
		geom.NewLineStringXY(0, 0, 10, 10),
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
		geom.NewLineStringXY(10, 10, 20, 20),
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
		geom.NewLineStringXY(0, 0, 10, 10),
	})

	clone := gc.Clone().(*geom.GeometryCollection)

	if !clone.EqualsExact(gc, 0.0001) {
		t.Error("Clone should be equal to original")
	}

	// Verify deep copy
	gc.Add(geom.NewPoint(100, 100))
	if clone.NumGeometries() == gc.NumGeometries() {
		t.Error("Clone should be independent of original")
	}
}

// TestGeometryCollection_Normalize tests normalization
func TestGeometryCollection_Normalize(t *testing.T) {
	gc := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(20, 20),
		geom.NewPoint(0, 0),
		geom.NewPoint(10, 10),
	})

	gc.Normalize()

	// After normalization, geometries should be sorted
	// (exact order depends on Compare implementation)
	if gc.NumGeometries() != 3 {
		t.Errorf("Expected 3 geometries after normalize, got %d", gc.NumGeometries())
	}
}

// TestGeometryCollection_EqualsExact tests exact equality
func TestGeometryCollection_EqualsExact(t *testing.T) {
	gc1 := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		geom.NewLineStringXY(0, 0, 10, 10),
	})

	gc2 := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		geom.NewLineStringXY(0, 0, 10, 10),
	})

	gc3 := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		geom.NewLineStringXY(0, 0, 20, 20),
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
		geom.NewLineStringXY(20, 10, 30, 80),
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

	// Test that adding invalidates envelope cache
	gc.Add(geom.NewPoint(200, 200))
	env3 := gc.Envelope()
	if env3.MaxX != 200 || env3.MaxY != 200 {
		t.Error("Envelope should be recalculated after adding geometry")
	}
}

// TestGeometryCollection_String tests WKT output
func TestGeometryCollection_String(t *testing.T) {
	gc := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		geom.NewLineStringXY(0, 0, 10, 10),
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
		geom.NewLineStringXY(0, 0, 10, 10),
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
		geom.NewLineStringXY(0, 0, 10, 10),                                            // 1D
		geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil), // 2D
	})

	// Should return max dimension
	if gc.Dimension() != geom.DimensionArea {
		t.Errorf("Expected dimension %d, got %d", geom.DimensionArea, gc.Dimension())
	}

	if gc.NumGeometries() != 3 {
		t.Errorf("Expected 3 geometries, got %d", gc.NumGeometries())
	}
}
