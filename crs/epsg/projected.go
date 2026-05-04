package epsg

import (
	"github.com/exergy-dev/go-topology-suite/crs"
	"github.com/exergy-dev/go-topology-suite/crs/proj"
)

// init registers the bulk EPSG ranges that aren't worth hand-naming:
//
//   - WGS84 / UTM zones 1N..60N  (32601..32660)
//   - WGS84 / UTM zones 1S..60S  (32701..32760)
//   - NAD83  / UTM zones 1N..23N (26901..26923)
//   - ETRS89 / UTM zones 32N..35N (25832..25835)
//
// Definition is wired at registration time (rather than in a second
// init in definitions.go) so the named CRS list and the UTM ranges can
// each populate themselves without depending on init-order.
func init() {
	registerUTMRange(32601, 32660, false, crs.DatumWGS84, 32600)
	registerUTMRange(32701, 32760, true, crs.DatumWGS84, 32700)
	registerUTMRange(26901, 26923, false, crs.DatumNAD83, 26900)
	registerUTMRange(25832, 25835, false, crs.DatumETRS89, 25800)
}

// registerUTMRange registers UTM zones in a contiguous EPSG code block.
// codeBase is the value such that (code - codeBase) yields the zone
// number (e.g. 32600 → zones 1..60 on the northern hemisphere).
func registerUTMRange(first, last int, southern bool, datum crs.Datum, codeBase int) {
	for code := first; code <= last; code++ {
		zone := code - codeBase
		register(&crs.CRS{
			Authority: "EPSG",
			Code:      code,
			Kind:      crs.Projected,
			Definition: &crs.Definition{
				Datum:      datum,
				Projection: proj.UTM(zone, southern, datum.Ellipsoid),
			},
		})
	}
}
