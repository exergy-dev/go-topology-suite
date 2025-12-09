// Package relate provides DE-9IM spatial relationship computation.
//
// The Dimensionally Extended 9-Intersection Model (DE-9IM) describes
// spatial relationships between two geometries using a 3x3 matrix that
// records the dimensions of intersections between the Interior, Boundary,
// and Exterior of each geometry.
package relate

import (
	"fmt"
	"strings"
)

// Dimension represents the dimension of a geometric intersection.
type Dimension int

const (
	// DimFalse indicates no intersection (empty set).
	DimFalse Dimension = -1
	// DimPoint indicates a 0-dimensional intersection (point).
	DimPoint Dimension = 0
	// DimLine indicates a 1-dimensional intersection (line).
	DimLine Dimension = 1
	// DimArea indicates a 2-dimensional intersection (area).
	DimArea Dimension = 2
	// DimDontCare indicates any dimension is acceptable (for pattern matching).
	DimDontCare Dimension = -2
	// DimTrue indicates any non-empty dimension (for pattern matching).
	DimTrue Dimension = -3
)

// String returns the string representation of a dimension.
func (d Dimension) String() string {
	switch d {
	case DimFalse:
		return "F"
	case DimPoint:
		return "0"
	case DimLine:
		return "1"
	case DimArea:
		return "2"
	case DimDontCare:
		return "*"
	case DimTrue:
		return "T"
	default:
		return "?"
	}
}

// Location indices for the matrix.
const (
	Interior = 0
	Boundary = 1
	Exterior = 2
)

// IntersectionMatrix represents a DE-9IM matrix.
// The matrix is indexed as [row][col] where:
//   - row 0-2: Interior, Boundary, Exterior of geometry A
//   - col 0-2: Interior, Boundary, Exterior of geometry B
type IntersectionMatrix [3][3]Dimension

// NewIntersectionMatrix creates a new matrix with all values set to DimFalse.
func NewIntersectionMatrix() *IntersectionMatrix {
	return &IntersectionMatrix{
		{DimFalse, DimFalse, DimFalse},
		{DimFalse, DimFalse, DimFalse},
		{DimFalse, DimFalse, DimFalse},
	}
}

// NewIntersectionMatrixFromString parses a 9-character DE-9IM string.
func NewIntersectionMatrixFromString(s string) (*IntersectionMatrix, error) {
	if len(s) != 9 {
		return nil, fmt.Errorf("DE-9IM string must be 9 characters, got %d", len(s))
	}

	m := NewIntersectionMatrix()
	for i, c := range strings.ToUpper(s) {
		row := i / 3
		col := i % 3
		dim, err := parseDimension(byte(c))
		if err != nil {
			return nil, err
		}
		m[row][col] = dim
	}
	return m, nil
}

func parseDimension(c byte) (Dimension, error) {
	switch c {
	case 'F', 'f':
		return DimFalse, nil
	case '0':
		return DimPoint, nil
	case '1':
		return DimLine, nil
	case '2':
		return DimArea, nil
	case 'T', 't':
		return DimTrue, nil
	case '*':
		return DimDontCare, nil
	default:
		return DimFalse, fmt.Errorf("invalid dimension character: %c", c)
	}
}

// Get returns the dimension at the specified location.
func (m *IntersectionMatrix) Get(locA, locB int) Dimension {
	return m[locA][locB]
}

// Set sets the dimension at the specified location.
func (m *IntersectionMatrix) Set(locA, locB int, dim Dimension) {
	m[locA][locB] = dim
}

// SetAtLeast sets the dimension at the specified location to at least the given value.
func (m *IntersectionMatrix) SetAtLeast(locA, locB int, dim Dimension) {
	if dim > m[locA][locB] {
		m[locA][locB] = dim
	}
}

// SetAtLeastIfValid sets the dimension if both locations are valid (not -1).
func (m *IntersectionMatrix) SetAtLeastIfValid(locA, locB int, dim Dimension) {
	if locA >= 0 && locB >= 0 {
		m.SetAtLeast(locA, locB, dim)
	}
}

// String returns the 9-character DE-9IM string representation.
func (m *IntersectionMatrix) String() string {
	var sb strings.Builder
	for row := 0; row < 3; row++ {
		for col := 0; col < 3; col++ {
			sb.WriteString(m[row][col].String())
		}
	}
	return sb.String()
}

// Matches tests if this matrix matches the given pattern string.
// Pattern characters:
//   - T: matches any non-empty dimension (0, 1, 2)
//   - F: matches empty dimension (-1)
//   - *: matches any dimension
//   - 0, 1, 2: matches that specific dimension
func (m *IntersectionMatrix) Matches(pattern string) bool {
	if len(pattern) != 9 {
		return false
	}

	pattern = strings.ToUpper(pattern)
	for i, c := range pattern {
		row := i / 3
		col := i % 3
		if !matchesDimension(m[row][col], byte(c)) {
			return false
		}
	}
	return true
}

func matchesDimension(actual Dimension, pattern byte) bool {
	switch pattern {
	case '*':
		return true
	case 'T':
		return actual >= DimPoint
	case 'F':
		return actual == DimFalse
	case '0':
		return actual == DimPoint
	case '1':
		return actual == DimLine
	case '2':
		return actual == DimArea
	default:
		return false
	}
}

