package ntcanvas_test

import (
	"fmt"
	"image"
	"nimbleterminal.dev/ntcanvas"
)

func ExampleRasterLine() {
	grid := ntcanvas.NewGraphemeGrid(4, 3)
	ntcanvas.RasterLine(grid.Bounds(),
		ntcanvas.Float64Point{X: -100, Y: -100},
		ntcanvas.Float64Point{X: 100, Y: 100},
		func(p ntcanvas.Point) { grid.SetGrapheme(p, "·", ntcanvas.CellStyle{}) })
	fmt.Printf("%q\n", grid.Render(grid.Bounds()))
	// Output: "·   \n ·  \n  · "
}

func ExampleRasterLineAA() {
	ntcanvas.RasterLineAA(image.Rect(0, 0, 3, 2),
		ntcanvas.Float64Point{X: 0, Y: 0.5},
		ntcanvas.Float64Point{X: 2, Y: 0.5},
		func(p ntcanvas.Point, coverage float64) {
			// Use coverage to blend application colors or select glyph intensity.
			fmt.Printf("(%d,%d): %.2f\n", p.X, p.Y, coverage)
		})
	// Output:
	// (0,0): 0.25
	// (0,1): 0.25
	// (1,0): 0.50
	// (1,1): 0.50
	// (2,0): 0.25
	// (2,1): 0.25
}
