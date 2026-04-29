// Package geom is the foundation of Terra: it defines the Geometry interface,
// the seven OGC Simple Features types, the coordinate value types, the Layout
// enum, and the Envelope.
//
// Every other Terra subpackage depends on geom. The interfaces here are
// stable for the lifetime of a sync gate (see the parallel implementation
// plan); breaking changes happen only at gate boundaries.
package geom
