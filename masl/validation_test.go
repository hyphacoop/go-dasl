package masl_test

import (
	"testing"

	"github.com/hyphacoop/go-dasl/cid"
	"github.com/hyphacoop/go-dasl/masl"
)

func TestDocumentValidation(t *testing.T) {
	t.Run("single-mode document should have valid CID", func(t *testing.T) {
		validCid := cid.HashBytes([]byte("test data"))
		doc := &masl.Document{
			Resource: masl.Resource{
				Src: validCid,
			},
		}

		if !doc.Valid() {
			t.Fatal("expected valid document to pass validation")
		}

		if doc.IsBundle() {
			t.Fatal("expected single-mode document, got bundle")
		}
	})

	t.Run("single-mode document with empty CID should still be valid", func(t *testing.T) {
		doc := &masl.Document{
			Resource: masl.Resource{
				Src: cid.Cid{}, // empty CID
			},
		}

		// Current API doesn't validate CID emptiness, so this should pass
		if !doc.Valid() {
			t.Fatal("expected document to be valid")
		}

		if doc.IsBundle() {
			t.Fatal("expected single-mode document, got bundle")
		}
	})

	t.Run("bundle-mode document should have resources", func(t *testing.T) {
		doc := &masl.Document{
			Resources: map[string]*masl.Resource{
				"/valid": {
					Src: cid.HashBytes([]byte("resource1")),
				},
				"/another": {
					Src: cid.HashBytes([]byte("resource2")),
				},
			},
		}

		if !doc.Valid() {
			t.Fatal("expected valid bundle document to pass validation")
		}

		if !doc.IsBundle() {
			t.Fatal("expected bundle-mode document, got single-mode")
		}

		if len(doc.Resources) != 2 {
			t.Fatalf("expected 2 resources, got %d", len(doc.Resources))
		}
	})

	t.Run("empty document should still be valid", func(t *testing.T) {
		doc := &masl.Document{}

		// Current API always returns true for Valid()
		if !doc.Valid() {
			t.Fatal("expected empty document to be valid")
		}

		if doc.IsBundle() {
			t.Fatal("expected single-mode document, got bundle")
		}
	})
}

func TestResourceValidation(t *testing.T) {
	t.Run("resource with valid CID should be accessible", func(t *testing.T) {
		resource := &masl.Resource{
			Src: cid.HashBytes([]byte("test data")),
		}

		if !resource.Src.Defined() {
			t.Fatal("expected resource to have defined CID")
		}
	})

	t.Run("resource with empty CID should have undefined CID", func(t *testing.T) {
		resource := &masl.Resource{
			Src: cid.Cid{}, // empty CID
		}

		if resource.Src.Defined() {
			t.Fatal("expected resource to have undefined CID")
		}
	})

	t.Run("resource without CID should have undefined CID", func(t *testing.T) {
		resource := &masl.Resource{}
		// Don't set any CID

		if resource.Src.Defined() {
			t.Fatal("expected resource to have undefined CID")
		}
	})
}

func TestDocumentTypeValidation(t *testing.T) {
	t.Run("bundle-mode document should be identified correctly", func(t *testing.T) {
		doc := &masl.Document{
			Resources: map[string]*masl.Resource{
				"/test": {
					Src: cid.HashBytes([]byte("test")),
				},
			},
		}

		if !doc.IsBundle() {
			t.Fatal("expected bundle-mode document")
		}

		// Access resource directly from Resources map
		resource := doc.Resources["/test"]
		if resource == nil {
			t.Fatal("expected to find resource at /test")
		}

		if !resource.Src.Defined() {
			t.Fatal("expected resource to have valid CID")
		}
	})

	t.Run("single-mode document should be identified correctly", func(t *testing.T) {
		validCid := cid.HashBytes([]byte("test"))
		doc := &masl.Document{
			Resource: masl.Resource{
				Src: validCid,
			},
		}

		if doc.IsBundle() {
			t.Fatal("expected single-mode document")
		}

		// Access resource directly from embedded Resource
		if !doc.Resource.Src.Defined() {
			t.Fatal("expected document resource to have valid CID")
		}

		if !doc.Resource.Src.Equal(validCid) {
			t.Fatal("expected document resource CID to match")
		}
	})

	t.Run("document with both Resource and Resources should be bundle-mode", func(t *testing.T) {
		doc := &masl.Document{
			Resource: masl.Resource{
				Src: cid.HashBytes([]byte("single")),
			},
			Resources: map[string]*masl.Resource{
				"/test": {
					Src: cid.HashBytes([]byte("bundle")),
				},
			},
		}

		// IsBundle() returns true if Resources is not nil
		if !doc.IsBundle() {
			t.Fatal("expected bundle-mode document when Resources is present")
		}
	})
}

