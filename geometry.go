package ntcanvas

import (
	"charm.land/lipgloss/v2"
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

// SizeLimits bounds a checked allocation. Zero fields use defaults: dimensions
// at most 4096 and at most 65536 cells. These bound one grid, not total process
// memory; Resize temporarily retains the previous grid, and styles may reference
// additional data. Negative limits are invalid.
type SizeLimits struct{ MaxWidth, MaxHeight, MaxCells int }

// ValidateSize checks dimensions without allocating. It also rejects sizes whose
// cell or row storage would overflow an int, including on 32-bit/WASM targets.
func ValidateSize(w, h int, limits SizeLimits) error {
	if limits.MaxWidth < 0 || limits.MaxHeight < 0 || limits.MaxCells < 0 || !validAllocation(w, h) {
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

func validAllocation(w, h int) bool {
	if w < 0 || h < 0 {
		return false
	}
	maxInt := int(^uint(0) >> 1)
	rowBytes := int(unsafe.Sizeof(CellLine{}))
	cellBytes := int(unsafe.Sizeof(Cell{}))
	if h > maxInt/rowBytes {
		return false
	}
	// Guard combined storage, not merely the cell count.
	remaining := maxInt - h*rowBytes
	return h == 0 || w <= remaining/cellBytes/h
}

// FillRect fills r intersected with Bounds. It does not alter the default style.
func (m *StyledRuneGrid) FillRect(r image.Rectangle, c Cell) {
	if c.Rune != 0 {
		c.Rune = plainRune(c.Rune)
	}
	r = r.Intersect(m.area)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			m.content[y][x] = c
		}
	}
}

// ClearRect resets r intersected with Bounds to empty, unstyled cells.
func (m *StyledRuneGrid) ClearRect(r image.Rectangle) { m.FillRect(r, Cell{}) }

// SetStyleRect replaces styles in r intersected with Bounds, preserving runes
// and the default Style used for future writes. It does not merge attributes.
func (m *StyledRuneGrid) SetStyleRect(r image.Rectangle, s lipgloss.Style) {
	r = r.Intersect(m.area)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			m.content[y][x].Style = s
		}
	}
}

// CopyRect copies src's cells in r to dst, mapping r.Min to dst. Both canvases
// clip the copy without changing that mapping. Blank cells overwrite destination
// cells; no transparency or style merging is applied. Grids have no viewport state.
// Overlapping storage (including copies within the same grid) behaves as if the source
// cells were snapshotted first. Cell/style values are copied, not callback state.
// The temporary allocation is proportional to the clipped copied area.
func (m *StyledRuneGrid) CopyRect(dst Point, src *StyledRuneGrid, r image.Rectangle) {
	m.copyRect(dst, &src.storage, r)
}

// Unsigned distances avoid signed overflow for rectangles spanning MinInt to
// MaxInt. Conversion back to int happens only after clipping to a canvas axis.
func copyAxis(dst, lo, hi, srcSize, dstSize int) (src, target, length int) {
	src, end := max(0, lo), min(srcSize, hi)
	if src >= end || dstSize == 0 || dst >= dstSize {
		return 0, 0, 0
	}
	offset := uint(src) - uint(lo)
	if offset > uint(dstSize-1)-uint(dst) {
		return 0, 0, 0
	}
	if dst < 0 {
		before := uint(0) - uint(dst)
		if offset < before {
			skip := before - offset
			if skip >= uint(end-src) {
				return 0, 0, 0
			}
			src += int(skip)
			offset += skip
		}
	}
	target = int(uint(dst) + offset)
	return src, target, min(end-src, dstSize-target)
}
