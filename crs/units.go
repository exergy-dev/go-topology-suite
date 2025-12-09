package crs

import "math"

// Unit represents a unit of measurement for coordinates.
type Unit struct {
	// Name is the unit name (e.g., "metre", "degree").
	Name string

	// Type indicates whether this is a linear or angular unit.
	Type UnitType

	// ToMetres is the conversion factor to metres (for linear units).
	// For angular units, this is set to 1.0.
	toMetres float64

	// ToDegrees is the conversion factor to degrees (for angular units).
	// For linear units, this is set to 1.0.
	toDegrees float64
}

// UnitType indicates whether a unit is linear (distance) or angular (angle).
type UnitType int

const (
	// Linear indicates a linear (distance) unit.
	Linear UnitType = iota
	// Angular indicates an angular unit.
	Angular
)

// String returns the string representation of a UnitType.
func (t UnitType) String() string {
	switch t {
	case Linear:
		return "Linear"
	case Angular:
		return "Angular"
	default:
		return "Unknown"
	}
}

// ToMetres returns the conversion factor from this unit to metres.
// For angular units, returns 1.0 (no meaningful conversion).
func (u Unit) ToMetres() float64 {
	return u.toMetres
}

// ToDegrees returns the conversion factor from this unit to degrees.
// For linear units, returns 1.0 (no meaningful conversion).
func (u Unit) ToDegrees() float64 {
	return u.toDegrees
}

// IsLinear returns true if this is a linear (distance) unit.
func (u Unit) IsLinear() bool {
	return u.Type == Linear
}

// IsAngular returns true if this is an angular unit.
func (u Unit) IsAngular() bool {
	return u.Type == Angular
}

// Common angular units.
var (
	// Degree is the unit for degrees of arc.
	Degree = Unit{
		Name:       "degree",
		Type:       Angular,
		toMetres:   1.0,
		toDegrees:  1.0,
	}

	// Radian is the unit for radians.
	Radian = Unit{
		Name:       "radian",
		Type:       Angular,
		toMetres:   1.0,
		toDegrees:  180.0 / math.Pi,
	}

	// Gradian is the unit for gradians (400 gradians = 360 degrees).
	Gradian = Unit{
		Name:       "gradian",
		Type:       Angular,
		toMetres:   1.0,
		toDegrees:  0.9,
	}
)

// Common linear units.
var (
	// Metre is the SI base unit for length.
	Metre = Unit{
		Name:       "metre",
		Type:       Linear,
		toMetres:   1.0,
		toDegrees:  1.0,
	}

	// Kilometre is 1000 metres.
	Kilometre = Unit{
		Name:       "kilometre",
		Type:       Linear,
		toMetres:   1000.0,
		toDegrees:  1.0,
	}

	// Foot is the international foot (0.3048 metres exactly).
	Foot = Unit{
		Name:       "foot",
		Type:       Linear,
		toMetres:   0.3048,
		toDegrees:  1.0,
	}

	// USSurveyFoot is the US survey foot (1200/3937 metres).
	USSurveyFoot = Unit{
		Name:       "US survey foot",
		Type:       Linear,
		toMetres:   1200.0 / 3937.0,
		toDegrees:  1.0,
	}

	// Mile is the international mile (1609.344 metres).
	Mile = Unit{
		Name:       "mile",
		Type:       Linear,
		toMetres:   1609.344,
		toDegrees:  1.0,
	}

	// NauticalMile is the international nautical mile (1852 metres).
	NauticalMile = Unit{
		Name:       "nautical mile",
		Type:       Linear,
		toMetres:   1852.0,
		toDegrees:  1.0,
	}
)

// ConvertValue converts a value from one unit to another.
// Returns an error if the units are not compatible (e.g., converting
// between linear and angular units).
func ConvertValue(value float64, from, to Unit) (float64, error) {
	if from.Type != to.Type {
		return 0, ErrIncompatibleUnits
	}

	if from.Type == Linear {
		// Convert: from -> metres -> to
		return value * from.toMetres / to.toMetres, nil
	}

	// Angular units
	// Convert: from -> degrees -> to
	return value * from.toDegrees / to.toDegrees, nil
}
