package relateng

import (
	"sort"

	"github.com/terra-geo/terra/geom"
)

// convertPolygonNodeSections takes the contiguous polygon sections that
// share an element id and converts the OGC "touching-rings" structure
// into the equivalent self-touch / inverted-ring structure that
// RelateNode's edge ordering expects.
//
// Port of org.locationtech.jts.operation.relateng.PolygonNodeConverter.
//
// Inputs are assumed to have canonical orientation (CW shells, CCW
// holes) and to be topologically valid (no crossing or collinear
// segments). The same number of output sections is produced as input
// sections; each output corner encloses an area that lies entirely
// inside the polygon.
func convertPolygonNodeSections(polySections []*NodeSection) []*NodeSection {
	if len(polySections) <= 1 {
		return polySections
	}
	// Sort by edge angle around the node so that adjacent corner pairs
	// are contiguous in the slice.
	sorted := make([]*NodeSection, len(polySections))
	copy(sorted, polySections)
	sort.SliceStable(sorted, func(i, j int) bool {
		return polygonNodeAngleLess(sorted[i], sorted[j])
	})

	sections := extractUniquePolygonSections(sorted)
	if len(sections) == 1 {
		return sections
	}

	shellIdx := findShellIndex(sections)
	if shellIdx < 0 {
		return convertHoleSections(sections)
	}
	out := make([]*NodeSection, 0, len(sections))
	next := shellIdx
	for {
		next = convertShellAndHoles(sections, next, &out)
		if next == shellIdx {
			break
		}
	}
	return out
}

func convertShellAndHoles(sections []*NodeSection, shellIdx int, out *[]*NodeSection) int {
	shell := sections[shellIdx]
	if shell.V0 == nil || shell.V1 == nil {
		// Degenerate shell — fall through and skip.
		return nextSectionIdx(sections, shellIdx)
	}
	inVertex := *shell.V0
	i := nextSectionIdx(sections, shellIdx)
	for !sections[i].IsShell() {
		hole := sections[i]
		if hole.V0 == nil || hole.V1 == nil {
			i = nextSectionIdx(sections, i)
			continue
		}
		outVertex := *hole.V1
		*out = append(*out, createConvertedSection(shell, inVertex, outVertex))
		inVertex = *hole.V0
		i = nextSectionIdx(sections, i)
	}
	outVertex := *shell.V1
	*out = append(*out, createConvertedSection(shell, inVertex, outVertex))
	return i
}

func convertHoleSections(sections []*NodeSection) []*NodeSection {
	out := make([]*NodeSection, 0, len(sections))
	template := sections[0]
	for i := 0; i < len(sections); i++ {
		next := nextSectionIdx(sections, i)
		cur := sections[i]
		nx := sections[next]
		if cur.V0 == nil || nx.V1 == nil {
			continue
		}
		out = append(out, createConvertedSection(template, *cur.V0, *nx.V1))
	}
	return out
}

func createConvertedSection(template *NodeSection, v0, v1 geom.XY) *NodeSection {
	v0c := v0
	v1c := v1
	return NewNodeSection(template.IsA, DimA, template.ID, 0, template.Polygon,
		template.IsNodeAtVertex, &v0c, template.NodePt, &v1c)
}

func extractUniquePolygonSections(sections []*NodeSection) []*NodeSection {
	out := make([]*NodeSection, 0, len(sections))
	last := sections[0]
	out = append(out, last)
	for _, ns := range sections[1:] {
		if !sameAngle(last, ns) {
			out = append(out, ns)
			last = ns
		}
	}
	return out
}

func findShellIndex(sections []*NodeSection) int {
	for i, ns := range sections {
		if ns.IsShell() {
			return i
		}
	}
	return -1
}

func nextSectionIdx(sections []*NodeSection, i int) int {
	if i+1 >= len(sections) {
		return 0
	}
	return i + 1
}

// polygonNodeAngleLess orders sections by the angle of their incoming
// edge (V0) around the node point, increasing CCW. Mirrors the JTS
// NodeSection.EdgeAngleComparator key.
func polygonNodeAngleLess(a, b *NodeSection) bool {
	if a.V0 == nil || b.V0 == nil {
		return a.V0 != nil && b.V0 == nil
	}
	return compareAngle(a.NodePt, *a.V0, *b.V0) < 0
}

// sameAngle reports whether two sections share the same V0 angle around
// the common node point — i.e. whether they would compare equal under
// polygonNodeAngleLess.
func sameAngle(a, b *NodeSection) bool {
	if a.V0 == nil || b.V0 == nil {
		return false
	}
	return compareAngle(a.NodePt, *a.V0, *b.V0) == 0
}
