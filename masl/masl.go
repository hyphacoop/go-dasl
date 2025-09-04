package masl

import (
	"fmt"
	"reflect"

	"github.com/hyphacoop/go-dasl/cid"
	"github.com/hyphacoop/go-dasl/drisl"
)

// Document represents a MASL document with arbitrary metadata
type Document struct {
	data map[string]any // Raw deserialized data
}

// Resource represents a MASL resource with arbitrary attributes
type Resource struct {
	data map[string]any // Raw deserialized data
}

// NewDocument creates a new empty MASL document
func NewDocument() Document {
	return Document{
		data: make(map[string]any),
	}
}

// NewResource creates a new empty MASL resource
func NewResource() Resource {
	return Resource{
		data: make(map[string]any),
	}
}

// Marshal serializes the document to DRISL bytes
func (d Document) Marshal() ([]byte, error) {
	// Validate the document before marshaling
	if err := d.validate(); err != nil {
		return nil, fmt.Errorf("invalid document: %w", err)
	}
	return drisl.Marshal(d.data)
}

// Unmarshal deserializes DRISL bytes into the document
func (d *Document) Unmarshal(data []byte) error {
	d.data = make(map[string]any)
	if err := drisl.Unmarshal(data, &d.data); err != nil {
		return err
	}
	// Validate the document after unmarshaling
	if err := d.validate(); err != nil {
		return fmt.Errorf("invalid document after unmarshal: %w", err)
	}
	return nil
}

// GetResourceFromBundle retrieves a resource from the bundle by path
func (d Document) GetResourceFromBundle(name string) (*Resource, error) {
	// Check if this is a single-mode document (has src field)
	if _, hasSrc := d.data["src"]; hasSrc {
		return nil, fmt.Errorf("cannot call GetResourceFromBundle on single-mode document (use GetResource instead)")
	}

	resources, ok := d.data["resources"]
	if !ok {
		return nil, fmt.Errorf("document has no resources bundle")
	}

	resourcesMap, ok := resources.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("resources field is not a map")
	}

	resourceData, ok := resourcesMap[name]
	if !ok {
		return nil, fmt.Errorf("resource %q not found", name)
	}

	resourceMap, ok := resourceData.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("resource data is not a map")
	}

	return &Resource{data: resourceMap}, nil
}

// GetResource retrieves the single resource (for single-mode documents)
func (d Document) GetResource() (*Resource, error) {
	// Check if this is a bundle-mode document (has resources field)
	if _, hasResources := d.data["resources"]; hasResources {
		return nil, fmt.Errorf("cannot call GetResource on bundle-mode document (use GetResourceFromBundle instead)")
	}

	_, ok := d.data["src"]
	if !ok {
		return nil, fmt.Errorf("document has no src field")
	}

	return &Resource{data: d.data}, nil
}

// AddResourceToBundle adds a resource to the document's bundle
func (d *Document) AddResourceToBundle(resource Resource, name string) error {
	if d.data == nil {
		d.data = make(map[string]any)
	}

	resources, ok := d.data["resources"]
	if !ok {
		resources = make(map[string]any)
		d.data["resources"] = resources
	}

	resourcesMap, ok := resources.(map[string]any)
	if !ok {
		return fmt.Errorf("resources field is not a map")
	}

	resourcesMap[name] = resource.data
	return nil
}

// SetSrc sets the src field for single-mode documents
func (d *Document) SetSrc(c cid.Cid) {
	if d.data == nil {
		d.data = make(map[string]any)
	}
	d.data["src"] = c
}

// SetAttribute sets an arbitrary attribute on the document
func (d *Document) SetAttribute(name string, value any) {
	if d.data == nil {
		d.data = make(map[string]any)
	}
	d.data[name] = value
}

// GetAttribute gets an arbitrary attribute from the document
func (d Document) GetAttribute(name string) (any, bool) {
	if d.data == nil {
		return nil, false
	}
	value, ok := d.data[name]
	return value, ok
}

// GetCID retrieves the CID from the resource's src field
func (r Resource) GetCID() cid.Cid {
	src, ok := r.data["src"]
	if !ok {
		return cid.Cid{}
	}

	// Handle CID stored as a link object
	if linkMap, ok := src.(map[string]any); ok {
		if cidValue, ok := linkMap["/"]; ok {
			if cidStr, ok := cidValue.(string); ok {
				c, err := cid.NewCidFromString(cidStr)
				if err != nil {
					return cid.Cid{}
				}
				return c
			}
		}
	}

	// Handle CID stored directly
	if c, ok := src.(cid.Cid); ok {
		return c
	}

	return cid.Cid{}
}

// SetCID sets the CID in the resource's src field
func (r *Resource) SetCID(c cid.Cid) {
	if r.data == nil {
		r.data = make(map[string]any)
	}
	r.data["src"] = c
}

// GetAttribute retrieves an arbitrary attribute from the resource
func (r Resource) GetAttribute(name string) (string, error) {
	if r.data == nil {
		return "", fmt.Errorf("resource has no data")
	}

	value, ok := r.data[name]
	if !ok {
		return "", fmt.Errorf("attribute %q not found", name)
	}

	// Try to convert to string
	if str, ok := value.(string); ok {
		return str, nil
	}

	return fmt.Sprintf("%v", value), nil
}

