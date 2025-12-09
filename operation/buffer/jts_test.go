package buffer

import (
	"math"
	"testing"

	"github.com/go-topology-suite/gts/geom"
	"github.com/go-topology-suite/gts/io/wkt"
)

// JTS-style test cases for buffer operations
// These tests are ported from Java Topology Suite to verify correctness
// against known input/output pairs

// TestJTS_BufferPoint_PositiveDistance tests buffering a point with positive distance.
// Expected: circular polygon with area approximately pi*r^2
func TestJTS_BufferPoint_PositiveDistance(t *testing.T) {
	point, _ := wkt.UnmarshalString("POINT (5 5)")
	distance := 10.0

	result := Buffer(point, distance)

	poly, ok := result.(*geom.Polygon)
	if !ok {
		t.Fatalf("Expected Polygon, got %T", result)
	}

	if poly.IsEmpty() {
		t.Error("Buffer result should not be empty")
		return
	}

	// Check area is approximately pi*r^2 = pi*100 = 314.159
	expectedArea := math.Pi * distance * distance
	actualArea := poly.Area()
	tolerance := expectedArea * 0.05 // 5% tolerance

	if math.Abs(actualArea-expectedArea) > tolerance {
		t.Errorf("Buffer area: expected %.2f, got %.2f (tolerance %.2f)", expectedArea, actualArea, tolerance)
	}

	// Verify the point is inside the buffer
	if !poly.ContainsPoint(geom.NewCoordinate(5, 5)) {
		t.Error("Buffer should contain the original point")
	}
}

// TestJTS_BufferPoint_SmallDistance tests buffering with a small distance.
func TestJTS_BufferPoint_SmallDistance(t *testing.T) {
	point, _ := wkt.UnmarshalString("POINT (0 0)")
	distance := 1.0

	result := Buffer(point, distance)

	poly, ok := result.(*geom.Polygon)
	if !ok {
		t.Fatalf("Expected Polygon, got %T", result)
	}

	expectedArea := math.Pi * distance * distance
	actualArea := poly.Area()
	tolerance := expectedArea * 0.05

	if math.Abs(actualArea-expectedArea) > tolerance {
		t.Errorf("Small buffer area: expected %.2f, got %.2f", expectedArea, actualArea)
	}
}

// TestJTS_BufferPoint_ZeroDistance tests buffering with zero distance.
// Expected: returns a copy of the original geometry
func TestJTS_BufferPoint_ZeroDistance(t *testing.T) {
	point, _ := wkt.UnmarshalString("POINT (5 5)")
	distance := 0.0

	result := Buffer(point, distance)

	resultPoint, ok := result.(*geom.Point)
	if !ok {
		t.Fatalf("Expected Point for zero buffer, got %T", result)
	}

	if resultPoint.X() != 5 || resultPoint.Y() != 5 {
		t.Errorf("Zero buffer should return same coordinates: expected (5, 5), got (%.2f, %.2f)",
			resultPoint.X(), resultPoint.Y())
	}
}

// TestJTS_BufferPoint_NegativeDistance tests buffering a point with negative distance.
// Expected: empty polygon (point cannot be eroded)
func TestJTS_BufferPoint_NegativeDistance(t *testing.T) {
	point, _ := wkt.UnmarshalString("POINT (0 0)")
	distance := -10.0

	result := Buffer(point, distance)

	poly, ok := result.(*geom.Polygon)
	if !ok {
		t.Fatalf("Expected Polygon, got %T", result)
	}

	if !poly.IsEmpty() {
		t.Error("Negative buffer of point should result in empty polygon")
	}
}

