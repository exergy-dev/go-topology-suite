package crs

// init initializes all common CRS, datum, and ellipsoid instances
// in the correct order to avoid nil pointer panics.
func init() {
	// Initialize in dependency order:
	// 1. Ellipsoids (no dependencies)
	initEllipsoids()

	// 2. Datums (depend on ellipsoids)
	initDatums()

	// 3. Geographic CRS (depend on datums)
	initGeographicCRS()

	// 4. Projected CRS (depend on geographic CRS)
	initProjectedCRS()
}
