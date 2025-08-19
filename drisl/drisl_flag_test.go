package drisl_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/hyphacoop/go-dasl/drisl"
)

var intRangeTests = []struct {
	name string
	in   any
}{
	{"2^63", uint64(math.MaxInt64 + 1)},
	{"big.Int(2^63)", new(big.Int).SetUint64(math.MaxInt64 + 1)},
	{"big.Int(-(2^63)-1)", new(big.Int).Sub(big.NewInt(math.MinInt64), big.NewInt(1))},
}

func TestIntRangeMarshal(t *testing.T) {
	em, _ := drisl.EncOptions{Int64RangeOnly: true}.EncMode()
	for _, tt := range intRangeTests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := em.Marshal(uint64(math.MaxInt64 + 1))
			if err == nil {
				t.Errorf("%x - want error", b)
				return
			}
			t.Log(err)
		})
	}
}

func TestIntRangeUnmarshal(t *testing.T) {
	dm, _ := drisl.DecOptions{Int64RangeOnly: true}.DecMode()
	var v any
	err := dm.Unmarshal(hexDecode("1b8000000000000000"), &v)
	if err == nil {
		t.Errorf("2^63: %v - want error", v)
		return
	}
	t.Log(err)
}
func TestIntRangeUnmarshal2(t *testing.T) {
	dm, _ := drisl.DecOptions{Int64RangeOnly: true}.DecMode()
	var v any
	err := dm.Unmarshal(hexDecode("3b8000000000000000"), &v)
	if err == nil {
		t.Errorf("-(2^63): %v - want error", v)
		return
	}
	t.Log(err)
}

func TestUndefinedUnmarshal(t *testing.T) {
	dm, _ := drisl.DecOptions{AllowUndefined: true}.DecMode()
	var v any
	err := dm.Unmarshal([]byte{0xf7}, &v)
	if err != nil || v != nil {
		t.Errorf("got %v, %v - want error", v, err)
	}
}
