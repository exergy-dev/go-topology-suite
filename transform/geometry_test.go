package transform

import (
	"math"
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransformPoint(t *testing.T) {
	translation := NewAffineTranslation(10, 20)

	point := geom.NewPoint(5, 15)
	point.SetSRID(4326)

	result, err := TransformGeometry(translation, point)
	require.NoError(t, err, "TransformGeometry() error")

	resultPoint, ok := result.(*geom.Point)
	require.True(t, ok, "Result is not a Point: %T", result)

	assert.InDelta(t, 15.0, resultPoint.X(), epsilon)
	assert.InDelta(t, 35.0, resultPoint.Y(), epsilon)

	// Check SRID is preserved
	assert.Equal(t, 4326, resultPoint.SRID(), "SRID should be preserved")
}

func TestTransformEmptyPoint(t *testing.T) {
	translation := NewAffineTranslation(10, 20)
	point := geom.NewPointEmpty()

	result, err := TransformGeometry(translation, point)
	require.NoError(t, err, "TransformGeometry() error")

	assert.True(t, result.IsEmpty(), "Transformed empty point should remain empty")
}

func TestTransformLineString(t *testing.T) {
	scale := NewAffineScale(2, 3)

	ls := geom.NewLineStringXY(0, 0, 10, 10, 20, 20)
	ls.SetSRID(4326)

	result, err := TransformGeometry(scale, ls)
	require.NoError(t, err, "TransformGeometry() error")

	resultLS, ok := result.(*geom.LineString)
	require.True(t, ok, "Result is not a LineString: %T", result)

	expected := []struct{ x, y float64 }{
		{0, 0},
		{20, 30},
		{40, 60},
	}

	coords := resultLS.Coordinates()
	require.Len(t, coords, len(expected), "Coordinate count mismatch")

	for i, exp := range expected {
		assert.InDelta(t, exp.x, coords[i].X, epsilon, "Coordinate %d X", i)
		assert.InDelta(t, exp.y, coords[i].Y, epsilon, "Coordinate %d Y", i)
	}

	// Check SRID is preserved
	assert.Equal(t, 4326, resultLS.SRID(), "SRID should be preserved")
}

func TestTransformPolygon(t *testing.T) {
	rotation := NewAffineRotation(math.Pi / 2) // 90 degrees counter-clockwise

	// Create a square polygon
	shell := geom.NewLinearRingXY(0, 0, 4, 0, 4, 4, 0, 4, 0, 0)
	polygon := geom.NewPolygon(shell, []*geom.LinearRing{})
	polygon.SetSRID(4326)

	result, err := TransformGeometry(rotation, polygon)
	require.NoError(t, err, "TransformGeometry() error")

	resultPoly, ok := result.(*geom.Polygon)
	require.True(t, ok, "Result is not a Polygon: %T", result)

	// After 90-degree rotation, (4, 0) should become approximately (0, 4)
	coords := resultPoly.ExteriorRing().Coordinates()

	// Find the coordinate that should be (0, 4) - it was (4, 0) before rotation
	found := false
	for _, coord := range coords {
		if math.Abs(coord.X-0) < 0.01 && math.Abs(coord.Y-4) < 0.01 {
			found = true
			break
		}
	}

	assert.True(t, found, "Expected coordinate (0, 4) not found in rotated polygon")

	// Check SRID is preserved
	assert.Equal(t, 4326, resultPoly.SRID(), "SRID should be preserved")
}

func TestTransformPolygonWithHoles(t *testing.T) {
	translation := NewAffineTranslation(10, 20)

	// Create a polygon with a hole
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	hole := geom.NewLinearRingXY(2, 2, 8, 2, 8, 8, 2, 8, 2, 2)
	polygon := geom.NewPolygon(shell, []*geom.LinearRing{hole})

	result, err := TransformGeometry(translation, polygon)
	require.NoError(t, err, "TransformGeometry() error")

	resultPoly, ok := result.(*geom.Polygon)
	require.True(t, ok, "Result is not a Polygon: %T", result)

	assert.Equal(t, 1, resultPoly.NumInteriorRings(), "Should have 1 hole")

	// Check that shell was transformed
	shellCoords := resultPoly.ExteriorRing().Coordinates()
	assert.InDelta(t, 10.0, shellCoords[0].X, epsilon, "Shell first X")
	assert.InDelta(t, 20.0, shellCoords[0].Y, epsilon, "Shell first Y")

	// Check that hole was transformed
	holeCoords := resultPoly.InteriorRingN(0).Coordinates()
	assert.InDelta(t, 12.0, holeCoords[0].X, epsilon, "Hole first X")
	assert.InDelta(t, 22.0, holeCoords[0].Y, epsilon, "Hole first Y")
}

func TestTransformMultiPoint(t *testing.T) {
	scale := NewAffineScale(2, 2)

	points := []*geom.Point{
		geom.NewPoint(1, 2),
		geom.NewPoint(3, 4),
		geom.NewPoint(5, 6),
	}
	mp := geom.NewMultiPoint(points)
	mp.SetSRID(4326)

	result, err := TransformGeometry(scale, mp)
	require.NoError(t, err, "TransformGeometry() error")

	resultMP, ok := result.(*geom.MultiPoint)
	require.True(t, ok, "Result is not a MultiPoint: %T", result)

	expected := []struct{ x, y float64 }{
		{2, 4},
		{6, 8},
		{10, 12},
	}

	require.Equal(t, len(expected), resultMP.NumGeometries(), "Result point count")

	for i, exp := range expected {
		point := resultMP.GeometryN(i).(*geom.Point)
		assert.InDelta(t, exp.x, point.X(), epsilon, "Point %d X", i)
		assert.InDelta(t, exp.y, point.Y(), epsilon, "Point %d Y", i)
	}

	// Check SRID is preserved
	assert.Equal(t, 4326, resultMP.SRID(), "SRID should be preserved")
}

func TestTransformMultiLineString(t *testing.T) {
	translation := NewAffineTranslation(5, 10)

	linestrings := []*geom.LineString{
		geom.NewLineStringXY(0, 0, 10, 0),
		geom.NewLineStringXY(0, 10, 10, 10),
	}
	mls := geom.NewMultiLineString(linestrings)
	mls.SetSRID(4326)

	result, err := TransformGeometry(translation, mls)
	require.NoError(t, err, "TransformGeometry() error")

	resultMLS, ok := result.(*geom.MultiLineString)
	require.True(t, ok, "Result is not a MultiLineString: %T", result)

	require.Equal(t, 2, resultMLS.NumGeometries(), "Result linestring count")

	// Check first linestring
	ls1 := resultMLS.GeometryN(0).(*geom.LineString)
	coords1 := ls1.Coordinates()
	assert.InDelta(t, 5.0, coords1[0].X, epsilon, "First linestring start X")
	assert.InDelta(t, 10.0, coords1[0].Y, epsilon, "First linestring start Y")

	// Check SRID is preserved
	assert.Equal(t, 4326, resultMLS.SRID(), "SRID should be preserved")
}

func TestTransformMultiPolygon(t *testing.T) {
	scale := NewAffineScale(2, 2)

	polygons := []*geom.Polygon{
		geom.NewPolygon(geom.NewLinearRingXY(0, 0, 5, 0, 5, 5, 0, 5, 0, 0), []*geom.LinearRing{}),
		geom.NewPolygon(geom.NewLinearRingXY(10, 10, 15, 10, 15, 15, 10, 15, 10, 10), []*geom.LinearRing{}),
	}
	mpoly := geom.NewMultiPolygon(polygons)
	mpoly.SetSRID(4326)

	result, err := TransformGeometry(scale, mpoly)
	require.NoError(t, err, "TransformGeometry() error")

	resultMPoly, ok := result.(*geom.MultiPolygon)
	require.True(t, ok, "Result is not a MultiPolygon: %T", result)

	require.Equal(t, 2, resultMPoly.NumGeometries(), "Result polygon count")

	// Check first polygon
	poly1 := resultMPoly.GeometryN(0).(*geom.Polygon)
	coords1 := poly1.ExteriorRing().Coordinates()
	// (5, 5) should become (10, 10)
	found := false
	for _, coord := range coords1 {
		if math.Abs(coord.X-10) < epsilon && math.Abs(coord.Y-10) < epsilon {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected coordinate (10, 10) not found in scaled polygon")

	// Check SRID is preserved
	assert.Equal(t, 4326, resultMPoly.SRID(), "SRID should be preserved")
}

func TestTransformGeometryCollection(t *testing.T) {
	translation := NewAffineTranslation(10, 20)

	geometries := []geom.Geometry{
		geom.NewPoint(0, 0),
		geom.NewLineStringXY(0, 0, 10, 10),
		geom.NewPolygon(geom.NewLinearRingXY(0, 0, 5, 0, 5, 5, 0, 5, 0, 0), []*geom.LinearRing{}),
	}
	gc := geom.NewGeometryCollection(geometries)
	gc.SetSRID(4326)

	result, err := TransformGeometry(translation, gc)
	require.NoError(t, err, "TransformGeometry() error")

	resultGC, ok := result.(*geom.GeometryCollection)
	require.True(t, ok, "Result is not a GeometryCollection: %T", result)

	require.Equal(t, 3, resultGC.NumGeometries(), "Result geometry count")

	// Check point
	point := resultGC.GeometryN(0).(*geom.Point)
	assert.InDelta(t, 10.0, point.X(), epsilon, "Point X")
	assert.InDelta(t, 20.0, point.Y(), epsilon, "Point Y")

	// Check linestring
	ls := resultGC.GeometryN(1).(*geom.LineString)
	coords := ls.Coordinates()
	assert.InDelta(t, 10.0, coords[0].X, epsilon, "LineString start X")
	assert.InDelta(t, 20.0, coords[0].Y, epsilon, "LineString start Y")

	// Check polygon
	poly := resultGC.GeometryN(2).(*geom.Polygon)
	polyCoords := poly.ExteriorRing().Coordinates()
	assert.InDelta(t, 10.0, polyCoords[0].X, epsilon, "Polygon start X")
	assert.InDelta(t, 20.0, polyCoords[0].Y, epsilon, "Polygon start Y")

	// Check SRID is preserved
	assert.Equal(t, 4326, resultGC.SRID(), "SRID should be preserved")
}

func TestTransformNestedGeometryCollection(t *testing.T) {
	scale := NewAffineScale(2, 2)

	// Create a nested collection
	innerCollection := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(1, 2),
		geom.NewPoint(3, 4),
	})

	outerCollection := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(5, 6),
		innerCollection,
	})

	result, err := TransformGeometry(scale, outerCollection)
	require.NoError(t, err, "TransformGeometry() error")

	resultGC, ok := result.(*geom.GeometryCollection)
	require.True(t, ok, "Result is not a GeometryCollection: %T", result)

	// Check outer point
	point := resultGC.GeometryN(0).(*geom.Point)
	assert.InDelta(t, 10.0, point.X(), epsilon, "Outer point X")
	assert.InDelta(t, 12.0, point.Y(), epsilon, "Outer point Y")

	// Check inner collection
	innerResult := resultGC.GeometryN(1).(*geom.GeometryCollection)
	innerPoint := innerResult.GeometryN(0).(*geom.Point)
	assert.InDelta(t, 2.0, innerPoint.X(), epsilon, "Inner point X")
	assert.InDelta(t, 4.0, innerPoint.Y(), epsilon, "Inner point Y")
}

