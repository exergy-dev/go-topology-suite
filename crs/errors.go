package crs

import "errors"

var (
	// ErrIncompatibleUnits is returned when trying to convert between incompatible units.
	ErrIncompatibleUnits = errors.New("cannot convert between linear and angular units")

	// ErrSRIDMismatch is returned when operations are attempted between geometries
	// with incompatible coordinate reference systems.
	ErrSRIDMismatch = errors.New("SRID mismatch between geometries")
)
