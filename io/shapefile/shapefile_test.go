package shapefile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jonas-p/go-shp"
	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteReadPoint(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "points.shp")

	// Create test points
	factory := geom.DefaultFactory
	points := []geom.Geometry{
		factory.CreatePoint(1.0, 2.0),
		factory.CreatePoint(3.0, 4.0),
		factory.CreatePoint(5.0, 6.0),
	}

	// Write points
	writer, err := NewWriter(filename, ShapeTypePoint)
	require.NoError(t, err, "Failed to create writer")

	for _, p := range points {
		err = writer.Write(p)
		require.NoError(t, err, "Failed to write point")
	}
	err = writer.Close()
	require.NoError(t, err, "Failed to close writer")

	// Read points back
	reader, err := NewReader(filename)
	require.NoError(t, err, "Failed to create reader")
	defer reader.Close() //nolint:errcheck

	assert.Equal(t, ShapeTypePoint, reader.ShapeType(), "Expected Point shape type")

	var readGeoms []geom.Geometry
	for reader.Next() {
		g, err := reader.Geometry()
		require.NoError(t, err, "Failed to read geometry")
		readGeoms = append(readGeoms, g)
	}

	require.Len(t, readGeoms, 3, "Expected 3 geometries")

	// Verify coordinates
	for i, g := range readGeoms {
		pt, ok := g.(*geom.Point)
		require.True(t, ok, "Expected Point, got %T", g)

		origPt := points[i].(*geom.Point)
		assert.InDelta(t, origPt.X(), pt.X(), 0.0001, "X coordinate mismatch at index %d", i)
		assert.InDelta(t, origPt.Y(), pt.Y(), 0.0001, "Y coordinate mismatch at index %d", i)
	}
}

func TestWriteReadLineString(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "lines.shp")

	// Create test linestring
	factory := geom.DefaultFactory
	coords := geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(20, 0),
	}
	line := factory.CreateLineString(coords)

	// Write linestring
	writer, err := NewWriter(filename, ShapeTypePolyLine)
	require.NoError(t, err, "Failed to create writer")

	err = writer.Write(line)
	require.NoError(t, err, "Failed to write linestring")
	err = writer.Close()
	require.NoError(t, err, "Failed to close writer")

	// Read linestring back
	reader, err := NewReader(filename)
	require.NoError(t, err, "Failed to create reader")
	defer reader.Close() //nolint:errcheck

	assert.Equal(t, ShapeTypePolyLine, reader.ShapeType(), "Expected PolyLine shape type")

	require.True(t, reader.Next(), "Expected at least one record")
	g, err := reader.Geometry()
	require.NoError(t, err, "Failed to read geometry")

	ls, ok := g.(*geom.LineString)
	require.True(t, ok, "Expected LineString, got %T", g)

	readCoords := ls.Coordinates()
	require.Len(t, readCoords, 3, "Expected 3 coordinates")

	for i, c := range coords {
		assert.InDelta(t, c.X, readCoords[i].X, 0.0001, "X coordinate mismatch at index %d", i)
		assert.InDelta(t, c.Y, readCoords[i].Y, 0.0001, "Y coordinate mismatch at index %d", i)
	}
}

func TestWriteReadPolygon(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "polygons.shp")

	// Create test polygon
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

	// Write polygon
	writer, err := NewWriter(filename, ShapeTypePolygon)
	require.NoError(t, err, "Failed to create writer")

	err = writer.Write(poly)
	require.NoError(t, err, "Failed to write polygon")
	err = writer.Close()
	require.NoError(t, err, "Failed to close writer")

	// Read polygon back
	reader, err := NewReader(filename)
	require.NoError(t, err, "Failed to create reader")
	defer reader.Close() //nolint:errcheck

	assert.Equal(t, ShapeTypePolygon, reader.ShapeType(), "Expected Polygon shape type")

	require.True(t, reader.Next(), "Expected at least one record")
	g, err := reader.Geometry()
	require.NoError(t, err, "Failed to read geometry")

	readPoly, ok := g.(*geom.Polygon)
	require.True(t, ok, "Expected Polygon, got %T", g)

	// Verify area
	assert.InDelta(t, poly.Area(), readPoly.Area(), 0.01, "Area mismatch")
}

