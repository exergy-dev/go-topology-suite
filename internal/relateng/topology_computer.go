package relateng

import "github.com/terra-geo/terra/geom"

// TopologyComputer is the orchestrator that drives a TopologyPredicate
// to a final answer by exploring the topological interaction between
// two RelateNG geometries. Port of
// org.locationtech.jts.operation.relateng.TopologyComputer.
//
// Coverage in this Go port:
//
//   - InitExteriorDims handles all dimension-driven exterior fills.
//   - AddPointOnPointInterior / AddPointOnPointExterior covers P/P.
//   - AddPointOnGeometry covers P/L and P/A from the point side.
//   - AddLineEndOnGeometry covers L/L and L/A line-end nodes.
//   - AddAreaVertex covers A/L and A/A area-vertex nodes.
//
// Not yet ported:
//
//   - The edge-segment intersector pipeline (EdgeSegmentIntersector,
//     EdgeSetIntersector, RelateNode, RelateEdge). When edges of A and
//     B cross between vertices, those interior intersection points
//     are not added. This is sufficient for the (large) class of
//     predicates whose answer is determined by point-locator results
//     on existing vertices alone — the higher-level RelateNG wrapper
//     in predicate/relateng.go will fall back to the legacy
//     predicate path for inputs whose answer depends on non-vertex
//     edge intersections.
//
// The exposed API mirrors JTS so the missing pieces can be slotted
// in later (TopologyComputer.AddIntersection, EvaluateNodes) without
// disturbing callers.
type TopologyComputer struct {
	predicate TopologyPredicate
	geomA     *Geometry
	geomB     *Geometry
	// nodeMap stores per-coordinate sections discovered by
	// AddIntersection, for later evaluation. Until the edge pipeline
	// lands the map is unused but kept so the field layout matches
	// JTS for forward compatibility.
	nodeMap map[geom.XY][]*NodeSection
}

// NewTopologyComputer constructs a TopologyComputer bound to the
// given predicate and operands. The exterior-dim seed is performed
// up front (mirroring JTS).
func NewTopologyComputer(p TopologyPredicate, a, b *Geometry) *TopologyComputer {
	tc := &TopologyComputer{
		predicate: p,
		geomA:     a,
		geomB:     b,
		nodeMap:   make(map[geom.XY][]*NodeSection),
	}
	tc.initExteriorDims()
	return tc
}

// Geometry returns the A or B operand as selected by isA.
func (tc *TopologyComputer) Geometry(isA bool) *Geometry {
	if isA {
		return tc.geomA
	}
	return tc.geomB
}

// GetDimension returns the dimension of the A or B operand.
func (tc *TopologyComputer) GetDimension(isA bool) int {
	return tc.Geometry(isA).Dimension()
}

// IsAreaArea reports whether both inputs have area dimension.
func (tc *TopologyComputer) IsAreaArea() bool {
	return tc.GetDimension(true) == DimA && tc.GetDimension(false) == DimA
}

// IsSelfNodingRequired indicates whether the inputs require self-
// noding for correct evaluation. Mirrors JTS exactly.
func (tc *TopologyComputer) IsSelfNodingRequired() bool {
	if !tc.predicate.RequireSelfNoding() {
		return false
	}
	if tc.geomA.IsSelfNodingRequired() {
		return true
	}
	if tc.geomB.HasAreaAndLine() {
		return true
	}
	return false
}

// IsExteriorCheckRequired delegates to the predicate.
func (tc *TopologyComputer) IsExteriorCheckRequired(isA bool) bool {
	return tc.predicate.RequireExteriorCheck(isA)
}

// IsResultKnown reports whether the bound predicate has resolved.
func (tc *TopologyComputer) IsResultKnown() bool { return tc.predicate.IsKnown() }

// Result returns the predicate's final boolean value.
func (tc *TopologyComputer) Result() bool { return tc.predicate.Value() }

// Finish drives the predicate's final settlement step.
func (tc *TopologyComputer) Finish() { tc.predicate.Finish() }

func (tc *TopologyComputer) updateDim(locA, locB, dim int) {
	tc.predicate.UpdateDimension(locA, locB, dim)
}

