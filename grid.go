package ntcanvas

import (
	"image"

	ntcanvasgrid "nimbleterminal.dev/ntcanvas/grid"
	"nimbleterminal.dev/ntcanvas/internal/cellstore"
)

// The grids, their cell style, points, and the line rasterizers live in
// package ntcanvasgrid, which links none of the terminal UI stack. They are
// re-exported here under the names they have always had: each type below is
// the same type as its ntcanvasgrid counterpart, methods and all.

type (
	// Point is alias for image.Point
	Point = ntcanvasgrid.Point
	// Float64Point represents a point in a coordinate system with floating
	// point precision.
	Float64Point = ntcanvasgrid.Float64Point
	// Color is a terminal color. Its zero value selects the terminal default.
	// Construct nondefault colors with IndexedColor or RGBColor.
	Color = ntcanvasgrid.Color
	// Attributes are terminal SGR flags, not layout or text transformations.
	Attributes = ntcanvasgrid.Attributes
	// CellStyle is a compact, comparable style shared by RuneGrid and GraphemeGrid.
	CellStyle = ntcanvasgrid.CellStyle
	// RuneCell is one RuneGrid cell: a rune and its style.
	RuneCell = ntcanvasgrid.RuneCell
	// RuneGrid stores one rune per terminal cell with a compact CellStyle.
	RuneGrid = ntcanvasgrid.RuneGrid
	// GraphemeCell is a read-only snapshot returned by GraphemeGrid.Cell.
	GraphemeCell = ntcanvasgrid.GraphemeCell
	// GraphemeGrid stores grapheme clusters in terminal columns.
	GraphemeGrid = ntcanvasgrid.GraphemeGrid
	// SizeLimits bounds a checked allocation; zero fields use defaults.
	SizeLimits = ntcanvasgrid.SizeLimits
)

const (
	// Bold requests increased text intensity.
	Bold = ntcanvasgrid.Bold
	// Dim requests reduced text intensity.
	Dim = ntcanvasgrid.Dim
	// Italic requests italic text.
	Italic = ntcanvasgrid.Italic
	// Underline requests underlined text.
	Underline = ntcanvasgrid.Underline
	// Blink requests blinking text where supported by the terminal.
	Blink = ntcanvasgrid.Blink
	// Reverse swaps foreground and background colors.
	Reverse = ntcanvasgrid.Reverse
	// Strikethrough requests a line through the text.
	Strikethrough = ntcanvasgrid.Strikethrough

	// MaxGraphemeRunes bounds stored and rendered cluster complexity.
	MaxGraphemeRunes = ntcanvasgrid.MaxGraphemeRunes
)

var (
	// ErrInvalidSize indicates negative dimensions, invalid limits, or integer overflow.
	ErrInvalidSize = ntcanvasgrid.ErrInvalidSize
	// ErrSizeLimit indicates a request exceeds a caller's allocation limits.
	ErrSizeLimit = ntcanvasgrid.ErrSizeLimit
)

// ValidateSize checks dimensions without allocating. It also rejects sizes whose
// cell or row storage would overflow an int, including on 32-bit/WASM targets.
// It allows for the largest cell of any grid, a StyledRuneGrid's.
func ValidateSize(w, h int, limits SizeLimits) error {
	return cellstore.ValidateSize[Cell](w, h, cellstore.Limits(limits))
}

// IndexedColor selects an entry in the terminal's 256-color palette.
func IndexedColor(index uint8) Color { return ntcanvasgrid.IndexedColor(index) }

// RGBColor creates a true-color terminal color from red, green, and blue bytes.
func RGBColor(r, g, b uint8) Color { return ntcanvasgrid.RGBColor(r, g, b) }

// NewRuneGrid allocates an uncapped grid; see ntcanvasgrid.NewRuneGrid.
func NewRuneGrid(w, h int) *RuneGrid { return ntcanvasgrid.NewRuneGrid(w, h) }

