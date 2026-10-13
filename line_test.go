package ntcanvas

import (
	"image"
	"math"
	"reflect"
	"testing"
)

func linePoints(bounds image.Rectangle, a, b Float64Point) []Point {
	var points []Point
	RasterLine(bounds, a, b, func(p Point) { points = append(points, p) })
	return points
}

func TestRasterLine(t *testing.T) {
	bounds := image.Rect(0, 0, 5, 5)
	for _, tc := range []struct {
		name string
		a, b Float64Point
		want []Point
	}{
		{"horizontal", Float64Point{X: -100, Y: 2}, Float64Point{X: 100, Y: 2}, []Point{{X: 0, Y: 2}, {X: 1, Y: 2}, {X: 2, Y: 2}, {X: 3, Y: 2}, {X: 4, Y: 2}}},
		{"reverse", Float64Point{X: 100, Y: 2}, Float64Point{X: -100, Y: 2}, []Point{{X: 4, Y: 2}, {X: 3, Y: 2}, {X: 2, Y: 2}, {X: 1, Y: 2}, {X: 0, Y: 2}}},
		{"vertical", Float64Point{X: 2, Y: -100}, Float64Point{X: 2, Y: 100}, []Point{{X: 2, Y: 0}, {X: 2, Y: 1}, {X: 2, Y: 2}, {X: 2, Y: 3}, {X: 2, Y: 4}}},
		{"shallow", Float64Point{X: 0, Y: 0}, Float64Point{X: 4, Y: 2}, []Point{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 1}, {X: 3, Y: 1}, {X: 4, Y: 2}}},
		{"steep", Float64Point{X: 0, Y: 0}, Float64Point{X: 2, Y: 4}, []Point{{X: 0, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 2}, {X: 1, Y: 3}, {X: 2, Y: 4}}},
		{"negative slope", Float64Point{X: 0, Y: 4}, Float64Point{X: 4, Y: 0}, []Point{{X: 0, Y: 4}, {X: 1, Y: 3}, {X: 2, Y: 2}, {X: 3, Y: 1}, {X: 4, Y: 0}}},
		{"fractional", Float64Point{X: 0.5, Y: 1.5}, Float64Point{X: 3.4, Y: 1.5}, []Point{{X: 1, Y: 2}, {X: 2, Y: 2}, {X: 3, Y: 2}}},
		{"degenerate", Float64Point{X: 2.2, Y: 3.5}, Float64Point{X: 2.2, Y: 3.5}, []Point{{X: 2, Y: 4}}},
		{"outside degenerate", Float64Point{X: -0.1, Y: 2}, Float64Point{X: -0.1, Y: 2}, nil},
		{"parallel outside", Float64Point{X: -10, Y: 5}, Float64Point{X: 10, Y: 5}, nil},
		{"corner touch", Float64Point{X: -1, Y: 1}, Float64Point{X: 1, Y: -1}, []Point{{X: 0, Y: 0}}},
		{"extreme diagonal", Float64Point{X: -math.MaxFloat64, Y: -math.MaxFloat64}, Float64Point{X: math.MaxFloat64, Y: math.MaxFloat64}, []Point{{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 2, Y: 2}, {X: 3, Y: 3}, {X: 4, Y: 4}}},
		{"extreme horizontal", Float64Point{X: math.MaxFloat64, Y: 2}, Float64Point{X: -math.MaxFloat64, Y: 2}, []Point{{X: 4, Y: 2}, {X: 3, Y: 2}, {X: 2, Y: 2}, {X: 1, Y: 2}, {X: 0, Y: 2}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := linePoints(bounds, tc.a, tc.b); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("%v != %v", got, tc.want)
			}
		})
	}
}

