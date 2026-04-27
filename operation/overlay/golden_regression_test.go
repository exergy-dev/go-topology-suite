package overlay

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/internal/fixture"
	"github.com/robert-malhotra/go-topology-suite/io/wkt"
	"github.com/stretchr/testify/require"
)

func mustWKTGeometry(t *testing.T, text string) geom.Geometry {
	t.Helper()

	return fixture.MustGeometry(t, text)
}

func TestGoldenRegressionPolygonOverlayFixtures(t *testing.T) {
	tests := []fixture.WKTCase{
		{
			Name:        "touching polygons at a point intersect as point",
			A:           "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
			B:           "POLYGON ((10 10, 20 10, 20 20, 10 20, 10 10))",
			Operation:   "intersection",
			ExpectedWKT: "POINT (10 10)",
			Source:      "GEOS/JTS parity fixture",
		},
		{
			Name:        "touching polygons at an edge intersect as line",
			A:           "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
			B:           "POLYGON ((10 0, 20 0, 20 10, 10 10, 10 0))",
			Operation:   "intersection",
			ExpectedWKT: "LINESTRING (10 0, 10 10)",
		},
		{
			Name:        "edge-adjacent polygons union",
			A:           "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
			B:           "POLYGON ((10 0, 20 0, 20 10, 10 10, 10 0))",
			Operation:   "union",
			ExpectedWKT: "POLYGON ((0 0, 10 0, 20 0, 20 10, 10 10, 0 10, 0 0))",
		},
		{
			Name:        "polygon with hole contains shell island",
			A:           "POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0), (5 5, 15 5, 15 15, 5 15, 5 5))",
			B:           "POLYGON ((1 1, 4 1, 4 4, 1 4, 1 1))",
			Operation:   "intersection",
			ExpectedWKT: "POLYGON ((1 1, 4 1, 4 4, 1 4, 1 1))",
		},
		{
			Name:        "polygon with hole difference adds shell island as hole",
			A:           "POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0), (5 5, 15 5, 15 15, 5 15, 5 5))",
			B:           "POLYGON ((1 1, 4 1, 4 4, 1 4, 1 1))",
			Operation:   "difference",
			ExpectedWKT: "POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0), (1 1, 1 4, 4 4, 4 1, 1 1), (5 5, 5 15, 15 15, 15 5, 5 5))",
		},
		{
			Name:        "polygon with hole intersection clips around hole",
			A:           "POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0), (5 5, 15 5, 15 15, 5 15, 5 5))",
			B:           "POLYGON ((2 2, 18 2, 18 8, 2 8, 2 2))",
			Operation:   "intersection",
			ExpectedWKT: "POLYGON ((2 2, 18 2, 18 8, 15 8, 15 5, 5 5, 5 8, 2 8, 2 2))",
		},
		{
			Name:        "polygon with hole is disjoint from outside polygon",
			A:           "POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0), (5 5, 15 5, 15 15, 5 15, 5 5))",
			B:           "POLYGON ((30 30, 35 30, 35 35, 30 35, 30 30))",
			Operation:   "intersection",
			ExpectedWKT: "GEOMETRYCOLLECTION EMPTY",
		},
		{
			Name:        "polygon touching hole boundary intersects as boundary line",
			A:           "POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0), (5 5, 15 5, 15 15, 5 15, 5 5))",
			B:           "POLYGON ((5 6, 7 6, 7 8, 5 8, 5 6))",
			Operation:   "intersection",
			ExpectedWKT: "LINESTRING (5 8, 5 6)",
		},
		{
			Name:        "polygon touching hole corner intersects as point",
			A:           "POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0), (5 5, 15 5, 15 15, 5 15, 5 5))",
			B:           "POLYGON ((5 5, 7 6, 6 7, 5 5))",
			Operation:   "intersection",
			ExpectedWKT: "POINT (5 5)",
		},
		{
			Name:        "line crossing polygon clips to polygon interior",
			A:           "LINESTRING (-5 5, 15 5)",
			B:           "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
			Operation:   "intersection",
			ExpectedWKT: "LINESTRING (0 5, 10 5)",
		},
		{
			Name:        "line crossing polygon difference keeps exterior segments",
			A:           "LINESTRING (-5 5, 15 5)",
			B:           "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
			Operation:   "difference",
			ExpectedWKT: "MULTILINESTRING ((-5 5, 0 5), (10 5, 15 5))",
		},
		{
			Name:        "nested multipolygon intersection returns contained polygon",
			A:           "MULTIPOLYGON (((0 0, 20 0, 20 20, 0 20, 0 0)), ((30 30, 40 30, 40 40, 30 40, 30 30)))",
			B:           "POLYGON ((2 2, 4 2, 4 4, 2 4, 2 2))",
			Operation:   "intersection",
			ExpectedWKT: "POLYGON ((2 2, 4 2, 4 4, 2 4, 2 2))",
		},
		{
			Name:        "multipolygon intersection returns two clipped polygons",
			A:           "MULTIPOLYGON (((0 0, 10 0, 10 10, 0 10, 0 0)), ((20 0, 30 0, 30 10, 20 10, 20 0)))",
			B:           "POLYGON ((5 5, 25 5, 25 15, 5 15, 5 5))",
			Operation:   "intersection",
			ExpectedWKT: "MULTIPOLYGON (((5 5, 10 5, 10 10, 5 10, 5 5)), ((20 5, 25 5, 25 10, 20 10, 20 5)))",
		},
		{
			Name:        "multipolygon union merges one component and keeps another",
			A:           "MULTIPOLYGON (((0 0, 10 0, 10 10, 0 10, 0 0)), ((20 0, 30 0, 30 10, 20 10, 20 0)))",
			B:           "POLYGON ((5 0, 15 0, 15 10, 5 10, 5 0))",
			Operation:   "union",
			ExpectedWKT: "MULTIPOLYGON (((0 0, 5 0, 10 0, 15 0, 15 10, 10 10, 5 10, 0 10, 0 0)), ((20 0, 30 0, 30 10, 20 10, 20 0)))",
		},
		{
			Name:        "multipolygon union keeps disjoint polygon components",
			A:           "MULTIPOLYGON (((0 0, 10 0, 10 10, 0 10, 0 0)), ((20 0, 30 0, 30 10, 20 10, 20 0)))",
			B:           "POLYGON ((40 0, 50 0, 50 10, 40 10, 40 0))",
			Operation:   "union",
			ExpectedWKT: "MULTIPOLYGON (((0 0, 10 0, 10 10, 0 10, 0 0)), ((20 0, 30 0, 30 10, 20 10, 20 0)), ((40 0, 50 0, 50 10, 40 10, 40 0)))",
		},
		{
			Name:        "overlapping squares symmetric difference returns polygonal ring",
			A:           "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
			B:           "POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5))",
			Operation:   "symdifference",
			ExpectedWKT: "POLYGON ((0 0, 10 0, 10 5, 15 5, 15 15, 5 15, 5 10, 0 10, 0 0), (5 5, 5 10, 10 10, 10 5, 5 5))",
		},
		{
			Name:        "line on polygon boundary intersects as boundary line",
			A:           "LINESTRING (0 0, 10 0)",
			B:           "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
			Operation:   "intersection",
			ExpectedWKT: "LINESTRING (0 0, 10 0)",
		},
		{
			Name:        "line on polygon boundary has empty difference",
			A:           "LINESTRING (0 0, 10 0)",
			B:           "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
			Operation:   "difference",
			ExpectedWKT: "LINESTRING EMPTY",
		},
		{
			Name:        "collection polygon member intersects nested polygon",
			A:           "GEOMETRYCOLLECTION (POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0)), LINESTRING (30 0, 40 0))",
			B:           "POLYGON ((5 5, 10 5, 10 10, 5 10, 5 5))",
			Operation:   "intersection",
			ExpectedWKT: "POLYGON ((5 5, 10 5, 10 10, 5 10, 5 5))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			a := mustWKTGeometry(t, tt.A)
			b := mustWKTGeometry(t, tt.B)

			var result geom.Geometry
			switch tt.Operation {
			case "intersection":
				result = Intersection(a, b)
			case "union":
				result = Union(a, b)
			case "difference":
				result = Difference(a, b)
			case "symdifference":
				result = SymDifference(a, b)
			default:
				t.Fatalf("unknown overlay operation %q", tt.Operation)
			}

			require.Equal(t, tt.ExpectedWKT, wkt.MarshalString(result))
		})
	}
}
