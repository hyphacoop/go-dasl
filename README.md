# go-dasl

A Go reference library for [DASL](https://dasl.ing).

🚧 Not ready for production yet 🚧

See an overview at [pkg.go.dev](https://pkg.go.dev/github.com/hyphacoop/go-dasl).
In-depth documentation is coming.

## Submodules

DASL has many specs, only some of which are implemented here.

- DRISL (dag-cbor): implemented
- CID: working, independent implementation is planned (including BDASL)
- RASL: stretch goal
- MASL: stretch goal
- CAR: currently out of scope

## Versioning

This library follows [Semantic Versioning](https://semver.org/).

- Pre-release versions (like x.y.z-alpha) are not intended to be used by the public
- Zero versions (0.y.z) are ready to use, but may have breaking changes
- 1.0.0 will be released once the DRISL spec has been finalized

## Funding

Development and maintenance of this library is funded by [IPFS](https://ipfs.tech)
via the [Open Impact Foundation](https://openimpact.foundation/).
Work is performed by [Hypha](https://hypha.coop/).

## Contributing

At this stage bug reports and fixes are welcome, but feature requests are out of scope.
Feature issues and PRs will be considered in the future.
File an issue or discussion and let's talk!

## License

This library is dual-licensed under MIT or Apache 2.0.