func TestRasterLineOffsetAndInvalidBounds(t *testing.T) {
	bounds := image.Rect(-4, -4, 0, 0)
	got := linePoints(bounds, Float64Point{X: -3.5, Y: -2.5}, Float64Point{X: -1.5, Y: -2.5})
	if want := []Point{{X: -4, Y: -3}, {X: -3, Y: -3}, {X: -2, Y: -3}}; !reflect.DeepEqual(got, want) {
		t.Fatal(got)
	}
	for _, r := range []image.Rectangle{{}, image.Rect(0, 0, 0, 5), {Min: Point{X: 5, Y: 5}, Max: Point{X: 1, Y: 1}}} {
		RasterLine(r, Float64Point{}, Float64Point{X: 1, Y: 1}, func(Point) { t.Fatal("invalid bounds") })
	}
	for _, v := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		for _, pair := range [][2]Float64Point{{{X: v, Y: 0}, {}}, {{X: 0, Y: v}, {}}, {{}, {X: v, Y: 0}}, {{}, {X: 0, Y: v}}} {
			RasterLine(image.Rect(0, 0, 5, 5), pair[0], pair[1], func(Point) { t.Fatal("nonfinite point") })
		}
	}
	RasterLine(bounds, Float64Point{}, Float64Point{}, nil)
	// Only one row/column is visible, even with distant diagonal endpoints.
	for _, r := range []image.Rectangle{image.Rect(2, 0, 3, 5), image.Rect(0, 2, 5, 3), image.Rect(2, 2, 3, 3)} {
		got := linePoints(r, Float64Point{X: -1e100, Y: -1e100}, Float64Point{X: 1e100, Y: 1e100})
		if !reflect.DeepEqual(got, []Point{{X: 2, Y: 2}}) {
			t.Fatal(got)
		}
	}
	// Bounds at machine-int edges must neither overflow nor round outside bounds.
	maxInt := int(^uint(0) >> 1)
	for _, r := range []image.Rectangle{image.Rect(maxInt-3, 0, maxInt, 1), image.Rect(-maxInt-1, 0, -maxInt+2, 1)} {
		got := linePoints(r, Float64Point{X: -math.MaxFloat64, Y: 0}, Float64Point{X: math.MaxFloat64, Y: 0})
		want := []Point{{X: r.Min.X, Y: 0}, {X: r.Min.X + 1, Y: 0}, {X: r.Min.X + 2, Y: 0}}
		if !reflect.DeepEqual(got, want) {
			t.Fatal(got, want)
		}
	}
}

func FuzzRasterLine(f *testing.F) {
	for _, v := range [][4]float64{{-100, 2, 100, 2}, {-math.MaxFloat64, -math.MaxFloat64, math.MaxFloat64, math.MaxFloat64}, {0.5, 1.5, 3.4, 1.5}, {0, 0, 4, 2}} {
		f.Add(v[0], v[1], v[2], v[3])
	}
	f.Fuzz(func(t *testing.T, x, y, xx, yy float64) {
		bounds := image.Rect(-5, -3, 6, 4)
		a, b := Float64Point{X: x, Y: y}, Float64Point{X: xx, Y: yy}
		var last Point
		count := 0
		RasterLine(bounds, a, b, func(p Point) {
			count++
			if count > 11 || !p.In(bounds) {
				t.Fatal("unbounded or out of bounds", p, count)
			}
			if count > 1 && (p == last || p.X-last.X < -1 || p.X-last.X > 1 || p.Y-last.Y < -1 || p.Y-last.Y > 1) {
				t.Fatal("duplicate or disconnected", last, p)
			}
			last = p
		})
		if finitePoint(a) && finitePoint(b) {
			p, q, ok := clipLine(bounds, a, b)
			rp, rq, rok := clipLineExact(bounds, a, b)
			if ok != rok || ok && (p != rp || q != rq) {
				t.Fatalf("clip mismatch: %v %v %v vs %v %v %v; %v %v", p, q, ok, rp, rq, rok, a, b)
			}
		} else if count != 0 {
			t.Fatal("nonfinite input")
		}
	})
}

func BenchmarkRasterLine(b *testing.B) {
	for _, tc := range []struct {
		name string
		a, z Float64Point
	}{
		{"visible", Float64Point{X: 0, Y: 0}, Float64Point{X: 79, Y: 23}},
		{"clipped", Float64Point{X: -1000, Y: -100}, Float64Point{X: 1000, Y: 120}},
		{"extreme", Float64Point{X: -math.MaxFloat64, Y: 10}, Float64Point{X: math.MaxFloat64, Y: 10}},
	} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			count := 0
			for b.Loop() {
				RasterLine(image.Rect(0, 0, 80, 24), tc.a, tc.z, func(Point) { count++ })
			}
			if count == 0 {
				b.Fatal("no output")
			}
		})
	}
}
