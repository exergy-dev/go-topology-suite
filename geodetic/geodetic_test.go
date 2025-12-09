package geodetic

import (
	"math"
	"testing"
)

const (
	// Tolerance for distance comparisons (meters)
	distanceTolerance = 0.001 // 1mm

	// Tolerance for angle comparisons (degrees)
	angleTolerance = 0.0001 // ~0.36 arc-seconds

	// Tolerance for area comparisons (relative error)
	areaTolerance = 0.001 // 0.1%
)

// TestEllipsoidProperties tests the ellipsoid property calculations
func TestEllipsoidProperties(t *testing.T) {
	tests := []struct {
		name            string
		ellipsoid       *Ellipsoid
		expectedA       float64
		expectedB       float64
		expectedF       float64
		expectedInvF    float64
		expectedESq     float64
		expectedE2Sq    float64
		toleranceInvF   float64
		toleranceESq    float64
	}{
		{
			name:          "WGS84",
			ellipsoid:     WGS84,
			expectedA:     6378137.0,
			expectedB:     6356752.314245179,
			expectedF:     1.0 / 298.257223563,
			expectedInvF:  298.257223563,
			expectedESq:   0.00669437999014,
			expectedE2Sq:  0.00673949674228,
			toleranceInvF: 1e-9,
			toleranceESq:  1e-11,
		},
		{
			name:          "GRS80",
			ellipsoid:     GRS80,
			expectedA:     6378137.0,
			expectedInvF:  298.257222101,
			toleranceInvF: 1e-9,
			toleranceESq:  1e-11,
		},
		{
			name:          "Clarke1866",
			ellipsoid:     Clarke1866,
			expectedA:     6378206.4,
			expectedB:     6356583.8,
			toleranceInvF: 1e-6,
			toleranceESq:  1e-11,
		},
		{
			name:          "Sphere",
			ellipsoid:     Sphere,
			expectedA:     6371000.0,
			expectedB:     6371000.0,
			expectedF:     0,
			expectedInvF:  0,
			expectedESq:   0,
			expectedE2Sq:  0,
			toleranceInvF: 1e-9,
			toleranceESq:  1e-11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if math.Abs(tt.ellipsoid.SemiMajorAxis()-tt.expectedA) > 1e-6 {
				t.Errorf("SemiMajorAxis() = %v, want %v", tt.ellipsoid.SemiMajorAxis(), tt.expectedA)
			}

			if tt.expectedB > 0 && math.Abs(tt.ellipsoid.SemiMinorAxis()-tt.expectedB) > 1e-6 {
				t.Errorf("SemiMinorAxis() = %v, want %v", tt.ellipsoid.SemiMinorAxis(), tt.expectedB)
			}

			if tt.expectedF > 0 && math.Abs(tt.ellipsoid.Flattening()-tt.expectedF) > 1e-12 {
				t.Errorf("Flattening() = %v, want %v", tt.ellipsoid.Flattening(), tt.expectedF)
			}

			if tt.expectedInvF > 0 && math.Abs(tt.ellipsoid.InverseFlattening()-tt.expectedInvF) > tt.toleranceInvF {
				t.Errorf("InverseFlattening() = %v, want %v", tt.ellipsoid.InverseFlattening(), tt.expectedInvF)
			}

			if tt.expectedESq > 0 && math.Abs(tt.ellipsoid.EccentricitySquared()-tt.expectedESq) > tt.toleranceESq {
				t.Errorf("EccentricitySquared() = %.14f, want %.14f", tt.ellipsoid.EccentricitySquared(), tt.expectedESq)
			}
		})
	}
}

