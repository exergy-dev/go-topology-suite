// Package epsg provides a registry of well-known coordinate reference
// systems addressed by EPSG code.
//
// The registry is populated at package init time. Use Lookup to retrieve a
// *crs.CRS by EPSG code, or reference one of the exported variables (for
// example WGS84, BritishNationalGrid, Lambert93) directly.
//
// This package depends on package crs but adds no runtime overhead beyond
// the construction of its built-in registry. CRS values returned by Lookup
// share the same underlying *crs.CRS instances as the exported variables;
// callers must treat them as read-only.
package epsg
