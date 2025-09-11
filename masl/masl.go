package masl

import (
	"github.com/hyphacoop/go-dasl/cid"
)

// HTTPHeaders contains all supported HTTP response headers for MASL
type HTTPHeaders struct {
	MediaType             string `cbor:"mediaType,omitempty"`
	ContentDisposition    string `cbor:"content-disposition,omitempty"`
	ContentEncoding       string `cbor:"content-encoding,omitempty"`
	ContentLanguage       string `cbor:"content-language,omitempty"`
	ContentSecurityPolicy string `cbor:"content-security-policy,omitempty"`
	Link                  string `cbor:"link,omitempty"`
	PermissionsPolicy     string `cbor:"permissions-policy,omitempty"`
	ReferrerPolicy        string `cbor:"referrer-policy,omitempty"`
	ServiceWorkerAllowed  string `cbor:"service-worker-allowed,omitempty"`
	Sourcemap             string `cbor:"sourcemap,omitempty"`
	SpeculationRules      string `cbor:"speculation-rules,omitempty"`
	SupportsLoadingMode   string `cbor:"supports-loading-mode,omitempty"`
	XContentTypeOptions   string `cbor:"x-content-type-options,omitempty"`
}

// Icon represents an icon entry in the app manifest
type Icon struct {
	Src     string `cbor:"src"`
	Sizes   string `cbor:"sizes,omitempty"`
	Purpose string `cbor:"purpose,omitempty"`
}

// Screenshot represents a screenshot entry in the app manifest
type Screenshot struct {
	Src        string `cbor:"src"`
	Sizes      string `cbor:"sizes,omitempty"`
	Label      string `cbor:"label,omitempty"`
	FormFactor string `cbor:"form_factor,omitempty"`
	Platform   string `cbor:"platform,omitempty"`
}

// AppManifest contains Web App Manifest metadata fields
type AppManifest struct {
	BackgroundColor string       `cbor:"background_color,omitempty"`
	Categories      []string     `cbor:"categories,omitempty"`
	Description     string       `cbor:"description,omitempty"`
	Icons           []Icon       `cbor:"icons,omitempty"`
	ID              string       `cbor:"id,omitempty"`
	Screenshots     []Screenshot `cbor:"screenshots,omitempty"`
	ShortName       string       `cbor:"short_name,omitempty"`
	ThemeColor      string       `cbor:"theme_color,omitempty"`
}

// CARCompatibility contains fields required for CAR file compatibility
type CARCompatibility struct {
	Version int        `cbor:"version,omitempty"`
	Roots   []*cid.Cid `cbor:"roots,omitempty"`
}

// ATCompatibility contains fields for AT Protocol compatibility
type ATCompatibility struct {
	Type string `cbor:"$type,omitempty"`
}

// Versioning contains fields for version tracking
type Versioning struct {
	Prev *cid.Cid `cbor:"prev,omitempty"`
}

type Resource struct {
	Src  *cid.Cid `cbor:"src,omitempty"`
	Name string   `cbor:"name,omitempty"`

	HTTPHeaders
	AppManifest

	Attributes map[string]string `cbor:",unknown"`
}

type Document struct {
	Resource
	Resources map[string]*Resource `cbor:"resources,omitempty"`

	// Additional document-level fields
	CARCompatibility
	ATCompatibility
	Versioning
}

func (r *Document) IsBundle() bool {
	return r.Resources != nil
}
