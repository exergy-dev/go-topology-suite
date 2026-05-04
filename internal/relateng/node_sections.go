package relateng

import (
	"sort"

	"github.com/exergy-dev/go-topology-suite/geom"
)

// NodeSections is the per-coordinate aggregator built by the topology
// computer. As segment-pair intersections are reported, each one
// contributes a NodeSection to the bucket keyed by its node point.
// Once the segment-pair pass is finished, NodeSections.CreateNode
// constructs a fully-ordered RelateNode whose Finish() drives the
// final classification step.
//
// Port of org.locationtech.jts.operation.relateng.NodeSections.
type NodeSections struct {
	NodePt   geom.XY
	Sections []*NodeSection
}

// NewNodeSections allocates an empty bucket anchored at pt.
func NewNodeSections(pt geom.XY) *NodeSections {
	return &NodeSections{NodePt: pt}
}

// Add appends ns to the bucket.
func (n *NodeSections) Add(ns *NodeSection) {
	n.Sections = append(n.Sections, ns)
}

// HasInteractionAB reports whether the bucket has at least one A and
// one B section. Only nodes with AB interaction need to be evaluated.
func (n *NodeSections) HasInteractionAB() bool {
	hasA, hasB := false, false
	for _, ns := range n.Sections {
		if ns.IsA {
			hasA = true
		} else {
			hasB = true
		}
		if hasA && hasB {
			return true
		}
	}
	return false
}

// Polygonal returns the parent polygon of the first section matching
// isA, or nil if none. Used by the area-interior check.
func (n *NodeSections) Polygonal(isA bool) geom.Geometry {
	for _, ns := range n.Sections {
		if ns.IsA == isA && ns.Polygon != nil {
			return ns.Polygon
		}
	}
	return nil
}

// CreateNode produces a RelateNode by feeding sections to the
// CCW-ordered edge list. Sections from the same polygon are grouped so
// PolygonNodeConverter can normalise their topology before insertion.
func (n *NodeSections) CreateNode() *RelateNode {
	n.prepare()
	node := NewRelateNode(n.NodePt)
	i := 0
	for i < len(n.Sections) {
		ns := n.Sections[i]
		if ns.IsArea() && hasMultiplePolygonSections(n.Sections, i) {
			poly := collectPolygonSections(n.Sections, i)
			converted := convertPolygonNodeSections(poly)
			for _, c := range converted {
				node.AddEdgesFromSection(c)
			}
			i += len(poly)
			continue
		}
		node.AddEdgesFromSection(ns)
		i++
	}
	return node
}

// prepare sorts sections so that lines come before areas, and edges
// from the same polygon are contiguous.
func (n *NodeSections) prepare() {
	sort.SliceStable(n.Sections, func(i, j int) bool {
		return compareNodeSection(n.Sections[i], n.Sections[j]) < 0
	})
}

// compareNodeSection orders sections by:
//
//  1. isA (false < true so B sections come first when sorted ascending; JTS sorts so that line<area)
//  2. dimension (line < area)
//  3. element id
//  4. ring id
//
// This matches the JTS NodeSection.compareTo ordering well enough that
// PolygonNodeConverter can run on contiguous polygon sections.
func compareNodeSection(a, b *NodeSection) int {
	// isA: A=true sorts later (B first) — actually JTS sorts isA ascending (false=0, true=1).
	if a.IsA != b.IsA {
		if !a.IsA {
			return -1
		}
		return 1
	}
	if a.Dim != b.Dim {
		if a.Dim < b.Dim {
			return -1
		}
		return 1
	}
	if a.ID != b.ID {
		if a.ID < b.ID {
			return -1
		}
		return 1
	}
	if a.RingID != b.RingID {
		if a.RingID < b.RingID {
			return -1
		}
		return 1
	}
	return 0
}

func hasMultiplePolygonSections(sections []*NodeSection, i int) bool {
	if i >= len(sections)-1 {
		return false
	}
	return sections[i].IsSamePolygon(sections[i+1])
}

func collectPolygonSections(sections []*NodeSection, i int) []*NodeSection {
	first := sections[i]
	out := []*NodeSection{first}
	for j := i + 1; j < len(sections); j++ {
		if !first.IsSamePolygon(sections[j]) {
			break
		}
		out = append(out, sections[j])
	}
	return out
}