// IsDisjoint returns true if the geometries are disjoint.
func (m *IntersectionMatrix) IsDisjoint() bool {
	return m[Interior][Interior] == DimFalse &&
		m[Interior][Boundary] == DimFalse &&
		m[Boundary][Interior] == DimFalse &&
		m[Boundary][Boundary] == DimFalse
}

// IsIntersects returns true if the geometries intersect.
func (m *IntersectionMatrix) IsIntersects() bool {
	return !m.IsDisjoint()
}

// IsTouches returns true if the geometries touch.
// They touch if they have at least one point in common,
// but their interiors do not intersect.
func (m *IntersectionMatrix) IsTouches(dimA, dimB int) bool {
	if dimA > dimB {
		// Swap for consistent checking
		return m.Transpose().IsTouches(dimB, dimA)
	}

	// Interior/Interior must be empty
	if m[Interior][Interior] != DimFalse {
		return false
	}

	// Must have at least one point in common
	return m[Interior][Boundary] >= DimPoint ||
		m[Boundary][Interior] >= DimPoint ||
		m[Boundary][Boundary] >= DimPoint
}

// IsCrosses returns true if the geometries cross.
func (m *IntersectionMatrix) IsCrosses(dimA, dimB int) bool {
	// Point/Point or Area/Area cannot cross
	if (dimA == 0 && dimB == 0) || (dimA == 2 && dimB == 2) {
		return false
	}

	// Line/Line
	if dimA == 1 && dimB == 1 {
		return m[Interior][Interior] == DimPoint
	}

	// Line/Area or Point/Line
	if dimA < dimB {
		return m[Interior][Interior] >= DimPoint && m[Interior][Exterior] >= DimPoint
	}

	// Area/Line or Line/Point
	return m[Interior][Interior] >= DimPoint && m[Exterior][Interior] >= DimPoint
}

// IsWithin returns true if geometry A is within geometry B.
func (m *IntersectionMatrix) IsWithin() bool {
	return m[Interior][Interior] >= DimPoint &&
		m[Interior][Exterior] == DimFalse &&
		m[Boundary][Exterior] == DimFalse
}

// IsContains returns true if geometry A contains geometry B.
func (m *IntersectionMatrix) IsContains() bool {
	return m[Interior][Interior] >= DimPoint &&
		m[Exterior][Interior] == DimFalse &&
		m[Exterior][Boundary] == DimFalse
}

// IsOverlaps returns true if the geometries overlap.
func (m *IntersectionMatrix) IsOverlaps(dimA, dimB int) bool {
	// Must be same dimension, and must be point or area
	if dimA != dimB {
		// Line/Area case
		if (dimA == 1 && dimB == 2) || (dimA == 2 && dimB == 1) {
			return m[Interior][Interior] == DimLine &&
				m[Interior][Exterior] >= DimPoint &&
				m[Exterior][Interior] >= DimPoint
		}
		return false
	}

	// Point/Point or Area/Area
	if dimA == 0 || dimA == 2 {
		return m[Interior][Interior] >= DimPoint &&
			m[Interior][Exterior] >= DimPoint &&
			m[Exterior][Interior] >= DimPoint
	}

	// Line/Line
	if dimA == 1 {
		return m[Interior][Interior] == DimLine &&
			m[Interior][Exterior] >= DimPoint &&
			m[Exterior][Interior] >= DimPoint
	}

	return false
}

// IsEquals returns true if the geometries are topologically equal.
func (m *IntersectionMatrix) IsEquals(dimA, dimB int) bool {
	if dimA != dimB {
		return false
	}
	return m[Interior][Interior] >= DimPoint &&
		m[Interior][Exterior] == DimFalse &&
		m[Boundary][Exterior] == DimFalse &&
		m[Exterior][Interior] == DimFalse &&
		m[Exterior][Boundary] == DimFalse
}

// IsCovers returns true if geometry A covers geometry B.
func (m *IntersectionMatrix) IsCovers() bool {
	return (m[Interior][Interior] >= DimPoint ||
		m[Interior][Boundary] >= DimPoint ||
		m[Boundary][Interior] >= DimPoint ||
		m[Boundary][Boundary] >= DimPoint) &&
		m[Exterior][Interior] == DimFalse &&
		m[Exterior][Boundary] == DimFalse
}

// IsCoveredBy returns true if geometry A is covered by geometry B.
func (m *IntersectionMatrix) IsCoveredBy() bool {
	return (m[Interior][Interior] >= DimPoint ||
		m[Interior][Boundary] >= DimPoint ||
		m[Boundary][Interior] >= DimPoint ||
		m[Boundary][Boundary] >= DimPoint) &&
		m[Interior][Exterior] == DimFalse &&
		m[Boundary][Exterior] == DimFalse
}

// Transpose returns a new matrix with rows and columns swapped.
// This effectively swaps geometry A and B.
func (m *IntersectionMatrix) Transpose() *IntersectionMatrix {
	t := NewIntersectionMatrix()
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			t[i][j] = m[j][i]
		}
	}
	return t
}