func TestTransformNilGeometry(t *testing.T) {
	translation := NewAffineTranslation(10, 20)

	result, err := TransformGeometry(translation, nil)
	require.NoError(t, err, "TransformGeometry() error")

	assert.Nil(t, result, "Transforming nil geometry should return nil")
}

func BenchmarkTransformPoint(b *testing.B) {
	transform := NewAffineRotation(math.Pi / 4)
	point := geom.NewPoint(123.456, 789.012)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = TransformGeometry(transform, point)
	}
}

func BenchmarkTransformLineString(b *testing.B) {
	transform := NewAffineRotation(math.Pi / 4)

	// Create a linestring with 100 points
	coords := make(geom.CoordinateSequence, 100)
	for i := 0; i < 100; i++ {
		coords[i] = geom.NewCoordinate(float64(i), float64(i)*2)
	}
	ls := geom.NewLineString(coords)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = TransformGeometry(transform, ls)
	}
}

func BenchmarkTransformPolygon(b *testing.B) {
	transform := NewAffineRotation(math.Pi / 4)

	// Create a polygon with 100 points
	coords := make(geom.CoordinateSequence, 101)
	for i := 0; i < 100; i++ {
		angle := float64(i) * 2 * math.Pi / 100
		coords[i] = geom.NewCoordinate(100*math.Cos(angle), 100*math.Sin(angle))
	}
	coords[100] = coords[0] // Close the ring

	shell := geom.NewLinearRing(coords)
	poly := geom.NewPolygon(shell, []*geom.LinearRing{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = TransformGeometry(transform, poly)
	}
}
