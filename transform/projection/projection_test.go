package projection

import (
	"math"
	"testing"
)

const epsilon = 1e-6 // More relaxed tolerance for projection calculations

func TestWebMercatorOrigin(t *testing.T) {
	wm := WebMercator()

	// Test origin (0, 0) should map to (0, 0)
	x, y, err := wm.Forward(0, 0)
	if err != nil {
		t.Fatalf("Forward() error = %v", err)
	}

	if math.Abs(x) > epsilon || math.Abs(y) > epsilon {
		t.Errorf("Forward(0, 0) = (%f, %f), want (0, 0)", x, y)
	}
}

func TestWebMercatorSanFrancisco(t *testing.T) {
	wm := WebMercator()

	// San Francisco coordinates
	lon, lat := -122.4194, 37.7749

	x, y, err := wm.Forward(lon, lat)
	if err != nil {
		t.Fatalf("Forward() error = %v", err)
	}

	// Expected values (approximate)
	expectedX := -13627665.0
	expectedY := 4548431.0

	// Allow 1km tolerance for these approximate values
	tolerance := 1000.0

	if math.Abs(x-expectedX) > tolerance || math.Abs(y-expectedY) > tolerance {
		t.Errorf("Forward(%f, %f) = (%f, %f), want approximately (%f, %f)",
			lon, lat, x, y, expectedX, expectedY)
	}

	// Test inverse
	lonInv, latInv, err := wm.Inverse(x, y)
	if err != nil {
		t.Fatalf("Inverse() error = %v", err)
	}

	// Should get back original coordinates (within small tolerance)
	if math.Abs(lonInv-lon) > 1e-6 || math.Abs(latInv-lat) > 1e-6 {
		t.Errorf("Round trip failed: got (%f, %f), want (%f, %f)",
			lonInv, latInv, lon, lat)
	}
}

func TestWebMercatorRoundTrip(t *testing.T) {
	wm := WebMercator()

	tests := []struct {
		name string
		lon  float64
		lat  float64
	}{
		{"origin", 0, 0},
		{"prime meridian", 0, 51.5},     // London latitude
		{"equator", -122.4, 0},           // San Francisco longitude
		{"New York", -74.006, 40.7128},
		{"Tokyo", 139.6917, 35.6895},
		{"Sydney", 151.2093, -33.8688},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Forward transformation
			x, y, err := wm.Forward(tt.lon, tt.lat)
			if err != nil {
				t.Fatalf("Forward() error = %v", err)
			}

			// Inverse transformation
			lonInv, latInv, err := wm.Inverse(x, y)
			if err != nil {
				t.Fatalf("Inverse() error = %v", err)
			}

			// Check round trip accuracy
			if math.Abs(lonInv-tt.lon) > 1e-9 || math.Abs(latInv-tt.lat) > 1e-9 {
				t.Errorf("Round trip failed: got (%f, %f), want (%f, %f)",
					lonInv, latInv, tt.lon, tt.lat)
			}
		})
	}
}

func TestWebMercatorPoleErrors(t *testing.T) {
	wm := WebMercator()

	// North pole should error
	_, _, err := wm.Forward(0, 90)
	if err == nil {
		t.Error("Forward(0, 90) should return error for north pole")
	}

	// South pole should error
	_, _, err = wm.Forward(0, -90)
	if err == nil {
		t.Error("Forward(0, -90) should return error for south pole")
	}
}

func TestMercatorEllipsoidal(t *testing.T) {
	// Create Mercator with WGS84 ellipsoid (not spherical)
	merc := NewMercator(WGS84(), 0, 0, 0)

	lon, lat := -122.4194, 37.7749

	// Forward transformation
	x, y, err := merc.Forward(lon, lat)
	if err != nil {
		t.Fatalf("Forward() error = %v", err)
	}

	// Inverse transformation
	lonInv, latInv, err := merc.Inverse(x, y)
	if err != nil {
		t.Fatalf("Inverse() error = %v", err)
	}

	// Check round trip accuracy
	if math.Abs(lonInv-lon) > 1e-6 || math.Abs(latInv-lat) > 1e-6 {
		t.Errorf("Ellipsoidal round trip failed: got (%f, %f), want (%f, %f)",
			lonInv, latInv, lon, lat)
	}
}

func TestUTMZoneCalculation(t *testing.T) {
	tests := []struct {
		zone           int
		north          bool
		expectedCM     float64
		expectedFE     float64
		expectedFN     float64
	}{
		{1, true, -177, 500000, 0},
		{30, true, -3, 500000, 0},      // Prime meridian zone
		{31, true, 3, 500000, 0},
		{60, true, 177, 500000, 0},
		{30, false, -3, 500000, 10000000}, // Southern hemisphere
	}

	for _, tt := range tests {
		name := "UTM Zone " + string(rune('0'+tt.zone))
		if tt.north {
			name += "N"
		} else {
			name += "S"
		}

		t.Run(name, func(t *testing.T) {
			utm := UTM(tt.zone, tt.north, nil)

			if math.Abs(utm.CentralMeridian-tt.expectedCM) > epsilon {
				t.Errorf("Central meridian = %f, want %f",
					utm.CentralMeridian, tt.expectedCM)
			}

			if math.Abs(utm.FalseEasting-tt.expectedFE) > epsilon {
				t.Errorf("False easting = %f, want %f",
					utm.FalseEasting, tt.expectedFE)
			}

			if math.Abs(utm.FalseNorthing-tt.expectedFN) > epsilon {
				t.Errorf("False northing = %f, want %f",
					utm.FalseNorthing, tt.expectedFN)
			}

			if math.Abs(utm.ScaleFactor-0.9996) > epsilon {
				t.Errorf("Scale factor = %f, want 0.9996", utm.ScaleFactor)
			}
		})
	}
}

