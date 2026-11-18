package ntcanvasgrid

import "nimbleterminal.dev/ntcanvas/internal/cellstore"

var (
	// ErrInvalidSize indicates negative dimensions, invalid limits, or integer overflow.
	ErrInvalidSize = cellstore.ErrInvalidSize
	// ErrSizeLimit indicates a request exceeds a caller's allocation limits.
	ErrSizeLimit = cellstore.ErrSizeLimit
)

// SizeLimits bounds a checked allocation. Zero fields use defaults: dimensions
// at most 4096 and at most 65536 cells. These bound one grid, not total process
// memory; Resize temporarily retains the previous grid, and styles may reference
// additional data. Negative limits are invalid.
type SizeLimits struct{ MaxWidth, MaxHeight, MaxCells int }

// ValidateSize checks dimensions without allocating. It also rejects sizes whose
// cell or row storage would overflow an int, including on 32-bit/WASM targets,
// for the larger of this package's cells, GraphemeCell.
func ValidateSize(w, h int, limits SizeLimits) error {
	return cellstore.ValidateSize[GraphemeCell](w, h, cellstore.Limits(limits))
}
