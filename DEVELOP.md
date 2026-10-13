# Development

Use Go 1.27.2, [Go Task](https://taskfile.dev), Node 26, and a C compiler for
race tests. Python 3 is needed to serve the demo.
Vulnerability scans and first-time tool installation require network access.
The Taskfile installs pinned golangci-lint into the ignored `.task/bin` directory.

Go 1.26.8 is the supported minimum. CI tests native and WASM builds with both
1.26.8 and 1.27.2, with automatic toolchain upgrades disabled; the full tooling
and site checks run on 1.27.2.

Node 26 is an intentional choice for the browser/WASM development tools: this
project tests against the current Node release line rather than waiting for LTS.
The local CI suite has been validated with Node 26.11.1. Node is not a dependency
of native library users.

## Commands

```sh
task test             # native library and example tests
task test-wasm        # browser-workspace tests under Node
task vet              # Go vet
task lint             # pinned golangci-lint, library and examples
task ci               # full check before submitting code changes
task run-minimal      # small Bubble Tea example
task run-logo         # original styled-cell example
task run-grids        # grid and Unicode comparison
task run-lines        # rasterization and coverage
task run-viewports    # shared storage and model interaction
task serve-wasm-site  # browser demo at http://localhost:8000/web/
```

`task ci` also runs lint and checks formatting, module metadata, race/no-CGO tests,
native/Windows/WASM builds, browser input, golden terminal screens, license notices, and vulnerabilities.
It does not need sibling repositories. Use `task --list-all` for other commands.

Run allocation benchmarks with:

```sh
go test -run='^$' -bench=. -benchmem .
```

## Demo recording

Install `ntrecord` with `brew install --cask nimbleterminal/tap/ntrecord`, then run
`task record-demo`. It builds the examples and runs [examples/demo.tape](examples/demo.tape)
with ntrecord's deterministic clock. ntrecord supports VHS tapes and Kitty graphics;
it renders on a real PTY without needing a browser, GUI terminal, or FFmpeg for GIFs.

The tape records panning, focus switching, and a shared edit to
`dist/viewports.gif` and `dist/viewports.png`, then captures the grid comparison
to `dist/grids.png`. The task copies the viewport PNG to
`docs/images/viewports.png` for the README. Review the
image before committing it; recordings and intermediate captures stay ignored
under `dist/`. Screen waits synchronize each scene, and shell prompts are hidden.
Recording is an optional local task, separate from `task ci`.
The tape explicitly enables color even when the parent shell sets `NO_COLOR`.
Joined emoji in the grid capture may differ with the recorder/terminal font;
the README uses the shared-viewports scene, which avoids that font limitation.

## Reference tests

Keep native Bubble Tea in `go.mod` and the browser fork in `wasm.work`.
The root module pins upstream extraction dependencies; examples and WASM may
select newer versions. Test both graphs when changing dependencies.

Native and WASM tests compare against checked-in upstream output. Follow the
[fixture instructions](testdata/upstream/README.md) to regenerate it; CI never does.

## Browser checks

Run `task serve-wasm-site` and check:

- Help buttons appear above the terminal.
- Shared viewports: keys, wheel, and drag move only the focused canvas; Enter
  switches focus, e edits both views, g resizes storage, and b blurs both.
- Grid comparison: s toggles single-column rendering and c clears a wide cell.
- Lines: arrows adjust the endpoint and g switches glyphs.
- Quit, Restart, keyboard focus, and narrow layouts work.

Automated checks verify bridge input, viewport results, and interpreted terminal
screens. Run `task test-screen` after building the demo runtime; see the
[screen fixtures](testdata/screen/README.md). Browser painting and physical
pointer gestures still require visual verification.

Keep generated binaries, WASM/runtime assets, and site output ignored. Run
`task notices` after dependency changes and review `THIRD_PARTY_NOTICES.md`.
