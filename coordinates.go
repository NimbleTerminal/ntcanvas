// ntcharts - Copyright (c) 2024 Neomantra Corp.

package ntcanvas

import (
	"image"
	"math"
)

// Point is alias for image.Point
type Point = image.Point

// Float64Point represents a point in a coordinate system
// with floating point precision.
type Float64Point struct {
	X float64
	Y float64
}

// NewFloat64PointFromPoint returns a new Float64Point from a given Point.
func NewFloat64PointFromPoint(p Point) Float64Point {
	return Float64Point{X: float64(p.X), Y: float64(p.Y)}
}

// Mul returns a Float64Point with both X and Y values
// multiplied by the X and Y values of the given Float64Point.
func (p Float64Point) Mul(f Float64Point) Float64Point {
	return Float64Point{p.X * f.X, p.Y * f.Y}
}

// Add returns a Float64Point with both X and Y values
// added to the X and Y values of the given Float64Point.
func (p Float64Point) Add(f Float64Point) Float64Point {
	return Float64Point{p.X + f.X, p.Y + f.Y}
}

// Sub returns a Float64Point with both X and Y values
// subtracted by the X and Y values of the given Float64Point.
func (p Float64Point) Sub(f Float64Point) Float64Point {
	return Float64Point{p.X - f.X, p.Y - f.Y}
}

// NewPointFromFloat64Point returns a new Point from a given Float64Point.
func NewPointFromFloat64Point(f Float64Point) Point {
	return Point{X: int(math.Round(f.X)), Y: int(math.Round(f.Y))}
}

// CanvasYCoordinates returns a sequence of Y coordinates in the
// canvas coordinates system (X,Y is top left) from a given sequence of Y coordinates
// in the Cartesian coordinates system (X,Y is bottom left)
// by passing the graph X axis in the canvas coordinates system.
func CanvasYCoordinates(xAxis int, seq []int) (r []int) {
	r = make([]int, 0, len(seq))
	for _, v := range seq {
		r = append(r, CanvasYCoordinate(xAxis, v))
	}
	return
}

// CanvasYCoordinate returns a Y coordinates in the
// canvas coordinates system (X,Y is top left) from a Y coordinate
// in the Cartesian coordinates system (X,Y is bottom left)
// by passing the graph X axis in the canvas coordinates system.
func CanvasYCoordinate(xAxis int, y int) (r int) {
	return xAxis - y
}

// CanvasPoints returns a sequence of Points in the
// canvas coordinates system (X,Y is top left) from a given sequence of Points
// in the Cartesian coordinates system (X,Y is bottom left)
// by passing the graph origin in the canvas coordinates system.
func CanvasPoints(origin Point, seq []Point) (r []Point) {
	r = make([]Point, 0, len(seq))
	for _, v := range seq {
		r = append(r, CanvasPoint(origin, v))
	}
	return
}

// CanvasPoint returns a Point in the
// canvas coordinates system (X,Y is top left) from a given Point
// in the Cartesian coordinates system (X,Y is bottom left)
// by passing the graph origin in the canvas coordinates system.
func CanvasPoint(origin Point, p Point) (r Point) {
	return Point{X: origin.X + p.X, Y: origin.Y - p.Y}
}

// CanvasPointFromFloat64Point returns a Point
// in the canvas coordinates systems (X,Y is top left)
// from a canvas Float64Point in the
// Cartesian coordinates system (X,Y is bottom left)
// by passing the graph origin in the canvas coordinates system.
func CanvasPointFromFloat64Point(origin Point, f Float64Point) Point {
	// round coordinates to nearest integer and
	// convert Cartesian coordinates to canvas coordinates
	return CanvasPoint(origin, NewPointFromFloat64Point(f))
}

// CanvasFloat64Point returns a Float64Point in the
// canvas coordinates system (X,Y is top left) from a given Float64Point
// in the Cartesian coordinates system (X,Y is bottom left)
// by passing the graph origin in the canvas coordinates system.
func CanvasFloat64Point(origin Point, p Float64Point) (r Float64Point) {
	return Float64Point{float64(origin.X) + p.X, float64(origin.Y) - p.Y}
}
