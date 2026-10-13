# Examples

Run these from the repository root. Each browser demo publishes its keys through
booba so the same actions appear as buttons above the terminal.

| Command | What it shows | Controls |
| --- | --- | --- |
| `task run-grids` | StyledRuneGrid, RuneGrid, and GraphemeGrid with the same Unicode input | s: SingleColumn, c: clear a cell, r: reset |
| `task run-lines` | RasterLine beside RasterLineAA; explicit background/color blending | up/down: endpoint, g: dot/block glyph, r: reset |
| `task run-viewports` | Two Models viewing one GraphemeGrid, independent cursors and focus | Enter: switch, arrows/HJKL/wheel/drag: pan, e: edit, g: resize storage, b: blur, r: reset views |
| `task run-logo` | Original styled-cell canvas and mouse interaction | Enter: switch, arrows/HJKL/wheel/drag: pan |
| `task run-minimal` | Small Bubble Tea host around a GraphemeGrid | arrows/HJKL: pan |

Use q or Ctrl+C to quit. `task serve-wasm-site` serves the four interactive demos
at http://localhost:8000/web/. The minimal host remains a native teaching example.

Grid comparison intentionally shows rune-grid column drift with wide/combining
text. SingleColumn changes rendered output without rewriting cells. Clearing the
second half of a wide grapheme clears both halves, while a rune grid clears just
one rune. The line demo maps coverage to a foreground color against an explicit
background; the callback owns that policy.

Shared viewports keep grid storage separate from Model state. Terminal resizing
changes each view window; g explicitly resizes the grid. Both Models see shared
edits immediately. Their cursors remain independent and clamp to storage bounds.
Clicking outside both views blurs them; Enter or a click restores focus.

The examples module owns go-booba. Only browser builds select `wasm.work` and its
Bubble Tea replacement. No demo dependencies are added to ntcanvas's root module.
