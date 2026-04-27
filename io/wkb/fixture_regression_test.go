package wkb

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/internal/fixture"
	"github.com/stretchr/testify/require"
)

func TestGoldenRegressionWKBParserFixtures(t *testing.T) {
	tests := []fixture.WKTCase{
		{
			Name:      "EWKB nested collection preserves SRID",
			A:         "GEOMETRYCOLLECTION (POINT (1 2), GEOMETRYCOLLECTION (POINT (3 4)))",
			Operation: "ewkb-roundtrip",
			Predicate: "srid:4326",
			Source:    "EWKB nested SRID regression fixture",
		},
		{
			Name:      "WKB nested collection preserves Z and M",
			A:         "GEOMETRYCOLLECTION ZM (POINT (1 2 3 4), GEOMETRYCOLLECTION ZM (LINESTRING (0 0 5 6, 1 1 7 8)))",
			Operation: "wkb-zm-roundtrip",
			Predicate: "nested-zm",
			Source:    "WKB nested dimensional metadata regression fixture",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			g := fixture.MustGeometry(t, tt.A)
			switch tt.Operation {
			case "ewkb-roundtrip":
				setSRIDRecursive(g, 4326)
				data, err := MarshalEWKB(g)
				require.NoError(t, err)

				roundTrip, err := Unmarshal(data)
				require.NoError(t, err)
				require.Equal(t, 4326, roundTrip.SRID())
				nested := roundTrip.(*geom.GeometryCollection).GeometryN(1).(*geom.GeometryCollection)
				require.Equal(t, 4326, nested.SRID())
			case "wkb-zm-roundtrip":
				data, err := MarshalWithOptions(g, Options{OutputDimension: 4})
				require.NoError(t, err)

				roundTrip, err := Unmarshal(data)
				require.NoError(t, err)
				point := roundTrip.(*geom.GeometryCollection).GeometryN(0).(*geom.Point)
				require.True(t, point.Coordinate().HasZ())
				require.True(t, point.Coordinate().HasM())
				nested := roundTrip.(*geom.GeometryCollection).GeometryN(1).(*geom.GeometryCollection)
				line := nested.GeometryN(0).(*geom.LineString)
				require.True(t, line.Coordinates()[0].HasZ())
				require.True(t, line.Coordinates()[0].HasM())
			default:
				t.Fatalf("unknown WKB fixture operation %q", tt.Operation)
			}
		})
	}
}

func setSRIDRecursive(g geom.Geometry, srid int) {
	if setter, ok := g.(interface{ SetSRID(int) }); ok {
		setter.SetSRID(srid)
	}
	if gc, ok := g.(*geom.GeometryCollection); ok {
		for i := 0; i < gc.NumGeometries(); i++ {
			setSRIDRecursive(gc.GeometryN(i), srid)
		}
	}
}
