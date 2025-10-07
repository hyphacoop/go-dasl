# go-dasl

A Go reference library for [DASL](https://dasl.ing).

## Project Status (Sep 2025)

This project is active and works well. Breaking changes will still occur, but you can use it today.
Production use cases are recommended to wait until the API settles, by v1 at the latest.

## Usage

```
go get github.com/hyphacoop/go-dasl@latest
```

```go
package main

import (
	"fmt"
	"time"

	"github.com/hyphacoop/go-dasl/cid"
	"github.com/hyphacoop/go-dasl/drisl"
)

type Data struct {
    Name      string    `cbor:"name"`
    Count     int       `cbor:"count"`
    Timestamp time.Time `cbor:"timestamp"`
    ID        cid.Cid   `cbor:"id"`
    Ref       cid.Cid   `cbor:"ref"`
}

func main() {
    // Create a CID for some data
    id, _ := drisl.CidForValue(map[string]string{"hello": "world"})
    ref, _ := cid.NewCidFromString("bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4")

    data := Data{
        Name:      "example",
        Count:     42,
        Timestamp: time.Date(2023, 6, 15, 14, 30, 45, 0, time.UTC),
        ID:        id,
        Ref:       ref,
    }

    bytes, err := drisl.Marshal(data)
    if err != nil {
        panic(err)
    }

    fmt.Printf("%x\n", bytes)
}
```

See an overview at [pkg.go.dev](https://pkg.go.dev/github.com/hyphacoop/go-dasl).

### Struct Tag Options

Struct tags provide control over encoding and decoding behavior.

#### Reducing Encoded Size

These options automatically reduce encoded size of structs and improve speed:
- `toarray`: encode without field names (decode back to original struct)
- `omitempty`: omit empty field when encoding (same rules as encoding/json)
- `omitzero`: omit zero-value field when encoding (same rules as encoding/json)

NOTE: When a struct uses `toarray`, the encoder will ignore `omitempty` and `omitzero` to prevent position of encoded array elements from changing. This allows decoder to match encoded elements to their Go struct field.

Example usage:

```go
// Convert struct fields to CBOR array to reduce size
type myData struct {
    _       struct{} `cbor:",toarray"`
    Payload []byte
    Age     int
    Name    string
}

// Use omitempty, omitzero, and field naming
type myData2 struct {
    Payload []byte
    Age     int     `cbor:",omitempty"`
    Name    string  `cbor:"my_name"`
    Ref     cid.Cid `cbor:",omitzero"` // Use omitzero (not omitempty) for CIDs
}
```

#### Field Control

- `unknown`: collect unrecognized fields during decoding into this map field
- `-`: omit the field entirely from encoding and decoding

Example usage:

```go
// Collect unknown fields for forward compatibility
type myData3 struct {
    ID      string         `cbor:"id"`
    Value   int            `cbor:"value"`
    Secret  []byte         `cbor:"-"`        // Skip this field
    Unknown map[string]any `cbor:",unknown"` // Captures unrecognized fields
}
```

## Submodules

DASL has many specs, only some of which are implemented here.

- DRISL (dag-cbor): implemented
- CID: implemented (including BDASL)
- RASL: implemented
- MASL: implemented
- CAR: currently out of scope

## Versioning

This library follows [Semantic Versioning](https://semver.org/).

- Pre-release versions (like x.y.z-alpha) are not intended to be used by the public
- Zero versions (0.y.z) are ready to use, but may have breaking changes
- 1.0.0 will be released once the DRISL and CID specs have been finalized

Some submodules may be documented as experimental, which would allow for breaking changes
even in minor versions.

## Funding

Development and maintenance of this library is funded by [IPFS](https://ipfs.tech)
via the [Open Impact Foundation](https://openimpact.foundation/).
Work is performed by [Hypha](https://hypha.coop/).

## Contributing

At this stage bug reports and fixes are welcome, but feature requests are out of scope.
Feature issues and PRs will be considered in the future.
File an issue or discussion and let's talk!

See also [CONTRIBUTING.md](./CONTRIBUTING.md).

## License

This library is dual-licensed under MIT or Apache 2.0.
