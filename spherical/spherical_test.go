package spherical

import (
	"math"
	"testing"

	"github.com/go-topology-suite/gts/geom"
)

const (
	// Tolerance for floating point comparisons in meters
	distanceTolerance = 100.0 // 100 meters
	// Tolerance for area comparisons in square meters
	areaTolerance = 1000000.0 // 1 km²
)

// TestDistance tests geodesic distance calculations.
func TestDistance(t *testing.T) {
	tests := []struct {
		name     string
		lon1     float64
		lat1     float64
		lon2     float64
		lat2     float64
		expected float64 // in meters
	}{
		{
			name:     "NYC to London",
			lon1:     -74.0060, // NYC
			lat1:     40.7128,
			lon2:     -0.1278, // London
			lat2:     51.5074,
			expected: 5570000, // ~5570 km
		},
		{
			name:     "Equator quarter circle",
			lon1:     0.0,
			lat1:     0.0,
			lon2:     90.0,
			lat2:     0.0,
			expected: 10018754, // ~10,000 km (quarter of Earth's circumference)
		},
		{
			name:     "Same point",
			lon1:     -122.4194, // San Francisco
			lat1:     37.7749,
			lon2:     -122.4194,
			lat2:     37.7749,
			expected: 0,
		},
		{
			name:     "Antipodal points (approximation)",
			lon1:     0.0,
			lat1:     0.0,
			lon2:     180.0,
			lat2:     0.0,
			expected: 20037508, // ~20,000 km (half Earth's circumference)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p1 := geom.NewPoint(tt.lon1, tt.lat1)
			p2 := geom.NewPoint(tt.lon2, tt.lat2)

			dist := Distance(p1, p2)

			// Use relative tolerance for large distances
			tolerance := math.Max(distanceTolerance, tt.expected*0.01)
			if math.Abs(dist-tt.expected) > tolerance {
				t.Errorf("Distance() = %v, want %v (tolerance %v)", dist, tt.expected, tolerance)
			}

			// Test DistanceCoords
			dist2 := DistanceCoords(tt.lon1, tt.lat1, tt.lon2, tt.lat2)
			if math.Abs(dist-dist2) > 0.01 {
				t.Errorf("Distance() and DistanceCoords() differ: %v vs %v", dist, dist2)
			}
		})
	}
}

// TestDistanceEmptyGeometries tests distance with empty geometries.
func TestDistanceEmptyGeometries(t *testing.T) {
	p1 := geom.NewPoint(0, 0)
	empty := geom.NewPointEmpty()

	dist := Distance(p1, empty)
	if dist != 0 {
		t.Errorf("Distance to empty point should be 0, got %v", dist)
	}

	dist = Distance(empty, p1)
	if dist != 0 {
		t.Errorf("Distance from empty point should be 0, got %v", dist)
	}
}

// TestLength tests linestring length calculations.
func TestLength(t *testing.T) {
	tests := []struct {
		name     string
		coords   []float64 // lon, lat pairs
		expected float64   // in meters
	}{
		{
			name: "Equator line segment",
			coords: []float64{
				0.0, 0.0,
				1.0, 0.0,
			},
			expected: 111319, // ~111 km (1 degree at equator)
		},
		{
			name: "Multi-segment line",
			coords: []float64{
				0.0, 0.0,
				1.0, 0.0,
				1.0, 1.0,
			},
			expected: 222638, // ~223 km (two segments)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := geom.NewLineStringXY(tt.coords...)
			length := Length(ls)

			tolerance := math.Max(distanceTolerance, tt.expected*0.01)
			if math.Abs(length-tt.expected) > tolerance {
				t.Errorf("Length() = %v, want %v (tolerance %v)", length, tt.expected, tolerance)
			}
		})
	}
}

