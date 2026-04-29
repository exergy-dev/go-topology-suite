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

// Overlay is the entry point for the overlay-NG path. Default behaviour:
// no snap-rounding (preserves user input exactly). Use OverlayWithTolerance
// to handle inputs with near-coincident vertices.
func Overlay(subj, clip *geom.Polygon, op Op) (*geom.Polygon, []*geom.Polygon, error) {
	return OverlayWithTolerance(subj, clip, op, 0)
}

// OverlayWithTolerance is Overlay with explicit snap-rounding tolerance.
// Snapping the inputs to a common grid before noding eliminates the
// near-coincident-edge cases that defeat the brute-force segment
// intersector (parallel near-coincident edges return ok=false from
// SegmentIntersection, leaving the topology graph disconnected at
// near-shared vertices). A typical choice for unit-scale inputs is
// tolerance = 1e-9; for lon/lat data, ~1e-7 (~1 cm). Pass tolerance=0
// to skip the snap pass and use raw coordinates.
func OverlayWithTolerance(subj, clip *geom.Polygon, op Op, tolerance float64) (*geom.Polygon, []*geom.Polygon, error) {
	if !crs.Equal(subj.CRS(), clip.CRS()) {
		return nil, nil, terra.ErrCRSMismatch
	}
	if subj.IsEmpty() || clip.IsEmpty() {
		return geom.NewEmptyPolygon(subj.CRS(), geom.LayoutXY), nil, nil
	}

	subjRings := snapAllRings(subj, tolerance)
	clipRings := snapAllRings(clip, tolerance)
	if len(subjRings) == 0 || len(clipRings) == 0 {
		return geom.NewEmptyPolygon(subj.CRS(), geom.LayoutXY), nil, nil
	}
	return overlayCore(subj.CRS(), subjRings, clipRings, op)
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

// overlayCore is the shared body: node every ring → DCEL → classify
// faces against the original (multi-ring) polygons → trace result rings
// → reassemble outers and holes.
func overlayCore(c *crs.CRS, subjRings, clipRings [][]geom.XY, op Op) (*geom.Polygon, []*geom.Polygon, error) {
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
		// Topology graph disconnected — subj and clip share no boundary.
		// Hand off to the disjoint helper using shell-only inputs.
		return overlayDisjoint(
			geom.NewPolygon(c, subjRings...),
			geom.NewPolygon(c, clipRings...),
			op,
		)
	}

	classifyFacesPolygons(d, subjRings, clipRings)
	applyOp(d, op)
	rings := extractResultRings(d)
	if len(rings) == 0 {
		return geom.NewEmptyPolygon(c, geom.LayoutXY), nil, nil
	}
	return assembleOutputPolygons(c, rings)
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
