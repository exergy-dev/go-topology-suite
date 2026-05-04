// Package crs models coordinate reference systems as data, not metadata.
//
// Every Terra geometry carries a *CRS pointer. Operations that mix
// geometries with incompatible CRS return ErrCRSMismatch (defined in the
// top-level terra package) rather than silently producing nonsense; users
// must call terra.Transform explicitly to reconcile. The crs package
// itself supplies the underlying Operation graph (OperationFor) and the
// projection / datum primitives; terra.Transform is the geometry-aware
// wrapper.
//
// The Authority+Code form ("EPSG", 4326) is the canonical identity used
// for equality. The optional WKT2 string defines CRSes outside the built-in
// registry; equality on those is structural over the WKT2 source.
package crs
