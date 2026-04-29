package geojson

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/wkt"
)

func TestEncodeAllTypes(t *testing.T) {
	cases := []struct {
		w    string
		want string
	}{
		{"POINT (1 2)", `{"type":"Point","coordinates":[1,2]}`},
		{"LINESTRING (0 0, 1 1)", `{"type":"LineString","coordinates":[[0,0],[1,1]]}`},
		{"POLYGON ((0 0, 0 1, 1 1, 1 0, 0 0))", `{"type":"Polygon","coordinates":[[[0,0],[0,1],[1,1],[1,0],[0,0]]]}`},
		{"MULTIPOINT ((1 2), (3 4))", `{"type":"MultiPoint","coordinates":[[1,2],[3,4]]}`},
		{"MULTILINESTRING ((0 0, 1 1), (2 2, 3 3))", `{"type":"MultiLineString","coordinates":[[[0,0],[1,1]],[[2,2],[3,3]]]}`},
	}
	for _, tc := range cases {
		t.Run(tc.w, func(t *testing.T) {
			g, err := wkt.Unmarshal(tc.w)
			if err != nil {
				t.Fatal(err)
			}
			got, err := Marshal(g)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tc.want {
				t.Errorf("got  %s\nwant %s", got, tc.want)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	wkts := []string{
		"POINT (1 2)",
		"LINESTRING (0 0, 1 1, 2 2)",
		"POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0))",
		"POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0), (2 2, 2 4, 4 4, 4 2, 2 2))",
		"MULTIPOINT ((1 2), (3 4))",
		"MULTILINESTRING ((0 0, 1 1), (2 2, 3 3))",
		"MULTIPOLYGON (((0 0, 0 1, 1 1, 1 0, 0 0)))",
	}
	for _, w := range wkts {
		t.Run(w, func(t *testing.T) {
			g, err := wkt.Unmarshal(w)
			if err != nil {
				t.Fatal(err)
			}
			data, err := Marshal(g)
			if err != nil {
				t.Fatal(err)
			}
			back, err := Unmarshal(data)
			if err != nil {
				t.Fatalf("Unmarshal: %v\ndata: %s", err, data)
			}
			out, _ := wkt.Marshal(back)
			if out != w {
				t.Errorf("round-trip differs:\n got %q\nwant %q", out, w)
			}
		})
	}
}

func TestPointXYZ(t *testing.T) {
	p := geom.NewPointXYZ(nil, geom.XYZ{X: 1, Y: 2, Z: 3})
	got, _ := Marshal(p)
	want := `{"type":"Point","coordinates":[1,2,3]}`
	if string(got) != want {
		t.Errorf("got %s, want %s", got, want)
	}

	back, err := Unmarshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if back.Layout() != geom.LayoutXYZ {
		t.Errorf("layout = %v", back.Layout())
	}
}

func TestFeatureRoundTrip(t *testing.T) {
	src := `{"type":"Feature","id":42,"bbox":[0,0,1,1],"geometry":{"type":"Point","coordinates":[0.5,0.5]},"properties":{"name":"origin"}}`
	var f Feature
	if err := json.Unmarshal([]byte(src), &f); err != nil {
		t.Fatal(err)
	}
	if f.Geometry.Type() != geom.PointType {
		t.Errorf("expected Point, got %v", f.Geometry.Type())
	}
	if f.Properties["name"] != "origin" {
		t.Errorf("properties = %v", f.Properties)
	}
	out, err := json.Marshal(&f)
	if err != nil {
		t.Fatal(err)
	}
	// Round trip through json.Unmarshal again to compare semantically.
	var f2 Feature
	if err := json.Unmarshal(out, &f2); err != nil {
		t.Fatalf("re-decode: %v\ndata: %s", err, out)
	}
	if f2.Properties["name"] != "origin" {
		t.Errorf("after round-trip: %v", f2.Properties)
	}
}

func TestFeatureCollectionForeign(t *testing.T) {
	src := `{"type":"FeatureCollection","features":[],"title":"test","attribution":"me"}`
	var fc FeatureCollection
	if err := json.Unmarshal([]byte(src), &fc); err != nil {
		t.Fatal(err)
	}
	if len(fc.Foreign) != 2 {
		t.Errorf("Foreign len = %d, want 2", len(fc.Foreign))
	}
	out, _ := json.Marshal(&fc)
	if !strings.Contains(string(out), `"title":"test"`) {
		t.Errorf("foreign 'title' lost: %s", out)
	}
}

func TestUnmarshalEmptyPoint(t *testing.T) {
	// Empty arrays decode to POINT EMPTY for cross-format compat.
	g, err := Unmarshal([]byte(`{"type":"Point","coordinates":[]}`))
	if err != nil {
		t.Fatal(err)
	}
	if !g.IsEmpty() {
		t.Errorf("empty-array point should be empty")
	}
}

func TestUnmarshalErrors(t *testing.T) {
	cases := []string{
		`{"type":"Foo"}`,
		`{"type":"Point"}`,
		`{"type":"Point","coordinates":"bad"}`,
		`not json`,
	}
	for _, c := range cases {
		if _, err := Unmarshal([]byte(c)); err == nil {
			t.Errorf("expected error for %q", c)
		}
	}
}
