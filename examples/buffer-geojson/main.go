// Buffer a geometry by a fixed distance and emit the result as
// RFC 7946 GeoJSON (exterior rings forced CCW).
package main

import (
	"fmt"
	"log"

	"github.com/exergy-dev/go-topology-suite/buffer"
	"github.com/exergy-dev/go-topology-suite/geojson"
	"github.com/exergy-dev/go-topology-suite/wkt"
)

func main() {
	line, err := wkt.Unmarshal("LINESTRING (0 0, 10 0, 10 10)")
	if err != nil {
		log.Fatal(err)
	}

	buffered, err := buffer.Buffer(line, 1.5)
	if err != nil {
		log.Fatal(err)
	}

	out, err := geojson.Marshal(buffered,
		geojson.WithPrecision(3),
		geojson.WithForceCCW(),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(out))
}
