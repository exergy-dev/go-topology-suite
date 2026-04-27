package buffer

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/internal/fixture"
	"github.com/robert-malhotra/go-topology-suite/io/wkt"
	"github.com/stretchr/testify/require"
)

func TestGoldenRegressionBufferFixtures(t *testing.T) {
	tests := []fixture.WKTCase{
		{
			Name:        "zero distance point returns equivalent point",
			A:           "POINT (5 5)",
			Operation:   "buffer",
			ExpectedWKT: "POINT (5 5)",
			Source:      "JTS/GEOS parity fixture",
		},
		{
			Name:        "negative polygon buffer collapses to empty",
			A:           "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
			Operation:   "buffer:-5",
			ExpectedWKT: "POLYGON EMPTY",
			Source:      "JTS/GEOS parity fixture",
		},
		{
			Name:      "collection buffer dissolves overlapping point buffers",
			A:         "GEOMETRYCOLLECTION (POINT (0 0), POINT (5 0))",
			Operation: "buffer:5",
			Predicate: "valid-polygonal",
			Source:    "JTS/GEOS parity fixture",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			g := fixture.MustGeometry(t, tt.A)
			result := bufferFixtureOperation(t, tt.Operation, g)

			if tt.ExpectedWKT != "" {
				require.Equal(t, tt.ExpectedWKT, wkt.MarshalString(result))
			}
			if tt.Predicate == "valid-polygonal" {
				require.False(t, result.IsEmpty())
				require.True(t, result.IsValid())
				switch result.(type) {
				case *geom.Polygon, *geom.MultiPolygon:
				default:
					t.Fatalf("expected polygonal result, got %T", result)
				}
			}
		})
	}
}

func bufferFixtureOperation(t *testing.T, op string, g geom.Geometry) geom.Geometry {
	t.Helper()

	switch op {
	case "buffer":
		return Buffer(g, 0)
	case "buffer:-5":
		return Buffer(g, -5)
	case "buffer:5":
		return Buffer(g, 5)
	default:
		t.Fatalf("unknown buffer fixture operation %q", op)
		return nil
	}
}
