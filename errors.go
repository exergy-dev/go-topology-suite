// Package gts is the top-level facade for the go-topology-suite geospatial library.
//
// Most users will import the format and operation subpackages directly:
//
//	import (
//	    "github.com/exergy-dev/go-topology-suite/geom"
//	    "github.com/exergy-dev/go-topology-suite/geojson"
//	    "github.com/exergy-dev/go-topology-suite/predicate"
//	)
//
// The gts package itself only re-exports the sentinel errors and a small
// number of convenience constructors.
package gts

import "errors"

var (
	// ErrEmpty is returned when an operation is undefined on an empty geometry.
	ErrEmpty = errors.New("gts: operation undefined on empty geometry")

	// ErrCRSMismatch is returned when two operands have differing CRS.
	// Callers must transform explicitly via gts.Transform.
	ErrCRSMismatch = errors.New("gts: operands have different CRS")

	// ErrUnsupportedKernel is returned when the requested kernel does not
	// implement the requested operation (e.g. spherical Buffer in v1).
	ErrUnsupportedKernel = errors.New("gts: kernel does not support operation")

	// ErrInvalidGeometry is returned when an input geometry violates an
	// invariant required by the operation (self-intersection, unclosed ring,
	// etc.). Use validate.Validate for detailed defect reports.
	ErrInvalidGeometry = errors.New("gts: invalid geometry")
)
