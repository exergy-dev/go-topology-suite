// Package wkb encodes and decodes geometries in Well-Known Binary format.
//
// Two flavours are supported:
//
//   - PostGIS EWKB (default): the ubiquitous extension that ORs the type
//     code with high-bit flags for Z (0x80000000), M (0x40000000), and
//     SRID (0x20000000). When SRID is set, a 4-byte little-endian SRID
//     follows the type code.
//
//   - ISO 13249-3 (opt-in via WithISO): encodes Z/M variants as separate
//     base type codes (POINT=1, POINTZ=1001, POINTM=2001, POINTZM=3001).
//     The output is round-trippable by standards-conformant decoders that
//     reject EWKB high-bit flags.
//
// The decoder accepts both flavours, auto-detecting from the type-code
// shape. Byte order is preserved on round-trip; encode default is little
// endian (NDR) since it dominates on x86 and ARM.
package wkb
