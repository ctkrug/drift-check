# Pinset vision

## The problem

A polyglot monorepo can name one runtime version in a language pin file,
another in GitHub Actions, and a third on a maintainer's machine. Those claims
live far enough apart that a routine upgrade often changes only one of them.
The mismatch stays quiet until a build fails or a contributor gets different
behavior from CI.

## Who Pinset serves

Pinset is for platform engineers and senior maintainers responsible for Node,
Python, Go, and Ruby projects in one repository. They need a quick answer to
"do our toolchain versions agree?" without migrating the team to a new version
manager or maintaining a custom audit script.

## Product boundary

Pinset performs read-only reconciliation. It discovers supported pin files,
reads every matching GitHub Actions setup step, checks installed toolchains,
and reports each disagreement. It does not install runtimes, rewrite files, or
choose the correct version for the team.

That narrow boundary keeps adoption small: run one static binary, inspect one
report, and use its exit code as a CI gate.

## Design principles

- **One repository view.** Nested projects and root workflows belong in the
  same audit, with source paths preserved in the output.
- **Missing is evidence.** If a pinned toolchain is not installed, the report
  says `installed=not found` and counts drift.
- **Every claim participates.** All matching workflow pins are retained, and
  every version pin is compared with every other pin in its result.
- **Broad pins remain useful.** A major-minor pin such as `1.24` agrees with an
  installed patch such as `1.24.3`; two different exact patches do not agree.
- **Text and JSON share a contract.** Human and machine-readable output use the
  same results and the same exit status.
- **No runtime dependency.** Go and the standard library produce a single
  static binary that can audit other language runtimes without depending on
  npm, pip, or Bundler.

## Version 1 contract

Given a repository containing any supported Node, Python, Go, or Ruby pin,
Pinset produces a deterministic report that:

- names each file, workflow, and installed source it checked;
- marks each ecosystem result as matching or drifted;
- reports nested projects with repository-relative paths;
- emits valid JSON when `--json` is set;
- exits nonzero exactly when drift or a scan failure occurs;
- never mutates the audited repository.

The committed end-to-end fixture and release workflow exercise that contract
against the compiled binary and tagged static builds.