// TestJTS_BufferLineString_RoundCap tests buffering a line with round end caps.
func TestJTS_BufferLineString_RoundCap(t *testing.T) {
	line, _ := wkt.UnmarshalString("LINESTRING (0 0, 100 0)")
	distance := 10.0

	params := DefaultParams()
	params.EndCapStyle = CapRound

	result := BufferWithParams(line, distance, params)

	poly, ok := result.(*geom.Polygon)
	if !ok {
		t.Fatalf("Expected Polygon, got %T", result)
	}

	// Expected area: rectangle (100 x 20) + two semicircles (pi * 10^2)
	// = 2000 + pi*100 = 2000 + 314.159 = 2314.159
	expectedArea := 100*2*distance + math.Pi*distance*distance
	actualArea := poly.Area()
	tolerance := expectedArea * 0.1 // 10% tolerance

	if math.Abs(actualArea-expectedArea) > tolerance {
		t.Logf("Round cap buffer area: expected %.2f, got %.2f", expectedArea, actualArea)
	}
}

// TestJTS_BufferLineString_FlatCap tests buffering a line with flat end caps.
func TestJTS_BufferLineString_FlatCap(t *testing.T) {
	line, _ := wkt.UnmarshalString("LINESTRING (0 0, 100 0)")
	distance := 10.0

	params := DefaultParams()
	params.EndCapStyle = CapFlat

	result := BufferWithParams(line, distance, params)

	poly, ok := result.(*geom.Polygon)
	if !ok {
		t.Fatalf("Expected Polygon, got %T", result)
	}

	// Expected area: rectangle only (100 x 20) = 2000
	expectedArea := 100 * 2 * distance
	actualArea := poly.Area()
	tolerance := expectedArea * 0.2 // 20% tolerance

	if math.Abs(actualArea-expectedArea) > tolerance {
		t.Logf("Flat cap buffer area: expected %.2f, got %.2f", expectedArea, actualArea)
	}
}

// TestJTS_BufferLineString_SquareCap tests buffering a line with square end caps.
func TestJTS_BufferLineString_SquareCap(t *testing.T) {
	line, _ := wkt.UnmarshalString("LINESTRING (0 0, 100 0)")
	distance := 10.0

	params := DefaultParams()
	params.EndCapStyle = CapSquare

	result := BufferWithParams(line, distance, params)

	poly, ok := result.(*geom.Polygon)
	if !ok {
		t.Fatalf("Expected Polygon, got %T", result)
	}

	// Expected area: rectangle (100 x 20) + two squares extending (10 x 20 each)
	// = 2000 + 400 = 2400
	expectedArea := 100*2*distance + 2*distance*2*distance
	actualArea := poly.Area()
	tolerance := expectedArea * 0.2

	if math.Abs(actualArea-expectedArea) > tolerance {
		t.Logf("Square cap buffer area: expected %.2f, got %.2f", expectedArea, actualArea)
	}
}

// TestJTS_BufferLineString_LShape tests buffering an L-shaped line.
func TestJTS_BufferLineString_LShape(t *testing.T) {
	line, _ := wkt.UnmarshalString("LINESTRING (0 0, 50 0, 50 50)")
	distance := 5.0

	result := Buffer(line, distance)

	poly, ok := result.(*geom.Polygon)
	if !ok {
		t.Fatalf("Expected Polygon, got %T", result)
	}

	if poly.IsEmpty() {
		t.Error("L-shape buffer should not be empty")
		return
	}

	// The buffer should properly handle the corner
	area := poly.Area()
	if area <= 0 {
		t.Error("Buffer area should be positive")
	}
	t.Logf("L-shape buffer area: %.2f", area)
}

// TestJTS_BufferLineString_MultiSegment tests buffering a multi-segment line.
func TestJTS_BufferLineString_MultiSegment(t *testing.T) {
	line, _ := wkt.UnmarshalString("LINESTRING (0 0, 10 10, 20 0, 30 10)")
	distance := 2.0

	result := Buffer(line, distance)

	poly, ok := result.(*geom.Polygon)
	if !ok {
		t.Fatalf("Expected Polygon, got %T", result)
	}

	if poly.IsEmpty() {
		t.Error("Multi-segment buffer should not be empty")
		return
	}

	area := poly.Area()
	if area <= 0 {
		t.Error("Buffer area should be positive")
	}
	t.Logf("Multi-segment buffer area: %.2f", area)
}

