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

	// Reject undefined
	svr, err := cbor.NewSimpleValueRegistryFromDefaults(
		cbor.WithRejectedSimpleValue(cbor.SimpleValue(23)),
	)
	if err != nil {
		panic(err)
	}

	drislDecMode, err = cbor.DecOptions{
		// Easier to re-encode to JSON later
		DefaultMapType: reflect.TypeOf(map[string]any{}),

		// Try to be strict
		DupMapKey:        cbor.DupMapKeyEnforcedAPF,
		TimeTag:          cbor.DecTagOptional,
		IndefLength:      cbor.IndefLengthForbidden,
		MapKeyByteString: cbor.MapKeyByteStringForbidden,
		SimpleValues:     svr,
		NaN:              cbor.NaNDecodeForbidden,
		Inf:              cbor.InfDecodeForbidden,
		BignumTag:        cbor.BignumTagForbidden,
	}.DecModeWithSharedTags(cborTags)
	if err != nil {
		panic(err)
	}

	drislEncMode, err = cbor.EncOptions{
		// Try to be strict
		Sort:          cbor.SortLengthFirst,
		ShortestFloat: cbor.ShortestFloatNone,
		NaNConvert:    cbor.NaNConvertReject,
		InfConvert:    cbor.InfConvertReject,
		BigIntConvert: cbor.BigIntConvertShortest,
		TimeTag:       cbor.EncTagNone,
		IndefLength:   cbor.IndefLengthForbidden,
	}.EncModeWithSharedTags(cborTags)
}

func Marshal(v any) ([]byte, error) {
	return drislEncMode.Marshal(v)
}

func Unmarshal(data []byte, v any) error {
	return drislDecMode.Unmarshal(data, v)
}

func Valid(data []byte) bool {
	return drislDecMode.Wellformed(data) == nil
}