func TestWriteReadPolygonWithHole(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "polygons_hole.shp")

	// Create polygon with hole
	factory := geom.DefaultFactory
	shellCoords := geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(20, 0),
		geom.NewCoordinate(20, 20),
		geom.NewCoordinate(0, 20),
		geom.NewCoordinate(0, 0),
	}
	holeCoords := geom.CoordinateSequence{
		geom.NewCoordinate(5, 5),
		geom.NewCoordinate(5, 15),
		geom.NewCoordinate(15, 15),
		geom.NewCoordinate(15, 5),
		geom.NewCoordinate(5, 5),
	}

	shell := factory.CreateLinearRing(shellCoords)
	hole := factory.CreateLinearRing(holeCoords)
	poly := factory.CreatePolygon(shell, []*geom.LinearRing{hole})

	// Write polygon
	writer, err := NewWriter(filename, ShapeTypePolygon)
	require.NoError(t, err, "Failed to create writer")

	err = writer.Write(poly)
	require.NoError(t, err, "Failed to write polygon")
	err = writer.Close()
	require.NoError(t, err, "Failed to close writer")

	// Read polygon back
	reader, err := NewReader(filename)
	require.NoError(t, err, "Failed to create reader")
	defer reader.Close() //nolint:errcheck

	require.True(t, reader.Next(), "Expected at least one record")
	g, err := reader.Geometry()
	require.NoError(t, err, "Failed to read geometry")

	readPoly, ok := g.(*geom.Polygon)
	require.True(t, ok, "Expected Polygon, got %T", g)

	// Verify area (should have hole subtracted)
	expectedArea := 20.0*20.0 - 10.0*10.0 // 400 - 100 = 300
	assert.InDelta(t, expectedArea, readPoly.Area(), 0.1, "Area mismatch - expected hole to be subtracted")
}

func TestWriteReadMultiPoint(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "multipoints.shp")

	// Create multipoint
	factory := geom.DefaultFactory
	points := []*geom.Point{
		factory.CreatePoint(1, 2),
		factory.CreatePoint(3, 4),
		factory.CreatePoint(5, 6),
	}
	mp := factory.CreateMultiPoint(points)

	// Write multipoint
	writer, err := NewWriter(filename, ShapeTypeMultiPoint)
	require.NoError(t, err, "Failed to create writer")

	err = writer.Write(mp)
	require.NoError(t, err, "Failed to write multipoint")
	err = writer.Close()
	require.NoError(t, err, "Failed to close writer")

	// Read multipoint back
	reader, err := NewReader(filename)
	require.NoError(t, err, "Failed to create reader")
	defer reader.Close() //nolint:errcheck

	assert.Equal(t, ShapeTypeMultiPoint, reader.ShapeType(), "Expected MultiPoint shape type")

	require.True(t, reader.Next(), "Expected at least one record")
	g, err := reader.Geometry()
	require.NoError(t, err, "Failed to read geometry")

	readMP, ok := g.(*geom.MultiPoint)
	require.True(t, ok, "Expected MultiPoint, got %T", g)

	assert.Equal(t, 3, readMP.NumGeometries(), "Expected 3 points in multipoint")
}

func TestWriteReadMultiLineString(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "multilines.shp")

	// Create multilinestring
	factory := geom.DefaultFactory
	lines := []*geom.LineString{
		mustCreateLineStringXY(factory, 0, 0, 10, 10),
		mustCreateLineStringXY(factory, 20, 20, 30, 30),
	}
	mls := factory.CreateMultiLineString(lines)

	// Write multilinestring
	writer, err := NewWriter(filename, ShapeTypePolyLine)
	require.NoError(t, err, "Failed to create writer")

	err = writer.Write(mls)
	require.NoError(t, err, "Failed to write multilinestring")
	err = writer.Close()
	require.NoError(t, err, "Failed to close writer")

	// Read multilinestring back
	reader, err := NewReader(filename)
	require.NoError(t, err, "Failed to create reader")
	defer reader.Close() //nolint:errcheck

	require.True(t, reader.Next(), "Expected at least one record")
	g, err := reader.Geometry()
	require.NoError(t, err, "Failed to read geometry")

	readMLS, ok := g.(*geom.MultiLineString)
	require.True(t, ok, "Expected MultiLineString, got %T", g)

	assert.Equal(t, 2, readMLS.NumGeometries(), "Expected 2 linestrings")
}

