/*
Package cid is an implementation of Content Identifiers, specifically the restricted DASL subset of CIDs.

https://dasl.ing/cid.html
*/
package cid

import (
	"bytes"
	"crypto/sha256"
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/hyphacoop/cbor/v2"
	"github.com/multiformats/go-varint"
)

// Codec is the encoding of the data represented by the CID.
type Codec byte

// HashType is the algorithm of the hash digest embedded in the CID.
type HashType byte

const (
	minCidLength = 4
	cidTagNumber = 42

	CidVersion              = 0x01
	CodecRaw       Codec    = 0x55
	CodecDrisl     Codec    = 0x71
	HashTypeSha256 HashType = 0x12
	HashTypeBlake3 HashType = 0x1e

	// The usual hash length
	HashLength = 32
)

var (
	// https://github.com/multiformats/multibase
	// The encoding referred to by "base32" or "b".
	// "RFC4648 case-insensitive - no padding"
	multibaseBase32 = base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567").WithPadding(base32.NoPadding)

	// EmptyCid is a raw SHA-256 CID that represents no data.
	// Its digest is the SHA-256 hash of the empty byte string "".
	// Try not to use it as a sentinel value.
	EmptyCid = Cid{b: hexDecode("01551220e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")}
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
//
// It is always valid, unless created directly such as cid.Cid{} or new(cid.Cid).
// That will cause panics if methods are called upon it, unless otherwise documented.
//
// Programs using CIDs should typically store and pass them as values, not pointers.
// That is, CID variables and struct fields should be of type cid.Cid, not *cid.Cid.
//
// https://dasl.ing/cid.html
type Cid struct {
	// b is the binary CID data
	b []byte
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
	if len(in) < minCidLength {
		return Cid{}, &ForbiddenCidError{"too few bytes"}
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
	hashSize, idx, err := varint.FromUvarint(in[3:])
	if err != nil {
		return Cid{}, &ForbiddenCidError{err.Error()}
	}
	idx += 3 // Index of `in` where the hash starts
	if len(in[idx:]) != int(hashSize) {
		return Cid{}, &ForbiddenCidError{"remaining data doesn't match stated hash size"}
	}

	// Create Cid, copying input data so it can't be changed by the caller
	b := make([]byte, len(in))
	copy(b, in)
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

	cid := Cid{b: make([]byte, 0, minCidLength)}

	b, err := r.ReadByte()
	if err != nil {
		return Cid{}, fixErr(err)
	}
	if b != CidVersion {
		return Cid{}, &ForbiddenCidError{"invalid version"}
	}
	cid.b = append(cid.b, b)

	b, err = r.ReadByte()
	if err != nil {
		return Cid{}, fixErr(err)
	}
	if b != byte(CodecRaw) && b != byte(CodecDrisl) {
		return Cid{}, &ForbiddenCidError{"invalid codec"}
	}
	cid.b = append(cid.b, b)

	b, err = r.ReadByte()
	if err != nil {
		return Cid{}, fixErr(err)
	}
	if b != byte(HashTypeSha256) && b != byte(HashTypeBlake3) {
		return Cid{}, &ForbiddenCidError{"invalid hash type"}
	}
	cid.b = append(cid.b, b)

	hashSize, bs, err := readUvarint(r)
	if err != nil {
		if err == io.ErrUnexpectedEOF {
			return Cid{}, err
		}
		return Cid{}, &ForbiddenCidError{err.Error()}
	}
	cid.b = append(cid.b, bs...)

	digest := make([]byte, hashSize)
	_, err = io.ReadFull(r, digest)
	if err != nil {
		return Cid{}, fixErr(err)
	}
	cid.b = append(cid.b, digest...)

	return cid, nil
}

// NewCidFromString creates a new DASL CID from the given string.
// A ForbiddenCidError is returned if it is invalid DASL.
func NewCidFromString(s string) (Cid, error) {
	if len(s) == 0 {
		return Cid{}, &ForbiddenCidError{"empty"}
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
// Currently the length of the digest is not validated.
func NewCidFromInfo(codec Codec, hashType HashType, digest []byte) (Cid, error) {
	if codec != CodecRaw && codec != CodecDrisl {
		return Cid{}, &ForbiddenCidError{"invalid codec"}
	}
	if hashType != HashTypeSha256 && hashType != HashTypeBlake3 {
		return Cid{}, &ForbiddenCidError{"invalid hash type"}
	}
	b := make([]byte, 3, 4+len(digest)) // Assume hash size varint is 1 byte long
	b[0] = CidVersion
	b[1] = byte(codec)
	b[2] = byte(hashType)

	// Go slice size limits prevent this varint from being larger than the max allowed by
	// the IPFS unsigned-varint spec: https://github.com/multiformats/unsigned-varint
	// So I can just add it and move on
	b = binary.AppendUvarint(b, uint64(len(digest)))

	b = append(b, digest...)
	return Cid{b}, nil
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
	return Cid{append([]byte{CidVersion, byte(CodecRaw), byte(HashTypeSha256), HashLength}, digest[:]...)}
}

// HashReader creates a raw SHA-256 CID by hashing all the data in the reader.
// Any error returned comes from the reader.
func HashReader(r io.Reader) (Cid, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, r); err != nil {
		return Cid{}, err
	}
	// Quick version of NewCidFromInfo
	return Cid{append([]byte{CidVersion, byte(CodecRaw), byte(HashTypeSha256), HashLength}, hasher.Sum(nil)...)}, nil
}

// Bytes returns the CID in binary format.
// It is safe to modify.
// Note this is not the representation used in DRISL (CBOR)
func (c Cid) Bytes() []byte {
	b := make([]byte, len(c.b))
	copy(b, c.b)
	return b
}

// String returns the CID in string format.
func (c Cid) String() string {
	s, _ := c.MarshalText()
	return string(s)
}

// Equals returns true if the two CIDs are exactly the same.
//
// CIDs with the same hash type and digest but different codecs are not considered equal.
//
// This is equivalent to comparing the .Bytes() or .String() output of two CIDs,
// but more efficient.
func (c Cid) Equals(o Cid) bool {
	return bytes.Equal(c.b, o.b)
}

// Codec returns the codec of the CID.
func (c Cid) Codec() Codec {
	return Codec(c.b[1])
}

// HashType returns the hash type of the CID.
func (c Cid) HashType() HashType {
	return HashType(c.b[2])
}

// HashSize returns the size of the hash digest stored in the CID.
func (c Cid) HashSize() uint64 {
	// TODO: optimize this and the next func?
	// By extracting hash digest/size/index or something on creation

	// Skip version, codec, hash type
	size, _, _ := varint.FromUvarint(c.b[3:])
	return size
}

// Digest returns the hash digest stored in the CID.
// It is safe to modify.
func (c Cid) Digest() []byte {
	// Move past varint to get start of digest
	i := 3 // Skip version, codec, hash type
	for c.b[i] > 0x80 {
		i++
	}
	i++ // Skip last byte of varint

	// Copy digest into new slice for safety and return
	digest := make([]byte, len(c.b)-i)
	copy(digest, c.b[i:])
	return digest
}

// Defined returns false if this Cid has no data due to being created with
// cid.Cid{} or new(cid.Cid).
// This method shouldn't be needed as long as you don't do this.
func (c Cid) Defined() bool {
	return c.b != nil
}

// MarshalCBOR fulfills the drisl.Marshaler interface.
// It does not panic if the Cid is nil or empty, instead an error is returned.
func (c Cid) MarshalCBOR() ([]byte, error) {
	if c.b == nil {
		return nil, errors.New("go-dasl/cid: cannot marshal nil Cid")
	}
	// CID in CBOR is just CID bytes with 0x00 prepended
	return cbor.Marshal(cbor.Tag{
		Number:  cidTagNumber,
		Content: append([]byte{0x00}, c.b...),
	})
}

// MarshalJSON fulfills the json.Marshaler interface.
// It follows the dag-json standard: {"/": "bafkr..."}
// This just for simple display purposes.
func (c Cid) MarshalJSON() ([]byte, error) {
	// Pre-calculate buffer size: {"/": + opening quote + CID string + closing quote + }
	// CID string length = 1 (multibase prefix 'b') + base32 encoded length
	cidLen := 1 + multibaseBase32.EncodedLen(len(c.b))
	bufLen := 6 + cidLen + 2 // {"/": + } = 6, plus quotes for CID

	buf := make([]byte, 0, bufLen)
	buf = append(buf, `{"/":"`...)

	// Inline the string encoding to avoid c.String() allocation
	buf = append(buf, 'b')
	// Encode directly into the buffer
	oldLen := len(buf)
	buf = buf[:oldLen+cidLen-1]
	multibaseBase32.Encode(buf[oldLen:], c.b)

	buf = append(buf, `"}`...)
	return buf, nil
}

// MarshalText fulfills the encoding.TextMarshaler interface.
// It is equivalent to String().
func (c Cid) MarshalText() ([]byte, error) {
	text := make([]byte, multibaseBase32.EncodedLen(len(c.b))+1)
	text[0] = 'b'
	multibaseBase32.Encode(text[1:], c.b)
	return text, nil
}

// MarshalBinary fulfills the encoding.BinaryMarshaler interface.
// It is equivalent to Bytes().
func (c Cid) MarshalBinary() ([]byte, error) {
	return c.Bytes(), nil
}

// AppendBinary fulfills the encoding.BinaryAppender interface.
// It simply appends the Bytes() output.
func (c Cid) AppendBinary(b []byte) ([]byte, error) {
	return append(b, c.b...), nil
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

// UnmarshalText fulfills the encoding.TextUnmarshaler interface.
// It is equivalent to NewCidFromString.
func (c *Cid) UnmarshalText(text []byte) error {
	parsed, err := NewCidFromString(string(text))
	if err != nil {
		return err
	}
	*c = parsed
	return nil
}

// UnmarshalBinary fulfills the encoding.BinaryUnmarshaler interface.
// It is equivalent to NewCidFromBytes.
func (c *Cid) UnmarshalBinary(data []byte) error {
	parsed, err := NewCidFromBytes(data)
	if err != nil {
		return err
	}
	*c = parsed
	return nil
}

// readUvarint reads a unsigned varint from the given reader.
// Modified from go-varint v0.1.0 under MIT license.
func readUvarint(r io.ByteReader) (uint64, []byte, error) {
	// Modified from the go standard library. Copyright the Go Authors and
	// released under the BSD License.
	var x uint64
	var s uint
	bs := make([]byte, 0, 1)
	for s = 0; ; s += 7 {
		b, err := r.ReadByte()
		if err != nil {
			if err == io.EOF && s != 0 {
				// "eof" will look like a success.
				// If we've read part of a value, this is not a
				// success.
				err = io.ErrUnexpectedEOF
			}
			return 0, nil, err
		}
		bs = append(bs, b)
		if (s == 56 && b >= 0x80) || s >= (7*varint.MaxLenUvarint63) {
			// this is the 9th and last byte we're willing to read, but it
			// signals there's more (1 in MSB).
			// or this is the >= 10th byte, and for some reason we're still here.
			return 0, nil, varint.ErrOverflow
		}
		if b < 0x80 {
			if b == 0 && s > 0 {
				return 0, nil, varint.ErrNotMinimal
			}
			return x | uint64(b)<<s, bs, nil
		}
		x |= uint64(b&0x7f) << s
	}
}
