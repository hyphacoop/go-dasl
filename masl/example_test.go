package masl_test

import (
	"encoding/json"
	"testing"

	"github.com/hyphacoop/go-dasl/cid"
	"github.com/hyphacoop/go-dasl/drisl"
	"github.com/hyphacoop/go-dasl/masl"
	"github.com/stretchr/testify/require"
)

func TestMASLSingleMode(t *testing.T) {
	// craft a single masl doc
	single := `{
  "src": "bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4",
  "mediaType": "text/html",
  "content-language": "en",
  "service-worker-allowed": "/",
  "custom-header": "custom-value",
  "x-my-app-version": "1.2.3"
}`
	var singleMap map[string]any
	err := json.Unmarshal([]byte(single), &singleMap)
	require.NoError(t, err)
	src, ok := singleMap["src"].(string)
	require.True(t, ok)
	cid, err := cid.NewCidFromString(src)
	require.NoError(t, err)
	singleMap["src"] = cid
	cborBz, err := drisl.Marshal(singleMap)
	require.NoError(t, err)

	maslDoc := masl.Document{}
	err = drisl.Unmarshal(cborBz, &maslDoc)
	require.NoError(t, err, "Could not unmarshal %x", cborBz)

	require.Equal(t, maslDoc.Src, &cid)
	require.Equal(t, maslDoc.MediaType, "text/html")
	require.Equal(t, maslDoc.ContentLanguage, "en")
	require.Equal(t, maslDoc.ServiceWorkerAllowed, "/")
	require.Equal(t, maslDoc.Attributes["custom-header"], "custom-value")
	require.Equal(t, maslDoc.Attributes["x-my-app-version"], "1.2.3")
	require.Len(t, maslDoc.Attributes, 2) // two unknown attributes
	require.False(t, maslDoc.IsBundle())

	roundTripped, err := drisl.Marshal(maslDoc)
	require.NoError(t, err)
	require.Equal(t, cborBz, roundTripped)
}

func TestMASLBundleMode(t *testing.T) {
	bundle := `
{
  "name": "My Doc",
  "resources": {
    "/": {
      "mediaType": "text/html",
      "content-encoding": "gzip",
      "content-language": "fr",
      "x-cache-control": "max-age=3600"
    },
    "/interactive.js": {
      "mediaType": "application/javascript",
      "sourcemap": "/interactive.js.map",
      "x-build-version": "2.1.0"
    },
    "/interactive.js.map": {
      "mediaType": "application/json"
    },
    "/picture.jpg": {
      "mediaType": "image/jpeg",
      "x-photographer": "Jane Doe"
    }
  }
}`
	var bundleMap map[string]any
	err := json.Unmarshal([]byte(bundle), &bundleMap)
	require.NoError(t, err)
	resourcesMap, ok := bundleMap["resources"].(map[string]any)
	require.True(t, ok)
	for key, resource := range resourcesMap {
		resourceMap, ok := resource.(map[string]any)
		require.True(t, ok)
		resourceMap["src"] = cid.HashBytes([]byte(key))
	}
	cborBz, err := drisl.Marshal(bundleMap)
	require.NoError(t, err)

	maslDoc := masl.Document{}
	err = drisl.Unmarshal(cborBz, &maslDoc)
	require.NoError(t, err)
	require.Nil(t, maslDoc.Src)
	require.Equal(t, maslDoc.Name, "My Doc")
	require.Len(t, maslDoc.Resources, 4)

	// Check root resource "/"
	rootResource := maslDoc.Resources["/"]
	require.NotNil(t, rootResource)
	require.Equal(t, cid.HashBytes([]byte("/")), *rootResource.Src)
	require.Equal(t, "text/html", rootResource.MediaType)
	require.Equal(t, "gzip", rootResource.ContentEncoding)
	require.Equal(t, "fr", rootResource.ContentLanguage)
	require.Equal(t, "max-age=3600", rootResource.Attributes["x-cache-control"])
	require.Len(t, rootResource.Attributes, 1) // one unknown attribute

	// Check JavaScript resource "/interactive.js"
	jsResource := maslDoc.Resources["/interactive.js"]
	require.NotNil(t, jsResource)
	require.Equal(t, cid.HashBytes([]byte("/interactive.js")), *jsResource.Src)
	require.Equal(t, "application/javascript", jsResource.MediaType)
	require.Equal(t, "/interactive.js.map", jsResource.Sourcemap)
	require.Equal(t, "2.1.0", jsResource.Attributes["x-build-version"])
	require.Len(t, jsResource.Attributes, 1) // one unknown attribute

	// Check source map resource "/interactive.js.map"
	mapResource := maslDoc.Resources["/interactive.js.map"]
	require.NotNil(t, mapResource)
	require.Equal(t, cid.HashBytes([]byte("/interactive.js.map")), *mapResource.Src)
	require.Equal(t, "application/json", mapResource.MediaType)
	require.Len(t, mapResource.Attributes, 0) // no additional attributes

	// Check image resource "/picture.jpg"
	imgResource := maslDoc.Resources["/picture.jpg"]
	require.NotNil(t, imgResource)
	require.Equal(t, cid.HashBytes([]byte("/picture.jpg")), *imgResource.Src)
	require.Equal(t, "image/jpeg", imgResource.MediaType)
	require.Equal(t, "Jane Doe", imgResource.Attributes["x-photographer"])
	require.Len(t, imgResource.Attributes, 1) // one unknown attribute

	require.True(t, maslDoc.IsBundle())

	roundTripped, err := drisl.Marshal(maslDoc)
	require.NoError(t, err)
	require.Equal(t, cborBz, roundTripped)
}