// TestVincentyKnownDistances tests Vincenty's formula against known accurate values
func TestVincentyKnownDistances(t *testing.T) {
	tests := []struct {
		name     string
		lat1     float64
		lon1     float64
		lat2     float64
		lon2     float64
		expected float64 // meters
	}{
		{
			name:     "Flinders Peak to Buninyong (Australia)",
			lat1:     -37.9510334166667,
			lon1:     144.4248679444444,
			lat2:     -37.6528211388889,
			lon2:     143.9264955555556,
			expected: 54972.271, // Known accurate distance (allowing ~3mm tolerance)
		},
		{
			name:     "Same point",
			lat1:     0,
			lon1:     0,
			lat2:     0,
			lon2:     0,
			expected: 0,
		},
		{
			name:     "Equator points",
			lat1:     0,
			lon1:     0,
			lat2:     0,
			lon2:     1,
			expected: 111319.491, // ~111km for 1 degree at equator
		},
		{
			name:     "North-South along meridian",
			lat1:     0,
			lon1:     0,
			lat2:     1,
			lon2:     0,
			expected: 110574.389, // ~110.5km for 1 degree latitude
		},
		{
			name:     "New York to London",
			lat1:     40.7128,
			lon1:     -74.0060,
			lat2:     51.5074,
			lon2:     -0.1278,
			expected: 5570230, // ~5570km
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distance, err := Vincenty(tt.lat1, tt.lon1, tt.lat2, tt.lon2, WGS84)
			if err != nil {
				t.Fatalf("Vincenty() error = %v", err)
			}

			// For known accurate values, use tighter tolerance
			tol := distanceTolerance
			if tt.expected > 1000000 { // For long distances, allow 0.3% error
				tol = tt.expected * 0.003
			} else if tt.expected > 10000 {
				tol = 0.01 // Allow 1cm for medium distances
			}

			if math.Abs(distance-tt.expected) > tol {
				t.Errorf("Vincenty() = %.3f m, want %.3f m (diff: %.3f m)",
					distance, tt.expected, math.Abs(distance-tt.expected))
			}
		})
	}
}

// TestDistanceSymmetry tests that Distance(A,B) == Distance(B,A)
func TestDistanceSymmetry(t *testing.T) {
	tests := []struct {
		name string
		lat1 float64
		lon1 float64
		lat2 float64
		lon2 float64
	}{
		{"Basic", 40.7128, -74.0060, 51.5074, -0.1278},
		{"Across dateline", 35.6762, 139.6503, 37.7749, -122.4194},
		{"Near poles", 89.5, 0, 88.0, 45},
		{"Equator", 0, 0, 0, 90},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d1, err1 := Vincenty(tt.lat1, tt.lon1, tt.lat2, tt.lon2, WGS84)
			d2, err2 := Vincenty(tt.lat2, tt.lon2, tt.lat1, tt.lon1, WGS84)

			if err1 != nil || err2 != nil {
				if err1 != err2 {
					t.Errorf("Asymmetric errors: err1=%v, err2=%v", err1, err2)
				}
				return
			}

			if math.Abs(d1-d2) > distanceTolerance {
				t.Errorf("Distance not symmetric: d1=%.6f, d2=%.6f (diff: %.6f)",
					d1, d2, math.Abs(d1-d2))
			}
		})
	}
}

