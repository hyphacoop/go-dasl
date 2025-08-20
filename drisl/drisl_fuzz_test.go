package drisl_test

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/hyphacoop/go-dasl/cid"
	"github.com/hyphacoop/go-dasl/drisl"
	"pgregory.net/rapid"
)

func seeds() [][]byte {
	b, err := os.ReadFile("testdata/fuzz/cbor_seeds")
	if err != nil {
		panic(err)
	}
	hexSeeds := bytes.Split(b, []byte("\n"))
	seeds := make([][]byte, len(hexSeeds))
	for i, hs := range hexSeeds {
		seeds[i] = make([]byte, hex.DecodedLen(len(hs)))
		hex.Decode(seeds[i], hs)
	}
	return seeds
}

func FuzzUnmarshal(f *testing.F) {
	for _, seed := range seeds() {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, val []byte) {
		var v any
		err := drisl.Unmarshal(val, &v)
		if err == nil {
			result, err := drisl.Marshal(v)
			if err != nil {
				t.Errorf("Example %x produced marshaling error %v", val, err)
			} else if !bytes.Equal(result, val) {
				t.Errorf("got %x, want %x", result, val)
			}
		}
	})
}

type marshaler struct{ val []byte }

func (m marshaler) MarshalCBOR() ([]byte, error) {
	return m.val, nil
}

// FuzzMarshaler tests that cbor.Marshaler won't be accepted by drisl.Marshal unless it
// outputs valid DRISL.
func FuzzMarshaler(f *testing.F) {
	for _, seed := range seeds() {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, val []byte) {
		b, err := drisl.Marshal(marshaler{val})
		var v any
		if err == nil {
			err = drisl.Unmarshal(b, &v)
			if err != nil && validMarshalerError(err) {
				t.Errorf("Marshal produced invalid DRISL: %x -> %x -> %v", val, b, err)
			} else if !bytes.Equal(b, val) {
				t.Errorf("Marshal changed input: %x -> %x", val, b)
			}
		}
	})
}

// validMarshalerError checks whether the error from unmarshalling a Marshaler's output is worth raising.
// Some errors are not worth chasing down right now.
func validMarshalerError(err error) bool {
	// Invalid UTF-8 is allowed since it is still wellformed CBOR and DRISL
	if err.Error() == "cbor: invalid UTF-8 string" {
		return false
	}
	// CID errors are also allowed for now since it would be difficult to validate.
	// https://github.com/hyphacoop/go-dasl/issues/9
	if strings.HasPrefix(err.Error(), "invalid cid:") {
		return false
	}
	// Marshaler is allowed to output things that are too large for the default unmarshaller.
	// That's not invalid by any spec.
	if strings.HasPrefix(err.Error(), "cbor: exceeded max") {
		return false
	}
	// Not worth checking for now, it's unlikely a good Marshaler would do this
	if strings.HasPrefix(err.Error(), "cbor: found duplicate map key") {
		return false
	}
	return true
}

type unmarshaler struct {
	t   *testing.T
	val []byte
}

func (u *unmarshaler) UnmarshalCBOR([]byte) error {
	u.t.Errorf("invalid data hit unmarshaler: %x", u.val)
	return nil
}

func FuzzUnmarshaler(f *testing.F) {
	for _, seed := range seeds() {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, val []byte) {
		var v any
		err := drisl.Unmarshal(val, &v)
		if err == nil || !validMarshalerError(err) {
			// Valid so ignore
			return
		}
		u := unmarshaler{t, val}
		drisl.Unmarshal(val, &u)
		t.Logf("first unmarshal error: %v", err)
		// If UnmarshalCBOR gets called the test will fail
	})
}