// NewRuneGridChecked validates dimensions against limits before allocating.
func NewRuneGridChecked(w, h int, l SizeLimits) (*RuneGrid, error) {
	if err := ValidateSize(w, h, l); err != nil {
		return nil, err
	}
	return ntcanvasgrid.NewRuneGrid(w, h), nil
}

// NewGraphemeGrid allocates an uncapped grid measured in terminal columns; see
// ntcanvasgrid.NewGraphemeGrid.
func NewGraphemeGrid(w, h int) *GraphemeGrid { return ntcanvasgrid.NewGraphemeGrid(w, h) }

// NewGraphemeGridChecked validates dimensions against limits before allocating.
func NewGraphemeGridChecked(w, h int, l SizeLimits) (*GraphemeGrid, error) {
	if err := ValidateSize(w, h, l); err != nil {
		return nil, err
	}
	return ntcanvasgrid.NewGraphemeGrid(w, h), nil
}

// SanitizeText replaces control, format, and line/paragraph separator runes with
// U+FFFD; see ntcanvasgrid.SanitizeText.
func SanitizeText(s string) string { return ntcanvasgrid.SanitizeText(s) }

// RasterLine visits the cells of the line from one point to another, clipped
// to bounds; see ntcanvasgrid.RasterLine.
func RasterLine(bounds image.Rectangle, from, to Float64Point, plot func(Point)) {
	ntcanvasgrid.RasterLine(bounds, from, to, plot)
}

// RasterLineAA is RasterLine with a coverage for each cell; see
// ntcanvasgrid.RasterLineAA.
func RasterLineAA(bounds image.Rectangle, from, to Float64Point, plot func(Point, float64)) {
	ntcanvasgrid.RasterLineAA(bounds, from, to, plot)
}

// NewFloat64PointFromPoint returns a new Float64Point from a given Point.
func NewFloat64PointFromPoint(p Point) Float64Point { return ntcanvasgrid.NewFloat64PointFromPoint(p) }

// NewPointFromFloat64Point returns a new Point from a given Float64Point.
func NewPointFromFloat64Point(f Float64Point) Point { return ntcanvasgrid.NewPointFromFloat64Point(f) }

// CanvasYCoordinates converts Cartesian Y coordinates to canvas rows; see
// ntcanvasgrid.CanvasYCoordinates.
func CanvasYCoordinates(xAxis int, seq []int) []int {
	return ntcanvasgrid.CanvasYCoordinates(xAxis, seq)
}

// CanvasYCoordinate converts a Cartesian Y coordinate to a canvas row; see
// ntcanvasgrid.CanvasYCoordinate.
func CanvasYCoordinate(xAxis int, y int) int { return ntcanvasgrid.CanvasYCoordinate(xAxis, y) }

// CanvasPoints converts Cartesian points to canvas points; see
// ntcanvasgrid.CanvasPoints.
func CanvasPoints(origin Point, seq []Point) []Point { return ntcanvasgrid.CanvasPoints(origin, seq) }

// CanvasPoint converts a Cartesian point to a canvas point; see
// ntcanvasgrid.CanvasPoint.
func CanvasPoint(origin Point, p Point) Point { return ntcanvasgrid.CanvasPoint(origin, p) }

// CanvasPointFromFloat64Point converts a Cartesian Float64Point to a canvas
// point; see ntcanvasgrid.CanvasPointFromFloat64Point.
func CanvasPointFromFloat64Point(origin Point, f Float64Point) Point {
	return ntcanvasgrid.CanvasPointFromFloat64Point(origin, f)
}

// CanvasFloat64Point converts a Cartesian Float64Point to a canvas
// Float64Point; see ntcanvasgrid.CanvasFloat64Point.
func CanvasFloat64Point(origin Point, p Float64Point) Float64Point {
	return ntcanvasgrid.CanvasFloat64Point(origin, p)
}
