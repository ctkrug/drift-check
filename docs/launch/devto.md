---
title: "Building Pinset: finding toolchain drift across a monorepo"
published: false
description: "How I built a static Go CLI that compares language pins, CI setup steps, and installed runtimes."
tags: go, cli, devops, monorepo
---

# Building Pinset: finding toolchain drift across a monorepo

I kept seeing the same small failure in repositories that use more than one
language. A runtime gets upgraded in the obvious file, but the same version is
also pinned in CI and installed on developer machines. One of those copies gets
missed. The pull request looks fine until a build runs under a different
toolchain.

I built [Pinset](https://apps.charliekrug.com/drift-check/) to make that
disagreement visible with one command. It scans a repository for Node, Python,
Go, and Ruby pins, reads matching GitHub Actions setup steps, checks the
toolchains on `PATH`, and exits nonzero when any two claims disagree. The source
is on [GitHub](https://github.com/ctkrug/drift-check).

## A static binary that audits other runtimes

The first decision was to write Pinset in Go and keep the implementation on the
standard library. A tool that diagnoses a broken Node or Python setup should not
need npm or pip before it can start. `go build` produces one binary, and the
release workflow cross-compiles it for Linux and macOS.

That choice shaped the GitHub Actions parser. Pinset does not need a complete
YAML object model. It needs a narrow sequence: find a `uses:` entry for
`actions/setup-go`, `actions/setup-node`, `actions/setup-python`, or
`ruby/setup-ruby`, then read the requested version key from the sibling `with:`
block. A line-oriented scanner handles that shape with no runtime dependency.

The trade-off is explicit. This scanner is not a general YAML parser. If a file
contains a line larger than the scanner limit, Pinset warns and skips that file
instead of crashing or hiding pins found elsewhere. Tests cover quoted values,
comments, lookalike action names, duplicate steps, multiple workflow files, and
oversized garbage input.

## Prefix matching needed pairwise checks

Version comparison looked simple until patch versions entered the picture. A
pin such as `go 1.24` should agree with an installed `1.24.3`, so Pinset treats
the shorter dotted version as a prefix. That avoids reporting normal
major-minor versus patch differences as drift.

My first implementation compared every value with the first pin. That is wrong
because prefix compatibility is not transitive. The set `1.24`, `1.24.1`, and
`1.24.2` makes the broad value agree with both exact values, even though the two
patch versions disagree with each other. The fix was to compare every pair.
The number of pins per ecosystem is tiny, so the quadratic loop is clearer and
safer than a more elaborate normalization scheme.

Monorepo boundaries caused a similar correction. Pin files belong to nested
project roots, but GitHub only reads workflows from the repository root. Each
detector now receives both paths. A file such as `services/api/go.mod` keeps its
project-relative label while it is reconciled with every matching root workflow
step. The end-to-end fixture compiles the real binary and checks that behavior
against a golden report.

## What I would do differently

I would model the repository root and project root as separate concepts from
the first detector interface. Starting with one `root` argument made the early
implementation tidy, but it also made it easy to scan for CI beside a nested
module, where GitHub would never look. I would also write the multiple-workflow
case before the first parser implementation; retaining only the first match is
an easy mistake when a helper returns one string instead of a slice.

For a later release, I would consider a YAML library if real repositories show
that anchors, expressions, or uncommon step shapes matter enough to justify the
dependency. For version 1, the constrained scanner keeps the binary small and
the accepted input shape easy to test.

Pinset is available at
[apps.charliekrug.com/drift-check](https://apps.charliekrug.com/drift-check/),
with installation instructions and source at
[github.com/ctkrug/drift-check](https://github.com/ctkrug/drift-check).
