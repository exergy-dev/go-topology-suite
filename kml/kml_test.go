package kml

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// xmlDecodes confirms the fragment parses without error. KML fragments
// are not whole documents, so we wrap them in a synthetic <Root> element.
func xmlDecodes(t *testing.T, s string) {
	t.Helper()
	dec := xml.NewDecoder(strings.NewReader("<Root>" + s + "</Root>"))
	for {
		_, err := dec.Token()
		if err != nil {
			if err.Error() == "EOF" {
				return
			}
			require.NoErrorf(t, err, "xml decode failed\nfragment:\n%s", s)
		}
	}
}

// hasElementWithBody reports whether the fragment contains a <name> ...
// </name> pair (non-empty body verified by the trailing close-tag test).
func hasElement(s, name string) bool {
	return strings.Contains(s, "<"+name+">") && strings.Contains(s, "</"+name+">")
}

func TestMarshal_Point(t *testing.T) {
	p := geom.NewPoint(nil, geom.XY{X: 1, Y: 2})
	got, err := Marshal(p)
	require.NoError(t, err)
	xmlDecodes(t, got)
	require.Truef(t, hasElement(got, "Point") && hasElement(got, "coordinates"), "missing elements:\n%s", got)
	require.Truef(t, strings.Contains(got, "1,2"), "expected x,y tuple in: %s", got)
}

func TestMarshal_PointWithZOverride(t *testing.T) {
	p := geom.NewPoint(nil, geom.XY{X: 1, Y: 2})
	got, err := Marshal(p, WithZ(99))
	require.NoError(t, err)
	xmlDecodes(t, got)
	require.Truef(t, strings.Contains(got, "1,2,99"), "expected x,y,z in: %s", got)
}

func TestMarshal_LineString(t *testing.T) {
	ls := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 2, Y: 0}})
	got, err := Marshal(ls)
	require.NoError(t, err)
	xmlDecodes(t, got)
	require.Truef(t, hasElement(got, "LineString") && hasElement(got, "coordinates"), "missing elements:\n%s", got)
}

func TestMarshal_PolygonWithHole(t *testing.T) {
	outer := []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0}}
	hole := []geom.XY{{X: 2, Y: 2}, {X: 4, Y: 2}, {X: 4, Y: 4}, {X: 2, Y: 4}, {X: 2, Y: 2}}
	p := geom.NewPolygon(nil, outer, hole)
	got, err := Marshal(p)
	require.NoError(t, err)
	xmlDecodes(t, got)
	require.Truef(t, hasElement(got, "Polygon"), "missing Polygon: %s", got)
	require.Truef(t, hasElement(got, "outerBoundaryIs") && hasElement(got, "innerBoundaryIs"), "missing boundaries:\n%s", got)
	// Two LinearRing tags (outer + 1 hole)
	assert.Equalf(t, 2, strings.Count(got, "<LinearRing>"), "expected 2 LinearRing tags:\n%s", got)
}

func TestMarshal_MultiPolygon(t *testing.T) {
	a := geom.NewPolygon(nil, []geom.XY{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 0}})
	b := geom.NewPolygon(nil, []geom.XY{{X: 5, Y: 5}, {X: 6, Y: 5}, {X: 6, Y: 6}, {X: 5, Y: 5}})
	mp := geom.NewMultiPolygon(nil, a, b)
	got, err := Marshal(mp)
	require.NoError(t, err)
	xmlDecodes(t, got)
	require.Truef(t, hasElement(got, "MultiGeometry"), "missing MultiGeometry:\n%s", got)
	assert.Equalf(t, 2, strings.Count(got, "<Polygon>"), "expected 2 Polygon tags:\n%s", got)
}

func TestMarshal_Modifiers(t *testing.T) {
	p := geom.NewPoint(nil, geom.XY{X: 1, Y: 2})
	got, err := Marshal(p,
		WithExtrude(true),
		WithTesselate(true),
		WithAltitudeMode(AltitudeModeAbsolute))
	require.NoError(t, err)
	xmlDecodes(t, got)
	for _, want := range []string{"<extrude>1</extrude>", "<tesselate>1</tesselate>",
		"<altitudeMode>absolute</altitudeMode>"} {
		require.Truef(t, strings.Contains(got, want), "expected %q in:\n%s", want, got)
	}
}

func TestMarshal_LinePrefix(t *testing.T) {
	p := geom.NewPoint(nil, geom.XY{X: 1, Y: 2})
	got, err := Marshal(p, WithLinePrefix(">>"))
	require.NoError(t, err)
	require.Truef(t, strings.HasPrefix(got, ">><Point>"), "expected prefix on first line:\n%s", got)
}

func TestMarshal_Precision(t *testing.T) {
	p := geom.NewPoint(nil, geom.XY{X: 1.0 / 3, Y: 2.0 / 3})
	got, err := Marshal(p, WithPrecision(3))
	require.NoError(t, err)
	require.Truef(t, strings.Contains(got, "0.333,0.667"), "expected fixed-precision in: %s", got)
}

func TestMarshal_NilGeometry(t *testing.T) {
	_, err := Marshal(nil)
	require.Error(t, err)
}
