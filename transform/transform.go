// Package transform provides coordinate transformation functionality
// for geometric objects. It includes support for affine transformations,
// map projections, and composite transformations.
package transform

import (
	"fmt"

	"github.com/go-topology-suite/gts/geom"
)

// Transform defines an interface for coordinate transformations.
// Transformations can convert coordinates between different coordinate
// reference systems (CRS) or apply geometric transformations.
type Transform interface {
	// Forward transforms from source to target CRS.
	// Returns the transformed x and y coordinates, or an error if the
	// transformation fails.
	Forward(x, y float64) (float64, float64, error)

	// Inverse transforms from target to source CRS.
	// Returns the inverse-transformed x and y coordinates, or an error
	// if the transformation fails.
	Inverse(x, y float64) (float64, float64, error)
}

// Identity represents a pass-through transformation that returns
// coordinates unchanged. Useful as a null object or for testing.
type Identity struct{}

// NewIdentity creates a new identity transformation.
func NewIdentity() *Identity {
	return &Identity{}
}

// Forward returns the input coordinates unchanged.
func (i *Identity) Forward(x, y float64) (float64, float64, error) {
	return x, y, nil
}

// Inverse returns the input coordinates unchanged.
func (i *Identity) Inverse(x, y float64) (float64, float64, error) {
	return x, y, nil
}

// InverseTransform wraps a transform to swap its forward and inverse operations.
// This allows you to use a transform in the reverse direction without
// implementing a separate reverse transform.
type InverseTransform struct {
	T Transform
}

// NewInverse creates a new inverse transform that swaps forward and inverse
// operations of the provided transform.
func NewInverse(t Transform) *InverseTransform {
	return &InverseTransform{T: t}
}

// Forward applies the inverse transformation of the wrapped transform.
func (it *InverseTransform) Forward(x, y float64) (float64, float64, error) {
	return it.T.Inverse(x, y)
}

// Inverse applies the forward transformation of the wrapped transform.
func (it *InverseTransform) Inverse(x, y float64) (float64, float64, error) {
	return it.T.Forward(x, y)
}

// Composite chains multiple transforms together, applying them in sequence.
// The forward transformation applies transforms in order (first to last).
// The inverse transformation applies transforms in reverse order (last to first).
type Composite struct {
	Transforms []Transform
}

// NewComposite creates a new composite transformation from the given transforms.
// Transforms are applied in the order they appear in the slice.
func NewComposite(transforms ...Transform) *Composite {
	return &Composite{Transforms: transforms}
}

// Forward applies all transforms in sequence (first to last).
func (c *Composite) Forward(x, y float64) (float64, float64, error) {
	var err error
	for i, t := range c.Transforms {
		x, y, err = t.Forward(x, y)
		if err != nil {
			return 0, 0, fmt.Errorf("transform %d forward failed: %w", i, err)
		}
	}
	return x, y, nil
}

// Inverse applies all transforms in reverse sequence (last to first).
func (c *Composite) Inverse(x, y float64) (float64, float64, error) {
	var err error
	for i := len(c.Transforms) - 1; i >= 0; i-- {
		x, y, err = c.Transforms[i].Inverse(x, y)
		if err != nil {
			return 0, 0, fmt.Errorf("transform %d inverse failed: %w", i, err)
		}
	}
	return x, y, nil
}

// TransformCoordinate applies a transformation to a single coordinate.
// Returns a new transformed coordinate.
func TransformCoordinate(t Transform, coord geom.Coordinate) (geom.Coordinate, error) {
	x, y, err := t.Forward(coord.X, coord.Y)
	if err != nil {
		return geom.Coordinate{}, err
	}

	result := geom.NewCoordinate(x, y)

	// Copy Z and M values if present (these are not transformed)
	if coord.Z != nil {
		z := *coord.Z
		result.Z = &z
	}
	if coord.M != nil {
		m := *coord.M
		result.M = &m
	}

	return result, nil
}

// TransformCoordinates applies a transformation to a sequence of coordinates.
// Returns a new transformed coordinate sequence.
func TransformCoordinates(t Transform, coords geom.CoordinateSequence) (geom.CoordinateSequence, error) {
	if len(coords) == 0 {
		return geom.CoordinateSequence{}, nil
	}

	result := make(geom.CoordinateSequence, len(coords))
	for i, coord := range coords {
		transformed, err := TransformCoordinate(t, coord)
		if err != nil {
			return nil, fmt.Errorf("coordinate %d: %w", i, err)
		}
		result[i] = transformed
	}

	return result, nil
}