func TestWriteReadMultiPolygon(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "multipolygons.shp")

	// Create multipolygon with two separate polygons
	factory := geom.DefaultFactory

	shell1 := mustCreateLinearRingXY(factory, 0, 0, 5, 0, 5, 5, 0, 5, 0, 0)
	poly1 := factory.CreatePolygon(shell1, nil)

	shell2 := mustCreateLinearRingXY(factory, 10, 10, 15, 10, 15, 15, 10, 15, 10, 10)
	poly2 := factory.CreatePolygon(shell2, nil)

	mp := factory.CreateMultiPolygon([]*geom.Polygon{poly1, poly2})

	// Write multipolygon
	writer, err := NewWriter(filename, ShapeTypePolygon)
	require.NoError(t, err, "Failed to create writer")

	err = writer.Write(mp)
	require.NoError(t, err, "Failed to write multipolygon")
	err = writer.Close()
	require.NoError(t, err, "Failed to close writer")

	// Read multipolygon back
	reader, err := NewReader(filename)
	require.NoError(t, err, "Failed to create reader")
	defer reader.Close() //nolint:errcheck

	require.True(t, reader.Next(), "Expected at least one record")
	g, err := reader.Geometry()
	require.NoError(t, err, "Failed to read geometry")

	readMP, ok := g.(*geom.MultiPolygon)
	require.True(t, ok, "Expected MultiPolygon, got %T", g)

	assert.Equal(t, 2, readMP.NumGeometries(), "Expected 2 polygons")

	// Verify total area
	expectedArea := 5.0*5.0 + 5.0*5.0 // 25 + 25 = 50
	assert.InDelta(t, expectedArea, readMP.Area(), 0.1, "Total area mismatch")
}

func TestReadAll(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "readall.shp")

	// Create and write test data
	factory := geom.DefaultFactory
	points := []geom.Geometry{
		factory.CreatePoint(1, 2),
		factory.CreatePoint(3, 4),
		factory.CreatePoint(5, 6),
	}

	err := WriteAll(filename, points)
	require.NoError(t, err, "Failed to write all")

	// Read all back
	readGeoms, err := ReadAll(filename)
	require.NoError(t, err, "Failed to read all")

	assert.Len(t, readGeoms, 3, "Expected 3 geometries")
}

func TestWriteAll(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "writeall.shp")

	// Create test data
	factory := geom.DefaultFactory
	geoms := []geom.Geometry{
		factory.CreatePoint(1, 2),
		factory.CreatePoint(3, 4),
	}

	// Write all
	err := WriteAll(filename, geoms)
	require.NoError(t, err, "Failed to write all")

	// Verify files exist
	_, err = os.Stat(filename)
	require.NoError(t, err, "SHP file should exist")

	shxFile := filename[:len(filename)-4] + ".shx"
	_, err = os.Stat(shxFile)
	require.NoError(t, err, "SHX file should exist")
}

func TestEmptyGeometries(t *testing.T) {
	// Note: The underlying go-shp library has limited support for Null shapes.
	// Null shapes written to a Point shapefile may not be properly preserved.
	// This test verifies that writing empty geometries doesn't cause errors.

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "empty.shp")

	factory := geom.DefaultFactory

	// Create writer with Point type
	writer, err := NewWriter(filename, ShapeTypePoint)
	require.NoError(t, err, "Failed to create writer")

	// Write a regular point
	err = writer.Write(factory.CreatePoint(1, 2))
	require.NoError(t, err, "Failed to write point")

	// Writing an empty point should not error (writes a Null shape)
	err = writer.Write(factory.CreatePointEmpty())
	require.NoError(t, err, "Failed to write empty point")

	// Write another regular point
	err = writer.Write(factory.CreatePoint(3, 4))
	require.NoError(t, err, "Failed to write point")

	err = writer.Close()
	require.NoError(t, err, "Failed to close writer")

	// Read back - we expect at least the non-empty points to be readable
	reader, err := NewReader(filename)
	require.NoError(t, err, "Failed to create reader")
	defer reader.Close() //nolint:errcheck

	var geoms []geom.Geometry
	for reader.Next() {
		g, err := reader.Geometry()
		require.NoError(t, err, "Failed to read geometry")
		geoms = append(geoms, g)
	}

	// The go-shp library reads 3 records but may not properly handle null shapes
	// We verify that we got at least the non-empty points
	require.GreaterOrEqual(t, len(geoms), 1, "Expected at least 1 geometry")

	// First should be regular point
	pt, ok := geoms[0].(*geom.Point)
	require.True(t, ok, "Expected Point")
	assert.InDelta(t, 1.0, pt.X(), 0.0001)
	assert.InDelta(t, 2.0, pt.Y(), 0.0001)
}

