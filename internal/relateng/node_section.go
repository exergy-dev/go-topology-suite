package relateng

import "github.com/terra-geo/terra/geom"

// NodeSection is one half of a topology node in JTS RelateNG: a
// node point plus its incident edge vertices. Port of
// org.locationtech.jts.operation.relateng.NodeSection.
//
// A node in an areal geometry always has both incident vertices
// populated (the entering and exiting edge tangents).
//
// A node in a linear geometry may have one of the incident vertices
// missing — when the node is at a line endpoint.
//
// The "edges of an area node are CW-shell-oriented" requirement
// (JTS norm) is the caller's responsibility; the data class merely
// stores what it is given.
type NodeSection struct {
	IsA            bool          // true if this section belongs to geometry A
	Dim            int           // DimP / DimL / DimA
	ID             int           // element id within the parent geometry
	RingID         int           // ring id within a polygon (0 = shell)
	IsNodeAtVertex bool          // true if the node coincides with an existing vertex
	NodePt         geom.XY       // the node point itself
	V0, V1         *geom.XY      // optional incident vertices
	Polygon        geom.Geometry // the polygon this section is part of (nil for non-area)
}

// NewNodeSection constructs a NodeSection with the given fields.
// Either or both of v0/v1 may be nil for line-endpoint cases.
func NewNodeSection(isA bool, dim, id, ringID int, poly geom.Geometry,
	isNodeAtVertex bool, v0 *geom.XY, nodePt geom.XY, v1 *geom.XY) *NodeSection {
	return &NodeSection{
		IsA:            isA,
		Dim:            dim,
		ID:             id,
		RingID:         ringID,
		IsNodeAtVertex: isNodeAtVertex,
		NodePt:         nodePt,
		V0:             v0,
		V1:             v1,
		Polygon:        poly,
	}
}

// Vertex returns V0 (i==0) or V1 (otherwise). May return nil for
// missing line-endpoint vertices.
func (s *NodeSection) Vertex(i int) *geom.XY {
	if i == 0 {
		return s.V0
	}
	return s.V1
}

// IsShell reports whether this section is on the shell (ring 0).
func (s *NodeSection) IsShell() bool { return s.RingID == 0 }

// IsArea reports whether the parent component is areal.
func (s *NodeSection) IsArea() bool { return s.Dim == DimA }

// IsSameGeometry reports whether s and o belong to the same input
// geometry (both A or both B).
func (s *NodeSection) IsSameGeometry(o *NodeSection) bool {
	return s.IsA == o.IsA
}

// IsSamePolygon reports whether s and o are in the same polygon
// element of the same parent geometry.
func (s *NodeSection) IsSamePolygon(o *NodeSection) bool {
	return s.IsA == o.IsA && s.ID == o.ID
}

// IsProper reports whether the node is NOT at an existing vertex
// (i.e. it's a "proper" mid-segment intersection).
func (s *NodeSection) IsProper() bool { return !s.IsNodeAtVertex }

// IsAreaArea reports whether both sections represent area edges.
func IsAreaArea(a, b *NodeSection) bool {
	return a.Dim == DimA && b.Dim == DimA
}

// IsProperPair reports whether both sections are proper (not at
// vertex). Mirrors NodeSection.isProper(NodeSection,NodeSection).
func IsProperPair(a, b *NodeSection) bool { return a.IsProper() && b.IsProper() }