func (tc *TopologyComputer) updateDimAB(isAB bool, loc1, loc2, dim int) {
	if isAB {
		tc.updateDim(loc1, loc2, dim)
	} else {
		tc.updateDim(loc2, loc1, dim)
	}
}

// initExteriorDims seeds the matrix with cells that are determined
// purely by the input dimensions.
func (tc *TopologyComputer) initExteriorDims() {
	dimA := tc.geomA.DimensionReal()
	dimB := tc.geomB.DimensionReal()

	switch {
	case dimA == DimP && dimB == DimL:
		tc.updateDim(LocExterior, LocInterior, DimL)
	case dimA == DimL && dimB == DimP:
		tc.updateDim(LocInterior, LocExterior, DimL)
	case dimA == DimP && dimB == DimA:
		tc.updateDim(LocExterior, LocInterior, DimA)
		tc.updateDim(LocExterior, LocBoundary, DimL)
	case dimA == DimA && dimB == DimP:
		tc.updateDim(LocInterior, LocExterior, DimA)
		tc.updateDim(LocBoundary, LocExterior, DimL)
	case dimA == DimL && dimB == DimA:
		tc.updateDim(LocExterior, LocInterior, DimA)
	case dimA == DimA && dimB == DimL:
		tc.updateDim(LocInterior, LocExterior, DimA)
	case dimA == DimFalse || dimB == DimFalse:
		if dimA != DimFalse {
			tc.initExteriorEmpty(true)
		}
		if dimB != DimFalse {
			tc.initExteriorEmpty(false)
		}
	}
}

func (tc *TopologyComputer) initExteriorEmpty(isA bool) {
	dim := tc.GetDimension(isA)
	switch dim {
	case DimP:
		tc.updateDimAB(isA, LocInterior, LocExterior, DimP)
	case DimL:
		if tc.Geometry(isA).HasBoundary() {
			tc.updateDimAB(isA, LocBoundary, LocExterior, DimP)
		}
		tc.updateDimAB(isA, LocInterior, LocExterior, DimL)
	case DimA:
		tc.updateDimAB(isA, LocBoundary, LocExterior, DimL)
		tc.updateDimAB(isA, LocInterior, LocExterior, DimA)
	}
}

// AddPointOnPointInterior records a point/point coincidence at p.
func (tc *TopologyComputer) AddPointOnPointInterior(p geom.XY) {
	tc.updateDim(LocInterior, LocInterior, DimP)
}

// AddPointOnPointExterior records a point that is in the exterior
// of the other point set. isA indicates which input contributed the
// point.
func (tc *TopologyComputer) AddPointOnPointExterior(isA bool, p geom.XY) {
	tc.updateDimAB(isA, LocInterior, LocExterior, DimP)
}

// AddPointOnGeometry records a point of the isPointA input vs the
// other input at the given target dim/loc.
func (tc *TopologyComputer) AddPointOnGeometry(isPointA bool, locTarget, dimTarget int, p geom.XY) {
	tc.updateDimAB(isPointA, LocInterior, locTarget, DimP)
	if tc.Geometry(!isPointA).IsEmpty() {
		return
	}
	switch dimTarget {
	case DimP, DimL:
		return
	case DimA:
		tc.updateDimAB(isPointA, LocExterior, LocInterior, DimA)
		tc.updateDimAB(isPointA, LocExterior, LocBoundary, DimL)
	}
}

// AddLineEndOnGeometry records a line endpoint of the isLineA input.
func (tc *TopologyComputer) AddLineEndOnGeometry(isLineA bool, locLineEnd, locTarget, dimTarget int, p geom.XY) {
	tc.updateDimAB(isLineA, locLineEnd, locTarget, DimP)
	if tc.Geometry(!isLineA).IsEmpty() {
		return
	}
	switch dimTarget {
	case DimP:
		return
	case DimL:
		if locTarget == LocExterior {
			tc.updateDimAB(isLineA, LocInterior, LocExterior, DimL)
		}
	case DimA:
		if locTarget != LocBoundary {
			tc.updateDimAB(isLineA, LocInterior, locTarget, DimL)
			tc.updateDimAB(isLineA, LocExterior, locTarget, DimA)
		}
	}
}

