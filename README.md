# Pinset

**▶ Live page: [apps.charliekrug.com/drift-check](https://apps.charliekrug.com/drift-check/)**

**Keep every toolchain pin in agreement.**

[![CI](https://github.com/ctkrug/drift-check/actions/workflows/ci.yml/badge.svg)](https://github.com/ctkrug/drift-check/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-175cd3.svg)](LICENSE)

Pinset is a read-only CLI for platform engineers and maintainers of polyglot
monorepos. One command finds Node, Python, Go, and Ruby version pins, compares
them with GitHub Actions and the toolchains on `PATH`, and names every mismatch.

```text
$ drift-check .
Go      go.mod=1.24  .github/workflows/ci.yml=1.23  installed=1.24.3               ⚠ DRIFT
Node    .nvmrc=20.11.0  .github/workflows/ci.yml=20.11.0  installed=20.11.0        ✓
Python  .python-version=3.12  .github/workflows/ci.yml=3.11  installed=3.12.1     ⚠ DRIFT
Ruby    .ruby-version=3.3.0  Gemfile.lock=3.3.0  installed=3.3.0                  ✓

2 drift(s) found across 4 ecosystem(s).
```

Pinset never edits a pin file or installs a runtime. It reports the repository
that already exists, then exits nonzero when any pair of version claims
disagrees.

## Install

Install the latest tagged release with Go:

```sh
go install github.com/ctkrug/drift-check@latest
```

Or download the static binary for Linux or macOS from
[GitHub Releases](https://github.com/ctkrug/drift-check/releases).

## Usage

Audit the current repository:

```sh
drift-check
```

Audit another path:

```sh
drift-check ../platform-monorepo
```

Emit JSON for scripts and CI annotations:

```sh
drift-check --json . | jq .
```

The command exits with:

- `0` when every detected pin agrees, or no supported pin files exist
- `1` when drift is found or the scan cannot complete
- `2` when command-line arguments are invalid

Nested projects are discovered recursively. Pinset skips `.git`,
`node_modules`, and `vendor`, and compares every nested project with every
matching setup step in the repository's root GitHub Actions workflows.

## Supported version sources

| Ecosystem | Repository pins | GitHub Actions pin | Installed command |
|---|---|---|---|
| Go | `go.mod` | `actions/setup-go` | `go version` |
| Node | `.nvmrc` | `actions/setup-node` | `node -v` |
| Python | `.python-version` | `actions/setup-python` | `python3 --version`, then `python --version` |
| Ruby | `.ruby-version`, `Gemfile.lock` | `ruby/setup-ruby` | `ruby -v` |

A missing installed toolchain appears as `installed=not found` and counts as
drift. Pinset does not silently turn an incomplete machine into a clean report.

## Version matching

Versions use dotted-prefix compatibility. A broad pin such as `1.24` agrees
with `1.24.3`; two exact pins such as `1.24.1` and `1.24.2` do not. Every pin is
compared with every other pin in the same ecosystem result, so a broad file pin
cannot hide disagreement between CI and the installed runtime.

## Add Pinset to CI

Run the released module directly in a GitHub Actions job:

```yaml
- name: Check toolchain version drift
  run: go run github.com/ctkrug/drift-check@latest .
```

The step fails when Pinset finds drift, which makes the exit code the CI
contract. No configuration file or repository migration is required.

## Develop locally

Pinset uses Go 1.22 and the standard library.

```sh
make build
make test
make vet
make fmt
```

The suite includes detector tests, malformed-workflow probes, report golden
files, and an end-to-end test that compiles the binary and audits a nested
four-language fixture.

Read [the architecture map](docs/ARCHITECTURE.md),
[the product brief](docs/POSITIONING.md), and
[the design system](docs/DESIGN.md) for more detail.

## License

MIT. See [LICENSE](LICENSE).

More of Charlie's projects → [apps.charliekrug.com](https://apps.charliekrug.com)
