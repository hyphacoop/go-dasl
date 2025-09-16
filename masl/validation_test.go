package masl_test

// import (
// 	"testing"

// 	"github.com/hyphacoop/go-dasl/cid"
// 	"github.com/hyphacoop/go-dasl/masl"
// )

// func TestDocumentValidation(t *testing.T) {
// 	t.Run("single-mode document should have valid CID", func(t *testing.T) {
// 		validCid := cid.HashBytes([]byte("test data"))
// 		doc, err := masl.NewSingleModeDocument(validCid)
// 		if err != nil {
// 			t.Fatalf("failed to create single-mode document: %v", err)
// 		}

// 		err = doc.Validate()
// 		if err != nil {
// 			t.Fatalf("expected valid document to pass validation, got: %v", err)
// 		}
// 	})

// 	t.Run("single-mode document with empty CID should fail validation", func(t *testing.T) {
// 		_, err := masl.NewSingleModeDocument(cid.Cid{}) // empty CID

// 		if err == nil {
// 			t.Fatal("expected NewSingleModeDocument with empty CID to fail")
// 		}
// 	})

// 	t.Run("bundle-mode document should validate all resources", func(t *testing.T) {
// 		doc := masl.NewDocument()

// 		// Add valid resource
// 		resource1 := masl.NewResource()
// 		err := resource1.SetCID(cid.HashBytes([]byte("resource1")))
// 		if err != nil {
// 			t.Fatalf("failed to set valid CID: %v", err)
// 		}
// 		doc.AddResourceToBundle("/valid", resource1)

// 		// Add invalid resource - this should fail at SetCID
// 		resource2 := masl.NewResource()
// 		err = resource2.SetCID(cid.Cid{}) // empty CID
// 		if err == nil {
// 			t.Fatal("expected SetCID with empty CID to fail")
// 		}
// 	})

// 	t.Run("empty document should fail validation", func(t *testing.T) {
// 		doc := masl.NewDocument()

// 		err := doc.Validate()
// 		if err == nil {
// 			t.Fatal("expected empty document to fail validation")
// 		}
// 	})
// }

// func TestResourceValidation(t *testing.T) {
// 	t.Run("resource with valid CID should pass validation", func(t *testing.T) {
// 		resource := masl.NewResource()
// 		err := resource.SetCID(cid.HashBytes([]byte("test data")))
// 		if err != nil {
// 			t.Fatalf("failed to set valid CID: %v", err)
// 		}

// 		err = resource.Validate()
// 		if err != nil {
// 			t.Fatalf("expected valid resource to pass validation, got: %v", err)
// 		}
// 	})

// 	t.Run("resource with empty CID should fail validation", func(t *testing.T) {
// 		resource := masl.NewResource()
// 		err := resource.SetCID(cid.Cid{}) // empty CID

// 		if err == nil {
// 			t.Fatal("expected SetCID with empty CID to fail")
// 		}
// 	})

// 	t.Run("resource without CID should fail validation", func(t *testing.T) {
// 		resource := masl.NewResource()
// 		// Don't set any CID

// 		err := resource.Validate()
// 		if err == nil {
// 			t.Fatal("expected resource without CID to fail validation")
// 		}
// 	})
// }

// func TestDocumentTypeValidation(t *testing.T) {
// 	t.Run("GetResource should fail on bundle-mode document", func(t *testing.T) {
// 		doc := masl.NewDocument()

// 		// Make it a bundle-mode document
// 		resource := masl.NewResource()
// 		err := resource.SetCID(cid.HashBytes([]byte("test")))
// 		if err != nil {
// 			t.Fatalf("failed to set valid CID: %v", err)
// 		}
// 		doc.AddResourceToBundle("/test", resource)

// 		// This should panic, not return an error
// 		defer func() {
// 			if r := recover(); r == nil {
// 				t.Fatal("expected GetResource to panic on bundle-mode document")
// 			}
// 		}()
// 		doc.GetResource()
// 	})

// 	t.Run("GetResourceFromBundle should fail on single-mode document", func(t *testing.T) {
// 		validCid := cid.HashBytes([]byte("test"))
// 		doc, err := masl.NewSingleModeDocument(validCid)
// 		if err != nil {
// 			t.Fatalf("failed to create single-mode document: %v", err)
// 		}

// 		// This should panic, not return an error
// 		defer func() {
// 			if r := recover(); r == nil {
// 				t.Fatal("expected GetResourceFromBundle to panic on single-mode document")
// 			}
// 		}()
// 		doc.GetResourceFromBundle("/test")
// 	})

// 	t.Run("GetResource should work on single-mode document", func(t *testing.T) {
// 		validCid := cid.HashBytes([]byte("test"))
// 		doc, err := masl.NewSingleModeDocument(validCid)
// 		if err != nil {
// 			t.Fatalf("failed to create single-mode document: %v", err)
// 		}

// 		resource := doc.GetResource()
// 		if resource == nil {
// 			t.Fatal("expected GetResource to return a resource")
// 		}

// 		if !resource.GetCID().Defined() {
// 			t.Fatal("expected resource to have valid CID")
// 		}
// 	})

// 	t.Run("GetResourceFromBundle should work on bundle-mode document", func(t *testing.T) {
// 		doc := masl.NewDocument()

// 		resource := masl.NewResource()
// 		err := resource.SetCID(cid.HashBytes([]byte("test")))
// 		if err != nil {
// 			t.Fatalf("failed to set valid CID: %v", err)
// 		}
// 		doc.AddResourceToBundle("/test", resource)

// 		retrieved := doc.GetResourceFromBundle("/test")
// 		if retrieved == nil {
// 			t.Fatal("expected GetResourceFromBundle to return a resource")
// 		}

// 		if !retrieved.GetCID().Defined() {
// 			t.Fatal("expected retrieved resource to have valid CID")
// 		}
// 	})
// }
