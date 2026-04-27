// Package topology is the v2 API surface for go-topology-suite.
//
// The v2 API favors explicit errors for operations that can fail because of
// nil inputs, invalid geometries, unsupported operations, or strict parser
// failures. RelatePattern validates DE-9IM patterns strictly before running
// the relation operation. Buffer validates custom buffer parameters, including
// quadrant segments, cap style, join style, and mitre limits.
//
// It is currently a compatibility facade over the pure-Go v1 implementation
// while the shared topology engine is hardened behind it.
package topology
