# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
make build          # Build both binaries to bin/
make build-cli      # Build CLI only: bin/adr
make build-web      # Build web server only: bin/adr-web
make test           # Run all Go tests
make test-verbose   # Run tests with -v
make test-cover     # Generate coverage report (coverage.html)
make vet            # Run go vet

# Run a single test
go test -run TestInitCmd_WritesConfigFile ./internal/cli/

# Frontend (from web/ directory)
cd web && npm run dev    # Vite dev server (proxies /api and /health to localhost:8080)
cd web && npm run build  # Production build (vue-tsc + vite build)
```

## Architecture

Two binaries, one domain package:

- **`cmd/adr-cli/`** — CLI binary using Cobra via `cli.NewRootCmd()`
- **`cmd/adr-web/`** — Web server binary using `flag` package, `--addr` flag (default `:8080`)
- **`internal/adr/`** — Domain layer: `ADR` struct, `Status` enum, `Repository` interface (with `context.Context`), `Config` (.adr.json read/write), and embedded ADR templates (`templates/*.md` via `//go:embed`)
- **`internal/cli/`** — Cobra command factories (`NewRootCmd()`, `NewInitCmd()`)
- **`internal/web/`** — Chi router with `Server` struct exposing `Handler()` for httptest and `ListenAndServe()` for production
- **`web/`** — Vue 3 + Vite + TypeScript frontend; Vite proxies `/api` and `/health` to the Go backend in dev

## Development Patterns

- **TDD**: RED → GREEN → REFACTOR. Write failing test first, make it pass, then clean up.
- **Cobra commands**: Always use factory functions (`NewXxxCmd()`) returning `*cobra.Command` for testability. Test via `root.SetArgs([]string{...})` + `root.Execute()`.
- **Web server tests**: Use `httptest.NewRecorder()` against `server.Handler()` — no real TCP listener needed.
- **Repository interface**: Defined in the domain package (`internal/adr/`), consumed there. Implementations live elsewhere. Use compile-time checks: `var _ adr.Repository = (*myImpl)(nil)`.
- **Config**: `.adr.json` in project root stores version, ADR directory path, and template name. Managed by `adr.SaveConfig()`/`adr.LoadConfig()`.
- **Templates**: Embedded at compile time from `internal/adr/templates/`. Three variants: `nygard`, `madr-minimal`, `madr-full`.