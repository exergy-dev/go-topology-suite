package crs

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEllipsoid tests ellipsoid creation and calculations.
func TestEllipsoid(t *testing.T) {
	tests := []struct {
		name              string
		semiMajorAxis     float64
		inverseFlattening float64
		wantSemiMinor     float64
		wantEccentricity  float64
		wantError         bool
	}{
		{
			name:              "WGS84",
			semiMajorAxis:     6378137.0,
			inverseFlattening: 298.257223563,
			wantSemiMinor:     6356752.314245179,
			wantEccentricity:  0.081819190842622,
			wantError:         false,
		},
		{
			name:              "Sphere",
			semiMajorAxis:     6371000.0,
			inverseFlattening: 0.0,
			wantSemiMinor:     6371000.0,
			wantEccentricity:  0.0,
			wantError:         false,
		},
		{
			name:              "Invalid negative semi-major axis",
			semiMajorAxis:     -1000.0,
			inverseFlattening: 298.0,
			wantError:         true,
		},
		{
			name:              "Invalid negative inverse flattening",
			semiMajorAxis:     6378137.0,
			inverseFlattening: -10.0,
			wantError:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ellipsoid, err := NewEllipsoid(tt.name, tt.semiMajorAxis, tt.inverseFlattening)

			if tt.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tt.name, ellipsoid.Name())
			assert.Equal(t, tt.semiMajorAxis, ellipsoid.SemiMajorAxis())
			assert.Equal(t, tt.inverseFlattening, ellipsoid.InverseFlattening())
			assert.InDelta(t, tt.wantSemiMinor, ellipsoid.SemiMinorAxis(), 1e-6)
			assert.InDelta(t, tt.wantEccentricity, ellipsoid.Eccentricity(), 1e-12)

			// Verify eccentricity squared
			wantEccentricitySq := tt.wantEccentricity * tt.wantEccentricity
			assert.InDelta(t, wantEccentricitySq, ellipsoid.EccentricitySquared(), 1e-12)
		})
	}
}

// TestEllipsoidFromAF tests ellipsoid creation from flattening.
func TestEllipsoidFromAF(t *testing.T) {
	// WGS84: f = 1/298.257223563 ≈ 0.003352810664747
	flattening := 1.0 / 298.257223563
	ellipsoid, err := NewEllipsoidFromAF("WGS84", 6378137.0, flattening)
	require.NoError(t, err)

	assert.InDelta(t, 298.257223563, ellipsoid.InverseFlattening(), 1e-6)

	// Test invalid flattening
	_, err = NewEllipsoidFromAF("Invalid", 6378137.0, 1.5)
	require.Error(t, err)

	_, err = NewEllipsoidFromAF("Invalid", 6378137.0, -0.1)
	require.Error(t, err)
}

// TestCommonEllipsoids tests the predefined common ellipsoids.
func TestCommonEllipsoids(t *testing.T) {
	tests := []struct {
		ellipsoid         Ellipsoid
		name              string
		semiMajorAxis     float64
		inverseFlattening float64
	}{
		{WGS84Ellipsoid, "WGS 84", 6378137.0, 298.257223563},
		{GRS80Ellipsoid, "GRS 1980", 6378137.0, 298.257222101},
		{Clarke1866Ellipsoid, "Clarke 1866", 6378206.4, 294.978698214},
		{Airy1830Ellipsoid, "Airy 1830", 6377563.396, 299.3249646},
		{Bessel1841Ellipsoid, "Bessel 1841", 6377397.155, 299.1528128},
		{International1924Ellipsoid, "International 1924", 6378388.0, 297.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.name, tt.ellipsoid.Name())
			assert.Equal(t, tt.semiMajorAxis, tt.ellipsoid.SemiMajorAxis())
			assert.InDelta(t, tt.inverseFlattening, tt.ellipsoid.InverseFlattening(), 1e-6)

			// Verify semi-minor axis is less than semi-major axis
			assert.True(t, tt.ellipsoid.SemiMinorAxis() < tt.ellipsoid.SemiMajorAxis())

			// Verify eccentricity is in valid range [0, 1)
			e := tt.ellipsoid.Eccentricity()
			assert.True(t, e >= 0 && e < 1, "Eccentricity() = %v, want [0, 1)", e)
		})
	}
}

