# ADR Helper

A CLI tool for managing [Architecture Decision Records](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions). Inspired by [adr-tools](https://github.com/npryce/adr-tools).

## Installation

```bash
make build       # builds bin/adr and bin/adr-web
make build-cli   # builds bin/adr only
make build-web   # builds bin/adr-web only
```

## Quick Start

```bash
adr init docs/decisions              # initialize ADR directory
adr new "Use PostgreSQL"             # create a new ADR
adr show 1                           # display ADR in terminal
adr update 1 accepted                # update ADR status
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

Display an ADR in the terminal.

| Flag | Description |
|------|-------------|
| `--plain` | Disable colored output |
| `--json` | Output as JSON |

### `adr update <id> [status]`

Update the status of an existing ADR. When status is omitted, an interactive menu is shown.

Valid statuses: `proposed`, `accepted`, `rejected`, `deprecated`, `superseded`.

## Web Server

```bash
adr-web             # starts on :8080
adr-web --addr :3000
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
