package masl_test

import (
	"reflect"
	"testing"

	"github.com/hyphacoop/go-dasl/cid"
	"github.com/hyphacoop/go-dasl/drisl"
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
		data, err := drisl.Marshal(doc)
		if err != nil {
			t.Fatalf("failed to marshal document: %v", err)
		}

		// Unmarshal the data back into a document
		unmarshaled := masl.Document{Resources: make(map[string]*masl.Resource)}
		err = drisl.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Fatalf("failed to unmarshal document: %v", err)
		}

		if !reflect.DeepEqual(doc, unmarshaled) {
			t.Fatalf("documents not equal after roundtrip:\noriginal: %+v\nunmarshaled: %+v", doc, unmarshaled)
		}
	})
}

// maslDocumentGenerator creates a generator for MASL documents
func maslDocumentGenerator() *rapid.Generator[masl.Document] {
	return rapid.Custom(func(t *rapid.T) masl.Document {
		// Generate either a single-mode or bundle-mode document
		isBundleMode := rapid.Bool().Draw(t, "is_bundle_mode")

		doc := masl.Document{Resources: make(map[string]*masl.Resource), Resource: masl.Resource{Attributes: make(map[string]any)}}

		if isBundleMode {
			// Bundle mode: generate resources map (ensure at least one resource)
			resources := rapid.MapOfN(
				pathGenerator(),
				resourceGenerator(),
				1, 5, // At least 1, at most 5 resources
			).Draw(t, "resources")

			// Set the resources on the document
			for path, resource := range resources {
				doc.Resources[path] = &resource
			}
		} else {
			doc.Resource = resourceGenerator().Draw(t, "single_resource")
		}

		return doc
	})
}

// resourceGenerator creates a generator for MASL resources
func resourceGenerator() *rapid.Generator[masl.Resource] {
	return rapid.Custom(func(t *rapid.T) masl.Resource {
		resource := masl.Resource{Attributes: make(map[string]any)}

		// Generate a CID for the resource
		resourceCid := cidGenerator().Draw(t, "cid")
		resource.Src = resourceCid

		// Generate arbitrary attributes
		numAttributes := rapid.IntRange(0, 5).Draw(t, "num_attributes")
		for i := 0; i < numAttributes; i++ {
			attrName := rapid.String().Draw(t, "attr_name")
			attrValue := rapid.String().Draw(t, "attr_value")
			resource.Attributes[attrName] = attrValue
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
