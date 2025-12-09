package transform

import (
	"math"
	"testing"

	"github.com/go-topology-suite/gts/geom"
)

func TestTransformPoint(t *testing.T) {
	translation := NewAffineTranslation(10, 20)

	point := geom.NewPoint(5, 15)
	point.SetSRID(4326)

	result, err := TransformGeometry(translation, point)
	if err != nil {
		t.Fatalf("TransformGeometry() error = %v", err)
	}

	resultPoint, ok := result.(*geom.Point)
	if !ok {
		t.Fatalf("Result is not a Point: %T", result)
	}

	if math.Abs(resultPoint.X()-15) > epsilon || math.Abs(resultPoint.Y()-35) > epsilon {
		t.Errorf("Transformed point = (%f, %f), want (15, 35)",
			resultPoint.X(), resultPoint.Y())
	}

	// Check SRID is preserved
	if resultPoint.SRID() != 4326 {
		t.Errorf("SRID = %d, want 4326", resultPoint.SRID())
	}
}

func TestTransformEmptyPoint(t *testing.T) {
	translation := NewAffineTranslation(10, 20)
	point := geom.NewPointEmpty()

	result, err := TransformGeometry(translation, point)
	if err != nil {
		t.Fatalf("TransformGeometry() error = %v", err)
	}

	if !result.IsEmpty() {
		t.Error("Transformed empty point should remain empty")
	}
}

func TestTransformLineString(t *testing.T) {
	scale := NewAffineScale(2, 3)

	ls := geom.NewLineStringXY(0, 0, 10, 10, 20, 20)
	ls.SetSRID(4326)

	result, err := TransformGeometry(scale, ls)
	if err != nil {
		t.Fatalf("TransformGeometry() error = %v", err)
	}

	resultLS, ok := result.(*geom.LineString)
	if !ok {
		t.Fatalf("Result is not a LineString: %T", result)
	}

	expected := []struct{ x, y float64 }{
		{0, 0},
		{20, 30},
		{40, 60},
	}

	coords := resultLS.Coordinates()
	if len(coords) != len(expected) {
		t.Fatalf("Result has %d coordinates, want %d", len(coords), len(expected))
	}

	for i, exp := range expected {
		if math.Abs(coords[i].X-exp.x) > epsilon || math.Abs(coords[i].Y-exp.y) > epsilon {
			t.Errorf("Coordinate %d = (%f, %f), want (%f, %f)",
				i, coords[i].X, coords[i].Y, exp.x, exp.y)
		}
	}

	// Check SRID is preserved
	if resultLS.SRID() != 4326 {
		t.Errorf("SRID = %d, want 4326", resultLS.SRID())
	}
}

func TestTransformPolygon(t *testing.T) {
	rotation := NewAffineRotation(math.Pi / 2) // 90 degrees counter-clockwise

	// Create a square polygon
	shell := geom.NewLinearRingXY(0, 0, 4, 0, 4, 4, 0, 4, 0, 0)
	polygon := geom.NewPolygon(shell, []*geom.LinearRing{})
	polygon.SetSRID(4326)

	result, err := TransformGeometry(rotation, polygon)
	if err != nil {
		t.Fatalf("TransformGeometry() error = %v", err)
	}

	resultPoly, ok := result.(*geom.Polygon)
	if !ok {
		t.Fatalf("Result is not a Polygon: %T", result)
	}

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

	if !found {
		t.Error("Expected coordinate (0, 4) not found in rotated polygon")
	}

	// Check SRID is preserved
	if resultPoly.SRID() != 4326 {
		t.Errorf("SRID = %d, want 4326", resultPoly.SRID())
	}
}

