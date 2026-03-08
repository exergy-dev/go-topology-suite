// Package ioutil provides shared utilities for geometry I/O packages.
package ioutil

import (
	"fmt"
	"strings"
)

// WriteNumber writes a float64 to a string builder with the given precision.
// A negative precision uses Go's default %g formatting.
func WriteNumber(sb *strings.Builder, n float64, precision int) {
	if precision < 0 {
		fmt.Fprintf(sb, "%g", n)
	} else {
		fmt.Fprintf(sb, "%.*f", precision, n)
	}
}
