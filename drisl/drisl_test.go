package drisl_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/hyphacoop/cbor/v2"
	"github.com/hyphacoop/go-dasl/cid"
	"github.com/hyphacoop/go-dasl/drisl"
)

type daslTestCase struct {
	Type string
	Data string
	Tags []string
	Name string
}

func TestDaslJson(t *testing.T) {
	// Tests from https://github.com/hyphacoop/dasl-testing
	// Parse all the JSON files and run all the relevant tests as subtests
	// Kind of like table-driven testing, but on the fly
	// https://go.dev/wiki/TableDrivenTests

	err := filepath.WalkDir("dasl-testing/fixtures/cbor/", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, ".json") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var tests []*daslTestCase
		if err := json.Unmarshal(b, &tests); err != nil {
			return err
		}
		runTests(t, tests)
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func runTests(t *testing.T, tests []*daslTestCase) {
	// Use decoder with increased depth to pass depth test
	dec, err := drisl.DecOptions{
		MaxNestedLevels: 3001,
	}.DecMode()
	if err != nil {
		panic(err)
	}

	// Pass datetime test
	enc, err := drisl.EncOptions{
		Time: drisl.TimeModeReject,
	}.EncMode()
	if err != nil {
		panic(err)
	}

	for _, test := range tests {
		isRelevantTest := slices.ContainsFunc(test.Tags, func(tag string) bool {
			return tag == "basic" || tag == "dag-cbor" || tag == "dasl-cid"
		})
		if !isRelevantTest {
			continue
		}
		testData := hexDecode(test.Data)
		test.Name = fmt.Sprintf("%s-%s", test.Type, test.Name)

		switch test.Type {
		case "roundtrip":
			t.Run(test.Name, func(t *testing.T) {
				var v any
				if err := dec.Unmarshal(testData, &v); err != nil {
					t.Errorf("Unmarshal error: %v", err)
					return
				}
				b, err := enc.Marshal(v)
				if err != nil {
					t.Errorf("Marshal error: %v", err)
					return
				}
				if !bytes.Equal(testData, b) {
					t.Errorf("got %x, want %x", b, testData)
				}
			})
			// t.Run(test.Name+"-Valid", func(t *testing.T) {
			// 	if !drisl.Valid(testData) {
			// 		t.Errorf("got false, want true")
			// 	}
			// })
		case "invalid_in":
			t.Run(test.Name, func(t *testing.T) {
				var v any
				if err := dec.Unmarshal(testData, &v); err == nil {
					t.Error("Unmarshal didn't raise an error")
				}
			})
			// t.Run(test.Name+"-Valid", func(t *testing.T) {
			// 	if drisl.Valid(testData) {
			// 		t.Errorf("got true, want false")
			// 	}
			// })
		case "invalid_out":
			t.Run(test.Name, func(t *testing.T) {
				// Decode data with neutral CBOR decoder, then confirm it cannot be
				// encoded by the strict DRISL encoder.
				var v any
				if err := cbor.Unmarshal(testData, &v); err != nil {
					t.Errorf("cbor.Unmarshal failed unexpectedly: %v", err)
					return
				}
				if _, err := enc.Marshal(v); err == nil {
					t.Error("Marshal didn't raise an error")
				}
			})
			// t.Run(test.Name+"-Valid", func(t *testing.T) {
			// 	if drisl.Valid(testData) {
			// 		t.Errorf("got true, want false")
			// 	}
			// })
		default:
			panic(fmt.Errorf("unknown test type '%s'", test.Type))
		}
	}
}

var marshalTests = []struct {
	name string
	in   any
	out  string
}{
	{"small big.Int", big.NewInt(1), "01"},
	{"small negative big.Int", big.NewInt(-1), "20"},
	{"float32", float32(1.5), "fb3ff8000000000000"},
	{"reduceable int32", int32(1), "01"},
	{"cid tag", cbor.Tag{
		Number:  drisl.CidTagNumber,
		Content: hexDecode("00015512205891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03")},
		"d82a582500015512205891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03"},
	{"time.Time", time.Unix(0, 0).UTC(), "74313937302d30312d30315430303a30303a30305a"},
}

// TestMarshal tests encoding Go objects with more versatility than the DASL test suite.
// Objects that are invalid to encode are not tested in this function.
func TestMarshal(t *testing.T) {
	for _, tt := range marshalTests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := drisl.Marshal(tt.in)
			if !bytes.Equal(hexDecode(tt.out), b) || err != nil {
				t.Errorf(`Marshal(%#v) = %x, %v, want match for %s, nil`, tt.in, b, err, tt.out)
			}
		})
	}
}

