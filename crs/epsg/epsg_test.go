package epsg

import (
	"fmt"
	"testing"

	"github.com/robert-malhotra/go-topology-suite/crs"
)

// TestLookupCommonCRS tests that common EPSG codes can be looked up.
func TestLookupCommonCRS(t *testing.T) {
	tests := []struct {
		code int
		name string
	}{
		{4326, "WGS 84"},
		{4269, "NAD83"},
		{4267, "NAD27"},
		{4258, "ETRS89"},
		{3857, "WGS 84 / Pseudo-Mercator"},
		{32610, "WGS 84 / UTM zone 10N"},
		{32617, "WGS 84 / UTM zone 17N"},
		{32632, "WGS 84 / UTM zone 32N"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			crs, err := Lookup(tt.code)
			if err != nil {
				t.Fatalf("Lookup(%d) error: %v", tt.code, err)
			}
			if crs == nil {
				t.Fatalf("Lookup(%d) returned nil", tt.code)
			}
			expectedCode := fmt.Sprintf("EPSG:%d", tt.code)
			if crs.Code() != expectedCode {
				t.Errorf("Code() = %q, want %q", crs.Code(), expectedCode)
			}
			if crs.Name() != tt.name {
				t.Errorf("Name() = %q, want %q", crs.Name(), tt.name)
			}
		})
	}
}

// TestLookupNotFound tests that looking up an unregistered code returns an error.
func TestLookupNotFound(t *testing.T) {
	_, err := Lookup(99999)
	if err == nil {
		t.Error("Lookup(99999) expected error, got nil")
	}
}

// TestMustLookup tests the MustLookup function.
func TestMustLookup(t *testing.T) {
	// Should not panic for valid code
	crs := MustLookup(4326)
	if crs == nil {
		t.Fatal("MustLookup(4326) returned nil")
	}
	if crs.Code() != "EPSG:4326" {
		t.Errorf("Code() = %q, want %q", crs.Code(), "EPSG:4326")
	}
}

// TestMustLookupPanics tests that MustLookup panics for invalid codes.
func TestMustLookupPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustLookup(99999) did not panic")
		}
	}()
	MustLookup(99999)
}

// TestWGS84IsGeographic tests that WGS84 is identified as geographic.
func TestWGS84IsGeographic(t *testing.T) {
	if !WGS84.IsGeographic() {
		t.Error("WGS84.IsGeographic() = false, want true")
	}
	if WGS84.Type() != crs.Geographic {
		t.Errorf("WGS84.Type() = %v, want Geographic", WGS84.Type())
	}
}

// TestWebMercatorIsProjected tests that Web Mercator is identified as projected.
func TestWebMercatorIsProjected(t *testing.T) {
	if WebMercator.IsGeographic() {
		t.Error("WebMercator.IsGeographic() = true, want false")
	}
	if WebMercator.Type() != crs.Projected {
		t.Errorf("WebMercator.Type() = %v, want Projected", WebMercator.Type())
	}
}

// TestUTMZoneGeneration tests the UTMZone function.
func TestUTMZoneGeneration(t *testing.T) {
	tests := []struct {
		zone     int
		north    bool
		wantCode string
		wantName string
	}{
		{1, true, "EPSG:32601", "WGS 84 / UTM zone 1N"},
		{60, true, "EPSG:32660", "WGS 84 / UTM zone 60N"},
		{1, false, "EPSG:32701", "WGS 84 / UTM zone 1S"},
		{60, false, "EPSG:32760", "WGS 84 / UTM zone 60S"},
		{10, true, "EPSG:32610", "WGS 84 / UTM zone 10N"},
		{32, true, "EPSG:32632", "WGS 84 / UTM zone 32N"},
	}

	for _, tt := range tests {
		t.Run(tt.wantName, func(t *testing.T) {
			crs := UTMZone(tt.zone, tt.north)
			if crs.Code() != tt.wantCode {
				t.Errorf("UTMZone(%d, %v).Code() = %q, want %q",
					tt.zone, tt.north, crs.Code(), tt.wantCode)
			}
			if crs.Name() != tt.wantName {
				t.Errorf("UTMZone(%d, %v).Name() = %q, want %q",
					tt.zone, tt.north, crs.Name(), tt.wantName)
			}
			if crs.IsGeographic() {
				t.Errorf("UTMZone(%d, %v).IsGeographic() = true, want false",
					tt.zone, tt.north)
			}
		})
	}
}

