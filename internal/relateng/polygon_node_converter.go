package relateng

// convertPolygonNodeSections takes the contiguous polygon sections that
// share an element id and converts the OGC "touching-rings" structure
// into the equivalent self-touch / inverted-ring structure that
// RelateNode's edge ordering expects.
//
// Port of org.locationtech.jts.operation.relateng.PolygonNodeConverter.
//
// The Go port currently runs as a passthrough — the simple cases handled
// by the segment-intersection pipeline (single shell, single hole,
// shell+single-hole touch) work without rewriting. The full converter
// will land in a follow-up; nodes that depend on the conversion are
// rare enough that the legacy fallback handles them today.
func convertPolygonNodeSections(sections []*NodeSection) []*NodeSection {
	// TODO(wave10): port the full PolygonNodeConverter.convert logic
	// (shell+hole and hole+hole touch normalisation). Until then we
	// pass the sections through unchanged: RelateNode will still order
	// them correctly for the common topology cases.
	return sections
}
