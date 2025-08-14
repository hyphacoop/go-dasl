package drisl

import (
	"errors"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-cid"
)

const CidTagNumber = 42

type ForbiddenCidError struct {
	c Cid
}

func (e *ForbiddenCidError) Error() string {
	return fmt.Sprintf("invalid cid: does not conform to DASL CID specification: %s", e.c.String())
}

// Cid is go-cid with CBOR marshalling support.
type Cid struct {
	cid.Cid
}

func NewCidFromBytes(b []byte) (Cid, error) {
	c, err := cid.Cast(b)
	if err != nil {
		return Cid{}, err
	}
	dc := Cid{c}
	if !dc.isDASl() {
		return Cid{}, &ForbiddenCidError{dc}
	}
	return dc, nil
}

func NewCidFromString(s string) (Cid, error) {
	c, err := cid.Decode(s)
	if err != nil {
		return Cid{}, err
	}
	dc := Cid{c}
	if !dc.isDASl() {
		return Cid{}, &ForbiddenCidError{dc}
	}
	return dc, nil
}

func (c Cid) isDASl() bool {
	// https://dasl.ing/cid.html
	// https://dasl.ing/bdasl.html

	if !c.Defined() {
		return false
	}
	if c.Version() != 1 {
		return false
	}
	prefix := c.Prefix()
	if prefix.Codec != 0x55 && prefix.Codec != 0x71 {
		return false
	}
	if prefix.MhType != 0x12 && prefix.MhType != 0x1e {
		return false
	}
	return true
}

func (c Cid) MarshalCBOR() ([]byte, error) {
	if !c.isDASl() {
		return nil, &ForbiddenCidError{c}
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
		return fmt.Errorf("invalid cid: unmarshal: got tag content type %T, expect tag content []byte", tag.Content)
	}

	// Verify CBOR CID encoding
	if len(cidData) == 0 {
		return errors.New("invalid cid: unmarshal: no data")
	}
	if cidData[0] != 0x00 {
		return fmt.Errorf("invalid cid: unmarshal: got CBOR CID prefix 0x%02x, expect prefix 0x00", cidData[0])
	}

	// Skip 0x00 prefix
	parsed, err := cid.Cast(cidData[1:])
	if err != nil {
		return err
	}
	*c = Cid{parsed}

	if !c.isDASl() {
		return &ForbiddenCidError{*c}
	}

	return nil
}
