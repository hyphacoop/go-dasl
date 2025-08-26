package cid

import (
	"errors"
	"fmt"

	"github.com/hyphacoop/cbor/v2"
)

// RawCid is an unvalidated CID.
//
// It is used to store CID information when decoding non-DASL CIDs from DRISL
// is enabled. You can also use it to encode non-DASL CIDs.
// Only do this if you are not working in a DASL-compliant ecosystem!
//
// It holds the bytes of a binary CID. It is not the same bytes as CID-in-CBOR.
//
// There are no guarantees this is a valid CID by any spec.
//
// You can pass it to NewCidFromBytes to validate it as a DASL CID, or to the
// Cast function of the go-cid library to parse it as an original IPFS CID.
type RawCid []byte

func (c RawCid) MarshalCBOR() ([]byte, error) {
	// CID in CBOR is just CID bytes with 0x00 prepended
	return cbor.Marshal(cbor.Tag{
		Number:  cidTagNumber,
		Content: append([]byte{0x00}, c...),
	})
}

func (c *RawCid) UnmarshalCBOR(b []byte) error {
	var tag cbor.Tag
	if err := cbor.Unmarshal(b, &tag); err != nil {
		return err
	}

	// Check tag number
	if tag.Number != cidTagNumber {
		return fmt.Errorf("got tag number %d, expect tag number %d", tag.Number, cidTagNumber)
	}

	// Check tag content
	cidData, isByteString := tag.Content.([]byte)
	if !isByteString {
		return fmt.Errorf("invalid cid: unmarshal: got tag content type %T, expect tag content []byte", tag.Content)
	}

	// Verify CBOR CID encoding
	if len(cidData) == 0 {
		return errors.New("invalid cid: unmarshal: no data")
	}
	if cidData[0] != 0x00 {
		return fmt.Errorf("invalid cid: unmarshal: got CBOR CID prefix 0x%02x, expect prefix 0x00", cidData[0])
	}

	// Skip 0x00 prefix and copy data
	*c = make([]byte, len(cidData[1:]))
	copy(*c, cidData[1:])
	return nil
}

func (c RawCid) String() string {
	s := multibaseBase32.EncodeToString(c)
	return "b" + s
}
