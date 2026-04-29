package wkt2

import (
	"testing"
)

// FuzzParse exercises the WKT2 CRS parser with arbitrary string input.
// The wkt2 package is parse-only (no encoder in v0.1), so the only
// invariant is that Parse never panics — every malformed input must
// surface as an error.
func FuzzParse(f *testing.F) {
	seeds := []string{
		`GEOGCRS["WGS 84",DATUM["World Geodetic System 1984",ELLIPSOID["WGS 84",6378137,298.257223563,LENGTHUNIT["metre",1]]],PRIMEM["Greenwich",0],ID["EPSG",4326]]`,
		`GEODCRS["WGS 84",DATUM["WGS_1984",ELLIPSOID["WGS 84",6378137,298.257223563]],ID["EPSG",4326]]`,
		`PROJCRS["WGS 84 / Pseudo-Mercator",BASEGEOGCRS["WGS 84",DATUM["WGS_1984",ELLIPSOID["WGS 84",6378137,298.257223563]]],ID["EPSG",3857]]`,
		`BOUNDCRS[SOURCECRS[GEOGCRS["WGS 84",ID["EPSG",4326]]],TARGETCRS[GEOGCRS["NAD83",ID["EPSG",4269]]],ABRIDGEDTRANSFORMATION["NAD83 to WGS 84"]]`,
		`GEOGCRS["x"]`,
		`GEOGCRS["x",ID["EPSG",4326]]`,
		``,
		` `,
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		// Parse must not panic on any input. Errors are expected and fine.
		_, _ = Parse(s)
	})
}
