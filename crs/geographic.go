package crs

import "fmt"

// GeographicCRS represents a geographic coordinate reference system.
// Geographic CRS use latitude and longitude coordinates on an ellipsoidal
// Earth model.
type GeographicCRS struct {
	code              string
	name              string
	datum             Datum
	coordinateSystem  CoordinateSystem
	areaOfUse         [4]float64 // minLon, minLat, maxLon, maxLat
}

// NewGeographicCRS creates a new geographic CRS.
// The area of use is optional and can be specified as nil for global coverage,
// or as a 4-element slice [minLon, minLat, maxLon, maxLat] in degrees.
func NewGeographicCRS(code, name string, datum Datum, cs CoordinateSystem, areaOfUse []float64) (*GeographicCRS, error) {
	if datum == nil {
		return nil, fmt.Errorf("datum cannot be nil")
	}
	if cs == nil {
		return nil, fmt.Errorf("coordinate system cannot be nil")
	}

	crs := &GeographicCRS{
		code:             code,
		name:             name,
		datum:            datum,
		coordinateSystem: cs,
		areaOfUse:        [4]float64{-180, -90, 180, 90}, // Default: global
	}

	if len(areaOfUse) > 0 {
		if len(areaOfUse) != 4 {
			return nil, fmt.Errorf("areaOfUse must have 4 elements [minLon, minLat, maxLon, maxLat], got %d", len(areaOfUse))
		}
		copy(crs.areaOfUse[:], areaOfUse)
	}

	return crs, nil
}

// Code returns the authority code (e.g., "EPSG:4326").
func (g *GeographicCRS) Code() string {
	return g.code
}

// Name returns the name of the CRS.
func (g *GeographicCRS) Name() string {
	return g.name
}

// Type returns Geographic.
func (g *GeographicCRS) Type() CRSType {
	return Geographic
}

// IsGeographic returns true.
func (g *GeographicCRS) IsGeographic() bool {
	return true
}

// Datum returns the datum.
func (g *GeographicCRS) Datum() Datum {
	return g.datum
}

// CoordinateSystem returns the coordinate system.
func (g *GeographicCRS) CoordinateSystem() CoordinateSystem {
	return g.coordinateSystem
}

// AreaOfUse returns the geographic area where this CRS is valid.
func (g *GeographicCRS) AreaOfUse() (minLon, minLat, maxLon, maxLat float64) {
	return g.areaOfUse[0], g.areaOfUse[1], g.areaOfUse[2], g.areaOfUse[3]
}

// WKT returns a simplified Well-Known Text representation.
func (g *GeographicCRS) WKT() string {
	ellipsoid := g.datum.Ellipsoid()
	return fmt.Sprintf(`GEOGCS["%s",DATUM["%s",SPHEROID["%s",%.1f,%.9f]],PRIMEM["Greenwich",0],UNIT["degree",0.0174532925199433]]`,
		g.name,
		g.datum.Name(),
		ellipsoid.Name(),
		ellipsoid.SemiMajorAxis(),
		ellipsoid.InverseFlattening())
}

// Common geographic coordinate reference systems.
var (
	// WGS84 is the World Geodetic System 1984 geographic CRS (EPSG:4326).
	// This is the most widely used geographic CRS, used by GPS and web mapping.
	WGS84 *GeographicCRS

	// NAD83 is the North American Datum 1983 geographic CRS (EPSG:4269).
	// Used in North America for surveying and mapping.
	NAD83 *GeographicCRS

	// NAD27 is the North American Datum 1927 geographic CRS (EPSG:4267).
	// Historical datum used in North America before NAD83.
	NAD27 *GeographicCRS

	// OSGB36 is the Ordnance Survey Great Britain 1936 geographic CRS (EPSG:4277).
	// Used in the United Kingdom.
	OSGB36 *GeographicCRS

	// ED50 is the European Datum 1950 geographic CRS (EPSG:4230).
	// Historical datum used in Europe.
	ED50 *GeographicCRS
)

// initGeographicCRS initializes the common geographic CRS instances.
// This is called from init.go after datums are initialized.
func initGeographicCRS() {
	var err error

	// WGS 84 (EPSG:4326)
	WGS84, err = NewGeographicCRS("EPSG:4326", "WGS 84", WGS84Datum, EllipsoidalCS2D, nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create WGS84 CRS: %v", err))
	}

	// NAD83 (EPSG:4269)
	NAD83, err = NewGeographicCRS("EPSG:4269", "NAD83", NAD83Datum, EllipsoidalCS2D,
		[]float64{-180, 14.92, 180, 86.46}) // North America
	if err != nil {
		panic(fmt.Sprintf("failed to create NAD83 CRS: %v", err))
	}

	// NAD27 (EPSG:4267)
	NAD27, err = NewGeographicCRS("EPSG:4267", "NAD27", NAD27Datum, EllipsoidalCS2D,
		[]float64{-180, 14.92, 180, 86.46}) // North America
	if err != nil {
		panic(fmt.Sprintf("failed to create NAD27 CRS: %v", err))
	}

	// OSGB 1936 (EPSG:4277)
	OSGB36, err = NewGeographicCRS("EPSG:4277", "OSGB 1936", OSGB36Datum, EllipsoidalCS2D,
		[]float64{-8.82, 49.79, 1.92, 60.94}) // United Kingdom
	if err != nil {
		panic(fmt.Sprintf("failed to create OSGB36 CRS: %v", err))
	}

	// ED50 (EPSG:4230)
	ED50, err = NewGeographicCRS("EPSG:4230", "ED50", ED50Datum, EllipsoidalCS2D,
		[]float64{-16.1, 32.88, 40.18, 84.17}) // Europe
	if err != nil {
		panic(fmt.Sprintf("failed to create ED50 CRS: %v", err))
	}
}