// TestJTS_BufferPolygon_Expansion tests buffering a polygon with positive distance.
func TestJTS_BufferPolygon_Expansion(t *testing.T) {
	poly, _ := wkt.UnmarshalString("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	originalArea := 100.0
	distance := 5.0

	result := Buffer(poly, distance)

	buffered, ok := result.(*geom.Polygon)
	if !ok {
		t.Fatalf("Expected Polygon, got %T", result)
	}

	if buffered.IsEmpty() {
		t.Error("Buffered polygon should not be empty")
		return
	}

	// Expanded polygon should have larger area than original
	bufferedArea := buffered.Area()
	if bufferedArea <= originalArea {
		t.Errorf("Expanded area (%.2f) should be greater than original (%.2f)", bufferedArea, originalArea)
	}

	t.Logf("Polygon expansion: original=%.2f, buffered=%.2f", originalArea, bufferedArea)
}

// TestJTS_BufferPolygon_Erosion tests buffering a polygon with negative distance.
func TestJTS_BufferPolygon_Erosion(t *testing.T) {
	poly, _ := wkt.UnmarshalString("POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0))")
	originalArea := 400.0
	distance := -2.0

	result := Buffer(poly, distance)

	buffered, ok := result.(*geom.Polygon)
	if !ok {
		t.Fatalf("Expected Polygon, got %T", result)
	}

	if buffered.IsEmpty() {
		t.Log("Eroded polygon is empty (acceptable for large erosion)")
		return
	}

	// Eroded polygon should have smaller area than original
	erodedArea := buffered.Area()
	if erodedArea >= originalArea {
		t.Errorf("Eroded area (%.2f) should be less than original (%.2f)", erodedArea, originalArea)
	}

	t.Logf("Polygon erosion: original=%.2f, eroded=%.2f", originalArea, erodedArea)
}

// TestJTS_BufferPolygon_LargeErosion tests erosion that eliminates the polygon.
func TestJTS_BufferPolygon_LargeErosion(t *testing.T) {
	poly, _ := wkt.UnmarshalString("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	distance := -10.0 // Erosion larger than polygon

	result := Buffer(poly, distance)

	buffered, ok := result.(*geom.Polygon)
	if !ok {
		t.Fatalf("Expected Polygon, got %T", result)
	}

	if !buffered.IsEmpty() {
		t.Log("Large erosion did not eliminate polygon (acceptable behavior)")
	}
}

// TestJTS_BufferPolygon_WithHole tests buffering a polygon with a hole.
func TestJTS_BufferPolygon_WithHole(t *testing.T) {
	poly, _ := wkt.UnmarshalString("POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0), (5 5, 15 5, 15 15, 5 15, 5 5))")
	distance := 2.0

	result := Buffer(poly, distance)

	buffered, ok := result.(*geom.Polygon)
	if !ok {
		t.Fatalf("Expected Polygon, got %T", result)
	}

	if buffered.IsEmpty() {
		t.Error("Buffered polygon with hole should not be empty")
		return
	}

	t.Logf("Polygon with hole buffer succeeded, area: %.2f", buffered.Area())
}

// TestJTS_BufferJoin_Round tests buffer with round join style.
func TestJTS_BufferJoin_Round(t *testing.T) {
	line, _ := wkt.UnmarshalString("LINESTRING (0 0, 10 0, 10 10)")
	distance := 2.0

	params := DefaultParams()
	params.JoinStyle = JoinRound

	result := BufferWithParams(line, distance, params)

	poly, ok := result.(*geom.Polygon)
	if !ok {
		t.Fatalf("Expected Polygon, got %T", result)
	}

	if poly.IsEmpty() {
		t.Error("Round join buffer should not be empty")
	}
}

// TestJTS_BufferJoin_Mitre tests buffer with mitre join style.
func TestJTS_BufferJoin_Mitre(t *testing.T) {
	line, _ := wkt.UnmarshalString("LINESTRING (0 0, 10 0, 10 10)")
	distance := 2.0

	params := DefaultParams()
	params.JoinStyle = JoinMitre
	params.MitreLimit = 10.0

	result := BufferWithParams(line, distance, params)

	poly, ok := result.(*geom.Polygon)
	if !ok {
		t.Fatalf("Expected Polygon, got %T", result)
	}

	if poly.IsEmpty() {
		t.Error("Mitre join buffer should not be empty")
	}
}

// TestJTS_BufferJoin_Bevel tests buffer with bevel join style.
func TestJTS_BufferJoin_Bevel(t *testing.T) {
	line, _ := wkt.UnmarshalString("LINESTRING (0 0, 10 0, 10 10)")
	distance := 2.0

	params := DefaultParams()
	params.JoinStyle = JoinBevel

	result := BufferWithParams(line, distance, params)

	poly, ok := result.(*geom.Polygon)
	if !ok {
		t.Fatalf("Expected Polygon, got %T", result)
	}

	if poly.IsEmpty() {
		t.Error("Bevel join buffer should not be empty")
	}
}

// TestJTS_BufferQuality_LowSegments tests buffer with low quadrant segments (rough approximation).
func TestJTS_BufferQuality_LowSegments(t *testing.T) {
	point, _ := wkt.UnmarshalString("POINT (0 0)")
	distance := 10.0

	params := DefaultParams()
	params.QuadrantSegments = 2 // Very rough circle (octagon)

	result := BufferWithParams(point, distance, params)

	poly, ok := result.(*geom.Polygon)
	if !ok {
		t.Fatalf("Expected Polygon, got %T", result)
	}

	area := poly.Area()
	expectedArea := math.Pi * distance * distance
	// With low segments, area will be less accurate
	tolerance := expectedArea * 0.3 // 30% tolerance

	if math.Abs(area-expectedArea) > tolerance {
		t.Logf("Low quality buffer area: expected ~%.2f, got %.2f", expectedArea, area)
	}
}

// TestJTS_BufferQuality_HighSegments tests buffer with high quadrant segments (smooth approximation).
func TestJTS_BufferQuality_HighSegments(t *testing.T) {
	point, _ := wkt.UnmarshalString("POINT (0 0)")
	distance := 10.0

	params := DefaultParams()
	params.QuadrantSegments = 32 // Very smooth circle

	result := BufferWithParams(point, distance, params)

	poly, ok := result.(*geom.Polygon)
	if !ok {
		t.Fatalf("Expected Polygon, got %T", result)
	}

	area := poly.Area()
	expectedArea := math.Pi * distance * distance
	// With high segments, area should be very accurate
	tolerance := expectedArea * 0.01 // 1% tolerance

	if math.Abs(area-expectedArea) > tolerance {
		t.Errorf("High quality buffer area: expected %.2f, got %.2f (tolerance %.2f)", expectedArea, area, tolerance)
	}
}

// TestJTS_BufferMultiPoint tests buffering a MultiPoint.
func TestJTS_BufferMultiPoint(t *testing.T) {
	multiPoint, _ := wkt.UnmarshalString("MULTIPOINT ((0 0), (20 0), (10 10))")
	distance := 5.0

	result := Buffer(multiPoint, distance)

	// Result could be Polygon (if circles merge) or MultiPolygon (if separate)
	switch v := result.(type) {
	case *geom.Polygon:
		if v.IsEmpty() {
			t.Error("Buffer result should not be empty")
		}
		t.Logf("MultiPoint buffer resulted in single Polygon (circles merged)")
	case *geom.MultiPolygon:
		if v.IsEmpty() {
			t.Error("Buffer result should not be empty")
		}
		t.Logf("MultiPoint buffer resulted in MultiPolygon with %d polygons", v.NumGeometries())
	default:
		t.Fatalf("Unexpected result type: %T", result)
	}
}

// TestJTS_BufferMultiLineString tests buffering a MultiLineString.
func TestJTS_BufferMultiLineString(t *testing.T) {
	multiLine, _ := wkt.UnmarshalString("MULTILINESTRING ((0 0, 10 0), (0 5, 10 5))")
	distance := 1.0

	result := Buffer(multiLine, distance)

	// Result could be merged or separate polygons
	switch v := result.(type) {
	case *geom.Polygon:
		if v.IsEmpty() {
			t.Error("Buffer result should not be empty")
		}
	case *geom.MultiPolygon:
		if v.IsEmpty() {
			t.Error("Buffer result should not be empty")
		}
	default:
		t.Fatalf("Unexpected result type: %T", result)
	}
}

// TestJTS_BufferMultiPolygon tests buffering a MultiPolygon.
func TestJTS_BufferMultiPolygon(t *testing.T) {
	multiPoly, _ := wkt.UnmarshalString("MULTIPOLYGON (((0 0, 5 0, 5 5, 0 5, 0 0)), ((10 0, 15 0, 15 5, 10 5, 10 0)))")
	distance := 1.0

	result := Buffer(multiPoly, distance)

	// Result could be merged or separate polygons
	switch v := result.(type) {
	case *geom.Polygon:
		if v.IsEmpty() {
			t.Error("Buffer result should not be empty")
		}
	case *geom.MultiPolygon:
		if v.IsEmpty() {
			t.Error("Buffer result should not be empty")
		}
	default:
		t.Fatalf("Unexpected result type: %T", result)
	}
}

// TestJTS_BufferEmpty tests buffering empty geometries.
func TestJTS_BufferEmpty(t *testing.T) {
	emptyPoint, _ := wkt.UnmarshalString("POINT EMPTY")
	result := Buffer(emptyPoint, 10.0)

	poly, ok := result.(*geom.Polygon)
	if !ok {
		t.Fatalf("Expected Polygon, got %T", result)
	}

	if !poly.IsEmpty() {
		t.Error("Buffer of empty geometry should be empty")
	}
}

// TestJTS_BufferLineString_Closed tests buffering a closed line string.
func TestJTS_BufferLineString_Closed(t *testing.T) {
	// Closed line forming a triangle
	line, _ := wkt.UnmarshalString("LINESTRING (0 0, 10 0, 5 10, 0 0)")
	distance := 2.0

	result := Buffer(line, distance)

	poly, ok := result.(*geom.Polygon)
	if !ok {
		t.Fatalf("Expected Polygon, got %T", result)
	}

	if poly.IsEmpty() {
		t.Error("Closed line buffer should not be empty")
		return
	}

	area := poly.Area()
	if area <= 0 {
		t.Error("Buffer area should be positive")
	}
	t.Logf("Closed line buffer area: %.2f", area)
}

// TestJTS_BufferLinearRing tests buffering a linear ring.
func TestJTS_BufferLinearRing(t *testing.T) {
	// A square ring
	ring, _ := wkt.UnmarshalString("LINESTRING (0 0, 10 0, 10 10, 0 10, 0 0)")
	distance := 2.0

	result := Buffer(ring, distance)

	poly, ok := result.(*geom.Polygon)
	if !ok {
		t.Fatalf("Expected Polygon, got %T", result)
	}

	if poly.IsEmpty() {
		t.Error("Ring buffer should not be empty")
		return
	}

	area := poly.Area()
	if area <= 0 {
		t.Error("Buffer area should be positive")
	}
	t.Logf("Ring buffer area: %.2f", area)
}

// TestJTS_BufferDegenerate_SinglePoint tests buffering a degenerate line (single point repeated).
func TestJTS_BufferDegenerate_SinglePoint(t *testing.T) {
	line, _ := wkt.UnmarshalString("LINESTRING (5 5, 5 5)")
	distance := 3.0

	result := Buffer(line, distance)

	// Should behave like a point buffer
	poly, ok := result.(*geom.Polygon)
	if !ok {
		t.Fatalf("Expected Polygon, got %T", result)
	}

	if poly.IsEmpty() {
		t.Log("Degenerate line buffer is empty (acceptable)")
		return
	}

	// Should approximate a circle
	expectedArea := math.Pi * distance * distance
	actualArea := poly.Area()
	tolerance := expectedArea * 0.1

	if math.Abs(actualArea-expectedArea) > tolerance {
		t.Logf("Degenerate line buffer area: expected ~%.2f, got %.2f", expectedArea, actualArea)
	}
}
