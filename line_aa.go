package ntcanvas

import (
	"image"
	"math"
	"math/big"
)

// RasterLineAA clips a segment to cell centers and emits Wu-style coverage in
// (0,1] for up to two neighboring cells per major-axis step. Zero weights and
// out-of-bounds neighbors are omitted; each cell is emitted at most once.
// Steps follow from-to order. Within a step, minor coordinates increase for a
// forward walk and decrease for a reverse walk. Reversing reverses the output.
//
// Coverage includes endpoint caps: integer endpoints normally have half weight.
// Short segments within one step are combined, not emitted twice. A degenerate
// visible point emits its nearest cell (half ties away from zero) at coverage 1.
// Coverage is a geometric weight, not a color/alpha or terminal rendering promise.
// The caller chooses blending, glyph intensity, and handling of overlapping lines.
//
// Bounds, invalid input, work limits, and exact clipping fallback follow RasterLine.
// Clipping precedes antialiasing; coverage outside the cell-center box is lost,
// and clipping creates new endpoint caps. This is not an infinite-line area filter.
func RasterLineAA(bounds image.Rectangle, from, to Float64Point, plot func(Point, float64)) {
	if plot == nil || bounds.Empty() || !finitePoint(from) || !finitePoint(to) {
		return
	}
	// Canonicalize before floating-point clipping so reversal cannot move a
	// near-integer sample to the opposite neighboring cell through roundoff.
	reversedInput := from.X > to.X || from.X == to.X && from.Y > to.Y
	if reversedInput {
		from, to = to, from
	}
	a, b, visible, exact := clipLineFast(bounds, from, to)
	if exact {
		if reversedInput {
			from, to = to, from
		}
		rasterLineAAExact(bounds, from, to, plot)
		return
	}
	if !visible {
		return
	}
	if reversedInput {
		a, b = b, a
	}
	rasterLineAAFloat(bounds, a, b, plot)
}

func rasterLineAAFloat(bounds image.Rectangle, a, b Float64Point, plot func(Point, float64)) {
	if a == b {
		plot(NewPointFromFloat64Point(a), 1)
		return
	}
	steep := math.Abs(b.Y-a.Y) > math.Abs(b.X-a.X)
	if steep {
		a.X, a.Y = a.Y, a.X
		b.X, b.Y = b.Y, b.X
	}
	reverse := a.X > b.X
	if reverse {
		a, b = b, a
	}
	gradient := (b.Y - a.Y) / (b.X - a.X)
	first, last := int(math.Floor(a.X+0.5)), int(math.Floor(b.X+0.5))
	step := 1
	if reverse {
		first, last = last, first
		step = -1
	}
	emit := func(x, y int, weight float64) {
		p := Point{X: x, Y: y}
		if steep {
			p.X, p.Y = p.Y, p.X
		}
		if weight > 0 && p.In(bounds) {
			plot(p, min(1, weight))
		}
	}
	for x := first; ; x += step {
		// Intersection of the segment's major-axis extent with this cell's footprint.
		cap := min(b.X, float64(x)+0.5) - max(a.X, float64(x)-0.5)
		if cap > 0 {
			y := math.FMA(float64(x)-a.X, gradient, a.Y)
			low := math.Floor(y)
			fraction := y - low
			if reverse {
				emit(x, int(low)+1, cap*fraction)
				emit(x, int(low), cap*(1-fraction))
			} else {
				emit(x, int(low), cap*(1-fraction))
				emit(x, int(low)+1, cap*fraction)
			}
		}
		if x == last {
			break
		}
	}
}

// The exact path keeps absolute coordinates as integers/rationals, so bounds
// near machine-int limits still produce distinct cells and meaningful fractions.
func rasterLineAAExact(bounds image.Rectangle, from, to Float64Point, plot func(Point, float64)) {
	ax, ay, bx, by, ok := clipLineRational(bounds, from, to)
	if !ok {
		return
	}
	dx, dy := new(big.Rat).Sub(bx, ax), new(big.Rat).Sub(by, ay)
	if dx.Sign() == 0 && dy.Sign() == 0 {
		plot(Point{X: roundRational(ax), Y: roundRational(ay)}, 1)
		return
	}
	// Once extreme input has been clipped into ordinary coordinates, use the
	// allocation-free coverage walker. Keep rationals only when absolute integer
	// coordinates or a collapsed float conversion require their extra precision.
	fx, fy := rationalFloat(ax), rationalFloat(ay)
	tx, ty := rationalFloat(bx), rationalFloat(by)
	if max(math.Abs(fx), math.Abs(fy), math.Abs(tx), math.Abs(ty)) <= 1<<26 && (fx != tx || fy != ty) {
		rasterLineAAFloat(bounds, Float64Point{fx, fy}, Float64Point{tx, ty}, plot)
		return
	}
	steep := new(big.Rat).Abs(dy).Cmp(new(big.Rat).Abs(dx)) > 0
	if steep {
		ax, ay = ay, ax
		bx, by = by, bx
		dx, dy = dy, dx
	}
	reverse := ax.Cmp(bx) > 0
	if reverse {
		ax, bx = bx, ax
		ay = by
	} // dy/dx is unchanged
	gradient := new(big.Rat).Quo(dy, dx)
	half, one := big.NewRat(1, 2), big.NewRat(1, 1)
	nearest := func(v *big.Rat) int {
		shifted := new(big.Rat).Add(v, half)
		return int(new(big.Int).Div(shifted.Num(), shifted.Denom()).Int64())
	}
	first, last := nearest(ax), nearest(bx)
	step := 1
	if reverse {
		first, last = last, first
		step = -1
	}
	minMinor, maxMinor := bounds.Min.Y, bounds.Max.Y
	if steep {
		minMinor, maxMinor = bounds.Min.X, bounds.Max.X
	}
	lowBound, highBound := big.NewInt(int64(minMinor)), big.NewInt(int64(maxMinor))
	emit := func(x int, y *big.Int, weight *big.Rat) {
		if weight.Sign() <= 0 || y.Cmp(lowBound) < 0 || y.Cmp(highBound) >= 0 {
			return
		}
		coverage, _ := weight.Float64()
		if coverage == 0 {
			return
		} // an unrepresentably small positive weight
		p := Point{X: x, Y: int(y.Int64())}
		if steep {
			p.X, p.Y = p.Y, p.X
		}
		plot(p, min(1, coverage))
	}
	for x := first; ; x += step {
		center := new(big.Rat).SetInt64(int64(x))
		left, right := new(big.Rat).Sub(center, half), new(big.Rat).Add(center, half)
		if left.Cmp(ax) < 0 {
			left = ax
		}
		if right.Cmp(bx) > 0 {
			right = bx
		}
		cap := new(big.Rat).Sub(right, left)
		if cap.Sign() > 0 {
			y := new(big.Rat).Add(ay, new(big.Rat).Mul(gradient, new(big.Rat).Sub(center, ax)))
			low := new(big.Int).Div(y.Num(), y.Denom())
			high := new(big.Int).Add(low, big.NewInt(1))
			fraction := new(big.Rat).Sub(y, new(big.Rat).SetInt(low))
			lower := new(big.Rat).Mul(cap, new(big.Rat).Sub(one, fraction))
			upper := new(big.Rat).Mul(cap, fraction)
			if reverse {
				emit(x, high, upper)
				emit(x, low, lower)
			} else {
				emit(x, low, lower)
				emit(x, high, upper)
			}
		}
		if x == last {
			break
		}
	}
}

func rationalFloat(v *big.Rat) float64 { f, _ := v.Float64(); return f }
