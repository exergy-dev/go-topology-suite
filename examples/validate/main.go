// Validate a geometry, then repair a deliberately broken polygon with
// validate.Fix. Fix produces a valid geometry by snapping, splitting,
// and re-noding rings; the original is unchanged.
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/validate"
	"github.com/exergy-dev/go-topology-suite/wkt"
)

func main() {
	good := mustWKT("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	if err := validate.Validate(good); err != nil {
		log.Fatalf("expected good polygon to validate: %v", err)
	}
	fmt.Println("good polygon: valid")

	// Bowtie: the ring crosses itself.
	bad := mustWKT("POLYGON ((0 0, 10 10, 10 0, 0 10, 0 0))")
	err := validate.Validate(bad)
	var verr *validate.ValidationError
	if errors.As(err, &verr) {
		fmt.Printf("bad polygon: invalid — %s\n", verr.Error())
	}

	fixed := validate.Fix(bad)
	fixedText, err := wkt.Marshal(fixed)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("fixed polygon: %s\n", fixedText)

	if err := validate.Validate(fixed); err != nil {
		log.Fatalf("Fix() output is still invalid: %v", err)
	}
	fmt.Println("fixed polygon: valid")
}

func mustWKT(s string) geom.Geometry {
	g, err := wkt.Unmarshal(s)
	if err != nil {
		log.Fatalf("wkt.Unmarshal(%q): %v", s, err)
	}
	return g
}
