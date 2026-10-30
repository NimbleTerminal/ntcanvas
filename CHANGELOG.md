# Changelog

## Unreleased

## 0.1.3 — 2026-10-30

- Update the examples to go-booba v0.10.7 for browser terminal box-drawing
  continuity, improved font defaults, and Safari WebTransport connection fixes.

## 0.1.2 — 2026-10-28

- Update example dependencies to go-booba v0.10.5 and the browser-only Bubble
  Tea fork to its latest WASM revision, including browser clipboard, mouse-wheel,
  and GPU color-emoji fixes.
- Refresh native and example dependencies to Bubble Tea v2.1.1 and the latest
  Ultraviolet revision. Update development and CI to Go 1.27.3, validate with
  Node 26.12.1, and refresh pinned Go lint and vulnerability tools.

## 0.1.1 — 2026-10-19

- Ignore key releases in the logo example, document shared mouse-zone IDs in
  model copies, and avoid reallocating grids when dimensions are unchanged.
- Complete public API documentation and correct PgUp/PgDown help labels.
- Test the minimum supported Go version and clarify development requirements.

## 0.1.0 — 2026-10-13

Initial release of ntcanvas, a terminal-cell canvas library for Go extracted
from NTCharts.

- Styled rune, compact rune, and Unicode grapheme grids.
- Bubble Tea viewports with keyboard and mouse panning, including shared grids.
- Cell styling, clipped drawing and copying, resizing, and antialiased lines.
- Native examples and interactive browser demos.
