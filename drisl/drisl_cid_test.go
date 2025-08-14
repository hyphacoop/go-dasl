package drisl_test

import (
	"bytes"
	"testing"

	"github.com/hyphacoop/go-dasl/drisl"
)

// Other CID tests are in drisl_test.go. This just sanity checks the constructors.

func TestCidFromString(t *testing.T) {
	s := "bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4"
	b := hexDecode("01551220adee2e8fb5459c9bcf07d7d78d1183bf40a7f60f57a54a19194801c9a27ead87")
	c, err := drisl.NewCidFromString(s)
	if err != nil {
		t.Error(err)
	}
	if c.String() != s {
		t.Errorf("want %s, got %s", s, c.String())
	}
	if !bytes.Equal(c.Bytes(), b) {
		t.Errorf("want %x, got %x", b, c.Bytes())
	}
}

func TestCidFromBytes(t *testing.T) {
	s := "bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4"
	b := hexDecode("01551220adee2e8fb5459c9bcf07d7d78d1183bf40a7f60f57a54a19194801c9a27ead87")
	c, err := drisl.NewCidFromBytes(b)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(c.Bytes(), b) {
		t.Errorf("want %x, got %x", b, c.Bytes())
	}
	if c.String() != s {
		t.Errorf("want %s, got %s", s, c.String())
	}
}