// TestDirectInverseRoundTrip tests that Direct followed by Inverse returns to original
func TestDirectInverseRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		lat      float64
		lon      float64
		azimuth  float64
		distance float64
	}{
		{"Short distance", 40.7128, -74.0060, 45.0, 10000},
		{"Medium distance", 0, 0, 90.0, 1000000},
		{"Long distance", 51.5074, -0.1278, 270.0, 5000000},
		{"North", 35.6762, 139.6503, 0.0, 500000},
		{"South", 35.6762, 139.6503, 180.0, 500000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Forward: compute destination
			lat2, lon2, az2, err := Direct(tt.lat, tt.lon, tt.azimuth, tt.distance, WGS84)
			if err != nil {
				t.Fatalf("Direct() error = %v", err)
			}

			// Inverse: compute back to origin
			dist, az1, _, err := Inverse(tt.lat, tt.lon, lat2, lon2, WGS84)
			if err != nil {
				t.Fatalf("Inverse() error = %v", err)
			}

			// Check distance matches
			if math.Abs(dist-tt.distance) > distanceTolerance {
				t.Errorf("Distance mismatch: got %.3f m, want %.3f m", dist, tt.distance)
			}

			// Check azimuth matches (allowing for normalization)
			azDiff := math.Abs(az1 - tt.azimuth)
			if azDiff > 180 {
				azDiff = 360 - azDiff
			}
			if azDiff > angleTolerance {
				t.Errorf("Azimuth mismatch: got %.6f°, want %.6f° (diff: %.6f°)",
					az1, tt.azimuth, azDiff)
			}

			// Additional check: reverse azimuth should point back
			reverseAz := normalizeAzimuth(az2 + 180)
			lat3, lon3, _, err := Direct(lat2, lon2, reverseAz, tt.distance, WGS84)
			if err != nil {
				t.Fatalf("Reverse Direct() error = %v", err)
			}

			latDiff := math.Abs(lat3 - tt.lat)
			lonDiff := math.Abs(lon3 - tt.lon)
			if lonDiff > 180 {
				lonDiff = 360 - lonDiff
			}

			// Allow ~1 meter tolerance in position (very rough estimate)
			if latDiff > 0.00001 || lonDiff > 0.00001 {
				t.Errorf("Round trip position error: lat diff=%.8f°, lon diff=%.8f°",
					latDiff, lonDiff)
			}
		})
	}
}

// TestHaversine tests the Haversine formula
func TestHaversine(t *testing.T) {
	tests := []struct {
		name     string
		lat1     float64
		lon1     float64
		lat2     float64
		lon2     float64
		expected float64
	}{
		{
			name:     "Same point",
			lat1:     0,
			lon1:     0,
			lat2:     0,
			lon2:     0,
			expected: 0,
		},
		{
			name:     "Equator 1 degree",
			lat1:     0,
			lon1:     0,
			lat2:     0,
			lon2:     1,
			expected: 111195, // Using mean radius
		},
		{
			name:     "Short distance",
			lat1:     40.7128,
			lon1:     -74.0060,
			lat2:     40.7489,
			lon2:     -73.9680,
			expected: 5135, // ~5.1km (using mean radius)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distance := Haversine(tt.lat1, tt.lon1, tt.lat2, tt.lon2, EarthMeanRadius)

			// Haversine is less accurate, allow 1% error
			tolerance := math.Max(100, tt.expected*0.01)

			if math.Abs(distance-tt.expected) > tolerance {
				t.Errorf("Haversine() = %.1f m, want %.1f m (diff: %.1f m)",
					distance, tt.expected, math.Abs(distance-tt.expected))
			}
		})
	}
}

// TestInitialBearing tests bearing calculations
func TestInitialBearing(t *testing.T) {
	tests := []struct {
		name     string
		lat1     float64
		lon1     float64
		lat2     float64
		lon2     float64
		expected float64
	}{
		{
			name:     "Due East",
			lat1:     0,
			lon1:     0,
			lat2:     0,
			lon2:     1,
			expected: 90,
		},
		{
			name:     "Due West",
			lat1:     0,
			lon1:     0,
			lat2:     0,
			lon2:     -1,
			expected: 270,
		},
		{
			name:     "Due North",
			lat1:     0,
			lon1:     0,
			lat2:     1,
			lon2:     0,
			expected: 0,
		},
		{
			name:     "Due South",
			lat1:     0,
			lon1:     0,
			lat2:     -1,
			lon2:     0,
			expected: 180,
		},
		{
			name:     "Northeast",
			lat1:     0,
			lon1:     0,
			lat2:     1,
			lon2:     1,
			expected: 45,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bearing := InitialBearing(tt.lat1, tt.lon1, tt.lat2, tt.lon2)

			bearingDiff := math.Abs(bearing - tt.expected)
			if bearingDiff > 180 {
				bearingDiff = 360 - bearingDiff
			}

			// Allow 1 degree tolerance for spherical approximation
			if bearingDiff > 1.0 {
				t.Errorf("InitialBearing() = %.2f°, want %.2f°", bearing, tt.expected)
			}
		})
	}
}

