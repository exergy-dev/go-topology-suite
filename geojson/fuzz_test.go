package geojson

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// FuzzUnmarshal exercises the GeoJSON parser with arbitrary byte input.
// It asserts:
//  1. Unmarshal never panics — malformed JSON or geometry must surface
//     as an error, not a runtime panic.
//  2. If Unmarshal succeeds, Marshal must succeed and the result must
//     reach a fixed point on a second round-trip (re-decode + re-encode
//     yields byte-identical output).
//
// We test idempotence rather than first-encoding equality because input
// formatting (whitespace, key order in raw JSON) need not match Marshal's
// canonical form.
func FuzzUnmarshal(f *testing.F) {
	seeds := []string{
		`{"type":"Point","coordinates":[1,2]}`,
		`{"type":"Point","coordinates":[1,2,3]}`,
		`{"type":"LineString","coordinates":[[0,0],[1,1]]}`,
		`{"type":"Polygon","coordinates":[[[0,0],[1,0],[1,1],[0,1],[0,0]]]}`,
		`{"type":"MultiPoint","coordinates":[[1,2],[3,4]]}`,
		`{"type":"MultiLineString","coordinates":[[[0,0],[1,1]],[[2,2],[3,3]]]}`,
		`{"type":"MultiPolygon","coordinates":[[[[0,0],[0,1],[1,1],[1,0],[0,0]]]]}`,
		`{"type":"GeometryCollection","geometries":[{"type":"Point","coordinates":[1,2]}]}`,
	}
	for _, s := range seeds {
		f.Add([]byte(s))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		g, err := Unmarshal(data)
		if err != nil {
			return // expected for most random input
		}
		require.NotNilf(t, g, "Unmarshal returned (nil, nil) for %q", data)
		first, err := Marshal(g)
		require.NoErrorf(t, err, "Marshal of parsed geometry failed\ninput: %q", data)
		g2, err := Unmarshal(first)
		require.NoErrorf(t, err, "re-Unmarshal of own Marshal output failed\nfirst: %s", first)
		second, err := Marshal(g2)
		require.NoError(t, err, "re-Marshal failed")
		require.Equalf(t, first, second, "round-trip not idempotent:\nfirst:  %s\nsecond: %s\ninput: %q", first, second, data)
	})
}
