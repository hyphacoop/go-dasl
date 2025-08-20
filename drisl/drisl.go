package drisl

import (
	"crypto/sha256"
	"reflect"

	"github.com/fxamacker/cbor/v2"
	"github.com/hyphacoop/go-dasl/cid"
)

var (
	drislDecMode DecMode
	drislEncMode EncMode
	cborTags     cbor.TagSet
	svr          *cbor.SimpleValueRegistry
	svrUndefined *cbor.SimpleValueRegistry
)

const CidTagNumber = 42

func init() {
	cborTags = cbor.NewTagSet()
	err := cborTags.Add(
		cbor.TagOptions{EncTag: cbor.EncTagRequired, DecTag: cbor.DecTagRequired},
		reflect.TypeOf(cid.Cid{}),
		CidTagNumber,
	)
	if err != nil {
		panic(err)
	}

	svr = cbor.NewSimpleValueRegistryStrict()
	svrUndefined = cbor.NewSimpleValueRegistryStrictUndefined()

	drislDecMode, err = DecOptions{}.DecMode()
	if err != nil {
		panic(err)
	}
	drislEncMode, err = EncOptions{}.EncMode()
	if err != nil {
		panic(err)
	}
}

func Marshal(v any) ([]byte, error) {
	return drislEncMode.Marshal(v)
}

func Unmarshal(data []byte, v any) error {
	return drislDecMode.Unmarshal(data, v)
}

// func Valid(data []byte) bool {
// 	// XXX: this is correct but inefficient
// 	var v any
// 	return drislDecMode.Unmarshal(data, &v) == nil
// }

// DecOptions specifies decoding options.
type DecOptions struct {
	// MaxNestedLevels specifies the max nested levels allowed for any combination of CBOR array, maps, and tags.
	// Default is 32 levels and it can be set to [4, 65535]. Note that higher maximum levels of nesting can
	// require larger amounts of stack to deserialize. Don't increase this higher than you require.
	MaxNestedLevels int

	// MaxArrayElements specifies the max number of elements for CBOR arrays.
	// Default is 128*1024=131072 and it can be set to [16, 2147483647]
	MaxArrayElements int

	// MaxMapPairs specifies the max number of key-value pairs for CBOR maps.
	// Default is 128*1024=131072 and it can be set to [16, 2147483647]
	MaxMapPairs int

	// Int64RangeOnly reduces the range of valid integers when decoding to the range
	// supported by the int64 type: [-(2^63), 2^63-1].
	Int64RangeOnly bool

	// AllowUndefined accepts CBOR's 'undefined' simple value when decoding, silently
	// turning it into Go's nil.
	AllowUndefined bool
}

// DecMode is the main interface for decoding.
type DecMode interface {
	// Unmarshal parses the CBOR-encoded data into the value pointed to by v
	// using the decoding mode.  If v is nil, not a pointer, or a nil pointer,
	// Unmarshal returns an error.
	//
	// See the documentation for Unmarshal for details.
	Unmarshal(data []byte, v any) error
}

func (opts DecOptions) DecMode() (DecMode, error) {
	thisSvr := svr
	if opts.AllowUndefined {
		thisSvr = svrUndefined
	}
	return cbor.DecOptions{
		// Try to be strict
		DupMapKey:          cbor.DupMapKeyEnforcedAPF,
		IndefLength:        cbor.IndefLengthForbidden,
		DefaultMapType:     reflect.TypeOf(map[string]any{}),
		MapKeyByteString:   cbor.MapKeyByteStringForbidden,
		SimpleValues:       thisSvr,
		NaN:                cbor.NaNDecodeForbidden,
		Inf:                cbor.InfDecodeForbidden,
		BignumTag:          cbor.BignumTagForbidden,
		Float64Only:        true,
		TagsMd:             cbor.TagsLimited,
		EnforceIntPrefEnc:  true,
		MapKeyTypeStrict:   true,
		DisableKeyAsInt:    true,
		EnforceSort:        true,
		KeepFloatPrecision: true,
		MaxNestedLevels:    opts.MaxNestedLevels,
		MaxArrayElements:   opts.MaxArrayElements,
		MaxMapPairs:        opts.MaxMapPairs,
		Int64RangeOnly:     opts.Int64RangeOnly,
	}.DecModeWithSharedTags(cborTags)
}

// EncOptions specifies encoding options.
type EncOptions struct {
	// Int64RangeOnly reduces the range of valid integers when encoding to the range
	// supported by the int64 type: [-(2^63), 2^63-1]
	Int64RangeOnly bool
}

// EncMode is the main interface for encoding.
type EncMode interface {
	Marshal(v any) ([]byte, error)
}

func (opts EncOptions) EncMode() (EncMode, error) {
	return cbor.EncOptions{
		// Try to be strict
		Sort:             cbor.SortBytewiseLexical,
		ShortestFloat:    cbor.ShortestFloatNone,
		NaNConvert:       cbor.NaNConvertReject,
		InfConvert:       cbor.InfConvertReject,
		BigIntConvert:    cbor.BigIntConvertOnly,
		Time:             cbor.TimeModeReject,
		TimeTag:          cbor.EncTagNone,
		IndefLength:      cbor.IndefLengthForbidden,
		MapKeyStringOnly: true,
		SimpleValues:     svr,
		Float64Only:      true,
		DisableKeyAsInt:  true,
		TagsMd:           cbor.TagsLimited,
		Int64RangeOnly:   opts.Int64RangeOnly,
	}.EncModeWithSharedTags(cborTags)
}

// Marshaler is the interface implemented by types that can marshal themselves
// into valid CBOR.
//
// If the CBOR is not DRISL-compliant it will be rejected. drisl.Marshal should
// be used to get output for this function.
type Marshaler interface {
	MarshalCBOR() ([]byte, error)
}

// Unmarshaler is the interface implemented by types that wish to unmarshal
// CBOR data themselves.  The input is a valid CBOR value. UnmarshalCBOR
// must copy the CBOR data if it needs to use it after returning.
//
// Only DRISL-compliant CBOR will be provided to this function. drisl.Unmarshal
// should be used to decode it.
type Unmarshaler interface {
	UnmarshalCBOR([]byte) error
}

// CalculateCidForValue calculates the DRISL SHA-256 CID for the given Go value.
// This is achieved by marshalling it into DRISL and then hashing those bytes.
// An error is returned if the value could not be marshalled
func CalculateCidForValue(v any) (cid.Cid, error) {
	b, err := Marshal(v)
	if err != nil {
		return cid.Cid{}, err
	}
	digest := sha256.Sum256(b)
	return cid.NewCidFromInfo(cid.CodecDrisl, cid.HashTypeSha256, digest[:])
}
