package topology

import (
	"math"
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/io/wkb"
	"github.com/robert-malhotra/go-topology-suite/operation/buffer"
)

func TestNewPolygonRejectsInvalidByDefault(t *testing.T) {
	_, err := NewPolygon(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(0, 0),
	}, nil)
	if err == nil {
		t.Fatal("expected invalid bow-tie polygon to be rejected")
	}
}

func TestNewPointRejectsInvalidByDefault(t *testing.T) {
	if _, err := NewPoint(math.NaN(), 0); err == nil {
		t.Fatal("expected invalid point to be rejected")
	}

	point, err := NewPoint(math.NaN(), 0, ConstructorOptions{AllowInvalid: true})
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}
	if point.IsValid() {
		t.Fatal("expected constructed point to remain invalid")
	}
}

func TestNewMultiPointConstructors(t *testing.T) {
	factory := NewGeometryFactory(4326)
	point, err := NewPoint(1, 2)
	if err != nil {
		t.Fatalf("unexpected point constructor error: %v", err)
	}

	multiPoint, err := NewMultiPoint([]*geom.Point{point}, ConstructorOptions{Factory: factory})
	if err != nil {
		t.Fatalf("unexpected multipoint constructor error: %v", err)
	}
	if multiPoint.NumGeometries() != 1 {
		t.Fatalf("expected one point, got %d", multiPoint.NumGeometries())
	}
	if multiPoint.SRID() != 4326 {
		t.Fatalf("expected SRID 4326, got %d", multiPoint.SRID())
	}

	fromCoords, err := NewMultiPointFromCoords(geom.CoordinateSequence{
		geom.NewCoordinate(3, 4),
	})
	if err != nil {
		t.Fatalf("unexpected multipoint from coords constructor error: %v", err)
	}
	if fromCoords.NumGeometries() != 1 {
		t.Fatalf("expected one point, got %d", fromCoords.NumGeometries())
	}

	empty, err := NewMultiPointEmpty()
	if err != nil {
		t.Fatalf("unexpected empty multipoint constructor error: %v", err)
	}
	if !empty.IsEmpty() {
		t.Fatal("expected empty multipoint")
	}
}

func TestNewMultiPointRejectsInvalidAndNilInputs(t *testing.T) {
	invalidPoint := geom.NewPoint(math.NaN(), 0)
	if _, err := NewMultiPoint([]*geom.Point{invalidPoint}); err == nil {
		t.Fatal("expected invalid point component to be rejected")
	}
	if _, err := NewMultiPoint([]*geom.Point{invalidPoint}, ConstructorOptions{AllowInvalid: true}); err != nil {
		t.Fatalf("unexpected constructor error with AllowInvalid: %v", err)
	}
	if _, err := NewMultiPoint(nil); err == nil {
		t.Fatal("expected nil points slice to be rejected")
	}
	if _, err := NewMultiPoint([]*geom.Point{nil}); err == nil {
		t.Fatal("expected nil point component to be rejected")
	}
	if _, err := NewMultiPointFromCoords(nil); err == nil {
		t.Fatal("expected nil coordinates to be rejected")
	}
	if _, err := NewMultiPointFromCoords(geom.CoordinateSequence{
		geom.NewCoordinate(math.NaN(), 0),
	}); err == nil {
		t.Fatal("expected invalid coordinate to be rejected")
	}
	if _, err := NewMultiPointFromCoords(geom.CoordinateSequence{
		geom.NewCoordinate(math.NaN(), 0),
	}, ConstructorOptions{AllowInvalid: true}); err != nil {
		t.Fatalf("unexpected constructor error with AllowInvalid: %v", err)
	}
}

func TestNewMultiLineStringConstructors(t *testing.T) {
	line, err := NewLineString(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(1, 1),
	})
	if err != nil {
		t.Fatalf("unexpected line constructor error: %v", err)
	}

	multiLine, err := NewMultiLineString([]*geom.LineString{line}, ConstructorOptions{Factory: NewGeometryFactory(3857)})
	if err != nil {
		t.Fatalf("unexpected multilinestring constructor error: %v", err)
	}
	if multiLine.NumGeometries() != 1 {
		t.Fatalf("expected one line, got %d", multiLine.NumGeometries())
	}
	if multiLine.SRID() != 3857 {
		t.Fatalf("expected SRID 3857, got %d", multiLine.SRID())
	}

	empty, err := NewMultiLineStringEmpty()
	if err != nil {
		t.Fatalf("unexpected empty multilinestring constructor error: %v", err)
	}
	if !empty.IsEmpty() {
		t.Fatal("expected empty multilinestring")
	}
}

