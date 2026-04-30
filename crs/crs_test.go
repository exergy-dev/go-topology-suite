package crs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEqualByAuthorityCode(t *testing.T) {
	a := &CRS{Authority: "EPSG", Code: 4326, Kind: Geographic}
	b := &CRS{Authority: "EPSG", Code: 4326}
	assert.True(t, Equal(a, b), "matching EPSG codes should be Equal regardless of kind")
	c := &CRS{Authority: "EPSG", Code: 3857}
	assert.False(t, Equal(a, c), "different codes should not be Equal")
}

func TestEqualNilHandling(t *testing.T) {
	assert.True(t, Equal(nil, nil), "two nil CRSes should be Equal")
	assert.False(t, Equal(nil, WGS84), "nil and non-nil should not be Equal")
}

func TestEqualWKT2Fallback(t *testing.T) {
	wkt := `GEOGCRS["custom",...]`
	a := &CRS{WKT2: wkt}
	b := &CRS{WKT2: wkt}
	assert.True(t, Equal(a, b), "matching WKT2 should be Equal")
	c := &CRS{WKT2: `GEOGCRS["other",...]`}
	assert.False(t, Equal(a, c), "differing WKT2 should not be Equal")
}

func TestKindHelpers(t *testing.T) {
	assert.True(t, WGS84.IsGeographic(), "WGS84 should be geographic")
	assert.False(t, WGS84.IsProjected(), "WGS84 should not be projected")
	assert.True(t, WebMercator.IsProjected(), "WebMercator should be projected")
	var nilCRS *CRS
	assert.False(t, nilCRS.IsGeographic() || nilCRS.IsProjected(), "nil CRS should be neither geographic nor projected")
}
