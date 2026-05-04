package crs

// Kind classifies a CRS as geographic (lon/lat on the ellipsoid),
// projected (Cartesian X/Y in some unit, typically meters), or unspecified.
//
// Kind drives the default-kernel selection logic in the predicate and
// measure packages: a geographic-CRS Distance defaults to the geodesic
// kernel; a projected-CRS Distance defaults to the planar kernel.
type Kind uint8

const (
	UnknownKind Kind = iota
	Geographic
	Projected
)

// CRS identifies a coordinate reference system.
//
// In the common case (Authority+Code refers to a registered EPSG code),
// the WKT2 field is empty and the Kind is supplied either by the registry
// or — for ad-hoc CRSes — by the caller.
//
// Definition is optional: it carries the parameters needed by Transform
// to actually convert coordinates (datum, projection). CRSes used only
// for identity comparison can leave it nil.
type CRS struct {
	Authority  string
	Code       int
	WKT2       string
	Kind       Kind
	Definition *Definition
}

// Equal reports whether two CRSes refer to the same coordinate reference
// system. Identity is by (Authority, Code) when both have authority codes;
// otherwise structural over the WKT2 string. Two nil pointers compare equal.
func Equal(a, b *CRS) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.Authority != "" && b.Authority != "" && a.Code != 0 && b.Code != 0 {
		return a.Authority == b.Authority && a.Code == b.Code
	}
	return a.WKT2 != "" && a.WKT2 == b.WKT2
}

// IsGeographic reports whether c is known to be a geographic CRS.
// nil and unknown-kind CRSes return false.
func (c *CRS) IsGeographic() bool { return c != nil && c.Kind == Geographic }

// IsProjected reports whether c is known to be a projected CRS.
func (c *CRS) IsProjected() bool { return c != nil && c.Kind == Projected }

// Pre-defined CRSes. These are placeholders sized for v0.1: the
// crs/epsg subpackage will hold the full registry.
var (
	WGS84 = &CRS{
		Authority: "EPSG", Code: 4326, Kind: Geographic,
	}
	WebMercator = &CRS{
		Authority: "EPSG", Code: 3857, Kind: Projected,
	}
	NAD83 = &CRS{
		Authority: "EPSG", Code: 4269, Kind: Geographic,
	}
)
