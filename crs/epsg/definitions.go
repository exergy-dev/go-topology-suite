package epsg

import (
	"math"

	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/crs/proj"
)

// init populates the Definition pointer on every named CRS this package
// registers. UTM-range Definitions are wired in projected.go.
//
// AxisLonLat is the default even for CRSes (notably EPSG:4326) that
// EPSG defines as (lat, lon): real-world data and JTS/GeoTools-style
// stacks both store (lon, lat). Users needing strict-EPSG axis order
// construct their own *crs.CRS.
func init() {
	wireGeographic()
	wireProjected()
}

func wireGeographic() {
	WGS84.Definition = &crs.Definition{Datum: crs.DatumWGS84}
	NAD83.Definition = &crs.Definition{Datum: crs.DatumNAD83}
	NAD27.Definition = &crs.Definition{Datum: crs.DatumNAD27}
	WGS72.Definition = &crs.Definition{Datum: crs.DatumWGS72}
	ETRS89.Definition = &crs.Definition{Datum: crs.DatumETRS89}
	WGS84_3D.Definition = &crs.Definition{Datum: crs.DatumWGS84}
	CGCS2000.Definition = &crs.Definition{Datum: crs.DatumCGCS2000}
	Beijing1954.Definition = &crs.Definition{Datum: crs.DatumBeijing1954}
}

func wireProjected() {
	const d2r = math.Pi / 180.0

	// EPSG:3857 — Web Mercator pseudo-projection, spherical math.
	WebMercator.Definition = &crs.Definition{
		Datum:      crs.DatumWebMercator,
		Projection: proj.NewWebMercator(),
	}

	// EPSG:2154 — RGF93 / Lambert-93 (LCC 2SP).
	Lambert93.Definition = &crs.Definition{
		Datum: crs.DatumRGF93,
		Projection: proj.NewLambertConformalConic2SP(
			crs.GRS80Ellipsoid.A, crs.GRS80Ellipsoid.E2(),
			3.0*d2r, 46.5*d2r, 49.0*d2r, 44.0*d2r,
			700000.0, 6600000.0,
		),
	}

	// EPSG:27700 — OSGB36 / British National Grid (Transverse Mercator).
	BritishNationalGrid.Definition = &crs.Definition{
		Datum: crs.DatumOSGB36,
		Projection: proj.NewTransverseMercator(
			crs.Airy1830Ellipsoid.A, crs.Airy1830Ellipsoid.E2(),
			-2.0*d2r, 49.0*d2r, 0.9996012717,
			400000.0, -100000.0,
		),
	}

	// EPSG:5070 — NAD83 / Conus Albers (Albers Equal-Area).
	ConusAlbers.Definition = &crs.Definition{
		Datum: crs.DatumNAD83,
		Projection: proj.NewAlbersEqualAreaConic(
			crs.GRS80Ellipsoid.A, crs.GRS80Ellipsoid.E2(),
			-96.0*d2r, 23.0*d2r, 29.5*d2r, 45.5*d2r,
			0.0, 0.0,
		),
	}

	// EPSG:3035 — ETRS89 / LAEA Europe (Lambert Azimuthal Equal-Area).
	EuropeLAEA.Definition = &crs.Definition{
		Datum: crs.DatumETRS89,
		Projection: proj.NewLambertAzimuthalEqualArea(
			crs.GRS80Ellipsoid.A, crs.GRS80Ellipsoid.E2(),
			10.0*d2r, 52.0*d2r,
			4321000.0, 3210000.0,
		),
	}

}