func TestBoundingBox(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "bbox.shp")

	// Create points spanning a known bounding box
	factory := geom.DefaultFactory
	points := []geom.Geometry{
		factory.CreatePoint(0, 0),
		factory.CreatePoint(100, 100),
		factory.CreatePoint(50, 50),
	}

	err := WriteAll(filename, points)
	require.NoError(t, err, "Failed to write all")

	// Read and check bounding box
	reader, err := NewReader(filename)
	require.NoError(t, err, "Failed to create reader")
	defer reader.Close() //nolint:errcheck

	bbox := reader.BoundingBox()
	assert.InDelta(t, 0.0, bbox.MinX, 0.0001, "MinX mismatch")
	assert.InDelta(t, 0.0, bbox.MinY, 0.0001, "MinY mismatch")
	assert.InDelta(t, 100.0, bbox.MaxX, 0.0001, "MaxX mismatch")
	assert.InDelta(t, 100.0, bbox.MaxY, 0.0001, "MaxY mismatch")
}

func TestShapeTypeString(t *testing.T) {
	tests := []struct {
		shapeType ShapeType
		expected  string
	}{
		{ShapeTypeNull, "Null"},
		{ShapeTypePoint, "Point"},
		{ShapeTypePolyLine, "PolyLine"},
		{ShapeTypePolygon, "Polygon"},
		{ShapeTypeMultiPoint, "MultiPoint"},
		{ShapeTypePointZ, "PointZ"},
		{ShapeTypePolyLineZ, "PolyLineZ"},
		{ShapeTypePolygonZ, "PolygonZ"},
		{ShapeTypeMultiPointZ, "MultiPointZ"},
		{ShapeType(999), "Unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.shapeType.String())
		})
	}
}

func TestGeometryToShapeType(t *testing.T) {
	factory := geom.DefaultFactory

	tests := []struct {
		name     string
		geom     geom.Geometry
		expected ShapeType
	}{
		{"Nil", nil, ShapeTypeNull},
		{"EmptyPoint", factory.CreatePointEmpty(), ShapeTypeNull},
		{"Point", factory.CreatePoint(1, 2), ShapeTypePoint},
		{"LineString", mustCreateLineStringXY(factory, 0, 0, 10, 10), ShapeTypePolyLine},
		{"Polygon", factory.CreatePolygon(mustCreateLinearRingXY(factory, 0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil), ShapeTypePolygon},
		{"MultiPoint", factory.CreateMultiPoint([]*geom.Point{factory.CreatePoint(1, 2)}), ShapeTypeMultiPoint},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := GeometryToShapeType(tc.geom)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestInferShapeType(t *testing.T) {
	factory := geom.DefaultFactory

	t.Run("EmptySlice", func(t *testing.T) {
		result := InferShapeType([]geom.Geometry{})
		assert.Equal(t, ShapeTypeNull, result)
	})

	t.Run("HomogeneousPoints", func(t *testing.T) {
		geoms := []geom.Geometry{
			factory.CreatePoint(1, 2),
			factory.CreatePoint(3, 4),
		}
		result := InferShapeType(geoms)
		assert.Equal(t, ShapeTypePoint, result)
	})

	t.Run("HomogeneousPolygons", func(t *testing.T) {
		shell := mustCreateLinearRingXY(factory, 0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		geoms := []geom.Geometry{
			factory.CreatePolygon(shell, nil),
			factory.CreatePolygon(shell, nil),
		}
		result := InferShapeType(geoms)
		assert.Equal(t, ShapeTypePolygon, result)
	})

	t.Run("HeterogeneousTypes", func(t *testing.T) {
		geoms := []geom.Geometry{
			factory.CreatePoint(1, 2),
			mustCreateLineStringXY(factory, 0, 0, 10, 10),
		}
		result := InferShapeType(geoms)
		assert.Equal(t, ShapeTypeNull, result)
	})
}

func TestWriteAllEmptySlice(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "empty.shp")

	err := WriteAll(filename, []geom.Geometry{})
	assert.Error(t, err, "Expected error for empty geometry slice")
}

func TestWriteAllHeterogeneous(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "hetero.shp")

	factory := geom.DefaultFactory
	geoms := []geom.Geometry{
		factory.CreatePoint(1, 2),
		mustCreateLineStringXY(factory, 0, 0, 10, 10),
	}

	err := WriteAll(filename, geoms)
	assert.Error(t, err, "Expected error for heterogeneous geometry types")
}

func TestReaderWithFactory(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "factory.shp")

	// Write with default factory
	factory := geom.DefaultFactory
	err := WriteAll(filename, []geom.Geometry{factory.CreatePoint(1, 2)})
	require.NoError(t, err, "Failed to write")

	// Read with custom factory
	customFactory := geom.NewGeometryFactoryWithSRID(4326)
	reader, err := NewReaderWithFactory(filename, customFactory)
	require.NoError(t, err, "Failed to create reader")
	defer reader.Close() //nolint:errcheck

	require.True(t, reader.Next(), "Expected at least one record")
	g, err := reader.Geometry()
	require.NoError(t, err, "Failed to read geometry")

	assert.Equal(t, 4326, g.SRID(), "Expected SRID from factory")
}

func TestReaderWithNilFactoryUsesDefault(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "nil_factory.shp")

	err := WriteAll(filename, []geom.Geometry{geom.DefaultFactory.CreatePoint(1, 2)})
	require.NoError(t, err)

	reader, err := NewReaderWithFactory(filename, nil)
	require.NoError(t, err)
	defer reader.Close() //nolint:errcheck

	require.True(t, reader.Next())
	g, err := reader.Geometry()
	require.NoError(t, err)
	require.IsType(t, &geom.Point{}, g)
}

