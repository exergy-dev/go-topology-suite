package epsg_test

import (
	"testing"

	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/crs/epsg"
)

// namedCase is one row in the table-driven sanity test of every named
// EPSG code this package exposes.
type namedCase struct {
	name string
	v    *crs.CRS
	code int
	kind crs.Kind
}

func namedCases() []namedCase {
	return []namedCase{
		{"WGS84", epsg.WGS84, 4326, crs.Geographic},
		{"NAD83", epsg.NAD83, 4269, crs.Geographic},
		{"NAD27", epsg.NAD27, 4267, crs.Geographic},
		{"WGS72", epsg.WGS72, 4322, crs.Geographic},
		{"ETRS89", epsg.ETRS89, 4258, crs.Geographic},
		{"WGS84_3D", epsg.WGS84_3D, 4979, crs.Geographic},
		{"CGCS2000", epsg.CGCS2000, 4490, crs.Geographic},
		{"Beijing1954", epsg.Beijing1954, 4214, crs.Geographic},
		{"WebMercator", epsg.WebMercator, 3857, crs.Projected},
		{"Lambert93", epsg.Lambert93, 2154, crs.Projected},
		{"BritishNationalGrid", epsg.BritishNationalGrid, 27700, crs.Projected},
		{"ConusAlbers", epsg.ConusAlbers, 5070, crs.Projected},
		{"EuropeLAEA", epsg.EuropeLAEA, 3035, crs.Projected},
	}
}

func TestNamedLookups(t *testing.T) {
	for _, tc := range namedCases() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.v == nil {
				t.Fatalf("named var %s is nil", tc.name)
			}
			if tc.v.Authority != "EPSG" {
				t.Errorf("Authority = %q, want EPSG", tc.v.Authority)
			}
			if tc.v.Code != tc.code {
				t.Errorf("Code = %d, want %d", tc.v.Code, tc.code)
			}
			if tc.v.Kind != tc.kind {
				t.Errorf("Kind = %v, want %v", tc.v.Kind, tc.kind)
			}
			got := epsg.Lookup(tc.code)
			if got == nil {
				t.Fatalf("Lookup(%d) returned nil", tc.code)
			}
			if got != tc.v {
				t.Errorf("Lookup(%d) returned a different pointer than the named var", tc.code)
			}
		})
	}
}

func TestLookupUnknown(t *testing.T) {
	if got := epsg.Lookup(99999); got != nil {
		t.Errorf("Lookup(99999) = %+v, want nil", got)
	}
	if got := epsg.Lookup(0); got != nil {
		t.Errorf("Lookup(0) = %+v, want nil", got)
	}
	if got := epsg.Lookup(-1); got != nil {
		t.Errorf("Lookup(-1) = %+v, want nil", got)
	}
}

func TestWGS84EqualsCRSWGS84(t *testing.T) {
	got := epsg.Lookup(4326)
	if got == nil {
		t.Fatal("Lookup(4326) = nil")
	}
	if !crs.Equal(got, crs.WGS84) {
		t.Errorf("Lookup(4326) not Equal to crs.WGS84: %+v vs %+v", got, crs.WGS84)
	}
	// And the cross-package WebMercator/NAD83 comparisons should also hold,
	// since they share authority+code with the upstream crs vars.
	if !crs.Equal(epsg.WebMercator, crs.WebMercator) {
		t.Errorf("epsg.WebMercator not Equal to crs.WebMercator")
	}
	if !crs.Equal(epsg.NAD83, crs.NAD83) {
		t.Errorf("epsg.NAD83 not Equal to crs.NAD83")
	}
}

func TestUTMZoneCoverage(t *testing.T) {
	// Spot-check the four programmatic ranges. Every code must resolve to
	// a Projected CRS with Authority=EPSG.
	ranges := []struct {
		first, last int
	}{
		{32601, 32660},
		{32701, 32760},
		{26901, 26923},
		{25832, 25835},
	}
	for _, r := range ranges {
		for code := r.first; code <= r.last; code++ {
			c := epsg.Lookup(code)
			if c == nil {
				t.Errorf("Lookup(%d) = nil, want non-nil", code)
				continue
			}
			if c.Authority != "EPSG" || c.Code != code {
				t.Errorf("Lookup(%d) = %+v, want Authority=EPSG Code=%d", code, c, code)
			}
			if c.Kind != crs.Projected {
				t.Errorf("Lookup(%d).Kind = %v, want Projected", code, c.Kind)
			}
		}
	}

	// Sanity bounds: nothing immediately outside each range was registered
	// as a side-effect.
	for _, code := range []int{32600, 32661, 32700, 32761, 26900, 26924, 25831, 25836} {
		if got := epsg.Lookup(code); got != nil {
			t.Errorf("Lookup(%d) = %+v, want nil (outside registered range)", code, got)
		}
	}
}

func TestCodesReturnsAllRegistered(t *testing.T) {
	codes := epsg.Codes()
	// 8 named geographic + 5 named projected + 60 + 60 + 23 + 4 UTM = 160.
	const want = 8 + 5 + 60 + 60 + 23 + 4
	if len(codes) != want {
		t.Errorf("Codes() returned %d entries, want %d", len(codes), want)
	}
	// Verify ordering.
	for i := 1; i < len(codes); i++ {
		if codes[i-1] >= codes[i] {
			t.Errorf("Codes() not strictly sorted at index %d: %d >= %d", i, codes[i-1], codes[i])
			break
		}
	}
}
