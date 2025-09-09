package cid_test

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/hyphacoop/go-dasl/cid"
)

// More CID tests are in drisl_test.go

var (
	cidStr    = "bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4"
	cidBytes  = hexDecode("01551220adee2e8fb5459c9bcf07d7d78d1183bf40a7f60f57a54a19194801c9a27ead87")
	cidDigest = hexDecode("adee2e8fb5459c9bcf07d7d78d1183bf40a7f60f57a54a19194801c9a27ead87")
	cidCid    = MustCid(cidStr)
)

func MustCid(s string) cid.Cid {
	c, err := cid.NewCidFromString(s)
	if err != nil {
		panic(err)
	}
	return c
}

func hexDecode(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func TestNewCidFromString(t *testing.T) {
	c, err := cid.NewCidFromString(cidStr)
	if err != nil {
		t.Fatal(err)
	}
	if c.String() != cidStr {
		t.Fatalf("want %s, got %s", cidStr, c.String())
	}
	if !bytes.Equal(c.Bytes(), cidBytes) {
		t.Fatalf("want %x, got %x", cidBytes, c.Bytes())
	}
}

func TestNewCidFromBytes(t *testing.T) {
	c, err := cid.NewCidFromBytes(cidBytes)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(c.Bytes(), cidBytes) {
		t.Fatalf("want %x, got %x", cidBytes, c.Bytes())
	}
	if c.String() != cidStr {
		t.Fatalf("want %s, got %s", cidStr, c.String())
	}
}

func TestCidWithInvalidMultibase(t *testing.T) {
	s := "zb2rhiMENf9e7DtGrsW46yLSw743GN1T6g7QFjUvXCxmvNnSr"
	c, err := cid.NewCidFromString(s)
	if err == nil {
		t.Fatalf("invalid CID was parsed from string: %v", c)
	}
	t.Log(err)
}

func TestNewCidFromReader(t *testing.T) {
	r := bytes.NewReader(append(cidBytes, []byte("foobar")...))
	c, err := cid.NewCidFromReader(r)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(c.Bytes(), cidBytes) {
		t.Fatalf("want %x, got %x", cidBytes, c.Bytes())
	}
	if b, _ := r.ReadByte(); b != 'f' {
		t.Fatalf("can't read later data")
	}
}

func TestNewCidFromReader2(t *testing.T) {
	f, err := os.Open("testdata/cid_file")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	br := bufio.NewReader(f)
	c, err := cid.NewCidFromReader(br)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(c.Bytes(), cidBytes) {
		t.Fatalf("want %x, got %x", cidBytes, c.Bytes())
	}
	trailer, err := io.ReadAll(br)
	if err != nil {
		t.Fatalf("can't read later data, got error: %v", err)
	}
	if !bytes.Equal(trailer, []byte("foobar")) {
		t.Fatalf("trailing data: want %x got %x", []byte("foobar"), trailer)
	}
}

func TestNewCidFromInfo(t *testing.T) {
	c, err := cid.NewCidFromInfo(cid.CodecRaw, cid.HashTypeSha256, cidDigest)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(c.Bytes(), cidBytes) {
		t.Fatalf("want %x, got %x", cidBytes, c.Bytes())
	}
}

func TestHashSize(t *testing.T) {
	if cidCid.HashSize() != 32 {
		t.Errorf(".HashSize() = %d, want 32", cidCid.HashSize())
	}
}

func TestDigest(t *testing.T) {
	if !bytes.Equal(cidCid.Digest(), cidDigest) {
		t.Errorf(".Digest() = %x, want %x", cidCid.Digest(), cidDigest)
	}
}

func TestUnmarshalText(t *testing.T) {
	var c cid.Cid
	err := c.UnmarshalText([]byte(cidStr))
	if err != nil {
		t.Error(err)
	} else if c.String() != cidStr {
		t.Fatalf("want %s, got %s", cidStr, c.String())
	}
}

func TestBinary(t *testing.T) {
	b, _ := cidCid.MarshalBinary()
	b2, _ := cidCid.AppendBinary(nil)
	if !bytes.Equal(b, b2) {
		t.Errorf("MarshalBinary and AppendBinary are not equal: %x - %x", b, b2)
	}
}

func TestJson(t *testing.T) {
	j, err := json.Marshal(cidCid)
	if err != nil {
		t.Fatal(err)
	}
	if string(j) != `{"/":"bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4"}` {
		t.Fatalf("bad json: %s", j)
	}
}

func TestMarshalJson(t *testing.T) {
	j, _ := cidCid.MarshalJSON()
	if len(j) != cap(j) {
		t.Fatalf("len %d, cap %d", len(j), cap(j))
	}
}
