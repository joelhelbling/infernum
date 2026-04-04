# ollama-bench (CLI)

Go CLI for running Ollama benchmarks, publishing results to the ollama-bench web service, and querying comparison reports. Open source, distributed as signed binaries.

**Spec:** `docs/specs/platform-design.md`
**Current plan:** `docs/plans/2026-04-04-ollama-bench-cli.md`

## Tech stack

- Go 1.22+
- Cobra (CLI framework)
- tablewriter (ASCII tables)
- Standard library for HTTP, crypto

## Conventions

- **`pkg/`** — shared, importable by external projects (notably the web backend). Types only, minimal surface. Breaking changes here break downstream consumers.
- **`internal/`** — CLI-private logic, not importable
- **`cmd/ollama-bench/main.go`** — entry point, wires version flag

## Testing

- `go test -short ./...` — unit tests (no network, no ollama, no DB)
- `go test ./...` — includes integration tests (require real ollama running locally)
- Tests use `FakeOllamaRunner` to avoid depending on real ollama for unit testing

## Build

```bash
make build    # builds ollama-bench binary with version injection
make test     # unit tests only
```

## Trust story

The CLI is open for review, closed for contribution. Users must be able to clone, review, and build themselves. Signed binary distribution (Sigstore/cosign) is required at release time.
