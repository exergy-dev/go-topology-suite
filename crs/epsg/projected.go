package epsg

import (
	"fmt"

	"github.com/robert-malhotra/go-topology-suite/crs"
)

// Common projected CRS definitions with EPSG codes.

// WebMercator is the WGS 84 / Pseudo-Mercator projected CRS (EPSG:3857).
//
// This is the de facto standard for web mapping, used by Google Maps,
// OpenStreetMap, Bing Maps, and most tile-based web maps. It uses a
// simplified Mercator projection that treats the Earth as a sphere.
//
// Properties:
//   - EPSG Code: 3857 (also known as EPSG:900913, "Google")
//   - Datum: WGS 84
//   - Projection: Mercator (Spherical)
//   - Axis order: Easting (East), Northing (North)
//   - Bounds: Longitude [-180, 180], Latitude ~[-85.06, 85.06]
//   - Units: Meters
//   - False Easting: 0
//   - False Northing: 0
//
// Important: This projection has significant distortion at high latitudes.
// The latitude range is limited to approximately ±85.06 degrees due to
// the Mercator projection becoming infinite at the poles.
//
// Also known as: Web Mercator, Google Mercator, WGS 84 / Pseudo-Mercator
var WebMercator = crs.WebMercator

// UTMZone returns the UTM CRS for the specified zone and hemisphere.
//
// The Universal Transverse Mercator (UTM) system divides the Earth into
// 60 zones, each 6 degrees of longitude wide. Each zone has a northern
// and southern hemisphere variant.
//
// Parameters:
//   - zone: UTM zone number (1-60)
//   - north: true for northern hemisphere, false for southern
//
// Returns:
//   - CRS for the specified UTM zone
//   - EPSG code: 326xx for northern hemisphere, 327xx for southern
//     where xx is the zero-padded zone number (01-60)
//
// Example:
//
//	utm10n := UTMZone(10, true)  // EPSG:32610 (US West Coast)
//	utm17n := UTMZone(17, true)  // EPSG:32617 (US East Coast)
//	utm32n := UTMZone(32, true)  // EPSG:32632 (Central Europe)
//	utm50s := UTMZone(50, false) // EPSG:32750 (Australia)
//
// Returns an error if zone is not in the range [1, 60].
func UTMZone(zone int, north bool) (crs.CRS, error) {
	if zone < 1 || zone > 60 {
		return nil, fmt.Errorf("invalid UTM zone: %d (must be 1-60)", zone)
	}

	var code string
	var hemisphere string
	var minLat, maxLat float64

	if north {
		code = fmt.Sprintf("EPSG:%d", 32600+zone)
		hemisphere = "N"
		minLat = 0
		maxLat = 84 // UTM northern limit
	} else {
		code = fmt.Sprintf("EPSG:%d", 32700+zone)
		hemisphere = "S"
		minLat = -80 // UTM southern limit
		maxLat = 0
	}

	name := fmt.Sprintf("WGS 84 / UTM zone %d%s", zone, hemisphere)

	// Calculate the longitude bounds for this zone
	// Zone 1 is centered at -177°, zone 60 is centered at 177°
	centralMeridian := float64(zone)*6.0 - 183.0
	minLon := centralMeridian - 3.0
	maxLon := centralMeridian + 3.0

	utm, err := crs.NewProjectedCRS(
		code,
		name,
		crs.WGS84,
		crs.CartesianCS2D,
		"Transverse Mercator",
		[]float64{minLon, minLat, maxLon, maxLat},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create UTM zone %d%s: %w", zone, hemisphere, err)
	}

	return utm, nil
}

// must is a helper that panics on error, used for package-level
// variable initialization where failure indicates a programming error.
func must(c crs.CRS, err error) crs.CRS {
	if err != nil {
		panic(err)
	}
	return c
}

// Common UTM zone CRS for convenience.

// UTM10N is the WGS 84 / UTM zone 10N projected CRS (EPSG:32610).
//
// Covers the US West Coast (including parts of California, Oregon, Washington).
//
// Properties:
//   - EPSG Code: 32610
//   - Datum: WGS 84
//   - Central Meridian: -123° (123°W)
//   - Zone: 10N (120°W to 126°W)
//   - Units: Meters
var UTM10N crs.CRS

// UTM17N is the WGS 84 / UTM zone 17N projected CRS (EPSG:32617).
//
// Covers the US East Coast (including parts of New York, Pennsylvania, Virginia).
//
// Properties:
//   - EPSG Code: 32617
//   - Datum: WGS 84
//   - Central Meridian: -81° (81°W)
//   - Zone: 17N (78°W to 84°W)
//   - Units: Meters
var UTM17N crs.CRS

// UTM32N is the WGS 84 / UTM zone 32N projected CRS (EPSG:32632).
//
// Covers Central Europe (including parts of Germany, Austria, Italy).
//
// Properties:
//   - EPSG Code: 32632
//   - Datum: WGS 84
//   - Central Meridian: 9° (9°E)
//   - Zone: 32N (6°E to 12°E)
//   - Units: Meters
var UTM32N crs.CRS

func init() {
	UTM10N = must(UTMZone(10, true))
	UTM17N = must(UTMZone(17, true))
	UTM32N = must(UTMZone(32, true))

	// Register all projected CRS
	registerCRS(WebMercator)
	registerCRS(UTM10N)
	registerCRS(UTM17N)
	registerCRS(UTM32N)
}