// AddAreaVertex records an area vertex interaction with a target
// element.
func (tc *TopologyComputer) AddAreaVertex(isAreaA bool, locArea, locTarget, dimTarget int, p geom.XY) {
	if locTarget == LocExterior {
		tc.updateDimAB(isAreaA, LocInterior, LocExterior, DimA)
		if locArea == LocBoundary {
			tc.updateDimAB(isAreaA, LocBoundary, LocExterior, DimL)
			tc.updateDimAB(isAreaA, LocExterior, LocExterior, DimA)
		}
		return
	}
	switch dimTarget {
	case DimP:
		tc.addAreaVertexOnPoint(isAreaA, locArea, p)
	case DimL:
		tc.addAreaVertexOnLine(isAreaA, locArea, locTarget, p)
	case DimA:
		tc.addAreaVertexOnArea(isAreaA, locArea, locTarget, p)
	}
}

func (tc *TopologyComputer) addAreaVertexOnPoint(isAreaA bool, locArea int, p geom.XY) {
	tc.updateDimAB(isAreaA, locArea, LocInterior, DimP)
	tc.updateDimAB(isAreaA, LocInterior, LocExterior, DimA)
	if locArea == LocBoundary {
		tc.updateDimAB(isAreaA, LocBoundary, LocExterior, DimL)
		tc.updateDimAB(isAreaA, LocExterior, LocExterior, DimA)
	}
}

func (tc *TopologyComputer) addAreaVertexOnLine(isAreaA bool, locArea, locTarget int, p geom.XY) {
	tc.updateDimAB(isAreaA, locArea, locTarget, DimP)
	if locArea == LocInterior {
		tc.updateDimAB(isAreaA, LocInterior, LocExterior, DimA)
	}
}

func (tc *TopologyComputer) addAreaVertexOnArea(isAreaA bool, locArea, locTarget int, p geom.XY) {
	if locTarget == LocBoundary {
		if locArea == LocBoundary {
			tc.updateDimAB(isAreaA, LocBoundary, LocBoundary, DimP)
		} else {
			tc.updateDimAB(isAreaA, LocInterior, LocInterior, DimA)
			tc.updateDimAB(isAreaA, LocInterior, LocBoundary, DimL)
			tc.updateDimAB(isAreaA, LocInterior, LocExterior, DimA)
		}
	} else {
		tc.updateDimAB(isAreaA, LocInterior, locTarget, DimA)
		if locArea == LocBoundary {
			tc.updateDimAB(isAreaA, LocBoundary, locTarget, DimL)
			tc.updateDimAB(isAreaA, LocExterior, locTarget, DimA)
		}
	}
}

// AddIntersection is the entry point for edge-intersection-derived
// nodes. The Go port currently only collects the sections; the
// node-evaluation step (RelateNode.finish + per-edge L-cell update)
// is not yet ported, so the higher-level driver should not depend
// on AddIntersection for correctness in cases where edges cross
// between vertices.
func (tc *TopologyComputer) AddIntersection(a, b *NodeSection) {
	if !a.IsSameGeometry(b) {
		tc.updateIntersectionAB(a, b)
	}
	tc.nodeMap[a.NodePt] = append(tc.nodeMap[a.NodePt], a, b)
}

func (tc *TopologyComputer) updateIntersectionAB(a, b *NodeSection) {
	if IsAreaArea(a, b) && IsProperPair(a, b) {
		tc.updateDim(LocInterior, LocInterior, DimA)
	}
	pt := a.NodePt
	locA := tc.geomA.LocateNode(pt, a.Polygon)
	locB := tc.geomB.LocateNode(pt, b.Polygon)
	tc.updateDim(locA, locB, DimP)
}

// EvaluateNodes is the placeholder for the future RelateNode-driven
// node evaluation pass. Today it is a no-op; once RelateNode +
// RelateEdge are ported, this method will iterate the nodeMap and
// finish each node's side-location propagation.
func (tc *TopologyComputer) EvaluateNodes() {
	// no-op until RelateNode/RelateEdge are ported.
	_ = tc.nodeMap
}
