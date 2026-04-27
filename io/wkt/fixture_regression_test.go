package wkt_test

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/internal/fixture"
	"github.com/robert-malhotra/go-topology-suite/io/wkt"
	"github.com/stretchr/testify/require"
)

func TestGoldenRegressionWKTParserFixtures(t *testing.T) {
	tests := []fixture.WKTCase{
		{
			Name:      "strict parser rejects trailing tokens",
			A:         "POINT (1 2) garbage",
			Operation: "parse",
			Predicate: "error",
			Source:    "strict parser regression fixture",
		},
		{
			Name:      "strict parser rejects malformed geometry collection",
			A:         "GEOMETRYCOLLECTION (POINT (0 0),)",
			Operation: "parse",
			Predicate: "error",
			Source:    "strict parser regression fixture",
		},
		{
			Name:      "strict parser rejects malformed multipolygon",
			A:         "MULTIPOLYGON (((0 0, 1 0, 1 1, 0 1)))",
			Operation: "parse",
			Predicate: "error",
			Source:    "strict parser regression fixture",
		},
		{
			Name:      "nested geometry collection with Z values",
			A:         "GEOMETRYCOLLECTION Z (POINT (1 2 3), LINESTRING (0 0 1, 1 1 2))",
			Operation: "parse",
			Predicate: "nested-z",
			Source:    "strict parser regression fixture",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			g, err := wkt.UnmarshalString(tt.A)
			if tt.Predicate == "error" {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.Predicate == "nested-z" {
				gc := g.(*geom.GeometryCollection)
				point := gc.GeometryN(0).(*geom.Point)
				require.True(t, point.Coordinate().HasZ())
				require.Equal(t, 3.0, point.Coordinate().GetZ())
				line := gc.GeometryN(1).(*geom.LineString)
				require.True(t, line.Coordinates()[0].HasZ())
				require.Equal(t, 1.0, line.Coordinates()[0].GetZ())
				return
			}
			if tt.ExpectedWKT != "" {
				require.Equal(t, tt.ExpectedWKT, wkt.MarshalString(g))
			}
		})
	}
}