func TestTransformPolygonWithHoles(t *testing.T) {
	translation := NewAffineTranslation(10, 20)

	// Create a polygon with a hole
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	hole := geom.NewLinearRingXY(2, 2, 8, 2, 8, 8, 2, 8, 2, 2)
	polygon := geom.NewPolygon(shell, []*geom.LinearRing{hole})

	result, err := TransformGeometry(translation, polygon)
	if err != nil {
		t.Fatalf("TransformGeometry() error = %v", err)
	}

	resultPoly, ok := result.(*geom.Polygon)
	if !ok {
		t.Fatalf("Result is not a Polygon: %T", result)
	}

	if resultPoly.NumInteriorRings() != 1 {
		t.Errorf("Result has %d holes, want 1", resultPoly.NumInteriorRings())
	}

	// Check that shell was transformed
	shellCoords := resultPoly.ExteriorRing().Coordinates()
	if math.Abs(shellCoords[0].X-10) > epsilon || math.Abs(shellCoords[0].Y-20) > epsilon {
		t.Errorf("Shell first coordinate = (%f, %f), want (10, 20)",
			shellCoords[0].X, shellCoords[0].Y)
	}

	// Check that hole was transformed
	holeCoords := resultPoly.InteriorRingN(0).Coordinates()
	if math.Abs(holeCoords[0].X-12) > epsilon || math.Abs(holeCoords[0].Y-22) > epsilon {
		t.Errorf("Hole first coordinate = (%f, %f), want (12, 22)",
			holeCoords[0].X, holeCoords[0].Y)
	}
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
	if err != nil {
		t.Fatalf("TransformGeometry() error = %v", err)
	}

	resultMP, ok := result.(*geom.MultiPoint)
	if !ok {
		t.Fatalf("Result is not a MultiPoint: %T", result)
	}

	expected := []struct{ x, y float64 }{
		{2, 4},
		{6, 8},
		{10, 12},
	}

	if resultMP.NumGeometries() != len(expected) {
		t.Fatalf("Result has %d points, want %d", resultMP.NumGeometries(), len(expected))
	}

	for i, exp := range expected {
		point := resultMP.GeometryN(i).(*geom.Point)
		if math.Abs(point.X()-exp.x) > epsilon || math.Abs(point.Y()-exp.y) > epsilon {
			t.Errorf("Point %d = (%f, %f), want (%f, %f)",
				i, point.X(), point.Y(), exp.x, exp.y)
		}
	}

	// Check SRID is preserved
	if resultMP.SRID() != 4326 {
		t.Errorf("SRID = %d, want 4326", resultMP.SRID())
	}
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
	if err != nil {
		t.Fatalf("TransformGeometry() error = %v", err)
	}

	resultMLS, ok := result.(*geom.MultiLineString)
	if !ok {
		t.Fatalf("Result is not a MultiLineString: %T", result)
	}

	if resultMLS.NumGeometries() != 2 {
		t.Fatalf("Result has %d linestrings, want 2", resultMLS.NumGeometries())
	}

	// Check first linestring
	ls1 := resultMLS.GeometryN(0).(*geom.LineString)
	coords1 := ls1.Coordinates()
	if math.Abs(coords1[0].X-5) > epsilon || math.Abs(coords1[0].Y-10) > epsilon {
		t.Errorf("First linestring start = (%f, %f), want (5, 10)",
			coords1[0].X, coords1[0].Y)
	}

	// Check SRID is preserved
	if resultMLS.SRID() != 4326 {
		t.Errorf("SRID = %d, want 4326", resultMLS.SRID())
	}
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
	if err != nil {
		t.Fatalf("TransformGeometry() error = %v", err)
	}

	resultMPoly, ok := result.(*geom.MultiPolygon)
	if !ok {
		t.Fatalf("Result is not a MultiPolygon: %T", result)
	}

	if resultMPoly.NumGeometries() != 2 {
		t.Fatalf("Result has %d polygons, want 2", resultMPoly.NumGeometries())
	}

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
	if !found {
		t.Error("Expected coordinate (10, 10) not found in scaled polygon")
	}

	// Check SRID is preserved
	if resultMPoly.SRID() != 4326 {
		t.Errorf("SRID = %d, want 4326", resultMPoly.SRID())
	}
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
	if err != nil {
		t.Fatalf("TransformGeometry() error = %v", err)
	}

	resultGC, ok := result.(*geom.GeometryCollection)
	if !ok {
		t.Fatalf("Result is not a GeometryCollection: %T", result)
	}

	if resultGC.NumGeometries() != 3 {
		t.Fatalf("Result has %d geometries, want 3", resultGC.NumGeometries())
	}

	// Check point
	point := resultGC.GeometryN(0).(*geom.Point)
	if math.Abs(point.X()-10) > epsilon || math.Abs(point.Y()-20) > epsilon {
		t.Errorf("Point = (%f, %f), want (10, 20)", point.X(), point.Y())
	}

	// Check linestring
	ls := resultGC.GeometryN(1).(*geom.LineString)
	coords := ls.Coordinates()
	if math.Abs(coords[0].X-10) > epsilon || math.Abs(coords[0].Y-20) > epsilon {
		t.Errorf("LineString start = (%f, %f), want (10, 20)",
			coords[0].X, coords[0].Y)
	}

	// Check polygon
	poly := resultGC.GeometryN(2).(*geom.Polygon)
	polyCoords := poly.ExteriorRing().Coordinates()
	if math.Abs(polyCoords[0].X-10) > epsilon || math.Abs(polyCoords[0].Y-20) > epsilon {
		t.Errorf("Polygon start = (%f, %f), want (10, 20)",
			polyCoords[0].X, polyCoords[0].Y)
	}

	// Check SRID is preserved
	if resultGC.SRID() != 4326 {
		t.Errorf("SRID = %d, want 4326", resultGC.SRID())
	}
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
	if err != nil {
		t.Fatalf("TransformGeometry() error = %v", err)
	}

	resultGC, ok := result.(*geom.GeometryCollection)
	if !ok {
		t.Fatalf("Result is not a GeometryCollection: %T", result)
	}

	// Check outer point
	point := resultGC.GeometryN(0).(*geom.Point)
	if math.Abs(point.X()-10) > epsilon || math.Abs(point.Y()-12) > epsilon {
		t.Errorf("Outer point = (%f, %f), want (10, 12)", point.X(), point.Y())
	}

	// Check inner collection
	innerResult := resultGC.GeometryN(1).(*geom.GeometryCollection)
	innerPoint := innerResult.GeometryN(0).(*geom.Point)
	if math.Abs(innerPoint.X()-2) > epsilon || math.Abs(innerPoint.Y()-4) > epsilon {
		t.Errorf("Inner point = (%f, %f), want (2, 4)", innerPoint.X(), innerPoint.Y())
	}
}

func TestTransformNilGeometry(t *testing.T) {
	translation := NewAffineTranslation(10, 20)

	result, err := TransformGeometry(translation, nil)
	if err != nil {
		t.Fatalf("TransformGeometry() error = %v", err)
	}

	if result != nil {
		t.Error("Transforming nil geometry should return nil")
	}
}

func BenchmarkTransformPoint(b *testing.B) {
	transform := NewAffineRotation(math.Pi / 4)
	point := geom.NewPoint(123.456, 789.012)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TransformGeometry(transform, point)
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
		TransformGeometry(transform, ls)
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
		TransformGeometry(transform, poly)
	}
}
