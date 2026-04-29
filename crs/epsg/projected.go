package epsg

import "github.com/terra-geo/terra/crs"

// init registers the bulk EPSG ranges that aren't worth hand-naming:
//
//   - WGS84 / UTM zones 1N..60N  (32601..32660)
//   - WGS84 / UTM zones 1S..60S  (32701..32760)
//   - NAD83  / UTM zones 1N..23N (26901..26923)
//   - ETRS89 / UTM zones 32N..35N (25832..25835)
//
// Each entry is a fresh *crs.CRS with Kind=Projected and the conventional
// EPSG authority code; WKT2 is left empty (callers can fetch the WKT from
// the upstream EPSG dataset if they need it).
func init() {
	registerRange("EPSG", 32601, 32660, crs.Projected) // WGS84 UTM N
	registerRange("EPSG", 32701, 32760, crs.Projected) // WGS84 UTM S
	registerRange("EPSG", 26901, 26923, crs.Projected) // NAD83 UTM N
	registerRange("EPSG", 25832, 25835, crs.Projected) // ETRS89 UTM N
}

// registerRange inserts a contiguous block of EPSG codes [first, last] with
// the same authority and kind. It is used to declaratively register the UTM
// zone families.
func registerRange(authority string, first, last int, kind crs.Kind) {
	for code := first; code <= last; code++ {
		register(&crs.CRS{Authority: authority, Code: code, Kind: kind})
	}
}
