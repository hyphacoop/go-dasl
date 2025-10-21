/*
Package masl is an implementation of MASL (Metadata for Arbitrary Structures & Links),
the metadata system from DASL.

MASL provides a CBOR-based metadata format for content-addressed and decentralized systems.
It enables self-contained, self-certified content distribution by bundling resources with
their metadata.

https://dasl.ing/masl.html

# Document Modes

MASL documents can operate in two modes:

Single Mode: Metadata for a single resource, identified by the Src field.
The document directly contains metadata fields for that resource.

Bundle Mode: Metadata for multiple resources, organized in a Resources map.
Each resource is keyed by its path (must start with "/") and contains its own metadata.

# Usage

Basic single-mode document:

	doc := masl.Document{
		Resource: masl.Resource{
			Src: someCid,
			ContentType: "text/html",
			ContentLanguage: "en",
		},
	}
	data, err := drisl.Marshal(doc)

Bundle-mode document with multiple resources:

	doc := masl.Document{
		Resources: map[string]*masl.Resource{
			"/": {
				Src: indexCid,
				ContentType: "text/html",
			},
			"/app.js": {
				Src: jsCid,
				ContentType: "application/javascript",
				Sourcemap: "/app.js.map",
			},
		},
	}

# Well-Known Attributes

MASL defines several categories of well-known attributes:

HTTP Response Headers: ContentType, ContentEncoding, ContentLanguage, etc.
These map to standard HTTP headers for web content delivery.

Web App Manifest: Name, Icons, Screenshots, BackgroundColor, etc.
These enable Progressive Web App (PWA) metadata.

Versioning: Prev field links to previous document version via CID.

Compatibility Fields:
  - CAR: Version and Roots for CAR file metadata
  - AT Protocol: Type field (typically "ing.dasl.masl")

# Validation

Documents can be validated using the Valid() method, which checks:
  - CAR version must be 0 or 1
  - AT type must be empty or "ing.dasl.masl"
  - Bundle mode: resource paths must start with "/"
  - Bundle mode: all resources must have Src CID
  - Bundle mode: sourcemap/speculation-rules must reference existing resources
  - Bundle mode: icon/screenshot src must reference existing resources
  - Single mode: icons/screenshots cannot have src references

# Encoding

MASL documents are encoded using DRISL (dag-cbor). All fields use omitempty
or omitzero to minimize encoding size. Unknown attributes are stored in the
Attributes map and preserved during roundtrip encoding.
*/
package masl

import (
	"strings"

	"github.com/hyphacoop/go-dasl/cid"
)

// Icon represents an icon entry in a Web App Manifest.
//
// In bundle mode, Src references a path in the Resources map (e.g., "/icon.png").
// In single mode, Src must be empty as there is no Resources map to reference.
//
// Sizes specifies image dimensions (e.g., "512x512" or "192x192 512x512").
// Purpose indicates icon usage context (e.g., "any", "maskable", "monochrome").
type Icon struct {
	Src     string `cbor:"src"`
	Sizes   string `cbor:"sizes,omitempty"`
	Purpose string `cbor:"purpose,omitempty"`
}

// Screenshot represents a screenshot entry in a Web App Manifest.
//
// In bundle mode, Src references a path in the Resources map (e.g., "/screenshot1.png").
// In single mode, Src must be empty as there is no Resources map to reference.
//
// Sizes specifies image dimensions (e.g., "1280x720").
// Label provides an accessible description of the screenshot.
// FormFactor indicates the display format (e.g., "wide", "narrow").
// Platform specifies the target platform (e.g., "windows", "macos", "android").
type Screenshot struct {
	Src        string `cbor:"src"`
	Sizes      string `cbor:"sizes,omitempty"`
	Label      string `cbor:"label,omitempty"`
	FormFactor string `cbor:"form_factor,omitempty"`
	Platform   string `cbor:"platform,omitempty"`
}

