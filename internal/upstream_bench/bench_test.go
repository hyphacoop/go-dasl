// This file is a copy of drisl_bench_test.go but using the upstream CBOR lib instead.
// It configures the CBOR lib to match ours as closely as possible.
package upstreambench

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/hyphacoop/go-dasl/cid"
)

func hexDecode(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func decOptions() cbor.DecOptions {
	return cbor.DecOptions{
		DupMapKey:        cbor.DupMapKeyEnforcedAPF,
		IndefLength:      cbor.IndefLengthForbidden,
		DefaultMapType:   reflect.TypeOf(map[string]any{}),
		MapKeyByteString: cbor.MapKeyByteStringForbidden,
		NaN:              cbor.NaNDecodeForbidden,
		Inf:              cbor.InfDecodeForbidden,
		BignumTag:        cbor.BignumTagForbidden,
		TagsMd:           cbor.TagsAllowed,
	}
}

func decMode() cbor.DecMode {
	cborTags := cbor.NewTagSet()
	err := cborTags.Add(
		cbor.TagOptions{EncTag: cbor.EncTagRequired, DecTag: cbor.DecTagRequired},
		reflect.TypeOf(cid.Cid{}),
		42,
	)
	if err != nil {
		panic(err)
	}
	dm, err := decOptions().DecModeWithSharedTags(cborTags)
	if err != nil {
		panic(err)
	}
	return dm
}
func encMode() cbor.EncMode {
	cborTags := cbor.NewTagSet()
	err := cborTags.Add(
		cbor.TagOptions{EncTag: cbor.EncTagRequired, DecTag: cbor.DecTagRequired},
		reflect.TypeOf(cid.Cid{}),
		42,
	)
	if err != nil {
		panic(err)
	}
	dm, err := cbor.EncOptions{
		Sort:          cbor.SortBytewiseLexical,
		ShortestFloat: cbor.ShortestFloatNone,
		NaNConvert:    cbor.NaNConvertReject,
		InfConvert:    cbor.InfConvertReject,
		BigIntConvert: cbor.BigIntConvertShortest,
		Time:          cbor.TimeRFC3339Nano,
		TimeTag:       cbor.EncTagNone,
		IndefLength:   cbor.IndefLengthForbidden,
		TagsMd:        cbor.TagsAllowed,
	}.EncModeWithSharedTags(cborTags)
	if err != nil {
		panic(err)
	}
	return dm
}

type T1 struct {
	T    bool
	UI   uint
	I    int
	F    float64
	B    []byte
	S    string
	Slci []int
	Mss  map[string]string
}

type T3 struct {
	_    struct{} `cbor:",toarray"`
	T    bool
	UI   uint
	I    int
	F    float64
	B    []byte
	S    string
	Slci []int
	Mss  map[string]string
}

type ManyFieldsOneOmitEmpty struct {
	F01, F02, F03, F04, F05, F06, F07, F08, F09, F10, F11, F12, F13, F14, F15, F16 int
	F17, F18, F19, F20, F21, F22, F23, F24, F25, F26, F27, F28, F29, F30, F31      int

	F32 int `cbor:",omitempty"`
}

type SomeFieldsOneOmitEmpty struct {
	F01, F02, F03, F04, F05, F06, F07 int

	F08 int `cbor:",omitempty"`
}

type ManyFieldsAllOmitEmpty struct {
	F01 int `cbor:",omitempty"`
	F02 int `cbor:",omitempty"`
	F03 int `cbor:",omitempty"`
	F04 int `cbor:",omitempty"`
	F05 int `cbor:",omitempty"`
	F06 int `cbor:",omitempty"`
	F07 int `cbor:",omitempty"`
	F08 int `cbor:",omitempty"`
	F09 int `cbor:",omitempty"`
	F10 int `cbor:",omitempty"`
	F11 int `cbor:",omitempty"`
	F12 int `cbor:",omitempty"`
	F13 int `cbor:",omitempty"`
	F14 int `cbor:",omitempty"`
	F15 int `cbor:",omitempty"`
	F16 int `cbor:",omitempty"`
	F17 int `cbor:",omitempty"`
	F18 int `cbor:",omitempty"`
	F19 int `cbor:",omitempty"`
	F20 int `cbor:",omitempty"`
	F21 int `cbor:",omitempty"`
	F22 int `cbor:",omitempty"`
	F23 int `cbor:",omitempty"`
	F24 int `cbor:",omitempty"`
	F25 int `cbor:",omitempty"`
	F26 int `cbor:",omitempty"`
	F27 int `cbor:",omitempty"`
	F28 int `cbor:",omitempty"`
	F29 int `cbor:",omitempty"`
	F30 int `cbor:",omitempty"`
	F31 int `cbor:",omitempty"`
	F32 int `cbor:",omitempty"`
}

type SomeFieldsAllOmitEmpty struct {
	F01 int `cbor:",omitempty"`
	F02 int `cbor:",omitempty"`
	F03 int `cbor:",omitempty"`
	F04 int `cbor:",omitempty"`
	F05 int `cbor:",omitempty"`
	F06 int `cbor:",omitempty"`
	F07 int `cbor:",omitempty"`
	F08 int `cbor:",omitempty"`
}

var (
	typeIntf            = reflect.TypeOf([]any(nil)).Elem()
	typeTime            = reflect.TypeOf(time.Time{})
	typeBigInt          = reflect.TypeOf(big.Int{})
	typeString          = reflect.TypeOf("")
	typeByteSlice       = reflect.TypeOf([]byte(nil))
	typeBool            = reflect.TypeOf(true)
	typeUint8           = reflect.TypeOf(uint8(0))
	typeUint16          = reflect.TypeOf(uint16(0))
	typeUint32          = reflect.TypeOf(uint32(0))
	typeUint64          = reflect.TypeOf(uint64(0))
	typeInt8            = reflect.TypeOf(int8(0))
	typeInt16           = reflect.TypeOf(int16(0))
	typeInt32           = reflect.TypeOf(int32(0))
	typeInt64           = reflect.TypeOf(int64(0))
	typeFloat32         = reflect.TypeOf(float32(0))
	typeFloat64         = reflect.TypeOf(float64(0))
	typeByteArray       = reflect.TypeOf([5]byte{})
	typeIntSlice        = reflect.TypeOf([]int{})
	typeStringSlice     = reflect.TypeOf([]string{})
	typeMapIntfIntf     = reflect.TypeOf(map[any]any{})
	typeMapStringInt    = reflect.TypeOf(map[string]int{})
	typeMapStringString = reflect.TypeOf(map[string]string{})
	typeMapStringIntf   = reflect.TypeOf(map[string]any{})
	typeCID             = reflect.TypeOf(cid.Cid{})
)

var decodeBenchmarks = []struct {
	name          string
	data          []byte
	decodeToTypes []reflect.Type
}{
	{
		name:          "bool",
		data:          hexDecode("f5"),
		decodeToTypes: []reflect.Type{typeIntf, typeBool},
	}, // true
	{
		name:          "positive int",
		data:          hexDecode("1bffffffffffffffff"),
		decodeToTypes: []reflect.Type{typeIntf, typeUint64},
	}, // uint64(18446744073709551615)
	{
		name:          "negative int",
		data:          hexDecode("3903e7"),
		decodeToTypes: []reflect.Type{typeIntf, typeInt64},
	}, // int64(-1000)
	{
		name:          "float",
		data:          hexDecode("fbc010666666666666"),
		decodeToTypes: []reflect.Type{typeIntf, typeFloat64},
	}, // float64(-4.1)
	{
		name:          "bytes",
		data:          hexDecode("581a0102030405060708090a0b0c0d0e0f101112131415161718191a"),
		decodeToTypes: []reflect.Type{typeIntf, typeByteSlice},
	}, // []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26}
	{
		name:          "text",
		data:          hexDecode("782b54686520717569636b2062726f776e20666f78206a756d7073206f76657220746865206c617a7920646f67"),
		decodeToTypes: []reflect.Type{typeIntf, typeString},
	}, // "The quick brown fox jumps over the lazy dog"
	{
		name:          "array",
		data:          hexDecode("981a0102030405060708090a0b0c0d0e0f101112131415161718181819181a"),
		decodeToTypes: []reflect.Type{typeIntf, typeIntSlice},
	}, // []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26}
	{
		name:          "map",
		data:          hexDecode("ad616161416162614261636143616461446165614561666146616761476168614861696149616a614a616c614c616d614d616e614e"),
		decodeToTypes: []reflect.Type{typeIntf, typeMapStringIntf, typeMapStringString},
	}, // map[string]string{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E", "f": "F", "g": "G", "h": "H", "i": "I", "j": "J", "l": "L", "m": "M", "n": "N"}}
	{
		name:          "cid",
		data:          hexDecode("d82a582500015512205891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03"),
		decodeToTypes: []reflect.Type{typeCID},
	},
}

var encodeBenchmarks = []struct {
	name   string
	data   []byte
	values []any
}{
	{
		name:   "bool",
		data:   hexDecode("f5"),
		values: []any{true},
	},
	{
		name:   "positive int",
		data:   hexDecode("1bffffffffffffffff"),
		values: []any{uint64(18446744073709551615)},
	},
	{
		name:   "negative int",
		data:   hexDecode("3903e7"),
		values: []any{int64(-1000)},
	},
	{
		name:   "float",
		data:   hexDecode("fbc010666666666666"),
		values: []any{float64(-4.1), float32(-4.1)},
	},
	{
		name:   "bytes",
		data:   hexDecode("581a0102030405060708090a0b0c0d0e0f101112131415161718191a"),
		values: []any{[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26}},
	},
	{
		name:   "text",
		data:   hexDecode("782b54686520717569636b2062726f776e20666f78206a756d7073206f76657220746865206c617a7920646f67"),
		values: []any{"The quick brown fox jumps over the lazy dog"},
	},
	{
		name:   "array",
		data:   hexDecode("981a0102030405060708090a0b0c0d0e0f101112131415161718181819181a"),
		values: []any{[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26}},
	},
	{
		name:   "map",
		data:   hexDecode("ad616161416162614261636143616461446165614561666146616761476168614861696149616a614a616c614c616d614d616e614e"),
		values: []any{map[string]string{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E", "f": "F", "g": "G", "h": "H", "i": "I", "j": "J", "l": "L", "m": "M", "n": "N"}},
	},
	{
		name:   "cid",
		data:   hexDecode("D82A582401551220ADEE2E8FB5459C9BCF07D7D78D1183BF40A7F60F57A54A19194801C9A27EAD87"),
		values: []any{cid.MustNewCidFromString("bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4")},
	},
}

func BenchmarkUnmarshal(b *testing.B) {
	dm := decMode()
	for _, bm := range decodeBenchmarks {
		for _, t := range bm.decodeToTypes {
			name := "CBOR " + bm.name + " to Go " + t.String()
			if t.Kind() == reflect.Struct {
				name = "CBOR " + bm.name + " to Go " + t.Kind().String()
			}
			b.Run(name, func(b *testing.B) {
				b.SetBytes(int64(len(bm.data)))
				for b.Loop() {
					vPtr := reflect.New(t).Interface()
					if err := dm.Unmarshal(bm.data, vPtr); err != nil {
						b.Fatal("Unmarshal:", err)
					}
				}
			})
		}
	}
	var moreBenchmarks = []struct {
		name         string
		data         []byte
		decodeToType reflect.Type
	}{
		// Unmarshal CBOR map with string key to map[string]interface{}.
		{
			name:         "CBOR map to Go map[string]interface{}",
			data:         hexDecode("A76141F561443903E76145581A0102030405060708090A0B0C0D0E0F101112131415161718191A6146782B54686520717569636B2062726F776E20666F78206A756D7073206F76657220746865206C617A7920646F676243691BFFFFFFFFFFFFFFFF634D7373AD6163614361656145616661466167614761686148616D614E616E614D616F61416170614261716144617261496173614A6174614C64476C6369981A0102030405060708090A0B0C0D0E0F101112131415161718181819181A"),
			decodeToType: reflect.TypeOf(map[string]any{}),
		},
		// Unmarshal CBOR map with string key to struct.
		{
			name:         "CBOR map to Go struct",
			data:         hexDecode("A76141F56142581A0102030405060708090A0B0C0D0E0F101112131415161718191A6143FBC0106666666666666144782B54686520717569636B2062726F776E20666F78206A756D7073206F76657220746865206C617A7920646F676255491BFFFFFFFFFFFFFFFF634D7373AD6163614361656145616661466167614761686148616D614E616E614D616F61416170614261716144617261496173614A6174614C64456C6369981A0102030405060708090A0B0C0D0E0F101112131415161718181819181A"),
			decodeToType: reflect.TypeOf(T1{}),
		},
		// Unmarshal CBOR array of known sequence of data types, such as signed/maced/encrypted CWT, to []interface{}.
		{
			name:         "CBOR array to Go []interface{}",
			data:         hexDecode("88F51BFFFFFFFFFFFFFFFF3903E7FBC010666666666666581A0102030405060708090A0B0C0D0E0F101112131415161718191A782B54686520717569636B2062726F776E20666F78206A756D7073206F76657220746865206C617A7920646F67981A0102030405060708090A0B0C0D0E0F101112131415161718181819181AAD616261426163614361646144616561456166614661696149616D614E616E6141616F6147617061486171614A6172614C6173614D"),
			decodeToType: reflect.TypeOf([]any{}),
		},
		// Unmarshal CBOR array of known sequence of data types, such as signed/maced/encrypted CWT, to struct.
		{
			name:         "CBOR array to Go struct toarray",
			data:         hexDecode("88F51BFFFFFFFFFFFFFFFF3903E7FBC010666666666666581A0102030405060708090A0B0C0D0E0F101112131415161718191A782B54686520717569636B2062726F776E20666F78206A756D7073206F76657220746865206C617A7920646F67981A0102030405060708090A0B0C0D0E0F101112131415161718181819181AAD616261426163614361646144616561456166614661696149616D614E616E6141616F6147617061486171614A6172614C6173614D"),
			decodeToType: reflect.TypeOf(T3{}),
		},
	}
	for _, bm := range moreBenchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.SetBytes(int64(len(bm.data)))
			for b.Loop() {
				vPtr := reflect.New(bm.decodeToType).Interface()
				if err := dm.Unmarshal(bm.data, vPtr); err != nil {
					b.Fatal("Unmarshal:", err)
				}
			}
		})
	}
}

