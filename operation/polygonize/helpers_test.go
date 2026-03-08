package polygonize

import "github.com/robert-malhotra/go-topology-suite/geom"

func mustLineStringXY(values ...float64) *geom.LineString {
	seq, err := geom.NewCoordinateSequenceXY(values...)
	if err != nil {
		panic(err)
	}
	return geom.NewLineString(seq)
}
