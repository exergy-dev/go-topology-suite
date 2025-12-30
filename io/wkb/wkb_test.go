package wkb

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/go-topology-suite/gts/geom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalUnmarshalPoint(t *testing.T) {
	factory := geom.DefaultFactory
	p := factory.CreatePoint(1.5, 2.5)

	// Write to WKB
	data, err := Marshal(p)
	require.NoError(t, err, "Failed to marshal")
	require.NotEmpty(t, data, "Expected non-empty WKB output")

	// Read back
	g, err := Unmarshal(data)
	require.NoError(t, err, "Failed to unmarshal")

	point, ok := g.(*geom.Point)
	require.True(t, ok, "Expected Point, got %T", g)

	coord := point.Coordinate()
	assert.Equal(t, 1.5, coord.X)
	assert.Equal(t, 2.5, coord.Y)
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
	require.NoError(t, err, "Failed to marshal")
	require.NotEmpty(t, data, "Expected non-empty WKB output")

	// Read back
	g, err := Unmarshal(data)
	require.NoError(t, err, "Failed to unmarshal")

	lineString, ok := g.(*geom.LineString)
	require.True(t, ok, "Expected LineString, got %T", g)

	readCoords := lineString.Coordinates()
	assert.Len(t, readCoords, 3, "Expected 3 coordinates")
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
	require.NoError(t, err, "Failed to marshal")
	require.NotEmpty(t, data, "Expected non-empty WKB output")

	// Read back
	g, err := Unmarshal(data)
	require.NoError(t, err, "Failed to unmarshal")

	polygon, ok := g.(*geom.Polygon)
	require.True(t, ok, "Expected Polygon, got %T", g)

	assert.False(t, polygon.IsEmpty(), "Expected non-empty polygon")

	extRing := polygon.ExteriorRing()
	assert.Len(t, extRing.Coordinates(), 5, "Expected 5 coordinates in exterior ring")
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
	require.NoError(t, err, "Failed to marshal")
	require.NotEmpty(t, data, "Expected non-empty WKB output")

	// Read back
	g, err := Unmarshal(data)
	require.NoError(t, err, "Failed to unmarshal")

	multiPoint, ok := g.(*geom.MultiPoint)
	require.True(t, ok, "Expected MultiPoint, got %T", g)

	assert.Equal(t, 3, multiPoint.NumGeometries(), "Expected 3 points")
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
	require.NoError(t, err, "Failed to marshal")
	require.NotEmpty(t, data, "Expected non-empty WKB output")

	// Read back
	g, err := Unmarshal(data)
	require.NoError(t, err, "Failed to unmarshal")

	collection, ok := g.(*geom.GeometryCollection)
	require.True(t, ok, "Expected GeometryCollection, got %T", g)

	assert.Equal(t, 2, collection.NumGeometries(), "Expected 2 geometries")
}

func TestEWKBWithSRID(t *testing.T) {
	factory := geom.NewGeometryFactoryWithSRID(4326)
	p := factory.CreatePoint(1.5, 2.5)

	// Write to EWKB with SRID
	data, err := MarshalEWKB(p)
	require.NoError(t, err, "Failed to marshal")
	require.NotEmpty(t, data, "Expected non-empty EWKB output")

	// Read back
	g, err := Unmarshal(data)
	require.NoError(t, err, "Failed to unmarshal")

	point, ok := g.(*geom.Point)
	require.True(t, ok, "Expected Point, got %T", g)

	assert.Equal(t, 4326, point.SRID(), "Expected SRID 4326")
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
	require.NoError(t, err, "Failed to marshal")

	// First byte should be XDR (big endian)
	assert.Equal(t, byte(wkbXDR), data[0], "Expected XDR byte order marker")

	// Read back
	g, err := Unmarshal(data)
	require.NoError(t, err, "Failed to unmarshal")

	point, ok := g.(*geom.Point)
	require.True(t, ok, "Expected Point, got %T", g)

	coord := point.Coordinate()
	assert.Equal(t, 1.5, coord.X)
	assert.Equal(t, 2.5, coord.Y)
}

func TestEmptyGeometries(t *testing.T) {
	factory := geom.DefaultFactory

	// Test empty point
	emptyPoint := factory.CreatePointEmpty()
	data, err := Marshal(emptyPoint)
	require.NoError(t, err, "Failed to marshal empty point")
	g, err := Unmarshal(data)
	require.NoError(t, err, "Failed to unmarshal empty point")
	assert.True(t, g.IsEmpty(), "Expected empty point")

	// Test empty polygon
	emptyPoly := factory.CreatePolygonEmpty()
	data, err = Marshal(emptyPoly)
	require.NoError(t, err, "Failed to marshal empty polygon")
	g, err = Unmarshal(data)
	require.NoError(t, err, "Failed to unmarshal empty polygon")
	assert.True(t, g.IsEmpty(), "Expected empty polygon")
}

func TestInvalidWKB(t *testing.T) {
	// Too short
	_, err := Unmarshal([]byte{1})
	assert.Error(t, err, "Expected error for too short WKB")

	// Invalid byte order
	_, err = Unmarshal([]byte{5, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	assert.Error(t, err, "Expected error for invalid byte order")

	// Empty data
	_, err = Unmarshal([]byte{})
	assert.Error(t, err, "Expected error for empty data")
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
	require.NoError(t, err, "Failed to marshal")

	// Read back
	g, err := Unmarshal(data)
	require.NoError(t, err, "Failed to unmarshal")

	polygon, ok := g.(*geom.Polygon)
	require.True(t, ok, "Expected Polygon, got %T", g)

	assert.Equal(t, 1, polygon.NumInteriorRings(), "Expected 1 interior ring")
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
			require.NoError(t, err, "Failed to marshal")
			g, err := Unmarshal(data)
			require.NoError(t, err, "Failed to unmarshal")

			// Write again and compare
			data2, err := Marshal(g)
			require.NoError(t, err, "Failed to re-marshal")
			assert.True(t, bytes.Equal(data, data2), "Round-trip WKB mismatch")
		})
	}
}

func TestUnmarshalWithFactory(t *testing.T) {
	factory := geom.NewGeometryFactoryWithSRID(4326)
	p := factory.CreatePoint(1, 2)

	data, err := Marshal(p)
	require.NoError(t, err, "Failed to marshal")

	g, err := UnmarshalWithFactory(data, factory)
	require.NoError(t, err, "Failed to unmarshal")

	assert.NotNil(t, g, "Expected non-nil geometry")
}

func TestMarshalWithOptions(t *testing.T) {
	factory := geom.DefaultFactory
	p := factory.CreatePoint(1, 2)

	opts := Options{
		ByteOrder:       binary.LittleEndian,
		OutputDimension: 2,
	}

	data, err := MarshalWithOptions(p, opts)
	require.NoError(t, err, "Failed to marshal")

	assert.NotEmpty(t, data, "Expected non-empty data")

	// First byte should be NDR (little endian)
	assert.Equal(t, byte(wkbNDR), data[0], "Expected NDR byte order marker")
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
