package noding

import "github.com/terra-geo/terra/geom"

// SegmentString is a sequence of vertices defining a connected polyline.
// When the first and last vertices are equal the string represents a
// polygon ring boundary.
//
// Tag carries arbitrary caller-defined data through the noder unchanged.
// The overlay package uses Tag to mark which input geometry (e.g. polygon
// A vs polygon B) an edge originated from.
//
// SegmentString is a value carrier: the noder reads from input strings
// and returns freshly-allocated output strings; it never mutates input.
type SegmentString struct {
	Coords []geom.XY
	Tag    int
}

// NumSegments returns the number of edges in the string. A string with
// fewer than two coordinates has zero segments.
func (s *SegmentString) NumSegments() int {
	if len(s.Coords) < 2 {
		return 0
	}
	return len(s.Coords) - 1
}

// Segment returns the endpoints of the i-th edge.
func (s *SegmentString) Segment(i int) (geom.XY, geom.XY) {
	return s.Coords[i], s.Coords[i+1]
}
