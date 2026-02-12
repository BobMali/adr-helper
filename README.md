# ADR Helper

A CLI tool and web UI for managing [Architecture Decision Records](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions). Inspired by [adr-tools](https://github.com/npryce/adr-tools).

## Installation

```bash
make build       # builds bin/adr and bin/adr-web (includes frontend)
make build-cli   # builds bin/adr only
make build-web   # builds bin/adr-web only (includes frontend build)
```

### All Makefile Targets

| Target | Description |
|--------|-------------|
| `make build` | Build both CLI and web binaries |
| `make build-cli` | Build CLI binary (`bin/adr`) |
| `make build-web` | Build web binary with embedded frontend (`bin/adr-web`) |
| `make build-frontend` | Build the Vue frontend only |
| `make test` | Run all Go tests |
| `make test-verbose` | Run Go tests with verbose output |
| `make test-cover` | Generate HTML coverage report (`coverage.html`) |
| `make test-frontend` | Run frontend tests (Vitest) |
| `make test-all` | Run Go and frontend tests |
| `make vet` | Run `go vet` |
| `make clean` | Remove build artifacts and coverage files |

## Quick Start

```bash
adr init docs/decisions              # initialize ADR directory
adr new "Use PostgreSQL"             # create a new ADR
adr show 1                           # display ADR in terminal
adr update 1 accepted                # update ADR status
adr list                             # list all ADRs in a table
```

## Commands

### `adr init [path]`

Initialize a new ADR directory with a template. Defaults to the current directory.

| Flag | Description |
|------|-------------|
| `-t, --template <name>` | Template format: `nygard` (default), `madr-minimal`, `madr-full` |
| `--template-file <file>` | Output template filename (default: `template.md`) |
| `-f, --force` | Overwrite existing files |

### `adr new <title>`

Create a new ADR.

| Flag | Description |
|------|-------------|
| `-s, --supersedes <id>[,<id>...]` | IDs of ADRs that the new record supersedes |

```bash
adr new "Migrate to PostgreSQL" --supersedes 3,5
```

### `adr show <id>`

Display an ADR in the terminal with syntax highlighting.

| Flag | Description |
|------|-------------|
| `--plain` | Disable colored output |
| `--json` | Output as JSON |

### `adr update <id> [status]`

Update the status of an existing ADR. When status is omitted, an interactive menu is shown.

Valid statuses: `proposed`, `accepted`, `rejected`, `deprecated`, `superseded`.

### `adr list`

List all ADRs in a table with columns: ID, Date, Title, Status.

| Flag | Description |
|------|-------------|
| `--plain` | Disable colored output |
| `--json` | Output as JSON array |
| `-s, --search <query>` | Filter ADRs by title or number |
| `--count` | Show status counts instead of listing ADRs |

```bash
adr list                          # list all ADRs
adr list --search "database"      # filter by title or number
adr list --count                  # show counts grouped by status
adr list --count --json           # status counts as JSON
```

### Environment Variables

`adr show` and `adr list` respect the `NO_COLOR` environment variable. When set (to any value), colored output is disabled — equivalent to passing `--plain`.

## Web UI

```bash
adr-web             # starts on :8080
adr-web --addr :3000
```

The web server embeds a Vue 3 single-page application that provides:

- **List view** — browse all ADRs with real-time search filtering
- **Detail view** — view full ADR content with markdown rendering
- **Status updates** — change an ADR's status directly from the UI
- **Supersede flow** — mark an ADR as superseded and link the superseding record
- **Dark mode** — automatic dark/light theme support

### API Endpoints

The Go server exposes the following REST API:

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check (`{"status":"ok"}`) |
| `GET` | `/api/adr` | List all ADRs (supports `?q=<query>` for search) |
| `GET` | `/api/adr/statuses` | List valid status values |
| `GET` | `/api/adr/{number}` | Get a single ADR with full content |
| `PATCH` | `/api/adr/{number}/status` | Update an ADR's status |

The `PATCH` endpoint accepts a JSON body:

```json
{ "status": "accepted" }
```

For superseding, include the superseding ADR number:

```json
{ "status": "superseded", "supersededBy": 4 }
```

## Development

### Frontend Dev Server

```bash
cd web && npm install
cd web && npm run dev    # Vite dev server on :5173
```

The Vite dev server proxies `/api` and `/health` requests to the Go backend at `localhost:8080`. Run `adr-web` in a separate terminal for the full stack.

### Running Tests

```bash
make test                # Go tests
make test-frontend       # Frontend tests (Vitest)
make test-all            # Both
```

## Templates

Three built-in templates are available (set via `adr init --template`):

- **nygard** (default) — Context, Decision, Consequences
- **madr-minimal** — Context and Problem Statement, Considered Options, Decision Outcome
- **madr-full** — Full MADR format with YAML frontmatter and extended sections

## Configuration

`adr init` creates an `.adr.json` file in the project root:

```json
{
  "version": "1",
  "directory": "docs/decisions",
  "template": "nygard",
  "templateFile": "template.md"
}
```
