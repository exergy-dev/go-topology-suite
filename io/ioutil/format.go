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
		sb.WriteString(fmt.Sprintf("%g", n))
	} else {
		sb.WriteString(fmt.Sprintf("%.*f", precision, n))
	}
}
