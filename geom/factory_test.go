package geom_test

import (
	"math"
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

// TestGeometryFactory_Default tests default factory
func TestGeometryFactory_Default(t *testing.T) {
	factory := geom.NewGeometryFactoryDefault()

	if factory.SRID() != 0 {
		t.Errorf("Default factory SRID should be 0, got %d", factory.SRID())
	}

	pm := factory.PrecisionModel()
	if pm == nil {
		t.Error("Default factory should have a precision model")
	}
}

// TestGeometryFactory_WithSRID tests factory with custom SRID
func TestGeometryFactory_WithSRID(t *testing.T) {
	factory := geom.NewGeometryFactoryWithSRID(4326)

	if factory.SRID() != 4326 {
		t.Errorf("Expected SRID 4326, got %d", factory.SRID())
	}

	p := factory.CreatePoint(10, 20)
	if p.SRID() != 4326 {
		t.Errorf("Created point should have SRID 4326, got %d", p.SRID())
	}
}

// TestGeometryFactory_WithPrecision tests factory with custom precision
func TestGeometryFactory_WithPrecision(t *testing.T) {
	pm := geom.NewFixedPrecision(100) // 2 decimal places
	factory := geom.NewGeometryFactory(pm, 0)

	p := factory.CreatePoint(1.23456, 2.34567)

	// Should be rounded to 2 decimal places
	if math.Abs(p.X()-1.23) > 0.001 {
		t.Errorf("Expected X=1.23, got %f", p.X())
	}
	if math.Abs(p.Y()-2.35) > 0.001 {
		t.Errorf("Expected Y=2.35, got %f", p.Y())
	}
}

// TestGeometryFactory_FullConstructor tests factory with precision and SRID
func TestGeometryFactory_FullConstructor(t *testing.T) {
	pm := geom.NewFixedPrecision(10)
	factory := geom.NewGeometryFactory(pm, 4326)

	if factory.SRID() != 4326 {
		t.Errorf("Expected SRID 4326, got %d", factory.SRID())
	}

	if factory.PrecisionModel() != pm {
		t.Error("Factory should use provided precision model")
	}
}

// TestGeometryFactory_NilPrecisionModel tests factory with nil precision model
func TestGeometryFactory_NilPrecisionModel(t *testing.T) {
	factory := geom.NewGeometryFactory(nil, 0)

	// Should use default floating precision
	if factory.PrecisionModel() == nil {
		t.Error("Factory should have default precision model when nil provided")
	}
}

// TestGeometryFactory_CreatePoint tests point creation
func TestGeometryFactory_CreatePoint(t *testing.T) {
	factory := geom.NewGeometryFactoryWithSRID(4326)

	p := factory.CreatePoint(10, 20)

	if p.X() != 10 || p.Y() != 20 {
		t.Errorf("Expected point (10, 20), got (%f, %f)", p.X(), p.Y())
	}

	if p.SRID() != 4326 {
		t.Errorf("Expected SRID 4326, got %d", p.SRID())
	}
}

// TestGeometryFactory_CreatePointFromCoordinate tests point creation from coordinate
func TestGeometryFactory_CreatePointFromCoordinate(t *testing.T) {
	factory := geom.NewGeometryFactoryWithSRID(4326)
	coord := geom.NewCoordinate(10, 20)

	p := factory.CreatePointFromCoordinate(coord)

	if p.X() != 10 || p.Y() != 20 {
		t.Errorf("Expected point (10, 20), got (%f, %f)", p.X(), p.Y())
	}

	if p.SRID() != 4326 {
		t.Errorf("Expected SRID 4326, got %d", p.SRID())
	}

	// Verify coordinate was cloned
	coord.X = 999
	if p.X() == 999 {
		t.Error("Factory should clone coordinate, not reference it")
	}
}

// TestGeometryFactory_CreatePointEmpty tests empty point creation
func TestGeometryFactory_CreatePointEmpty(t *testing.T) {
	factory := geom.NewGeometryFactoryWithSRID(4326)

	p := factory.CreatePointEmpty()

	if !p.IsEmpty() {
		t.Error("Created point should be empty")
	}

	if p.SRID() != 4326 {
		t.Errorf("Expected SRID 4326, got %d", p.SRID())
	}
}

// TestGeometryFactory_CreateLineString tests linestring creation
func TestGeometryFactory_CreateLineString(t *testing.T) {
	factory := geom.NewGeometryFactoryWithSRID(4326)
	coords := mustCoordsXY(0, 0, 10, 10, 20, 0)

	ls := factory.CreateLineString(coords)

	if ls.NumPoints() != 3 {
		t.Errorf("Expected 3 points, got %d", ls.NumPoints())
	}

	if ls.SRID() != 4326 {
		t.Errorf("Expected SRID 4326, got %d", ls.SRID())
	}
}

// TestGeometryFactory_CreateLineStringFromXYPairs tests linestring creation from xy pairs
func TestGeometryFactory_CreateLineStringFromXYPairs(t *testing.T) {
	factory := geom.NewGeometryFactoryDefault()

	ls := mustCreateLineStringXY(factory, 0, 0, 10, 10, 20, 0)

	if ls.NumPoints() != 3 {
		t.Errorf("Expected 3 points, got %d", ls.NumPoints())
	}

	expected := math.Sqrt(200) + math.Sqrt(200) // sqrt(10^2+10^2) + sqrt(10^2+10^2)
	if math.Abs(ls.Length()-expected) > 0.01 {
		t.Errorf("Expected length ~%f, got %f", expected, ls.Length())
	}
}

// TestGeometryFactory_CreateLineStringEmpty tests empty linestring creation
func TestGeometryFactory_CreateLineStringEmpty(t *testing.T) {
	factory := geom.NewGeometryFactoryWithSRID(4326)

	ls := factory.CreateLineStringEmpty()

	if !ls.IsEmpty() {
		t.Error("Created linestring should be empty")
	}

	if ls.SRID() != 4326 {
		t.Errorf("Expected SRID 4326, got %d", ls.SRID())
	}
}

// TestGeometryFactory_CreateLinearRing tests linear ring creation
func TestGeometryFactory_CreateLinearRing(t *testing.T) {
	factory := geom.NewGeometryFactoryWithSRID(4326)
	coords := mustCoordsXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)

	lr := factory.CreateLinearRing(coords)

	if !lr.IsClosed() {
		t.Error("Linear ring should be closed")
	}

	if lr.SRID() != 4326 {
		t.Errorf("Expected SRID 4326, got %d", lr.SRID())
	}
}

