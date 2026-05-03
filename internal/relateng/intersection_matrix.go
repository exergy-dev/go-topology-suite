package relateng

// IntersectionMatrix is a 3x3 DE-9IM matrix indexed by Location codes.
// Indices follow JTS convention: row = geometry-A location, column =
// geometry-B location, with EXTERIOR=0, BOUNDARY=1, INTERIOR=2.
//
// Cell values use JTS Dimension semantics:
//   -1 (DimFalse / "F"): empty intersection
//    0 / 1 / 2 (DimP / DimL / DimA): non-empty intersection of given dim
//   -2 (DimDontCare): pattern wildcard, only used for parsing patterns
//   -3 (DimTrue): pattern "T", only used for parsing patterns
//
// The runtime matrix only ever contains -1..2.
type IntersectionMatrix struct {
	cells [3][3]int
}

// Pattern wildcard sentinels (only valid in pattern matrices).
const (
	DimDontCare = -2
	DimTrue     = -3
)

// NewIntersectionMatrix returns an empty matrix (all DimFalse).
func NewIntersectionMatrix() *IntersectionMatrix {
	im := &IntersectionMatrix{}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			im.cells[i][j] = DimFalse
		}
	}
	return im
}

// patternOrder maps DE-9IM string positions [0..8] to (row, col)
// pairs of the internal matrix. Patterns are written in OGC order
// (II IB IE BI BB BE EI EB EE) but the matrix indexes by Location
// constants (E=0, B=1, I=2). This table bridges the two.
var patternOrder = [9][2]int{
	{LocInterior, LocInterior}, // II
	{LocInterior, LocBoundary}, // IB
	{LocInterior, LocExterior}, // IE
	{LocBoundary, LocInterior}, // BI
	{LocBoundary, LocBoundary}, // BB
	{LocBoundary, LocExterior}, // BE
	{LocExterior, LocInterior}, // EI
	{LocExterior, LocBoundary}, // EB
	{LocExterior, LocExterior}, // EE
}

// NewPatternMatrix parses a 9-char DE-9IM pattern (e.g.
// "T*F**FFF*") into a matrix populated with DimTrue / DimDontCare /
// DimFalse / 0..2 cell values. Returns nil for malformed input.
//
// Pattern characters are written in OGC order
// (II IB IE BI BB BE EI EB EE); the resulting cells are stored under
// the matrix's Location-indexed layout (E=0, B=1, I=2).
func NewPatternMatrix(pattern string) *IntersectionMatrix {
	if len(pattern) != 9 {
		return nil
	}
	im := &IntersectionMatrix{}
	for k := 0; k < 9; k++ {
		i := patternOrder[k][0]
		j := patternOrder[k][1]
		c := pattern[k]
		switch c {
		case '*':
			im.cells[i][j] = DimDontCare
		case 'T':
			im.cells[i][j] = DimTrue
		case 'F':
			im.cells[i][j] = DimFalse
		case '0':
			im.cells[i][j] = DimP
		case '1':
			im.cells[i][j] = DimL
		case '2':
			im.cells[i][j] = DimA
		default:
			return nil
		}
	}
	return im
}

// Get returns the cell at (locA, locB).
func (im *IntersectionMatrix) Get(locA, locB int) int {
	return im.cells[locA][locB]
}

// Set assigns the cell at (locA, locB).
func (im *IntersectionMatrix) Set(locA, locB int, dim int) {
	im.cells[locA][locB] = dim
}

// SetAtLeast updates the cell only if `dim` exceeds the current
// value, preserving the strictly-monotonic-build behaviour the JTS
// TopologyComputer assumes.
func (im *IntersectionMatrix) SetAtLeast(locA, locB int, dim int) {
	if dim > im.cells[locA][locB] {
		im.cells[locA][locB] = dim
	}
}

// Matches reports whether im satisfies the supplied DE-9IM pattern.
// Pattern characters: '*' wildcard, 'T' any non-F (>=0), 'F' must be
// F, '0'/'1'/'2' exact match.
//
// Patterns are interpreted in OGC order (II IB IE BI BB BE EI EB EE);
// see patternOrder.
func (im *IntersectionMatrix) Matches(pattern string) bool {
	if len(pattern) != 9 {
		return false
	}
	for k := 0; k < 9; k++ {
		i := patternOrder[k][0]
		j := patternOrder[k][1]
		c := im.cells[i][j]
		p := pattern[k]
		switch p {
		case '*':
			continue
		case 'T':
			if c < DimP {
				return false
			}
		case 'F':
			if c != DimFalse {
				return false
			}
		case '0', '1', '2':
			if c != int(p-'0') {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// String renders the matrix as a 9-char DE-9IM string.
//
// The string is in OGC order (II IB IE BI BB BE EI EB EE), the same
// layout as predicate.DE9IM.
func (im *IntersectionMatrix) String() string {
	out := make([]byte, 9)
	for k := 0; k < 9; k++ {
		i := patternOrder[k][0]
		j := patternOrder[k][1]
		switch v := im.cells[i][j]; v {
		case DimFalse:
			out[k] = 'F'
		case DimDontCare:
			out[k] = '*'
		case DimTrue:
			out[k] = 'T'
		default:
			if v >= 0 && v <= 2 {
				out[k] = byte('0' + v)
			} else {
				out[k] = '?'
			}
		}
	}
	return string(out)
}
