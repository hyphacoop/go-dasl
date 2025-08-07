package drisl_test

import (
	"bytes"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/hyphacoop/go-dasl/drisl"
	"pgregory.net/rapid"
)

var seeds = [][]byte{
	// from cid.json:
	hexDecode("d82a582500015512205891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03"), // Valid CID
	hexDecode("d82a4100"), // Simple invalid CID
	// from floats.json
	hexDecode("fb3ff8000000000000"), // A valid float
	hexDecode("fb7ff8000000000000"), // An invalid NaN
	// from utf8.json
	hexDecode("7873d8a7d984d986d8b561f09f94a5f09f918bf09f8fbc5acda7cc91cc93cca4cd9461cc88cc88cc87cd96ccad6ccdaecc92cdab67cc8ccc9acc97cd9a6fcc94cdaecc87cd90cc87cc99f09fa79fe2808de29980efb88fe2808d20f09f8fb3efb88fe2808de29aa7efb88fe794b0e2808b20e2808b"), // Bunch of UTF-8 characters
	hexDecode("62c328"), // Invalid utf-8
	// A fully invalid string
	[]byte("She sells seashells by the sea shore"),
}

func FuzzUnmarshal(f *testing.F) {
	for _, seed := range seeds {
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

func treeGenerator() *rapid.Generator[map[any]any] {
	terminatorGens := []*rapid.Generator[any]{
		rapid.Bool().AsAny(),
		rapid.Float64().AsAny(),
		rapid.Int64().AsAny(),
		rapid.String().AsAny(),
		rapid.Custom(func(t *rapid.T) time.Time {
			sec, nsec := rapid.Int64().Draw(t, "sec"), rapid.Int64().Draw(t, "nsec")
			return time.Unix(sec, nsec)
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
	return rapid.MapOf(rapid.OneOf(terminatorGens...), rapid.OneOf(generators...))
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
				t.Errorf("Struct %+v produced unmarshal error %v", value, err)
			} else if !deepEqualWithNumericConversion(value, unmarshaled) {
				t.Errorf("Expected %#v got %#v, (bz: %x)", value, unmarshaled, bz)
			}
		}
	}))
}

// deepEqualWithNumericConversion compares two values with special handling for integer types
func deepEqualWithNumericConversion(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	va, vb := reflect.ValueOf(a), reflect.ValueOf(b)

	// Handle integer conversions (but leave floats alone)
	if isInteger(va) && isInteger(vb) {
		return integerEqual(va, vb)
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

	// For all other types (including floats), use standard reflection comparison
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

func integerEqual(a, b reflect.Value) bool {
	// Convert both to int64 for comparison, handling signed/unsigned conversions
	aInt, aOk := toInt64(a)
	bInt, bOk := toInt64(b)

	if !aOk || !bOk {
		return false
	}

	return aInt == bInt
}

func toInt64(v reflect.Value) (int64, bool) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// Handle potential overflow when converting uint to int64
		uval := v.Uint()
		if uval <= math.MaxInt64 {
			return int64(uval), true
		}
		return 0, false
	}
	return 0, false
}
