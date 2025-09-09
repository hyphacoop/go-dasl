package rasl_test

import (
	"reflect"
	"testing"

	"github.com/hyphacoop/go-dasl/cid"
	"github.com/hyphacoop/go-dasl/rasl"
)

var parseTests = []struct {
	name string
	in   string
	out  *rasl.URL
}{
	{
		"regular",
		"web+rasl://bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4;berjon.com,bsky.app/",
		&rasl.URL{
			Cid:   cid.MustNewCidFromString("bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4"),
			Hints: []string{"berjon.com", "bsky.app"},
			Path:  "/",
		},
	}, {
		"no hints",
		"web+rasl://bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4/",
		&rasl.URL{
			Cid:  cid.MustNewCidFromString("bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4"),
			Path: "/",
		},
	}, {
		"no hints no path",
		"web+rasl://bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4",
		&rasl.URL{
			Cid: cid.MustNewCidFromString("bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4"),
		},
	}, {
		"many hints",
		"web+rasl://bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4;berjon.com,1.1.1.1,1.2.3.4:1234,[::1],[::1]:1234/",
		&rasl.URL{
			Cid: cid.MustNewCidFromString("bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4"),
			Hints: []string{
				"berjon.com", "1.1.1.1", "1.2.3.4:1234", "[::1]", "[::1]:1234",
			},
			Path: "/",
		},
	},
}

func TestParse(t *testing.T) {
	for _, tt := range parseTests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := rasl.Parse(tt.in)
			if err != nil || !reflect.DeepEqual(tt.out, u) {
				t.Errorf("Parse(%s) = %#v, %v - want %#v", tt.in, u, err, tt.out)
			}
		})
	}
}

var badParseTests = []struct {
	name string
	in   string
}{
	{
		"no hints semicolon",
		"web+rasl://bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4;/",
	},
}

func TestParseBad(t *testing.T) {
	for _, tt := range badParseTests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := rasl.Parse(tt.in)
			if err == nil {
				t.Errorf("Parse(%s) = %#v - want error", tt.in, u)
			}
		})
	}
}