func TestUTMRoundTrip(t *testing.T) {
	// Test UTM Zone 10N (covers San Francisco area)
	utm := UTM(10, true, WGS84())

	tests := []struct {
		name string
		lon  float64
		lat  float64
	}{
		{"center of zone", -123, 37.7749},  // Near San Francisco
		{"equator", -123, 0},
		{"north", -123, 60},
		{"edge west", -126, 40},
		{"edge east", -120, 40},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Forward transformation
			x, y, err := utm.Forward(tt.lon, tt.lat)
			if err != nil {
				t.Fatalf("Forward() error = %v", err)
			}

			// Inverse transformation
			lonInv, latInv, err := utm.Inverse(x, y)
			if err != nil {
				t.Fatalf("Inverse() error = %v", err)
			}

			// Check round trip accuracy (more relaxed tolerance for UTM)
			tolLon := 1e-7
			tolLat := 1e-7

			if math.Abs(lonInv-tt.lon) > tolLon || math.Abs(latInv-tt.lat) > tolLat {
				t.Errorf("Round trip failed: got (%f, %f), want (%f, %f)",
					lonInv, latInv, tt.lon, tt.lat)
			}
		})
	}
}

func TestTransverseMercatorOrigin(t *testing.T) {
	// Create a simple TM projection at the origin
	tm := NewTransverseMercator(WGS84(), 0, 0, 1.0, 0, 0)

	// Origin should map to origin
	x, y, err := tm.Forward(0, 0)
	if err != nil {
		t.Fatalf("Forward() error = %v", err)
	}

	if math.Abs(x) > epsilon || math.Abs(y) > epsilon {
		t.Errorf("Forward(0, 0) = (%f, %f), want (0, 0)", x, y)
	}
}

func TestTransverseMercatorRoundTrip(t *testing.T) {
	tm := NewTransverseMercator(WGS84(), 0, 0, 1.0, 0, 0)

	tests := []struct {
		name string
		lon  float64
		lat  float64
	}{
		{"origin", 0, 0},
		{"prime meridian north", 0, 45},
		{"prime meridian south", 0, -45},
		{"east of cm", 5, 45},
		{"west of cm", -5, 45},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Forward transformation
			x, y, err := tm.Forward(tt.lon, tt.lat)
			if err != nil {
				t.Fatalf("Forward() error = %v", err)
			}

			// Inverse transformation
			lonInv, latInv, err := tm.Inverse(x, y)
			if err != nil {
				t.Fatalf("Inverse() error = %v", err)
			}

			// Check round trip accuracy (more relaxed tolerance)
			tolLon := 1e-5
			tolLat := 1e-5

			if math.Abs(lonInv-tt.lon) > tolLon || math.Abs(latInv-tt.lat) > tolLat {
				t.Errorf("Round trip failed: got (%.10f, %.10f), want (%.10f, %.10f)",
					lonInv, latInv, tt.lon, tt.lat)
			}
		})
	}
}

func TestEllipsoids(t *testing.T) {
	tests := []struct {
		name      string
		ellipsoid *Ellipsoid
		expectedA float64
		spherical bool
	}{
		{"WGS84", WGS84(), 6378137.0, false},
		{"GRS80", GRS80(), 6378137.0, false},
		{"Clarke1866", Clarke1866(), 6378206.4, false},
		{"Sphere default", Sphere(0), 6378137.0, true},
		{"Sphere custom", Sphere(6371000), 6371000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if math.Abs(tt.ellipsoid.A-tt.expectedA) > epsilon {
				t.Errorf("Semi-major axis = %f, want %f",
					tt.ellipsoid.A, tt.expectedA)
			}

			if tt.ellipsoid.IsSpherical() != tt.spherical {
				t.Errorf("IsSpherical() = %v, want %v",
					tt.ellipsoid.IsSpherical(), tt.spherical)
			}
		})
	}
}

func TestLongitudeNormalization(t *testing.T) {
	wm := WebMercator()

	// Test that longitude gets normalized to [-180, 180]
	tests := []struct {
		name        string
		lon         float64
		expectedLon float64
	}{
		{"normal", 45, 45},
		{"just over 180", 181, -179},
		{"just under -180", -181, 179},
		{"360", 360, 0},
		{"-360", -360, 0},
		{"large positive", 540, -180},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Project and inverse project
			x, y, err := wm.Forward(tt.lon, 0)
			if err != nil {
				t.Fatalf("Forward() error = %v", err)
			}

			lonInv, _, err := wm.Inverse(x, y)
			if err != nil {
				t.Fatalf("Inverse() error = %v", err)
			}

			// The normalized longitude should be in [-180, 180]
			if lonInv < -180 || lonInv > 180 {
				t.Errorf("Inverse longitude %f is not in [-180, 180]", lonInv)
			}
		})
	}
}

func BenchmarkWebMercatorForward(b *testing.B) {
	wm := WebMercator()
	lon, lat := -122.4194, 37.7749

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = wm.Forward(lon, lat)
	}
}

func BenchmarkWebMercatorInverse(b *testing.B) {
	wm := WebMercator()
	x, y := -13627665.0, 4548431.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = wm.Inverse(x, y)
	}
}

func BenchmarkUTMForward(b *testing.B) {
	utm := UTM(10, true, WGS84())
	lon, lat := -122.4194, 37.7749

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = utm.Forward(lon, lat)
	}
}

func BenchmarkUTMInverse(b *testing.B) {
	utm := UTM(10, true, WGS84())
	x, y := 550000.0, 4180000.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = utm.Inverse(x, y)
	}
}
