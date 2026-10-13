// ntcharts - Copyright (c) 2024 Neomantra Corp.

package ntcanvas

// Options configure either StyledRuneGrid or Model; storage and input stay separate.

import (
	"charm.land/lipgloss/v2"
	zone "github.com/lrstanley/bubblezone/v2"
)

// ModelOption configures viewport and interaction state in NewModel.
type ModelOption func(*Model)

// WithStyle sets the default Cell style.
func WithStyle(s lipgloss.Style) StyledOption {
	return func(m *StyledRuneGrid) {
		m.Style = s
	}
}

// WithKeyMap installs a separate copy of k for each model configured by this
// option. As with key.Binding, key-name slices inside bindings are shared.
func WithKeyMap(k KeyMap) ModelOption {
	return func(m *Model) {
		keys := k
		m.KeyMap = &keys
	}
}

// WithUpdateHandler sets the UpdateHandler used
// when processing bubbletea Msg events in Update().
func WithUpdateHandler(h UpdateHandler) ModelOption {
	return func(m *Model) {
		m.UpdateHandler = h
	}
}

// WithZoneManager sets the bubblezone Manager used
// when processing bubbletea Msg mouse events in Update().
func WithZoneManager(zm *zone.Manager) ModelOption {
	return func(m *Model) {
		m.SetZoneManager(zm)
	}
}

// WithCursor sets the cursor starting position for the viewport.
func WithCursor(p Point) ModelOption {
	return func(m *Model) {
		m.SetCursor(p)
	}
}

// WithFocus sets the canvas to be focused.
func WithFocus() ModelOption {
	return func(m *Model) {
		m.Focus()
	}
}

// WithContent copies the given []string into
// the contents of the canvas with default style.
// Each string will be copied into a row starting
// from the top of the canvas to bottom.
// Use option WithStyle() to set canvas style
// before using WithContent() for styling.
func WithContent(l []string) StyledOption {
	return func(m *StyledRuneGrid) {
		m.SetLines(l)
	}
}

// WithViewWidth sets the viewport width of the canvas.
func WithViewWidth(w int) ModelOption {
	return func(m *Model) {
		m.ViewWidth = w
	}
}

// WithViewHeight sets the viewport height of the canvas.
func WithViewHeight(h int) ModelOption {
	return func(m *Model) {
		m.ViewHeight = h
	}
}
