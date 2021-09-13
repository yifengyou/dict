package main

import (
	"fmt"

	"github.com/srfrog/dict"
)

type Car struct {
	Model, BrandID string
	Recalls        int
}

func main() {
	// Map of car models, indexed by VIN.
	// Data source: NHTSA.gov
	vins := map[string]*Car{
		"2C3KA43R08H129584": &Car{
			Model:   "2008 CHRYSLER 300",
			BrandID: "ACB9976A-DB5F-4D57-B9A8-9F5C53D87C7C",
			Recalls: 1,
		},
		"1N6AD07U78C416152": &Car{
			Model:   "2008 NISSAN FRONTIER SE-V6 RWD",
			BrandID: "003096EE-C8FC-4C2F-ADEF-406F86C1F70B",
			Recalls: 5,
		},
		"WDDGF8AB8EA940372": &Car{
			Model:   "2014 Mercedes-Benz C300W4",
			BrandID: "57B7B707-4357-4306-9FD6-1EDCA43CF77B",
			Recalls: 4,
		},
	}

	// Create new dict and initialize with vins map.
	d := dict.New(vins)

	// Add a couple more VINs.
	d.Set("1N4AL2AP4BN404580", &Car{
		Model:   "2011 NISSAN ALTIMA 2.5 S CVT",
		BrandID: "003096EE-C8FC-4C2F-ADEF-406F86C1F70B",
		Recalls: 2,
	})
	d.Set("4T1BE46K48U762452", &Car{
		Model:   "2008 TOYOTA Camry",
		BrandID: "C5764FE4-F1E8-46BE-AFC6-A2FC90110387",
		Recalls: 5,
	})

	// Check current total
	fmt.Println("Total VIN Count:", d.Len())

	// Print VINs that have 3 or more recalls
	for item := range d.Items() {
		car, ok := item.Value.(*Car)
		if !ok {
			continue // Not a Car
		}
		if car.Recalls < 3 {
			continue // Not enough recalls
		}
		fmt.Println("---")
		fmt.Println("VIN:", item.Key)
	}
}
