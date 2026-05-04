package conformance

import (
	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/measure"
	"github.com/exergy-dev/go-topology-suite/overlay"
	"github.com/exergy-dev/go-topology-suite/predicate"
)

// gtsImpl adapts the github.com/exergy-dev/go-topology-suite packages to the Impl
// interface. It is the "reference" implementation against which every
// other Impl is compared.
type gtsImpl struct{}

// NewGTS returns the go-topology-suite adapter. The harness uses it as
// the reference Impl: its outputs define the expected behaviour and
// other Impls are recorded as agreeing or disagreeing with it.
func NewGTS() Impl { return gtsImpl{} }

func (gtsImpl) Name() string { return "gts" }

func (gtsImpl) Intersection(a, b geom.Geometry) (geom.Geometry, error) {
	return overlay.Intersection(a, b)
}

func (gtsImpl) Union(a, b geom.Geometry) (geom.Geometry, error) {
	return overlay.Union(a, b)
}

func (gtsImpl) Difference(a, b geom.Geometry) (geom.Geometry, error) {
	return overlay.Difference(a, b)
}

func (gtsImpl) Area(g geom.Geometry) (float64, error) {
	return measure.Area(g), nil
}

func (gtsImpl) Length(g geom.Geometry) (float64, error) {
	return measure.Length(g), nil
}

func (gtsImpl) Relate(a, b geom.Geometry) (string, error) {
	im, err := predicate.Relate(a, b)
	if err != nil {
		return "", err
	}
	return string(im), nil
}
