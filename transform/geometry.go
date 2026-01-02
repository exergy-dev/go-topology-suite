package transform

import (
	"fmt"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

type transformFilter struct {
	t   Transform
	err error
}

func (f *transformFilter) Filter(coord *geom.Coordinate) {
	if f.err != nil {
		return
	}
	updated, err := TransformCoordinate(f.t, *coord)
	if err != nil {
		f.err = err
		return
	}
	*coord = updated
}

// TransformGeometry applies a transformation to all coordinates in a geometry.
// Returns a new transformed geometry with the same type and structure as the input.
// Handles all geometry types including collections.
func TransformGeometry(t Transform, g geom.Geometry) (geom.Geometry, error) {
	if g == nil || g.IsEmpty() {
		return g, nil
	}

	clone := g.Clone()
	filter := &transformFilter{t: t}
	if cf, ok := clone.(geom.CoordinateFilterer); ok {
		cf.ApplyCoordinateFilter(filter)
	} else {
		return nil, fmt.Errorf("unsupported geometry type: %T", g)
	}
	if filter.err != nil {
		return nil, fmt.Errorf("transforming %T: %w", g, filter.err)
	}
	return clone, nil
}