// TestGeometryFactory_CreateLinearRingFromXYPairs tests linear ring creation from xy pairs
func TestGeometryFactory_CreateLinearRingFromXYPairs(t *testing.T) {
	factory := geom.NewGeometryFactoryDefault()

	lr := mustCreateLinearRingXY(factory, 0, 0, 10, 0, 10, 10, 0, 10, 0, 0)

	if !lr.IsClosed() {
		t.Error("Linear ring should be closed")
	}
}

// TestGeometryFactory_CreateLinearRingEmpty tests empty linear ring creation
func TestGeometryFactory_CreateLinearRingEmpty(t *testing.T) {
	factory := geom.NewGeometryFactoryWithSRID(4326)

	lr := factory.CreateLinearRingEmpty()

	if !lr.IsEmpty() {
		t.Error("Created linear ring should be empty")
	}

	if lr.SRID() != 4326 {
		t.Errorf("Expected SRID 4326, got %d", lr.SRID())
	}
}

// TestGeometryFactory_CreatePolygon tests polygon creation
func TestGeometryFactory_CreatePolygon(t *testing.T) {
	factory := geom.NewGeometryFactoryWithSRID(4326)
	shell := mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	hole := mustLinearRingXY(2, 2, 8, 2, 8, 8, 2, 8, 2, 2)

	poly := factory.CreatePolygon(shell, []*geom.LinearRing{hole})

	if poly.IsEmpty() {
		t.Error("Polygon should not be empty")
	}

	if poly.SRID() != 4326 {
		t.Errorf("Expected SRID 4326, got %d", poly.SRID())
	}

	area := poly.Area()
	expected := 100.0 - 36.0
	if math.Abs(area-expected) > 0.001 {
		t.Errorf("Expected area %f, got %f", expected, area)
	}
}

