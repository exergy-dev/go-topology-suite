// Package valid provides geometry validation according to OGC Simple Features.
package valid

import (
	"fmt"
	"math"

	"github.com/go-topology-suite/gts/algorithm"
	"github.com/go-topology-suite/gts/geom"
)

// ValidationErrorType describes the type of validation error.
type ValidationErrorType int

const (
	// ErrNone indicates no error.
	ErrNone ValidationErrorType = iota
	// ErrInvalidCoordinate indicates a coordinate has NaN or Inf values.
	ErrInvalidCoordinate
	// ErrTooFewPoints indicates not enough points for the geometry type.
	ErrTooFewPoints
	// ErrRingNotClosed indicates a ring is not closed.
	ErrRingNotClosed
	// ErrSelfIntersection indicates a ring has a self-intersection.
	ErrSelfIntersection
	// ErrDuplicateRings indicates duplicate rings in a polygon.
	ErrDuplicateRings
	// ErrNestedHoles indicates a hole contains another hole.
	ErrNestedHoles
	// ErrHoleOutsideShell indicates a hole is outside the shell.
	ErrHoleOutsideShell
	// ErrDisconnectedInterior indicates the interior is disconnected.
	ErrDisconnectedInterior
	// ErrNestedShells indicates shells of a MultiPolygon are nested.
	ErrNestedShells
	// ErrInvalidOrientation indicates incorrect ring orientation.
	ErrInvalidOrientation
)

// String returns a human-readable description of the error type.
func (e ValidationErrorType) String() string {
	switch e {
	case ErrNone:
		return "no error"
	case ErrInvalidCoordinate:
		return "invalid coordinate (NaN or Inf)"
	case ErrTooFewPoints:
		return "too few points"
	case ErrRingNotClosed:
		return "ring not closed"
	case ErrSelfIntersection:
		return "self-intersection"
	case ErrDuplicateRings:
		return "duplicate rings"
	case ErrNestedHoles:
		return "nested holes"
	case ErrHoleOutsideShell:
		return "hole outside shell"
	case ErrDisconnectedInterior:
		return "disconnected interior"
	case ErrNestedShells:
		return "nested shells in MultiPolygon"
	case ErrInvalidOrientation:
		return "invalid ring orientation"
	default:
		return "unknown error"
	}
}

// ValidationError represents a geometry validation error.
type ValidationError struct {
	// Type is the type of validation error.
	Type ValidationErrorType
	// Location is the coordinate where the error was detected (if applicable).
	Location *geom.Coordinate
	// Message provides additional details about the error.
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.Location != nil {
		return fmt.Sprintf("%s at (%g, %g): %s", e.Type.String(), e.Location.X, e.Location.Y, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Type.String(), e.Message)
}

// ValidationResult holds the result of geometry validation.
type ValidationResult struct {
	// IsValid is true if the geometry is valid.
	IsValid bool
	// Errors contains all validation errors found.
	Errors []*ValidationError
}

// Error returns the first error message if invalid, empty string if valid.
func (r *ValidationResult) Error() string {
	if r.IsValid || len(r.Errors) == 0 {
		return ""
	}
	return r.Errors[0].Error()
}

// Validate validates a geometry and returns a detailed result.
func Validate(g geom.Geometry) *ValidationResult {
	if g == nil {
		return &ValidationResult{IsValid: true}
	}

	switch v := g.(type) {
	case *geom.Point:
		return validatePoint(v)
	case *geom.LineString:
		return validateLineString(v)
	case *geom.LinearRing:
		return validateLinearRing(v)
	case *geom.Polygon:
		return validatePolygon(v)
	case *geom.MultiPoint:
		return validateMultiPoint(v)
	case *geom.MultiLineString:
		return validateMultiLineString(v)
	case *geom.MultiPolygon:
		return validateMultiPolygon(v)
	case *geom.GeometryCollection:
		return validateGeometryCollection(v)
	default:
		return &ValidationResult{IsValid: true}
	}
}

// IsValid returns true if the geometry is valid.
func IsValid(g geom.Geometry) bool {
	return Validate(g).IsValid
}

func validatePoint(p *geom.Point) *ValidationResult {
	result := &ValidationResult{IsValid: true}

	if p.IsEmpty() {
		return result
	}

	coord := p.Coordinate()
	if err := validateCoordinate(coord); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, err)
	}

	return result
}

