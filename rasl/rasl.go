// Package rasl implements the RASL URL scheme for retrieving content-addressed resources.
//
// Example:
//
//	rasl://bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4/?hint=berjon.com&hint=bsky.app
//
// https://dasl.ing/rasl.html
package rasl

import (
	"errors"
	"net/url"
	"strings"

	"github.com/hyphacoop/go-dasl/cid"
)

// URL stores all the information about a RASL URL.
// You can construct this manually or use Parse to get it from a string.
//
// If constructed manually, you are responsible for making sure it is valid.
type URL struct {
	// Cid represents the content.
	Cid cid.Cid

	// Hints is a slice of hosts.
	// Examples: domain.com, 1.2.3.4, 1.2.3.4:1234.
	//
	// Do not modify Hints while fetching data.
	Hints []string

	// Path is the optional URL path.
	// This can be used if the CID resolves to MASL data.
	Path string
}

// Parse parses out the information from a RASL URL, if it's valid.
func Parse(rawUrl string) (*URL, error) {
	if !strings.HasPrefix(rawUrl, "rasl://") {
		return nil, errors.New("invalid scheme")
	}
	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}
	// Validate
	if u.User != nil {
		return nil, errors.New("user info not allowed")
	}
	if u.Fragment != "" {
		return nil, errors.New("fragment not allowed")
	}

	var ru URL
	ru.Path = u.Path
	ru.Cid, err = cid.NewCidFromString(u.Host)
	if err != nil {
		return nil, err
	}

	// Extract hints
	// URL must be valid
	query, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return nil, err
	}
	if !query.Has("hint") {
		// No hints, which is allowed
		return &ru, nil
	}

	hints := make([]string, 0)
	for _, hint := range query["hint"] {
		// Required to ignore invalid hints
		if hint == "" {
			continue
		}
		_, err = parseHost(hint)
		if err == nil {
			hints = append(hints, hint)
		}
	}
	if len(hints) > 0 {
		ru.Hints = hints
	}
	return &ru, nil
}

// String creates a RASL URL from the struct data.
// It does not do validation.
func (ru *URL) String() string {
	var sb strings.Builder
	sb.WriteString("rasl://")
	sb.WriteString(ru.Cid.String())
	sb.WriteString(ru.Path)
	if len(ru.Hints) > 0 {
		sb.WriteByte('?')
		query := url.Values{"hint": ru.Hints}
		sb.WriteString(query.Encode())
	}
	return sb.String()
}
