// rasl implements the RASL URL scheme retrieving content-addressed resources.
//
// Example:
//
//	web+rasl://bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4;berjon.com,bsky.app/
//
// https://dasl.ing/rasl.html
package rasl

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/hyphacoop/go-dasl/cid"
)

type URL struct {
	// Cid represents the content.
	Cid cid.Cid

	// Hints is a slice of authorities.
	// Examples: domain.com, 1.2.3.4, 1.2.3.4:1234, user:password@example.com
	//
	// Do not modify Hints while fetching data.
	Hints []string

	// Path is the optional URL path.
	// This can be used if the CID resolves to MASL data.
	Path string
}

// Parse parses out the information from a RASL URL, if it's valid.
func Parse(rawUrl string) (*URL, error) {
	if !strings.HasPrefix(rawUrl, "web+rasl://") {
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
	if u.RawQuery != "" {
		return nil, errors.New("query not allowed")
	}
	if u.Fragment != "" {
		return nil, errors.New("fragment not allowed")
	}

	var ru URL
	ru.Path = u.Path

	// Extract cid
	semicolonIdx := strings.IndexByte(u.Host, ';')
	if semicolonIdx == -1 {
		ru.Cid, err = cid.NewCidFromString(u.Host)
		if err != nil {
			return nil, err
		}
		// No hints so we're done
		return &ru, nil
	} else {
		ru.Cid, err = cid.NewCidFromString(u.Host[:semicolonIdx])
		if err != nil {
			return nil, err
		}
	}

	// Extract hints
	ru.Hints = strings.Split(u.Host[semicolonIdx+1:], ",")
	for _, hint := range ru.Hints {
		// Make sure they are valid authorities
		if len(hint) == 0 {
			return nil, errors.New("cannot have semicolon with no hints")
		}
		_, _, err = parseAuthority(hint)
		if err != nil {
			return nil, fmt.Errorf("%v: %s", err, hint)
		}
	}
	return &ru, nil
}

func (ru *URL) String() string {
	var sb strings.Builder
	sb.WriteString("web+rasl://")
	sb.WriteString(ru.Cid.String())
	if len(ru.Hints) > 0 {
		sb.WriteByte(';')
		sb.WriteString(ru.Hints[0])
		for _, hint := range ru.Hints[1:] {
			sb.WriteByte(',')
			sb.WriteString(hint)
		}
	}
	sb.WriteString(ru.Path)
	return sb.String()
}
