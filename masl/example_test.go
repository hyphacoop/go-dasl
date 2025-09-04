package masl_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/hyphacoop/go-dasl/cid"
	"github.com/hyphacoop/go-dasl/drisl"
	"github.com/hyphacoop/go-dasl/masl"
)

func ExampleDocument_single_mode() {
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
	json.Unmarshal([]byte(single), &singleMap)
	src := singleMap["src"].(string)
	cid := cid.MustNewCidFromString(src)
	singleMap["src"] = cid
	cborBz, err := drisl.Marshal(singleMap)
	if err != nil {
		panic(err)
	}

	maslDoc := masl.Document{}
	err = drisl.Unmarshal(cborBz, &maslDoc)
	if err != nil {
		panic(err)
	}

	if !maslDoc.IsBundle() {
		// We can access properties of a non-bundle document through the document object...
		fmt.Println(maslDoc.Src)
		fmt.Println(maslDoc.MediaType)
		// ...or as a resource
		resource := maslDoc.Resource
		fmt.Println(resource.ContentLanguage)
		fmt.Println(resource.ServiceWorkerAllowed)
		fmt.Println(resource.Attributes["custom-header"])
		fmt.Println(resource.Attributes["x-my-app-version"])
	} else {
		panic("Single-mode document marked bundle!")
	}
	roundTripped, err := drisl.Marshal(maslDoc)
	if err != nil {
		panic(err)
	}
	fmt.Println(hex.EncodeToString(roundTripped))
	// The roundTripped bytes should be the same as the original:
	if !bytes.Equal(roundTripped, cborBz) {
		panic(fmt.Sprintf("RoundTripped and original bytes differ! Round tripp result: %s", hex.EncodeToString(roundTripped)))
	}
	// Output: bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4
	// text/html
	// en
	// /
	// custom-value
	// 1.2.3
	// a663737263d82a58250001551220adee2e8fb5459c9bcf07d7d78d1183bf40a7f60f57a54a19194801c9a27ead87696d656469615479706569746578742f68746d6c6d637573746f6d2d6865616465726c637573746f6d2d76616c756570636f6e74656e742d6c616e677561676562656e70782d6d792d6170702d76657273696f6e65312e322e3376736572766963652d776f726b65722d616c6c6f776564612f
}

