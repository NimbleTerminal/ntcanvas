# API guide

Import `nimbleterminal.dev/ntcanvas`; its package name is `ntcanvas`.

## Surface and Model

```go
type Surface interface {
    Bounds() image.Rectangle
    Render(image.Rectangle) string
}
```

`Render` clips to bounds, returns empty output for an empty intersection, and
joins rows without a trailing newline. It returns content, not terminal cursor
commands or a Bubble Tea root view. Custom surfaces may use nonzero origins;
their bounds must have nonnegative dimensions representable by `int`.

`NewModel(surface, options...)` adds viewport, focus, keys, and mouse handling.
Retain the model returned by `Update`. Use `SetViewport`, `SetCursor`, or movement
methods to pan; `View` renders the visible region.
`SetCursor` clamps to the last cell and permits a partial viewport at an edge.
The `Move` methods stop at the last full viewport for interactive panning.
An empty `SetViewport` intersection sets zero view dimensions and anchors the
cursor at `Bounds().Min`, including for surfaces with nonzero origins. `Surface()` returns the wrapped
surface. A nil interface creates an empty model; typed nil pointers are invalid.

`WithFocus`, `WithCursor`, `WithViewWidth`, `WithViewHeight`, `WithKeyMap`,
`WithUpdateHandler`, and `WithZoneManager` configure the model. Mouse dragging
requires a scanned bubblezone manager and mouse reporting in the host. Dragging
starts with a click and ends on release or blur.
Keyboard panning handles key presses (including repeats) and ignores key releases.
PgUp and PgDown move one row, like Up and Down; they do not zoom or page by the
viewport height.

Read methods, `Init`, and `Update` use value receivers. Mutating setters use
pointer receivers. Retain the model returned by `Update`.

Model copies share the surface and `*KeyMap`, but copy viewport, focus, and
built-in drag state. Constructors allocate separate default key maps.
Copies also share the installed zone manager and zone ID. To render independent
mouse zones, construct separate models with `NewModel` rather than copying an
already zoned model; those models can still share the same surface and manager.
`WithKeyMap(value)` installs a separate shallow copy per model; binding key-name
slices retain their existing aliasing. To detach a copied model's bindings, copy
`*model.KeyMap` and assign a pointer to that copy. A nil KeyMap disables keyboard
bindings; mouse handling still works.
Custom callbacks can still share captured data. The model has no drawing methods
and never resizes storage. After a grid resize, its visible cursor is clamped to
the new bounds; viewport dimensions remain unchanged.

## Grids and ownership

All constructors return mutable pointers. Keep one owner/event loop; do not copy
grid values or mutate them concurrently. A grid resize is visible to every model
that references that grid. Cell reads return values, not mutable storage aliases.
There is no public generic grid or renderer framework.

- `NewStyledRuneGrid(w, h, options...)`: `Cell` contains a rune and Lip Gloss style.
- `NewRuneGrid(w, h)`: `RuneCell` contains a rune and `CellStyle`.
- `NewGraphemeGrid(w, h)`: `GraphemeCell` snapshots expose text, style, width,
  and whether a cell is a continuation of the preceding wide character.

All grids support `Bounds`, `Width`, `Height`, `Render`, `Resize`, `Clear`,
`Cell`, `SetText`, `SetTextWithStyle`, `SetLines`, `SetLinesWithStyle`, `Fill`,
`FillLine`, `FillRect`, `ClearRect`, `SetStyle`, `SetStyleRect`, the four
`Shift` methods, and same-type `CopyRect`. Cross-type conversion is explicit;
full Lip Gloss layout cannot be silently converted into terminal attributes.

`SetLines` and `SetLinesWithStyle` return no value. They clip extra lines to the
grid height and leave cells beyond each written line unchanged.

`Fill` and `FillLine` accept a `Cell` on StyledRuneGrid, a `RuneCell` on
RuneGrid, or `(text, CellStyle)` on GraphemeGrid. Grapheme fills return false for
an invalid cluster (or an invalid row for `FillLine`). Rune fills ignore invalid
rows. All shifts move one column/row and clear the exposed edge.

`CopyRect(dst, source, rect)` maps `rect.Min` to `dst`, clips to both grids, and
copies blank cells as well as content and style. Overlaps use a snapshot of the
clipped area. Temporary storage is proportional to the copied area. It does not
copy viewport, focus, or default writing style.

`SetStyle` and `SetStyleRect` replace styles; it does not merge attributes or change the default
`Style` for future writes. Styles on the full styled grid can retain callbacks.

## Text and layout trust

Use `GraphemeGrid` for untrusted text in composed layouts. Rune-grid sanitization
blocks terminal controls but does not preserve columns: wide characters,
combining marks, and shaping sequences can shift or truncate trusted content
beside the grid. This is a layout trust boundary, not just a visual caveat.

