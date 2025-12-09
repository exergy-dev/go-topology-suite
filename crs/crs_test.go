package crs

import (
	"math"
	"testing"
)

const epsilon = 1e-9

func almostEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

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
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if ellipsoid.Name() != tt.name {
				t.Errorf("Name() = %v, want %v", ellipsoid.Name(), tt.name)
			}

			if ellipsoid.SemiMajorAxis() != tt.semiMajorAxis {
				t.Errorf("SemiMajorAxis() = %v, want %v", ellipsoid.SemiMajorAxis(), tt.semiMajorAxis)
			}

			if ellipsoid.InverseFlattening() != tt.inverseFlattening {
				t.Errorf("InverseFlattening() = %v, want %v", ellipsoid.InverseFlattening(), tt.inverseFlattening)
			}

			if !almostEqual(ellipsoid.SemiMinorAxis(), tt.wantSemiMinor, 1e-6) {
				t.Errorf("SemiMinorAxis() = %v, want %v", ellipsoid.SemiMinorAxis(), tt.wantSemiMinor)
			}

			if !almostEqual(ellipsoid.Eccentricity(), tt.wantEccentricity, 1e-12) {
				t.Errorf("Eccentricity() = %v, want %v", ellipsoid.Eccentricity(), tt.wantEccentricity)
			}

			// Verify eccentricity squared
			wantEccentricitySq := tt.wantEccentricity * tt.wantEccentricity
			if !almostEqual(ellipsoid.EccentricitySquared(), wantEccentricitySq, 1e-12) {
				t.Errorf("EccentricitySquared() = %v, want %v", ellipsoid.EccentricitySquared(), wantEccentricitySq)
			}
		})
	}
}

// TestEllipsoidFromAF tests ellipsoid creation from flattening.
func TestEllipsoidFromAF(t *testing.T) {
	// WGS84: f = 1/298.257223563 ≈ 0.003352810664747
	flattening := 1.0 / 298.257223563
	ellipsoid, err := NewEllipsoidFromAF("WGS84", 6378137.0, flattening)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !almostEqual(ellipsoid.InverseFlattening(), 298.257223563, 1e-6) {
		t.Errorf("InverseFlattening() = %v, want 298.257223563", ellipsoid.InverseFlattening())
	}

	// Test invalid flattening
	_, err = NewEllipsoidFromAF("Invalid", 6378137.0, 1.5)
	if err == nil {
		t.Error("expected error for flattening >= 1, got nil")
	}

	_, err = NewEllipsoidFromAF("Invalid", 6378137.0, -0.1)
	if err == nil {
		t.Error("expected error for negative flattening, got nil")
	}
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
			if tt.ellipsoid.Name() != tt.name {
				t.Errorf("Name() = %v, want %v", tt.ellipsoid.Name(), tt.name)
			}

			if tt.ellipsoid.SemiMajorAxis() != tt.semiMajorAxis {
				t.Errorf("SemiMajorAxis() = %v, want %v", tt.ellipsoid.SemiMajorAxis(), tt.semiMajorAxis)
			}

			if !almostEqual(tt.ellipsoid.InverseFlattening(), tt.inverseFlattening, 1e-6) {
				t.Errorf("InverseFlattening() = %v, want %v", tt.ellipsoid.InverseFlattening(), tt.inverseFlattening)
			}

			// Verify semi-minor axis is less than semi-major axis
			if tt.ellipsoid.SemiMinorAxis() >= tt.ellipsoid.SemiMajorAxis() {
				t.Errorf("SemiMinorAxis() >= SemiMajorAxis()")
			}

			// Verify eccentricity is in valid range [0, 1)
			e := tt.ellipsoid.Eccentricity()
			if e < 0 || e >= 1 {
				t.Errorf("Eccentricity() = %v, want [0, 1)", e)
			}
		})
	}
}