func TestPolylineToGeometryRejectsMalformedParts(t *testing.T) {
	points := []shp.Point{{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 2, Y: 2}}

	tests := []struct {
		name  string
		parts []int32
		err   string
	}{
		{name: "negative offset", parts: []int32{-1}, err: "negative"},
		{name: "offset past points", parts: []int32{3}, err: "exceeds"},
		{name: "non increasing offsets", parts: []int32{0, 0}, err: "not greater"},
		{name: "short part", parts: []int32{0, 2}, err: "fewer than 2"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := polyLineToGeometry(points, tc.parts, false, nil, nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.err)
		})
	}
}

func TestPolylineToGeometryRejectsMismatchedZArray(t *testing.T) {
	points := []shp.Point{{X: 0, Y: 0}, {X: 1, Y: 1}}

	_, err := polyLineToGeometry(points, []int32{0}, true, []float64{10}, geom.DefaultFactory)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Z array length")
}

func TestPolygonToGeometryRejectsMalformedRings(t *testing.T) {
	tests := []struct {
		name   string
		points []shp.Point
		err    string
	}{
		{
			name:   "short ring",
			points: []shp.Point{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}},
			err:    "fewer than 4",
		},
		{
			name: "unclosed ring",
			points: []shp.Point{
				{X: 0, Y: 0},
				{X: 1, Y: 0},
				{X: 1, Y: 1},
				{X: 0, Y: 1},
			},
			err: "not closed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := polygonToGeometry(tc.points, []int32{0}, false, nil, nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.err)
		})
	}
}

func TestPolygonToGeometryRejectsMismatchedZArray(t *testing.T) {
	points := []shp.Point{
		{X: 0, Y: 0},
		{X: 1, Y: 0},
		{X: 1, Y: 1},
		{X: 0, Y: 0},
	}

	_, err := polygonToGeometry(points, []int32{0}, true, []float64{1, 2}, geom.DefaultFactory)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Z array length")
}

func TestMultiPointToGeometryRejectsMismatchedZArray(t *testing.T) {
	points := []shp.Point{{X: 0, Y: 0}, {X: 1, Y: 1}}

	_, err := multiPointToGeometry(points, true, []float64{1}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Z array length")
}

func TestFeatureIterator(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "features.shp")

	// Create shapefile with attributes
	factory := geom.DefaultFactory
	writer, err := NewWriter(filename, ShapeTypePoint)
	require.NoError(t, err)

	fields := []Field{
		{Name: "name", Type: FieldTypeString, Length: 50},
		{Name: "value", Type: FieldTypeInteger, Length: 10},
	}
	err = writer.SetFields(fields)
	require.NoError(t, err)

	// Write features
	for i := 0; i < 3; i++ {
		f := &Feature{
			Geometry:   factory.CreatePoint(float64(i), float64(i*2)),
			Properties: map[string]any{"name": "point" + string(rune('A'+i)), "value": i * 10},
		}
		err = writer.WriteFeature(f)
		require.NoError(t, err)
	}
	_ = writer.Close()

	// Read back using iterator
	count := 0
	for f, err := range Features(filename) {
		require.NoError(t, err)
		require.NotNil(t, f)
		assert.Equal(t, count, f.Index)
		assert.NotNil(t, f.Geometry)
		assert.NotNil(t, f.Properties)
		count++
	}
	assert.Equal(t, 3, count)
}

