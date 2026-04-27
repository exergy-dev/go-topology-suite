package relate

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/internal/fixture"
	"github.com/stretchr/testify/require"
)

func mustRelateWKTGeometry(t *testing.T, text string) geom.Geometry {
	t.Helper()

	return fixture.MustGeometry(t, text)
}

func TestGoldenRegressionDE9IMFixtures(t *testing.T) {
	tests := []fixture.WKTCase{
		{
			Name:          "polygons touching at point",
			A:             "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
			B:             "POLYGON ((10 10, 20 10, 20 20, 10 20, 10 10))",
			DimA:          2,
			DimB:          2,
			ExpectedDE9IM: "FF2F01212",
			Intersects:    true,
			Touches:       true,
		},
		{
			Name:          "polygons touching at edge",
			A:             "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
			B:             "POLYGON ((10 0, 20 0, 20 10, 10 10, 10 0))",
			DimA:          2,
			DimB:          2,
			ExpectedDE9IM: "FF2F11212",
			Intersects:    true,
			Touches:       true,
		},
		{
			Name:          "overlapping polygons",
			A:             "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
			B:             "POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5))",
			DimA:          2,
			DimB:          2,
			ExpectedDE9IM: "212101212",
			Intersects:    true,
		},
		{
			Name:          "polygon with hole contains polygon in shell",
			A:             "POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0), (5 5, 15 5, 15 15, 5 15, 5 5))",
			B:             "POLYGON ((1 1, 4 1, 4 4, 1 4, 1 1))",
			DimA:          2,
			DimB:          2,
			ExpectedDE9IM: "212FF1FF2",
			Intersects:    true,
			Contains:      true,
		},
		{
			Name:          "polygon with hole disjoint from outside polygon",
			A:             "POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0), (5 5, 15 5, 15 15, 5 15, 5 5))",
			B:             "POLYGON ((30 30, 35 30, 35 35, 30 35, 30 30))",
			DimA:          2,
			DimB:          2,
			ExpectedDE9IM: "FF2FF1212",
			Disjoint:      true,
		},
		{
			Name:          "polygon touching inner hole boundary",
			A:             "POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0), (5 5, 15 5, 15 15, 5 15, 5 5))",
			B:             "POLYGON ((5 6, 7 6, 7 8, 5 8, 5 6))",
			DimA:          2,
			DimB:          2,
			ExpectedDE9IM: "FF2F11212",
			Intersects:    true,
			Touches:       true,
		},
		{
			Name:          "polygon exactly fills hole boundary",
			A:             "POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0), (5 5, 15 5, 15 15, 5 15, 5 5))",
			B:             "POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5))",
			DimA:          2,
			DimB:          2,
			ExpectedDE9IM: "FF2F112F2",
			Intersects:    true,
			Touches:       true,
		},
		{
			Name:          "polygon with hole disjoint from polygon inside hole",
			A:             "POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0), (5 5, 15 5, 15 15, 5 15, 5 5))",
			B:             "POLYGON ((8 8, 12 8, 12 12, 8 12, 8 8))",
			DimA:          2,
			DimB:          2,
			ExpectedDE9IM: "FF2FF1212",
			Disjoint:      true,
		},
		{
			Name:          "polygon boundary overlaps hole boundary",
			A:             "POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0), (5 5, 15 5, 15 15, 5 15, 5 5))",
			B:             "POLYGON ((5 6, 7 6, 7 8, 5 8, 5 6))",
			DimA:          2,
			DimB:          2,
			ExpectedDE9IM: "FF2F11212",
			Intersects:    true,
			Touches:       true,
		},
		{
			Name:          "line through polygon boundary",
			A:             "LINESTRING (-5 5, 15 5)",
			B:             "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
			DimA:          1,
			DimB:          2,
			ExpectedDE9IM: "101FF0212",
			Intersects:    true,
		},
		{
			Name:          "line inside polygon",
			A:             "LINESTRING (2 2, 8 8)",
			B:             "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
			DimA:          1,
			DimB:          2,
			ExpectedDE9IM: "1FF0FF212",
			Intersects:    true,
			Within:        true,
		},
		{
			Name:          "line lies on polygon boundary",
			A:             "LINESTRING (0 0, 10 0)",
			B:             "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
			DimA:          1,
			DimB:          2,
			ExpectedDE9IM: "F1FF0F212",
			Intersects:    true,
			Touches:       true,
		},
		{
			Name:          "nested multipolygon contains polygon member",
			A:             "MULTIPOLYGON (((0 0, 20 0, 20 20, 0 20, 0 0)), ((30 30, 40 30, 40 40, 30 40, 30 30)))",
			B:             "POLYGON ((2 2, 4 2, 4 4, 2 4, 2 2))",
			DimA:          2,
			DimB:          2,
			ExpectedDE9IM: "212FF1FF2",
			Intersects:    true,
			Contains:      true,
		},
		{
			Name:          "multipolygon contains one multipolygon component and is disjoint from another",
			A:             "MULTIPOLYGON (((0 0, 20 0, 20 20, 0 20, 0 0)), ((40 0, 60 0, 60 20, 40 20, 40 0)))",
			B:             "MULTIPOLYGON (((2 2, 4 2, 4 4, 2 4, 2 2)), ((80 80, 82 80, 82 82, 80 82, 80 80)))",
			DimA:          2,
			DimB:          2,
			ExpectedDE9IM: "212FF1212",
			Intersects:    true,
		},
		{
			Name:          "collection polygon member relates to contained polygon",
			A:             "GEOMETRYCOLLECTION (POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0)), LINESTRING (30 0, 40 0))",
			B:             "POLYGON ((5 5, 10 5, 10 10, 5 10, 5 5))",
			DimA:          2,
			DimB:          2,
			ExpectedDE9IM: "212FF1FF2",
			Intersects:    true,
			Contains:      true,
		},
		{
			Name:          "geometry collection polygon contains line",
			A:             "GEOMETRYCOLLECTION (POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0)), LINESTRING (20 0, 30 0))",
			B:             "LINESTRING (2 2, 8 8)",
			DimA:          2,
			DimB:          1,
			ExpectedDE9IM: "102FF1FF2",
			Intersects:    true,
			Contains:      true,
		},
		{
			Name:          "polygonal collection shared edge is interior",
			A:             "GEOMETRYCOLLECTION (POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0)), POLYGON ((10 0, 20 0, 20 10, 10 10, 10 0)))",
			B:             "LINESTRING (10 2, 10 8)",
			DimA:          2,
			DimB:          1,
			ExpectedDE9IM: "102FF1FF2",
			Intersects:    true,
			Contains:      true,
		},
		{
			Name:          "polygonal collection relates as dissolved polygon set",
			A:             "GEOMETRYCOLLECTION (POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0)), POLYGON ((10 0, 20 0, 20 10, 10 10, 10 0)))",
			B:             "POLYGON ((0 0, 20 0, 20 10, 0 10, 0 0))",
			DimA:          2,
			DimB:          2,
			ExpectedDE9IM: "2FFF1FFF2",
			Intersects:    true,
			Contains:      true,
			Within:        true,
		},
		{
			Name:          "mixed collection point member contains point outside polygon",
			A:             "GEOMETRYCOLLECTION (POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0)), POINT (20 20))",
			B:             "POINT (20 20)",
			DimA:          2,
			DimB:          0,
			ExpectedDE9IM: "0F2FF1FF2",
			Intersects:    true,
			Contains:      true,
		},
		{
			Name:          "mixed collection point member promotes polygon boundary endpoint",
			A:             "GEOMETRYCOLLECTION (POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0)), POINT (20 20))",
			B:             "GEOMETRYCOLLECTION (LINESTRING (20 20, 30 30))",
			DimA:          2,
			DimB:          1,
			ExpectedDE9IM: "F02FF1102",
			Intersects:    true,
			Touches:       true,
		},
		{
			Name:          "mixed collection shared polygon boundary line is contained",
			A:             "GEOMETRYCOLLECTION (POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0)), POLYGON ((10 0, 20 0, 20 10, 10 10, 10 0)), LINESTRING (10 2, 10 8))",
			B:             "LINESTRING (10 2, 10 8)",
			DimA:          2,
			DimB:          1,
			ExpectedDE9IM: "1F2F01FF2",
			Intersects:    true,
			Contains:      true,
		},
		{
			Name:          "nested mixed collection contains point through child collection",
			A:             "GEOMETRYCOLLECTION (GEOMETRYCOLLECTION (POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))), POINT (20 20))",
			B:             "POINT (5 5)",
			DimA:          2,
			DimB:          0,
			ExpectedDE9IM: "0F2FF1FF2",
			Intersects:    true,
			Contains:      true,
		},
		{
			Name:          "mixed collection equals polygon despite lower dimensional member",
			A:             "GEOMETRYCOLLECTION (POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0)), LINESTRING (2 2, 8 8))",
			B:             "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
			DimA:          2,
			DimB:          2,
			ExpectedDE9IM: "2FF01FFF2",
			Intersects:    true,
			Contains:      true,
			Within:        true,
		},
		{
			Name:          "mixed collection contains lower dimensional collection",
			A:             "GEOMETRYCOLLECTION (POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0)), POINT (20 20))",
			B:             "GEOMETRYCOLLECTION (POINT (5 5), POINT (20 20))",
			DimA:          2,
			DimB:          0,
			ExpectedDE9IM: "0F2FF1FF2",
			Intersects:    true,
			Contains:      true,
		},
		{
			Name:          "mixed collection equals collection with contained line",
			A:             "GEOMETRYCOLLECTION (POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0)), LINESTRING (2 2, 8 8))",
			B:             "GEOMETRYCOLLECTION (POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0)))",
			DimA:          2,
			DimB:          2,
			ExpectedDE9IM: "2FF01FFF2",
			Intersects:    true,
			Contains:      true,
			Within:        true,
		},
		{
			Name:          "mixed collections overlap by polygon and point member",
			A:             "GEOMETRYCOLLECTION (POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0)), POINT (20 20))",
			B:             "GEOMETRYCOLLECTION (POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5)), POINT (20 20))",
			DimA:          2,
			DimB:          2,
			ExpectedDE9IM: "212101212",
			Intersects:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			a := mustRelateWKTGeometry(t, tt.A)
			b := mustRelateWKTGeometry(t, tt.B)

			m := Relate(a, b)
			require.Equal(t, tt.ExpectedDE9IM, m.String())
			require.Equal(t, tt.Intersects, m.IsIntersects())
			require.Equal(t, tt.Disjoint, m.IsDisjoint())
			require.Equal(t, tt.Touches, m.IsTouches(tt.DimA, tt.DimB))
			require.Equal(t, tt.Contains, m.IsContains())
			require.Equal(t, tt.Within, m.IsWithin())

			if _, ok := a.(*geom.GeometryCollection); ok {
				transpose := Relate(b, a)
				require.Equal(t, m.Transpose().String(), transpose.String())
			} else if _, ok := b.(*geom.GeometryCollection); ok {
				transpose := Relate(b, a)
				require.Equal(t, m.Transpose().String(), transpose.String())
			}
		})
	}
}