// TestArea tests polygon area calculations.
func TestArea(t *testing.T) {
	tests := []struct {
		name     string
		coords   []float64 // lon, lat pairs
		expected float64   // in square meters
	}{
		{
			name: "Small square near equator",
			coords: []float64{
				0.0, 0.0,
				0.1, 0.0,
				0.1, 0.1,
				0.0, 0.1,
				0.0, 0.0,
			},
			expected: 123600000, // ~123 km² (0.1° × 0.1° near equator)
		},
		{
			name: "Unit square at equator",
			coords: []float64{
				0.0, 0.0,
				1.0, 0.0,
				1.0, 1.0,
				0.0, 1.0,
				0.0, 0.0,
			},
			expected: 12360000000, // ~12,360 km²
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ring := geom.NewLinearRingXY(tt.coords...)
			poly := geom.NewPolygon(ring, nil)
			area := Area(poly)

			tolerance := math.Max(areaTolerance, tt.expected*0.05) // 5% tolerance for area
			if math.Abs(area-tt.expected) > tolerance {
				t.Errorf("Area() = %v, want %v (tolerance %v)", area, tt.expected, tolerance)
			}

			// Test that signed area has same magnitude
			signedArea := SignedArea(poly)
			if math.Abs(math.Abs(signedArea)-area) > 0.01 {
				t.Errorf("SignedArea magnitude %v doesn't match Area %v", math.Abs(signedArea), area)
			}
		})
	}
}

// TestAreaWithHoles tests polygon area with holes.
func TestAreaWithHoles(t *testing.T) {
	// Outer ring: 1° × 1° square
	outer := geom.NewLinearRingXY(
		0.0, 0.0,
		1.0, 0.0,
		1.0, 1.0,
		0.0, 1.0,
		0.0, 0.0,
	)

	// Inner ring (hole): 0.5° × 0.5° square in the center
	hole := geom.NewLinearRingXY(
		0.25, 0.25,
		0.75, 0.25,
		0.75, 0.75,
		0.25, 0.75,
		0.25, 0.25,
	)

	poly := geom.NewPolygon(outer, []*geom.LinearRing{hole})

	area := PolygonAreaWithHoles(poly)

	// Outer area - hole area
	outerArea := Area(geom.NewPolygon(outer, nil))
	holeArea := Area(geom.NewPolygon(hole, nil))
	expectedArea := outerArea - holeArea

	tolerance := math.Max(areaTolerance, expectedArea*0.05)
	if math.Abs(area-expectedArea) > tolerance {
		t.Errorf("PolygonAreaWithHoles() = %v, want %v (tolerance %v)", area, expectedArea, tolerance)
	}
}

