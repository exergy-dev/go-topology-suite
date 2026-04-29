package overlay

import (
	"fmt"

	terra "github.com/terra-geo/terra"
	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/kernel/planar"
)

// Intersection returns subject ∩ clipper.
//
// Fast path: if clipper is a convex single-ring polygon, Sutherland-Hodgman
// is used (numerically robust, simple).
//
// General path: Greiner-Hormann is used for arbitrary simple polygons.
// See greiner_hormann.go for v0.1 limitations (no holes, vertex-coincident
// inputs unreliable).
func Intersection(subject, clipper geom.Geometry) (geom.Geometry, error) {
	if !crs.Equal(subject.CRS(), clipper.CRS()) {
		return nil, terra.ErrCRSMismatch
	}
	if subject.IsEmpty() || clipper.IsEmpty() {
		return geom.NewEmptyPolygon(subject.CRS(), geom.LayoutXY), nil
	}
	subj, ok := subject.(*geom.Polygon)
	if !ok {
		return nil, fmt.Errorf("overlay: subject must be Polygon (got %T): %w",
			subject, terra.ErrUnsupportedKernel)
	}
	clip, ok := clipper.(*geom.Polygon)
	if !ok {
		return nil, fmt.Errorf("overlay: clipper must be Polygon (got %T): %w",
			clipper, terra.ErrUnsupportedKernel)
	}
	// Convex fast-path.
	if clip.NumRings() == 1 && isConvexCCW(clip.Ring(0)) {
		clipRing := clip.Ring(0)
		rings := make([][]geom.XY, 0, subj.NumRings())
		for r := 0; r < subj.NumRings(); r++ {
			clipped := sutherlandHodgman(subj.Ring(r), clipRing)
			if len(clipped) >= 4 {
				rings = append(rings, clipped)
			}
		}
		if len(rings) == 0 {
			return geom.NewEmptyPolygon(subject.CRS(), geom.LayoutXY), nil
		}
		return geom.NewPolygon(subject.CRS(), rings...), nil
	}
	// General path: Greiner-Hormann.
	return IntersectionGeneral(subject, clipper)
}

// isConvexCCW returns true iff the ring is convex and counter-clockwise.
// Convex iff every consecutive triple has the same chirality (CCW).
func isConvexCCW(ring []geom.XY) bool {
	if len(ring) < 4 {
		return false
	}
	k := planar.Default
	prev := 0 // 0 means undecided
	for i := 0; i+2 < len(ring); i++ {
		o := k.Orient(ring[i], ring[i+1], ring[i+2])
		switch o {
		case 1: // CCW
			if prev == -1 {
				return false
			}
			prev = 1
		case -1:
			if prev == 1 {
				return false
			}
			prev = -1
		}
	}
	return prev == 1
}

// sutherlandHodgman clips subject by the convex CCW clip polygon,
// returning the intersection ring. The standard textbook algorithm: clip
// the subject one edge of the clipper at a time, keeping/cutting points
// based on which side they fall on relative to the (infinite) clip edge.
func sutherlandHodgman(subject, clip []geom.XY) []geom.XY {
	output := append([]geom.XY(nil), subject...)
	if len(output) > 0 && output[0] == output[len(output)-1] {
		output = output[:len(output)-1]
	}
	for i := 0; i+1 < len(clip); i++ {
		if len(output) == 0 {
			return nil
		}
		ce1, ce2 := clip[i], clip[i+1]
		// Snapshot input — output's backing storage will be reused for the
		// new ring, so we must not let `input` alias future writes.
		input := append([]geom.XY(nil), output...)
		output = output[:0]
		if len(input) == 0 {
			break
		}
		s := input[len(input)-1]
		for _, e := range input {
			if leftOfOrOn(ce1, ce2, e) {
				if !leftOfOrOn(ce1, ce2, s) {
					if ip, ok := lineLineIntersect(s, e, ce1, ce2); ok {
						output = append(output, ip)
					}
				}
				output = append(output, e)
			} else if leftOfOrOn(ce1, ce2, s) {
				if ip, ok := lineLineIntersect(s, e, ce1, ce2); ok {
					output = append(output, ip)
				}
			}
			s = e
		}
	}
	if len(output) > 0 {
		output = append(output, output[0])
	}
	return output
}

// leftOfOrOn reports whether p lies to the left of or on the directed
// line from a to b. CCW convention.
func leftOfOrOn(a, b, p geom.XY) bool {
	cross := (b.X-a.X)*(p.Y-a.Y) - (b.Y-a.Y)*(p.X-a.X)
	return cross >= 0
}

// lineLineIntersect intersects two infinite lines. Used by S-H to compute
// the cutoff point of a subject edge against a clip edge.
func lineLineIntersect(a1, a2, b1, b2 geom.XY) (geom.XY, bool) {
	rx := a2.X - a1.X
	ry := a2.Y - a1.Y
	sx := b2.X - b1.X
	sy := b2.Y - b1.Y
	denom := rx*sy - ry*sx
	if denom == 0 {
		return geom.XY{}, false
	}
	tNum := (b1.X-a1.X)*sy - (b1.Y-a1.Y)*sx
	t := tNum / denom
	return geom.XY{X: a1.X + t*rx, Y: a1.Y + t*ry}, true
}
