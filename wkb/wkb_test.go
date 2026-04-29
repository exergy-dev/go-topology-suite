package wkb

import (
	"encoding/binary"
	"testing"

	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/wkt"
)

// roundTrip encodes g, decodes the result, and re-encodes — round-trip
// equality on the WKT projection is the simplest cross-format invariant.
func roundTrip(t *testing.T, g geom.Geometry, opts ...Option) geom.Geometry {
	t.Helper()
	data, err := Marshal(g, opts...)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	return got
}

func TestRoundTripAllTypes(t *testing.T) {
	wkts := []string{
		"POINT (1 2)",
		"LINESTRING (0 0, 1 1, 2 2)",
		"POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0))",
		"POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0), (2 2, 2 4, 4 4, 4 2, 2 2))",
		"MULTIPOINT ((1 2), (3 4))",
		"MULTILINESTRING ((0 0, 1 1), (2 2, 3 3))",
		"MULTIPOLYGON (((0 0, 0 1, 1 1, 1 0, 0 0)), ((2 2, 2 3, 3 3, 3 2, 2 2)))",
		"GEOMETRYCOLLECTION (POINT (1 2), LINESTRING (0 0, 1 1))",
	}
	for _, in := range wkts {
		t.Run(in, func(t *testing.T) {
			g, err := wkt.Unmarshal(in)
			if err != nil {
				t.Fatal(err)
			}
			got := roundTrip(t, g)
			out, _ := wkt.Marshal(got)
			if out != in {
				t.Errorf("round-trip differs:\n got %q\nwant %q", out, in)
			}
		})
	}
}

func TestEWKBSRIDPreserved(t *testing.T) {
	g, err := wkt.Unmarshal("POINT (1 2)")
	if err != nil {
		t.Fatal(err)
	}
	// Attach CRS via a fresh point (Unmarshal without SRID prefix has nil CRS).
	p := geom.NewPoint(crs.WGS84, geom.XY{X: 1, Y: 2})
	data, err := Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Unmarshal(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.CRS() == nil || got.CRS().Code != 4326 {
		t.Errorf("SRID lost in round-trip: %+v", got.CRS())
	}

	// Now without CRS — no SRID flag should be set.
	data2, _ := Marshal(g) // g has nil CRS
	got2, _ := Unmarshal(data2)
	if got2.CRS() != nil {
		t.Errorf("unexpected CRS: %+v", got2.CRS())
	}
}

func TestBigEndianRoundTrip(t *testing.T) {
	p := geom.NewPoint(nil, geom.XY{X: 1, Y: 2})
	data, err := Marshal(p, WithByteOrder(binary.BigEndian))
	if err != nil {
		t.Fatal(err)
	}
	if data[0] != 0 {
		t.Errorf("byte-order tag = %d, want 0 (XDR)", data[0])
	}
	got, err := Unmarshal(data)
	if err != nil {
		t.Fatal(err)
	}
	pp := got.(*geom.Point)
	if pp.XY().X != 1 || pp.XY().Y != 2 {
		t.Errorf("XY = %+v", pp.XY())
	}
}

func TestISOMode(t *testing.T) {
	// XYZ point under ISO encoding has type code 1001 (POINT Z).
	p := geom.NewPointXYZ(nil, geom.XYZ{X: 1, Y: 2, Z: 3})
	data, err := Marshal(p, WithISO())
	if err != nil {
		t.Fatal(err)
	}
	// type code lives at bytes [1..5] little-endian by default
	tc := binary.LittleEndian.Uint32(data[1:5])
	if tc != 1001 {
		t.Errorf("ISO Z type code = %d, want 1001", tc)
	}
	got, err := Unmarshal(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.Layout() != geom.LayoutXYZ {
		t.Errorf("layout after ISO round-trip = %v", got.Layout())
	}
}

func TestPointEmpty(t *testing.T) {
	e := geom.NewEmptyPoint(nil, geom.LayoutXY)
	data, err := Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Unmarshal(data)
	if err != nil {
		t.Fatal(err)
	}
	if !got.IsEmpty() {
		t.Errorf("empty point did not survive round-trip")
	}
}

func TestUnmarshalErrors(t *testing.T) {
	cases := [][]byte{
		{},                 // empty
		{2},                // bad byte-order tag
		{1, 99, 0, 0, 0},   // unknown type code
		{1, 1, 0, 0, 0, 1}, // truncated point body
	}
	for _, b := range cases {
		if _, err := Unmarshal(b); err == nil {
			t.Errorf("expected error for %v", b)
		}
	}
}
