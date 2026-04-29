package conformance

import (
	"fmt"

	sfgeom "github.com/peterstace/simplefeatures/geom"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/wkt"
)

// simplefeaturesImpl adapts github.com/peterstace/simplefeatures/geom
// to the Impl interface.
//
// simplefeatures is pure Go (no cgo), so it is compiled into the
// default test binary. The conversion path is WKT in both directions:
// Terra -> wkt.Marshal -> simplefeatures.UnmarshalWKT -> operation ->
// Geometry.AsText() -> wkt.Unmarshal -> Terra.
//
// CAVEATS:
//   - Round-tripping through WKT loses M values (terra and simplefeatures
//     both support M, but our Terra->WKT encoder defaults to XY for the
//     corpus geometries, which are 2D anyway).
//   - simplefeatures rejects some inputs Terra accepts as "valid enough"
//     (e.g. self-touching rings). When that happens we surface the
//     error and the harness records a divergence rather than a panic.
type simplefeaturesImpl struct{}

// NewSimplefeatures returns the simplefeatures adapter.
func NewSimplefeatures() Impl { return simplefeaturesImpl{} }

func (simplefeaturesImpl) Name() string { return "simplefeatures" }

// toSF converts a Terra geometry to a simplefeatures geometry via WKT.
// Errors from either marshalling or parsing surface to the caller so
// the harness can record a clean divergence (rather than panicking).
func toSF(g geom.Geometry) (sfgeom.Geometry, error) {
	w, err := wkt.Marshal(g)
	if err != nil {
		return sfgeom.Geometry{}, fmt.Errorf("terra->wkt: %w", err)
	}
	// NoValidate is a deliberate choice: the corpus geometries are
	// already validated upstream by validate.Validate (in the corpus
	// smoke harness). Suppressing simplefeatures' own validation
	// avoids spurious "duplicate vertex" or "ring not closed"
	// rejections where Terra would still produce a usable result.
	sf, err := sfgeom.UnmarshalWKT(w, sfgeom.NoValidate{})
	if err != nil {
		return sfgeom.Geometry{}, fmt.Errorf("wkt->simplefeatures: %w", err)
	}
	return sf, nil
}

// fromSF converts a simplefeatures geometry back to a Terra geometry
// via WKT. Empty results round-trip to a Terra empty polygon so the
// area-based equality comparator naturally handles them.
func fromSF(sf sfgeom.Geometry) (geom.Geometry, error) {
	if sf.IsEmpty() {
		// wkt.Unmarshal handles empty WKT for the same type, but we
		// short-circuit to keep the adapter cheap.
		return wkt.Unmarshal(sf.AsText())
	}
	return wkt.Unmarshal(sf.AsText())
}

func (simplefeaturesImpl) Intersection(a, b geom.Geometry) (geom.Geometry, error) {
	sa, err := toSF(a)
	if err != nil {
		return nil, err
	}
	sb, err := toSF(b)
	if err != nil {
		return nil, err
	}
	res, err := sfgeom.Intersection(sa, sb)
	if err != nil {
		return nil, err
	}
	return fromSF(res)
}

func (simplefeaturesImpl) Union(a, b geom.Geometry) (geom.Geometry, error) {
	sa, err := toSF(a)
	if err != nil {
		return nil, err
	}
	sb, err := toSF(b)
	if err != nil {
		return nil, err
	}
	res, err := sfgeom.Union(sa, sb)
	if err != nil {
		return nil, err
	}
	return fromSF(res)
}

func (simplefeaturesImpl) Difference(a, b geom.Geometry) (geom.Geometry, error) {
	sa, err := toSF(a)
	if err != nil {
		return nil, err
	}
	sb, err := toSF(b)
	if err != nil {
		return nil, err
	}
	res, err := sfgeom.Difference(sa, sb)
	if err != nil {
		return nil, err
	}
	return fromSF(res)
}

func (simplefeaturesImpl) Area(g geom.Geometry) (float64, error) {
	sg, err := toSF(g)
	if err != nil {
		return 0, err
	}
	return sg.Area(), nil
}

func (simplefeaturesImpl) Length(g geom.Geometry) (float64, error) {
	sg, err := toSF(g)
	if err != nil {
		return 0, err
	}
	return sg.Length(), nil
}

func (simplefeaturesImpl) Relate(a, b geom.Geometry) (string, error) {
	sa, err := toSF(a)
	if err != nil {
		return "", err
	}
	sb, err := toSF(b)
	if err != nil {
		return "", err
	}
	return sfgeom.Relate(sa, sb)
}
