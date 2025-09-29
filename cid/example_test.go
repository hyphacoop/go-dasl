package cid_test

import (
	"fmt"

	"github.com/hyphacoop/go-dasl/cid"
)

func Example() {
	// Create a CID from a string representation
	cidString := "bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4"
	c1, err := cid.NewCidFromString(cidString)
	if err != nil {
		panic(err)
	}
	fmt.Printf("CID from string: %s\n", c1.String())

	// Create a CID from binary bytes
	cidBytes := []byte{0x01, 0x55, 0x12, 0x20, 0xad, 0xee, 0x2e, 0x8f, 0xb5, 0x45, 0x9c, 0x9b, 0xcf, 0x07, 0xd7, 0xd7, 0x8d, 0x11, 0x83, 0xbf, 0x40, 0xa7, 0xf6, 0x0f, 0x57, 0xa5, 0x4a, 0x19, 0x19, 0x48, 0x01, 0xc9, 0xa2, 0x7e, 0xad, 0x87}
	c2, err := cid.NewCidFromBytes(cidBytes)
	if err != nil {
		panic(err)
	}
	fmt.Printf("CID from bytes: %s\n", c2.String())

	// Both should be equal
	fmt.Printf("CIDs are equal: %t\n", c1.Equal(c2))

	// Show CID properties
	fmt.Printf("Codec: %d\n", c1.Codec())
	fmt.Printf("Hash type: %d\n", c1.HashType())
	// Output:
	// CID from string: bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4
	// CID from bytes: bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4
	// CIDs are equal: true
	// Codec: 85
	// Hash type: 18
}