func TestBigBigInt(t *testing.T) {
	var i big.Int
	i.SetString("18446744073709551616", 10)
	b, err := drisl.Marshal(i)
	if err == nil {
		t.Errorf(`Marshal(big.Int(2^64)) = %x, %v, want error`, b, err)
	}
}

func TestMaxCborInt(t *testing.T) {
	var i int64
	err := drisl.Unmarshal(hexDecode("1b8000000000000000"), &i)
	if err == nil {
		t.Errorf("Unmarshal(2^63) into int64, got %d wanted error", i)
		return
	}
	t.Log(err)
}

// TestFloat32UnmarshalSafe tests decoding a float64 into a float32 for a value that fits.
func TestFloat32UnmarshalSafe(t *testing.T) {
	var v float32
	if err := drisl.Unmarshal(hexDecode("fb3ff8000000000000"), &v); err != nil {
		t.Errorf(`Unmarshal(1.5) returned err: %v`, err)
		return
	}
	if v != 1.5 {
		t.Errorf(`Unmarshal(1.5) = %f, want 1.5`, v)
	}
}

// TestFloat32UnmarshalUnsafe tests decoding a float64 into a float32 for a value that doesn't fit.
// https://github.com/hyphacoop/go-dasl/issues/7
func TestFloat32UnmarshalUnsafe(t *testing.T) {
	var v float32
	if err := drisl.Unmarshal(hexDecode("fb3ff8000000000001"), &v); err == nil {
		t.Errorf(`Unmarshal(1.5000000000000002) = %f, wanted error`, v)
	}
}

func TestIntUnmarshalUnsafe(t *testing.T) {
	var v uint8
	if err := drisl.Unmarshal(hexDecode("190100"), &v); err == nil {
		t.Errorf("Unmarshal(256) = %d, wanted error", v)
	}
}

func TestUnassignedSimpleValueMarshal(t *testing.T) {
	b, err := drisl.Marshal(cbor.SimpleValue(0))
	if err == nil {
		t.Errorf(`Marshal(SimpleValue(0)) = %x, %v, want error`, b, err)
	}
}

func TestUnassignedSimpleValueUnmarshal(t *testing.T) {
	var v any
	if err := drisl.Unmarshal([]byte{0xe0}, &v); err == nil {
		t.Errorf("Unmarshal(SimpleValue(0)) = %v, wanted error", v)
	}
}

func TestBuiltinTagUnmarshal(t *testing.T) {
	var v any
	// Decode tag type 1 (epoch int)
	// Normally this would decode into a time.Time
	// But here should fail because tags are banned
	if err := drisl.Unmarshal(hexDecode("c11a514b67b0"), &v); err == nil {
		t.Errorf("Unmarshal(epoch tag) = %v, wanted error", v)
	}
}

func TestTimeStringUnmarshal(t *testing.T) {
	var v any
	if err := drisl.Unmarshal(hexDecode("c07819323032352d30352d32365431363a31383a31372d30343a3030"), &v); err == nil {
		t.Errorf("Unmarshal(time tag) = %v, wanted error", v)
	}
}

func TestCborTagUnmarshal(t *testing.T) {
	var v cbor.Tag
	if err := drisl.Unmarshal(hexDecode("c11a514b67b0"), &v); err == nil {
		t.Errorf("Unmarshal tag into cbor.Tag: got %v, wanted error", v)
	}
}

