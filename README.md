# Drift Check

**One command audits your polyglot monorepo and tells you exactly where pinned
versions disagree — with each other, or with what's actually installed.**

```
$ drift-check
go.mod        says go1.24   →  CI pins go1.23   →  installed go1.22   ⚠ DRIFT
.nvmrc        says 20.11.0  →  CI pins 20.11.0   →  installed 20.11.0  ✓
.python-version says 3.12.1 →  CI pins 3.11.7    →  installed 3.12.1   ⚠ DRIFT
Gemfile.lock  says 3.3.0    →  (no CI pin found) →  installed 3.3.0    ✓

2 drifts found across 4 ecosystems.
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

## Planned features

- **Ecosystem detectors**: `.nvmrc` / `package.json` engines (Node),
  `.python-version` / `pyproject.toml` (Python), `go.mod` / `go.sum` (Go),
  `.ruby-version` / `Gemfile.lock` (Ruby).
- **Installed-toolchain resolution**: shells out to each ecosystem's version
  command and normalizes the output for comparison.
- **CI pin extraction**: parses `.github/workflows/*.yml` for
  `actions/setup-node`, `actions/setup-go`, `actions/setup-python`,
  `ruby/setup-ruby`, and equivalents, pulling the pinned version out of each.
- **Drift report**: a single reconciled table (pin file vs. CI vs.
  installed) with clear pass/fail per ecosystem, plus a non-zero exit code
  when drift is found so it's CI-gateable.
- **Machine-readable output**: `--json` for scripting and dashboards.
- **Monorepo-aware**: walks subdirectories so a single run covers every
  package/service in a polyglot repo, not just the root.

## Stack

Go (single static binary, stdlib-first). No runtime dependencies beyond the
Go standard library where possible, so `go build` produces one binary you
can drop anywhere — no interpreter, no package manager, no install step.

## Status

Early scaffold. See [`docs/VISION.md`](docs/VISION.md) for the design and
[`docs/BACKLOG.md`](docs/BACKLOG.md) for the build plan.

## License

MIT — see [`LICENSE`](LICENSE).