// TestInvalidDocuments tests documents that should fail validation according to MASL spec
func TestInvalidDocuments(t *testing.T) {
	t.Run("bundle with resource path not starting with slash should be invalid", func(t *testing.T) {
		doc := &masl.Document{
			Resources: map[string]*masl.Resource{
				"invalid-path": { // MUST start with /
					Src: cid.HashBytes([]byte("test")),
				},
			},
		}

		if doc.Valid() {
			t.Fatal("expected document with invalid resource path to be invalid")
		}
	})

	t.Run("bundle with empty resource path should be invalid", func(t *testing.T) {
		doc := &masl.Document{
			Resources: map[string]*masl.Resource{
				"": { // Empty path is invalid
					Src: cid.HashBytes([]byte("test")),
				},
			},
		}

		if doc.Valid() {
			t.Fatal("expected document with empty resource path to be invalid")
		}
	})

	t.Run("bundle with resource missing src should be invalid", func(t *testing.T) {
		doc := &masl.Document{
			Resources: map[string]*masl.Resource{
				"/valid": {
					// Missing Src field - MUST have src
				},
			},
		}

		if doc.Valid() {
			t.Fatal("expected document with resource missing src to be invalid")
		}
	})

	t.Run("CAR compatibility with invalid version should be invalid", func(t *testing.T) {
		doc := &masl.Document{
			Resource: masl.Resource{
				Src: cid.HashBytes([]byte("test")),
			},
			Version: 2, // MUST be 1
		}

		if doc.Valid() {
			t.Fatal("expected document with invalid CAR version to be invalid")
		}
	})

	t.Run("AT compatibility with invalid type should be invalid", func(t *testing.T) {
		doc := &masl.Document{
			Resource: masl.Resource{
				Src: cid.HashBytes([]byte("test")),
			},
			Type: "invalid.type", // SHOULD be "ing.dasl.masl"
		}

		if doc.Valid() {
			t.Fatal("expected document with invalid AT type to be invalid")
		}
	})

	t.Run("app manifest icon with non-existent src should be invalid", func(t *testing.T) {
		doc := &masl.Document{
			Resources: map[string]*masl.Resource{
				"/": {
					Src: cid.HashBytes([]byte("index")),
				},
			},
		}
		// AppManifest is embedded in Document via Resource
		doc.Icons = []masl.Icon{
			{
				Src: "/non-existent.png", // MUST reference existing resource
			},
		}

		if doc.Valid() {
			t.Fatal("expected document with icon referencing non-existent resource to be invalid")
		}
	})

	t.Run("app manifest screenshot with non-existent src should be invalid", func(t *testing.T) {
		doc := &masl.Document{
			Resources: map[string]*masl.Resource{
				"/": {
					Src: cid.HashBytes([]byte("index")),
				},
			},
		}
		// AppManifest is embedded in Document via Resource
		doc.Screenshots = []masl.Screenshot{
			{
				Src: "/missing-screenshot.png", // MUST reference existing resource
			},
		}

		if doc.Valid() {
			t.Fatal("expected document with screenshot referencing non-existent resource to be invalid")
		}
	})

	t.Run("sourcemap referencing non-existent resource should be invalid", func(t *testing.T) {
		doc := &masl.Document{
			Resources: map[string]*masl.Resource{
				"/app.js": {
					Src:       cid.HashBytes([]byte("javascript")),
					Sourcemap: "/app.js.map", // MUST reference existing resource
				},
			},
		}

		if doc.Valid() {
			t.Fatal("expected document with sourcemap referencing non-existent resource to be invalid")
		}
	})

	t.Run("speculation-rules referencing non-existent resource should be invalid", func(t *testing.T) {
		doc := &masl.Document{
			Resources: map[string]*masl.Resource{
				"/": {
					Src:              cid.HashBytes([]byte("index")),
					SpeculationRules: "/speculation.json", // MUST reference existing resource
				},
			},
		}

		if doc.Valid() {
			t.Fatal("expected document with speculation-rules referencing non-existent resource to be invalid")
		}
	})

	t.Run("single-mode document with icon src should be invalid", func(t *testing.T) {
		doc := &masl.Document{
			Resource: masl.Resource{
				Src: cid.HashBytes([]byte("test")),
			},
		}
		// AppManifest is embedded in Document via Resource
		doc.Icons = []masl.Icon{
			{
				Src: "/icon.png", // Invalid in single mode - no resources map
			},
		}

		if doc.Valid() {
			t.Fatal("expected single-mode document with icon src to be invalid")
		}
	})

	t.Run("single-mode document with screenshot src should be invalid", func(t *testing.T) {
		doc := &masl.Document{
			Resource: masl.Resource{
				Src: cid.HashBytes([]byte("test")),
			},
		}
		// AppManifest is embedded in Document via Resource
		doc.Screenshots = []masl.Screenshot{
			{
				Src: "/screenshot.png", // Invalid in single mode - no resources map
			},
		}

		if doc.Valid() {
			t.Fatal("expected single-mode document with screenshot src to be invalid")
		}
	})
}