// SetAttribute sets an arbitrary attribute on the resource
func (r *Resource) SetAttribute(name string, value any) {
	if r.data == nil {
		r.data = make(map[string]any)
	}
	r.data[name] = value
}

// Equal compares two documents for equality
func (d Document) Equal(other Document) bool {
	return deepEqualWithTypeConversion(d.data, other.data)
}

// Equal compares two resources for equality
func (r Resource) Equal(other Resource) bool {
	return deepEqualWithTypeConversion(r.data, other.data)
}

// deepEqualWithTypeConversion compares two values with special handling for type conversions
// that can occur during CBOR marshaling/unmarshaling
func deepEqualWithTypeConversion(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	va, vb := reflect.ValueOf(a), reflect.ValueOf(b)

	// Handle maps recursively
	if va.Kind() == reflect.Map && vb.Kind() == reflect.Map {
		if va.Len() != vb.Len() {
			return false
		}
		for _, key := range va.MapKeys() {
			aVal := va.MapIndex(key)
			bVal := vb.MapIndex(key)
			if !bVal.IsValid() {
				return false
			}
			if !deepEqualWithTypeConversion(aVal.Interface(), bVal.Interface()) {
				return false
			}
		}
		return true
	}

	// Handle slices recursively with type conversion
	if va.Kind() == reflect.Slice && vb.Kind() == reflect.Slice {
		if va.Len() != vb.Len() {
			return false
		}

		// Special case: empty slices are equal regardless of element type
		if va.Len() == 0 && vb.Len() == 0 {
			return true
		}

		for i := 0; i < va.Len(); i++ {
			if !deepEqualWithTypeConversion(va.Index(i).Interface(), vb.Index(i).Interface()) {
				return false
			}
		}
		return true
	}

	// Handle CID comparisons
	if aCid, aOk := a.(cid.Cid); aOk {
		if bCid, bOk := b.(cid.Cid); bOk {
			return aCid.Equals(bCid)
		}
		return false
	}

	// Handle numeric type conversions (CBOR can change int to uint64, etc.)
	if isNumeric(va) && isNumeric(vb) {
		return numericEqual(va, vb)
	}

	// For all other types use standard reflection comparison
	return reflect.DeepEqual(a, b)
}

// isNumeric checks if a value is a numeric type
func isNumeric(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// numericEqual compares two numeric values for equality
func numericEqual(a, b reflect.Value) bool {
	// Convert both to float64 for comparison
	var aFloat, bFloat float64

	switch a.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		aFloat = float64(a.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		aFloat = float64(a.Uint())
	case reflect.Float32, reflect.Float64:
		aFloat = a.Float()
	}

	switch b.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		bFloat = float64(b.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		bFloat = float64(b.Uint())
	case reflect.Float32, reflect.Float64:
		bFloat = b.Float()
	}

	return aFloat == bFloat
}

// validate checks if the document is valid (internal method)
func (d Document) validate() error {
	return validateValue(d.data)
}

// Validate checks if the document is valid and properly structured
func (d Document) Validate() error {
	// Check if it's a single-mode document
	if src, hasSrc := d.data["src"]; hasSrc {
		// Validate the src CID
		if c, ok := src.(cid.Cid); ok {
			if !c.Defined() {
				return fmt.Errorf("document src contains empty CID")
			}
		} else {
			return fmt.Errorf("document src is not a valid CID")
		}
	}

	// Check if it's a bundle-mode document
	if resources, hasResources := d.data["resources"]; hasResources {
		resourcesMap, ok := resources.(map[string]any)
		if !ok {
			return fmt.Errorf("resources field is not a map")
		}

		// Validate each resource in the bundle
		for path, resourceData := range resourcesMap {
			resourceMap, ok := resourceData.(map[string]any)
			if !ok {
				return fmt.Errorf("resource at path %q is not a map", path)
			}

			resource := &Resource{data: resourceMap}
			if err := resource.Validate(); err != nil {
				return fmt.Errorf("resource at path %q is invalid: %w", path, err)
			}
		}
	}

	// Validate any other values in the document
	return validateValue(d.data)
}

// validate checks if the resource is valid (internal method)
func (r Resource) validate() error {
	return validateValue(r.data)
}

// Validate checks if the resource is valid and has a proper CID
func (r Resource) Validate() error {
	// A resource must have a src field with a valid CID
	src, ok := r.data["src"]
	if !ok {
		return fmt.Errorf("resource missing required src field")
	}

	// Check if it's a valid CID
	if c, ok := src.(cid.Cid); ok {
		if !c.Defined() {
			return fmt.Errorf("resource src contains empty CID")
		}
	} else {
		return fmt.Errorf("resource src is not a valid CID")
	}

	// Validate any other values in the resource
	return validateValue(r.data)
}

// validateValue recursively validates a value for empty CIDs
func validateValue(value any) error {
	switch v := value.(type) {
	case cid.Cid:
		if !v.Defined() {
			return fmt.Errorf("empty CID found")
		}
	case map[string]any:
		for key, val := range v {
			if err := validateValue(val); err != nil {
				return fmt.Errorf("in field %q: %w", key, err)
			}
		}
	case []any:
		for i, val := range v {
			if err := validateValue(val); err != nil {
				return fmt.Errorf("in array index %d: %w", i, err)
			}
		}
	case []cid.Cid:
		for i, c := range v {
			if !c.Defined() {
				return fmt.Errorf("empty CID found in array index %d", i)
			}
		}
	}
	return nil
}