func TestNewMultiLineStringRejectsInvalidAndNilInputs(t *testing.T) {
	invalidLine := geom.NewLineString(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
	})
	if _, err := NewMultiLineString([]*geom.LineString{invalidLine}); err == nil {
		t.Fatal("expected invalid line component to be rejected")
	}
	if _, err := NewMultiLineString([]*geom.LineString{invalidLine}, ConstructorOptions{AllowInvalid: true}); err != nil {
		t.Fatalf("unexpected constructor error with AllowInvalid: %v", err)
	}
	if _, err := NewMultiLineString(nil); err == nil {
		t.Fatal("expected nil lines slice to be rejected")
	}
	if _, err := NewMultiLineString([]*geom.LineString{nil}); err == nil {
		t.Fatal("expected nil line component to be rejected")
	}
}

func TestNewMultiPolygonConstructors(t *testing.T) {
	polygon, err := NewPolygon(square(0, 0, 1), nil)
	if err != nil {
		t.Fatalf("unexpected polygon constructor error: %v", err)
	}

	multiPolygon, err := NewMultiPolygon([]*geom.Polygon{polygon}, ConstructorOptions{Factory: NewGeometryFactory(32633)})
	if err != nil {
		t.Fatalf("unexpected multipolygon constructor error: %v", err)
	}
	if multiPolygon.NumGeometries() != 1 {
		t.Fatalf("expected one polygon, got %d", multiPolygon.NumGeometries())
	}
	if multiPolygon.SRID() != 32633 {
		t.Fatalf("expected SRID 32633, got %d", multiPolygon.SRID())
	}

	empty, err := NewMultiPolygonEmpty()
	if err != nil {
		t.Fatalf("unexpected empty multipolygon constructor error: %v", err)
	}
	if !empty.IsEmpty() {
		t.Fatal("expected empty multipolygon")
	}
}

func TestNewMultiPolygonRejectsInvalidAndNilInputs(t *testing.T) {
	invalidPolygon := geom.NewPolygon(
		geom.NewLinearRing(geom.CoordinateSequence{
			geom.NewCoordinate(0, 0),
			geom.NewCoordinate(10, 10),
			geom.NewCoordinate(0, 10),
			geom.NewCoordinate(10, 0),
			geom.NewCoordinate(0, 0),
		}),
		nil,
	)
	if _, err := NewMultiPolygon([]*geom.Polygon{invalidPolygon}); err == nil {
		t.Fatal("expected invalid polygon component to be rejected")
	}
	if _, err := NewMultiPolygon([]*geom.Polygon{invalidPolygon}, ConstructorOptions{AllowInvalid: true}); err != nil {
		t.Fatalf("unexpected constructor error with AllowInvalid: %v", err)
	}
	if _, err := NewMultiPolygon(nil); err == nil {
		t.Fatal("expected nil polygons slice to be rejected")
	}
	if _, err := NewMultiPolygon([]*geom.Polygon{nil}); err == nil {
		t.Fatal("expected nil polygon component to be rejected")
	}
}

func TestNewGeometryCollectionConstructors(t *testing.T) {
	point, err := NewPoint(1, 2)
	if err != nil {
		t.Fatalf("unexpected point constructor error: %v", err)
	}

	collection, err := NewGeometryCollection([]geom.Geometry{point}, ConstructorOptions{Factory: NewGeometryFactory(4269)})
	if err != nil {
		t.Fatalf("unexpected geometrycollection constructor error: %v", err)
	}
	if collection.NumGeometries() != 1 {
		t.Fatalf("expected one geometry, got %d", collection.NumGeometries())
	}
	if collection.SRID() != 4269 {
		t.Fatalf("expected SRID 4269, got %d", collection.SRID())
	}

	empty, err := NewGeometryCollectionEmpty()
	if err != nil {
		t.Fatalf("unexpected empty geometrycollection constructor error: %v", err)
	}
	if !empty.IsEmpty() {
		t.Fatal("expected empty geometrycollection")
	}
}