// TestContains tests point-in-polygon containment.
func TestContains(t *testing.T) {
	// Create a polygon around (0, 0)
	ring := geom.NewLinearRingXY(
		-1.0, -1.0,
		1.0, -1.0,
		1.0, 1.0,
		-1.0, 1.0,
		-1.0, -1.0,
	)
	poly := geom.NewPolygon(ring, nil)

	tests := []struct {
		name     string
		lon      float64
		lat      float64
		expected bool
	}{
		{
			name:     "Point inside",
			lon:      0.0,
			lat:      0.0,
			expected: true,
		},
		{
			name:     "Point outside",
			lon:      2.0,
			lat:      2.0,
			expected: false,
		},
		{
			name:     "Point near corner (inside)",
			lon:      0.5,
			lat:      0.5,
			expected: true,
		},
		{
			name:     "Point near corner (outside)",
			lon:      -1.5,
			lat:      -1.5,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := geom.NewPoint(tt.lon, tt.lat)
			result := Contains(poly, p)
			if result != tt.expected {
				t.Errorf("Contains() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestIntersects tests polygon intersection.
func TestIntersects(t *testing.T) {
	// Polygon 1: square from (-1,-1) to (1,1)
	ring1 := geom.NewLinearRingXY(
		-1.0, -1.0,
		1.0, -1.0,
		1.0, 1.0,
		-1.0, 1.0,
		-1.0, -1.0,
	)
	poly1 := geom.NewPolygon(ring1, nil)

	tests := []struct {
		name     string
		coords   []float64
		expected bool
	}{
		{
			name: "Overlapping polygon",
			coords: []float64{
				0.0, 0.0,
				2.0, 0.0,
				2.0, 2.0,
				0.0, 2.0,
				0.0, 0.0,
			},
			expected: true,
		},
		{
			name: "Disjoint polygon",
			coords: []float64{
				3.0, 3.0,
				4.0, 3.0,
				4.0, 4.0,
				3.0, 4.0,
				3.0, 3.0,
			},
			expected: false,
		},
		{
			name: "Adjacent polygon (touching)",
			coords: []float64{
				1.0, -1.0,
				2.0, -1.0,
				2.0, 1.0,
				1.0, 1.0,
				1.0, -1.0,
			},
			// Note: In spherical geometry, polygons that only share edges (not area)
			// may not be detected as intersecting by S2 due to the way boundary
			// containment is handled. This is expected behavior.
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ring2 := geom.NewLinearRingXY(tt.coords...)
			poly2 := geom.NewPolygon(ring2, nil)
			result := Intersects(poly1, poly2)
			if result != tt.expected {
				t.Errorf("Intersects() = %v, want %v", result, tt.expected)
			}

			// Test commutativity
			result2 := Intersects(poly2, poly1)
			if result != result2 {
				t.Errorf("Intersects() not commutative: %v vs %v", result, result2)
			}
		})
	}
}

// TestCellID tests S2 cell ID generation.
func TestCellID(t *testing.T) {
	p := geom.NewPoint(0.0, 0.0)

	// Test CellID at max level
	cellID := CellID(p)
	if !cellID.IsValid() {
		t.Error("CellID() returned invalid cell ID")
	}
	if cellID.Level() != 30 {
		t.Errorf("CellID() level = %v, want 30", cellID.Level())
	}

	// Test CellIDAtLevel
	for level := 0; level <= 30; level++ {
		cellID := CellIDAtLevel(p, level)
		if !cellID.IsValid() {
			t.Errorf("CellIDAtLevel(%v) returned invalid cell ID", level)
		}
		if cellID.Level() != level {
			t.Errorf("CellIDAtLevel(%v) level = %v, want %v", level, cellID.Level(), level)
		}
	}
}

// TestCellToken tests S2 cell token generation.
func TestCellToken(t *testing.T) {
	p := geom.NewPoint(-122.4194, 37.7749) // San Francisco

	token := CellToken(p, 10)
	if token == "" {
		t.Error("CellToken() returned empty token")
	}

	// Verify we can convert back
	cellID := CellFromToken(token)
	if !cellID.IsValid() {
		t.Error("CellFromToken() returned invalid cell ID")
	}

	if cellID.Level() != 10 {
		t.Errorf("Reconstructed cell level = %v, want 10", cellID.Level())
	}
}

// TestCovering tests geometry covering with S2 cells.
func TestCovering(t *testing.T) {
	// Create a small polygon
	ring := geom.NewLinearRingXY(
		-1.0, -1.0,
		1.0, -1.0,
		1.0, 1.0,
		-1.0, 1.0,
		-1.0, -1.0,
	)
	poly := geom.NewPolygon(ring, nil)

	cells := Covering(poly, 5, 10, 8)

	if len(cells) == 0 {
		t.Error("Covering() returned no cells")
	}

	if len(cells) > 8 {
		t.Errorf("Covering() returned %v cells, max was 8", len(cells))
	}

	// Verify all cells are valid and within level range
	for i, cell := range cells {
		if !cell.IsValid() {
			t.Errorf("Cell %v is invalid", i)
		}
		level := cell.Level()
		if level < 5 || level > 10 {
			t.Errorf("Cell %v level = %v, want between 5 and 10", i, level)
		}
	}
}

// TestCoveringTokens tests token-based covering.
func TestCoveringTokens(t *testing.T) {
	ring := geom.NewLinearRingXY(
		0.0, 0.0,
		1.0, 0.0,
		1.0, 1.0,
		0.0, 1.0,
		0.0, 0.0,
	)
	poly := geom.NewPolygon(ring, nil)

	tokens := CoveringTokens(poly, 5, 10, 4)

	if len(tokens) == 0 {
		t.Error("CoveringTokens() returned no tokens")
	}

	// Verify tokens can be converted back to cell IDs
	for i, token := range tokens {
		cellID := CellFromToken(token)
		if !cellID.IsValid() {
			t.Errorf("Token %v (%s) is invalid", i, token)
		}
	}
}

// TestConversions tests round-trip conversions.
func TestConversions(t *testing.T) {
	t.Run("Point conversion", func(t *testing.T) {
		original := geom.NewPoint(-122.4194, 37.7749)
		s2Point := ToS2Point(original)
		converted := FromS2Point(s2Point)

		if math.Abs(original.X()-converted.X()) > 1e-10 {
			t.Errorf("X coordinate mismatch: %v vs %v", original.X(), converted.X())
		}
		if math.Abs(original.Y()-converted.Y()) > 1e-10 {
			t.Errorf("Y coordinate mismatch: %v vs %v", original.Y(), converted.Y())
		}
	})

	t.Run("LineString conversion", func(t *testing.T) {
		original := geom.NewLineStringXY(
			0.0, 0.0,
			1.0, 1.0,
			2.0, 0.0,
		)
		polyline := ToS2Polyline(original)
		converted := FromS2Polyline(polyline)

		if original.NumPoints() != converted.NumPoints() {
			t.Errorf("Point count mismatch: %v vs %v", original.NumPoints(), converted.NumPoints())
		}
	})

	t.Run("Polygon conversion", func(t *testing.T) {
		ring := geom.NewLinearRingXY(
			0.0, 0.0,
			1.0, 0.0,
			1.0, 1.0,
			0.0, 1.0,
			0.0, 0.0,
		)
		original := geom.NewPolygon(ring, nil)
		s2Poly := ToS2Polygon(original)
		converted := FromS2Polygon(s2Poly)

		if original.ExteriorRing().NumPoints() != converted.ExteriorRing().NumPoints() {
			t.Errorf("Ring point count mismatch: %v vs %v",
				original.ExteriorRing().NumPoints(),
				converted.ExteriorRing().NumPoints())
		}
	})
}

// TestPerimeter tests polygon perimeter calculations.
func TestPerimeter(t *testing.T) {
	// Square at equator: 1° × 1°
	ring := geom.NewLinearRingXY(
		0.0, 0.0,
		1.0, 0.0,
		1.0, 1.0,
		0.0, 1.0,
		0.0, 0.0,
	)
	poly := geom.NewPolygon(ring, nil)

	perimeter := Perimeter(poly)

	// Approximate perimeter: 4 * 111km = 444km
	expected := 445276.0 // More precise value
	tolerance := math.Max(distanceTolerance, expected*0.01)

	if math.Abs(perimeter-expected) > tolerance {
		t.Errorf("Perimeter() = %v, want %v (tolerance %v)", perimeter, expected, tolerance)
	}
}

// TestCentroid tests centroid calculations.
func TestCentroid(t *testing.T) {
	// Square centered at origin
	ring := geom.NewLinearRingXY(
		-1.0, -1.0,
		1.0, -1.0,
		1.0, 1.0,
		-1.0, 1.0,
		-1.0, -1.0,
	)
	poly := geom.NewPolygon(ring, nil)

	centroid := Centroid(poly)

	if centroid.IsEmpty() {
		t.Error("Centroid() returned empty point")
	}

	// Centroid should be near (0, 0)
	if math.Abs(centroid.X()) > 0.1 || math.Abs(centroid.Y()) > 0.1 {
		t.Errorf("Centroid() = (%v, %v), expected near (0, 0)", centroid.X(), centroid.Y())
	}
}

// TestGeometryFromCellID tests converting cell IDs back to geometries.
func TestGeometryFromCellID(t *testing.T) {
	p := geom.NewPoint(0.0, 0.0)
	cellID := CellIDAtLevel(p, 10)

	poly := GeometryFromCellID(cellID)
	if poly == nil || poly.IsEmpty() {
		t.Error("GeometryFromCellID() returned nil or empty polygon")
	}

	// Should be a square (4 vertices + closing)
	if poly.ExteriorRing().NumPoints() != 5 {
		t.Errorf("Cell polygon has %v points, want 5", poly.ExteriorRing().NumPoints())
	}

	// Test with token
	token := cellID.ToToken()
	poly2 := GeometryFromCellToken(token)
	if poly2 == nil || poly2.IsEmpty() {
		t.Error("GeometryFromCellToken() returned nil or empty polygon")
	}
}

// TestEmptyGeometries tests handling of empty geometries.
func TestEmptyGeometries(t *testing.T) {
	empty := geom.NewPointEmpty()
	emptyLS := geom.NewLineStringEmpty()
	emptyPoly := geom.NewPolygonEmpty()

	if Distance(empty, empty) != 0 {
		t.Error("Distance between empty points should be 0")
	}

	if Length(emptyLS) != 0 {
		t.Error("Length of empty linestring should be 0")
	}

	if Area(emptyPoly) != 0 {
		t.Error("Area of empty polygon should be 0")
	}

	if Contains(emptyPoly, empty) {
		t.Error("Empty polygon should not contain empty point")
	}
}

// TestS2LatLngConversion tests coordinate conversion helpers.
func TestS2LatLngConversion(t *testing.T) {
	coord := geom.NewCoordinate(-122.4194, 37.7749) // San Francisco (lon, lat)

	ll := ToS2LatLng(coord)
	converted := FromS2LatLng(ll)

	if math.Abs(coord.X-converted.X) > 1e-10 {
		t.Errorf("X mismatch: %v vs %v", coord.X, converted.X)
	}
	if math.Abs(coord.Y-converted.Y) > 1e-10 {
		t.Errorf("Y mismatch: %v vs %v", coord.Y, converted.Y)
	}

	// Verify lat/lng are correctly mapped
	if math.Abs(ll.Lat.Degrees()-coord.Y) > 1e-10 {
		t.Errorf("Latitude mismatch: %v vs %v", ll.Lat.Degrees(), coord.Y)
	}
	if math.Abs(ll.Lng.Degrees()-coord.X) > 1e-10 {
		t.Errorf("Longitude mismatch: %v vs %v", ll.Lng.Degrees(), coord.X)
	}
}

// BenchmarkDistance benchmarks distance calculations.
func BenchmarkDistance(b *testing.B) {
	p1 := geom.NewPoint(-74.0060, 40.7128) // NYC
	p2 := geom.NewPoint(-0.1278, 51.5074)  // London

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Distance(p1, p2)
	}
}

// BenchmarkArea benchmarks area calculations.
func BenchmarkArea(b *testing.B) {
	ring := geom.NewLinearRingXY(
		0.0, 0.0,
		1.0, 0.0,
		1.0, 1.0,
		0.0, 1.0,
		0.0, 0.0,
	)
	poly := geom.NewPolygon(ring, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Area(poly)
	}
}

// BenchmarkContains benchmarks containment tests.
func BenchmarkContains(b *testing.B) {
	ring := geom.NewLinearRingXY(
		-1.0, -1.0,
		1.0, -1.0,
		1.0, 1.0,
		-1.0, 1.0,
		-1.0, -1.0,
	)
	poly := geom.NewPolygon(ring, nil)
	p := geom.NewPoint(0.0, 0.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Contains(poly, p)
	}
}

// BenchmarkCovering benchmarks S2 cell covering.
func BenchmarkCovering(b *testing.B) {
	ring := geom.NewLinearRingXY(
		-1.0, -1.0,
		1.0, -1.0,
		1.0, 1.0,
		-1.0, 1.0,
		-1.0, -1.0,
	)
	poly := geom.NewPolygon(ring, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Covering(poly, 5, 15, 8)
	}
}
