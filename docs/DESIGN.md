# Pinset design direction

## Aesthetic direction

**Calibration-lab instrument panel.** Pinset uses pale celadon enamel, graphite
labels, cobalt controls, and signal-orange drift marks to feel like a precise
bench instrument built for maintainers. This light industrial palette avoids the
risograph, editorial, blueprint, Swiss-grid, and warm reading-room directions in
recent portfolio projects.

## Tokens

| Token | Value | Use |
|---|---|---|
| `--bg` | `#dce8e1` | pale celadon page ground |
| `--surface-1` | `#f4f7f2` | raised instrument face |
| `--surface-2` | `#c8d7d0` | recessed rails and code wells |
| `--text` | `#16211d` | graphite primary text |
| `--text-muted` | `#465c53` | secondary labels and captions |
| `--accent` | `#175cd3` | cobalt mark, links, focus, and CTA |
| `--support` | `#c2411d` | signal-orange drift state |
| `--success` | `#16734a` | matching versions |
| `--danger` | `#b42318` | errors and failed checks |

- **Type:** Space Grotesk for the wordmark and headings; IBM Plex Mono for
  interface labels, body copy, commands, and terminal output. Each uses a
  geometric sans or monospace system fallback.
- **Spacing:** 4px base with 8, 12, 16, 24, 32, 48, 64, and 96px steps.
- **Corners:** 10px on instrument panels, 6px on controls, and 2px on status
  labels.
- **Depth:** a 1px graphite-tinted edge, a short 4px panel shadow, and a broad
  low-opacity cobalt shadow. Recessed code wells use an inset edge.
- **Motion:** 180ms ease-out for links, controls, and panels; 120ms ease-out for
  pressed states. Reduced-motion users receive color and opacity changes only.

## Layout intent

The real audit report is the hero. At 1440 by 900, a compact wordmark rail sits
above a two-column hero that fills at least 72vh. The left 42 percent carries the
benefit, install command, and one GitHub CTA. The right 58 percent is a large
terminal instrument showing a four-language drift report. Proof, supported pins,
and the CI contract continue below in an asymmetric grid, so no small card floats
inside an empty page.

At 768px, the hero stacks while the report keeps the full content width. At 390 by
844, the wordmark and GitHub link remain on one compact rail, copy and CTA come
first, and the terminal follows in a horizontally scrollable well. Page gutters
shrink to 16px, every interactive target remains at least 44px tall, and sections
use full-width panels instead of compressed columns.

## Signature detail

The Pinset wordmark sits beside a small alignment reticle made from four cobalt
corner ticks and one orange center pin. On first load, a thin cobalt calibration
rule sweeps across the reticle and terminal header, visually connecting the brand
to the audit. The sweep is disabled when reduced motion is requested.

## Interaction plan

This is a landing page for a CLI, not a game or interactive toy. The GitHub CTA,
copyable command treatment, and links receive themed hover, focus-visible, active,
and disabled states where applicable. The page has no audio and no win state.
