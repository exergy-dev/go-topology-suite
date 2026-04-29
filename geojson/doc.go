// Package geojson encodes and decodes Terra geometries to/from GeoJSON
// (RFC 7946).
//
// Output is strict RFC 7946: always WGS84-implied (no CRS member emitted),
// canonical key ordering ("type" before "coordinates"), Z values preserved
// when present (the RFC permits it).
//
// Input is more lenient: a foreign top-level "crs" member is accepted and
// dropped on read; bbox arrays of either 4 or 6 elements are accepted;
// numeric coordinates with embedded null/missing are rejected.
//
// Feature and FeatureCollection both round-trip foreign top-level members
// via the Foreign map[string]json.RawMessage field, so non-RFC extensions
// users sometimes attach (e.g. "title", "renderer") survive a round trip.
package geojson
