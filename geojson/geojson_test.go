package geojson

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/wkt"
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
			require.NoError(t, err)
			got, err := Marshal(g)
			require.NoError(t, err)
			assert.Equal(t, tc.want, string(got))
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
			require.NoError(t, err)
			data, err := Marshal(g)
			require.NoError(t, err)
			back, err := Unmarshal(data)
			require.NoErrorf(t, err, "Unmarshal\ndata: %s", data)
			out, _ := wkt.Marshal(back)
			assert.Equal(t, w, out, "round-trip differs")
		})
	}
}

func TestPointXYZ(t *testing.T) {
	p := geom.NewPointXYZ(nil, geom.XYZ{X: 1, Y: 2, Z: 3})
	got, _ := Marshal(p)
	want := `{"type":"Point","coordinates":[1,2,3]}`
	assert.Equal(t, want, string(got))

	back, err := Unmarshal(got)
	require.NoError(t, err)
	assert.Equal(t, geom.LayoutXYZ, back.Layout(), "layout")
}

func TestFeatureRoundTrip(t *testing.T) {
	src := `{"type":"Feature","id":42,"bbox":[0,0,1,1],"geometry":{"type":"Point","coordinates":[0.5,0.5]},"properties":{"name":"origin"}}`
	var f Feature
	require.NoError(t, json.Unmarshal([]byte(src), &f))
	assert.Equal(t, geom.PointType, f.Geometry.Type(), "expected Point")
	assert.Equal(t, "origin", f.Properties["name"], "properties")
	out, err := json.Marshal(&f)
	require.NoError(t, err)
	// Round trip through json.Unmarshal again to compare semantically.
	var f2 Feature
	require.NoErrorf(t, json.Unmarshal(out, &f2), "re-decode\ndata: %s", out)
	assert.Equal(t, "origin", f2.Properties["name"], "after round-trip")
}

func TestFeatureCollectionForeign(t *testing.T) {
	src := `{"type":"FeatureCollection","features":[],"title":"test","attribution":"me"}`
	var fc FeatureCollection
	require.NoError(t, json.Unmarshal([]byte(src), &fc))
	assert.Equal(t, 2, len(fc.Foreign), "Foreign len")
	out, _ := json.Marshal(&fc)
	assert.Truef(t, strings.Contains(string(out), `"title":"test"`), "foreign 'title' lost: %s", out)
}

func TestUnmarshalEmptyPoint(t *testing.T) {
	// Empty arrays decode to POINT EMPTY for cross-format compat.
	g, err := Unmarshal([]byte(`{"type":"Point","coordinates":[]}`))
	require.NoError(t, err)
	assert.True(t, g.IsEmpty(), "empty-array point should be empty")
}

func TestUnmarshalErrors(t *testing.T) {
	cases := []string{
		`{"type":"Foo"}`,
		`{"type":"Point"}`,
		`{"type":"Point","coordinates":"bad"}`,
		`not json`,
	}
	for _, c := range cases {
		_, err := Unmarshal([]byte(c))
		assert.Errorf(t, err, "expected error for %q", c)
	}
}
