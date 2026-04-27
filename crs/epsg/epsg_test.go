package epsg

import (
	"fmt"
	"testing"

	"github.com/robert-malhotra/go-topology-suite/crs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			require.NoError(t, err)
			require.NotNil(t, crs)
			expectedCode := fmt.Sprintf("EPSG:%d", tt.code)
			assert.Equal(t, expectedCode, crs.Code())
			assert.Equal(t, tt.name, crs.Name())
		})
	}
}

// TestLookupNotFound tests that looking up an unregistered code returns an error.
func TestLookupNotFound(t *testing.T) {
	_, err := Lookup(99999)
	require.Error(t, err)
}

// TestMustLookup tests the MustLookup function.
func TestMustLookup(t *testing.T) {
	// Should not panic for valid code
	crs := MustLookup(4326)
	require.NotNil(t, crs)
	assert.Equal(t, "EPSG:4326", crs.Code())
}

// TestMustLookupPanics tests that MustLookup panics for invalid codes.
func TestMustLookupPanics(t *testing.T) {
	assert.Panics(t, func() { MustLookup(99999) })
}

// TestWGS84IsGeographic tests that WGS84 is identified as geographic.
func TestWGS84IsGeographic(t *testing.T) {
	assert.True(t, WGS84.IsGeographic())
	assert.Equal(t, crs.Geographic, WGS84.Type())
}

// TestWebMercatorIsProjected tests that Web Mercator is identified as projected.
func TestWebMercatorIsProjected(t *testing.T) {
	assert.False(t, WebMercator.IsGeographic())
	assert.Equal(t, crs.Projected, WebMercator.Type())
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
			crs, err := UTMZone(tt.zone, tt.north)
			require.NoError(t, err)
			assert.Equal(t, tt.wantCode, crs.Code())
			assert.Equal(t, tt.wantName, crs.Name())
			assert.False(t, crs.IsGeographic())
		})
	}
}

// TestUTMZoneErrors tests that UTMZone returns errors for invalid zones.
func TestUTMZoneErrors(t *testing.T) {
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
			_, err := UTMZone(tt.zone, tt.north)
			require.Error(t, err)
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
	require.NoError(t, err)

	// Register the custom CRS
	err = Register(customCRS)
	require.NoError(t, err)

	// Verify it can be looked up
	retrieved, err := Lookup(customCode)
	require.NoError(t, err)
	assert.Equal(t, "EPSG:2154", retrieved.Code())
	assert.Equal(t, customCRS.Name(), retrieved.Name())

	// Clean up
	_ = Unregister(customCode)
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
	require.NoError(t, err)

	err = Register(customCRS)
	require.Error(t, err)
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
	require.NoError(t, err)

	// Register and then unregister
	_ = Register(customCRS)
	err = Unregister(customCode)
	require.NoError(t, err)

	// Verify it's no longer in the registry
	_, err = Lookup(customCode)
	require.Error(t, err)
}

// TestUnregisterNotFound tests unregistering a non-existent code.
func TestUnregisterNotFound(t *testing.T) {
	err := Unregister(99999)
	require.Error(t, err)
}

// TestCodes tests the Codes function.
func TestCodes(t *testing.T) {
	codes := Codes()
	assert.NotEmpty(t, codes)

	// Verify that codes are sorted
	for i := 1; i < len(codes); i++ {
		assert.Less(t, codes[i-1], codes[i], "Codes() not sorted at index %d", i)
	}

	// Verify some expected codes are present
	assert.Contains(t, codes, 4326)
	assert.Contains(t, codes, 4269)
	assert.Contains(t, codes, 3857)
}

// TestCount tests the Count function.
func TestCount(t *testing.T) {
	count := Count()
	assert.GreaterOrEqual(t, count, 8)

	// Add a custom CRS and verify count increases
	customCode := 9998
	customCRS, err := crs.NewGeographicCRS(
		"EPSG:9998",
		"Test",
		crs.WGS84Datum,
		crs.EllipsoidalCS2D,
		nil,
	)
	require.NoError(t, err)

	_ = Register(customCRS)
	newCount := Count()
	assert.Equal(t, count+1, newCount)

	// Clean up
	_ = Unregister(customCode)
}

// TestIsRegistered tests the IsRegistered function.
func TestIsRegistered(t *testing.T) {
	assert.True(t, IsRegistered(4326))
	assert.False(t, IsRegistered(99999))
}

// TestDatumProperties tests that datum properties are correct for geographic CRS.
func TestDatumProperties(t *testing.T) {
	tests := []struct {
		name          string
		crs           crs.CRS
		wantDatumName string
		wantEllipsoid crs.Ellipsoid
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
			assert.Equal(t, tt.wantDatumName, datum.Name())
			assert.Equal(t, tt.wantEllipsoid, datum.Ellipsoid())
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
			assert.Equal(t, tt.wantDimension, cs.Dimension())
			axis0, err := cs.Axis(0)
			require.NoError(t, err)
			assert.Equal(t, tt.wantUnit, axis0.Unit)
		})
	}
}

// TestAreaOfUse tests that area of use is defined for CRS.
func TestAreaOfUse(t *testing.T) {
	tests := []struct {
		name       string
		crs        crs.CRS
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
			assert.GreaterOrEqual(t, minLon, -180.0)
			assert.LessOrEqual(t, minLon, 180.0)
			assert.GreaterOrEqual(t, maxLon, -180.0)
			assert.LessOrEqual(t, maxLon, 180.0)
			assert.GreaterOrEqual(t, minLat, -90.0)
			assert.LessOrEqual(t, minLat, 90.0)
			assert.GreaterOrEqual(t, maxLat, -90.0)
			assert.LessOrEqual(t, maxLat, 90.0)

			if tt.wantGlobal {
				assert.Equal(t, -180.0, minLon)
				assert.Equal(t, 180.0, maxLon)
				assert.Equal(t, -90.0, minLat)
				assert.Equal(t, 90.0, maxLat)
			}
		})
	}
}

// TestWKT tests that WKT output is generated for CRS.
func TestWKT(t *testing.T) {
	tests := []struct {
		name         string
		crs          crs.CRS
		wantContains string
	}{
		{"WGS84", WGS84, "GEOGCS"},
		{"WebMercator", WebMercator, "PROJCS"},
		{"UTM10N", UTM10N, "PROJCS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wkt := tt.crs.WKT()
			assert.NotEmpty(t, wkt)
			assert.GreaterOrEqual(t, len(wkt), 50)
			assert.Contains(t, wkt, tt.wantContains)
		})
	}
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
		_, _ = UTMZone(10, true)
	}
}