// TestUTMZonePanics tests that UTMZone panics for invalid zones.
func TestUTMZonePanics(t *testing.T) {
	tests := []struct {
		zone  int
		north bool
	}{
		{0, true},
		{61, true},
		{-1, false},
		{100, false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("UTMZone(%d, %v) did not panic", tt.zone, tt.north)
				}
			}()
			UTMZone(tt.zone, tt.north)
		})
	}
}

// TestRegisterCustomCRS tests registering a custom CRS.
func TestRegisterCustomCRS(t *testing.T) {
	customCode := 2154
	customCRS, err := crs.NewProjectedCRS(
		"EPSG:2154",
		"RGF93 / Lambert-93",
		crs.WGS84,
		crs.CartesianCS2D,
		"Lambert Conformal Conic",
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create custom CRS: %v", err)
	}

	// Register the custom CRS
	err = Register(customCRS)
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	// Verify it can be looked up
	retrieved, err := Lookup(customCode)
	if err != nil {
		t.Fatalf("Lookup(%d) error: %v", customCode, err)
	}
	if retrieved.Code() != "EPSG:2154" {
		t.Errorf("Code() = %q, want %q", retrieved.Code(), "EPSG:2154")
	}
	if retrieved.Name() != customCRS.Name() {
		t.Errorf("Name() = %q, want %q", retrieved.Name(), customCRS.Name())
	}

	// Clean up
	Unregister(customCode)
}

// TestRegisterInvalidCode tests that registering a CRS with invalid code returns an error.
func TestRegisterInvalidCode(t *testing.T) {
	customCRS, err := crs.NewGeographicCRS(
		"INVALID:0",
		"Invalid CRS",
		crs.WGS84Datum,
		crs.EllipsoidalCS2D,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create custom CRS: %v", err)
	}

	err = Register(customCRS)
	if err == nil {
		t.Error("Register(CRS with invalid code) expected error, got nil")
	}
}

// TestUnregister tests unregistering a CRS.
func TestUnregister(t *testing.T) {
	customCode := 9999
	customCRS, err := crs.NewGeographicCRS(
		"EPSG:9999",
		"Test CRS",
		crs.WGS84Datum,
		crs.EllipsoidalCS2D,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create custom CRS: %v", err)
	}

	// Register and then unregister
	Register(customCRS)
	err = Unregister(customCode)
	if err != nil {
		t.Fatalf("Unregister(%d) error: %v", customCode, err)
	}

	// Verify it's no longer in the registry
	_, err = Lookup(customCode)
	if err == nil {
		t.Error("Lookup() after Unregister() should return error")
	}
}

// TestUnregisterNotFound tests unregistering a non-existent code.
func TestUnregisterNotFound(t *testing.T) {
	err := Unregister(99999)
	if err == nil {
		t.Error("Unregister(99999) expected error, got nil")
	}
}

// TestCodes tests the Codes function.
func TestCodes(t *testing.T) {
	codes := Codes()
	if len(codes) == 0 {
		t.Error("Codes() returned empty slice")
	}

	// Verify that codes are sorted
	for i := 1; i < len(codes); i++ {
		if codes[i] <= codes[i-1] {
			t.Errorf("Codes() not sorted: %d <= %d at index %d",
				codes[i], codes[i-1], i)
		}
	}

	// Verify some expected codes are present
	expectedCodes := []int{4326, 4269, 3857}
	for _, expected := range expectedCodes {
		found := false
		for _, code := range codes {
			if code == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Codes() missing expected code %d", expected)
		}
	}
}

// TestCount tests the Count function.
func TestCount(t *testing.T) {
	count := Count()
	if count < 8 {
		t.Errorf("Count() = %d, want at least 8", count)
	}

	// Add a custom CRS and verify count increases
	customCode := 9998
	customCRS, err := crs.NewGeographicCRS(
		"EPSG:9998",
		"Test",
		crs.WGS84Datum,
		crs.EllipsoidalCS2D,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create custom CRS: %v", err)
	}

	Register(customCRS)
	newCount := Count()
	if newCount != count+1 {
		t.Errorf("Count() after Register() = %d, want %d", newCount, count+1)
	}

	// Clean up
	Unregister(customCode)
}

// TestIsRegistered tests the IsRegistered function.
func TestIsRegistered(t *testing.T) {
	if !IsRegistered(4326) {
		t.Error("IsRegistered(4326) = false, want true")
	}
	if IsRegistered(99999) {
		t.Error("IsRegistered(99999) = true, want false")
	}
}

