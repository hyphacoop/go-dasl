module internal/upstream_bench

go 1.24.6

require (
	github.com/fxamacker/cbor/v2 v2.9.0
	github.com/hyphacoop/go-dasl v0.2.1
)

require (
	github.com/hyphacoop/cbor/v2 v2.0.0-20250827195905-4b6b4a1b5aef // indirect
	github.com/multiformats/go-varint v0.1.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
)

replace github.com/hyphacoop/go-dasl => ../..
