// Package terra is the top-level facade for the Terra geospatial library.
//
// Most users will import the format and operation subpackages directly:
//
//	import (
//	    "github.com/terra-geo/terra/geom"
//	    "github.com/terra-geo/terra/geojson"
//	    "github.com/terra-geo/terra/predicate"
//	)
//
// The terra package itself only re-exports the sentinel errors and a small
// number of convenience constructors.
package terra

import "errors"

var (
	// ErrEmpty is returned when an operation is undefined on an empty geometry.
	ErrEmpty = errors.New("terra: operation undefined on empty geometry")

	// ErrCRSMismatch is returned when two operands have differing CRS.
	// Callers must transform explicitly via crs.Transform.
	ErrCRSMismatch = errors.New("terra: operands have different CRS")

	// ErrUnsupportedKernel is returned when the requested kernel does not
	// implement the requested operation (e.g. spherical Buffer in v1).
	ErrUnsupportedKernel = errors.New("terra: kernel does not support operation")

	// ErrInvalidGeometry is returned when an input geometry violates an
	// invariant required by the operation (self-intersection, unclosed ring,
	// etc.). Use validate.Validate for detailed defect reports.
	ErrInvalidGeometry = errors.New("terra: invalid geometry")
)
