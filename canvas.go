// ntcharts - Copyright (c) 2024 Neomantra Corp.

// Package ntcanvas provides terminal-cell surfaces and a Bubble Tea viewport.
package ntcanvas

import (
	tea "charm.land/bubbletea/v2"
	zone "github.com/lrstanley/bubblezone/v2"
	"image"
)

// Surface renders a rectangle in cell coordinates. Render must clip to Bounds,
// return empty output for empty intersections, and omit a trailing newline.
// Rendering does not consume input or write directly to a terminal.
// Bounds must have nonnegative dimensions representable by int.
type Surface interface {
	Bounds() image.Rectangle
	Render(image.Rectangle) string
}

// Model displays a Surface. Copies share the surface and key map, but viewport, focus and
// built-in drag state are copied by value. Custom callbacks may share captures.
// Grid allocation/resizing belongs to the concrete surface, not this model.
type Model struct {
	surface               Surface
	KeyMap                *KeyMap
	UpdateHandler         UpdateHandler
	ViewWidth, ViewHeight int
	cursor                Point
	focus                 bool
	zoneManager           *zone.Manager
	zoneID                string
	lastMouse             Point
	dragging              bool
}

// NewModel wraps a surface. nil means an empty surface; typed nil pointers are
// not valid surfaces. Options execute in order. Retain the result of Update.
func NewModel(surface Surface, opts ...ModelOption) Model {
	keys := DefaultKeyMap()
	m := Model{surface: surface, KeyMap: &keys, UpdateHandler: DefaultUpdateHandler()}
	b := m.Bounds()
	m.cursor = b.Min
	m.ViewWidth = b.Dx()
	m.ViewHeight = b.Dy()
	for _, opt := range opts {
		opt(&m)
	}
	return m
}
func (m Model) Surface() Surface { return m.surface }
func (m Model) Bounds() image.Rectangle {
	if m.surface == nil {
		return image.Rectangle{}
	}
	return m.surface.Bounds()
}
func (m Model) Width() int    { return m.Bounds().Dx() }
func (m Model) Height() int   { return m.Bounds().Dy() }
func (m Model) Cursor() Point { return m.clampCursor(m.cursor) }
func (m Model) clampCursor(p Point) Point {
	b := m.Bounds()
	if b.Empty() {
		return b.Min
	}
	return Point{X: min(max(p.X, b.Min.X), b.Max.X-1), Y: min(max(p.Y, b.Min.Y), b.Max.Y-1)}
}

// SetCursor clamps to the last cell, allowing a partial viewport at an edge.
// Move methods instead stop at the last full viewport for interactive panning.
func (m *Model) SetCursor(p Point) { m.cursor = m.clampCursor(p) }

// Viewport returns the visible region, including after external grid resizing.
func (m Model) Viewport() image.Rectangle {
	b := m.Bounds()
	if b.Empty() || m.ViewWidth <= 0 || m.ViewHeight <= 0 {
		return image.Rectangle{}
	}
	p := m.Cursor()
	return image.Rectangle{Min: p, Max: Point{X: p.X + min(m.ViewWidth, b.Max.X-p.X), Y: p.Y + min(m.ViewHeight, b.Max.Y-p.Y)}}
}

// SetViewport clips a requested region to the surface. Window resizing never
// resizes the underlying surface automatically.
func (m *Model) SetViewport(r image.Rectangle) {
	b := m.Bounds()
	r = r.Intersect(b)
	m.cursor = r.Min
	if r.Empty() {
		m.cursor = b.Min
	}
	m.ViewWidth = r.Dx()
	m.ViewHeight = r.Dy()
}
func (m Model) View() string {
	var s string
	if r := m.Viewport(); m.surface != nil && !r.Empty() {
		s = m.surface.Render(r)
	}
	if m.zoneManager != nil {
		s = m.zoneManager.Mark(m.zoneID, s)
	}
	return s
}

func (m *Model) SetZoneManager(zm *zone.Manager) {
	m.dragging = false
	m.zoneManager = zm
	if (zm != nil) && (m.zoneID == "") {
		m.zoneID = zm.NewPrefix()
	}
}

func (m Model) ZoneManager() *zone.Manager {
	return m.zoneManager
}

func (m Model) ZoneID() string {
	return m.zoneID
}

func (m Model) Focused() bool {
	return m.focus
}

func (m *Model) Focus() {
	m.focus = true
}

func (m *Model) Blur() {
	m.focus = false
	m.dragging = false
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focus || m.UpdateHandler == nil {
		return m, nil
	}
	m.UpdateHandler(&m, msg)
	return m, nil
}