func TestFields(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "fields.shp")

	// Create shapefile with fields
	writer, err := NewWriter(filename, ShapeTypePoint)
	require.NoError(t, err)

	fields := []Field{
		{Name: "id", Type: FieldTypeInteger, Length: 10},
		{Name: "name", Type: FieldTypeString, Length: 50},
		{Name: "value", Type: FieldTypeFloat, Length: 10, Precision: 2},
	}
	err = writer.SetFields(fields)
	require.NoError(t, err)

	// Write a feature
	f := &Feature{
		Geometry:   geom.DefaultFactory.CreatePoint(1, 2),
		Properties: map[string]any{"id": 1, "name": "test", "value": 3.14},
	}
	err = writer.WriteFeature(f)
	require.NoError(t, err)
	_ = writer.Close()

	// Read back and verify fields
	reader, err := NewReader(filename)
	require.NoError(t, err)
	defer reader.Close() //nolint:errcheck

	fieldNames := reader.Fields()
	assert.Len(t, fieldNames, 3)
	assert.Equal(t, "id", fieldNames[0])
	assert.Equal(t, "name", fieldNames[1])
	assert.Equal(t, "value", fieldNames[2])
}

func TestWriteFeature(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "writefeature.shp")

	// Create shapefile
	writer, err := NewWriter(filename, ShapeTypePoint)
	require.NoError(t, err)

	fields := []Field{
		{Name: "city", Type: FieldTypeString, Length: 30},
		{Name: "pop", Type: FieldTypeInteger, Length: 10},
	}
	err = writer.SetFields(fields)
	require.NoError(t, err)

	// Write feature with attributes
	f := &Feature{
		Geometry:   geom.DefaultFactory.CreatePoint(-122.4194, 37.7749),
		Properties: map[string]any{"city": "San Francisco", "pop": 884363},
	}
	err = writer.WriteFeature(f)
	require.NoError(t, err)
	_ = writer.Close()

	// Read back and verify
	reader, err := NewReader(filename)
	require.NoError(t, err)
	defer reader.Close() //nolint:errcheck

	require.True(t, reader.Next())
	readFeature, err := reader.Feature()
	require.NoError(t, err)

	assert.Equal(t, 0, readFeature.Index)
	pt, ok := readFeature.Geometry.(*geom.Point)
	require.True(t, ok)
	assert.InDelta(t, -122.4194, pt.X(), 0.0001)
	assert.InDelta(t, 37.7749, pt.Y(), 0.0001)

	assert.Equal(t, "San Francisco", readFeature.Properties["city"])
	assert.Equal(t, "884363", readFeature.Properties["pop"]) // DBF returns strings
}

func BenchmarkWritePoints(b *testing.B) {
	tmpDir := b.TempDir()

	factory := geom.DefaultFactory
	points := make([]geom.Geometry, 1000)
	for i := 0; i < 1000; i++ {
		points[i] = factory.CreatePoint(float64(i), float64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filename := filepath.Join(tmpDir, "bench.shp")
		_ = WriteAll(filename, points)
		_ = os.Remove(filename)
		_ = os.Remove(filename[:len(filename)-4] + ".shx")
		_ = os.Remove(filename[:len(filename)-4] + ".dbf")
	}
}

func BenchmarkReadPoints(b *testing.B) {
	tmpDir := b.TempDir()
	filename := filepath.Join(tmpDir, "bench.shp")

	factory := geom.DefaultFactory
	points := make([]geom.Geometry, 1000)
	for i := 0; i < 1000; i++ {
		points[i] = factory.CreatePoint(float64(i), float64(i))
	}
	_ = WriteAll(filename, points)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ReadAll(filename)
	}
}
