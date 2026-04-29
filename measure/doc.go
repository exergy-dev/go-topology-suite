// Package measure computes scalar measurements (distance, length, area,
// centroid) on Terra geometries.
//
// Every measurement function is kernel-routed: pass WithKernel to override
// the default. For geographic CRSes the planar default is almost always
// wrong; until Phase 4 wires up automatic geodesic-default selection,
// callers using lon/lat data should pass WithKernel(geodesic.Default)
// explicitly for measurements that need to be in metres.
package measure