func ExampleDocument_bundle_mode() {
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
	if err != nil {
		panic(err)
	}
	resourcesMap := bundleMap["resources"].(map[string]any)
	for key, resource := range resourcesMap {
		resourceMap := resource.(map[string]any)
		resourceMap["src"] = cid.HashBytes([]byte(key))
	}
	cborBz, err := drisl.Marshal(bundleMap)
	if err != nil {
		panic(err)
	}

	maslDoc := masl.Document{}
	err = drisl.Unmarshal(cborBz, &maslDoc)
	if err != nil {
		panic(err)
	}

	// Bundles (by convention) have no Src defined.
	fmt.Println(maslDoc.Src.Defined())
	fmt.Println(maslDoc.Name)
	fmt.Println(len(maslDoc.Resources))

	// Bundle documents return IsBundle() = true
	if !maslDoc.IsBundle() {
		panic("Bundle doc reported as single mode!")
	}

	// You can get resources for the bundle doc through the Resources map:
	rootResource := maslDoc.Resources["/"]
	fmt.Println(rootResource.Src)
	// Resources make their attributes available:
	fmt.Println(rootResource.MediaType)
	fmt.Println(rootResource.ContentEncoding)
	fmt.Println(rootResource.ContentLanguage)
	fmt.Println(rootResource.Attributes["x-cache-control"])
	fmt.Println(len(rootResource.Attributes))

	jsResource := maslDoc.Resources["/interactive.js"]
	fmt.Println(jsResource.Src)
	fmt.Println(jsResource.MediaType)
	fmt.Println(jsResource.Sourcemap)
	fmt.Println(jsResource.Attributes["x-build-version"])
	fmt.Println(len(jsResource.Attributes))

	mapResource := maslDoc.Resources["/interactive.js.map"]
	fmt.Println(mapResource.Src)
	fmt.Println(mapResource.MediaType)
	fmt.Println(len(mapResource.Attributes))

	imgResource := maslDoc.Resources["/picture.jpg"]
	fmt.Println(imgResource.Src)
	fmt.Println(imgResource.MediaType)
	fmt.Println(imgResource.Attributes["x-photographer"])
	fmt.Println(len(imgResource.Attributes))

	roundTripped, err := drisl.Marshal(maslDoc)
	if err != nil {
		panic(err)
	}
	fmt.Println(hex.EncodeToString(roundTripped))
	// The roundTripped bytes should be the same as the original:
	if !bytes.Equal(roundTripped, cborBz) {
		panic(fmt.Sprintf("RoundTripped and original bytes differ! Round tripp result: %s", hex.EncodeToString(roundTripped)))
	}
	// Output: false
	// My Doc
	// 4
	// bafkreiekl3nlfatderbsdhqfdzfn4li5lo6gohdycbi36fbxrf6l37va6e
	// text/html
	// gzip
	// fr
	// max-age=3600
	// 1
	// bafkreifx36hhksvlghenzm6belsnjulbjrpanca6ro7kuscmga6b33whra
	// application/javascript
	// /interactive.js.map
	// 2.1.0
	// 1
	// bafkreif552xcbuuhm43vegovh5ezenqfc3tx4rzibmb46pqx5etzrrbe6e
	// application/json
	// 0
	// bafkreieneqao2af5fqo5hy326culolykokz5cmeiuprdy6irfz3h5ignti
	// image/jpeg
	// Jane Doe
	// 1
	// a2646e616d65664d7920446f63697265736f7572636573a4612fa563737263d82a582500015512208a5edab282632443219e051e4ade2d1d5bbc671c781051bf1437897cbdfea0f1696d656469615479706569746578742f68746d6c6f782d63616368652d636f6e74726f6c6c6d61782d6167653d3336303070636f6e74656e742d656e636f64696e6764677a697070636f6e74656e742d6c616e67756167656266726c2f706963747572652e6a7067a363737263d82a582500015512208d2400ed00bd2c1dd3e37af0a8b72f0a72b3d13088a3e23c79112e767ea0cd9a696d65646961547970656a696d6167652f6a7065676e782d70686f746f67726170686572684a616e6520446f656f2f696e7465726163746976652e6a73a463737263d82a58250001551220b7df8e754aab31c8dcb3c122e4d4d1614c5e06881e8bbeaa484c303c1deec788696d6564696154797065766170706c69636174696f6e2f6a61766173637269707469736f757263656d6170732f696e7465726163746976652e6a732e6d61706f782d6275696c642d76657273696f6e65322e312e30732f696e7465726163746976652e6a732e6d6170a263737263d82a58250001551220bdeeae20d28767375219d53f4992360516e77e47280b03cf3e17e92798c424f1696d6564696154797065706170706c69636174696f6e2f6a736f6e
}