// TestDatum tests datum creation and properties.
func TestDatum(t *testing.T) {
	ellipsoid := WGS84Ellipsoid

	tests := []struct {
		name          string
		datumName     string
		primeMeridian float64
		toWGS84Params []float64
		wantError     bool
	}{
		{
			name:          "WGS84",
			datumName:     "WGS 84",
			primeMeridian: 0.0,
			toWGS84Params: nil,
			wantError:     false,
		},
		{
			name:          "With transformation",
			datumName:     "Custom",
			primeMeridian: 0.0,
			toWGS84Params: []float64{1.0, 2.0, 3.0, 0.1, 0.2, 0.3, 0.5},
			wantError:     false,
		},
		{
			name:          "Invalid transformation params",
			datumName:     "Invalid",
			primeMeridian: 0.0,
			toWGS84Params: []float64{1.0, 2.0}, // Wrong number
			wantError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			datum, err := NewDatum(tt.datumName, ellipsoid, tt.primeMeridian, tt.toWGS84Params)

			if tt.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tt.datumName, datum.Name())
			assert.Equal(t, ellipsoid, datum.Ellipsoid())
			assert.Equal(t, tt.primeMeridian, datum.PrimeMeridian())

			// Check transformation parameters
			dx, dy, dz, rx, ry, rz, ds := datum.ToWGS84Params()
			if tt.toWGS84Params != nil {
				assert.Equal(t, tt.toWGS84Params[0], dx)
				assert.Equal(t, tt.toWGS84Params[1], dy)
				assert.Equal(t, tt.toWGS84Params[2], dz)
				assert.Equal(t, tt.toWGS84Params[3], rx)
				assert.Equal(t, tt.toWGS84Params[4], ry)
				assert.Equal(t, tt.toWGS84Params[5], rz)
				assert.Equal(t, tt.toWGS84Params[6], ds)
			}
		})
	}
}

// TestCommonDatums tests the predefined common datums.
func TestCommonDatums(t *testing.T) {
	tests := []struct {
		datum         Datum
		name          string
		ellipsoidName string
	}{
		{WGS84Datum, "WGS 84", "WGS 84"},
		{NAD83Datum, "NAD83", "GRS 1980"},
		{NAD27Datum, "NAD27", "Clarke 1866"},
		{OSGB36Datum, "OSGB 1936", "Airy 1830"},
		{ED50Datum, "ED50", "International 1924"},
		{TokyoDatum, "Tokyo", "Bessel 1841"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.name, tt.datum.Name())
			assert.Equal(t, tt.ellipsoidName, tt.datum.Ellipsoid().Name())

			// All common datums use Greenwich prime meridian
			assert.Equal(t, 0.0, tt.datum.PrimeMeridian())
		})
	}
}

// TestGeographicCRS tests geographic CRS creation and properties.
func TestGeographicCRS(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		crsName   string
		datum     Datum
		cs        CoordinateSystem
		areaOfUse []float64
		wantError bool
	}{
		{
			name:      "WGS84",
			code:      "EPSG:4326",
			crsName:   "WGS 84",
			datum:     WGS84Datum,
			cs:        EllipsoidalCS2D,
			areaOfUse: nil, // Global
			wantError: false,
		},
		{
			name:      "With area of use",
			code:      "EPSG:4269",
			crsName:   "NAD83",
			datum:     NAD83Datum,
			cs:        EllipsoidalCS2D,
			areaOfUse: []float64{-180, 14.92, 180, 86.46},
			wantError: false,
		},
		{
			name:      "Invalid area of use",
			code:      "TEST",
			crsName:   "Test",
			datum:     WGS84Datum,
			cs:        EllipsoidalCS2D,
			areaOfUse: []float64{-180, 14.92}, // Too few elements
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			crs, err := NewGeographicCRS(tt.code, tt.crsName, tt.datum, tt.cs, tt.areaOfUse)

			if tt.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tt.code, crs.Code())
			assert.Equal(t, tt.crsName, crs.Name())
			assert.Equal(t, Geographic, crs.Type())
			assert.True(t, crs.IsGeographic())
			assert.Equal(t, tt.datum, crs.Datum())
			assert.Equal(t, tt.cs, crs.CoordinateSystem())

			// Check area of use
			minLon, minLat, maxLon, maxLat := crs.AreaOfUse()
			if tt.areaOfUse != nil {
				assert.Equal(t, tt.areaOfUse[0], minLon)
				assert.Equal(t, tt.areaOfUse[1], minLat)
				assert.Equal(t, tt.areaOfUse[2], maxLon)
				assert.Equal(t, tt.areaOfUse[3], maxLat)
			} else {
				// Should default to global
				assert.Equal(t, -180.0, minLon)
				assert.Equal(t, -90.0, minLat)
				assert.Equal(t, 180.0, maxLon)
				assert.Equal(t, 90.0, maxLat)
			}

			// Check WKT is not empty
			assert.True(t, len(crs.WKT()) > 0, "WKT() returned empty string")
		})
	}
}

// TestCommonGeographicCRS tests the predefined common geographic CRS.
func TestCommonGeographicCRS(t *testing.T) {
	tests := []struct {
		crs  *GeographicCRS
		code string
		name string
	}{
		{WGS84, "EPSG:4326", "WGS 84"},
		{NAD83, "EPSG:4269", "NAD83"},
		{NAD27, "EPSG:4267", "NAD27"},
		{OSGB36, "EPSG:4277", "OSGB 1936"},
		{ED50, "EPSG:4230", "ED50"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.code, tt.crs.Code())
			assert.Equal(t, tt.name, tt.crs.Name())
			assert.True(t, tt.crs.IsGeographic())
		})
	}
}

