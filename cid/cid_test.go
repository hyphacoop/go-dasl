package cid_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/hyphacoop/go-dasl/cid"
)

// More CID tests are in drisl_test.go

func hexDecode(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func TestCidFromString(t *testing.T) {
	s := "bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4"
	b := hexDecode("01551220adee2e8fb5459c9bcf07d7d78d1183bf40a7f60f57a54a19194801c9a27ead87")
	c, err := cid.NewCidFromString(s)
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
	c, err := cid.NewCidFromBytes(b)
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

func TestCidWithInvalidMultibase(t *testing.T) {
	s := "zb2rhiMENf9e7DtGrsW46yLSw743GN1T6g7QFjUvXCxmvNnSr"
	c, err := cid.NewCidFromString(s)
	if err == nil {
		t.Errorf("invalid CID was parsed from string: %v", c)
	}
	t.Log(err)
}