func ExampleDocument_app_manifest() {
	// A MASL Document can specify an entire app manifest
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
	if err != nil {
		panic(err)
	}

	// Hack: fix version field type (JSON unmarshals numbers as float64)
	if version, ok := docMap["version"].(float64); ok {
		docMap["version"] = int(version)
	}

	// Add CIDs for all resources
	resourcesMap := docMap["resources"].(map[string]any)
	for key, resource := range resourcesMap {
		resourceMap := resource.(map[string]any)
		resourceMap["src"] = cid.HashBytes([]byte(key))
	}

	cborBz, err := drisl.Marshal(docMap)
	if err != nil {
		panic(err)
	}

	maslDoc := masl.Document{}
	err = drisl.Unmarshal(cborBz, &maslDoc)
	if err != nil {
		panic(err)
	}

	// App Manifest fields can be specified at the root level:
	fmt.Println(maslDoc.Name)
	fmt.Println(maslDoc.ShortName)
	fmt.Println(maslDoc.Description)
	fmt.Println(maslDoc.BackgroundColor)
	fmt.Println(maslDoc.ThemeColor)
	fmt.Println(maslDoc.Categories)

	// Icons:
	fmt.Println(len(maslDoc.Icons))
	fmt.Println(maslDoc.Icons[0].Src)
	fmt.Println(maslDoc.Icons[0].Sizes)
	fmt.Println(maslDoc.Icons[0].Purpose)

	// Screnshots:
	fmt.Println(len(maslDoc.Screenshots))
	fmt.Println(maslDoc.Screenshots[0].Src)
	fmt.Println(maslDoc.Screenshots[0].Sizes)
	fmt.Println(maslDoc.Screenshots[0].Label)
	fmt.Println(maslDoc.Screenshots[0].FormFactor)
	fmt.Println(maslDoc.Screenshots[0].Platform)

	// CAR Compatibility fields
	fmt.Println(maslDoc.Version)

	// AT Compatibility fields
	fmt.Println(maslDoc.Type)

	// App manifests are bundles...
	fmt.Println(maslDoc.IsBundle())
	fmt.Println(len(maslDoc.Resources))

	// ...with resources that are available through Resources:
	rootResource := maslDoc.Resources["/"]
	fmt.Println(rootResource.MediaType)
	fmt.Println(rootResource.ContentLanguage)
	fmt.Println(rootResource.ContentSecurityPolicy)
	fmt.Println(rootResource.ReferrerPolicy)
	fmt.Println(rootResource.XContentTypeOptions)
	fmt.Println(rootResource.Attributes["custom-app-header"])
	fmt.Println(len(rootResource.Attributes))

	// SVG resource
	svgResource := maslDoc.Resources["/unicorn.svg"]
	fmt.Println(svgResource.MediaType)
	fmt.Println(svgResource.ContentEncoding)
	fmt.Println(svgResource.Attributes["x-artist"])
	fmt.Println(len(svgResource.Attributes))

	// JavaScript resource with sourcemap
	jsResource := maslDoc.Resources["/app.js"]
	fmt.Println(jsResource.MediaType)
	fmt.Println(jsResource.Sourcemap)
	fmt.Println(jsResource.SupportsLoadingMode)
	fmt.Println(jsResource.Attributes["x-build-hash"])
	fmt.Println(len(jsResource.Attributes))

	roundTripped, err := drisl.Marshal(maslDoc)
	if err != nil {
		panic(err)
	}
	fmt.Println(hex.EncodeToString(roundTripped))

	// The roundTripped bytes should be the same as the original:
	if !bytes.Equal(roundTripped, cborBz) {
		panic(fmt.Sprintf("RoundTripped and original bytes differ! Round tripp result: %s", hex.EncodeToString(roundTripped)))
	}
	// Output: Unicorn Editor
	// Unicorn
	// This is simply the best app to edit unicorns with.
	// #00ff75
	// #ff0075
	// [productivity graphics]
	// 1
	// /unicorn.svg
	// 512x512
	// any
	// 1
	// /screenshot1.png
	// 1280x720
	// Main editing interface
	// wide
	// windows
	// 1
	// ing.dasl.masl
	// true
	// 5
	// text/html
	// en
	// default-src 'self'
	// strict-origin-when-cross-origin
	// nosniff
	// unicorn-v1
	// 1
	// image/svg+xml
	// gzip
	// Magic Designer
	// 1
	// application/javascript
	// /app.js.map
	// blocking
	// abc123def456
	// 1
	// ab646e616d656e556e69636f726e20456469746f726524747970656d696e672e6461736c2e6d61736c6569636f6e7381a3637372636c2f756e69636f726e2e7376676573697a6573673531327835313267707572706f736563616e796776657273696f6e01697265736f7572636573a5612fa763737263d82a582500015512208a5edab282632443219e051e4ade2d1d5bbc671c781051bf1437897cbdfea0f1696d656469615479706569746578742f68746d6c6f72656665727265722d706f6c696379781f7374726963742d6f726967696e2d7768656e2d63726f73732d6f726967696e70636f6e74656e742d6c616e677561676562656e71637573746f6d2d6170702d6865616465726a756e69636f726e2d763176782d636f6e74656e742d747970652d6f7074696f6e73676e6f736e69666677636f6e74656e742d73656375726974792d706f6c6963797264656661756c742d737263202773656c6627672f6170702e6a73a563737263d82a58250001551220d5e8a067e9f2ac7b37dd1c869137e3f0036c7c71699b9be0b0327ab7d3dad8b3696d6564696154797065766170706c69636174696f6e2f6a61766173637269707469736f757263656d61706b2f6170702e6a732e6d61706c782d6275696c642d686173686c61626331323364656634353675737570706f7274732d6c6f6164696e672d6d6f646568626c6f636b696e676b2f6170702e6a732e6d6170a363737263d82a5825000155122055c751187fe3404d61e6adaf4a7a8806e35c86635a0a14d9833963acba27fb7f696d6564696154797065706170706c69636174696f6e2f6a736f6e6c782d64656275672d696e666f67656e61626c65646c2f756e69636f726e2e737667a463737263d82a582500015512200489b30c08af4b87530eeeea4c7014173a41c2798910c45a0b68556cb387f14e68782d6172746973746e4d616769632044657369676e6572696d65646961547970656d696d6167652f7376672b786d6c70636f6e74656e742d656e636f64696e6764677a6970702f73637265656e73686f74312e706e67a363737263d82a582500015512209f0ba0c2b2f1a809c356356295960ef37c7dbc1b07567e9dedecedbc45f1f825696d656469615479706569696d6167652f706e676c782d7265736f6c7574696f6e64686967686a63617465676f72696573826c70726f6475637469766974796867726170686963736a73686f72745f6e616d6567556e69636f726e6b6465736372697074696f6e7832546869732069732073696d706c792074686520626573742061707020746f206564697420756e69636f726e7320776974682e6b73637265656e73686f747381a563737263702f73637265656e73686f74312e706e67656c6162656c764d61696e2065646974696e6720696e746572666163656573697a657368313238307837323068706c6174666f726d6777696e646f77736b666f726d5f666163746f7264776964656b7468656d655f636f6c6f726723666630303735706261636b67726f756e645f636f6c6f726723303066663735
}