func BenchmarkMarshal(b *testing.B) {
	em := encMode()
	for _, bm := range encodeBenchmarks {
		for _, v := range bm.values {
			name := "Go " + reflect.TypeOf(v).String() + " to CBOR " + bm.name
			if reflect.TypeOf(v).Kind() == reflect.Struct {
				name = "Go " + reflect.TypeOf(v).Kind().String() + " to CBOR " + bm.name
			}
			b.Run(name, func(b *testing.B) {
				b.SetBytes(int64(len(bm.data)))
				for b.Loop() {
					if _, err := em.Marshal(v); err != nil {
						b.Fatal("Marshal:", err)
					}
				}
			})
		}
	}
	// Marshal map[string]interface{} to CBOR map
	m1 := map[string]any{
		"T":    true,
		"UI":   uint(18446744073709551615),
		"I":    -1000,
		"F":    -4.1,
		"B":    []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26},
		"S":    "The quick brown fox jumps over the lazy dog",
		"Slci": []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26},
		"Mss":  map[string]string{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E", "f": "F", "g": "G", "h": "H", "i": "I", "j": "J", "l": "L", "m": "M", "n": "N"},
	}
	// Marshal struct to CBOR map
	v1 := T1{ //nolint:dupl
		T:    true,
		UI:   18446744073709551615,
		I:    -1000,
		F:    -4.1,
		B:    []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26},
		S:    "The quick brown fox jumps over the lazy dog",
		Slci: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26},
		Mss:  map[string]string{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E", "f": "F", "g": "G", "h": "H", "i": "I", "j": "J", "l": "L", "m": "M", "n": "N"},
	}
	// Marshal []interface to CBOR array.
	slc := []any{
		true,
		uint(18446744073709551615),
		-1000,
		-4.1,
		[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26},
		"The quick brown fox jumps over the lazy dog",
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26},
		map[string]string{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E", "f": "F", "g": "G", "h": "H", "i": "I", "j": "J", "l": "L", "m": "M", "n": "N"},
	}
	// Marshal struct toarray to CBOR array, such as signed/maced/encrypted CWT.
	v3 := T3{ //nolint:dupl
		T:    true,
		UI:   18446744073709551615,
		I:    -1000,
		F:    -4.1,
		B:    []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26},
		S:    "The quick brown fox jumps over the lazy dog",
		Slci: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26},
		Mss:  map[string]string{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E", "f": "F", "g": "G", "h": "H", "i": "I", "j": "J", "l": "L", "m": "M", "n": "N"},
	}
	var moreBenchmarks = []struct {
		name  string
		value any
	}{
		{
			name:  "Go map[string]interface{} to CBOR map",
			value: m1,
		},
		{
			name:  "Go struct to CBOR map",
			value: v1,
		},
		{
			name:  "Go struct many fields all omitempty all empty to CBOR map",
			value: ManyFieldsAllOmitEmpty{},
		},
		{
			name:  "Go struct some fields all omitempty all empty to CBOR map",
			value: SomeFieldsAllOmitEmpty{},
		},
		{
			name: "Go struct many fields all omitempty all nonempty to CBOR map",
			value: ManyFieldsAllOmitEmpty{
				F01: 1, F02: 1, F03: 1, F04: 1, F05: 1, F06: 1, F07: 1, F08: 1, F09: 1, F10: 1, F11: 1, F12: 1, F13: 1, F14: 1, F15: 1, F16: 1,
				F17: 1, F18: 1, F19: 1, F20: 1, F21: 1, F22: 1, F23: 1, F24: 1, F25: 1, F26: 1, F27: 1, F28: 1, F29: 1, F30: 1, F31: 1, F32: 1,
			},
		},
		{
			name: "Go struct some fields all omitempty all nonempty to CBOR map",
			value: SomeFieldsAllOmitEmpty{
				F01: 1, F02: 1, F03: 1, F04: 1, F05: 1, F06: 1, F07: 1, F08: 1,
			},
		},
		{
			name:  "Go struct many fields one omitempty to CBOR map",
			value: ManyFieldsOneOmitEmpty{},
		},
		{
			name:  "Go struct some fields one omitempty to CBOR map",
			value: SomeFieldsOneOmitEmpty{},
		},
		{
			name:  "Go []interface{} to CBOR map",
			value: slc,
		},
		{
			name:  "Go struct toarray to CBOR array",
			value: v3,
		},
	}
	for _, bm := range moreBenchmarks {
		b.Run(bm.name, func(b *testing.B) {
			bz, _ := em.Marshal(bm.value)
			b.ResetTimer()
			b.SetBytes(int64(len(bz)))
			for b.Loop() {
				if _, err := em.Marshal(bm.value); err != nil {
					b.Fatal("Marshal:", err)
				}
			}
		})
	}
}

