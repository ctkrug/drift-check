# Architecture

This map is for maintainers returning to Pinset after time away. See
[`POSITIONING.md`](POSITIONING.md) for the product boundary and
[`BACKLOG.md`](BACKLOG.md) for completed and deferred work.

## Layout

```text
main.go                          CLI flags, detector fan-out, source labels, exit code
internal/ecosystem/
  ecosystem.go                    shared Pin, Result, and Detector types
  reconcile.go                    pairwise dotted-version comparison
  golang.go                       Go detector
  node.go                         Node detector
  python.go                       Python detector
  ruby.go                         Ruby detector for two repository pin formats
  workflow.go                     GitHub Actions setup-step scanner
  walker.go                       recursive project discovery and skip rules
internal/report/
  report.go                       aligned text and JSON rendering
testdata/monorepo/               compiled-binary end-to-end fixture and golden output
.github/workflows/release.yml    tagged static-binary release build
docs/                            product, design, architecture, and launch notes
site/                            static product landing page
```

## Data flow

1. `main.run()` calls `ecosystem.FindProjectRoots(repositoryRoot)`. The walker
   finds directories containing a supported pin file and skips `.git`,
   `node_modules`, and `vendor` trees.
2. For each project root, `main.run()` calls each detector with both the project
   root and repository root. Repository pin files come from the project root;
   GitHub Actions pins come from the root `.github/workflows` directory.
3. A detector returns `nil` when its project pin is absent. When present, it
   gathers the file pin, every matching workflow setup-step pin, and the
   installed toolchain version. A missing executable becomes
   `installed=not found` rather than disappearing from the report.
4. `reconcile()` compares every pin pair. Dotted versions agree when the shorter
   value is a prefix of the longer value, so `1.24` matches `1.24.3` while
   `1.24.1` does not match `1.24.2`.
5. `main.run()` labels nested file sources relative to the requested repository
   without rewriting root workflow paths, then passes all results to the text or
   JSON renderer.
6. The process exits `0` when all results agree, `1` on drift or scan failure,
   and `2` on invalid command-line usage.

## Extension points

`ecosystem.Detector` is the language boundary. A new ecosystem implements
`Name()` and `Detect(projectRoot, repositoryRoot)`, then registers one detector
in `main.go`. Reporting and version comparison stay unchanged.

`workflow.go` is a narrow, line-oriented scanner rather than a general YAML
parser. It recognizes a setup action's `uses:` entry and the requested key in
its sibling `with:` block. It retains every matching step across every `.yml`
and `.yaml` file. A file that exceeds the scanner limit or cannot be read emits
a warning and does not hide valid pins from other workflow files.

`report.Write` measures columns from the current result set. Long nested paths
and workflow labels therefore keep the status column aligned. `report.WriteJSON`
uses the same result objects and drift count as text output.

## Validation

```sh
make build
make test
make vet
make fmt
```

`e2e_test.go` builds the real binary, supplies deterministic fake toolchains,
and compares a nested four-language audit with a committed golden report. Unit
tests cover missing tools, malformed workflows, duplicate setup steps, nested
project discovery, pairwise comparison, JSON shape, exit codes, and report
alignment.

When report formatting changes intentionally, update the internal golden file
with:

```sh
UPDATE_GOLDEN=1 go test ./internal/report/...
```
