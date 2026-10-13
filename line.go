package ntcanvas

import (
	"image"
	"math"
	"math/big"
)

// RasterLine clips a floating-point segment and visits a connected line of cells
// in from-to order. It clips to cell centers [Min, Max-1], rounds clipped endpoints
// to the nearest integer (half ties away from zero), and uses Bresenham's rule.
// Minor-axis half ties stay nearer the start; reversing may choose different tie
// cells. Both rounded endpoints are included, with no duplicate cells. A visible
// degenerate segment visits one cell. Clipping precedes rounding and walking;
// the result need not equal clipping an already rasterized unbounded line.
//
// Empty/inverted bounds, nonfinite endpoints, and a nil plot produce no calls.
// Work is O(1 + max(visible width, visible height)), not offscreen distance. Bounds
// are the caller's work budget; use actual grid bounds rather than huge rectangles.
// The callback controls glyphs, styles, and writes; Surface is not required.
func RasterLine(bounds image.Rectangle, from, to Float64Point, plot func(Point)) {
	if plot == nil || bounds.Empty() || !finitePoint(from) || !finitePoint(to) {
		return
	}
	a, b, ok := clipLine(bounds, from, to)
	if !ok {
		return
	}
	walkLine(a, b, plot)
}

func finitePoint(p Float64Point) bool {
	return !math.IsNaN(p.X) && !math.IsNaN(p.Y) && !math.IsInf(p.X, 0) && !math.IsInf(p.Y, 0)
}

func clipLine(bounds image.Rectangle, a, b Float64Point) (Point, Point, bool) {
	start, end, visible, exact := clipLineFast(bounds, a, b)
	if exact {
		return clipLineExact(bounds, a, b)
	}
	if !visible {
		return Point{}, Point{}, false
	}
	point := func(p Float64Point) Point {
		return Point{X: min(max(int(math.Round(p.X)), bounds.Min.X), bounds.Max.X-1), Y: min(max(int(math.Round(p.Y)), bounds.Min.Y), bounds.Max.Y-1)}
	}
	return point(start), point(end), true
}

// clipLineFast returns exact=true when robust clipping needs rational arithmetic.
func clipLineFast(bounds image.Rectangle, a, b Float64Point) (start, end Float64Point, visible, exact bool) {
	// Ordinary terminal coordinates use allocation-free floating-point clipping.
	// Beyond this range, small viewport intersections may disappear through
	// cancellation. Exact rational clipping handles every finite float64 value.
	const limit = 1 << 26
	if bounds.Min.X < -limit || bounds.Min.Y < -limit || bounds.Max.X > limit || bounds.Max.Y > limit ||
		math.Abs(a.X) > limit || math.Abs(a.Y) > limit || math.Abs(b.X) > limit || math.Abs(b.Y) > limit {
		return Float64Point{}, Float64Point{}, false, true
	}
	inside := func(p Float64Point) bool {
		return p.X >= float64(bounds.Min.X) && p.X <= float64(bounds.Max.X-1) && p.Y >= float64(bounds.Min.Y) && p.Y <= float64(bounds.Max.Y-1)
	}
	if inside(a) && inside(b) {
		return a, b, true, false
	}
	lo, hi := 0.0, 1.0
	dx, dy := b.X-a.X, b.Y-a.Y
	for _, edge := range [4][2]float64{
		{-dx, a.X - float64(bounds.Min.X)}, {dx, float64(bounds.Max.X-1) - a.X},
		{-dy, a.Y - float64(bounds.Min.Y)}, {dy, float64(bounds.Max.Y-1) - a.Y},
	} {
		p, q := edge[0], edge[1]
		if p == 0 {
			if q < 0 {
				return Float64Point{}, Float64Point{}, false, false
			}
			continue
		}
		t := q / p
		if math.IsInf(t, 0) {
			return Float64Point{}, Float64Point{}, false, true
		}
		if p < 0 {
			lo = math.Max(lo, t)
		} else {
			hi = math.Min(hi, t)
		}
		if lo > hi {
			if lo-hi < 1e-14 {
				return Float64Point{}, Float64Point{}, false, true
			}
			return Float64Point{}, Float64Point{}, false, false
		}
	}
	// Clipping arithmetic can move a half-integer tie to either side. Fall back
	// when its error could affect rounding, or when a near-tangent intersection
	// could have collapsed to a point. Ordinary nonboundary lines stay allocation-free.
	if math.Abs(hi-lo) < 1e-14 {
		return Float64Point{}, Float64Point{}, false, true
	}
	start = Float64Point{math.FMA(lo, dx, a.X), math.FMA(lo, dy, a.Y)}
	end = Float64Point{math.FMA(hi, dx, a.X), math.FMA(hi, dy, a.Y)}
	tolerance := 1e-14 * max(1, math.Abs(a.X), math.Abs(a.Y), math.Abs(b.X), math.Abs(b.Y))
	for _, v := range [...]float64{start.X, start.Y, end.X, end.Y} {
		if math.Abs(v-(math.Floor(v)+0.5)) <= tolerance {
			return Float64Point{}, Float64Point{}, false, true
		}
	}
	return start, end, true, false
}

