module github.com/hyphacoop/go-dasl

go 1.24.5

require (
	github.com/fxamacker/cbor/v2 v2.9.0
	github.com/multiformats/go-varint v0.1.0
	pgregory.net/rapid v1.2.0
)

require github.com/x448/float16 v0.8.4 // indirect

replace github.com/fxamacker/cbor/v2 => github.com/hyphacoop/cbor/v2 v2.0.0-20250820175754-6587bd559f7d
