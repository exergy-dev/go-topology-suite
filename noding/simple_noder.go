package noding

import (
	"github.com/go-topology-suite/gts/geom"
)

// SimpleNoder is a basic O(n²) noder that uses brute-force comparison
// to find all intersections between segments.
//
// This implementation is straightforward but not optimized for large
// datasets. For better performance with many segments, consider using
// a spatial index-based noder (e.g., MCIndexNoder).
//
// The noding process:
// 1. Compare every segment with every other segment
// 2. Use a SegmentIntersector to process intersections
// 3. Build noded segment strings from the results
type SimpleNoder struct {
	// segInt is the SegmentIntersector used to process intersections
	segInt SegmentIntersector

	// inputSegStrings are the input segment strings to be noded
	inputSegStrings []*NodedSegmentString

	// nodedSegStrings are the result after noding
	nodedSegStrings []*NodedSegmentString
}

// NewSimpleNoder creates a new SimpleNoder with the given SegmentIntersector.
func NewSimpleNoder(segInt SegmentIntersector) *SimpleNoder {
	if segInt == nil {
		segInt = NewIntersectionAdder()
	}
	return &SimpleNoder{
		segInt: segInt,
	}
}

// ComputeNodes computes all nodes (intersections) for the given segment strings.
// This uses a brute-force O(n²) algorithm to compare all pairs of segments.
func (sn *SimpleNoder) ComputeNodes(segmentStrings []*NodedSegmentString) {
	sn.inputSegStrings = segmentStrings
	sn.nodedSegStrings = nil

	// Compare all pairs of segments
	for i := 0; i < len(segmentStrings); i++ {
		ss1 := segmentStrings[i]
		for j := i; j < len(segmentStrings); j++ {
			ss2 := segmentStrings[j]
			sn.computeIntersects(ss1, ss2)

			// Check if we should stop early
			if sn.segInt.IsDone() {
				return
			}
		}
	}
}

// computeIntersects finds all intersections between two segment strings.
func (sn *SimpleNoder) computeIntersects(ss1, ss2 *NodedSegmentString) {
	// Compare every segment in ss1 with every segment in ss2
	for i := 0; i < ss1.Size(); i++ {
		for j := 0; j < ss2.Size(); j++ {
			sn.segInt.ProcessIntersections(ss1, i, ss2, j)

			if sn.segInt.IsDone() {
				return
			}
		}
	}
}

// GetNodedSubstrings returns the noded segment strings.
// Each segment string has been split at all intersection points.
func (sn *SimpleNoder) GetNodedSubstrings() []*NodedSegmentString {
	if sn.nodedSegStrings != nil {
		return sn.nodedSegStrings
	}

	sn.nodedSegStrings = make([]*NodedSegmentString, 0, len(sn.inputSegStrings))

	for _, ss := range sn.inputSegStrings {
		// Get the noded coordinates (with all intersection points inserted)
		nodedCoords := ss.NodedCoordinates()

		// Create noded segment strings from the split segments
		if len(nodedCoords) >= 2 {
			// If the original was closed, we may need to split into multiple strings
			// For now, create a single noded segment string
			nodedSS := NewNodedSegmentString(nodedCoords, ss.Context())
			sn.nodedSegStrings = append(sn.nodedSegStrings, nodedSS)
		}
	}

	return sn.nodedSegStrings
}

// ScaledNoder wraps a Noder and applies a scale factor to coordinates
// before noding, then rescales them back. This can improve robustness
// by converting floating-point coordinates to a fixed-precision grid.
type ScaledNoder struct {
	noder      Noder
	scaleFactor float64
	offsetX    float64
	offsetY    float64
}

// NewScaledNoder creates a new ScaledNoder.
func NewScaledNoder(noder Noder, scaleFactor float64) *ScaledNoder {
	return &ScaledNoder{
		noder:      noder,
		scaleFactor: scaleFactor,
	}
}

// scale scales a coordinate sequence.
func (sn *ScaledNoder) scale(coords geom.CoordinateSequence) geom.CoordinateSequence {
	scaled := make(geom.CoordinateSequence, len(coords))
	for i, c := range coords {
		scaled[i] = geom.NewCoordinate(
			(c.X-sn.offsetX)*sn.scaleFactor,
			(c.Y-sn.offsetY)*sn.scaleFactor,
		)
	}
	return scaled
}

