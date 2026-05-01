package overlayng

import (
	terra "github.com/terra-geo/terra"
	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/internal/noding"
	"github.com/terra-geo/terra/internal/snap"
)

// Op identifies which boolean polygon operation to perform.
type Op int

const (
	OpIntersection Op = iota
	OpUnion
	OpDifference
	OpSymDiff
)

// Overlay is the single-polygon entry point. Default behaviour: no
// snap-rounding (preserves user input exactly). Use OverlayWithTolerance
// to handle inputs with near-coincident vertices.
func Overlay(subj, clip *geom.Polygon, op Op) (*geom.Polygon, []*geom.Polygon, error) {
	return OverlayWithTolerance(subj, clip, op, 0)
}

// OverlayWithTolerance is Overlay with explicit snap-rounding tolerance.
// Snapping the inputs to a common grid before noding eliminates the
// near-coincident-edge cases that defeat the brute-force segment
// intersector. A typical choice for unit-scale inputs is tolerance =
// 1e-9; for lon/lat data, ~1e-7 (~1 cm). Pass tolerance=0 to skip the
// snap pass and use raw coordinates.
func OverlayWithTolerance(subj, clip *geom.Polygon, op Op, tolerance float64) (*geom.Polygon, []*geom.Polygon, error) {
	return OverlayPolygonalWithTolerance(
		[]*geom.Polygon{subj}, []*geom.Polygon{clip}, op, tolerance,
	)
}

// OverlayPolygonal accepts polygon slices for both subj and clip, so
// MultiPolygon overlay routes through the same DCEL-and-classifier path
// as single-polygon overlay. Each input is treated as the union of its
// constituent polygons (each of which carries its own outer ring + holes).
//
// Returned shape: one "first" polygon plus zero-or-more disjoint "rest"
// polygons, identical to the single-polygon entry point.
func OverlayPolygonal(subj, clip []*geom.Polygon, op Op) (*geom.Polygon, []*geom.Polygon, error) {
	return OverlayPolygonalWithTolerance(subj, clip, op, 0)
}

// OverlayPolygonalWithTolerance is the polygonal-input entry point with
// explicit snap-rounding tolerance.
func OverlayPolygonalWithTolerance(subj, clip []*geom.Polygon, op Op, tolerance float64) (*geom.Polygon, []*geom.Polygon, error) {
	c, err := commonCRS(subj, clip)
	if err != nil {
		return nil, nil, err
	}

	// Filter empties; flatten each polygon into its ring list.
	subjRings, subjPerPoly := snapAndPartition(subj, tolerance)
	clipRings, clipPerPoly := snapAndPartition(clip, tolerance)
	if len(subjRings) == 0 || len(clipRings) == 0 {
		return geom.NewEmptyPolygon(c, geom.LayoutXY), nil, nil
	}

	// Goodrich-Guibas hot-pixel pass: with subj and clip sharing a
	// hot-pixel set, a vertex from one input that snaps into the
	// other's segment path forces a split, preventing the DCEL from
	// disconnecting at near-vertices.
	if tolerance > 0 {
		subjRings, clipRings = hotPixelRoundCombined(subjRings, clipRings, tolerance)
		if len(subjRings) == 0 || len(clipRings) == 0 {
			return geom.NewEmptyPolygon(c, geom.LayoutXY), nil, nil
		}
	}

	return overlayCorePolygonal(c, subjRings, subjPerPoly, clipRings, clipPerPoly, op)
}

// hotPixelRoundCombined runs hot-pixel snap rounding across the
// combined subj+clip ring set so hot pixels from either side trigger
// splits in the other side's segments. Per-polygon partitions remain
// valid because hot-pixel processing only inserts vertices; no rings
// are dropped at this stage.
func hotPixelRoundCombined(subjRings, clipRings [][]geom.XY, tolerance float64) (subjOut, clipOut [][]geom.XY) {
	hp := snap.NewHotPixelSet(tolerance)
	for _, r := range subjRings {
		for _, v := range r {
			hp.Add(v)
		}
	}
	for _, r := range clipRings {
		for _, v := range r {
			hp.Add(v)
		}
	}
	subjOut = make([][]geom.XY, 0, len(subjRings))
	for _, r := range subjRings {
		if noded := hp.NodeRing(r); noded != nil {
			subjOut = append(subjOut, noded)
		}
	}
	clipOut = make([][]geom.XY, 0, len(clipRings))
	for _, r := range clipRings {
		if noded := hp.NodeRing(r); noded != nil {
			clipOut = append(clipOut, noded)
		}
	}
	return subjOut, clipOut
}

// commonCRS returns the CRS shared by all input polygons, or an error if
// they disagree (or both lists are empty).
func commonCRS(subj, clip []*geom.Polygon) (*crs.CRS, error) {
	var c *crs.CRS
	first := true
	for _, p := range subj {
		if first {
			c = p.CRS()
			first = false
			continue
		}
		if !crs.Equal(c, p.CRS()) {
			return nil, terra.ErrCRSMismatch
		}
	}
	for _, p := range clip {
		if first {
			c = p.CRS()
			first = false
			continue
		}
		if !crs.Equal(c, p.CRS()) {
			return nil, terra.ErrCRSMismatch
		}
	}
	return c, nil
}

