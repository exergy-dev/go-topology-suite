// Reproject a geometry between coordinate reference systems. The
// library never reprojects implicitly: predicates and overlays on
// geometries with mismatched CRS pointers return gts.ErrCRSMismatch.
package main

import (
	"fmt"
	"log"

	gts "github.com/exergy-dev/go-topology-suite"
	"github.com/exergy-dev/go-topology-suite/crs/epsg"
	"github.com/exergy-dev/go-topology-suite/geom"
)

func main() {
	// New York City in WGS84 lon/lat.
	nyc := geom.NewPoint(epsg.WGS84, geom.XY{X: -74.0060, Y: 40.7128})

	mercator, err := gts.Transform(nyc, epsg.WebMercator)
	if err != nil {
		log.Fatal(err)
	}

	xy := mercator.(*geom.Point).XY()
	fmt.Printf("NYC in EPSG:4326: %.4f, %.4f\n", -74.0060, 40.7128)
	fmt.Printf("NYC in EPSG:3857: %.2f, %.2f (metres)\n", xy.X, xy.Y)

	roundTrip, err := gts.Transform(mercator, epsg.WGS84)
	if err != nil {
		log.Fatal(err)
	}
	rt := roundTrip.(*geom.Point).XY()
	fmt.Printf("Round-trip back:  %.4f, %.4f\n", rt.X, rt.Y)
}
