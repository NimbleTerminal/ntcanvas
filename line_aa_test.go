package ntcanvas

import (
	"image"
	"math"
	"testing"
)

type coveredCell struct {
	p      Point
	weight float64
}

func coveredLine(bounds image.Rectangle, a, b Float64Point) []coveredCell {
	var cells []coveredCell
	RasterLineAA(bounds, a, b, func(p Point, c float64) { cells = append(cells, coveredCell{p, c}) })
	return cells
}
func equalCoverage(t *testing.T, got, want []coveredCell) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%v != %v", got, want)
	}
	for i := range got {
		if got[i].p != want[i].p || math.Abs(got[i].weight-want[i].weight) > 1e-12 {
			t.Fatalf("%v != %v", got, want)
		}
	}
}
func TestRasterLineAACoverage(t *testing.T) {
	for _, tc := range []struct {
		name string
		a, b Float64Point
		want []coveredCell
	}{
		{"horizontal", Float64Point{0, 1}, Float64Point{3, 1}, []coveredCell{{Point{X: 0, Y: 1}, .5}, {Point{X: 1, Y: 1}, 1}, {Point{X: 2, Y: 1}, 1}, {Point{X: 3, Y: 1}, .5}}},
		{"shallow", Float64Point{0, 0}, Float64Point{4, 2}, []coveredCell{{Point{}, .5}, {Point{X: 1}, .5}, {Point{X: 1, Y: 1}, .5}, {Point{X: 2, Y: 1}, 1}, {Point{X: 3, Y: 1}, .5}, {Point{X: 3, Y: 2}, .5}, {Point{X: 4, Y: 2}, .5}}},
		{"half row", Float64Point{0, .5}, Float64Point{2, .5}, []coveredCell{{Point{}, .25}, {Point{Y: 1}, .25}, {Point{X: 1}, .5}, {Point{X: 1, Y: 1}, .5}, {Point{X: 2}, .25}, {Point{X: 2, Y: 1}, .25}}},
		{"short", Float64Point{1.1, 1.2}, Float64Point{1.3, 1.2}, []coveredCell{{Point{X: 1, Y: 1}, .16}, {Point{X: 1, Y: 2}, .04}}},
		{"half endpoints", Float64Point{.5, 1}, Float64Point{1.5, 1}, []coveredCell{{Point{X: 1, Y: 1}, 1}}},
		{"point", Float64Point{1.5, 1.5}, Float64Point{1.5, 1.5}, []coveredCell{{Point{X: 2, Y: 2}, 1}}},
		{"corner touch", Float64Point{-1, 1}, Float64Point{1, -1}, []coveredCell{{Point{}, 1}}},
		{"extreme", Float64Point{-math.MaxFloat64, 1}, Float64Point{math.MaxFloat64, 1}, []coveredCell{{Point{Y: 1}, .5}, {Point{X: 1, Y: 1}, 1}, {Point{X: 2, Y: 1}, 1}, {Point{X: 3, Y: 1}, 1}, {Point{X: 4, Y: 1}, .5}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bounds := image.Rect(0, 0, 5, 5)
			equalCoverage(t, coveredLine(bounds, tc.a, tc.b), tc.want)
			rev := append([]coveredCell(nil), tc.want...)
			for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
				rev[i], rev[j] = rev[j], rev[i]
			}
			equalCoverage(t, coveredLine(bounds, tc.b, tc.a), rev)
			// Transposing covers steep and vertical lines with the same weights.
			for i := range tc.want {
				tc.want[i].p.X, tc.want[i].p.Y = tc.want[i].p.Y, tc.want[i].p.X
			}
			equalCoverage(t, coveredLine(bounds, Float64Point{tc.a.Y, tc.a.X}, Float64Point{tc.b.Y, tc.b.X}), tc.want)
		})
	}
}
func TestRasterLineAAInvalidAndOffset(t *testing.T) {
	for _, r := range []image.Rectangle{{}, {Min: Point{X: 3, Y: 3}, Max: Point{X: 1, Y: 1}}} {
		RasterLineAA(r, Float64Point{}, Float64Point{1, 1}, func(Point, float64) { t.Fatal("invalid bounds") })
	}
	bounds := image.Rect(-5, -5, 5, 5)
	RasterLineAA(bounds, Float64Point{}, Float64Point{1, 1}, nil)
	for _, v := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		for _, pair := range [][2]Float64Point{{{v, 0}, {}}, {{0, v}, {}}, {{}, {v, 0}}, {{}, {0, v}}} {
			if len(coveredLine(bounds, pair[0], pair[1])) != 0 {
				t.Fatal("nonfinite")
			}
		}
	}
	equalCoverage(t, coveredLine(bounds, Float64Point{-2.5, -3.5}, Float64Point{-2.5, -3.5}), []coveredCell{{Point{X: -3, Y: -4}, 1}})
	maxInt := int(^uint(0) >> 1)
	for _, r := range []image.Rectangle{image.Rect(maxInt-3, 0, maxInt, 1), image.Rect(-maxInt-1, 0, -maxInt+2, 1)} {
		equalCoverage(t, coveredLine(r, Float64Point{-math.MaxFloat64, 0}, Float64Point{math.MaxFloat64, 0}), []coveredCell{{Point{X: r.Min.X}, .5}, {Point{X: r.Min.X + 1}, 1}, {Point{X: r.Min.X + 2}, .5}})
	}
	if len(coveredLine(bounds, Float64Point{-100, 20}, Float64Point{100, 20})) != 0 {
		t.Fatal("offscreen line")
	}
}
func FuzzRasterLineAA(f *testing.F) {
	for _, v := range [][4]float64{{0, 0, 4, 2}, {.1, .2, .3, .2}, {-100, 2, 100, 2}, {-math.MaxFloat64, -math.MaxFloat64, math.MaxFloat64, math.MaxFloat64}} {
		f.Add(v[0], v[1], v[2], v[3])
	}
	f.Fuzz(func(t *testing.T, x, y, xx, yy float64) {
		bounds := image.Rect(-5, -3, 6, 4)
		a, b := Float64Point{x, y}, Float64Point{xx, yy}
		got := coveredLine(bounds, a, b)
		if len(got) > 22 {
			t.Fatal("unbounded work")
		}
		seen := map[Point]bool{}
		for _, c := range got {
			if seen[c.p] || !c.p.In(bounds) || math.IsNaN(c.weight) || c.weight <= 0 || c.weight > 1 {
				t.Fatal("invalid coverage", c)
			}
			seen[c.p] = true
		}
		if !finitePoint(a) || !finitePoint(b) {
			if len(got) != 0 {
				t.Fatal("nonfinite")
			}
			return
		}
		reverse := coveredLine(bounds, b, a)
		for i, j := 0, len(reverse)-1; i < j; i, j = i+1, j-1 {
			reverse[i], reverse[j] = reverse[j], reverse[i]
		}
		equalCoverage(t, got, reverse)
	})
}
func BenchmarkRasterLineAA(b *testing.B) {
	for _, tc := range []struct {
		name string
		a, z Float64Point
	}{
		{"visible", Float64Point{0, 0}, Float64Point{79, 23}},
		{"clipped", Float64Point{-1000, -100}, Float64Point{1000, 120}},
		{"extreme", Float64Point{-math.MaxFloat64, 10}, Float64Point{math.MaxFloat64, 10}},
	} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				RasterLineAA(image.Rect(0, 0, 80, 24), tc.a, tc.z, func(Point, float64) {})
			}
		})
	}
}
