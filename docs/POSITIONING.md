# Pinset positioning

## Audience brief

Pinset is for platform engineers and senior maintainers responsible for polyglot
monorepos. Their version requirements are scattered across language pin files,
GitHub Actions, and local toolchains, so a routine upgrade can leave CI and
developer machines on different releases. Pinset gives them one read-only report
that names every mismatch before it turns into a failed build.

## Product name

**Pinset**

The name treats a repository's version claims as one set that should agree. It is
short, easy to say, and specific to the tool's job without claiming to manage or
rewrite versions.

## Tagline

**Keep every toolchain pin in agreement.**

## Copy voice

Write for engineers who want evidence, not ceremony. Lead with the mismatch Pinset
finds, name the exact files and toolchains it reads, and show real terminal output.
Keep claims narrow: Pinset audits versions and exits nonzero on drift; it does not
install toolchains, edit files, or replace a version manager.
