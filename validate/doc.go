// Package validate detects defects in go-topology-suite geometries against the OGC
// Simple Features rules.
//
// Construction is non-validating by design: callers running
// performance-sensitive ingest paths should not pay for validation they
// did not ask for. Validate is the explicit opt-in.
//
// The reported defects in v0.1 cover the structural rules — ring closure,
// minimum vertex counts, ring orientation, hole containment. Detection of
// self-intersections is done via brute-force pairwise edge comparison;
// the Phase 3 overlay-NG port replaces it with a noding-based variant.
package validate
