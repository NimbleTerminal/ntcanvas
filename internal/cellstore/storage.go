// Package cellstore holds what ntcanvas's grids share and keep to themselves:
// rectangular cell storage, the size checks that guard it, and the rules for
// which runes may be stored and which cells would merge when written side by
// side. It imports nothing from the terminal UI stack, so the grid package
// built on it stays free of it too.
package cellstore

import (
	"errors"
	"image"
	"unsafe"
)

var (
	// ErrInvalidSize indicates negative dimensions, invalid limits, or integer overflow.
	ErrInvalidSize = errors.New("canvas: invalid size")
	// ErrSizeLimit indicates a request exceeds a caller's allocation limits.
	ErrSizeLimit = errors.New("canvas: size limit exceeded")
)

// Limits bounds a checked allocation; see the grids' SizeLimits.
type Limits struct{ MaxWidth, MaxHeight, MaxCells int }

// ValidateSize checks dimensions for a grid of T cells without allocating.
func ValidateSize[T any](w, h int, limits Limits) error {
	if limits.MaxWidth < 0 || limits.MaxHeight < 0 || limits.MaxCells < 0 || !ValidAllocation[T](w, h) {
		return ErrInvalidSize
	}
	if limits.MaxWidth == 0 {
		limits.MaxWidth = 4096
	}
	if limits.MaxHeight == 0 {
		limits.MaxHeight = 4096
	}
	if limits.MaxCells == 0 {
		limits.MaxCells = 65536
	}
	if w > limits.MaxWidth || h > limits.MaxHeight || (h > 0 && w > limits.MaxCells/h) {
		return ErrSizeLimit
	}
	return nil
}

// ValidAllocation reports whether w by h cells of T, and their rows, can be
// addressed without overflowing an int, including on 32-bit and WASM targets.
func ValidAllocation[T any](w, h int) bool {
	if w < 0 || h < 0 {
		return false
	}
	var cell T
	maxInt := int(^uint(0) >> 1)
	rowBytes := int(unsafe.Sizeof([]T(nil)))
	cellBytes := max(1, int(unsafe.Sizeof(cell)))
	if h > maxInt/rowBytes {
		return false
	}
	// Guard combined storage, not merely the cell count.
	remaining := maxInt - h*rowBytes
	return h == 0 || w <= remaining/cellBytes/h
}

// Storage shares only rectangular allocation and value copying. Grapheme
// mutations additionally enforce cluster boundaries in GraphemeGrid. A grid
// holds one as an unexported field, so none of this reaches its own API.
type Storage[T any] struct {
	Area    image.Rectangle
	Content [][]T
}

func (s *Storage[T]) Resize(w, h int) {
	w, h = max(0, w), max(0, h)
	if !ValidAllocation[T](w, h) {
		panic(ErrInvalidSize)
	}
	if w == s.Area.Dx() && h == s.Area.Dy() {
		return
	}
	rows := make([][]T, h)
	for y := range rows {
		rows[y] = make([]T, w)
		if y < len(s.Content) {
			copy(rows[y], s.Content[y])
		}
	}
	s.Area = image.Rect(0, 0, w, h)
	s.Content = rows
}
func (s *Storage[T]) Clear() {
	for _, row := range s.Content {
		clear(row)
	}
}
func (s *Storage[T]) FillRect(r image.Rectangle, c T) {
	r = r.Intersect(s.Area)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			s.Content[y][x] = c
		}
	}
}
func (s *Storage[T]) CopyRect(dst image.Point, src *Storage[T], r image.Rectangle) {
	sx, dx, w := CopyAxis(dst.X, r.Min.X, r.Max.X, src.Area.Dx(), s.Area.Dx())
	sy, dy, h := CopyAxis(dst.Y, r.Min.Y, r.Max.Y, src.Area.Dy(), s.Area.Dy())
	if w == 0 || h == 0 {
		return
	}
	cells := make([]T, w*h)
	for y := 0; y < h; y++ {
		copy(cells[y*w:(y+1)*w], src.Content[sy+y][sx:sx+w])
	}
	for y := 0; y < h; y++ {
		copy(s.Content[dy+y][dx:dx+w], cells[y*w:(y+1)*w])
	}
}

func (m *Storage[T]) ShiftUp() {
	c := m.Content
	if len(c) == 0 {
		return
	}
	row := c[0]
	copy(c, c[1:])
	clear(row)
	c[len(c)-1] = row
}

func (m *Storage[T]) ShiftDown() {
	c := m.Content
	if len(c) == 0 {
		return
	}
	row := c[len(c)-1]
	copy(c[1:], c)
	clear(row)
	c[0] = row
}

func (m *Storage[T]) ShiftLeft() {
	for i := range m.Content {
		cl := m.Content[i]
		if len(cl) == 0 {
			continue
		}
		copy(cl, cl[1:])
		var zero T
		cl[len(cl)-1] = zero
	}
}

func (m *Storage[T]) ShiftRight() {
	for i := range m.Content {
		cl := m.Content[i]
		if len(cl) == 0 {
			continue
		}
		copy(cl[1:], cl)
		var zero T
		cl[0] = zero
	}
}
