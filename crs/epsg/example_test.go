package epsg_test

import (
	"fmt"

	"github.com/robert-malhotra/go-topology-suite/crs/epsg"
)

// ExampleLookup demonstrates looking up a CRS by EPSG code.
func ExampleLookup() {
	crs, err := epsg.Lookup(4326)
	if err != nil {
		panic(err)
	}

	fmt.Println(crs.Name())
	fmt.Println(crs.Code())
	fmt.Println(crs.IsGeographic())

	// Output:
	// WGS 84
	// EPSG:4326
	// true
}

// ExampleMustLookup demonstrates the panic-on-error lookup.
func ExampleMustLookup() {
	crs := epsg.MustLookup(4326)
	fmt.Println(crs.Name())

	// Output:
	// WGS 84
}

// ExampleUTMZone demonstrates generating UTM zone CRS.
func ExampleUTMZone() {
	// Generate UTM zone 10 North (US West Coast)
	utm10n, err := epsg.UTMZone(10, true)
	if err != nil {
		panic(err)
	}
	fmt.Println(utm10n.Code())
	fmt.Println(utm10n.Name())

	// Generate UTM zone 50 South (Australia)
	utm50s, err := epsg.UTMZone(50, false)
	if err != nil {
		panic(err)
	}
	fmt.Println(utm50s.Code())

	// Output:
	// EPSG:32610
	// WGS 84 / UTM zone 10N
	// EPSG:32750
}

// ExampleCodes demonstrates listing all registered EPSG codes.
func ExampleCodes() {
	codes := epsg.Codes()
	fmt.Printf("Total CRS registered: %d\n", len(codes))
	fmt.Printf("First three codes: %v\n", codes[:3])

	// Output:
	// Total CRS registered: 8
	// First three codes: [3857 4258 4267]
}

// ExampleIsRegistered demonstrates checking if a code is registered.
func ExampleIsRegistered() {
	fmt.Println(epsg.IsRegistered(4326))
	fmt.Println(epsg.IsRegistered(99999))

	// Output:
	// true
	// false
}

// Example demonstrates comprehensive usage of the EPSG package.
func Example() {
	// Use predefined geographic CRS
	wgs84 := epsg.WGS84
	fmt.Printf("Name: %s\n", wgs84.Name())
	fmt.Printf("Code: %s\n", wgs84.Code())
	fmt.Printf("Is Geographic: %v\n", wgs84.IsGeographic())

	// Use predefined projected CRS
	webMercator := epsg.WebMercator
	fmt.Printf("\nWeb Mercator: %s\n", webMercator.Name())
	fmt.Printf("Is Projected: %v\n", !webMercator.IsGeographic())

	// Look up by code
	nad83, _ := epsg.Lookup(4269)
	fmt.Printf("\nNAD83: %s\n", nad83.Name())

	// Generate UTM zone
	utm32n, _ := epsg.UTMZone(32, true)
	fmt.Printf("\nUTM 32N: %s\n", utm32n.Code())

	// List registered codes
	fmt.Printf("\nTotal registered: %d\n", epsg.Count())

	// Output:
	// Name: WGS 84
	// Code: EPSG:4326
	// Is Geographic: true
	//
	// Web Mercator: WGS 84 / Pseudo-Mercator
	// Is Projected: true
	//
	// NAD83: NAD83
	//
	// UTM 32N: EPSG:32632
	//
	// Total registered: 8
}
