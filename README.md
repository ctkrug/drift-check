# Drift Check

**One command audits your polyglot monorepo and tells you exactly where pinned
versions disagree — with each other, or with what's actually installed.**

```
$ drift-check
Go      go.mod=1.24  .github/workflows/ci.yml=1.23  installed=1.22.2               ⚠ DRIFT
Node    .nvmrc=20.11.0  .github/workflows/ci.yml=20.11.0  installed=20.20.2        ⚠ DRIFT
Python  .python-version=3.12.1  .github/workflows/ci.yml=3.11.7  installed=3.12.3  ⚠ DRIFT
Ruby    Gemfile.lock=3.3.0                                                         ✓

3 drift(s) found across 4 ecosystem(s).
```

## Why

Polyglot monorepos accumulate version pins in a dozen places: `.nvmrc`,
`.python-version`, `go.mod`, `Gemfile.lock`, CI workflow files, Dockerfiles,
`mise`/`asdf` configs. Nothing keeps them honest against each other, or
against what's actually on a developer's machine. The drift is silent until
it breaks a build or — worse — behaves differently in prod than in CI. A
Pulumi engineer complained about exactly this on GitHub: `go.mod` said one
Go version, CI pinned another, and a contributor's shell had a third. Nobody
noticed until something broke.

Existing tools solve *part* of this: version managers (`asdf`, `mise`) let
you pin versions but don't audit for drift across ecosystems, and they
require migrating your whole team onto them. Linters catch syntax, not
version disagreement. Nothing reconciles "what every pin file says" against
"what's installed" and "what CI actually uses" in one pass.

## What it is

A single static Go binary, `drift-check`, that:

1. Walks a repo and parses every version-pinning file it recognizes across
   Node, Python, Go, and Ruby.
2. Resolves what's actually installed in the current shell (`node -v`,
   `python --version`, `go version`, `ruby -v`).
3. Parses CI workflow files (starting with GitHub Actions) for version pins
   used in the pipeline.
4. Reconciles all three into one drift report: which pins agree, which
   don't, and exactly where the disagreement is.

It is **read-only**. It never rewrites a version file, never installs a
toolchain, and never asks you to adopt a new version manager. It just tells
you the truth about what's already there.

## Usage

```
$ drift-check [flags] [path]
```

Install the latest tagged release with Go:

```
go install github.com/ctkrug/drift-check@latest
drift-check .
```

Or download the matching static binary from the project's GitHub Releases.
Defaults to the current directory. The scan walks nested project directories
and skips `.git`, `node_modules`, and `vendor`. Exits `0` when nothing drifts
(or no pin files are found at all), `1` when at least one ecosystem drifts —
so it drops straight into a CI job as a gate:

```yaml
- run: go run github.com/ctkrug/drift-check@latest
```

Flags:

- `--json` — emit a machine-readable report instead of text. Same exit
  code semantics as the text mode. Pipe through `jq .` to inspect.

## What's implemented

- **Ecosystem detectors**: `go.mod` (Go), `.nvmrc` (Node), `.python-version`
  (Python), and `.ruby-version` / `Gemfile.lock` (Ruby) — the latter two
  tracked as distinct pins so a mismatch between them is named directly.
- **Installed-toolchain resolution**: shells out to each ecosystem's version
  command (`go version`, `node -v`, `python3 --version`, `ruby -v`) and
  normalizes the output for comparison. A missing executable is reported as
  `installed=not found` and causes drift rather than being silently omitted.
- **CI pin extraction**: parses `.github/workflows/*.yml` for
  `actions/setup-go`, `actions/setup-node`, `actions/setup-python`, and
  `ruby/setup-ruby`, pulling the pinned version out of each. A workflow file
  that can't be parsed is skipped with a warning, not a crash.
- **Drift report**: a single reconciled table (pin file vs. CI vs.
  installed) with clear pass/fail per ecosystem, columns sized to the
  actual content so long CI paths don't break alignment.
- **Machine-readable output**: `--json` for scripting and dashboards.
- **Prefix-based version comparison**: a `go.mod` directive like `go 1.24`
  agrees with any installed `1.24.x`, matching the same rule across all
  four ecosystems.
- **Monorepo discovery**: recursively finds supported pin files and labels
  nested sources (for example, `services/api/go.mod`) in the report.
- **Release binaries**: version tags publish static binaries for Linux amd64,
  macOS amd64, and macOS arm64.

See [`docs/BACKLOG.md`](docs/BACKLOG.md) for the full build plan.

## Stack

Go (single static binary, stdlib-first). No runtime dependencies beyond the
Go standard library where possible, so `go build` produces one binary you
can drop anywhere — no interpreter, no package manager, no install step.

## Status

Core reconciliation, monorepo discovery, robustness checks, and distribution
automation are done and tested. See [`docs/VISION.md`](docs/VISION.md) for the
design and [`docs/BACKLOG.md`](docs/BACKLOG.md) for the delivery checklist.

## License

MIT — see [`LICENSE`](LICENSE).
