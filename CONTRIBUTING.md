# Contributing

## Development rules

- Clone the repo with `git clone --recurse-submodules` to include the dasl-testing submodule
- Support the latest two Go versions
- Add new tests for new code
- Don't add dependencies

## Testing

Run tests with `go test ./...`.

You can run fuzz tests by name, for example:

```
go test -fuzz='^FuzzMarshal$' -run='^FuzzMarshal$' -fuzztime=5m ./drisl
go test -fuzz='^FuzzUnmarshal$' -run='^FuzzUnmarshal$' -fuzztime=5m ./drisl
```

Benchmarking is available with `go test -bench=. ./drisl`.

The CI will check all of these for your PR.

## Licensing

Submitting a PR implies you release your changes under the dual MIT and Apache 2.0 license
of this repo. It does not relinquish authorship, and no CLA is required.