// TestValidDocuments tests documents that should pass validation
func TestValidDocuments(t *testing.T) {
	t.Run("valid bundle with proper paths and resources", func(t *testing.T) {
		doc := &masl.Document{
			Resources: map[string]*masl.Resource{
				"/": {
					Src: cid.HashBytes([]byte("index")),
				},
				"/style.css": {
					Src: cid.HashBytes([]byte("styles")),
				},
				"/app.js": {
					Src:       cid.HashBytes([]byte("javascript")),
					Sourcemap: "/app.js.map",
				},
				"/app.js.map": {
					Src: cid.HashBytes([]byte("sourcemap")),
				},
			},
		}

		// Set app manifest fields on the document (embedded via Resource)
		doc.Name = "Test App"
		doc.Icons = []masl.Icon{
			{
				Src: "/icon.png",
			},
		}

		// Add the icon resource
		doc.Resources["/icon.png"] = &masl.Resource{
			Src: cid.HashBytes([]byte("icon")),
		}

		if !doc.Valid() {
			t.Fatal("expected valid bundle document to pass validation")
		}
	})

	t.Run("valid CAR compatible document", func(t *testing.T) {
		doc := &masl.Document{
			Resource: masl.Resource{
				Src: cid.HashBytes([]byte("test")),
			},
			Version: 1, // Correct version
			Roots: []cid.Cid{
				cid.HashBytes([]byte("root1")),
				cid.HashBytes([]byte("root2")),
			},
		}

		if !doc.Valid() {
			t.Fatal("expected valid CAR compatible document to pass validation")
		}
	})

	t.Run("valid AT compatible document", func(t *testing.T) {
		doc := &masl.Document{
			Resource: masl.Resource{
				Src: cid.HashBytes([]byte("test")),
			},
			Type: "ing.dasl.masl", // Correct type
		}

		if !doc.Valid() {
			t.Fatal("expected valid AT compatible document to pass validation")
		}
	})

	t.Run("valid document with versioning", func(t *testing.T) {
		doc := &masl.Document{
			Resource: masl.Resource{
				Src: cid.HashBytes([]byte("test")),
			},
			Prev: cid.HashBytes([]byte("previous-version")),
		}

		if !doc.Valid() {
			t.Fatal("expected valid document with versioning to pass validation")
		}
	})
}
