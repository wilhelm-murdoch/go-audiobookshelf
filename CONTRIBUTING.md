# Contributing

Thanks for your interest in improving go-audiobookshelf! This is a thin,
idiomatic Go client for the Audiobookshelf API, and contributions that
keep it that way are very welcome.

## Getting started

You'll need **Go 1.26+**. Clone the repository and, from its root:

```sh
make test    # unit tests
make lint    # golangci-lint (pinned via go run, no global install needed)
make vet     # go vet
make cover   # unit tests with coverage
make fmt     # gofmt the tree
```

All of these run in CI, so it's worth getting them green locally first.
`make lint` downloads and runs the exact pinned linter version on first
use — no separate install step.

## How the code is organized

- One file per API resource group (`libraries.go`, `items.go`, …),
  mirroring the official API docs.
- `internal/rest` is a small, API-agnostic HTTP toolkit (transport, path
  building, errors, auth) that the typed client wraps. Prefer adding
  transport-level behavior there and Audiobookshelf-specific behavior in
  the root package.
- The full conventions — the resource-handle pattern, the path builder,
  flexible JSON decoding, the `Millis`/`Seconds` types — are documented in
  [`AGENTS.md`](AGENTS.md). Please skim it before a non-trivial change.

## Adding or changing endpoints

1. Build paths with the path builder (`apiPath(...)` / `rawPath(...)`),
   never by hand-concatenating strings.
2. Return resources with their client back-reference set so handle methods
   keep working.
3. Add a table entry in `endpoints_test.go` (and `handles_test.go` for a
   new handle method) asserting the method, path, and query.
4. If the server's JSON is loosely typed (numbers-as-strings, bools where
   a string is expected, etc.), add a tolerant decode path plus a unit
   test, following the existing examples in `types.go`.

## Tests

- Unit tests use `net/http/httptest` — no live server, no mocking.
- Functional tests live in `integration_test.go`, gated behind the
  `integration` build tag and the `ABS_BASE_URL` environment variable:

  ```sh
  ABS_BASE_URL=http://localhost:13378 make integration
  ```

  They run against a real Audiobookshelf server. The suite covers the
  media-free slice of the API; media-dependent coverage (items, covers,
  play sessions, scans) is a known gap and a great place to help.

When you touch the public API, please update the runnable examples in
`example_test.go` so the documentation stays compile-checked.

## Server compatibility

Each release is verified against a specific Audiobookshelf version
(`TestedServerVersion`, mirrored in the README matrix and both CI
pipelines). If you test against a newer server, mention the version in
your PR — and if it required decode changes, bump those three places
together.

## Pull requests

- Keep changes focused; separate unrelated work into separate PRs.
- Write clear, imperative commit messages explaining the *why*.
- Make sure `make test`, `make lint`, and `make vet` pass.
- New behavior needs tests.

By contributing, you agree that your work is licensed under the project's
[MIT License](LICENSE).