Both rune grids offer `SingleColumn = true`. During rendering it replaces each
nonblank rune whose Charm width is not one with U+FFFD, and sanitizes controls
(including raw styled text). Stored cells remain unchanged; the policy applies
to existing, filled, and copied cells too. It is opt-in and defaults to false.
Adjacent accepted runes that would merge into one grapheme are separated by
ZWNJ during rendering, using the same zero-width terminal assumption as
GraphemeGrid below. Boundaries reset at each rendered row and clipping edge.
It does not constrain Lip Gloss styles or guarantee width across terminal/font
policies. Prefer GraphemeGrid when preserving Unicode sequences matters.

All styles must come from the application, never untrusted text or configuration.
`SetCellStyle` can install arbitrary Lip Gloss behavior. `GetCellStyle` returns a
pointer to a copy of the style, not to grid storage; captured callbacks may still
share state. Neither rendering policy can restrict what a trusted style emits.

Legitimate Hebrew and Arabic text remains allowed. Terminals with bidi support
may visually reorder it; stored order is not a guarantee of displayed order.

## Styled rune grid

This stores one cell per rune and renders each cell with Lip Gloss.
Use `SetText`, `SetRunes`, `SetRune`, `SetCell`, `SetLines`, and their styled
variants. `WithStyle` and `WithContent` are grid options, applied in order.
Use `grid.Render(grid.Bounds())` for the whole grid or `Model.View()` for a viewport.

Writes sanitize text by default. `SetRawText` and `SetRawTextWithStyle` explicitly
preserve controls for trusted legacy content; compact grids have no raw mode.
`SanitizeText` replaces control/format runes and Unicode
line separators, including ESC, newlines, tabs, bidi controls, and joiners.
Styles are trusted: transforms, borders, padding, and arbitrary Unicode may expand
output beyond the grid's stored cells. Wide runes and combining marks still occupy
one stored cell each; clipping may split a grapheme.

## Compact styling and rune grid

```go
style := ntcanvas.CellStyle{
    Foreground: ntcanvas.RGBColor(80, 160, 240),
    Background: ntcanvas.IndexedColor(0),
    Attributes: ntcanvas.Bold | ntcanvas.Underline,
}
```

Zero colors select terminal defaults. `IndexedColor` supports the 256-color
palette; `RGBColor` emits true color. Attributes are bold, dim, italic, underline,
blink, reverse, and strikethrough. There are no layout rules or callbacks. The host
renderer/terminal controls color support. Equal styles render in runs.

`RuneGrid` offers `SetText`, `SetTextWithStyle`, `SetRune`, `SetRuneWithStyle`,
`SetCell`, and
`FillRect(rect, RuneCell)`. All writes neutralize controls as `SanitizeText` does.
Zero runes in explicit cells render as spaces; NUL in text becomes U+FFFD. Each rune still advances one stored cell regardless
of its terminal width. Text-write booleans mean a valid row, even if fully clipped.

## Grapheme grid

`SetText` and `SetTextWithStyle` segment text and advance by display columns using
the same Charm ANSI grapheme/width functions as the host dependency. Joined emoji
and combining sequences stay together. Controls and bidi formatting become
U+FFFD; ZWJ and ZWNJ remain available for shaping. Invalid UTF-8 becomes U+FFFD.
Clusters exceeding `MaxGraphemeRunes` (16 code points after sanitization) are
replaced wholesale with a single width-1 U+FFFD cell. This applies to text writes,
`SetGrapheme`, and fills, including overlong combining-mark and ZWJ chains.
Excess code points are not split into further cells. Normal standalone zero-width
clusters are discarded, never attached to earlier content. The limit bounds
stored/rendered cluster complexity, not input processing: sanitization and
segmentation still scan input. Applications should also bound incoming text size.
Short unsupported emoji sequences may still render differently on other fonts;
the cap does not establish universal terminal agreement.

`SetGrapheme(point, text, style)` accepts exactly one sanitized cluster of width
one or two. Invalid or partially clipped writes return false without changes.
`FillRect(rect, text, style)` tiles one such cluster; a spare column stays blank.

Wide-character rules:

- Overwriting either half clears the entire old cluster.
- Rendering a region that cuts a cluster emits blanks for its visible fragment.
- Copying a cut source cluster copies blanks. Destination clusters touched by a
  copy are cleared in full, possibly clearing a cell just outside the rectangle.
- Clearing or styling either half affects the whole cluster, including its other
  half outside the requested rectangle.
- Shrinking a grid clears a wide character cut by the new right edge.
- Shifts move one display column or row and clear the exposed edge. Horizontal
  shifts discard a wide character cut at either edge.