// TestProjectedCRS tests projected CRS creation and properties.
func TestProjectedCRS(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		crsName    string
		baseCRS    CRS
		cs         CoordinateSystem
		projection string
		areaOfUse  []float64
		wantError  bool
	}{
		{
			name:       "Web Mercator",
			code:       "EPSG:3857",
			crsName:    "WGS 84 / Pseudo-Mercator",
			baseCRS:    WGS84,
			cs:         CartesianCS2D,
			projection: "Mercator",
			areaOfUse:  []float64{-180, -85.06, 180, 85.06},
			wantError:  false,
		},
		{
			name:       "UTM Zone 33N",
			code:       "EPSG:32633",
			crsName:    "WGS 84 / UTM zone 33N",
			baseCRS:    WGS84,
			cs:         CartesianCS2D,
			projection: "Transverse Mercator",
			areaOfUse:  []float64{12, 0, 18, 84},
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			crs, err := NewProjectedCRS(tt.code, tt.crsName, tt.baseCRS, tt.cs, tt.projection, tt.areaOfUse)

			if tt.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tt.code, crs.Code())
			assert.Equal(t, tt.crsName, crs.Name())
			assert.Equal(t, Projected, crs.Type())
			assert.False(t, crs.IsGeographic())
			assert.Equal(t, tt.baseCRS, crs.BaseCRS())
			assert.Equal(t, tt.projection, crs.Projection())

			// Datum should come from base CRS
			assert.Equal(t, tt.baseCRS.Datum(), crs.Datum())

			// Check WKT is not empty
			assert.True(t, len(crs.WKT()) > 0, "WKT() returned empty string")
		})
	}
}

// TestCommonProjectedCRS tests the predefined common projected CRS.
func TestCommonProjectedCRS(t *testing.T) {
	tests := []struct {
		crs  *ProjectedCRS
		code string
		name string
	}{
		{WebMercator, "EPSG:3857", "WGS 84 / Pseudo-Mercator"},
		{UTM33N, "EPSG:32633", "WGS 84 / UTM zone 33N"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.code, tt.crs.Code())
			assert.Equal(t, tt.name, tt.crs.Name())
			assert.False(t, tt.crs.IsGeographic())
		})
	}
}

// TestCoordinateSystem tests coordinate system creation and properties.
func TestCoordinateSystem(t *testing.T) {
	axes := []Axis{
		{Name: "Longitude", Direction: East, Unit: Degree},
		{Name: "Latitude", Direction: North, Unit: Degree},
	}

	cs, err := NewCoordinateSystem(axes)
	require.NoError(t, err)

	assert.Equal(t, 2, cs.Dimension())

	axis0, err := cs.Axis(0)
	require.NoError(t, err)
	assert.Equal(t, "Longitude", axis0.Name)
	assert.Equal(t, East, axis0.Direction)
	assert.Equal(t, "degree", axis0.Unit.Name)

	axis1, err := cs.Axis(1)
	require.NoError(t, err)
	assert.Equal(t, "Latitude", axis1.Name)
	assert.Equal(t, North, axis1.Direction)

	// Test error cases
	_, err = NewCoordinateSystem([]Axis{})
	require.Error(t, err)

	_, err = cs.Axis(-1)
	require.Error(t, err)

	_, err = cs.Axis(2)
	require.Error(t, err)
}

// TestUnits tests unit conversion.
func TestUnits(t *testing.T) {
	tests := []struct {
		name    string
		value   float64
		from    Unit
		to      Unit
		want    float64
		wantErr bool
	}{
		{
			name:    "Metres to feet",
			value:   100.0,
			from:    Metre,
			to:      Foot,
			want:    328.08398950131235,
			wantErr: false,
		},
		{
			name:    "Feet to metres",
			value:   100.0,
			from:    Foot,
			to:      Metre,
			want:    30.48,
			wantErr: false,
		},
		{
			name:    "Degrees to radians",
			value:   180.0,
			from:    Degree,
			to:      Radian,
			want:    math.Pi,
			wantErr: false,
		},
		{
			name:    "Radians to degrees",
			value:   math.Pi,
			from:    Radian,
			to:      Degree,
			want:    180.0,
			wantErr: false,
		},
		{
			name:    "Metres to degrees (incompatible)",
			value:   100.0,
			from:    Metre,
			to:      Degree,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertValue(tt.value, tt.from, tt.to)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			assert.InDelta(t, tt.want, result, 1e-9)
		})
	}
}

// TestDirectionString tests Direction.String().
func TestDirectionString(t *testing.T) {
	tests := []struct {
		direction Direction
		want      string
	}{
		{North, "North"},
		{South, "South"},
		{East, "East"},
		{West, "West"},
		{Up, "Up"},
		{Down, "Down"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.direction.String())
		})
	}
}

// TestCRSTypeString tests CRSType.String().
func TestCRSTypeString(t *testing.T) {
	tests := []struct {
		crsType CRSType
		want    string
	}{
		{Geographic, "Geographic"},
		{Projected, "Projected"},
		{Geocentric, "Geocentric"},
		{Vertical, "Vertical"},
		{Compound, "Compound"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.crsType.String())
		})
	}
}