// snapAndPartition flattens a polygon list into a single ring list (for
// segment-string emission) plus a parallel "ring count per polygon"
// slice that lets the classifier reconstruct per-polygon containment.
func snapAndPartition(polys []*geom.Polygon, tolerance float64) (rings [][]geom.XY, perPoly []int) {
	for _, p := range polys {
		if p == nil || p.IsEmpty() {
			continue
		}
		r := snapAllRings(p, tolerance)
		if r == nil {
			continue
		}
		perPoly = append(perPoly, len(r))
		rings = append(rings, r...)
	}
	return rings, perPoly
}

// snapAllRings extracts every ring (outer + holes) from a polygon and
// optionally snap-rounds them. Rings that collapse under snap are
// dropped — except the outer ring; if that collapses we return nil so
// the caller can short-circuit to an empty result.
func snapAllRings(p *geom.Polygon, tolerance float64) [][]geom.XY {
	out := make([][]geom.XY, 0, p.NumRings())
	for r := 0; r < p.NumRings(); r++ {
		ring := p.Ring(r)
		if tolerance > 0 {
			ring = snap.New(tolerance).SnapRing(ring)
			if ring == nil {
				if r == 0 {
					return nil
				}
				continue
			}
		}
		out = append(out, ring)
	}
	return out
}

// overlayCore is the single-polygon shared body: node every ring →
// DCEL → classify faces against the original (multi-ring) polygons →
// trace result rings → reassemble outers and holes.
func overlayCore(c *crs.CRS, subjRings, clipRings [][]geom.XY, op Op) (*geom.Polygon, []*geom.Polygon, error) {
	// Single-polygon overlay routes through the polygonal entry point
	// with one polygon per side; perPoly slices encode that.
	return overlayCorePolygonal(c,
		subjRings, []int{len(subjRings)},
		clipRings, []int{len(clipRings)},
		op,
	)
}

// overlayCorePolygonal is the multi-aware shared body. ringsSubj is
// the flat ring list across all subj polygons; subjPerPoly[i] is the
// number of rings (outer + holes) belonging to subj polygon i. Same
// shape for clip.
func overlayCorePolygonal(
	c *crs.CRS,
	subjRings [][]geom.XY, subjPerPoly []int,
	clipRings [][]geom.XY, clipPerPoly []int,
	op Op,
) (*geom.Polygon, []*geom.Polygon, error) {
	segs := make([]*noding.SegmentString, 0, len(subjRings)+len(clipRings))
	for _, r := range subjRings {
		segs = append(segs, &noding.SegmentString{
			Coords: append([]geom.XY(nil), r...),
			Tag:    1,
		})
	}
	for _, r := range clipRings {
		segs = append(segs, &noding.SegmentString{
			Coords: append([]geom.XY(nil), r...),
			Tag:    2,
		})
	}
	noded := nodeAdaptive(segs)
	taggedSegs := flattenNoded(noded)
	d := buildDCEL(taggedSegs)
	d.traceFaces()

	if !d.isConnected() {
		// Topology graph disconnected — subj and clip share no
		// boundary. Hand off to the disjoint helper, which now
		// understands multi-polygon inputs.
		return overlayDisjointPolygonal(c,
			rebuildPolygons(c, subjRings, subjPerPoly),
			rebuildPolygons(c, clipRings, clipPerPoly),
			op,
		)
	}

	classifyFacesByPolygons(d, subjRings, subjPerPoly, clipRings, clipPerPoly)
	applyOp(d, op)
	rings := extractResultRings(d)
	if len(rings) == 0 {
		return geom.NewEmptyPolygon(c, geom.LayoutXY), nil, nil
	}
	return assembleOutputPolygons(c, rings)
}

// rebuildPolygons reconstructs the per-polygon slice from a flat ring
// list and a per-polygon ring-count partition. Used by the disjoint
// fallback, which needs per-polygon containment tests.
func rebuildPolygons(c *crs.CRS, rings [][]geom.XY, perPoly []int) []*geom.Polygon {
	out := make([]*geom.Polygon, 0, len(perPoly))
	off := 0
	for _, n := range perPoly {
		if n == 0 || off+n > len(rings) {
			continue
		}
		out = append(out, geom.NewPolygon(c, rings[off:off+n]...))
		off += n
	}
	return out
}

// flattenNoded turns a slice of noded SegmentStrings into our internal
// tagged 2-vertex edges. When two SegmentStrings produce the same
// directed segment, the DCEL builder merges them and ORs the tags so
// shared edges carry both source labels.
func flattenNoded(strings []*noding.SegmentString) []taggedSegment {
	var out []taggedSegment
	for _, s := range strings {
		if len(s.Coords) < 2 {
			continue
		}
		tag := uint8(s.Tag)
		for i := 0; i+1 < len(s.Coords); i++ {
			out = append(out, taggedSegment{
				p0:  s.Coords[i],
				p1:  s.Coords[i+1],
				tag: tag,
			})
		}
	}
	return out
}
