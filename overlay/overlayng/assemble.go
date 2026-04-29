package overlayng

import (
	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/geom"
)

// assembleOutputPolygons takes the boundary rings produced by
// extractResultRings and groups them into Polygons by detecting
// containment: a ring contained in exactly one other ring becomes a
// hole of that ring; doubly-contained rings (a ring inside a ring
// inside a ring) become separate outer polygons; etc.
//
// Algorithm:
//  1. For each ring, count how many OTHER rings contain its first
//     vertex. The depth tells us outer vs hole vs nested-outer.
//     - depth 0: outermost outer
//     - depth 1: hole
//     - depth 2: outer inside the hole
//     - depth 3: hole inside that
//  2. Each ring with even depth is an outer; assign it any rings of
//     depth+1 that are immediately contained (i.e. no intermediate).
//
// Edge case: if no rings have even depth (all-odd), treat shallowest as
// outer — defensive fallback for inputs the algorithm might mis-orient.
func assembleOutputPolygons(c *crs.CRS, rings [][]geom.XY) (*geom.Polygon, []*geom.Polygon, error) {
	if len(rings) == 0 {
		return geom.NewEmptyPolygon(c, geom.LayoutXY), nil, nil
	}
	if len(rings) == 1 {
		return geom.NewPolygon(c, rings[0]), nil, nil
	}

	depths := make([]int, len(rings))
	for i := range rings {
		for j := range rings {
			if i == j {
				continue
			}
			if pointInRing(rings[i][0], rings[j]) {
				depths[i]++
			}
		}
	}

	type group struct {
		outer int
		holes []int
	}
	var groups []group

	// For each even-depth ring, find its holes: rings of depth+1
	// contained directly in this ring (and not in any intermediate
	// ring of higher depth).
	for i := range rings {
		if depths[i]%2 != 0 {
			continue
		}
		g := group{outer: i}
		for j := range rings {
			if i == j || depths[j] != depths[i]+1 {
				continue
			}
			if !pointInRing(rings[j][0], rings[i]) {
				continue
			}
			// Confirm this is the IMMEDIATE outer: no other even-depth
			// ring of depth=depths[i]+? interposes. Simpler check: among
			// all even-depth rings containing j, i should be the deepest.
			deeperContainer := false
			for k := range rings {
				if k == i || depths[k] >= depths[i]+1 {
					continue
				}
				if !pointInRing(rings[j][0], rings[k]) {
					continue
				}
				if depths[k] > depths[i] {
					deeperContainer = true
					break
				}
			}
			if !deeperContainer {
				g.holes = append(g.holes, j)
			}
		}
		groups = append(groups, g)
	}

	if len(groups) == 0 {
		// All rings odd-depth — defensive fallback: emit each as a separate outer.
		first := geom.NewPolygon(c, rings[0])
		var rest []*geom.Polygon
		for i := 1; i < len(rings); i++ {
			rest = append(rest, geom.NewPolygon(c, rings[i]))
		}
		return first, rest, nil
	}

	// Build a polygon per group.
	polys := make([]*geom.Polygon, 0, len(groups))
	for _, g := range groups {
		all := make([][]geom.XY, 0, 1+len(g.holes))
		all = append(all, rings[g.outer])
		for _, h := range g.holes {
			all = append(all, rings[h])
		}
		polys = append(polys, geom.NewPolygon(c, all...))
	}

	if len(polys) == 1 {
		return polys[0], nil, nil
	}
	return polys[0], polys[1:], nil
}
