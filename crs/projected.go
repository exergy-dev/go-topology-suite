package crs

import "fmt"

// ProjectedCRS represents a projected coordinate reference system.
// A projected CRS uses a map projection to convert geographic coordinates
// (latitude/longitude) to planar coordinates (e.g., easting/northing).
type ProjectedCRS struct {
	code             string
	name             string
	baseCRS          CRS
	coordinateSystem CoordinateSystem
	areaOfUse        [4]float64 // minLon, minLat, maxLon, maxLat
	projection       string     // Projection name/method (e.g., "Transverse Mercator")
}

// NewProjectedCRS creates a new projected CRS.
// The baseCRS is typically a geographic CRS that provides the geodetic datum.
// The area of use is optional and can be specified as nil for global coverage,
// or as a 4-element slice [minLon, minLat, maxLon, maxLat] in degrees.
func NewProjectedCRS(code, name string, baseCRS CRS, cs CoordinateSystem, projection string, areaOfUse []float64) (*ProjectedCRS, error) {
	if baseCRS == nil {
		return nil, fmt.Errorf("base CRS cannot be nil")
	}
	if cs == nil {
		return nil, fmt.Errorf("coordinate system cannot be nil")
	}

	crs := &ProjectedCRS{
		code:             code,
		name:             name,
		baseCRS:          baseCRS,
		coordinateSystem: cs,
		projection:       projection,
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

// Code returns the authority code (e.g., "EPSG:32633" for UTM Zone 33N).
func (p *ProjectedCRS) Code() string {
	return p.code
}

// Name returns the name of the CRS.
func (p *ProjectedCRS) Name() string {
	return p.name
}

// Type returns Projected.
func (p *ProjectedCRS) Type() CRSType {
	return Projected
}

// IsGeographic returns false.
func (p *ProjectedCRS) IsGeographic() bool {
	return false
}

// Datum returns the datum from the base CRS.
func (p *ProjectedCRS) Datum() Datum {
	return p.baseCRS.Datum()
}

// CoordinateSystem returns the coordinate system.
func (p *ProjectedCRS) CoordinateSystem() CoordinateSystem {
	return p.coordinateSystem
}

// AreaOfUse returns the geographic area where this CRS is valid.
func (p *ProjectedCRS) AreaOfUse() (minLon, minLat, maxLon, maxLat float64) {
	return p.areaOfUse[0], p.areaOfUse[1], p.areaOfUse[2], p.areaOfUse[3]
}

// BaseCRS returns the base geographic CRS.
func (p *ProjectedCRS) BaseCRS() CRS {
	return p.baseCRS
}

// Projection returns the projection method name.
func (p *ProjectedCRS) Projection() string {
	return p.projection
}

// WKT returns a simplified Well-Known Text representation.
func (p *ProjectedCRS) WKT() string {
	ellipsoid := p.baseCRS.Datum().Ellipsoid()
	return fmt.Sprintf(`PROJCS["%s",GEOGCS["%s",DATUM["%s",SPHEROID["%s",%.1f,%.9f]],PRIMEM["Greenwich",0],UNIT["degree",0.0174532925199433]],PROJECTION["%s"],UNIT["metre",1]]`,
		p.name,
		p.baseCRS.Name(),
		p.baseCRS.Datum().Name(),
		ellipsoid.Name(),
		ellipsoid.SemiMajorAxis(),
		ellipsoid.InverseFlattening(),
		p.projection)
}

// Common projected coordinate reference systems.
var (
	// WebMercator is the Web Mercator projection (EPSG:3857).
	// Used by web mapping applications like Google Maps, OpenStreetMap.
	// Also known as WGS 84 / Pseudo-Mercator or Spherical Mercator.
	WebMercator *ProjectedCRS

	// UTM33N is UTM Zone 33N (EPSG:32633).
	// Covers parts of Europe including Norway, Sweden, and Germany.
	UTM33N *ProjectedCRS
)

// initProjectedCRS initializes the common projected CRS instances.
// This is called from init.go after geographic CRS are initialized.
func initProjectedCRS() {
	var err error

	// Web Mercator (EPSG:3857)
	WebMercator, err = NewProjectedCRS(
		"EPSG:3857",
		"WGS 84 / Pseudo-Mercator",
		WGS84,
		CartesianCS2D,
		"Mercator",
		[]float64{-180, -85.06, 180, 85.06}, // Web Mercator limits
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create Web Mercator CRS: %v", err))
	}

	// UTM Zone 33N (EPSG:32633)
	UTM33N, err = NewProjectedCRS(
		"EPSG:32633",
		"WGS 84 / UTM zone 33N",
		WGS84,
		CartesianCS2D,
		"Transverse Mercator",
		[]float64{12, 0, 18, 84}, // Zone 33N coverage
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create UTM33N CRS: %v", err))
	}
}
