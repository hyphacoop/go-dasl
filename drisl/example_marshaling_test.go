package drisl_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/hyphacoop/go-dasl/drisl"
)

type Animal int

const (
	Unknown Animal = iota
	Gopher
	Zebra
)

func (a *Animal) UnmarshalCBOR(b []byte) error {
	var s string
	if err := drisl.Unmarshal(b, &s); err != nil {
		return err
	}
	switch strings.ToLower(s) {
	default:
		*a = Unknown
	case "gopher":
		*a = Gopher
	case "zebra":
		*a = Zebra
	}
	return nil
}

func (a Animal) MarshalCBOR() ([]byte, error) {
	var s string
	switch a {
	default:
		s = "unknown"
	case Gopher:
		s = "gopher"
	case Zebra:
		s = "zebra"
	}
	return drisl.Marshal(s)
}

func Example_customMarshalCBOR() {
	// Create a slice of animals from string names
	animalNames := []string{"gopher", "armadillo", "zebra", "unknown", "gopher", "bee", "gopher", "zebra"}

	// Marshal the animal names to DRISL
	data, err := drisl.Marshal(animalNames)
	if err != nil {
		log.Fatal(err)
	}

	// Unmarshal into a slice of Animal enums (using custom UnmarshalCBOR)
	var zoo []Animal
	if err := drisl.Unmarshal(data, &zoo); err != nil {
		log.Fatal(err)
	}

	// Count the animals
	census := make(map[Animal]int)
	for _, animal := range zoo {
		census[animal] += 1
	}

	fmt.Printf("Zoo Census:\n* Gophers: %d\n* Zebras:  %d\n* Unknown: %d\n",
		census[Gopher], census[Zebra], census[Unknown])

	// Output:
	// Zoo Census:
	// * Gophers: 3
	// * Zebras:  2
	// * Unknown: 3
}
