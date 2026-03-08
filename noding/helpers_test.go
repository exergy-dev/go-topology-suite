package noding

import "github.com/robert-malhotra/go-topology-suite/geom"

func mustCoordsXY(values ...float64) geom.CoordinateSequence {
	seq, err := geom.NewCoordinateSequenceXY(values...)
	if err != nil {
		panic(err)
	}
	return seq
}
