package buffer

// CapStyle selects how the ends of an open LineString are closed.
type CapStyle int

const (
	// CapRound closes each endpoint with a semicircular arc whose radius
	// equals the buffer distance.
	CapRound CapStyle = iota
	// CapFlat (a.k.a. butt cap) terminates the buffer flush with the line's
	// endpoint — the cap is a straight segment perpendicular to the
	// terminal edge.
	CapFlat
	// CapSquare extends the buffer past the endpoint by the buffer distance
	// before closing with a perpendicular segment.
	CapSquare
)

// JoinStyle selects how the buffer turns at an interior vertex.
type JoinStyle int

const (
	// JoinRound rounds the corner with a circular arc of radius equal to
	// the buffer distance.
	JoinRound JoinStyle = iota
	// JoinMitre extends the two offset edges until they meet. If the
	// extension exceeds MitreLimit * distance, the join falls back to a
	// bevel.
	JoinMitre
	// JoinBevel cuts the corner with a straight segment connecting the two
	// offset endpoints.
	JoinBevel
)

// config holds resolved buffer options. Constructed via Option functions.
type config struct {
	cap          CapStyle
	join         JoinStyle
	mitreLimit   float64
	quadSegments int
}

// Option mutates a config. The zero config is invalid; defaultConfig
// supplies sensible starting values.
type Option func(*config)

func defaultConfig() config {
	return config{
		cap:          CapRound,
		join:         JoinRound,
		mitreLimit:   5.0,
		quadSegments: 8,
	}
}

// WithCapStyle sets the line endpoint cap style. Default: CapRound.
func WithCapStyle(s CapStyle) Option { return func(c *config) { c.cap = s } }

// WithJoinStyle sets the corner join style. Default: JoinRound.
func WithJoinStyle(s JoinStyle) Option { return func(c *config) { c.join = s } }

// WithMitreLimit sets the maximum mitre extension as a multiple of the
// buffer distance. Joins exceeding this limit fall back to bevel. Values
// ≤ 0 are ignored. Default: 5.0.
func WithMitreLimit(l float64) Option {
	return func(c *config) {
		if l > 0 {
			c.mitreLimit = l
		}
	}
}

// WithQuadSegments sets the number of straight-line segments used to
// approximate a quarter circle in round caps and joins. Values < 1 are
// clamped to 1. Default: 8.
func WithQuadSegments(n int) Option {
	return func(c *config) {
		if n < 1 {
			n = 1
		}
		c.quadSegments = n
	}
}
