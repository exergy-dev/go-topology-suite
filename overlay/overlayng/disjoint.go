package overlayng

import "github.com/terra-geo/terra/geom"

// overlayDisjoint handles the case where subj and clip share no boundary
// — either they're geometrically disjoint or one fully contains the
// other. Returns the appropriate result for each operation by checking
// containment via a single ray-cast per polygon.
func overlayDisjoint(subj, clip *geom.Polygon, op Op) (*geom.Polygon, []*geom.Polygon, error) {
	subjRing := subj.Ring(0)
	clipRing := clip.Ring(0)
	subjInClip := pointInRing(subjRing[0], clipRing)
	clipInSubj := pointInRing(clipRing[0], subjRing)

	switch op {
	case OpIntersection:
		switch {
		case subjInClip:
			return subj, nil, nil
		case clipInSubj:
			return clip, nil, nil
		}
		return geom.NewEmptyPolygon(subj.CRS(), geom.LayoutXY), nil, nil

	case OpUnion:
		switch {
		case subjInClip:
			return clip, nil, nil
		case clipInSubj:
			return subj, nil, nil
		}
		// Disjoint → both polygons in the result.
		return subj, []*geom.Polygon{clip}, nil

	case OpDifference:
		switch {
		case subjInClip:
			return geom.NewEmptyPolygon(subj.CRS(), geom.LayoutXY), nil, nil
		case clipInSubj:
			// subj with clip as a hole — outer + reversed inner ring.
			outer := append([]geom.XY(nil), subjRing...)
			hole := append([]geom.XY(nil), clipRing...)
			// Reverse to make hole CW relative to outer's CCW.
			for i, j := 0, len(hole)-1; i < j; i, j = i+1, j-1 {
				hole[i], hole[j] = hole[j], hole[i]
			}
			return geom.NewPolygon(subj.CRS(), outer, hole), nil, nil
		}
		return subj, nil, nil

	case OpSymDiff:
		switch {
		case subjInClip:
			outer := append([]geom.XY(nil), clipRing...)
			hole := append([]geom.XY(nil), subjRing...)
			for i, j := 0, len(hole)-1; i < j; i, j = i+1, j-1 {
				hole[i], hole[j] = hole[j], hole[i]
			}
			return geom.NewPolygon(subj.CRS(), outer, hole), nil, nil
		case clipInSubj:
			outer := append([]geom.XY(nil), subjRing...)
			hole := append([]geom.XY(nil), clipRing...)
			for i, j := 0, len(hole)-1; i < j; i, j = i+1, j-1 {
				hole[i], hole[j] = hole[j], hole[i]
			}
			return geom.NewPolygon(subj.CRS(), outer, hole), nil, nil
		}
		return subj, []*geom.Polygon{clip}, nil
	}
	return geom.NewEmptyPolygon(subj.CRS(), geom.LayoutXY), nil, nil
}