Separately written clusters can otherwise merge when rendered—for example two
regional indicators becoming one flag. The renderer inserts ZWNJ only at such
boundaries to retain their separate columns. Stored text remains unchanged.
This depends on the terminal treating U+200C as zero-width. The Ghostty screen
fixtures verify that behavior there; the library cannot detect a terminal that
renders ZWNJ visibly or assigns it a column.
Terminal/font Unicode policies may differ; no universal width guarantee is made.

## Clipped 2D lines

```go
ntcanvas.RasterLine(grid.Bounds(),
    ntcanvas.Float64Point{X: x, Y: y},
    ntcanvas.Float64Point{X: xx, Y: yy},
    func(p ntcanvas.Point) { grid.SetGrapheme(p, "·", style) })
```

`RasterLine(bounds, from, to, plot)` yields cell coordinates in caller endpoint
order. Convert projected float32 coordinates to float64 when constructing points.
The helper clips in floating-point cell coordinates to the **cell-center box**
`[Min.X, Max.X-1] × [Min.Y, Max.Y-1]`, then rounds the clipped endpoints to the
nearest integer, with half ties away from zero. Thus a degenerate point at
`X = -0.1` is outside a grid starting at zero, even though it would round to zero.

It walks an eight-connected Bresenham line, includes both rounded endpoints, and
visits each cell once. A visible degenerate segment visits one cell. When the
rasterizer's minor axis is exactly halfway, it stays nearer the starting endpoint;
reversing a segment can therefore select different tie cells. Clipping before
rounding/walking can also differ from clipping an already rasterized line.
This intentionally replaces the old capped float sampler's truncation and repeats.

Empty/inverted bounds, nonfinite coordinates, or a nil callback yield no cells.
Every callback point is inside bounds. Work depends on the clipped line's extent,
not endpoint distance. Ordinary clipping is allocation-free; extreme coordinates
and ambiguous rounding use exact rational arithmetic with bounded operand sizes.
The caller must use reasonable bounds; the helper imposes no separate cell budget.

The callback owns glyphs, styles, and ordering across segments. Projection, depth,
arrowhead construction, picking, and glyph/junction merging are outside this API.
The helper is independent of Surface and does not mutate a grid itself.

### Antialiased lines

`RasterLineAA(bounds, from, to, func(Point, float64))` emits Wu-style coverage.
The callback receives each cell at most once, with a weight in `(0, 1]`.
Up to two neighboring cells are emitted per major-axis step; zero weights and
out-of-bounds neighbors are omitted. Steps follow endpoint order, with increasing
minor coordinates on a forward walk and decreasing coordinates on a reverse walk.
Reversing the endpoints reverses the callback sequence without changing weights.

Endpoint caps reflect segment length: integer endpoints normally have half
coverage. Very short segments within one step combine into one pair of weights.
A visible degenerate point emits its nearest cell at full coverage. As with
RasterLine, clipping occurs at cell centers before walking. Clipped ends get new
endpoint caps; antialiasing does not extend the clipping box or renormalize weights
lost at its edges. This is a coverage approximation, not an exact area filter.

For a RuneGrid with an application-chosen black background, for example:

```go
ntcanvas.RasterLineAA(grid.Bounds(), from, to, func(p ntcanvas.Point, coverage float64) {
    level := uint8(math.Round(255 * coverage))
    grid.SetRuneWithStyle(p, '·', ntcanvas.CellStyle{
        Foreground: ntcanvas.RGBColor(level, level, level),
        Background: ntcanvas.RGBColor(0, 0, 0),
    })
})
```

Applications own blending against the actual background, gamma/color policy,
glyph choice, and accumulation across overlapping segments. Glyph shapes and
terminal color support affect the appearance. Coverage does not automatically
blend existing cells. Typical lines avoid allocations; the exact-arithmetic
fallback is slower and allocates, including during coverage calculation.

## Allocation

Each grid has a `New…Checked(w, h, SizeLimits)` constructor and `ResizeChecked`.
The styled constructor also accepts grid options. Invalid requests return errors;
a rejected resize leaves the grid unchanged. `ValidateSize` uses a conservative
allocation-overflow check sufficient for all three cell types.
Resizing to the current dimensions preserves storage without allocating or copying
cells. `ResizeChecked` still validates the requested limits before this no-op.

Zero limit fields default to width/height 4096 and 65536 cells. Limits apply per
call to one grid, not total process memory or retained text/style data. Resizing
and copying temporarily allocate extra storage. Unchecked constructors/Resize
clamp negative dimensions and panic on impossible allocation arithmetic. They
have no memory budget: a representable allocation can still exhaust memory
fatally. Use checked constructors **and `ResizeChecked`** for external dimensions,
including terminal/window resize messages, with application-appropriate limits.