func TestNewGeometryCollectionRejectsInvalidAndNilInputs(t *testing.T) {
	invalidLine := geom.NewLineString(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
	})
	if _, err := NewGeometryCollection([]geom.Geometry{invalidLine}); err == nil {
		t.Fatal("expected invalid geometry component to be rejected")
	}
	if _, err := NewGeometryCollection([]geom.Geometry{invalidLine}, ConstructorOptions{AllowInvalid: true}); err != nil {
		t.Fatalf("unexpected constructor error with AllowInvalid: %v", err)
	}
	if _, err := NewGeometryCollection(nil); err == nil {
		t.Fatal("expected nil geometries slice to be rejected")
	}
	if _, err := NewGeometryCollection([]geom.Geometry{nil}); err == nil {
		t.Fatal("expected nil geometry component to be rejected")
	}
}

func TestOverlayReturnsErrorsForNilAndInvalidInputs(t *testing.T) {
	point := geom.NewPoint(0, 0)
	if _, err := Intersection(nil, point); err == nil {
		t.Fatal("expected nil input error")
	}

	invalid, err := NewPolygon(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(0, 0),
	}, nil, ConstructorOptions{AllowInvalid: true})
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}
	if _, err := Intersection(invalid, point); err == nil {
		t.Fatal("expected invalid input error")
	}
}

func TestOverlayPrecisionModelSnapsInputs(t *testing.T) {
	left := geom.NewPoint(1.2, 2.2)
	right := geom.NewPoint(1.4, 2.4)

	result, err := Intersection(left, right, OverlayOptions{
		PrecisionModel: geom.NewFixedPrecision(1),
	})
	if err != nil {
		t.Fatalf("unexpected overlay error: %v", err)
	}
	point, ok := result.(*geom.Point)
	if !ok || point.IsEmpty() {
		t.Fatalf("expected snapped point intersection, got %T", result)
	}
	if point.X() != 1 || point.Y() != 2 {
		t.Fatalf("expected snapped point (1,2), got (%v,%v)", point.X(), point.Y())
	}
	if left.X() != 1.2 || right.X() != 1.4 {
		t.Fatal("precision option should not mutate input geometries")
	}
}

func TestOverlayPrecisionModelSnapsPolygonInputsWithoutMutation(t *testing.T) {
	left, err := NewPolygon(square(0.2, 0.2, 4), nil)
	if err != nil {
		t.Fatalf("unexpected left polygon constructor error: %v", err)
	}
	right, err := NewPolygon(square(2.6, 2.6, 4), nil)
	if err != nil {
		t.Fatalf("unexpected right polygon constructor error: %v", err)
	}

	opts := OverlayOptions{PrecisionModel: geom.NewFixedPrecision(1)}
	intersection, err := Intersection(left, right, opts)
	if err != nil {
		t.Fatalf("unexpected intersection error: %v", err)
	}
	union, err := Union(left, right, opts)
	if err != nil {
		t.Fatalf("unexpected union error: %v", err)
	}

	if intersection == nil || intersection.IsEmpty() {
		t.Fatal("expected non-empty snapped polygon intersection")
	}
	if union == nil || union.IsEmpty() {
		t.Fatal("expected non-empty snapped polygon union")
	}
	if area := areaOf(t, intersection); area != 1 {
		t.Fatalf("expected snapped intersection area 1, got %v", area)
	}
	if area := areaOf(t, union); area != 31 {
		t.Fatalf("expected snapped union area 31, got %v", area)
	}

	leftStart := left.ExteriorRing().Coordinates()[0]
	rightStart := right.ExteriorRing().Coordinates()[0]
	if leftStart.X != 0.2 || leftStart.Y != 0.2 || rightStart.X != 2.6 || rightStart.Y != 2.6 {
		t.Fatalf("precision option should not mutate input polygons, got left start (%v,%v), right start (%v,%v)",
			leftStart.X, leftStart.Y, rightStart.X, rightStart.Y)
	}
}

func TestBufferPrecisionModelSnapsInput(t *testing.T) {
	point := geom.NewPoint(1.2, 2.6)

	result, err := Buffer(point, 0, BufferOptions{
		PrecisionModel: geom.NewFixedPrecision(1),
	})
	if err != nil {
		t.Fatalf("unexpected buffer error: %v", err)
	}
	bufferedPoint, ok := result.(*geom.Point)
	if !ok {
		t.Fatalf("expected zero-distance point buffer, got %T", result)
	}
	if bufferedPoint.X() != 1 || bufferedPoint.Y() != 3 {
		t.Fatalf("expected snapped point (1,3), got (%v,%v)", bufferedPoint.X(), bufferedPoint.Y())
	}
	if point.X() != 1.2 || point.Y() != 2.6 {
		t.Fatal("precision option should not mutate input geometry")
	}
}