func validateCoordinate(c geom.Coordinate) *ValidationError {
	if math.IsNaN(c.X) || math.IsInf(c.X, 0) {
		return &ValidationError{
			Type:     ErrInvalidCoordinate,
			Location: &c,
			Message:  fmt.Sprintf("X coordinate is invalid: %v", c.X),
		}
	}
	if math.IsNaN(c.Y) || math.IsInf(c.Y, 0) {
		return &ValidationError{
			Type:     ErrInvalidCoordinate,
			Location: &c,
			Message:  fmt.Sprintf("Y coordinate is invalid: %v", c.Y),
		}
	}
	return nil
}

func validateCoordinates(coords geom.CoordinateSequence) *ValidationError {
	for _, c := range coords {
		if err := validateCoordinate(c); err != nil {
			return err
		}
	}
	return nil
}

func validateLineString(ls *geom.LineString) *ValidationResult {
	result := &ValidationResult{IsValid: true}

	if ls.IsEmpty() {
		return result
	}

	coords := ls.Coordinates()

	// Check for invalid coordinates
	if err := validateCoordinates(coords); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, err)
		return result
	}

	// LineString must have 0 or >= 2 points
	if len(coords) == 1 {
		result.IsValid = false
		result.Errors = append(result.Errors, &ValidationError{
			Type:     ErrTooFewPoints,
			Location: &coords[0],
			Message:  "LineString must have 0 or at least 2 points",
		})
	}

	return result
}

func validateLinearRing(lr *geom.LinearRing) *ValidationResult {
	result := &ValidationResult{IsValid: true}

	if lr.IsEmpty() {
		return result
	}

	coords := lr.Coordinates()

	// Check for invalid coordinates
	if err := validateCoordinates(coords); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, err)
		return result
	}

	// Ring must have at least 4 points (including closure)
	if len(coords) < 4 {
		result.IsValid = false
		result.Errors = append(result.Errors, &ValidationError{
			Type:    ErrTooFewPoints,
			Message: fmt.Sprintf("LinearRing must have at least 4 points, has %d", len(coords)),
		})
		return result
	}

	// Ring must be closed
	if !coords.IsClosed(geom.DefaultEpsilon) {
		result.IsValid = false
		result.Errors = append(result.Errors, &ValidationError{
			Type:     ErrRingNotClosed,
			Location: &coords[0],
			Message:  "ring is not closed",
		})
		return result
	}

	// Check for self-intersection
	if err := checkRingSelfIntersection(coords); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, err)
	}

	return result
}

func checkRingSelfIntersection(coords geom.CoordinateSequence) *ValidationError {
	n := len(coords)
	if n < 4 {
		return nil
	}

	// Check all non-adjacent segment pairs
	for i := 0; i < n-1; i++ {
		for j := i + 2; j < n-1; j++ {
			// Don't check first and last segments (they share closure point)
			if i == 0 && j == n-2 {
				continue
			}

			p1 := coords[i]
			p2 := coords[i+1]
			p3 := coords[j]
			p4 := coords[j+1]

			result := algorithm.LineIntersection(p1, p2, p3, p4)
			if result.HasIntersection && result.IsProper {
				return &ValidationError{
					Type:     ErrSelfIntersection,
					Location: &result.Intersection,
					Message:  fmt.Sprintf("ring self-intersects at segment %d and %d", i, j),
				}
			}
		}
	}

	return nil
}