// Exact fallback avoids parameter collapse for e.g. [-MaxFloat64, MaxFloat64]
// crossing a small viewport. float64 input bounds the rational operand sizes.
func clipLineExact(bounds image.Rectangle, a, b Float64Point) (Point, Point, bool) {
	ax, ay, bx, by, ok := clipLineRational(bounds, a, b)
	if !ok {
		return Point{}, Point{}, false
	}
	return Point{X: roundRational(ax), Y: roundRational(ay)}, Point{X: roundRational(bx), Y: roundRational(by)}, true
}

func roundRational(v *big.Rat) int {
	var whole, rem big.Int
	whole.QuoRem(v.Num(), v.Denom(), &rem)
	rem.Abs(&rem).Lsh(&rem, 1)
	if rem.Cmp(v.Denom()) >= 0 {
		whole.Add(&whole, big.NewInt(int64(v.Sign())))
	}
	return int(whole.Int64()) // clipped cell-center coordinates fit in int
}

func clipLineRational(bounds image.Rectangle, a, b Float64Point) (*big.Rat, *big.Rat, *big.Rat, *big.Rat, bool) {
	ax, ay := new(big.Rat).SetFloat64(a.X), new(big.Rat).SetFloat64(a.Y)
	dx := new(big.Rat).Sub(new(big.Rat).SetFloat64(b.X), ax)
	dy := new(big.Rat).Sub(new(big.Rat).SetFloat64(b.Y), ay)
	lo, hi := new(big.Rat), big.NewRat(1, 1)
	for _, axis := range []struct {
		start, delta *big.Rat
		min, max     int
	}{
		{ax, dx, bounds.Min.X, bounds.Max.X - 1}, {ay, dy, bounds.Min.Y, bounds.Max.Y - 1},
	} {
		low := new(big.Rat).Sub(new(big.Rat).SetInt64(int64(axis.min)), axis.start)
		high := new(big.Rat).Sub(new(big.Rat).SetInt64(int64(axis.max)), axis.start)
		if axis.delta.Sign() == 0 {
			if low.Sign() > 0 || high.Sign() < 0 {
				return nil, nil, nil, nil, false
			}
			continue
		}
		low.Quo(low, axis.delta)
		high.Quo(high, axis.delta)
		if low.Cmp(high) > 0 {
			low, high = high, low
		}
		if low.Cmp(lo) > 0 {
			lo.Set(low)
		}
		if high.Cmp(hi) < 0 {
			hi.Set(high)
		}
		if lo.Cmp(hi) > 0 {
			return nil, nil, nil, nil, false
		}
	}

	coordinate := func(start, delta, t *big.Rat) *big.Rat { return new(big.Rat).Add(start, new(big.Rat).Mul(delta, t)) }
	return coordinate(ax, dx, lo), coordinate(ay, dy, lo), coordinate(ax, dx, hi), coordinate(ay, dy, hi), true
}

func walkLine(a, b Point, plot func(Point)) {
	// Unsigned differences and subtract-before-add errors avoid overflow even for
	// a rectangle spanning negative and positive machine-int coordinates.
	distance := func(a, b int) uint {
		if a < b {
			return uint(b) - uint(a)
		}
		return uint(a) - uint(b)
	}
	dx, dy := distance(a.X, b.X), distance(a.Y, b.Y)
	sx, sy := 1, 1
	if a.X > b.X {
		sx = -1
	}
	if a.Y > b.Y {
		sy = -1
	}
	plot(a)
	if dx >= dy {
		err := dx / 2
		for a.X != b.X {
			a.X += sx
			if err < dy {
				err += dx - dy
				a.Y += sy
			} else {
				err -= dy
			}
			plot(a)
		}
	} else {
		err := dy / 2
		for a.Y != b.Y {
			a.Y += sy
			if err < dx {
				err += dy - dx
				a.X += sx
			} else {
				err -= dx
			}
			plot(a)
		}
	}
}
