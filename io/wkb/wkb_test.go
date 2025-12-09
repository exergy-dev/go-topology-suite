package wkb

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/go-topology-suite/gts/geom"
)

func TestMarshalUnmarshalPoint(t *testing.T) {
	factory := geom.DefaultFactory
	p := factory.CreatePoint(1.5, 2.5)

	// Write to WKB
	data, err := Marshal(p)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Expected non-empty WKB output")
	}

	// Read back
	g, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	point, ok := g.(*geom.Point)
	if !ok {
		t.Fatalf("Expected Point, got %T", g)
	}

	coord := point.Coordinate()
	if coord.X != 1.5 || coord.Y != 2.5 {
		t.Errorf("Expected (1.5, 2.5), got (%v, %v)", coord.X, coord.Y)
	}
}

func TestMarshalUnmarshalLineString(t *testing.T) {
	factory := geom.DefaultFactory
	coords := geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(1, 1),
		geom.NewCoordinate(2, 0),
	}
	ls := factory.CreateLineString(coords)

	// Write to WKB
	data, err := Marshal(ls)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Expected non-empty WKB output")
	}

	// Read back
	g, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	lineString, ok := g.(*geom.LineString)
	if !ok {
		t.Fatalf("Expected LineString, got %T", g)
	}

	readCoords := lineString.Coordinates()
	if len(readCoords) != 3 {
		t.Errorf("Expected 3 coordinates, got %d", len(readCoords))
	}
}

func TestMarshalUnmarshalPolygon(t *testing.T) {
	factory := geom.DefaultFactory
	shellCoords := geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(0, 0),
	}
	shell := factory.CreateLinearRing(shellCoords)
	poly := factory.CreatePolygon(shell, nil)

	// Write to WKB
	data, err := Marshal(poly)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Expected non-empty WKB output")
	}

	// Read back
	g, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	polygon, ok := g.(*geom.Polygon)
	if !ok {
		t.Fatalf("Expected Polygon, got %T", g)
	}

	if polygon.IsEmpty() {
		t.Error("Expected non-empty polygon")
	}

	extRing := polygon.ExteriorRing()
	if len(extRing.Coordinates()) != 5 {
		t.Errorf("Expected 5 coordinates in exterior ring, got %d", len(extRing.Coordinates()))
	}
}

func TestMarshalUnmarshalMultiPoint(t *testing.T) {
	factory := geom.DefaultFactory
	points := []*geom.Point{
		factory.CreatePoint(1, 2),
		factory.CreatePoint(3, 4),
		factory.CreatePoint(5, 6),
	}
	mp := factory.CreateMultiPoint(points)

	// Write to WKB
	data, err := Marshal(mp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Expected non-empty WKB output")
	}

	// Read back
	g, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	multiPoint, ok := g.(*geom.MultiPoint)
	if !ok {
		t.Fatalf("Expected MultiPoint, got %T", g)
	}

	if multiPoint.NumGeometries() != 3 {
		t.Errorf("Expected 3 points, got %d", multiPoint.NumGeometries())
	}
}

func TestMarshalUnmarshalGeometryCollection(t *testing.T) {
	factory := geom.DefaultFactory
	geoms := []geom.Geometry{
		factory.CreatePoint(1, 2),
		factory.CreateLineString(geom.CoordinateSequence{
			geom.NewCoordinate(0, 0),
			geom.NewCoordinate(1, 1),
		}),
	}
	gc := factory.CreateGeometryCollection(geoms)

	// Write to WKB
	data, err := Marshal(gc)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Expected non-empty WKB output")
	}

	// Read back
	g, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	collection, ok := g.(*geom.GeometryCollection)
	if !ok {
		t.Fatalf("Expected GeometryCollection, got %T", g)
	}

	if collection.NumGeometries() != 2 {
		t.Errorf("Expected 2 geometries, got %d", collection.NumGeometries())
	}
}

func TestEWKBWithSRID(t *testing.T) {
	factory := geom.NewGeometryFactoryWithSRID(4326)
	p := factory.CreatePoint(1.5, 2.5)

	// Write to EWKB with SRID
	data, err := MarshalEWKB(p)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Expected non-empty EWKB output")
	}

	// Read back
	g, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	point, ok := g.(*geom.Point)
	if !ok {
		t.Fatalf("Expected Point, got %T", g)
	}

	if point.SRID() != 4326 {
		t.Errorf("Expected SRID 4326, got %d", point.SRID())
	}
}

func TestByteOrderBigEndian(t *testing.T) {
	factory := geom.DefaultFactory
	p := factory.CreatePoint(1.5, 2.5)

	// Write with big endian
	opts := Options{
		ByteOrder:       binary.BigEndian,
		OutputDimension: 2,
	}
	data, err := MarshalWithOptions(p, opts)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// First byte should be XDR (big endian)
	if data[0] != wkbXDR {
		t.Errorf("Expected XDR byte order marker, got %d", data[0])
	}

	// Read back
	g, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	point, ok := g.(*geom.Point)
	if !ok {
		t.Fatalf("Expected Point, got %T", g)
	}

	coord := point.Coordinate()
	if coord.X != 1.5 || coord.Y != 2.5 {
		t.Errorf("Expected (1.5, 2.5), got (%v, %v)", coord.X, coord.Y)
	}
}