func TestMASAppMannifest(t *testing.T) {
	// Test a MASL Document that specifies an entire appManifest
	manifestDoc := `{
  "name": "Unicorn Editor",
  "short_name": "Unicorn",
  "description": "This is simply the best app to edit unicorns with.",
  "background_color": "#00ff75",
  "theme_color": "#ff0075",
  "categories": ["productivity", "graphics"],
  "icons": [
    {
      "src": "/unicorn.svg",
      "sizes": "512x512",
      "purpose": "any"
    }
  ],
  "screenshots": [
    {
      "src": "/screenshot1.png",
      "sizes": "1280x720",
      "label": "Main editing interface",
      "form_factor": "wide",
      "platform": "windows"
    }
  ],
  "version": 1,
  "roots": [],
  "$type": "ing.dasl.masl",
  "resources": {
    "/": {
      "mediaType": "text/html",
      "content-language": "en",
      "content-security-policy": "default-src 'self'",
      "referrer-policy": "strict-origin-when-cross-origin",
      "x-content-type-options": "nosniff",
      "custom-app-header": "unicorn-v1"
    },
    "/unicorn.svg": {
      "mediaType": "image/svg+xml",
      "content-encoding": "gzip",
      "x-artist": "Magic Designer"
    },
    "/screenshot1.png": {
      "mediaType": "image/png",
      "x-resolution": "high"
    },
    "/app.js": {
      "mediaType": "application/javascript",
      "sourcemap": "/app.js.map",
      "supports-loading-mode": "blocking",
      "x-build-hash": "abc123def456"
    },
    "/app.js.map": {
      "mediaType": "application/json",
      "x-debug-info": "enabled"
    }
  }
}`

	var docMap map[string]any
	err := json.Unmarshal([]byte(manifestDoc), &docMap)
	require.NoError(t, err)

	// Fix version field type (JSON unmarshals numbers as float64)
	if version, ok := docMap["version"].(float64); ok {
		docMap["version"] = int(version)
	}

	// Add CIDs for all resources
	resourcesMap, ok := docMap["resources"].(map[string]any)
	require.True(t, ok)
	for key, resource := range resourcesMap {
		resourceMap, ok := resource.(map[string]any)
		require.True(t, ok)
		resourceMap["src"] = cid.HashBytes([]byte(key))
	}

	cborBz, err := drisl.Marshal(docMap)
	require.NoError(t, err)

	maslDoc := masl.Document{}
	err = drisl.Unmarshal(cborBz, &maslDoc)
	require.NoError(t, err)

	// Test App Manifest fields
	require.Equal(t, "Unicorn Editor", maslDoc.Name)
	require.Equal(t, "Unicorn", maslDoc.ShortName)
	require.Equal(t, "This is simply the best app to edit unicorns with.", maslDoc.Description)
	require.Equal(t, "#00ff75", maslDoc.BackgroundColor)
	require.Equal(t, "#ff0075", maslDoc.ThemeColor)
	require.Equal(t, []string{"productivity", "graphics"}, maslDoc.Categories)

	// Test Icons
	require.Len(t, maslDoc.Icons, 1)
	require.Equal(t, "/unicorn.svg", maslDoc.Icons[0].Src)
	require.Equal(t, "512x512", maslDoc.Icons[0].Sizes)
	require.Equal(t, "any", maslDoc.Icons[0].Purpose)

	// Test Screenshots
	require.Len(t, maslDoc.Screenshots, 1)
	require.Equal(t, "/screenshot1.png", maslDoc.Screenshots[0].Src)
	require.Equal(t, "1280x720", maslDoc.Screenshots[0].Sizes)
	require.Equal(t, "Main editing interface", maslDoc.Screenshots[0].Label)
	require.Equal(t, "wide", maslDoc.Screenshots[0].FormFactor)
	require.Equal(t, "windows", maslDoc.Screenshots[0].Platform)

	// Test CAR Compatibility
	require.Equal(t, 1, maslDoc.Version)
	require.Len(t, maslDoc.Roots, 0)

	// Test AT Compatibility
	require.Equal(t, "ing.dasl.masl", maslDoc.Type)

	// Test Bundle mode
	require.True(t, maslDoc.IsBundle())
	require.Len(t, maslDoc.Resources, 5)

	// Test root resource with comprehensive HTTP headers
	rootResource := maslDoc.Resources["/"]
	require.NotNil(t, rootResource)
	require.Equal(t, "text/html", rootResource.MediaType)
	require.Equal(t, "en", rootResource.ContentLanguage)
	require.Equal(t, "default-src 'self'", rootResource.ContentSecurityPolicy)
	require.Equal(t, "strict-origin-when-cross-origin", rootResource.ReferrerPolicy)
	require.Equal(t, "nosniff", rootResource.XContentTypeOptions)
	require.Equal(t, "unicorn-v1", rootResource.Attributes["custom-app-header"])
	require.Len(t, rootResource.Attributes, 1)

	// Test SVG resource
	svgResource := maslDoc.Resources["/unicorn.svg"]
	require.NotNil(t, svgResource)
	require.Equal(t, "image/svg+xml", svgResource.MediaType)
	require.Equal(t, "gzip", svgResource.ContentEncoding)
	require.Equal(t, "Magic Designer", svgResource.Attributes["x-artist"])
	require.Len(t, svgResource.Attributes, 1)

	// Test JavaScript resource with sourcemap
	jsResource := maslDoc.Resources["/app.js"]
	require.NotNil(t, jsResource)
	require.Equal(t, "application/javascript", jsResource.MediaType)
	require.Equal(t, "/app.js.map", jsResource.Sourcemap)
	require.Equal(t, "blocking", jsResource.SupportsLoadingMode)
	require.Equal(t, "abc123def456", jsResource.Attributes["x-build-hash"])
	require.Len(t, jsResource.Attributes, 1)

	// Test roundtrip
	roundTripped, err := drisl.Marshal(maslDoc)
	require.NoError(t, err)
	require.Equal(t, cborBz, roundTripped)
}