func validatePolygon(p *geom.Polygon) *ValidationResult {
	result := &ValidationResult{IsValid: true}

	if p.IsEmpty() {
		return result
	}

	shell := p.ExteriorRing()

	// Validate shell
	shellResult := validateLinearRing(shell)
	if !shellResult.IsValid {
		result.IsValid = false
		result.Errors = append(result.Errors, shellResult.Errors...)
		return result
	}

	// Shell must be counter-clockwise (positive signed area)
	if !shell.IsCCW() {
		result.IsValid = false
		coords := shell.Coordinates()
		result.Errors = append(result.Errors, &ValidationError{
			Type:     ErrInvalidOrientation,
			Location: &coords[0],
			Message:  "exterior ring must be counter-clockwise",
		})
	}

	// Validate holes
	for i := 0; i < p.NumInteriorRings(); i++ {
		hole := p.InteriorRingN(i)
		holeResult := validateLinearRing(hole)
		if !holeResult.IsValid {
			result.IsValid = false
			result.Errors = append(result.Errors, holeResult.Errors...)
			continue
		}

		// Holes must be clockwise (negative signed area)
		if !hole.IsCW() {
			result.IsValid = false
			coords := hole.Coordinates()
			result.Errors = append(result.Errors, &ValidationError{
				Type:     ErrInvalidOrientation,
				Location: &coords[0],
				Message:  fmt.Sprintf("interior ring %d must be clockwise", i),
			})
		}

		// Hole must be inside shell
		if !isRingInsideRing(hole, shell) {
			result.IsValid = false
			coords := hole.Coordinates()
			result.Errors = append(result.Errors, &ValidationError{
				Type:     ErrHoleOutsideShell,
				Location: &coords[0],
				Message:  fmt.Sprintf("hole %d is not inside shell", i),
			})
		}
	}

	// Check holes don't nest within each other
	for i := 0; i < p.NumInteriorRings(); i++ {
		for j := i + 1; j < p.NumInteriorRings(); j++ {
			hole1 := p.InteriorRingN(i)
			hole2 := p.InteriorRingN(j)

			if isRingInsideRing(hole1, hole2) || isRingInsideRing(hole2, hole1) {
				result.IsValid = false
				coords := hole1.Coordinates()
				result.Errors = append(result.Errors, &ValidationError{
					Type:     ErrNestedHoles,
					Location: &coords[0],
					Message:  fmt.Sprintf("holes %d and %d are nested", i, j),
				})
			}
		}
	}

	// Check shell and holes don't cross
	for i := 0; i < p.NumInteriorRings(); i++ {
		hole := p.InteriorRingN(i)
		if err := checkRingsCross(shell, hole); err != nil {
			result.IsValid = false
			err.Message = fmt.Sprintf("shell crosses hole %d", i)
			result.Errors = append(result.Errors, err)
		}
	}

	// Check holes don't cross each other
	for i := 0; i < p.NumInteriorRings(); i++ {
		for j := i + 1; j < p.NumInteriorRings(); j++ {
			hole1 := p.InteriorRingN(i)
			hole2 := p.InteriorRingN(j)
			if err := checkRingsCross(hole1, hole2); err != nil {
				result.IsValid = false
				err.Message = fmt.Sprintf("hole %d crosses hole %d", i, j)
				result.Errors = append(result.Errors, err)
			}
		}
	}

	return result
}

// isRingInsideRing checks if inner is inside outer.
func isRingInsideRing(inner, outer *geom.LinearRing) bool {
	// Get a point from inner ring that is not on the boundary of outer
	innerCoords := inner.Coordinates()
	if len(innerCoords) == 0 {
		return false
	}

	// Test if a point from inner is inside outer
	// Use the first point for simplicity
	testPoint := innerCoords[0]
	return isPointInRing(testPoint, outer)
}

// isPointInRing uses the ray casting algorithm.
func isPointInRing(p geom.Coordinate, ring *geom.LinearRing) bool {
	coords := ring.Coordinates()
	n := len(coords)
	if n < 4 {
		return false
	}

	inside := false
	j := n - 2 // Second to last point (since ring is closed)

	for i := 0; i < n-1; i++ {
		xi, yi := coords[i].X, coords[i].Y
		xj, yj := coords[j].X, coords[j].Y

		if ((yi > p.Y) != (yj > p.Y)) &&
			(p.X < (xj-xi)*(p.Y-yi)/(yj-yi)+xi) {
			inside = !inside
		}
		j = i
	}

	return inside
}

// checkRingsCross checks if two rings cross each other.
func checkRingsCross(r1, r2 *geom.LinearRing) *ValidationError {
	coords1 := r1.Coordinates()
	coords2 := r2.Coordinates()

	for i := 0; i < len(coords1)-1; i++ {
		for j := 0; j < len(coords2)-1; j++ {
			result := algorithm.LineIntersection(coords1[i], coords1[i+1], coords2[j], coords2[j+1])
			if result.HasIntersection && result.IsProper {
				return &ValidationError{
					Type:     ErrSelfIntersection,
					Location: &result.Intersection,
					Message:  "rings cross",
				}
			}
		}
	}
	return nil
}

func validateMultiPoint(mp *geom.MultiPoint) *ValidationResult {
	result := &ValidationResult{IsValid: true}

	for i := 0; i < mp.NumGeometries(); i++ {
		p := mp.GeometryN(i).(*geom.Point)
		pResult := validatePoint(p)
		if !pResult.IsValid {
			result.IsValid = false
			result.Errors = append(result.Errors, pResult.Errors...)
		}
	}

	return result
}

