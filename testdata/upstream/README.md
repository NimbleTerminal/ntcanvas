# Upstream behavior fixture

`snapshot.json` was generated from unmodified NTCharts canvas code at
`bed05cf56936231c7c4b44833b7ad0457aedc5d6`, using that revision's go.mod/go.sum
and Go 1.27.2. The upstream MIT attribution is retained in the repository LICENSE.

The original scenarios in `scenarios.go.txt` use only upstream public APIs.
`internal/parity/scenarios.go` adapts them to the current API. They cover
coordinates/rounding, styled output, clipping, individual cell styling, fill,
resize, shifts, viewport edges, Unicode, and keyboard/wheel movement. Normal native and WASM tests compare the extracted implementation
with this checked-in result; no sibling checkout or fixture regeneration occurs
in CI. Aliasing and mouse-zone behavior have separate assertion-based tests.

To intentionally regenerate after reviewing a baseline/scenario change:

```
python3 scripts/generate-upstream-fixture.py --upstream /path/to/ntcharts
```

This reads the pinned git revision into a temporary module and runs the scenarios
against upstream, never the ntcanvas implementation. Review the resulting diff.

When numeric scaling was removed from this module, its six fixture entries were
deleted. All remaining upstream values are unchanged; they were not regenerated.

The surface split changes only the oversized-viewport trailing-newline rule in
these cases. The current adapter explicitly restores that one newline when
comparing with the historical fixture; direct Surface tests assert the new rule.
The fixture generator runs the archived upstream scenarios, never the adapter.

`canvas_test.go.txt` is a reference copy of the local NTCharts canvas tests inspected
on 2026-10-13 (checkout HEAD `0894ffed69fd00fcd9d6f9e3e4d0e3a4a9ae2007`).
The corresponding tests in the module root exercise the new grid
and model APIs. Model resize assertions now check independent viewport ownership.
