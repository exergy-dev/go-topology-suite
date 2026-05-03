package relateng

// Port of org.locationtech.jts.operation.relateng.DimensionLocation.
//
// DimLoc combines the topological dimension of an element (Point /
// Line / Area) with a coarse Location (Interior / Boundary / Exterior).
// This lets the locator return a single int that tells the caller both
// "where on the parent geometry the test point lies" and "what is the
// dimension of that element".
//
// JTS encodes the values as carefully chosen integers so that
// `dimension(dimLoc) >= dimension(otherDimLoc)` matches a sensible
// "this element overrides the other in mixed-dim collections" ordering.
// We mirror those literal integer values exactly so that the encoding
// is interchangeable with JTS test expectations.

// Topological dimension constants (mirror JTS Dimension).
const (
	DimFalse = -1 // empty
	DimP     = 0
	DimL     = 1
	DimA     = 2
)

// Coarse location constants (mirror JTS Location).
const (
	LocExterior int = 0
	LocBoundary int = 1
	LocInterior int = 2
)

// DimLoc combines an element dimension with a location. The integer
// values match JTS DimensionLocation exactly.
const (
	DLExterior     = LocExterior // 0
	DLPointInterior = 103
	DLLineInterior  = 110
	DLLineBoundary  = 111
	DLAreaInterior  = 120
	DLAreaBoundary  = 121
)

// LocationArea wraps a Location into the area dim/loc encoding.
func LocationArea(loc int) int {
	switch loc {
	case LocInterior:
		return DLAreaInterior
	case LocBoundary:
		return DLAreaBoundary
	}
	return DLExterior
}

// LocationLine wraps a Location into the line dim/loc encoding.
func LocationLine(loc int) int {
	switch loc {
	case LocInterior:
		return DLLineInterior
	case LocBoundary:
		return DLLineBoundary
	}
	return DLExterior
}

// LocationPoint wraps a Location into the point dim/loc encoding.
// Points only have an interior, so anything other than INTERIOR
// collapses to EXTERIOR.
func LocationPoint(loc int) int {
	if loc == LocInterior {
		return DLPointInterior
	}
	return DLExterior
}

// Location decodes a DimLoc into its coarse Location component.
func Location(dimLoc int) int {
	switch dimLoc {
	case DLPointInterior, DLLineInterior, DLAreaInterior:
		return LocInterior
	case DLLineBoundary, DLAreaBoundary:
		return LocBoundary
	}
	return LocExterior
}

// Dimension decodes a DimLoc into its element dimension.
func Dimension(dimLoc int) int {
	switch dimLoc {
	case DLPointInterior:
		return DimP
	case DLLineInterior, DLLineBoundary:
		return DimL
	case DLAreaInterior, DLAreaBoundary:
		return DimA
	}
	return DimFalse
}

// DimensionExt returns the dimension of dimLoc, substituting
// `exteriorDim` when dimLoc is DLExterior. Used by the topology
// computer to record the dimension of the exterior cell when only
// one geometry is being inspected.
func DimensionExt(dimLoc, exteriorDim int) int {
	if dimLoc == DLExterior {
		return exteriorDim
	}
	return Dimension(dimLoc)
}