func validateMultiLineString(mls *geom.MultiLineString) *ValidationResult {
	result := &ValidationResult{IsValid: true}

	for i := 0; i < mls.NumGeometries(); i++ {
		ls := mls.GeometryN(i).(*geom.LineString)
		lsResult := validateLineString(ls)
		if !lsResult.IsValid {
			result.IsValid = false
			result.Errors = append(result.Errors, lsResult.Errors...)
		}
	}

	return result
}

func validateMultiPolygon(mp *geom.MultiPolygon) *ValidationResult {
	result := &ValidationResult{IsValid: true}

	// Validate each polygon
	polygons := make([]*geom.Polygon, mp.NumGeometries())
	for i := 0; i < mp.NumGeometries(); i++ {
		poly := mp.GeometryN(i).(*geom.Polygon)
		polygons[i] = poly

		polyResult := validatePolygon(poly)
		if !polyResult.IsValid {
			result.IsValid = false
			result.Errors = append(result.Errors, polyResult.Errors...)
		}
	}

	// Check that polygons don't overlap (shells don't intersect except at points)
	for i := 0; i < len(polygons); i++ {
		for j := i + 1; j < len(polygons); j++ {
			if polygons[i].IsEmpty() || polygons[j].IsEmpty() {
				continue
			}

			shell1 := polygons[i].ExteriorRing()
			shell2 := polygons[j].ExteriorRing()

			// Check if shells cross
			if err := checkRingsCross(shell1, shell2); err != nil {
				result.IsValid = false
				err.Message = fmt.Sprintf("polygon %d and %d shells cross", i, j)
				result.Errors = append(result.Errors, err)
			}

			// Check if one shell is inside another (nested shells not allowed)
			if isRingInsideRing(shell1, shell2) {
				result.IsValid = false
				coords := shell1.Coordinates()
				result.Errors = append(result.Errors, &ValidationError{
					Type:     ErrNestedShells,
					Location: &coords[0],
					Message:  fmt.Sprintf("polygon %d is nested inside polygon %d", i, j),
				})
			} else if isRingInsideRing(shell2, shell1) {
				result.IsValid = false
				coords := shell2.Coordinates()
				result.Errors = append(result.Errors, &ValidationError{
					Type:     ErrNestedShells,
					Location: &coords[0],
					Message:  fmt.Sprintf("polygon %d is nested inside polygon %d", j, i),
				})
			}
		}
	}

	return result
}

func validateGeometryCollection(gc *geom.GeometryCollection) *ValidationResult {
	result := &ValidationResult{IsValid: true}

	for i := 0; i < gc.NumGeometries(); i++ {
		g := gc.GeometryN(i)
		gResult := Validate(g)
		if !gResult.IsValid {
			result.IsValid = false
			result.Errors = append(result.Errors, gResult.Errors...)
		}
	}

	return result
}

// MakeValid attempts to repair an invalid geometry.
// Returns the repaired geometry and whether repair was needed.
func MakeValid(g geom.Geometry) (geom.Geometry, bool) {
	if g == nil || IsValid(g) {
		return g, false
	}

	switch v := g.(type) {
	case *geom.Polygon:
		return makePolygonValid(v)
	case *geom.LinearRing:
		return makeLinearRingValid(v)
	default:
		// For other types, just return as-is for now
		return g, false
	}
}

func makeLinearRingValid(lr *geom.LinearRing) (geom.Geometry, bool) {
	coords := lr.Coordinates()
	if len(coords) == 0 {
		return lr, false
	}

	repaired := false

	// Ensure closure
	if !coords.IsClosed(geom.DefaultEpsilon) {
		coords = append(coords, coords[0].Clone())
		repaired = true
	}

	// Ensure minimum 4 points
	if len(coords) < 4 {
		// Can't repair - not enough points
		return lr, false
	}

	if repaired {
		return geom.NewLinearRing(coords), true
	}
	return lr, false
}

func makePolygonValid(p *geom.Polygon) (geom.Geometry, bool) {
	if p.IsEmpty() {
		return p, false
	}

	repaired := false
	shell := p.ExteriorRing()

	// Fix shell orientation if needed
	if !shell.IsCCW() {
		shell = shell.Reverse()
		repaired = true
	}

	// Fix hole orientations
	holes := make([]*geom.LinearRing, p.NumInteriorRings())
	for i := 0; i < p.NumInteriorRings(); i++ {
		hole := p.InteriorRingN(i)
		if !hole.IsCW() {
			hole = hole.Reverse()
			repaired = true
		}
		holes[i] = hole
	}

	if repaired {
		return geom.NewPolygon(shell, holes), true
	}
	return p, false
}