// TestDatum tests datum creation and properties.
func TestDatum(t *testing.T) {
	ellipsoid := WGS84Ellipsoid

	tests := []struct {
		name              string
		datumName         string
		primeMeridian     float64
		toWGS84Params     []float64
		wantError         bool
	}{
		{
			name:              "WGS84",
			datumName:         "WGS 84",
			primeMeridian:     0.0,
			toWGS84Params:     nil,
			wantError:         false,
		},
		{
			name:              "With transformation",
			datumName:         "Custom",
			primeMeridian:     0.0,
			toWGS84Params:     []float64{1.0, 2.0, 3.0, 0.1, 0.2, 0.3, 0.5},
			wantError:         false,
		},
		{
			name:              "Invalid transformation params",
			datumName:         "Invalid",
			primeMeridian:     0.0,
			toWGS84Params:     []float64{1.0, 2.0}, // Wrong number
			wantError:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			datum, err := NewDatum(tt.datumName, ellipsoid, tt.primeMeridian, tt.toWGS84Params)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if datum.Name() != tt.datumName {
				t.Errorf("Name() = %v, want %v", datum.Name(), tt.datumName)
			}

			if datum.Ellipsoid() != ellipsoid {
				t.Errorf("Ellipsoid() != expected ellipsoid")
			}

			if datum.PrimeMeridian() != tt.primeMeridian {
				t.Errorf("PrimeMeridian() = %v, want %v", datum.PrimeMeridian(), tt.primeMeridian)
			}

			// Check transformation parameters
			dx, dy, dz, rx, ry, rz, ds := datum.ToWGS84Params()
			if tt.toWGS84Params != nil {
				if dx != tt.toWGS84Params[0] || dy != tt.toWGS84Params[1] || dz != tt.toWGS84Params[2] {
					t.Errorf("Translation params incorrect")
				}
				if rx != tt.toWGS84Params[3] || ry != tt.toWGS84Params[4] || rz != tt.toWGS84Params[5] {
					t.Errorf("Rotation params incorrect")
				}
				if ds != tt.toWGS84Params[6] {
					t.Errorf("Scale param incorrect")
				}
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
			if tt.datum.Name() != tt.name {
				t.Errorf("Name() = %v, want %v", tt.datum.Name(), tt.name)
			}

			if tt.datum.Ellipsoid().Name() != tt.ellipsoidName {
				t.Errorf("Ellipsoid().Name() = %v, want %v", tt.datum.Ellipsoid().Name(), tt.ellipsoidName)
			}

			// All common datums use Greenwich prime meridian
			if tt.datum.PrimeMeridian() != 0.0 {
				t.Errorf("PrimeMeridian() = %v, want 0.0", tt.datum.PrimeMeridian())
			}
		})
	}
}

