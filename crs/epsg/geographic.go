package epsg

import (
	"github.com/robert-malhotra/go-topology-suite/crs"
)

// Common geographic CRS definitions with EPSG codes.

// WGS84 is the World Geodetic System 1984 geographic CRS (EPSG:4326).
//
// This is the most widely used geographic CRS, used by GPS, GeoJSON,
// and most web mapping applications. Coordinates are in decimal degrees.
//
// Properties:
//   - EPSG Code: 4326
//   - Datum: WGS 84
//   - Ellipsoid: WGS 84
//   - Axis order: Longitude (East), Latitude (North) in degrees
//   - Bounds: Longitude [-180, 180], Latitude [-90, 90]
//
// Important: The official EPSG definition uses Lat/Lon order, but many software
// systems (GeoJSON, PostGIS, etc.) use Lon/Lat order. This package follows
// the common Lon/Lat order convention.
var WGS84 = crs.WGS84

// NAD83 is the North American Datum 1983 geographic CRS (EPSG:4269).
//
// NAD83 is the standard datum for North America, used by US federal agencies
// and many state and local mapping systems. It uses the GRS80 ellipsoid
// which is nearly identical to WGS84.
//
// Properties:
//   - EPSG Code: 4269
//   - Datum: NAD83
//   - Ellipsoid: GRS 1980
//   - Axis order: Longitude (East), Latitude (North) in degrees
//   - Bounds: Primarily North America
//
// Note: NAD83 has been updated multiple times (NAD83(CORS96), NAD83(2011), etc.).
// This represents the original NAD83(1986).
var NAD83 = crs.NAD83

// NAD27 is the North American Datum 1927 geographic CRS (EPSG:4267).
//
// NAD27 is the older North American datum, replaced by NAD83 in the 1980s.
// It uses the Clarke 1866 ellipsoid and requires significant transformation
// to modern datums like WGS84 or NAD83.
//
// Properties:
//   - EPSG Code: 4267
//   - Datum: NAD27
//   - Ellipsoid: Clarke 1866
//   - Axis order: Longitude (East), Latitude (North) in degrees
//   - Bounds: Primarily North America
//
// Important: Transformation from NAD27 to WGS84/NAD83 varies by region.
// Differences can be 10+ meters. Use NADCON grids for accurate conversions.
var NAD27 = crs.NAD27

// ETRS89 is the European Terrestrial Reference System 1989 geographic CRS (EPSG:4258).
//
// ETRS89 is the standard datum for Europe, used by EU mapping agencies.
// It uses the GRS80 ellipsoid and is essentially equivalent to WGS84
// for most practical purposes.
//
// Properties:
//   - EPSG Code: 4258
//   - Datum: ETRS89 (based on GRS80 ellipsoid)
//   - Ellipsoid: GRS 1980
//   - Axis order: Longitude (East), Latitude (North) in degrees
//   - Bounds: Europe
//
// Note: ETRS89 is fixed to the European tectonic plate, while WGS84
// follows the global reference frame. The difference accumulates over
// time but is negligible (< 1m) for most applications.
var ETRS89 *crs.GeographicCRS

func init() {
	var err error

	// ETRS89 (EPSG:4258) - European Terrestrial Reference System 1989
	// We create this one manually since it's not in the base crs package
	etrs89Datum, err := crs.NewDatum("ETRS89", crs.GRS80Ellipsoid, 0.0,
		[]float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0})
	if err != nil {
		panic("failed to create ETRS89 datum: " + err.Error())
	}

	ETRS89, err = crs.NewGeographicCRS("EPSG:4258", "ETRS89", etrs89Datum,
		crs.EllipsoidalCS2D, []float64{-16.1, 32.88, 40.18, 84.73})
	if err != nil {
		panic("failed to create ETRS89 CRS: " + err.Error())
	}

	// Register all geographic CRS
	registerCRS(WGS84)
	registerCRS(NAD83)
	registerCRS(NAD27)
	registerCRS(ETRS89)
}
