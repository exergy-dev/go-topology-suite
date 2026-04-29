package wkb

import "encoding/binary"

// Option configures the WKB encoder.
type Option func(*config)

type config struct {
	order binary.ByteOrder
	iso   bool
	srid  int // 0 = inherit from CRS, negative = suppress
}

// WithByteOrder selects the byte order. Default: little endian (NDR).
func WithByteOrder(o binary.ByteOrder) Option { return func(c *config) { c.order = o } }

// WithISO selects ISO 13249-3 encoding (separate type codes for Z/M
// variants, no SRID slot). Mutually exclusive with EWKB SRID encoding.
func WithISO() Option { return func(c *config) { c.iso = true } }

// WithSRID overrides the SRID written in the EWKB header. By default the
// SRID is taken from the geometry's CRS (Authority "EPSG", non-zero Code).
// Pass a negative value to suppress the SRID slot entirely.
func WithSRID(srid int) Option { return func(c *config) { c.srid = srid } }

func defaults() config {
	return config{order: binary.LittleEndian, srid: 0}
}
