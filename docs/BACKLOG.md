# Backlog

Epics are ordered for build sequencing. Epic 1's first story is the first
useful report and must be reachable with only Epic 1 built.

## Epic 1: Core reconciliation and the wow-moment demo

Get from "reads go.mod" to "a real three-way drift report a user can run
in their own repo and trust."

- [x] **Run drift-check in a real polyglot monorepo and get an instant
      three-way Go drift report.** (WOW MOMENT)
  - Given a fixture repo with `go.mod` (`go 1.24`), a GitHub Actions
    workflow pinning `actions/setup-go@v5` to `go-version: "1.23"`, and
    the CLI run under a Go 1.22 toolchain, `drift-check` prints a report
    showing all three versions and marks the ecosystem as drifted.
  - The report names which specific sources disagree (e.g. "go.mod says
    1.24, ci says 1.23, installed says 1.22"), not just "drift: true".
  - Running `drift-check` in a repo where go.mod, CI, and installed all
    agree exits `0` and prints no drift.

- [x] **Parse GitHub Actions workflow files for version-setting steps.**
  - Given a `.github/workflows/*.yml` using `actions/setup-go` with a
    `go-version` input, drift-check extracts that version as a `ci`-source
    pin for the Go ecosystem.
  - A workflow file with no `setup-go` step produces no `ci` pin for Go
    (the ecosystem still reports pin-file vs. installed only) rather than
    erroring.
  - Malformed or unparseable YAML in a workflow file is skipped with a
    warning to stderr, not a crash.

- [x] **Add `--json` output for scripting and dashboards.**
  - `drift-check --json` on a repo with one drifted ecosystem produces
    valid JSON (verified by piping through `jq .`) containing the
    ecosystem name, every pin's source and version, and a `drift` boolean.
  - Exit code behavior (`0` clean, `1` drift found) is identical between
    text and JSON output modes.

- [x] **Exit non-zero exactly when drift is found, usable as a CI gate.**
  - A repo with zero drifted ecosystems exits `0`.
  - A repo with at least one drifted ecosystem exits `1`.
  - A repo with no recognized pin files at all (nothing to check) exits
    `0` and says so explicitly, rather than silently succeeding.

## Epic 2: Full ecosystem parity (Node, Python, Ruby)

Bring the three stub detectors up to the same reconciliation depth as Go:
pin file vs. installed vs. CI, not just presence detection.

- [x] **Reconcile Node version across .nvmrc, `node -v`, and CI.**
  - A repo with `.nvmrc` containing `20.11.0` and an installed Node
    reported by `node -v` as `v20.11.0` shows no drift.
  - A mismatch between `.nvmrc` and installed Node is reported with both
    versions named, matching the Go detector's message format.
  - `actions/setup-node`'s `node-version` input is parsed as the CI pin,
    same as `setup-go` in Epic 1.

- [x] **Reconcile Python version across .python-version, `python
      --version`, and CI.**
  - A repo with `.python-version` containing `3.12.1` and an installed
    interpreter reporting `Python 3.12.1` shows no drift.
  - Version comparison uses the same prefix-agreement rule as Go (a
    `.python-version` of `3.12` agrees with an installed `3.12.4`).
  - `actions/setup-python`'s `python-version` input is parsed as the CI
    pin.

- [x] **Reconcile Ruby version across .ruby-version, Gemfile.lock, `ruby
      -v`, and CI.**
  - A repo with both `.ruby-version` and a `Gemfile.lock` `RUBY VERSION`
    stanza pinning different patch versions is reported as drifted between
    those two sources specifically (not conflated into one).
  - `ruby/setup-ruby`'s `ruby-version` input is parsed as the CI pin.
  - A repo with only `Gemfile.lock` (no `.ruby-version`) still reconciles
    against installed and CI.

- [x] **Align report formatting across all four ecosystems.**
  - Running drift-check on a fixture with all four ecosystems present
    produces a report where every ecosystem's line is legible and
    consistently formatted (no column misalignment from long version
    strings), verified against a golden-file snapshot.

## Epic 3: Monorepo scale and reliability

A real polyglot monorepo has pin files nested in subdirectories, missing
toolchains, and occasionally malformed files. The tool needs to survive
all of that without crashing or lying.

- [x] **Walk subdirectories to find pin files in nested
      packages/services.**
  - A fixture with `services/api/go.mod` and `services/web/.nvmrc` (no
    pin files at repo root) is fully detected when drift-check is run
    from the repo root; both ecosystems appear in the report with the
    correct relative source paths.
  - A `.git`, `node_modules`, or `vendor` directory is not descended into
    (verified by timing or by planting a decoy pin file inside one and
    confirming it's excluded).

- [x] **Handle a missing toolchain gracefully.**
  - Running drift-check in a repo with a `.ruby-version` file but no
    `ruby` binary on `PATH` reports the pin-file version with the
    installed source explicitly marked "not found," rather than crashing
    or silently omitting the ecosystem.

- [x] **Handle incomplete pin files without a panic.**
  - A `go.mod` with no `go` directive line is skipped for that repo (no
    Go result), not a crash.
  - An empty `.nvmrc` file is treated as absent, not as a pin to an empty
    string.

- [x] **Add an end-to-end CLI integration test against a fixture
      monorepo.**
  - A committed `testdata/` fixture with all four ecosystems (some
    drifted, some not, one nested) is run through the actual compiled
    binary in CI, and its exit code and output are asserted against a
    golden file.

## Epic 4: Distribution polish

- [x] **Add install and quickstart instructions to the README.**
  - README documents `go install github.com/ctkrug/drift-check@latest`
    and shows a representative four-language report.

- [x] **Build and attach release binaries on tagged releases.**
  - Pushing a `v*` tag triggers a GitHub Actions job that cross-compiles
    for `linux/amd64`, `darwin/amd64`, and `darwin/arm64`, and attaches
    the binaries to a GitHub Release.
  - A downloaded release binary run with `--help` produces the same usage
    text as running from source.

## Epic 5: Closeout correctness and product presentation

- [x] **Compare every version pin pair.**
  - A broad `1.24` pin cannot hide disagreement between exact `1.24.1`
    and `1.24.2` pins.

- [x] **Apply repository workflows to nested projects.**
  - A nested `services/api/go.mod` is reconciled with setup steps in the
    root `.github/workflows` directory.
  - Every matching setup step in every workflow file is retained.

- [x] **Package Pinset as a portfolio project.**
  - Positioning, design tokens, README, architecture, landing page, and
    launch notes use one product name and one audience brief.