func BenchmarkUnmarshalMapToStruct(b *testing.B) {
	type S struct {
		A, B, C, D, E, F, G, H, I, J, K, L, M bool
	}

	var (
		allKnownFields   = hexDecode("ad6141f56142f56143f56144f56145f56146f56147f56148f56149f5614af5614bf5614cf5614df5") // {"A": true, ... "M": true }
		allUnknownFields = hexDecode("ad614ef5614ff56150f56151f56152f56153f56154f56155f56156f56157f56158f56159f5615af5") // {"N": true, ... "Z": true }
	)

	type ManyFields struct {
		AA, AB, AC, AD, AE, AF, AG, AH, AI, AJ, AK, AL, AM, AN, AO, AP, AQ, AR, AS, AT, AU, AV, AW, AX, AY, AZ bool
		BA, BB, BC, BD, BE, BF, BG, BH, BI, BJ, BK, BL, BM, BN, BO, BP, BQ, BR, BS, BT, BU, BV, BW, BX, BY, BZ bool
		CA, CB, CC, CD, CE, CF, CG, CH, CI, CJ, CK, CL, CM, CN, CO, CP, CQ, CR, CS, CT, CU, CV, CW, CX, CY, CZ bool
		DA, DB, DC, DD, DE, DF, DG, DH, DI, DJ, DK, DL, DM, DN, DO, DP, DQ, DR, DS, DT, DU, DV, DW, DX, DY, DZ bool
	}
	var manyFieldsOneKeyPerField []byte
	{
		// An EncOption that accepts a function to sort or shuffle keys might be useful for
		// cases like this. Here we are manually encoding the fields in reverse order to
		// target worst-case key-to-field matching.
		rt := reflect.TypeOf(ManyFields{})
		var buf bytes.Buffer
		if rt.NumField() > 255 {
			b.Fatalf("invalid test assumption: ManyFields expected to have no more than 255 fields, has %d", rt.NumField())
		}
		buf.WriteByte(0xb8)
		buf.WriteByte(byte(rt.NumField()))
		for i := 0; i < rt.NumField(); i++ {
			f := rt.Field(i)
			if len(f.Name) > 23 {
				b.Fatalf("invalid test assumption: field name %q longer than 23 bytes", f.Name)
			}
			buf.WriteByte(byte(0x60 + len(f.Name)))
			buf.WriteString(f.Name)
			buf.WriteByte(0xf5) // true
		}
		manyFieldsOneKeyPerField = buf.Bytes()
	}

	type input struct {
		name   string
		data   []byte
		into   any
		reject bool
	}

	for _, tc := range []*struct {
		name   string
		opts   cbor.DecOptions
		inputs []input
	}{
		{
			name: "default options",
			opts: decOptions(),
			inputs: []input{
				{
					name:   "all known fields",
					data:   allKnownFields,
					into:   S{},
					reject: false,
				},
				{
					name:   "all unknown fields",
					data:   allUnknownFields,
					into:   S{},
					reject: false,
				},
				{
					name:   "many fields one key per field",
					data:   manyFieldsOneKeyPerField,
					into:   ManyFields{},
					reject: false,
				},
			},
		},
	} {
		for _, in := range tc.inputs {
			b.Run(fmt.Sprintf("%s/%s", tc.name, in.name), func(b *testing.B) {
				dm, err := tc.opts.DecMode()
				if err != nil {
					b.Fatal(err)
				}

				dst := reflect.New(reflect.TypeOf(in.into)).Interface()

				b.ResetTimer()
				b.SetBytes(int64(len(in.data)))
				for b.Loop() {
					if err := dm.Unmarshal(in.data, dst); !in.reject && err != nil {
						b.Fatalf("unexpected error: %v", err)
					} else if in.reject && err == nil {
						b.Fatal("expected non-nil error")
					}
				}
			})
		}
	}
}