func TestCborTagMarshal(t *testing.T) {
	v := cbor.Tag{
		Number:  1,
		Content: 1363896240,
	}
	b, err := drisl.Marshal(&v)
	if err == nil {
		t.Errorf(`Marshal = %x, %v, want error`, b, err)
		return
	}
	t.Log(err)
}

func TestCidUnmarshal(t *testing.T) {
	var v cid.Cid
	if err := drisl.Unmarshal(hexDecode("d82a582500015512205891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03"), &v); err != nil {
		t.Errorf("Unmarshal(cid) into Cid: got error: %v", err)
	}
}

func TestCidUnmarshalAny(t *testing.T) {
	var v any
	if err := drisl.Unmarshal(hexDecode("d82a582500015512205891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03"), &v); err != nil {
		t.Errorf("Unmarshal(cid) into any: got error: %v", err)
	}
	if reflect.TypeOf(v) != reflect.TypeOf(cid.Cid{}) {
		t.Errorf("Unmarshal(cid) into any: got type %s", reflect.TypeOf(v).String())
	}
}

func TestLongNegIntUnmarshal(t *testing.T) {
	var v any
	if err := drisl.Unmarshal(hexDecode("3800"), &v); err == nil {
		t.Errorf("Unmarshal(-1_0) = %v, wanted error", v)
	}
}

func TestStructMapNullUnmarshal(t *testing.T) {
	// This CBOR map has a null key, which is illegal
	// Make sure it can't be unmarshalled, even to a struct
	var v struct {
		A int `cbor:""`
	}
	err := drisl.Unmarshal(hexDecode("a1f630"), &v)
	if err == nil {
		t.Errorf("Unmarshal map with null key into struct = %v, wanted error", v)
		return
	}
	t.Log(err)
}

func TestKeyasintMarshal(t *testing.T) {
	v := struct {
		A int `cbor:"1,keyasint"`
	}{1}
	b, err := drisl.Marshal(v)
	if err == nil {
		t.Errorf(`Marshal(keyasint) = %x, %v, want error`, b, err)
	}
}

func TestKeyasintUnmarshal(t *testing.T) {
	var v struct {
		A int `cbor:"1,keyasint"`
	}
	err := drisl.Unmarshal(hexDecode("a10101"), &v)
	if err == nil {
		t.Errorf("Unmarshal map with int key into keyasint struct = %v, wanted error", v)
		return
	}
	t.Log(err)
}

func TestStructMapOrderUnmarshal(t *testing.T) {
	var v struct {
		A int `cbor:"a"`
		B int `cbor:"b"`
	}
	err := drisl.Unmarshal(hexDecode("a2616201616102"), &v)
	if err == nil {
		t.Errorf(`Unmarshal({"b":1,"a":2}) = %v, wanted error`, v)
		return
	}
	t.Log(err)
}

func TestInvalidCidUnmarshal(t *testing.T) {
	var v any
	err := drisl.Unmarshal(hexDecode("d82a623030"), &v)
	if err == nil {
		t.Errorf("Unmarshal(bad Cid) = %v, wanted error", v)
		return
	}
	t.Log(err)
}

type invalidMarshaler struct{ v []byte }

func (im invalidMarshaler) MarshalCBOR() ([]byte, error) {
	return im.v, nil
}

var marshalerTests = []struct {
	name string
	in   []byte
}{
	{"float32", hexDecode("fa3fc00000")},
	{"map with int key", hexDecode("a10101")},
	{"banned tag", hexDecode("c11a514b67b0")},
}

func TestInvalidMarshaler(t *testing.T) {
	for _, tt := range marshalerTests {
		t.Run(tt.name, func(t *testing.T) {
			im := invalidMarshaler{tt.in}
			b, err := drisl.Marshal(&im)
			if err == nil {
				t.Errorf(`%x - want error`, b)
				return
			}
			t.Log(err)
		})
	}
}

