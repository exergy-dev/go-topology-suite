// Package precision provides models for controlling the precision
// of geometric coordinates. This package re-exports types from geom
// for backward compatibility.
package precision

import (
	"github.com/go-topology-suite/gts/geom"
)

// Type aliases for backward compatibility
type (
	PrecisionModel = geom.PrecisionModel
	PrecisionType  = geom.PrecisionType
)

// Constants re-exported from geom
const (
	FloatingPrecision       = geom.FloatingPrecision
	FloatingSinglePrecision = geom.FloatingSinglePrecision
	FixedPrecision          = geom.FixedPrecision
)

// Constructor functions
var (
	NewFloatingPrecision       = geom.NewFloatingPrecision
	NewFloatingSinglePrecision = geom.NewFloatingSinglePrecision
	NewFixedPrecision          = geom.NewFixedPrecision
)

// Common precision models
var (
	Floating       = geom.Floating
	FloatingSingle = geom.FloatingSingle
	Fixed1         = geom.Fixed1
	Fixed2         = geom.Fixed2
	Fixed3         = geom.Fixed3
	Fixed6         = geom.Fixed6
	Fixed8         = geom.Fixed8
)

// MakePreciseSequence applies a precision model to a coordinate sequence.
func MakePreciseSequence(pm PrecisionModel, coords geom.CoordinateSequence) {
	geom.MakePreciseSequence(pm, coords)
}

// Compare compares two precision models.
func Compare(pm1, pm2 PrecisionModel) int {
	return geom.ComparePrecision(pm1, pm2)
}

// MostPrecise returns the more precise of two precision models.
func MostPrecise(pm1, pm2 PrecisionModel) PrecisionModel {
	return geom.MostPrecise(pm1, pm2)
}

// LeastPrecise returns the less precise of two precision models.
func LeastPrecise(pm1, pm2 PrecisionModel) PrecisionModel {
	return geom.LeastPrecise(pm1, pm2)
}