func TestEmptyGeometries(t *testing.T) {
	factory := geom.DefaultFactory

	// Test empty point
	emptyPoint := factory.CreatePointEmpty()
	data, err := Marshal(emptyPoint)
	if err != nil {
		t.Fatalf("Failed to marshal empty point: %v", err)
	}
	g, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal empty point: %v", err)
	}
	if !g.IsEmpty() {
		t.Error("Expected empty point")
	}

	// Test empty polygon
	emptyPoly := factory.CreatePolygonEmpty()
	data, err = Marshal(emptyPoly)
	if err != nil {
		t.Fatalf("Failed to marshal empty polygon: %v", err)
	}
	g, err = Unmarshal(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal empty polygon: %v", err)
	}
	if !g.IsEmpty() {
		t.Error("Expected empty polygon")
	}
}

func TestInvalidWKB(t *testing.T) {
	// Too short
	_, err := Unmarshal([]byte{1})
	if err == nil {
		t.Error("Expected error for too short WKB")
	}

	// Invalid byte order
	_, err = Unmarshal([]byte{5, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	if err == nil {
		t.Error("Expected error for invalid byte order")
	}

	// Empty data
	_, err = Unmarshal([]byte{})
	if err == nil {
		t.Error("Expected error for empty data")
	}
}

func TestMarshalUnmarshalPolygonWithHole(t *testing.T) {
	factory := geom.DefaultFactory

	// Outer ring
	shellCoords := geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(20, 0),
		geom.NewCoordinate(20, 20),
		geom.NewCoordinate(0, 20),
		geom.NewCoordinate(0, 0),
	}
	shell := factory.CreateLinearRing(shellCoords)

	// Hole
	holeCoords := geom.CoordinateSequence{
		geom.NewCoordinate(5, 5),
		geom.NewCoordinate(15, 5),
		geom.NewCoordinate(15, 15),
		geom.NewCoordinate(5, 15),
		geom.NewCoordinate(5, 5),
	}
	hole := factory.CreateLinearRing(holeCoords)

	poly := factory.CreatePolygon(shell, []*geom.LinearRing{hole})

	// Write to WKB
	data, err := Marshal(poly)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Read back
	g, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	polygon, ok := g.(*geom.Polygon)
	if !ok {
		t.Fatalf("Expected Polygon, got %T", g)
	}

	if polygon.NumInteriorRings() != 1 {
		t.Errorf("Expected 1 interior ring, got %d", polygon.NumInteriorRings())
	}
}

func TestRoundTrip(t *testing.T) {
	factory := geom.DefaultFactory

	testCases := []struct {
		name string
		geom geom.Geometry
	}{
		{"Point", factory.CreatePoint(1.5, 2.5)},
		{"LineString", factory.CreateLineString(geom.CoordinateSequence{
			geom.NewCoordinate(0, 0),
			geom.NewCoordinate(1, 1),
			geom.NewCoordinate(2, 0),
		})},
		{"Polygon", factory.CreatePolygon(
			factory.CreateLinearRing(geom.CoordinateSequence{
				geom.NewCoordinate(0, 0),
				geom.NewCoordinate(10, 0),
				geom.NewCoordinate(10, 10),
				geom.NewCoordinate(0, 10),
				geom.NewCoordinate(0, 0),
			}), nil)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := Marshal(tc.geom)
			if err != nil {
				t.Fatalf("Failed to marshal: %v", err)
			}
			g, err := Unmarshal(data)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			// Write again and compare
			data2, err := Marshal(g)
			if err != nil {
				t.Fatalf("Failed to re-marshal: %v", err)
			}
			if !bytes.Equal(data, data2) {
				t.Error("Round-trip WKB mismatch")
			}
		})
	}
}

func TestUnmarshalWithFactory(t *testing.T) {
	factory := geom.NewGeometryFactoryWithSRID(4326)
	p := factory.CreatePoint(1, 2)

	data, err := Marshal(p)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	g, err := UnmarshalWithFactory(data, factory)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if g == nil {
		t.Error("Expected non-nil geometry")
	}
}

func TestMarshalWithOptions(t *testing.T) {
	factory := geom.DefaultFactory
	p := factory.CreatePoint(1, 2)

	opts := Options{
		ByteOrder:       binary.LittleEndian,
		OutputDimension: 2,
	}

	data, err := MarshalWithOptions(p, opts)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty data")
	}

	// First byte should be NDR (little endian)
	if data[0] != wkbNDR {
		t.Errorf("Expected NDR byte order marker, got %d", data[0])
	}
}

func BenchmarkMarshalPoint(b *testing.B) {
	p := geom.NewPoint(1.5, 2.5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Marshal(p)
	}
}

func BenchmarkUnmarshalPoint(b *testing.B) {
	p := geom.NewPoint(1.5, 2.5)
	data, _ := Marshal(p)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Unmarshal(data)
	}
}

func BenchmarkMarshalPolygon(b *testing.B) {
	factory := geom.DefaultFactory
	shell := factory.CreateLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(0, 0),
	})
	poly := factory.CreatePolygon(shell, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Marshal(poly)
	}
}

func BenchmarkUnmarshalPolygon(b *testing.B) {
	factory := geom.DefaultFactory
	shell := factory.CreateLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(0, 0),
	})
	poly := factory.CreatePolygon(shell, nil)
	data, _ := Marshal(poly)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Unmarshal(data)
	}
}