func treeGenerator() *rapid.Generator[map[string]any] {
	terminatorGens := []*rapid.Generator[any]{
		rapid.Bool().AsAny(),
		rapid.String().AsAny(),
		rapid.Float64().AsAny(),
		rapid.Float32().AsAny(),
		rapid.Byte().AsAny(),
		rapid.Int8().AsAny(),
		rapid.Int16().AsAny(),
		rapid.Int32().AsAny(),
		rapid.Int64().AsAny(),
		rapid.Int().AsAny(),
		rapid.Uint8().AsAny(),
		rapid.Uint16().AsAny(),
		rapid.Uint32().AsAny(),
		rapid.Uint64().AsAny(),
		rapid.Uint().AsAny(),
		rapid.Rune().AsAny(),
		rapid.Custom(func(t *rapid.T) cid.Cid {
			c, err := cid.NewCidFromBytes(rapid.SliceOf(rapid.Byte()).Draw(t, "cid"))
			if err != nil {
				return cid.Cid{}
			}
			return c
		}).Filter(func(c cid.Cid) bool {
			return c.Defined()
		}).AsAny(),
	}
	generators := []*rapid.Generator[any]{}
	for _, terminator := range terminatorGens {
		sliceGen := rapid.SliceOf(terminator)
		generators = append(generators, terminator, sliceGen.AsAny())
	}
	generators = append(generators,
		rapid.Deferred(treeGenerator).AsAny(),
		rapid.SliceOf(rapid.Deferred(treeGenerator)).AsAny(),
	)
	generators = append(generators,
		rapid.MapOf(rapid.String(), rapid.Deferred(treeGenerator).AsAny()).AsAny(),
	)
	return rapid.MapOf(rapid.String(), rapid.OneOf(generators...))
}

func FuzzMarshal(f *testing.F) {
	treeGen := treeGenerator()
	f.Fuzz(rapid.MakeFuzz(func(t *rapid.T) {
		value := treeGen.Draw(t, "value")
		bz, err := drisl.Marshal(value)
		if err == nil {
			unmarshaled := map[string]any{}
			err := drisl.Unmarshal(bz, &unmarshaled)
			if err != nil {
				if strings.HasPrefix(err.Error(), "cbor: exceeded max") {
					// This error is okay
					return
				}
				t.Errorf("Struct %+v produced unmarshal error %v", value, err)
			} else if !deepEqualWithNumericConversion(value, unmarshaled) {
				t.Errorf("Expected %#v got %#v, (bz: %x)", value, unmarshaled, bz)
			}
		} else {
			t.Errorf("struct %+v couldn't be marshalled: %v", value, err)
		}
	}))
}

// deepEqualWithNumericConversion compares two values with special handling for integer/float types
func deepEqualWithNumericConversion(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	va, vb := reflect.ValueOf(a), reflect.ValueOf(b)

	// Handle integer conversions
	if isInteger(va) && isInteger(vb) {
		return integerEqual(va, vb)
	}
	// Floats
	if isFloat(va) && isFloat(vb) {
		return floatEqual(va, vb)
	}

	// Handle slices recursively
	if va.Kind() == reflect.Slice && vb.Kind() == reflect.Slice {
		if va.Len() != vb.Len() {
			return false
		}
		for i := 0; i < va.Len(); i++ {
			if !deepEqualWithNumericConversion(va.Index(i).Interface(), vb.Index(i).Interface()) {
				return false
			}
		}
		return true
	}

	// Handle maps recursively
	if va.Kind() == reflect.Map && vb.Kind() == reflect.Map {
		if va.Len() != vb.Len() {
			return false
		}
		for _, key := range va.MapKeys() {
			aVal := va.MapIndex(key)
			bVal := vb.MapIndex(key)
			if !bVal.IsValid() {
				return false
			}
			if !deepEqualWithNumericConversion(aVal.Interface(), bVal.Interface()) {
				return false
			}
		}
		return true
	}

	// For all other types use standard reflection comparison
	return reflect.DeepEqual(a, b)
}

func isInteger(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	}
	return false
}

func isFloat(v reflect.Value) bool {
	return v.Kind() == reflect.Float32 || v.Kind() == reflect.Float64
}

func integerEqual(a, b reflect.Value) bool {
	// Convert both to int64 for comparison, handling signed/unsigned conversions
	aInt, aOk := toInt(a)
	bInt, bOk := toInt(b)
	if !aOk || !bOk {
		return false
	}
	return aInt.Cmp(bInt) == 0
}

func floatEqual(a, b reflect.Value) bool {
	return a.Float() == b.Float()
}

func toInt(v reflect.Value) (*big.Int, bool) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return big.NewInt(v.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var i big.Int
		return i.SetUint64(v.Uint()), true
	}
	return nil, false
}
