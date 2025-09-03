/*
Package cid is an implementation of Content Identifiers, specifically the restricted DASL subset of CIDs.

https://dasl.ing/cid.html
*/
package cid

import (
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/hyphacoop/cbor/v2"
)

// Codec is the encoding of the data represented by the CID.
type Codec byte

// HashType is the algorithm of the hash digest embedded in the CID.
type HashType byte

const (
	// All DASL CIDs are 36 bytes long
	// 4 header bytes + 32 hash digest bytes
	CidBinaryLength = 36

	// Encoded as a base32 string it is 58 bytes, plus the 'b' prefix.
	CidStrLength = 59

	cidTagNumber = 42

	CidVersion              = 0x01
	CodecRaw       Codec    = 0x55
	CodecDrisl     Codec    = 0x71
	HashTypeSha256 HashType = 0x12
	HashTypeBlake3 HashType = 0x1e

	// The hash digest length
	HashSize = 32

	// The hash digest starts on the 5th byte (index 4)
	// After version, codec, hash type, hash size
	dgIdx = 4
)

var (
	// https://github.com/multiformats/multibase
	// The encoding referred to by "base32" or "b".
	// "RFC4648 case-insensitive - no padding"
	multibaseBase32 = base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567").WithPadding(base32.NoPadding)

	// EmptyCid is a raw SHA-256 CID that represents no data.
	// Its digest is the SHA-256 hash of the empty byte string "".
	// Try not to use it as a sentinel value.
	EmptyCid = Cid{b: *(*[CidBinaryLength]byte)(hexDecode("01551220e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"))}
)

func hexDecode(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

// ForbiddenCidError is returned if the CID is invalid in some way according to the DRISL spec.
type ForbiddenCidError struct {
	msgOrCid string
}

func (e *ForbiddenCidError) Error() string {
	return fmt.Sprintf("invalid cid: does not conform to DASL CID specification: %s", e.msgOrCid)
}

// Cid is a DASL CID.
// It is always valid, unless created directly such as cid.Cid{} or new(cid.Cid).
// That will cause invalid output if methods are called upon it, unless otherwise documented.
//
// https://dasl.ing/cid.html
type Cid struct {
	// b is the binary CID data
	b [CidBinaryLength]byte
}

// NewCidFromBytes creates a new DASL CID from the given bytes.
// A ForbiddenCidError is returned if it is invalid in any way.
//
// Note this is not the same as the bytes for a CID encoded in DRISL (CBOR).
//
// The input data is copied and so can be modified afterward.
func NewCidFromBytes(in []byte) (Cid, error) {
	// Follow these steps:
	// https://dasl.ing/cid.html
	if len(in) != CidBinaryLength {
		return Cid{}, &ForbiddenCidError{"invalid length"}
	}
	if in[0] != CidVersion {
		return Cid{}, &ForbiddenCidError{"invalid version"}
	}
	if in[1] != byte(CodecRaw) && in[1] != byte(CodecDrisl) {
		return Cid{}, &ForbiddenCidError{"invalid codec"}
	}
	if in[2] != byte(HashTypeSha256) && in[2] != byte(HashTypeBlake3) {
		return Cid{}, &ForbiddenCidError{"invalid hash type"}
	}
	if in[3] != HashSize {
		return Cid{}, &ForbiddenCidError{"invalid hash size"}
	}

	// Remaining data is hash digest. It's length is already verified.

	// Create Cid, copying input data so it can't be changed by the caller
	var b [CidBinaryLength]byte
	copy(b[:], in)
	return Cid{b}, nil
}

type ReadByteReader interface {
	io.Reader
	io.ByteReader
}

// NewCidFromReader reads a binary DASL CID from the given reader.
// A ForbiddenCidError is returned if it is invalid in any way.
// Extra data after the CID is allowed.
//
// Note this is not the same as the bytes for a CID encoded in DRISL (CBOR).
//
// If your reader does not support io.ByteReader, you can easily fulfill this
// interface by wrapping it with bufio.NewReader.
func NewCidFromReader(r ReadByteReader) (Cid, error) {
	fixErr := func(e error) error {
		if e == io.EOF {
			return io.ErrUnexpectedEOF
		}
		return e
	}

	cid := Cid{}

	b, err := r.ReadByte()
	if err != nil {
		return Cid{}, fixErr(err)
	}
	if b != CidVersion {
		return Cid{}, &ForbiddenCidError{"invalid version"}
	}
	cid.b[0] = b

	b, err = r.ReadByte()
	if err != nil {
		return Cid{}, fixErr(err)
	}
	if b != byte(CodecRaw) && b != byte(CodecDrisl) {
		return Cid{}, &ForbiddenCidError{"invalid codec"}
	}
	cid.b[1] = b

	b, err = r.ReadByte()
	if err != nil {
		return Cid{}, fixErr(err)
	}
	if b != byte(HashTypeSha256) && b != byte(HashTypeBlake3) {
		return Cid{}, &ForbiddenCidError{"invalid hash type"}
	}
	cid.b[2] = b

	b, err = r.ReadByte()
	if err != nil {
		return Cid{}, fixErr(err)
	}
	if b != HashSize {
		return Cid{}, &ForbiddenCidError{"invalid hash size"}
	}
	cid.b[3] = b

	digest := make([]byte, HashSize)
	_, err = io.ReadFull(r, digest)
	if err != nil {
		return Cid{}, fixErr(err)
	}
	copy(cid.b[dgIdx:], digest)

	return cid, nil
}

// NewCidFromString creates a new DASL CID from the given string.
// A ForbiddenCidError is returned if it is invalid DASL.
func NewCidFromString(s string) (Cid, error) {
	if len(s) != CidStrLength {
		return Cid{}, &ForbiddenCidError{"invalid length"}
	}
	if s[0] != 'b' {
		return Cid{}, &ForbiddenCidError{"not base32 encoded"}
	}
	b, err := multibaseBase32.DecodeString(s[1:])
	if err != nil {
		return Cid{}, err
	}
	return NewCidFromBytes(b)
}

// NewCidFromInfo creates a DASL CID manually from the codec, hash type, and hash digest.
// An error is only returned if the codec or hashType provided don't conform to DASL.
func NewCidFromInfo(codec Codec, hashType HashType, digest [HashSize]byte) (Cid, error) {
	if codec != CodecRaw && codec != CodecDrisl {
		return Cid{}, &ForbiddenCidError{"invalid codec"}
	}
	if hashType != HashTypeSha256 && hashType != HashTypeBlake3 {
		return Cid{}, &ForbiddenCidError{"invalid hash type"}
	}
	cid := Cid{}
	cid.b[0] = CidVersion
	cid.b[1] = byte(codec)
	cid.b[2] = byte(hashType)
	cid.b[3] = HashSize
	copy(cid.b[dgIdx:], digest[:])
	return cid, nil
}

// MustNewCidFromString calls NewCidFromString and panics if it returns an error.
func MustNewCidFromString(s string) Cid {
	c, err := NewCidFromString(s)
	if err != nil {
		panic(err)
	}
	return c
}

// MustNewCidFromBytes calls NewCidFromBytes and panics if it returns an error.
func MustNewCidFromBytes(b []byte) Cid {
	c, err := NewCidFromBytes(b)
	if err != nil {
		panic(err)
	}
	return c
}

// HashBytes creates a raw SHA-256 CID by hashing the provided bytes.
func HashBytes(b []byte) Cid {
	digest := sha256.Sum256(b)
	// Quick version of NewCidFromInfo
	cid := Cid{[CidBinaryLength]byte{CidVersion, byte(CodecRaw), byte(HashTypeSha256), HashSize}}
	copy(cid.b[dgIdx:], digest[:])
	return cid
}

// HashReader creates a raw SHA-256 CID by hashing all the data in the reader.
// Any error returned comes from the reader.
func HashReader(r io.Reader) (Cid, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, r); err != nil {
		return Cid{}, err
	}
	// Quick version of NewCidFromInfo
	cid := Cid{[CidBinaryLength]byte{CidVersion, byte(CodecRaw), byte(HashTypeSha256), HashSize}}
	copy(cid.b[dgIdx:], hasher.Sum(nil))
	return cid, nil
}