// Resource represents metadata for a single resource in a MASL document.
//
// Resources can exist standalone (in single mode) or as part of a bundle.
// Each resource contains metadata describing its content type, encoding,
// language, security policies, and other HTTP-like headers.
//
// The Src field identifies the resource content by CID. In bundle mode,
// all resources must have a defined Src. In single mode, Src identifies
// the document's single resource.
//
// HTTP response headers map to standard web headers and control how the
// resource should be served and processed:
//   - ContentType: MIME type (e.g., "text/html", "application/javascript")
//   - ContentEncoding: Compression format (e.g., "gzip", "br")
//   - ContentLanguage: Language code (e.g., "en", "fr")
//   - ContentSecurityPolicy: CSP directives for security
//   - Sourcemap: Path to source map file (must reference existing resource in bundle)
//   - SpeculationRules: Path to speculation rules (must reference existing resource in bundle)
//
// Web App Manifest fields enable PWA metadata:
//   - Name, ShortName: Application names
//   - Description: Application description
//   - Icons, Screenshots: Visual assets (paths must reference existing resources in bundle)
//   - BackgroundColor, ThemeColor: Color scheme
//   - Categories: App categories (e.g., "productivity", "games")
//
// The Attributes map captures any additional metadata not covered by
// well-known fields. All unknown CBOR fields during unmarshaling are
// preserved in this map, enabling extensibility while maintaining
// roundtrip fidelity.
type Resource struct {
	Src  cid.Cid `cbor:"src,omitzero"`
	Name string  `cbor:"name,omitempty"`

	// HTTP response headers
	ContentType           string `cbor:"content-type,omitempty"`
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

	// Web App Manifest fields
	BackgroundColor string       `cbor:"background_color,omitempty"`
	Categories      []string     `cbor:"categories,omitempty"`
	Description     string       `cbor:"description,omitempty"`
	Icons           []Icon       `cbor:"icons,omitempty"`
	ID              string       `cbor:"id,omitempty"`
	Screenshots     []Screenshot `cbor:"screenshots,omitempty"`
	ShortName       string       `cbor:"short_name,omitempty"`
	ThemeColor      string       `cbor:"theme_color,omitempty"`

	Attributes map[string]any `cbor:",unknown"`
}

// Document represents a MASL document, which can operate in single mode or bundle mode.
//
// Document embeds Resource, so all Resource fields (Src, Name, ContentType, etc.)
// are directly accessible on the Document. This means a Document IS-A Resource
// with additional fields for organizing multiple resources and compatibility.
//
// # Single Mode
//
// Single mode is used when Src is set and Resources is nil. The Document's Resource
// fields describe metadata for the single resource identified by Src:
//
//	doc := masl.Document{
//		Resource: masl.Resource{
//			Src: contentCid,
//			ContentType: "text/html",
//		},
//	}
//
// # Bundle Mode
//
// Bundle mode is used when Resources is non-nil. Each entry in Resources describes
// metadata for one resource in the bundle, keyed by path. Paths must start with "/":
//
//	doc := masl.Document{
//		Resources: map[string]*masl.Resource{
//			"/": { Src: indexCid, ContentType: "text/html" },
//			"/app.js": { Src: jsCid, ContentType: "application/javascript" },
//		},
//	}
//
// # Compatibility Fields
//
// Additional document-level fields support versioning and ecosystem compatibility:
//
// CAR Compatibility (Version, Roots):
//   - Version: CAR format version (must be 0 or 1)
//   - Roots: Root CIDs for CAR files
//
// AT Protocol Compatibility (Type):
//   - Type: AT Protocol type identifier (typically "ing.dasl.masl")
//
// Versioning (Prev):
//   - Prev: Links to previous document version CID, creating version chains
//
// Use IsBundle() to check the document mode, and Valid() to validate
// according to the MASL specification.
type Document struct {
	Resource
	Resources map[string]*Resource `cbor:"resources,omitempty"`

	// CAR compatibility fields
	Version int       `cbor:"version,omitempty"`
	Roots   []cid.Cid `cbor:"roots,omitempty"`

	// AT Protocol compatibility fields
	Type string `cbor:"$type,omitempty"`

	// Versioning fields
	Prev cid.Cid `cbor:"prev,omitzero"`
}

