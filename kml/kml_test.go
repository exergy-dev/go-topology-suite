package kml

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/terra-geo/terra/geom"
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
			t.Fatalf("xml decode failed: %v\nfragment:\n%s", err, s)
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
	if err != nil {
		t.Fatal(err)
	}
	xmlDecodes(t, got)
	if !hasElement(got, "Point") || !hasElement(got, "coordinates") {
		t.Fatalf("missing elements:\n%s", got)
	}
	if !strings.Contains(got, "1,2") {
		t.Fatalf("expected x,y tuple in: %s", got)
	}
}

func TestMarshal_PointWithZOverride(t *testing.T) {
	p := geom.NewPoint(nil, geom.XY{X: 1, Y: 2})
	got, err := Marshal(p, WithZ(99))
	if err != nil {
		t.Fatal(err)
	}
	xmlDecodes(t, got)
	if !strings.Contains(got, "1,2,99") {
		t.Fatalf("expected x,y,z in: %s", got)
	}
}

func TestMarshal_LineString(t *testing.T) {
	ls := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 2, Y: 0}})
	got, err := Marshal(ls)
	if err != nil {
		t.Fatal(err)
	}
	xmlDecodes(t, got)
	if !hasElement(got, "LineString") || !hasElement(got, "coordinates") {
		t.Fatalf("missing elements:\n%s", got)
	}
}

func TestMarshal_PolygonWithHole(t *testing.T) {
	outer := []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0}}
	hole := []geom.XY{{X: 2, Y: 2}, {X: 4, Y: 2}, {X: 4, Y: 4}, {X: 2, Y: 4}, {X: 2, Y: 2}}
	p := geom.NewPolygon(nil, outer, hole)
	got, err := Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	xmlDecodes(t, got)
	if !hasElement(got, "Polygon") {
		t.Fatalf("missing Polygon: %s", got)
	}
	if !hasElement(got, "outerBoundaryIs") || !hasElement(got, "innerBoundaryIs") {
		t.Fatalf("missing boundaries:\n%s", got)
	}
	// Two LinearRing tags (outer + 1 hole)
	if c := strings.Count(got, "<LinearRing>"); c != 2 {
		t.Fatalf("expected 2 LinearRing tags, got %d:\n%s", c, got)
	}
}

func TestMarshal_MultiPolygon(t *testing.T) {
	a := geom.NewPolygon(nil, []geom.XY{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 0}})
	b := geom.NewPolygon(nil, []geom.XY{{X: 5, Y: 5}, {X: 6, Y: 5}, {X: 6, Y: 6}, {X: 5, Y: 5}})
	mp := geom.NewMultiPolygon(nil, a, b)
	got, err := Marshal(mp)
	if err != nil {
		t.Fatal(err)
	}
	xmlDecodes(t, got)
	if !hasElement(got, "MultiGeometry") {
		t.Fatalf("missing MultiGeometry:\n%s", got)
	}
	if c := strings.Count(got, "<Polygon>"); c != 2 {
		t.Fatalf("expected 2 Polygon tags, got %d:\n%s", c, got)
	}
}

func TestMarshal_Modifiers(t *testing.T) {
	p := geom.NewPoint(nil, geom.XY{X: 1, Y: 2})
	got, err := Marshal(p,
		WithExtrude(true),
		WithTesselate(true),
		WithAltitudeMode(AltitudeModeAbsolute))
	if err != nil {
		t.Fatal(err)
	}
	xmlDecodes(t, got)
	for _, want := range []string{"<extrude>1</extrude>", "<tesselate>1</tesselate>",
		"<altitudeMode>absolute</altitudeMode>"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in:\n%s", want, got)
		}
	}
}

func TestMarshal_LinePrefix(t *testing.T) {
	p := geom.NewPoint(nil, geom.XY{X: 1, Y: 2})
	got, err := Marshal(p, WithLinePrefix(">>"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, ">><Point>") {
		t.Fatalf("expected prefix on first line:\n%s", got)
	}
}

func TestMarshal_Precision(t *testing.T) {
	p := geom.NewPoint(nil, geom.XY{X: 1.0 / 3, Y: 2.0 / 3})
	got, err := Marshal(p, WithPrecision(3))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "0.333,0.667") {
		t.Fatalf("expected fixed-precision in: %s", got)
	}
}

func TestMarshal_NilGeometry(t *testing.T) {
	if _, err := Marshal(nil); err == nil {
		t.Fatalf("expected error for nil geometry")
	}
}
