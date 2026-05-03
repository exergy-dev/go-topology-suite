package noding

import (
	"github.com/terra-geo/terra/geom"
)

// DissolveSegmentStrings removes duplicate and reverse-equal segment
// strings from input, preserving the first occurrence of each unique
// vertex sequence. Two strings are considered equal if their coordinate
// arrays are identical, OR if reversing one yields the other.
//
// Mirrors org.locationtech.jts.noding.SegmentStringDissolver. The JTS
// class uses a TreeMap keyed by an OctagonalEnvelope hash to dedupe;
// we use a string-keyed Go map (the strings are already deduped by
// construction in our pipeline, so we don't need the secondary
// envelope hash).
//
// Tag is preserved from the first occurrence; duplicates' Tags are
// dropped on the floor (matching JTS behaviour: the first encounter
// wins).
func DissolveSegmentStrings(input []*SegmentString) []*SegmentString {
	if len(input) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(input))
	out := make([]*SegmentString, 0, len(input))
	for _, ss := range input {
		k := canonicalKey(ss.Coords)
		if _, dup := seen[k]; dup {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, &SegmentString{
			Coords: append([]geom.XY(nil), ss.Coords...),
			Tag:    ss.Tag,
		})
	}
	return out
}

// canonicalKey returns a deterministic key that is identical for a
// coord sequence and its reverse. The key is built from the
// lexicographically smaller of the forward and reverse encodings.
func canonicalKey(coords []geom.XY) string {
	if len(coords) == 0 {
		return ""
	}
	fwd := encodeCoords(coords, false)
	rev := encodeCoords(coords, true)
	if fwd <= rev {
		return fwd
	}
	return rev
}

// encodeCoords returns a deterministic byte-string encoding of coords,
// optionally walking them in reverse order.
func encodeCoords(coords []geom.XY, reverse bool) string {
	buf := make([]byte, 0, 32*len(coords))
	if reverse {
		for i := len(coords) - 1; i >= 0; i-- {
			buf = appendXY(buf, coords[i])
		}
	} else {
		for _, p := range coords {
			buf = appendXY(buf, p)
		}
	}
	return string(buf)
}
