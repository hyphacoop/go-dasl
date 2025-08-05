package drisl

import (
	"errors"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-cid"
)

const CidTagNumber = 42

// Cid is go-cid with CBOR marshalling support.
type Cid struct {
	cid.Cid
}

// TODO: DASL CID checks

func (c Cid) MarshalCBOR() ([]byte, error) {
	if !c.Defined() {
		return nil, errors.New("undefined CID")
	}
	// CID in CBOR is just CID bytes with 0x00 prepended
	return cbor.Marshal(cbor.Tag{
		Number:  CidTagNumber,
		Content: append([]byte{0x00}, c.Bytes()...),
	})
}

func (c *Cid) UnmarshalCBOR(b []byte) error {
	var tag cbor.Tag
	if err := cbor.Unmarshal(b, &tag); err != nil {
		return err
	}

	// Check tag number
	if tag.Number != CidTagNumber {
		return fmt.Errorf("got tag number %d, expect tag number %d", tag.Number, CidTagNumber)
	}

	// Check tag content
	cidData, isByteString := tag.Content.([]byte)
	if !isByteString {
		return fmt.Errorf("got tag content type %T, expect tag content []byte", tag.Content)
	}

	// Verify CBOR CID encoding
	if len(cidData) == 0 {
		return errors.New("zero-length CID")
	}
	if cidData[0] != 0x00 {
		return fmt.Errorf("got CBOR CID prefix 0x%02x, expect prefix 0x00", cidData[0])
	}

	parsed, err := cid.Cast(cidData)
	if err != nil {
		return err
	}
	*c = Cid{parsed}
	return nil
}
