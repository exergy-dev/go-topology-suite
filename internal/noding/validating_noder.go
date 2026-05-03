package noding

// ValidatingNoder wraps a base Noder and runs NodingValidator over its
// output. If validation fails the cached error is exposed via Err().
//
// Mirrors org.locationtech.jts.noding.ValidatingNoder. The JTS class
// throws a runtime exception on bad noding; we surface the failure
// through the Err accessor so the caller can decide how to react
// (panic, fall back, retry with a different noder, etc.).
type ValidatingNoder struct {
	inner Noder
	err   error
}

// NewValidatingNoder wraps inner with post-noding validation.
func NewValidatingNoder(inner Noder) *ValidatingNoder {
	return &ValidatingNoder{inner: inner}
}

// Node satisfies the Noder interface. The returned slice is whatever
// the inner noder produced; check Err() to see whether the result was
// well-noded.
func (n *ValidatingNoder) Node(input []*SegmentString) []*SegmentString {
	out := n.inner.Node(input)
	n.err = NewNodingValidator(out).CheckValid()
	return out
}

// Err returns the validation error from the most recent Node call, or
// nil if the noding was valid.
func (n *ValidatingNoder) Err() error { return n.err }