// TestDatumProperties tests that datum properties are correct for geographic CRS.
func TestDatumProperties(t *testing.T) {
	tests := []struct {
		name           string
		crs            crs.CRS
		wantDatumName  string
		wantEllipsoid  crs.Ellipsoid
	}{
		{
			name:          "WGS84",
			crs:           WGS84,
			wantDatumName: "WGS 84",
			wantEllipsoid: crs.WGS84Ellipsoid,
		},
		{
			name:          "NAD83",
			crs:           NAD83,
			wantDatumName: "NAD83",
			wantEllipsoid: crs.GRS80Ellipsoid,
		},
		{
			name:          "NAD27",
			crs:           NAD27,
			wantDatumName: "NAD27",
			wantEllipsoid: crs.Clarke1866Ellipsoid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			datum := tt.crs.Datum()
			if datum.Name() != tt.wantDatumName {
				t.Errorf("Datum().Name() = %q, want %q", datum.Name(), tt.wantDatumName)
			}
			if datum.Ellipsoid() != tt.wantEllipsoid {
				t.Errorf("Datum().Ellipsoid() != expected ellipsoid")
			}
		})
	}
}

// TestCoordinateSystem tests that coordinate systems are correctly set.
func TestCoordinateSystem(t *testing.T) {
	tests := []struct {
		name          string
		crs           crs.CRS
		wantDimension int
		wantUnit      crs.Unit
	}{
		{"WGS84", WGS84, 2, crs.Degree},
		{"NAD83", NAD83, 2, crs.Degree},
		{"WebMercator", WebMercator, 2, crs.Metre},
		{"UTM10N", UTM10N, 2, crs.Metre},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := tt.crs.CoordinateSystem()
			if cs.Dimension() != tt.wantDimension {
				t.Errorf("CoordinateSystem().Dimension() = %d, want %d",
					cs.Dimension(), tt.wantDimension)
			}
			axis0 := cs.Axis(0)
			if axis0.Unit != tt.wantUnit {
				t.Errorf("CoordinateSystem().Axis(0).Unit = %v, want %v",
					axis0.Unit, tt.wantUnit)
			}
		})
	}
}

// TestAreaOfUse tests that area of use is defined for CRS.
func TestAreaOfUse(t *testing.T) {
	tests := []struct {
		name      string
		crs       crs.CRS
		wantGlobal bool
	}{
		{"WGS84", WGS84, true},
		{"NAD83", NAD83, false},
		{"WebMercator", WebMercator, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minLon, minLat, maxLon, maxLat := tt.crs.AreaOfUse()

			// Check that values are reasonable
			if minLon < -180 || minLon > 180 {
				t.Errorf("AreaOfUse() minLon = %f, out of range [-180, 180]", minLon)
			}
			if maxLon < -180 || maxLon > 180 {
				t.Errorf("AreaOfUse() maxLon = %f, out of range [-180, 180]", maxLon)
			}
			if minLat < -90 || minLat > 90 {
				t.Errorf("AreaOfUse() minLat = %f, out of range [-90, 90]", minLat)
			}
			if maxLat < -90 || maxLat > 90 {
				t.Errorf("AreaOfUse() maxLat = %f, out of range [-90, 90]", maxLat)
			}

			if tt.wantGlobal {
				if minLon != -180 || maxLon != 180 || minLat != -90 || maxLat != 90 {
					t.Errorf("AreaOfUse() = (%f, %f, %f, %f), want global coverage (-180, -90, 180, 90)",
						minLon, minLat, maxLon, maxLat)
				}
			}
		})
	}
}

// TestWKT tests that WKT output is generated for CRS.
func TestWKT(t *testing.T) {
	tests := []struct {
		name        string
		crs         crs.CRS
		wantContains string
	}{
		{"WGS84", WGS84, "GEOGCS"},
		{"WebMercator", WebMercator, "PROJCS"},
		{"UTM10N", UTM10N, "PROJCS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wkt := tt.crs.WKT()
			if wkt == "" {
				t.Error("WKT() returned empty string")
			}
			if len(wkt) < 50 {
				t.Errorf("WKT() = %q, seems too short", wkt)
			}
			// Just check that it contains expected keyword
			if !contains(wkt, tt.wantContains) {
				t.Errorf("WKT() = %q, want to contain %q", wkt, tt.wantContains)
			}
		})
	}
}

// contains is a simple substring check helper.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// BenchmarkLookup benchmarks the Lookup function.
func BenchmarkLookup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Lookup(4326)
	}
}

// BenchmarkUTMZone benchmarks the UTMZone function.
func BenchmarkUTMZone(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = UTMZone(10, true)
	}
}