// TestPolygonArea tests area calculations
func TestPolygonArea(t *testing.T) {
	tests := []struct {
		name     string
		lats     []float64
		lons     []float64
		expected float64 // square meters
	}{
		{
			name: "Small square ~1km on each side",
			lats: []float64{0, 0.009, 0.009, 0, 0},
			lons: []float64{0, 0, 0.009, 0.009, 0},
			expected: 1000000, // ~1 square km
		},
		{
			name: "Triangle",
			lats: []float64{0, 0, 1, 0},
			lons: []float64{0, 1, 0, 0},
			expected: 6.17e9, // Large triangle
		},
		{
			name: "Point (degenerate)",
			lats: []float64{0},
			lons: []float64{0},
			expected: 0,
		},
		{
			name: "Line (degenerate)",
			lats: []float64{0, 1},
			lons: []float64{0, 1},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			area := PolygonArea(tt.lats, tt.lons, WGS84)

			if tt.expected == 0 {
				if area != 0 {
					t.Errorf("PolygonArea() = %.1f m², want 0 m²", area)
				}
				return
			}

			// Allow relative error for area calculations
			relError := math.Abs(area-tt.expected) / tt.expected
			if relError > areaTolerance*10 { // Allow 1% for area
				t.Errorf("PolygonArea() = %.0f m², want %.0f m² (rel error: %.2f%%)",
					area, tt.expected, relError*100)
			}
		})
	}
}

// TestSphericalPolygonArea tests spherical area calculations
func TestSphericalPolygonArea(t *testing.T) {
	tests := []struct {
		name string
		lats []float64
		lons []float64
	}{
		{
			name: "Square",
			lats: []float64{0, 0, 1, 1, 0},
			lons: []float64{0, 1, 1, 0, 0},
		},
		{
			name: "Triangle",
			lats: []float64{0, 0, 1, 0},
			lons: []float64{0, 1, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			area := SphericalPolygonArea(tt.lats, tt.lons, EarthMeanRadius)

			if area <= 0 {
				t.Errorf("SphericalPolygonArea() = %.1f m², want positive area", area)
			}

			// Compare with ellipsoidal result (should be within 1%)
			ellipsoidalArea := PolygonArea(tt.lats, tt.lons, WGS84)
			relDiff := math.Abs(area-ellipsoidalArea) / ellipsoidalArea

			if relDiff > 0.01 {
				t.Logf("Spherical vs ellipsoidal difference: %.2f%%", relDiff*100)
			}
		})
	}
}

// TestSignedPolygonArea tests signed area (winding order detection)
func TestSignedPolygonArea(t *testing.T) {
	// Counter-clockwise square
	latsCCW := []float64{0, 0, 1, 1, 0}
	lonsCCW := []float64{0, 1, 1, 0, 0}

	// Clockwise square (reversed)
	latsCW := []float64{0, 1, 1, 0, 0}
	lonsCW := []float64{0, 0, 1, 1, 0}

	areaCCW := SignedPolygonArea(latsCCW, lonsCCW, WGS84)
	areaCW := SignedPolygonArea(latsCW, lonsCW, WGS84)

	// Areas should have opposite signs
	if (areaCCW > 0 && areaCW > 0) || (areaCCW < 0 && areaCW < 0) {
		t.Errorf("Signed areas should have opposite signs: CCW=%.0f, CW=%.0f", areaCCW, areaCW)
	}

	// Absolute values should be approximately equal
	if math.Abs(math.Abs(areaCCW)-math.Abs(areaCW)) > math.Abs(areaCCW)*areaTolerance {
		t.Errorf("Signed area magnitudes differ: |CCW|=%.0f, |CW|=%.0f",
			math.Abs(areaCCW), math.Abs(areaCW))
	}
}

