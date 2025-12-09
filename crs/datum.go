package crs

import "fmt"

// datum is the default implementation of the Datum interface.
type datum struct {
	name           string
	ellipsoid      Ellipsoid
	primeMeridian  float64
	toWGS84Params  [7]float64 // dx, dy, dz, rx, ry, rz, ds
	hasTransform   bool
}

// NewDatum creates a new datum with the given parameters.
// The toWGS84Params slice should contain 7 values for the Helmert transformation:
//   - dx, dy, dz: translation in meters
//   - rx, ry, rz: rotation in arc-seconds
//   - ds: scale factor in parts per million
// If toWGS84Params is nil or empty, no transformation parameters are set.
func NewDatum(name string, ellipsoid Ellipsoid, primeMeridian float64, toWGS84Params []float64) (Datum, error) {
	if ellipsoid == nil {
		return nil, fmt.Errorf("ellipsoid cannot be nil")
	}

	d := &datum{
		name:          name,
		ellipsoid:     ellipsoid,
		primeMeridian: primeMeridian,
	}

	if len(toWGS84Params) > 0 {
		if len(toWGS84Params) != 7 {
			return nil, fmt.Errorf("toWGS84Params must have 7 elements, got %d", len(toWGS84Params))
		}
		copy(d.toWGS84Params[:], toWGS84Params)
		d.hasTransform = true
	}

	return d, nil
}

// Name returns the name of the datum.
func (d *datum) Name() string {
	return d.name
}

// Ellipsoid returns the ellipsoid used by this datum.
func (d *datum) Ellipsoid() Ellipsoid {
	return d.ellipsoid
}

// PrimeMeridian returns the longitude of the prime meridian in degrees.
func (d *datum) PrimeMeridian() float64 {
	return d.primeMeridian
}

// ToWGS84Params returns the 7-parameter Helmert transformation to WGS84.
func (d *datum) ToWGS84Params() (dx, dy, dz, rx, ry, rz, ds float64) {
	return d.toWGS84Params[0], d.toWGS84Params[1], d.toWGS84Params[2],
		d.toWGS84Params[3], d.toWGS84Params[4], d.toWGS84Params[5],
		d.toWGS84Params[6]
}

// Common datums used worldwide.
var (
	// WGS84Datum is the World Geodetic System 1984 datum.
	// Used by GPS and modern geographic systems. No transformation needed to itself.
	WGS84Datum Datum

	// NAD83Datum is the North American Datum 1983.
	// Used in North America. Very close to WGS84.
	NAD83Datum Datum

	// NAD27Datum is the North American Datum 1927.
	// Historical datum used in North America before NAD83.
	NAD27Datum Datum

	// OSGB36Datum is the Ordnance Survey Great Britain 1936 datum.
	// Used in the United Kingdom.
	OSGB36Datum Datum

	// ED50Datum is the European Datum 1950.
	// Historical datum used in Europe.
	ED50Datum Datum

	// TokyoDatum is the Tokyo datum.
	// Used in Japan and parts of East Asia.
	TokyoDatum Datum
)

// initDatums initializes the common datum instances.
// This is called from init.go after ellipsoids are initialized.
func initDatums() {
	var err error

	// WGS 84 datum - no transformation to itself
	WGS84Datum, err = NewDatum("WGS 84", WGS84Ellipsoid, 0.0, nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create WGS84 datum: %v", err))
	}

	// NAD83 datum - uses GRS80 ellipsoid, very close to WGS84
	// Transformation parameters are approximate (effectively zero for many purposes)
	NAD83Datum, err = NewDatum("NAD83", GRS80Ellipsoid, 0.0,
		[]float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0})
	if err != nil {
		panic(fmt.Sprintf("failed to create NAD83 datum: %v", err))
	}

	// NAD27 datum - uses Clarke 1866 ellipsoid
	// Transformation to WGS84 (CONUS mean values)
	NAD27Datum, err = NewDatum("NAD27", Clarke1866Ellipsoid, 0.0,
		[]float64{-8.0, 160.0, 176.0, 0.0, 0.0, 0.0, 0.0})
	if err != nil {
		panic(fmt.Sprintf("failed to create NAD27 datum: %v", err))
	}

	// OSGB36 datum - uses Airy 1830 ellipsoid
	// Transformation to WGS84 (OSGB36 to ETRS89 approximation)
	OSGB36Datum, err = NewDatum("OSGB 1936", Airy1830Ellipsoid, 0.0,
		[]float64{446.448, -125.157, 542.060, 0.1502, 0.2470, 0.8421, -20.4894})
	if err != nil {
		panic(fmt.Sprintf("failed to create OSGB36 datum: %v", err))
	}

	// ED50 datum - uses International 1924 ellipsoid
	// Transformation to WGS84 (mean values for Europe)
	ED50Datum, err = NewDatum("ED50", International1924Ellipsoid, 0.0,
		[]float64{-87.0, -98.0, -121.0, 0.0, 0.0, 0.0, 0.0})
	if err != nil {
		panic(fmt.Sprintf("failed to create ED50 datum: %v", err))
	}

	// Tokyo datum - uses Bessel 1841 ellipsoid
	// Transformation to WGS84 (mean values for Japan)
	TokyoDatum, err = NewDatum("Tokyo", Bessel1841Ellipsoid, 0.0,
		[]float64{-146.414, 507.337, 680.507, 0.0, 0.0, 0.0, 0.0})
	if err != nil {
		panic(fmt.Sprintf("failed to create Tokyo datum: %v", err))
	}
}
