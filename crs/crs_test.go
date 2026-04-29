package crs

import "testing"

func TestEqualByAuthorityCode(t *testing.T) {
	a := &CRS{Authority: "EPSG", Code: 4326, Kind: Geographic}
	b := &CRS{Authority: "EPSG", Code: 4326}
	if !Equal(a, b) {
		t.Errorf("matching EPSG codes should be Equal regardless of kind")
	}
	c := &CRS{Authority: "EPSG", Code: 3857}
	if Equal(a, c) {
		t.Errorf("different codes should not be Equal")
	}
}

func TestEqualNilHandling(t *testing.T) {
	if !Equal(nil, nil) {
		t.Errorf("two nil CRSes should be Equal")
	}
	if Equal(nil, WGS84) {
		t.Errorf("nil and non-nil should not be Equal")
	}
}

func TestEqualWKT2Fallback(t *testing.T) {
	wkt := `GEOGCRS["custom",...]`
	a := &CRS{WKT2: wkt}
	b := &CRS{WKT2: wkt}
	if !Equal(a, b) {
		t.Errorf("matching WKT2 should be Equal")
	}
	c := &CRS{WKT2: `GEOGCRS["other",...]`}
	if Equal(a, c) {
		t.Errorf("differing WKT2 should not be Equal")
	}
}

func TestKindHelpers(t *testing.T) {
	if !WGS84.IsGeographic() {
		t.Errorf("WGS84 should be geographic")
	}
	if WGS84.IsProjected() {
		t.Errorf("WGS84 should not be projected")
	}
	if !WebMercator.IsProjected() {
		t.Errorf("WebMercator should be projected")
	}
	var nilCRS *CRS
	if nilCRS.IsGeographic() || nilCRS.IsProjected() {
		t.Errorf("nil CRS should be neither geographic nor projected")
	}
}
