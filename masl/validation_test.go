package masl_test

import (
	"testing"

	"github.com/hyphacoop/go-dasl/cid"
	"github.com/hyphacoop/go-dasl/masl"
)

func TestDocumentValidation(t *testing.T) {
	t.Run("single-mode document should have valid CID", func(t *testing.T) {
		doc := masl.NewDocument()
		validCid := cid.HashBytes([]byte("test data"))
		doc.SetSrc(validCid)

		err := doc.Validate()
		if err != nil {
			t.Fatalf("expected valid document to pass validation, got: %v", err)
		}
	})

	t.Run("single-mode document with empty CID should fail validation", func(t *testing.T) {
		doc := masl.NewDocument()
		doc.SetSrc(cid.Cid{}) // empty CID

		err := doc.Validate()
		if err == nil {
			t.Fatal("expected document with empty CID to fail validation")
		}
	})

	t.Run("bundle-mode document should validate all resources", func(t *testing.T) {
		doc := masl.NewDocument()

		// Add valid resource
		resource1 := masl.NewResource()
		resource1.SetCID(cid.HashBytes([]byte("resource1")))
		doc.AddResourceToBundle(resource1, "/valid")

		// Add invalid resource
		resource2 := masl.NewResource()
		resource2.SetCID(cid.Cid{}) // empty CID
		doc.AddResourceToBundle(resource2, "/invalid")

		err := doc.Validate()
		if err == nil {
			t.Fatal("expected document with invalid resource to fail validation")
		}
	})

	t.Run("empty document should pass validation", func(t *testing.T) {
		doc := masl.NewDocument()

		err := doc.Validate()
		if err != nil {
			t.Fatalf("expected empty document to pass validation, got: %v", err)
		}
	})
}

func TestResourceValidation(t *testing.T) {
	t.Run("resource with valid CID should pass validation", func(t *testing.T) {
		resource := masl.NewResource()
		resource.SetCID(cid.HashBytes([]byte("test data")))

		err := resource.Validate()
		if err != nil {
			t.Fatalf("expected valid resource to pass validation, got: %v", err)
		}
	})

	t.Run("resource with empty CID should fail validation", func(t *testing.T) {
		resource := masl.NewResource()
		resource.SetCID(cid.Cid{}) // empty CID

		err := resource.Validate()
		if err == nil {
			t.Fatal("expected resource with empty CID to fail validation")
		}
	})

	t.Run("resource without CID should fail validation", func(t *testing.T) {
		resource := masl.NewResource()
		// Don't set any CID

		err := resource.Validate()
		if err == nil {
			t.Fatal("expected resource without CID to fail validation")
		}
	})
}

func TestDocumentTypeValidation(t *testing.T) {
	t.Run("GetResource should fail on bundle-mode document", func(t *testing.T) {
		doc := masl.NewDocument()

		// Make it a bundle-mode document
		resource := masl.NewResource()
		resource.SetCID(cid.HashBytes([]byte("test")))
		doc.AddResourceToBundle(resource, "/test")

		_, err := doc.GetResource()
		if err == nil {
			t.Fatal("expected GetResource to fail on bundle-mode document")
		}
	})

	t.Run("GetResourceFromBundle should fail on single-mode document", func(t *testing.T) {
		doc := masl.NewDocument()

		// Make it a single-mode document
		doc.SetSrc(cid.HashBytes([]byte("test")))

		_, err := doc.GetResourceFromBundle("/test")
		if err == nil {
			t.Fatal("expected GetResourceFromBundle to fail on single-mode document")
		}
	})

	t.Run("GetResource should work on single-mode document", func(t *testing.T) {
		doc := masl.NewDocument()
		doc.SetSrc(cid.HashBytes([]byte("test")))

		resource, err := doc.GetResource()
		if err != nil {
			t.Fatalf("expected GetResource to work on single-mode document, got: %v", err)
		}

		if !resource.GetCID().Defined() {
			t.Fatal("expected resource to have valid CID")
		}
	})

	t.Run("GetResourceFromBundle should work on bundle-mode document", func(t *testing.T) {
		doc := masl.NewDocument()

		resource := masl.NewResource()
		resource.SetCID(cid.HashBytes([]byte("test")))
		doc.AddResourceToBundle(resource, "/test")

		retrieved, err := doc.GetResourceFromBundle("/test")
		if err != nil {
			t.Fatalf("expected GetResourceFromBundle to work on bundle-mode document, got: %v", err)
		}

		if !retrieved.GetCID().Defined() {
			t.Fatal("expected retrieved resource to have valid CID")
		}
	})
}