func TestBufferRejectsInvalidParams(t *testing.T) {
	point := geom.NewPoint(0, 0)

	tests := []struct {
		name   string
		params buffer.Params
	}{
		{
			name: "quadrant segments",
			params: buffer.Params{
				QuadrantSegments: 0,
				EndCapStyle:      buffer.CapRound,
				JoinStyle:        buffer.JoinRound,
				MitreLimit:       5,
			},
		},
		{
			name: "end cap style",
			params: buffer.Params{
				QuadrantSegments: 8,
				EndCapStyle:      buffer.CapStyle(99),
				JoinStyle:        buffer.JoinRound,
				MitreLimit:       5,
			},
		},
		{
			name: "join style",
			params: buffer.Params{
				QuadrantSegments: 8,
				EndCapStyle:      buffer.CapRound,
				JoinStyle:        buffer.JoinStyle(99),
				MitreLimit:       5,
			},
		},
		{
			name: "mitre limit zero",
			params: buffer.Params{
				QuadrantSegments: 8,
				EndCapStyle:      buffer.CapRound,
				JoinStyle:        buffer.JoinMitre,
				MitreLimit:       0,
			},
		},
		{
			name: "mitre limit infinity",
			params: buffer.Params{
				QuadrantSegments: 8,
				EndCapStyle:      buffer.CapRound,
				JoinStyle:        buffer.JoinMitre,
				MitreLimit:       math.Inf(1),
			},
		},
		{
			name: "mitre limit nan",
			params: buffer.Params{
				QuadrantSegments: 8,
				EndCapStyle:      buffer.CapRound,
				JoinStyle:        buffer.JoinMitre,
				MitreLimit:       math.NaN(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Buffer(point, 1, BufferOptions{Params: &tt.params}); err == nil {
				t.Fatal("expected invalid buffer params error")
			}
		})
	}
}

func TestBufferAllowsValidCustomParams(t *testing.T) {
	point := geom.NewPoint(0, 0)
	params := buffer.Params{
		QuadrantSegments: 4,
		EndCapStyle:      buffer.CapSquare,
		JoinStyle:        buffer.JoinMitre,
		MitreLimit:       2,
	}

	result, err := Buffer(point, 1, BufferOptions{Params: &params})
	if err != nil {
		t.Fatalf("unexpected buffer error: %v", err)
	}
	if result == nil || result.IsEmpty() {
		t.Fatal("expected non-empty buffer result")
	}
}

func TestRelateAndPredicatesValidateInputs(t *testing.T) {
	point := geom.NewPoint(0, 0)
	invalid := geom.NewPoint(math.NaN(), 0)

	if _, err := Relate(nil, point); err == nil {
		t.Fatal("expected nil input error")
	}
	if _, err := Relate(invalid, point); err == nil {
		t.Fatal("expected invalid input error")
	}
	matrix, err := Relate(invalid, point, RelateOptions{AllowInvalidInputs: true})
	if err != nil {
		t.Fatalf("unexpected relate error with AllowInvalidInputs: %v", err)
	}
	if matrix == nil {
		t.Fatal("expected relate matrix")
	}
	if _, err := Intersects(invalid, point); err == nil {
		t.Fatal("expected predicate invalid input error")
	}
	if _, err := Intersects(invalid, point, RelateOptions{AllowInvalidInputs: true}); err != nil {
		t.Fatalf("unexpected predicate error with AllowInvalidInputs: %v", err)
	}
}

func TestRelatePatternAndPredicateHelpers(t *testing.T) {
	left, err := NewPolygon(square(0, 0, 10), nil)
	if err != nil {
		t.Fatalf("unexpected polygon constructor error: %v", err)
	}
	right, err := NewPolygon(square(5, 5, 10), nil)
	if err != nil {
		t.Fatalf("unexpected polygon constructor error: %v", err)
	}
	inside := geom.NewPoint(1, 1)
	outside := geom.NewPoint(20, 20)
	adjacent, err := NewPolygon(square(10, 0, 5), nil)
	if err != nil {
		t.Fatalf("unexpected polygon constructor error: %v", err)
	}
	crossingLine := geom.NewLineString(geom.CoordinateSequence{
		geom.NewCoordinate(-1, 5),
		geom.NewCoordinate(11, 5),
	})

	matrix, err := Relate(inside, left)
	if err != nil {
		t.Fatalf("unexpected relate error: %v", err)
	}
	if matrix.String() == "" {
		t.Fatal("expected non-empty relate matrix string")
	}

	matches, err := RelatePattern(inside, left, "T*F**F***")
	if err != nil {
		t.Fatalf("unexpected relate pattern error: %v", err)
	}
	if !matches {
		t.Fatal("expected point within polygon pattern to match")
	}
	if _, err := RelatePattern(inside, left, "T*F"); err == nil {
		t.Fatal("expected short relate pattern error")
	}
	if _, err := RelatePattern(inside, left, "T*F**F**X"); err == nil {
		t.Fatal("expected invalid relate pattern character error")
	}

	assertPredicate(t, "intersects", true, func() (bool, error) { return Intersects(left, right) })
	assertPredicate(t, "contains", true, func() (bool, error) { return Contains(left, inside) })
	assertPredicate(t, "within", true, func() (bool, error) { return Within(inside, left) })
	assertPredicate(t, "touches", true, func() (bool, error) { return Touches(left, adjacent) })
	assertPredicate(t, "crosses", true, func() (bool, error) { return Crosses(crossingLine, left) })
	assertPredicate(t, "overlaps", true, func() (bool, error) { return Overlaps(left, right) })
	assertPredicate(t, "equals", true, func() (bool, error) { return Equals(left, left.Clone()) })
	assertPredicate(t, "disjoint", true, func() (bool, error) { return Disjoint(left, outside) })
}

func TestStrictWKBWrapperRejectsTrailingBytes(t *testing.T) {
	data, err := WriteWKB(geom.NewPoint(1, 2), wkb.DefaultOptions())
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}
	data = append(data, 0xff)
	if _, err := ReadWKB(data); err == nil {
		t.Fatal("expected trailing byte error")
	}
}