// Bytes returns the CID in binary format.
// It is safe to modify. It is always CidBinaryLength bytes long.
// Note this is not the representation used in DRISL (CBOR).
func (c Cid) Bytes() []byte {
	// Slice returned over array for user convenience.
	// Writing to files, etc.
	b := make([]byte, CidBinaryLength)
	copy(b, c.b[:])
	return b
}

// String returns the CID in string format.
// It will always be CidStrLength bytes (or ASCII characters) long.
func (c Cid) String() string {
	s := multibaseBase32.EncodeToString(c.b[:])
	return "b" + s
}

// Equals returns true if the two CIDs are exactly the same.
//
// CIDs with the same hash type and digest but different codecs are not considered equal.
//
// This is equivalent to comparing the .Bytes() or .String() output of two CIDs,
// but more efficient.
func (c Cid) Equals(o Cid) bool {
	return c.b == o.b
}

// Codec returns the codec of the CID.
func (c Cid) Codec() Codec {
	return Codec(c.b[1])
}

// HashType returns the hash type of the CID.
func (c Cid) HashType() HashType {
	return HashType(c.b[2])
}

// Digest returns the hash digest stored in the CID.
// It is safe to modify.
func (c Cid) Digest() [HashSize]byte {
	// Copy digest into new slice for safety and return
	var digest [HashSize]byte
	copy(digest[:], c.b[dgIdx:])
	return digest
}

// Defined returns false if this Cid has no data due to being created with
// cid.Cid{} or new(cid.Cid).
// This method shouldn't be needed as long as you don't do this.
func (c Cid) Defined() bool {
	// Assume it's all zeros
	return c.b[0] != 0
}

// MarshalCBOR fulfills the drisl.Marshaler interface.
// It does not panic if the Cid is nil or empty, instead an error is returned.
func (c Cid) MarshalCBOR() ([]byte, error) {
	if !c.Defined() {
		return nil, errors.New("go-dasl/cid: will not marshal undefined Cid")
	}
	// Just return raw bytes instead of calling Marshal, because the header
	// is the same every time.
	return append([]byte{
		0xd8, 0x2a, // Tag 42
		0x58, 0x25, // 37 bytes
		0x00, // CID in CBOR prefix
	}, c.b[:]...), nil
}

// UnmarshalCBOR fulfills the drisl.Unmarshaler interface.
func (c *Cid) UnmarshalCBOR(b []byte) error {
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

	// Skip 0x00 prefix
	parsed, err := NewCidFromBytes(cidData[1:])
	if err != nil {
		return err
	}
	*c = parsed
	return nil
}