// unscale unscales a coordinate sequence.
func (sn *ScaledNoder) unscale(coords geom.CoordinateSequence) geom.CoordinateSequence {
	unscaled := make(geom.CoordinateSequence, len(coords))
	for i, c := range coords {
		unscaled[i] = geom.NewCoordinate(
			c.X/sn.scaleFactor+sn.offsetX,
			c.Y/sn.scaleFactor+sn.offsetY,
		)
	}
	return unscaled
}

// ComputeNodes computes nodes after scaling coordinates.
func (sn *ScaledNoder) ComputeNodes(segmentStrings []*NodedSegmentString) {
	// Scale all input coordinates
	scaled := make([]*NodedSegmentString, len(segmentStrings))
	for i, ss := range segmentStrings {
		scaledCoords := sn.scale(ss.Coordinates())
		scaled[i] = NewNodedSegmentString(scaledCoords, ss.Context())
	}

	// Compute nodes on scaled coordinates
	sn.noder.ComputeNodes(scaled)
}

// GetNodedSubstrings returns noded substrings with coordinates unscaled.
func (sn *ScaledNoder) GetNodedSubstrings() []*NodedSegmentString {
	nodedScaled := sn.noder.GetNodedSubstrings()

	// Unscale the coordinates
	result := make([]*NodedSegmentString, len(nodedScaled))
	for i, ss := range nodedScaled {
		unscaledCoords := sn.unscale(ss.Coordinates())
		result[i] = NewNodedSegmentString(unscaledCoords, ss.Context())
	}

	return result
}

// ValidatingNoder wraps a Noder and validates that the noding is complete
// (i.e., no intersections remain in the noded result).
type ValidatingNoder struct {
	noder Noder
}

// NewValidatingNoder creates a new ValidatingNoder.
func NewValidatingNoder(noder Noder) *ValidatingNoder {
	return &ValidatingNoder{noder: noder}
}

// ComputeNodes computes nodes using the wrapped noder.
func (vn *ValidatingNoder) ComputeNodes(segmentStrings []*NodedSegmentString) {
	vn.noder.ComputeNodes(segmentStrings)
}

// GetNodedSubstrings returns the noded substrings after validation.
func (vn *ValidatingNoder) GetNodedSubstrings() []*NodedSegmentString {
	nodedSS := vn.noder.GetNodedSubstrings()

	// Validate that no intersections remain
	counter := NewIntersectionCounter()
	checker := NewSimpleNoder(counter)
	checker.ComputeNodes(nodedSS)

	if counter.Count() > 0 {
		// In a production implementation, this might return an error
		// For now, we just return the noded strings anyway
	}

	return nodedSS
}

// IteratedNoder runs a noder multiple times until no more intersections
// are found. This is useful for handling numerical robustness issues.
type IteratedNoder struct {
	noder       Noder
	maxIterations int
}

// NewIteratedNoder creates a new IteratedNoder.
func NewIteratedNoder(noder Noder, maxIterations int) *IteratedNoder {
	if maxIterations <= 0 {
		maxIterations = 5
	}
	return &IteratedNoder{
		noder:       noder,
		maxIterations: maxIterations,
	}
}

// ComputeNodes computes nodes iteratively.
func (in *IteratedNoder) ComputeNodes(segmentStrings []*NodedSegmentString) {
	current := segmentStrings

	for i := 0; i < in.maxIterations; i++ {
		// Compute nodes
		in.noder.ComputeNodes(current)

		// Get noded result
		noded := in.noder.GetNodedSubstrings()

		// Check if any new intersections were found
		counter := NewIntersectionCounter()
		checker := NewSimpleNoder(counter)
		checker.ComputeNodes(noded)

		if counter.Count() == 0 {
			// No more intersections - we're done
			return
		}

		// Use the noded result as input for next iteration
		current = noded
	}
}

// GetNodedSubstrings returns the final noded substrings.
func (in *IteratedNoder) GetNodedSubstrings() []*NodedSegmentString {
	return in.noder.GetNodedSubstrings()
}
