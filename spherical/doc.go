// Package spherical provides spherical geometry operations for geographic coordinates.
//
// This package integrates with Google's S2 geometry library to provide accurate
// spherical calculations for WGS84 coordinates. It handles:
//   - Geodesic distance calculations
//   - Spherical polygon area
//   - Point-in-polygon tests on the sphere
//   - S2 cell indexing for spatial queries
//
// All coordinates are expected in WGS84 (EPSG:4326) with longitude as X and latitude as Y.
package spherical