// TestGeographicCRS tests geographic CRS creation and properties.
func TestGeographicCRS(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		crsName    string
		datum      Datum
		cs         CoordinateSystem
		areaOfUse  []float64
		wantError  bool
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
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if crs.Code() != tt.code {
				t.Errorf("Code() = %v, want %v", crs.Code(), tt.code)
			}

			if crs.Name() != tt.crsName {
				t.Errorf("Name() = %v, want %v", crs.Name(), tt.crsName)
			}

			if crs.Type() != Geographic {
				t.Errorf("Type() = %v, want Geographic", crs.Type())
			}

			if !crs.IsGeographic() {
				t.Error("IsGeographic() = false, want true")
			}

			if crs.Datum() != tt.datum {
				t.Errorf("Datum() != expected datum")
			}

			if crs.CoordinateSystem() != tt.cs {
				t.Errorf("CoordinateSystem() != expected coordinate system")
			}

			// Check area of use
			minLon, minLat, maxLon, maxLat := crs.AreaOfUse()
			if tt.areaOfUse != nil {
				if minLon != tt.areaOfUse[0] || minLat != tt.areaOfUse[1] ||
					maxLon != tt.areaOfUse[2] || maxLat != tt.areaOfUse[3] {
					t.Errorf("AreaOfUse() = [%v, %v, %v, %v], want %v",
						minLon, minLat, maxLon, maxLat, tt.areaOfUse)
				}
			} else {
				// Should default to global
				if minLon != -180 || minLat != -90 || maxLon != 180 || maxLat != 90 {
					t.Errorf("AreaOfUse() = [%v, %v, %v, %v], want global [-180, -90, 180, 90]",
						minLon, minLat, maxLon, maxLat)
				}
			}

			// Check WKT is not empty
			wkt := crs.WKT()
			if len(wkt) == 0 {
				t.Error("WKT() returned empty string")
			}
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
			if tt.crs.Code() != tt.code {
				t.Errorf("Code() = %v, want %v", tt.crs.Code(), tt.code)
			}

			if tt.crs.Name() != tt.name {
				t.Errorf("Name() = %v, want %v", tt.crs.Name(), tt.name)
			}

			if !tt.crs.IsGeographic() {
				t.Error("IsGeographic() = false, want true")
			}
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
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if crs.Code() != tt.code {
				t.Errorf("Code() = %v, want %v", crs.Code(), tt.code)
			}

			if crs.Name() != tt.crsName {
				t.Errorf("Name() = %v, want %v", crs.Name(), tt.crsName)
			}

			if crs.Type() != Projected {
				t.Errorf("Type() = %v, want Projected", crs.Type())
			}

			if crs.IsGeographic() {
				t.Error("IsGeographic() = true, want false")
			}

			if crs.BaseCRS() != tt.baseCRS {
				t.Errorf("BaseCRS() != expected base CRS")
			}

			if crs.Projection() != tt.projection {
				t.Errorf("Projection() = %v, want %v", crs.Projection(), tt.projection)
			}

			// Datum should come from base CRS
			if crs.Datum() != tt.baseCRS.Datum() {
				t.Errorf("Datum() != base CRS datum")
			}

			// Check WKT is not empty
			wkt := crs.WKT()
			if len(wkt) == 0 {
				t.Error("WKT() returned empty string")
			}
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
			if tt.crs.Code() != tt.code {
				t.Errorf("Code() = %v, want %v", tt.crs.Code(), tt.code)
			}

			if tt.crs.Name() != tt.name {
				t.Errorf("Name() = %v, want %v", tt.crs.Name(), tt.name)
			}

			if tt.crs.IsGeographic() {
				t.Error("IsGeographic() = true, want false")
			}
		})
	}
}

// TestCoordinateSystem tests coordinate system creation and properties.
func TestCoordinateSystem(t *testing.T) {
	axes := []Axis{
		{Name: "Longitude", Direction: East, Unit: Degree},
		{Name: "Latitude", Direction: North, Unit: Degree},
	}

	cs := NewCoordinateSystem(axes)

	if cs.Dimension() != 2 {
		t.Errorf("Dimension() = %v, want 2", cs.Dimension())
	}

	axis0 := cs.Axis(0)
	if axis0.Name != "Longitude" {
		t.Errorf("Axis(0).Name = %v, want Longitude", axis0.Name)
	}
	if axis0.Direction != East {
		t.Errorf("Axis(0).Direction = %v, want East", axis0.Direction)
	}
	if axis0.Unit.Name != "degree" {
		t.Errorf("Axis(0).Unit.Name = %v, want degree", axis0.Unit.Name)
	}

	axis1 := cs.Axis(1)
	if axis1.Name != "Latitude" {
		t.Errorf("Axis(1).Name = %v, want Latitude", axis1.Name)
	}
	if axis1.Direction != North {
		t.Errorf("Axis(1).Direction = %v, want North", axis1.Direction)
	}
}

// TestUnits tests unit conversion.
func TestUnits(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		from     Unit
		to       Unit
		want     float64
		wantErr  bool
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
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !almostEqual(result, tt.want, 1e-9) {
				t.Errorf("ConvertValue() = %v, want %v", result, tt.want)
			}
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
			if got := tt.direction.String(); got != tt.want {
				t.Errorf("Direction.String() = %v, want %v", got, tt.want)
			}
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
			if got := tt.crsType.String(); got != tt.want {
				t.Errorf("CRSType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