func ExampleDocument_versioning() {
	// Documents can be versioned via the prev field
	prevCid := cid.HashBytes([]byte("previous-version"))

	versionedDoc := `{
  "name": "Versioned Document",
  "src": "bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4",
  "mediaType": "text/html"
}`

	var versionedMap map[string]any
	err := json.Unmarshal([]byte(versionedDoc), &versionedMap)
	if err != nil {
		panic(err)
	}

	// Add src and prev CIDs
	src := versionedMap["src"].(string)
	srcCid, err := cid.NewCidFromString(src)
	if err != nil {
		panic(err)
	}
	versionedMap["src"] = srcCid
	versionedMap["prev"] = prevCid

	cborBz, err := drisl.Marshal(versionedMap)
	if err != nil {
		panic(err)
	}

	maslDoc := masl.Document{}
	err = drisl.Unmarshal(cborBz, &maslDoc)
	if err != nil {
		panic(err)
	}

	// Test versioning fields
	fmt.Println(maslDoc.Name)
	fmt.Println(maslDoc.Src)
	fmt.Println(maslDoc.Prev)
	fmt.Println(maslDoc.MediaType)
	fmt.Println(maslDoc.IsBundle())

	// Test roundtrip
	roundTripped, err := drisl.Marshal(maslDoc)
	if err != nil {
		panic(err)
	}

	// Validate roundtrip equality
	if hex.EncodeToString(cborBz) != hex.EncodeToString(roundTripped) {
		panic("roundtrip failed: original and roundtripped CBOR do not match")
	}
	fmt.Println(hex.EncodeToString(roundTripped))
	// Output: Versioned Document
	// bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4
	// bafkreidc6vzmk3fybiftl5rx6kbkz54fdpn5ahtvvleoxafmbanbxo4axa
	// text/html
	// false
	// a463737263d82a58250001551220adee2e8fb5459c9bcf07d7d78d1183bf40a7f60f57a54a19194801c9a27ead87646e616d657256657273696f6e656420446f63756d656e746470726576d82a5825000155122062f572c56cb80a0b35f637f282acf7851bdbd01e75aac8eb80ac081a1bbb80b8696d656469615479706569746578742f68746d6c
}