func TestRawTagMarshal(t *testing.T) {
	v := cbor.RawTag{Number: 123, Content: []byte{0x00}}
	b, err := drisl.Marshal(&v)
	if err == nil {
		t.Errorf(`Marshal = %x, %v, want error`, b, err)
		return
	}
	t.Log(err)
}

func TestCidNullUnmarshal(t *testing.T) {
	var c cid.Cid
	err := drisl.Unmarshal([]byte{0xf6}, &c)
	if err == nil {
		t.Errorf("Unmarshal null into Cid = %v - want error", c)
		return
	}
	t.Log(err)
}

func TestFloat64DetailedUnmarshal(t *testing.T) {
	var v float64
	// This float cannot be marshaled as a float32
	// So this tests that the float32 checks are not being triggered for a float64 receiver.
	err := drisl.Unmarshal(hexDecode("fbc010666666666666"), &v)
	if err != nil {
		t.Errorf("Unmarshal(-4.1) into float64 error: %v", err)
	}
}

func TestTimeStringUnmarshalTime(t *testing.T) {
	var v time.Time
	err := drisl.Unmarshal(hexDecode("74313937302d30312d30315430303a30303a30305a"), &v)
	if err != nil {
		t.Fatalf("Unmarshal(time str) to Time - error: %v", err)
	}
	if !v.Equal(time.Unix(0, 0)) {
		t.Fatalf("got %v want %v", v, time.Unix(0, 0).UTC())
	}
}

func TestRawMessageInvalidMarshal(t *testing.T) {
	var rm drisl.RawMessage = hexDecode("a10101")
	b, err := drisl.Marshal(rm)
	if err == nil {
		t.Fatalf("Marshal(RawMessage(invalid)) = %x - want error", b)
	}
	t.Log(err)
}

func TestRawMessageValidMarshal(t *testing.T) {
	var rm drisl.RawMessage = []byte{0x00}
	_, err := drisl.Marshal(rm)
	if err != nil {
		t.Fatalf("Marshal(RawMessage(valid)) got error: %v", err)
	}
}

func TestRawCidUnmarshal(t *testing.T) {
	var v any
	dm, _ := drisl.DecOptions{UseRawCid: true}.DecMode()
	err := dm.Unmarshal(hexDecode("d82a582300122022ad631c69ee983095b5b8acd029ff94aff1dc6c48837878589a92b90dfea317"), &v)
	if err != nil {
		t.Fatalf("Unmarshal(RawCid) failed: %v", err)
	}
	cidBytes := hexDecode("122022ad631c69ee983095b5b8acd029ff94aff1dc6c48837878589a92b90dfea317")
	if !bytes.Equal(v.(cid.RawCid), cidBytes) {
		t.Fatalf("got %x want %x", v, cidBytes)
	}
}

func TestRawCidMarshal(t *testing.T) {
	rc := cid.RawCid(hexDecode("122022ad631c69ee983095b5b8acd029ff94aff1dc6c48837878589a92b90dfea317"))
	b, err := drisl.Marshal(rc)
	if err != nil {
		t.Fatalf("Marshal(RawCid) failed: %v", err)
	}
	cborCid := hexDecode("d82a582300122022ad631c69ee983095b5b8acd029ff94aff1dc6c48837878589a92b90dfea317")
	if !bytes.Equal(b, cborCid) {
		t.Fatalf("got %x want %x", b, cborCid)
	}
}

func TestNoFloatsMarshal(t *testing.T) {
	var i float64 = 123.321
	enc, _ := drisl.EncOptions{NoFloats: true}.EncMode()
	b, err := enc.Marshal(i)
	if err == nil {
		t.Fatalf("want error, got %x", b)
	}
	t.Log(err)
}

func TestNoFloatsUnmarshal(t *testing.T) {
	var v any
	dec, _ := drisl.DecOptions{NoFloats: true}.DecMode()
	err := dec.Unmarshal(hexDecode("fb3ff8000000000000"), &v)
	if err == nil {
		t.Fatalf("want error")
	}
}
