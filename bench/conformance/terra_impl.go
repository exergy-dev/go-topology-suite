package conformance

import (
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/measure"
	"github.com/terra-geo/terra/overlay"
	"github.com/terra-geo/terra/predicate"
)

// terraImpl adapts the github.com/terra-geo/terra packages to the Impl
// interface. It is the "reference" implementation against which every
// other Impl is compared.
type terraImpl struct{}

// NewTerra returns the Terra adapter. The harness uses Terra as the
// reference Impl: its outputs define the expected behaviour and other
// Impls are recorded as agreeing or disagreeing with it.
func NewTerra() Impl { return terraImpl{} }

func (terraImpl) Name() string { return "terra" }

func (terraImpl) Intersection(a, b geom.Geometry) (geom.Geometry, error) {
	return overlay.Intersection(a, b)
}

func (terraImpl) Union(a, b geom.Geometry) (geom.Geometry, error) {
	return overlay.Union(a, b)
}

func (terraImpl) Difference(a, b geom.Geometry) (geom.Geometry, error) {
	return overlay.Difference(a, b)
}

func (terraImpl) Area(g geom.Geometry) (float64, error) {
	return measure.Area(g), nil
}

func (terraImpl) Length(g geom.Geometry) (float64, error) {
	return measure.Length(g), nil
}

func (terraImpl) Relate(a, b geom.Geometry) (string, error) {
	im, err := predicate.Relate(a, b)
	if err != nil {
		return "", err
	}
	return string(im), nil
}