func TestWriteWKTRejectsNil(t *testing.T) {
	if _, err := WriteWKT(nil); err == nil {
		t.Fatal("expected nil geometry error")
	}
}

func TestGeoJSONWrappersRoundTripAndRejectNil(t *testing.T) {
	data, err := WriteGeoJSON(geom.NewPoint(1, 2), GeoJSONOptions{Indent: "  "})
	if err != nil {
		t.Fatalf("unexpected geojson marshal error: %v", err)
	}

	g, err := ReadGeoJSON(data)
	if err != nil {
		t.Fatalf("unexpected geojson unmarshal error: %v", err)
	}
	if g.GeometryType() != "Point" {
		t.Fatalf("expected Point, got %s", g.GeometryType())
	}
	if g.SRID() != 4326 {
		t.Fatalf("expected GeoJSON SRID 4326, got %d", g.SRID())
	}
	if _, err := WriteGeoJSON(nil); err == nil {
		t.Fatal("expected nil geometry error")
	}
}

func TestKMLWrappersRoundTripAndRejectNil(t *testing.T) {
	data, err := WriteKML(geom.NewPoint(-122.084, 37.422), KMLOptions{
		Formatted: true,
		Precision: 3,
	})
	if err != nil {
		t.Fatalf("unexpected kml marshal error: %v", err)
	}

	g, err := ReadKML(data)
	if err != nil {
		t.Fatalf("unexpected kml unmarshal error: %v", err)
	}
	if g.GeometryType() != "Point" {
		t.Fatalf("expected Point, got %s", g.GeometryType())
	}
	if g.SRID() != 4326 {
		t.Fatalf("expected KML SRID 4326, got %d", g.SRID())
	}
	if _, err := WriteKML(nil); err == nil {
		t.Fatal("expected nil geometry error")
	}
}

func assertPredicate(t *testing.T, name string, expected bool, predicate func() (bool, error)) {
	t.Helper()
	actual, err := predicate()
	if err != nil {
		t.Fatalf("%s returned unexpected error: %v", name, err)
	}
	if actual != expected {
		t.Fatalf("%s = %v, want %v", name, actual, expected)
	}
}

func areaOf(t *testing.T, g geom.Geometry) float64 {
	t.Helper()
	areaGeometry, ok := g.(interface{ Area() float64 })
	if !ok {
		t.Fatalf("expected area geometry, got %T", g)
	}
	return areaGeometry.Area()
}

func square(minX, minY, size float64) geom.CoordinateSequence {
	maxX := minX + size
	maxY := minY + size
	return geom.CoordinateSequence{
		geom.NewCoordinate(minX, minY),
		geom.NewCoordinate(maxX, minY),
		geom.NewCoordinate(maxX, maxY),
		geom.NewCoordinate(minX, maxY),
		geom.NewCoordinate(minX, minY),
	}
}