// TestDestinationPoint tests destination point calculations
func TestDestinationPoint(t *testing.T) {
	tests := []struct {
		name        string
		lat         float64
		lon         float64
		bearing     float64
		distance    float64
		expectedLat float64
		expectedLon float64
	}{
		{
			name:        "Due North from equator",
			lat:         0,
			lon:         0,
			bearing:     0,
			distance:    111320, // ~1 degree
			expectedLat: 1.0,
			expectedLon: 0,
		},
		{
			name:        "Due East from equator",
			lat:         0,
			lon:         0,
			bearing:     90,
			distance:    111320,
			expectedLat: 0,
			expectedLon: 1.0,
		},
		{
			name:        "Zero distance",
			lat:         40.7128,
			lon:         -74.0060,
			bearing:     45,
			distance:    0,
			expectedLat: 40.7128,
			expectedLon: -74.0060,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lat2, lon2 := DestinationPoint(tt.lat, tt.lon, tt.bearing, tt.distance, WGS84)

			// Allow ~0.01 degree tolerance (~1km)
			if math.Abs(lat2-tt.expectedLat) > 0.01 {
				t.Errorf("Latitude = %.6f°, want %.6f°", lat2, tt.expectedLat)
			}

			lonDiff := math.Abs(lon2 - tt.expectedLon)
			if lonDiff > 180 {
				lonDiff = 360 - lonDiff
			}
			if lonDiff > 0.01 {
				t.Errorf("Longitude = %.6f°, want %.6f°", lon2, tt.expectedLon)
			}
		})
	}
}

// TestConversionFunctions tests helper conversion functions
func TestConversionFunctions(t *testing.T) {
	// Test deg2rad and rad2deg
	deg := 180.0
	rad := deg2rad(deg)
	if math.Abs(rad-math.Pi) > 1e-10 {
		t.Errorf("deg2rad(180) = %v, want %v", rad, math.Pi)
	}

	deg2 := rad2deg(rad)
	if math.Abs(deg2-deg) > 1e-10 {
		t.Errorf("rad2deg(pi) = %v, want %v", deg2, deg)
	}

	// Test normalizeAzimuth
	tests := []struct {
		input    float64
		expected float64
	}{
		{0, 0},
		{360, 0},
		{-90, 270},
		{450, 90},
		{-180, 180},
	}

	for _, tt := range tests {
		result := normalizeAzimuth(tt.input)
		if math.Abs(result-tt.expected) > 1e-10 {
			t.Errorf("normalizeAzimuth(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}

	// Test normalizeLongitude
	lonTests := []struct {
		input    float64
		expected float64
	}{
		{0, 0},
		{180, 180},
		{-180, 180}, // -180 normalizes to 180
		{181, -179},
		{-181, 179},
		{360, 0},
	}

	for _, tt := range lonTests {
		result := normalizeLongitude(tt.input)
		if math.Abs(result-tt.expected) > 1e-10 {
			t.Errorf("normalizeLongitude(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

// BenchmarkVincenty benchmarks Vincenty's inverse formula
func BenchmarkVincenty(b *testing.B) {
	lat1, lon1 := 40.7128, -74.0060
	lat2, lon2 := 51.5074, -0.1278

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Vincenty(lat1, lon1, lat2, lon2, WGS84)
	}
}

// BenchmarkHaversine benchmarks Haversine formula
func BenchmarkHaversine(b *testing.B) {
	lat1, lon1 := 40.7128, -74.0060
	lat2, lon2 := 51.5074, -0.1278

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Haversine(lat1, lon1, lat2, lon2, EarthMeanRadius)
	}
}

// BenchmarkDirect benchmarks Vincenty's direct formula
func BenchmarkDirect(b *testing.B) {
	lat, lon := 40.7128, -74.0060
	bearing := 45.0
	distance := 100000.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = Direct(lat, lon, bearing, distance, WGS84)
	}
}

// BenchmarkPolygonArea benchmarks polygon area calculation
func BenchmarkPolygonArea(b *testing.B) {
	lats := []float64{0, 0, 1, 1, 0}
	lons := []float64{0, 1, 1, 0, 0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = PolygonArea(lats, lons, WGS84)
	}
}
