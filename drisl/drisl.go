package drisl

import (
	"crypto/sha256"
	"reflect"

	"github.com/hyphacoop/cbor/v2"
	"github.com/hyphacoop/go-dasl/cid"
)

var (
	drislDecMode DecMode
	drislEncMode EncMode
	cborTags     cbor.TagSet
	svr          *cbor.SimpleValueRegistry
	svrUndefined *cbor.SimpleValueRegistry
)

// CidTagNumber is the number of the tag used to encode a CID in CBOR.
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

// Marshal returns the DRISL encoding of v using default encoding options.
// See EncOptions for encoding options.
//
// Marshal uses the following encoding rules:
//
// If value implements the Marshaler interface, Marshal calls its
// MarshalCBOR method. Invalid DRISL will still be rejected.
//
// If value implements encoding.BinaryMarshaler, Marhsal calls its
// MarshalBinary method and encode it as CBOR byte string.
//
// Boolean values encode as CBOR booleans (type 7).
//
// Positive integer values encode as CBOR positive integers (type 0).
//
// Negative integer values encode as CBOR negative integers (type 1).
//
// Floating point values encode as CBOR floating points (type 7).
// They are always 64 bits wide, even if a float32 type is used.
//
// String values encode as CBOR text strings (type 3).
//
// []byte values encode as CBOR byte strings (type 2).
//
// Array and slice values encode as CBOR arrays (type 4).
//
// Map values encode as CBOR maps (type 5).
//
// Struct values encode as CBOR maps (type 5).  Each exported struct field
// becomes a pair with field name encoded as CBOR text string (type 3) and
// field value encoded based on its type.
// See struct tag option "toarray" for special field "_" to encode struct values as
// CBOR array (type 4).
//
// Marshal supports format string stored under the "cbor" key in the struct
// field's tag.  CBOR format string can specify the name of the field,
// "omitempty", "omitzero", and special case "-" for
// field omission. If "cbor" key is absent, Marshal uses "json" key.
// When using the "json" key, the "omitzero" option is honored when building
// with Go 1.24+ to match stdlib encoding/json behavior.
//
// Special struct field "_" is used to specify struct level options, such as
// "toarray". "toarray" option enables Go struct to be encoded as CBOR array.
// "omitempty" and "omitzero" are disabled by "toarray" to ensure that the
// same number of elements are encoded every time.
//
// Anonymous struct fields are marshaled as if their exported fields
// were fields in the outer struct.  Marshal follows the same struct fields
// visibility rules used by JSON encoding package.
//
// time.Time encode as RFC3339 strings with nanosecond precision by default.
//
// big.Int values encode as CBOR integers (type 0 and 1) if values fit.
// Otherwise, an error is returned.
//
// Pointer values encode as the value pointed to.
//
// Interface values encode as the value stored in the interface.
//
// Nil slice/map/pointer/interface values encode as CBOR nulls (type 7).
//
// Values of other types cannot be encoded in CBOR.  Attempting
// to encode such a value causes Marshal to return an UnsupportedTypeError.
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

// TimeMode specifies how to encode time.Time values
type TimeMode int

// I use a custom type instead of cbor.TimeMode so that I can change the order and
// make TimeRFC3339Nano the default because I think it's obviously a better choice.

const (
	// TimeRFC3339Nano causes time.Time to encode to a CBOR time (tag 0) with a text string content
	// representing the time using 1-nanosecond precision in RFC3339 format.  If the time.Time has a
	// non-UTC timezone then a "localtime - UTC" numeric offset will be included as specified in RFC3339.
	// NOTE: User applications can avoid including the RFC3339 numeric offset by:
	// - providing a time.Time value set to UTC, or
	// - using the TimeUnix, TimeUnixMicro, or TimeUnixDynamic option instead of TimeRFC3339Nano.
	TimeRFC3339Nano TimeMode = iota

	// TimeUnix causes time.Time to encode to a CBOR time (tag 1) with an integer content
	// representing seconds elapsed (with 1-second precision) since UNIX Epoch UTC.
	// The TimeUnix option is location independent and has a clear precision guarantee.
	TimeUnix

	// TimeUnixMicro causes time.Time to encode to a CBOR time (tag 1) with a floating point content
	// representing seconds elapsed (with up to 1-microsecond precision) since UNIX Epoch UTC.
	// NOTE: The floating point content is encoded to the shortest floating-point encoding that preserves
	// the 64-bit floating point value. I.e., the floating point encoding can be IEEE 764:
	// binary64, binary32, or binary16 depending on the content's value.
	TimeUnixMicro

	// TimeUnixDynamic causes time.Time to encode to a CBOR time (tag 1) with either an integer content or
	// a floating point content, depending on the content's value.  This option is equivalent to dynamically
	// choosing TimeUnix if time.Time doesn't have fractional seconds, and using TimeUnixMicro if time.Time
	// has fractional seconds.
	TimeUnixDynamic

	// TimeRFC3339 causes time.Time to encode to a CBOR time (tag 0) with a text string content
	// representing the time using 1-second precision in RFC3339 format.  If the time.Time has a
	// non-UTC timezone then a "localtime - UTC" numeric offset will be included as specified in RFC3339.
	// NOTE: User applications can avoid including the RFC3339 numeric offset by:
	// - providing a time.Time value set to UTC, or
	// - using the TimeUnix, TimeUnixMicro, or TimeUnixDynamic option instead of TimeRFC3339.
	TimeRFC3339

	// TimeModeReject returns an UnsupportedTypeError instead of marshaling a time.Time.
	TimeModeReject
)

func (tm TimeMode) ToCborTimeMode() cbor.TimeMode {
	switch tm {
	case TimeRFC3339Nano:
		return cbor.TimeRFC3339Nano
	case TimeUnix:
		return cbor.TimeUnix
	case TimeUnixMicro:
		return cbor.TimeUnixMicro
	case TimeUnixDynamic:
		return cbor.TimeUnixDynamic
	case TimeRFC3339:
		return cbor.TimeRFC3339
	case TimeModeReject:
		return cbor.TimeModeReject
	default:
		// Shouldn't be here
		return cbor.TimeRFC3339Nano
	}
}

// EncOptions specifies encoding options.
type EncOptions struct {
	// Time specifies how to encode time.Time.
	// Note the default is RFC3339 string with nanosecond precision.
	Time TimeMode

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
		Time:             opts.Time.ToCborTimeMode(),
		TimeTag:          cbor.EncTagNone,
		IndefLength:      cbor.IndefLengthForbidden,
		MapKeyStringOnly: true,
		SimpleValues:     svr,
		Float64Only:      true,
		DisableKeyAsInt:  true,
		TagsMd:           cbor.TagsLimited,
		Int64RangeOnly:   opts.Int64RangeOnly,
		// I think this is more intuitive
		OmitEmpty: cbor.OmitEmptyGoValue,
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
