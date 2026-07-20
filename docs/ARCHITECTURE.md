# Architecture

A concise map of the codebase for anyone picking this project back up cold.
See [`VISION.md`](VISION.md) for why it exists and [`BACKLOG.md`](BACKLOG.md)
for what's built vs. planned.

## Layout

```
main.go                          CLI entrypoint: flags, detector fan-out, exit code
internal/ecosystem/
  ecosystem.go                     Pin, Result, Detector — the shared shape
  golang.go                        Go detector + reconcile()/versionsAgree() (shared logic)
  node.go                          Node detector
  python.go                        Python detector
  ruby.go                          Ruby detector (two pin files: .ruby-version, Gemfile.lock)
  workflow.go                      GitHub Actions workflow pin parser, shared by all 4 detectors
  walker.go                        Recursive project discovery, skips VCS/dependency trees
internal/report/
  report.go                        text (Write) and JSON (WriteJSON) rendering
testdata/monorepo/                 Stable fixture and golden report for compiled-binary e2e
.github/workflows/release.yml      Tagged static-binary release build
docs/
  VISION.md, BACKLOG.md, ARCHITECTURE.md
```

## Data flow

1. `main.run()` calls `ecosystem.FindProjectRoots(root)`, which walks nested
   directories containing recognized pin files and ignores `.git`,
   `node_modules`, and `vendor` trees.
2. For each discovered project root, `main.run()` builds one
   `ecosystem.Detector` per language and calls `Detect(projectRoot)`. Nested
   pin sources are reported relative to the requested root.
3. Each detector independently:
   - reads its pin file(s) (`go.mod`, `.nvmrc`, `.python-version`,
     `.ruby-version`/`Gemfile.lock`) — returns `nil, nil` if absent, so a
     missing ecosystem is silently skipped, not an error;
   - calls `findWorkflowPins(root, actionPrefix, inputKey)` to look for a
     matching `actions/setup-*` (or `ruby/setup-ruby`) step in
     `.github/workflows/*.yml` and adds a `ci`-sourced pin if found;
   - shells out to the installed toolchain (`go version`, `node -v`,
     `python3 --version`, `ruby -v`) and adds an `installed`-sourced pin;
     a missing executable is represented as `installed=not found`;
   - calls the shared `reconcile(pins)` (defined in `golang.go`, used by all
     four detectors) to compare every pin against the first and set
     `Result.Drift` / `Result.Detail`.
4. `main.run()` collects the non-nil `*ecosystem.Result`s and hands them to
   `report.Write` (text) or `report.WriteJSON` (`--json`), which return the
   count of drifted ecosystems.
5. Exit code: `0` if nothing drifted (including "no pin files found" at
   all), `1` if anything did, `2` on a flag-parsing error.

## Key design points

- **`ecosystem.Detector` interface**: adding a fifth ecosystem means one new
  file implementing `Name()` + `Detect(root)` and one registration in
  `main.go`; the report package remains unchanged.
- **`workflow.go` is a narrow, purpose-built line scanner**, not a general
  YAML parser — it only understands the specific "list item with a `uses:`
  key and a sibling `with:` map" shape GitHub Actions steps take. This
  keeps the stdlib-only dependency story from VISION.md intact. A file it
  can't scan (huge line, read error) is logged to stderr and skipped, not
  fatal.
- **`reconcile`/`versionsAgree` are ecosystem-agnostic** (live in
  `golang.go` for historical reasons, used by all four detectors): version
  comparison treats the shorter dotted string as a prefix of the longer one
  (`"1.24"` agrees with `"1.24.3"`), and drift detail names every pin's
  source, not just the first mismatching pair.
- **`report.Write` sizes columns to the actual content of the current run**
  (not a fixed width) so a long CI workflow path as a source label doesn't
  break alignment for shorter rows — see the golden-file test in
  `internal/report/golden_test.go`.

## How to run / test

```
make build   # go build -o drift-check .
make test    # go test ./...
make vet     # go vet ./...
make fmt     # gofmt -l .  (lists unformatted files; empty output = clean)
make run     # build + run against the current directory
```

Update `internal/report/testdata/golden_report.txt` with
`UPDATE_GOLDEN=1 go test ./internal/report/...` if you intentionally change
report formatting.

## Reliability checks

- `e2e_test.go` builds the actual binary and runs it against the committed
  nested fixture using deterministic fake toolchains, comparing output to a
  golden file.
- The release workflow cross-compiles static Linux and macOS binaries only
  for version tags, then publishes all assets from one job; normal CI
  continues to run build, vet, format, and tests.