func TestMASLVersioning(t *testing.T) {
	// Test versioning with prev field
	prevCid := cid.HashBytes([]byte("previous-version"))

	versionedDoc := `{
  "name": "Versioned Document",
  "src": "bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4",
  "mediaType": "text/html"
}`

	var versionedMap map[string]any
	err := json.Unmarshal([]byte(versionedDoc), &versionedMap)
	require.NoError(t, err)

	// Add src and prev CIDs
	src, ok := versionedMap["src"].(string)
	require.True(t, ok)
	srcCid, err := cid.NewCidFromString(src)
	require.NoError(t, err)
	versionedMap["src"] = srcCid
	versionedMap["prev"] = prevCid

	cborBz, err := drisl.Marshal(versionedMap)
	require.NoError(t, err)

	maslDoc := masl.Document{}
	err = drisl.Unmarshal(cborBz, &maslDoc)
	require.NoError(t, err)

	// Test versioning fields
	require.Equal(t, "Versioned Document", maslDoc.Name)
	require.Equal(t, &srcCid, maslDoc.Src)
	require.Equal(t, &prevCid, maslDoc.Prev)
	require.Equal(t, "text/html", maslDoc.MediaType)
	require.False(t, maslDoc.IsBundle())

	// Test roundtrip
	roundTripped, err := drisl.Marshal(maslDoc)
	require.NoError(t, err)
	require.Equal(t, cborBz, roundTripped)
}
