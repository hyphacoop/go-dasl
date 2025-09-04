package masl_test

import (
	"testing"

	"github.com/hyphacoop/go-dasl/cid"
	"github.com/hyphacoop/go-dasl/masl"
	"pgregory.net/rapid"
)

// TestMASLRoundtrip tests that MASL documents can be marshaled and unmarshaled
// while preserving their content using property-based testing
func TestMASLRoundtrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a random MASL document
		doc := maslDocumentGenerator().Draw(t, "document")

		// Marshal the document
		data, err := doc.Marshal()
		if err != nil {
			t.Fatalf("failed to marshal document: %v", err)
		}

		// Unmarshal the data back into a document
		var unmarshaled masl.Document
		err = unmarshaled.Unmarshal(data)
		if err != nil {
			t.Fatalf("failed to unmarshal document: %v", err)
		}

		// The roundtrip should preserve the document content
		// This will fail until we implement proper marshaling/unmarshaling
		if !documentsEqual(doc, unmarshaled) {
			t.Fatalf("roundtrip failed: original != unmarshaled")
		}
	})
}

// maslDocumentGenerator creates a generator for MASL documents
func maslDocumentGenerator() *rapid.Generator[masl.Document] {
	return rapid.Custom(func(t *rapid.T) masl.Document {
		// Generate either a single-mode or bundle-mode document
		isBundleMode := rapid.Bool().Draw(t, "is_bundle_mode")

		doc := masl.NewDocument()

		if isBundleMode {
			// Bundle mode: generate resources map
			resources := rapid.MapOf(
				pathGenerator(),
				resourceGenerator(),
			).Draw(t, "resources")

			// Set the resources on the document
			for path, resource := range resources {
				err := doc.AddResourceToBundle(resource, path)
				if err != nil {
					t.Fatalf("failed to add resource to bundle: %v", err)
				}
			}
		} else {
			// Single mode: generate a single resource with CID
			resourceCid := cidGenerator().Draw(t, "src_cid")
			doc.SetSrc(resourceCid)
		}

		// Add optional metadata fields
		if rapid.Bool().Draw(t, "has_name") {
			name := rapid.String().Draw(t, "name")
			doc.SetAttribute("name", name)
		}

		if rapid.Bool().Draw(t, "has_description") {
			description := rapid.String().Draw(t, "description")
			doc.SetAttribute("description", description)
		}

		// Add CAR compatibility fields if needed
		if rapid.Bool().Draw(t, "has_car_fields") {
			doc.SetAttribute("version", 1)
			// Only create non-empty roots arrays, or skip roots entirely
			if rapid.Bool().Draw(t, "has_roots") {
				roots := rapid.SliceOfN(cidGenerator(), 1, 5).Draw(t, "roots")
				doc.SetAttribute("roots", roots)
			}
		}

		// Add arbitrary attributes
		numAttributes := rapid.IntRange(0, 3).Draw(t, "num_doc_attributes")
		for i := 0; i < numAttributes; i++ {
			attrName := rapid.String().Draw(t, "doc_attr_name")
			attrValue := rapid.String().Draw(t, "doc_attr_value")
			doc.SetAttribute(attrName, attrValue)
		}

		return doc
	})
}

// resourceGenerator creates a generator for MASL resources
func resourceGenerator() *rapid.Generator[masl.Resource] {
	return rapid.Custom(func(t *rapid.T) masl.Resource {
		resource := masl.NewResource()

		// Generate a CID for the resource
		resourceCid := cidGenerator().Draw(t, "cid")
		resource.SetCID(resourceCid)

		// Generate arbitrary attributes
		numAttributes := rapid.IntRange(0, 5).Draw(t, "num_attributes")
		for i := 0; i < numAttributes; i++ {
			attrName := rapid.String().Draw(t, "attr_name")
			attrValue := rapid.String().Draw(t, "attr_value")
			resource.SetAttribute(attrName, attrValue)
		}

		return resource
	})
}

// pathGenerator creates a generator for arbitrary resource paths
func pathGenerator() *rapid.Generator[string] {
	return rapid.String()
}

// cidGenerator creates a generator for CIDs
func cidGenerator() *rapid.Generator[cid.Cid] {
	return rapid.Custom(func(t *rapid.T) cid.Cid {
		// Generate random bytes for CID content, ensuring we have enough for a valid CID
		data := rapid.SliceOfN(rapid.Byte(), 32, 32).Draw(t, "cid_data")

		// Create a CID by hashing the data (this always produces a valid CID)
		return cid.HashBytes(data)
	})
}

// documentsEqual compares two MASL documents for equality
func documentsEqual(a, b masl.Document) bool {
	return a.Equal(b)
}
