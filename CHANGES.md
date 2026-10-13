# Hardening after extraction

## Input and single-column boundary fixes

- Default keyboard panning ignores key releases, preventing a second movement
  when a host enables release reporting. Presses and repeats still pan.
- `SingleColumn` rendering in both rune grids separates adjacent runes that
  would merge into one grapheme (including Hangul Jamo and Unicode prepend
  characters). It uses the same ZWNJ boundary policy as GraphemeGrid, preserving
  stored cells and default rendering behavior. Styles remain trusted.

## Extraction hardening

The initial extraction preserved NTCharts behavior exactly. This follow-up
intentionally changes unsafe or ambiguous edge cases, while the upstream fixture
continues to guard ordinary cell output, clipping, styles, and coordinates.
The fixture is not regenerated to bless implementation changes.

- `New` now clamps negative dimensions like `Resize`. Both detect integer
  allocation overflow before allocating. They remain uncapped; `NewChecked`,
  `ResizeChecked`, and `ValidateSize` apply explicit `SizeLimits` and return errors.
- Empty canvases keep nonnegative cursor components. Empty/negative-size viewports
  render no content. Empty shifts are no-ops. An oversized positive viewport still
  retains the upstream bottom-edge trailing newline.
- Movement ignores nonpositive amounts and saturates without integer overflow.
  Right/down movement clamps to the last full viewport, including after an
  explicit edge cursor. Previously MoveRight checked one step then added any
  amount, and MoveDown could reject the entire oversized step.
- Focused Update with a nil handler is a no-op rather than a panic.

Additions: safe plain-text writers and sanitization, clipped geometry helpers,
checked sizing, and a minimal Bubble Tea example. Raw text writers preserve
terminal control sequences; the new safe writers require trusted styles.

Allocation improvements preserve observable cell contents and ownership:
Clear zeros existing rows; vertical shifts recycle a row; text writes decode only
as far as their right clipping boundary; View reserves for the visible grid.
Resize still detaches cell storage even when the dimensions are unchanged. Per-cell Lip Gloss rendering is retained:
combining runs could alter escape sequences, layout, or stateful transform calls.

Composition additions: `CopyRect` copies clipped regions with overlap-safe
snapshot semantics, including blank cells. `SetStyleRect` replaces cell styles
without changing runes or the default style. Existing methods are unchanged.

Scope: removed the unreleased `ntcanvas/buffer` package. Numeric scaling belongs
outside the terminal-cell canvas. No replacement package is provided here.

## Surface and grid API

This prerelease intentionally replaces `New(width, height)`/`NewChecked` with
concrete grid constructors and `NewModel(surface)`. Model no longer exposes cell
writes or storage resizing. Style/content options configure StyledRuneGrid;
viewport/input options configure Model. Old consumer aliases were removed.

Render output no longer has the historical oversized-viewport trailing newline.
Resizing a shared grid is visible to every model and does not reset viewports.
Built-in mouse tracking belongs to each Model, with explicit click/release state.

RuneGrid provides compact terminal styles. GraphemeGrid tracks display widths
and enforces complete-cluster mutations; see API.md for clipping and ownership.

## Consistent grid writes and rendering

- All grids use `SetText`/`SetTextWithStyle` and sanitize terminal controls by
  default. StyledRuneGrid replaces `SetString` and `SetPlainText` with this API;
  explicit `SetRawText` variants retain trusted legacy string writes. Cell writes
  and fills also sanitize controls, while zero-valued cells still mean blank.
  NUL in text becomes U+FFFD on every grid. Grapheme text retains shaping joiners.
- Grid `View` is removed. Use `Render(Bounds())` or wrap the grid in Model.
  All grids now provide sizing, line, fill, style, and shift conveniences.
  Grapheme shifts discard clusters cut at an edge.
- Styled blank rendering uses a constant space. Grapheme rendering skips cluster
  boundary checks between two single-byte ASCII cells. Allocation regression
  tests guard both paths; Unicode boundary checks remain in place.

## Package name

The root package is now `ntcanvas`, matching its module and repository. Imports
use `"nimbleterminal.dev/ntcanvas"` and refer to `ntcanvas.Surface`,
`ntcanvas.NewModel`, and the grid constructors without an alias. Existing callers
can explicitly alias the import as `canvas` while migrating. Attribution and
upstream test provenance are retained.

## Model receivers, compact SGR output, and lint gate

- `SetLines` and `SetLinesWithStyle` no longer return an always-true boolean.
- Model stores `*KeyMap` to reduce copy cost. Read methods, `Init`, and `Update`
  use value receivers; setters mutate through pointers. Model copies share key
  bindings, while separately constructed models get separate key maps.
  Empty viewport intersections anchor the cursor
  at the surface's minimum; explicit edge cursors and full-viewport panning retain
  their distinct behavior, documented in API.md.
- RuneGrid adds `SetRuneWithStyle`; text writes sanitize once per stored rune.
- Compact rendering builds SGR parameters on the stack and combines reset/new
  style in one sequence. Terminal output remains equivalent; escape bytes change
  for adjacent runs with different styles. StyledRuneGrid output is unchanged.
- CI now runs pinned golangci-lint on the library and examples. Existing lint
  findings are fixed. The ZWNJ terminal assumption is documented in API.md.

## Bounded Unicode rendering and workflow pins

- Grapheme clusters longer than 16 code points become a single U+FFFD cell on
  writes and fills. This bounds retained/rendered cluster complexity; applications
  must still limit input size and account for terminal/font differences.
- RuneGrid and StyledRuneGrid add opt-in `SingleColumn` rendering. Non-one-column
  runes become U+FFFD without changing stored cells. Default rune behavior stays
  unchanged; application styles remain trusted.
- API.md identifies layout trust boundaries, bidi limitations, style-copy
  ownership, and checked resizing for external dimensions.
- All GitHub Actions now use commit pins. Existing Dependabot configuration
  maintains GitHub Actions updates; Pages deployment permissions are unchanged.

## Clipped floating-point lines

`RasterLine` accepts Float64Point endpoints, clips to a cell-center rectangle,
and yields connected integer cells through a callback. It rounds to nearest with
half ties away from zero, includes endpoints once, rejects nonfinite input, and
bounds walking by the visible extent. Extreme coordinates and ambiguous rounding
use exact clipping; normal lines avoid allocations. See API.md for clipping and
rasterization tie rules. No grid, Surface, projection, or depth API changes.

## Antialiased line coverage

`RasterLineAA` adds callback-based Wu-style coverage without choosing colors or
glyphs. It emits unique in-bounds cells with endpoint caps, combines short-segment
weights, and preserves coverage when endpoints reverse. Degenerate points emit
one full-coverage cell. It shares clipping with RasterLine; exact fallback covers
extreme finite coordinates. API.md documents coverage and compositing semantics.
