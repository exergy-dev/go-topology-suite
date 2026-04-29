// Package simplify reduces vertex counts in geometries while preserving
// shape within a tolerance.
//
// Simplify uses Douglas-Peucker recursion: for any sequence of vertices,
// the vertex farthest from the segment connecting the endpoints is kept
// if its perpendicular distance exceeds the tolerance, and the algorithm
// recurses on each half. Otherwise all interior vertices collapse to the
// pair of endpoints.
//
// Simplify itself does NOT preserve topology — the simplified geometry
// can self-intersect or invert ring orientation. Callers needing
// guaranteed-simple output should use TopologyPreserving, which uses
// Visvalingam-Whyatt vertex elimination with a per-removal crossing
// safety check (vertices whose removal would introduce a self-
// intersection are kept).
package simplify
