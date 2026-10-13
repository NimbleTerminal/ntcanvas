package ntcanvas

import "image"

// storage shares only rectangular allocation and value copying. Grapheme
// mutations additionally enforce cluster boundaries in GraphemeGrid.
type storage[T any] struct {
	area    image.Rectangle
	content [][]T
}

func (s *storage[T]) resize(w, h int) {
	w, h = max(0, w), max(0, h)
	// All current cell types fit within Cell. Use the conservative common cap.
	if !validAllocation(w, h) {
		panic(ErrInvalidSize)
	}
	rows := make([][]T, h)
	for y := range rows {
		rows[y] = make([]T, w)
		if y < len(s.content) {
			copy(rows[y], s.content[y])
		}
	}
	s.area = image.Rect(0, 0, w, h)
	s.content = rows
}
func (s *storage[T]) clear() {
	for _, row := range s.content {
		clear(row)
	}
}
func (s *storage[T]) fillRect(r image.Rectangle, c T) {
	r = r.Intersect(s.area)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			s.content[y][x] = c
		}
	}
}
func (s *storage[T]) copyRect(dst Point, src *storage[T], r image.Rectangle) {
	sx, dx, w := copyAxis(dst.X, r.Min.X, r.Max.X, src.area.Dx(), s.area.Dx())
	sy, dy, h := copyAxis(dst.Y, r.Min.Y, r.Max.Y, src.area.Dy(), s.area.Dy())
	if w == 0 || h == 0 {
		return
	}
	cells := make([]T, w*h)
	for y := 0; y < h; y++ {
		copy(cells[y*w:(y+1)*w], src.content[sy+y][sx:sx+w])
	}
	for y := 0; y < h; y++ {
		copy(s.content[dy+y][dx:dx+w], cells[y*w:(y+1)*w])
	}
}

func (m *storage[T]) shiftUp() {
	c := m.content
	if len(c) == 0 {
		return
	}
	row := c[0]
	copy(c, c[1:])
	clear(row)
	c[len(c)-1] = row
}

func (m *storage[T]) shiftDown() {
	c := m.content
	if len(c) == 0 {
		return
	}
	row := c[len(c)-1]
	copy(c[1:], c)
	clear(row)
	c[0] = row
}

func (m *storage[T]) shiftLeft() {
	for i := range m.content {
		cl := m.content[i]
		if len(cl) == 0 {
			continue
		}
		copy(cl, cl[1:])
		var zero T
		cl[len(cl)-1] = zero
	}
}

func (m *storage[T]) shiftRight() {
	for i := range m.content {
		cl := m.content[i]
		if len(cl) == 0 {
			continue
		}
		copy(cl[1:], cl)
		var zero T
		cl[0] = zero
	}
}
