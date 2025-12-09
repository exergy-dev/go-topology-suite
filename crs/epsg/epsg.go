// Package epsg provides EPSG (European Petroleum Survey Group) coordinate
// reference system definitions and a registry for looking up CRS by code.
//
// The EPSG dataset is the de facto standard registry of coordinate reference
// systems used in GIS and surveying. Each CRS is identified by a unique integer
// code (e.g., 4326 for WGS84).
//
// This package includes the most commonly used CRS definitions:
//   - Geographic: WGS84 (4326), NAD83 (4269), NAD27 (4267), ETRS89 (4258)
//   - Projected: Web Mercator (3857), UTM zones (326xx, 327xx)
//
// Example usage:
//
//	// Look up a CRS by EPSG code
//	crs, err := epsg.Lookup(4326)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(crs.Name()) // "WGS 84"
//
//	// Use predefined CRS
//	wgs84 := epsg.WGS84
//	fmt.Println(wgs84.IsGeographic()) // true
//
//	// Get a UTM zone CRS
//	utm10n := epsg.UTMZone(10, true)
//	fmt.Println(utm10n.Code()) // "EPSG:32610"
package epsg

import (
	"fmt"
	"sort"
	"sync"

	"github.com/go-topology-suite/gts/crs"
)

var (
	// registry holds all registered CRS by their EPSG code.
	registry = make(map[int]crs.CRS)

	// mu protects the registry from concurrent access.
	mu sync.RWMutex
)

func init() {
	// Note: Registration is done lazily on first access to avoid
	// initialization order issues. The init functions in geographic.go
	// and projected.go will register their CRS definitions.
}

// registerCRS adds a CRS to the registry.
// Extracts the numeric code from the "EPSG:XXXX" string.
func registerCRS(c crs.CRS) {
	code := extractCode(c.Code())
	if code == 0 {
		return // Don't register CRS without a valid EPSG code
	}
	registry[code] = c
}

// extractCode extracts the numeric EPSG code from a string like "EPSG:4326".
// Returns 0 if the code cannot be parsed.
func extractCode(codeStr string) int {
	var code int
	_, err := fmt.Sscanf(codeStr, "EPSG:%d", &code)
	if err != nil {
		return 0
	}
	return code
}

// Lookup finds a CRS by its EPSG code.
//
// Returns an error if the code is not found in the registry.
// For codes not pre-registered, consider using Register to add custom CRS.
//
// Example:
//
//	crs, err := epsg.Lookup(4326)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(crs.Name()) // "WGS 84"
func Lookup(code int) (crs.CRS, error) {
	mu.RLock()
	defer mu.RUnlock()

	c, ok := registry[code]
	if !ok {
		return nil, fmt.Errorf("EPSG code %d not found in registry", code)
	}
	return c, nil
}

// MustLookup finds a CRS by its EPSG code.
//
// Panics if the code is not found in the registry.
// Use this only when you are certain the code exists (e.g., for well-known codes).
//
// Example:
//
//	wgs84 := epsg.MustLookup(4326)
//	fmt.Println(wgs84.Name()) // "WGS 84"
func MustLookup(code int) crs.CRS {
	c, err := Lookup(code)
	if err != nil {
		panic(err)
	}
	return c
}

// Register adds a custom CRS to the registry.
//
// This allows applications to define and register their own CRS definitions
// that can then be looked up by EPSG code.
//
// If a CRS with the same code already exists, it will be replaced.
// Returns an error if the CRS code cannot be parsed.
//
// Example:
//
//	customCRS, _ := crs.NewProjectedCRS("EPSG:2154", "RGF93 / Lambert-93", ...)
//	epsg.Register(customCRS)
func Register(c crs.CRS) error {
	code := extractCode(c.Code())
	if code == 0 {
		return fmt.Errorf("cannot register CRS with invalid code: %s", c.Code())
	}

	mu.Lock()
	defer mu.Unlock()

	registry[code] = c
	return nil
}

// Unregister removes a CRS from the registry by its EPSG code.
//
// This is useful for testing or when you need to replace a CRS definition.
// Returns an error if the code is not found.
//
// Example:
//
//	err := epsg.Unregister(2154)
//	if err != nil {
//	    log.Println(err)
//	}
func Unregister(code int) error {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := registry[code]; !ok {
		return fmt.Errorf("EPSG code %d not found in registry", code)
	}

	delete(registry, code)
	return nil
}

// Codes returns a sorted list of all EPSG codes in the registry.
//
// The returned slice is a copy, so modifying it will not affect the registry.
//
// Example:
//
//	codes := epsg.Codes()
//	fmt.Println(codes) // [3857 4258 4267 4269 4326 32610 32617 32632]
func Codes() []int {
	mu.RLock()
	defer mu.RUnlock()

	codes := make([]int, 0, len(registry))
	for code := range registry {
		codes = append(codes, code)
	}
	sort.Ints(codes)
	return codes
}

// Count returns the number of CRS definitions in the registry.
//
// Example:
//
//	count := epsg.Count()
//	fmt.Printf("Registry contains %d CRS definitions\n", count)
func Count() int {
	mu.RLock()
	defer mu.RUnlock()
	return len(registry)
}

// IsRegistered returns true if the given EPSG code is in the registry.
//
// Example:
//
//	if epsg.IsRegistered(4326) {
//	    fmt.Println("WGS84 is registered")
//	}
func IsRegistered(code int) bool {
	mu.RLock()
	defer mu.RUnlock()
	_, ok := registry[code]
	return ok
}
