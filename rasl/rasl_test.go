package rasl_test

import (
	"reflect"
	"testing"

	"github.com/hyphacoop/go-dasl/cid"
	"github.com/hyphacoop/go-dasl/rasl"
)

var parseTests = []struct {
	name   string
	in     string
	out    *rasl.URL
	outStr string // Leave blank for equal to in
}{
	{
		"regular",
		"rasl://bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4/?hint=berjon.com&hint=bsky.app",
		&rasl.URL{
			Cid:   cid.MustNewCidFromString("bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4"),
			Hints: []string{"berjon.com", "bsky.app"},
			Path:  "/",
		},
		"",
	}, {
		"no hints",
		"rasl://bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4/",
		&rasl.URL{
			Cid:  cid.MustNewCidFromString("bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4"),
			Path: "/",
		},
		"",
	}, {
		"no hints no path",
		"rasl://bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4",
		&rasl.URL{
			Cid: cid.MustNewCidFromString("bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4"),
		},
		"",
	}, {
		"many hints",
		"rasl://bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4/?hint=berjon.com&hint=1.1.1.1&hint=1.2.3.4%3A1234&hint=%5B%3A%3A1%5D&hint=%5B%3A%3A1%5D%3A1234",
		&rasl.URL{
			Cid: cid.MustNewCidFromString("bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4"),
			Hints: []string{
				"berjon.com", "1.1.1.1", "1.2.3.4:1234", "[::1]", "[::1]:1234",
			},
			Path: "/",
		},
		"",
	},
	{
		"blank hint",
		"rasl://bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4/?hint=",
		&rasl.URL{
			Cid:  cid.MustNewCidFromString("bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4"),
			Path: "/",
		},
		"rasl://bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4/",
	},
	{
		"non-host hint",
		"rasl://bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4/?hint=user%3Apass@example.com",
		&rasl.URL{
			Cid:  cid.MustNewCidFromString("bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4"),
			Path: "/",
		},
		"rasl://bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4/",
	}, {
		"other query string elements",
		"rasl://bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4/?foo=bar",
		&rasl.URL{
			Cid:  cid.MustNewCidFromString("bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4"),
			Path: "/",
		},
		"rasl://bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4/",
	},
}

func TestParse(t *testing.T) {
	for _, tt := range parseTests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := rasl.Parse(tt.in)
			if err != nil || !reflect.DeepEqual(tt.out, u) {
				t.Errorf("Parse(%s) = %#v, %v - want %#v", tt.in, u, err, tt.out)
			}
			if (tt.outStr == "" && u.String() != tt.in) || (tt.outStr != "" && u.String() != tt.outStr) {
				t.Errorf("String: got %s - want %s", u, tt.in)
			}
		})
	}
}

var badParseTests = []struct {
	name string
	in   string
}{
	{
		"port",
		"rasl://bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4:80",
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
