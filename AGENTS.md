# Working on ntcanvas

- Keep changes focused on this repository. Read nearby code and tests first.
- Keep the library independent of NTCharts, ntpicture, and booba. Examples use
  their own module; the browser Bubble Tea fork belongs only in `wasm.work`.
- Keep Surface small and Model independent of cell storage. Share storage
  internally; keep rune and grapheme mutation rules explicit. Document API and
  behavior changes in `CHANGES.md` and `API.md`.
- Keep upstream fixtures independent. Do not regenerate them just to make a
  failing test pass; follow `testdata/upstream/README.md`.
- Run relevant checks while working and `task ci` for code or dependency changes.
  See `DEVELOP.md` for commands. Report blockers and visual checks separately.
- Do not track generated binaries, WASM/runtime assets, or site output. Preserve
  license attribution and the README acknowledgements, license, and footer.
- Do not commit, push, publish, or release unless requested.