// TestGeometryFactory_CreatePolygonFromCoords tests polygon creation from coordinates
func TestGeometryFactory_CreatePolygonFromCoords(t *testing.T) {
	factory := geom.NewGeometryFactoryDefault()
	shell := mustCoordsXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	hole := mustCoordsXY(2, 2, 8, 2, 8, 8, 2, 8, 2, 2)

	poly := factory.CreatePolygonFromCoords(shell, hole)

	area := poly.Area()
	expected := 100.0 - 36.0
	if math.Abs(area-expected) > 0.001 {
		t.Errorf("Expected area %f, got %f", expected, area)
	}
}

// TestGeometryFactory_CreatePolygonEmpty tests empty polygon creation
func TestGeometryFactory_CreatePolygonEmpty(t *testing.T) {
	factory := geom.NewGeometryFactoryWithSRID(4326)

	poly := factory.CreatePolygonEmpty()

	if !poly.IsEmpty() {
		t.Error("Created polygon should be empty")
	}

	if poly.SRID() != 4326 {
		t.Errorf("Expected SRID 4326, got %d", poly.SRID())
	}
}

// TestGeometryFactory_CreateMultiPoint tests multipoint creation
func TestGeometryFactory_CreateMultiPoint(t *testing.T) {
	factory := geom.NewGeometryFactoryWithSRID(4326)
	points := []*geom.Point{
		geom.NewPoint(0, 0),
		geom.NewPoint(10, 10),
	}

	mp := factory.CreateMultiPoint(points)

	if mp.NumGeometries() != 2 {
		t.Errorf("Expected 2 points, got %d", mp.NumGeometries())
	}

	if mp.SRID() != 4326 {
		t.Errorf("Expected SRID 4326, got %d", mp.SRID())
	}
}

// TestGeometryFactory_CreateMultiPointFromCoords tests multipoint creation from coordinates
func TestGeometryFactory_CreateMultiPointFromCoords(t *testing.T) {
	factory := geom.NewGeometryFactoryDefault()
	coords := mustCoordsXY(0, 0, 10, 10, 20, 20)

	mp := factory.CreateMultiPointFromCoords(coords)

	if mp.NumGeometries() != 3 {
		t.Errorf("Expected 3 points, got %d", mp.NumGeometries())
	}
}

// TestGeometryFactory_CreateMultiPointEmpty tests empty multipoint creation
func TestGeometryFactory_CreateMultiPointEmpty(t *testing.T) {
	factory := geom.NewGeometryFactoryWithSRID(4326)

	mp := factory.CreateMultiPointEmpty()

	if !mp.IsEmpty() {
		t.Error("Created multipoint should be empty")
	}

	if mp.SRID() != 4326 {
		t.Errorf("Expected SRID 4326, got %d", mp.SRID())
	}
}

// TestGeometryFactory_CreateMultiLineString tests multilinestring creation
func TestGeometryFactory_CreateMultiLineString(t *testing.T) {
	factory := geom.NewGeometryFactoryWithSRID(4326)
	lines := []*geom.LineString{
		mustLineStringXY(0, 0, 10, 0),
		mustLineStringXY(20, 0, 30, 0),
	}

	mls := factory.CreateMultiLineString(lines)

	if mls.NumGeometries() != 2 {
		t.Errorf("Expected 2 lines, got %d", mls.NumGeometries())
	}

	if mls.SRID() != 4326 {
		t.Errorf("Expected SRID 4326, got %d", mls.SRID())
	}
}

// TestGeometryFactory_CreateMultiLineStringEmpty tests empty multilinestring creation
func TestGeometryFactory_CreateMultiLineStringEmpty(t *testing.T) {
	factory := geom.NewGeometryFactoryWithSRID(4326)

	mls := factory.CreateMultiLineStringEmpty()

	if !mls.IsEmpty() {
		t.Error("Created multilinestring should be empty")
	}

	if mls.SRID() != 4326 {
		t.Errorf("Expected SRID 4326, got %d", mls.SRID())
	}
}

