# go-dasl

A fast, ergonomic Go library for working with content-addressed data using [DASL](https://dasl.ing) specs. DASL is a streamlined distillation of IPFS specifications including CID and DAG-CBOR, making it easier to work with content-addressed data structures.

## Why go-dasl?

- **ATProto/Bluesky Ready**: Built specifically for [ATProtocol](https://atproto.com/) development, making it ideal for Bluesky and similar platforms
- **Performant**: Faster and more efficient than existing alternatives
- **Simple API**: Clean, intuitive interface designed for real-world use cases
- **Well-Typed**: Full Go type safety with struct tag support for fine-grained control

## Quick Start

Install the library:

```bash
go get github.com/hyphacoop/go-dasl@latest
```

Common use cases:

| Task | Module |
|------|--------|
| Decode CBOR data from the firehose | `drisl` |
| Create CBOR records | `drisl` |
| Parse and verify CIDs | `cid` |

## Project Status

✅ **Active and stable** - The library works well and is ready for use today.

⚠️ **Breaking changes possible** - The API may change before v1.0.0. Production users should pin versions or wait for the v1.0.0 release when the DRISL and CID specs are finalized.

## Example Usage

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

📚 **Full documentation**: See [pkg.go.dev/github.com/hyphacoop/go-dasl](https://pkg.go.dev/github.com/hyphacoop/go-dasl) for complete API reference.

## Struct Tag Options

Struct tags provide fine-grained control over encoding and decoding behavior.

### Size Optimization

Reduce encoded size and improve performance:

| Tag | Description |
|-----|-------------|
| `toarray` | Encode without field names (positional encoding) |
| `omitempty` | Omit empty fields (same rules as `encoding/json`) |
| `omitzero` | Omit zero-value fields (same rules as `encoding/json`) |

**Note**: When using `toarray`, the encoder ignores `omitempty` and `omitzero` to maintain consistent array element positions for decoding.

**Example:**

```go
// Positional encoding (smaller size)
type myData struct {
    _       struct{} `cbor:",toarray"`
    Payload []byte
    Age     int
    Name    string
}

// Field-level optimization
type myData2 struct {
    Payload []byte
    Age     int     `cbor:",omitempty"`
    Name    string  `cbor:"my_name"`
    Ref     cid.Cid `cbor:",omitzero"` // Use omitzero (not omitempty) for CIDs
}
```

### Field Control

| Tag | Description |
|-----|-------------|
| `unknown` | Collect unrecognized fields into a map (forward compatibility) |
| `-` | Skip field entirely during encoding and decoding |

**Example:**

```go
type myData3 struct {
    ID      string         `cbor:"id"`
    Value   int            `cbor:"value"`
    Secret  []byte         `cbor:"-"`        // Never encoded/decoded
    Unknown map[string]any `cbor:",unknown"` // Captures unknown fields
}
```

## Supported DASL Specs

| Spec | Status | Description |
|------|--------|-------------|
| **DRISL** (dag-cbor) | ✅ Implemented | CBOR encoding for content-addressed data |
| **CID** (including BDASL) | ✅ Implemented | Content Identifiers |
| **RASL** | ✅ Implemented | Record Addressing System Layer |
| **MASL** | ✅ Implemented | Merkle Array System Layer |
| **CAR** | ⭕ Out of scope | Content Addressable aRchives |

## Versioning

This library follows [Semantic Versioning](https://semver.org/):

- **Pre-release versions** (e.g., `x.y.z-alpha`): Not intended for public use
- **Zero versions** (`0.y.z`): Ready to use, but may include breaking changes
- **v1.0.0**: Will be released once DRISL and CID specs are finalized

Experimental submodules may have breaking changes in minor versions.

## Contributing

We welcome bug reports and fixes! Feature requests are currently out of scope but will be considered in the future.

**Want to contribute?** File an issue or start a discussion — let's talk!

See [CONTRIBUTING.md](./CONTRIBUTING.md) for details.

## Funding

This library is developed and maintained by [Hypha](https://hypha.coop/), funded by [IPFS](https://ipfs.tech) via the [Open Impact Foundation](https://openimpact.foundation/).

## License

Dual-licensed under **MIT** or **Apache 2.0** — your choice.
