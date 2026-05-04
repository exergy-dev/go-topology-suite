// Package index provides go-topology-suite's in-memory spatial index — an R-tree
// generic over the payload type. The index does not store geometries; it
// indexes envelopes plus a user-supplied value, leaving geometry ownership
// with the caller.
//
// The implementation is a basic R-tree with linear node split (Guttman 1984).
// A future iteration will switch to R*-tree splits and add STR bulk-load;
// the public API is stable across that internal change.
package index
