# Terminal screen fixtures

`frames.json` contains hand-authored expected screen rows and nondefault cell
styles. It is independent of the upstream byte-output fixture. Every unlisted
cell must have default colors and no attributes.

`internal/tools/screenframes` emits real ntcanvas output for composition,
panning, clearing, resizing, sanitized text, compact styles, and grapheme boundaries. A minimal host adds cursor
positioning, line erasure, and CRLF; ntcanvas does not own terminal framing.
`scripts/check-screen.mjs` feeds consecutive frames into the demo's pinned Ghostty
emulator in three-byte chunks, then checks interpreted text, colors, bold, grapheme extras, and column widths.
The same terminal is reused so stale output is observable.

Run `task build-wasm-site` to prepare the runtime and run the checks, or
`task test-screen` once the runtime is built. Both native and WASM output are
checked. `task ci` includes these checks; generated frames stay under `dist/`.

Review fixture edits by hand. There is no automatic golden-update command.
These tests cover terminal interpretation, not Bubble Tea's rendering loop,
a PTY, browser painting, physical input, or universal Unicode/font behavior.
