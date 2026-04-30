package wkt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// FuzzUnmarshal exercises the WKT/EWKT parser with arbitrary string input.
// It asserts two invariants:
//  1. Unmarshal never panics — malformed input must surface as an error.
//  2. If Unmarshal succeeds, Marshal of the result must itself parse, and
//     the produced WKT must be byte-identical on a second round-trip.
//
// Numeric value equality is not required (Marshal may canonicalise NaN
// representations, exponent forms etc.), only that re-encoding the
// re-decoded value reaches a fixed point.
func FuzzUnmarshal(f *testing.F) {
	seeds := []string{
		"POINT (1 2)",
		"POINT EMPTY",
		"POINT Z (1 2 3)",
		"POINT M (1 2 3)",
		"POINT ZM (1 2 3 4)",
		"LINESTRING (0 0, 1 1)",
		"LINESTRING EMPTY",
		"POLYGON ((0 0, 1 0, 1 1, 0 1, 0 0))",
		"POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0), (2 2, 2 4, 4 4, 4 2, 2 2))",
		"MULTIPOINT ((1 2), (3 4))",
		"MULTILINESTRING ((0 0, 1 1), (2 2, 3 3))",
		"MULTIPOLYGON (((0 0, 0 1, 1 1, 1 0, 0 0)), ((2 2, 2 3, 3 3, 3 2, 2 2)))",
		"GEOMETRYCOLLECTION (POINT (1 2), LINESTRING (0 0, 1 1))",
		"SRID=4326;POINT (1 2)",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		g, err := Unmarshal(s)
		if err != nil {
			return // expected for most random input
		}
		require.NotNil(t, g, "Unmarshal returned (nil, nil) for input %q", s)
		out, err := Marshal(g)
		require.NoErrorf(t, err, "Marshal of parsed geometry failed\ninput: %q", s)
		g2, err := Unmarshal(out)
		require.NoErrorf(t, err, "re-Unmarshal of own Marshal output failed\ninput: %q\nintermediate: %q", s, out)
		out2, err := Marshal(g2)
		require.NoErrorf(t, err, "re-Marshal failed\nintermediate: %q", out)
		require.Equalf(t, out, out2, "round-trip not idempotent:\nfirst:  %q\nsecond: %q\ninput: %q", out, out2, s)
	})
}