// IsBundle returns true if this document operates in bundle mode.
//
// Bundle mode is indicated by a non-nil Resources map, which contains
// multiple resources organized by path. Single mode documents have
// Resources == nil and describe metadata for one resource.
//
// Example:
//
//	if doc.IsBundle() {
//		// Access resources via doc.Resources map
//		rootResource := doc.Resources["/"]
//	} else {
//		// Access single resource via doc.Src and other Resource fields
//		fmt.Println(doc.Src)
//	}
func (d *Document) IsBundle() bool {
	return d.Resources != nil
}

// Valid validates the MASL document according to the MASL specification.
//
// Validation checks depend on document mode (single vs bundle) and include:
//
// CAR Compatibility:
//   - Version must be 0 or 1 (if specified)
//
// AT Protocol Compatibility:
//   - Type must be empty or "ing.dasl.masl" (if specified)
//
// Bundle Mode Validation:
//   - All resource paths must start with "/"
//   - All resource paths must be non-empty
//   - All resources must have a defined Src CID
//   - Sourcemap fields must reference existing resources in the Resources map
//   - SpeculationRules fields must reference existing resources in the Resources map
//   - Icon Src fields must reference existing resources in the Resources map
//   - Screenshot Src fields must reference existing resources in the Resources map
//
// Single Mode Validation:
//   - Icon Src fields must be empty (no Resources map to reference)
//   - Screenshot Src fields must be empty (no Resources map to reference)
//
// Returns true if the document is valid according to all applicable rules,
// false otherwise. This method does not return specific validation errors;
// it is intended for quick validity checks.
func (d *Document) Valid() bool {
	// Validate CAR compatibility fields
	if d.Version != 0 && d.Version != 1 {
		return false
	}

	// Validate AT compatibility fields
	if d.Type != "" && d.Type != "ing.dasl.masl" {
		return false
	}

	// Validate versioning fields
	// Prev field validation is handled by CID type itself

	if d.IsBundle() {
		// Bundle mode validation
		return d.validateBundle()
	} else {
		// Single mode validation
		return d.validateSingle()
	}
}

func (d *Document) validateBundle() bool {
	// Bundle mode: validate resources map
	for path, resource := range d.Resources {
		// Resource paths MUST start with /
		if path == "" || !strings.HasPrefix(path, "/") {
			return false
		}

		// Resources MUST have a src field
		if !resource.Src.Defined() {
			return false
		}

		// Validate HTTP header references within this resource
		if !d.validateResourceReferences(resource) {
			return false
		}
	}

	// Validate app manifest references (icons, screenshots)
	return d.validateAppManifestReferences()
}

func (d *Document) validateSingle() bool {
	// Single mode: app manifest icons/screenshots should not reference paths
	// since there's no resources map to resolve them against
	for _, icon := range d.Icons {
		if icon.Src != "" {
			return false
		}
	}

	for _, screenshot := range d.Screenshots {
		if screenshot.Src != "" {
			return false
		}
	}

	return true
}

func (d *Document) validateResourceReferences(resource *Resource) bool {
	// Validate sourcemap references
	if resource.Sourcemap != "" {
		if _, exists := d.Resources[resource.Sourcemap]; !exists {
			return false
		}
	}

	// Validate speculation-rules references
	if resource.SpeculationRules != "" {
		if _, exists := d.Resources[resource.SpeculationRules]; !exists {
			return false
		}
	}

	return true
}

func (d *Document) validateAppManifestReferences() bool {
	// Validate icon references
	for _, icon := range d.Icons {
		if icon.Src != "" {
			if _, exists := d.Resources[icon.Src]; !exists {
				return false
			}
		}
	}

	// Validate screenshot references
	for _, screenshot := range d.Screenshots {
		if screenshot.Src != "" {
			if _, exists := d.Resources[screenshot.Src]; !exists {
				return false
			}
		}
	}

	return true
}
