# Vision

## The problem

A polyglot monorepo — Node services next to a Python data pipeline next to
a Go CLI, maybe a Ruby admin app — accumulates version pins in a dozen
uncoordinated places: `.nvmrc`, `.python-version`, `go.mod`, `.ruby-version`,
`Gemfile.lock`, GitHub Actions workflow files, Dockerfiles, `mise`/`asdf`
configs. Every one of these is a promise about what version a tool should
run at. Nothing checks that the promises agree.

Drift accumulates silently: someone bumps `go.mod` to require 1.24 for a new
stdlib feature but forgets the CI workflow still pins `setup-go@1.23`.
Someone else has 1.22 in their local shell because they haven't reinstalled
in months. The build passes locally, fails in CI, or — worse — passes in
CI and behaves subtly differently in production. Nobody notices until it
breaks something, and by then the drift could be months old and span
several files nobody thought to cross-check.

This is exactly the complaint a Pulumi engineer raised publicly: three
different Go versions in play (module file, CI, local shell) with nothing
to catch the mismatch before it caused a real problem.

## Who it's for

Engineers and tech leads on polyglot codebases — teams with more than one
language's toolchain in the same repo — who want a fast, honest answer to
"are our version pins actually consistent right now?" without adopting a
new version manager, writing a bespoke script, or trusting institutional
memory.

## The core idea

**Read-only reconciliation, not enforcement.** `drift-check` doesn't manage
versions, install toolchains, or rewrite pin files. It has exactly one job:
find every version claim in a repo — from pin files, from CI config, and
from what's actually installed — and tell you where they disagree. It is a
detector, not an actor. That's what makes it adoptable in five minutes: no
migration, no new mandatory tool, no changed workflow. Point it at a repo,
get a report, done.

## Key design decisions

- **Single static binary, Go stdlib-first.** No runtime, no package
  manager, no install step beyond downloading one file. This matters
  precisely because the tool audits *other* language runtimes — it can't
  itself depend on npm or pip being correctly set up.
- **Detector interface per ecosystem.** Each language is an isolated
  `ecosystem.Detector` implementation returning a common `Result` shape.
  Adding a fifth ecosystem (Rust, Java, whatever) means writing one file,
  not touching the reconciliation or reporting logic.
- **Three-way reconciliation, not two-way.** Comparing "pin file vs.
  installed" alone misses the more insidious case: CI has its own
  independent pin that neither the pin file nor a local dev's shell agree
  with. All three sources are first-class.
- **Prefix-based version comparison.** A `go.mod` directive like `go 1.24`
  is a floor, not an exact pin — it should agree with any `1.24.x` install.
  Comparison logic treats the shorter version string as a prefix match
  against the longer one, so this doesn't produce false-positive drift.
- **Exit code as the CI contract.** Non-zero exit on any drift found. This
  is what makes the tool useful as a CI gate, not just an interactive
  report — "catch it before merge," not just "catch it eventually."
- **No config file for v1.** Zero-config by default (walk cwd, detect
  what's there). Config only gets added later if real usage demands it —
  premature configurability is a cost, not a feature.

## What "v1 done" looks like

Running `drift-check` in a real polyglot monorepo (Node + Python + Go +
Ruby, each with its own pin file and a GitHub Actions CI config that pins
versions independently) produces a single report that:

- Lists every ecosystem detected, with every version source found for it
  (pin file, CI, installed).
- Clearly marks which ecosystems agree and which drift, with the specific
  disagreeing sources named (not just "drift detected" — *which* sources
  and *what* versions).
- Exits non-zero if and only if drift was found, so it's usable unmodified
  as a CI step.
- Runs in well under a second on a normal-sized monorepo, since it's just
  file parsing and a handful of subprocess calls — no network access, no
  package registry queries.

That's the whole product. Everything else — JSON output, more ecosystems,
Dockerfile/`mise` support, a `--fix` suggestion mode — is explicitly
post-v1 and tracked in `BACKLOG.md` as stretch epics, not blockers.
