package drisl

import (
	"reflect"

	"github.com/fxamacker/cbor/v2"
)

var (
	drislDecMode cbor.DecMode
	drislEncMode cbor.EncMode
)

func init() {
	// Setup cbor lib options

	cborTags := cbor.NewTagSet()
	err := cborTags.Add(
		cbor.TagOptions{EncTag: cbor.EncTagRequired, DecTag: cbor.DecTagRequired},
		reflect.TypeOf(Cid{}),
		CidTagNumber,
	)
	if err != nil {
		panic(err)
	}

	svr := cbor.NewSimpleValueRegistryStrict()

	drislDecMode, err = cbor.DecOptions{
		// Try to be strict
		DupMapKey:        cbor.DupMapKeyEnforcedAPF,
		TimeTag:          cbor.DecTagOptional,
		IndefLength:      cbor.IndefLengthForbidden,
		DefaultMapType:   reflect.TypeOf(map[string]any{}),
		MapKeyByteString: cbor.MapKeyByteStringForbidden,
		SimpleValues:     svr,
		NaN:              cbor.NaNDecodeForbidden,
		Inf:              cbor.InfDecodeForbidden,
		BignumTag:        cbor.BignumTagForbidden,
		Float64Only:      true,
		TagsMd:           cbor.TagsLimited,
	}.DecModeWithSharedTags(cborTags)
	if err != nil {
		panic(err)
	}

	drislEncMode, err = cbor.EncOptions{
		// Try to be strict
		Sort:             cbor.SortLengthFirst,
		ShortestFloat:    cbor.ShortestFloatNone,
		NaNConvert:       cbor.NaNConvertReject,
		InfConvert:       cbor.InfConvertReject,
		BigIntConvert:    cbor.BigIntConvertOnly,
		Time:             cbor.TimeRFC3339Nano,
		TimeTag:          cbor.EncTagNone,
		IndefLength:      cbor.IndefLengthForbidden,
		MapKeyStringOnly: true,
		SimpleValues:     svr,
		Float64Only:      true,
	}.EncModeWithSharedTags(cborTags)
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