// TestGeometryFactory_CreateMultiPolygon tests multipolygon creation
func TestGeometryFactory_CreateMultiPolygon(t *testing.T) {
	factory := geom.NewGeometryFactoryWithSRID(4326)
	polys := []*geom.Polygon{
		geom.NewPolygon(mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
		geom.NewPolygon(mustLinearRingXY(20, 20, 30, 20, 30, 30, 20, 30, 20, 20), nil),
	}

	mp := factory.CreateMultiPolygon(polys)

	if mp.NumGeometries() != 2 {
		t.Errorf("Expected 2 polygons, got %d", mp.NumGeometries())
	}

	if mp.SRID() != 4326 {
		t.Errorf("Expected SRID 4326, got %d", mp.SRID())
	}
}

// TestGeometryFactory_CreateMultiPolygonEmpty tests empty multipolygon creation
func TestGeometryFactory_CreateMultiPolygonEmpty(t *testing.T) {
	factory := geom.NewGeometryFactoryWithSRID(4326)

	mp := factory.CreateMultiPolygonEmpty()

	if !mp.IsEmpty() {
		t.Error("Created multipolygon should be empty")
	}

	if mp.SRID() != 4326 {
		t.Errorf("Expected SRID 4326, got %d", mp.SRID())
	}
}

// TestGeometryFactory_CreateGeometryCollection tests geometry collection creation
func TestGeometryFactory_CreateGeometryCollection(t *testing.T) {
	factory := geom.NewGeometryFactoryWithSRID(4326)
	geoms := []geom.Geometry{
		geom.NewPoint(0, 0),
		mustLineStringXY(0, 0, 10, 10),
	}

	gc := factory.CreateGeometryCollection(geoms)

	if gc.NumGeometries() != 2 {
		t.Errorf("Expected 2 geometries, got %d", gc.NumGeometries())
	}

	if gc.SRID() != 4326 {
		t.Errorf("Expected SRID 4326, got %d", gc.SRID())
	}
}

// TestGeometryFactory_CreateGeometryCollectionEmpty tests empty geometry collection creation
func TestGeometryFactory_CreateGeometryCollectionEmpty(t *testing.T) {
	factory := geom.NewGeometryFactoryWithSRID(4326)

	gc := factory.CreateGeometryCollectionEmpty()

	if !gc.IsEmpty() {
		t.Error("Created geometry collection should be empty")
	}

	if gc.SRID() != 4326 {
		t.Errorf("Expected SRID 4326, got %d", gc.SRID())
	}
}

// TestGeometryFactory_PrecisionApplication tests precision model application
func TestGeometryFactory_PrecisionApplication(t *testing.T) {
	pm := geom.NewFixedPrecision(10) // 1 decimal place
	factory := geom.NewGeometryFactory(pm, 0)

	// Test point
	p := factory.CreatePoint(1.23456, 2.34567)
	if math.Abs(p.X()-1.2) > 0.01 || math.Abs(p.Y()-2.3) > 0.01 {
		t.Errorf("Point precision not applied: (%f, %f)", p.X(), p.Y())
	}

	// Test linestring
	ls := mustCreateLineStringXY(factory, 1.23456, 2.34567, 3.45678, 4.56789)
	coords := ls.Coordinates()
	if math.Abs(coords[0].X-1.2) > 0.01 {
		t.Errorf("LineString precision not applied: %f", coords[0].X)
	}

	// Test polygon
	shell := mustLinearRingXY(1.23, 2.34, 10.56, 2.34, 10.56, 10.78, 1.23, 10.78, 1.23, 2.34)
	poly := factory.CreatePolygon(shell, nil)
	polyCoords := poly.Coordinates()
	if math.Abs(polyCoords[0].X-1.2) > 0.01 {
		t.Errorf("Polygon precision not applied: %f", polyCoords[0].X)
	}
}

// TestGeometryFactory_DefaultFactory tests the default factory singleton
func TestGeometryFactory_DefaultFactory(t *testing.T) {
	factory := geom.DefaultFactory

	if factory == nil {
		t.Error("DefaultFactory should not be nil")
	}

	if factory.SRID() != 0 {
		t.Errorf("DefaultFactory SRID should be 0, got %d", factory.SRID())
	}

	// Should be able to create geometries
	p := factory.CreatePoint(1, 2)
	if p.X() != 1 || p.Y() != 2 {
		t.Error("DefaultFactory should create valid geometries")
	}
}

// TestGeometryFactory_CoordinateCloning tests that factories clone coordinates
func TestGeometryFactory_CoordinateCloning(t *testing.T) {
	factory := geom.NewGeometryFactoryDefault()

	coords := mustCoordsXY(0, 0, 10, 10)
	ls := factory.CreateLineString(coords)

	// Modify original
	coords[0].X = 999

	// Check that geometry wasn't affected
	if ls.Coordinates()[0].X == 999 {
		t.Error("Factory should clone coordinates, not reference them")
	}
}
