package wkt

import (
	"strings"
	"testing"

	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/geom"
)

func TestEncodePoint(t *testing.T) {
	p := geom.NewPoint(crs.WGS84, geom.XY{X: 1, Y: 2})
	got, err := Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if got != "POINT (1 2)" {
		t.Errorf("got %q", got)
	}
}

func TestEncodeEmptyPoint(t *testing.T) {
	p := geom.NewEmptyPoint(nil, geom.LayoutXY)
	got, _ := Marshal(p)
	if got != "POINT EMPTY" {
		t.Errorf("got %q", got)
	}
}

func TestEncodePolygonWithHole(t *testing.T) {
	outer := []geom.XY{{X: 0, Y: 0}, {X: 0, Y: 10}, {X: 10, Y: 10}, {X: 10, Y: 0}, {X: 0, Y: 0}}
	hole := []geom.XY{{X: 2, Y: 2}, {X: 2, Y: 4}, {X: 4, Y: 4}, {X: 4, Y: 2}, {X: 2, Y: 2}}
	p := geom.NewPolygon(nil, outer, hole)
	got, _ := Marshal(p)
	want := "POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0), (2 2, 2 4, 4 4, 4 2, 2 2))"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestRoundTripAllTypes(t *testing.T) {
	cases := []string{
		"POINT (1 2)",
		"POINT EMPTY",
		"LINESTRING (0 0, 1 1, 2 2)",
		"LINESTRING EMPTY",
		"POLYGON ((0 0, 0 1, 1 1, 1 0, 0 0))",
		"POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0), (2 2, 2 4, 4 4, 4 2, 2 2))",
		"POLYGON EMPTY",
		"MULTIPOINT ((1 2), (3 4))",
		"MULTIPOINT EMPTY",
		"MULTILINESTRING ((0 0, 1 1), (2 2, 3 3))",
		"MULTILINESTRING EMPTY",
		"MULTIPOLYGON (((0 0, 0 1, 1 1, 1 0, 0 0)), ((2 2, 2 3, 3 3, 3 2, 2 2)))",
		"MULTIPOLYGON EMPTY",
		"GEOMETRYCOLLECTION (POINT (1 2), LINESTRING (0 0, 1 1))",
		"GEOMETRYCOLLECTION EMPTY",
	}
	for _, in := range cases {
		t.Run(in, func(t *testing.T) {
			g, err := Unmarshal(in)
			if err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			out, err := Marshal(g)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			if out != in {
				t.Errorf("round-trip differs:\n got %q\nwant %q", out, in)
			}
		})
	}
}

func TestSRIDPrefix(t *testing.T) {
	g, err := Unmarshal("SRID=4326;POINT (1 2)")
	if err != nil {
		t.Fatal(err)
	}
	if g.CRS() == nil || g.CRS().Code != 4326 {
		t.Errorf("SRID prefix not attached: %+v", g.CRS())
	}
	out, _ := MarshalEWKT(g)
	if out != "SRID=4326;POINT (1 2)" {
		t.Errorf("EWKT round-trip = %q", out)
	}
}

func TestCaseInsensitive(t *testing.T) {
	g, err := Unmarshal("point (1 2)")
	if err != nil {
		t.Fatal(err)
	}
	if g.Type() != geom.PointType {
		t.Errorf("got %v", g.Type())
	}
}

func TestErrors(t *testing.T) {
	cases := []string{
		"FOO (1 2)",
		"POINT (1)",
		"POINT (1 2",       // missing close paren
		"POINT (1 2) junk", // trailing
	}
	for _, in := range cases {
		if _, err := Unmarshal(in); err == nil {
			t.Errorf("expected error for %q", in)
		}
	}
}

func TestEncodeXYZLayout(t *testing.T) {
	p := geom.NewPointXYZ(nil, geom.XYZ{X: 1, Y: 2, Z: 3})
	got, _ := Marshal(p)
	if got != "POINT Z (1 2 3)" {
		t.Errorf("got %q", got)
	}
}

func TestDecodeXYZLineString(t *testing.T) {
	g, err := Unmarshal("LINESTRING Z (0 0 1, 1 1 2, 2 2 3)")
	if err != nil {
		t.Fatal(err)
	}
	ls := g.(*geom.LineString)
	if ls.Layout() != geom.LayoutXYZ {
		t.Errorf("layout = %v, want XYZ", ls.Layout())
	}
	if ls.NumPoints() != 3 {
		t.Errorf("NumPoints = %d", ls.NumPoints())
	}
}

func TestEncodeOmitsTrailingZeros(t *testing.T) {
	p := geom.NewPoint(nil, geom.XY{X: 1.5, Y: 2})
	got, _ := Marshal(p)
	if !strings.Contains(got, "1.5") {
		t.Errorf("expected 1.5 in %q", got)
	}
	if strings.Contains(got, "2.0") {
		t.Errorf("did not expect 2.0 in %q", got)
	}
}
