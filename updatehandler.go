// ntcharts - Copyright (c) 2024 Neomantra Corp.

package ntcanvas

// File contains methods and objects used during canvas Model Update()
// to modify internal state.
// canvas is able to move the viewport displaying the contents
// either up, down, left and right

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// KeyMap configures the keyboard bindings used by DefaultUpdateHandler.
// PgUp and PgDown move one row, like Up and Down; no binding performs zooming.
type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	PgUp   key.Binding
	PgDown key.Binding
}

// DefaultKeyMap returns a default KeyMap for canvas.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "move left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "move right"),
		),
		PgUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("PgUp", "move up"),
		),
		PgDown: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("PgDown", "move down"),
		),
	}
}

// UpdateHandler callback invoked during an Update()
// and passes in the canvas *Model and bubbletea Msg.
type UpdateHandler func(*Model, tea.Msg)

// DefaultUpdateHandler is used by canvas chart to enable
// moving viewing window using the mouse wheel,
// holding down mouse left button and moving,
// and with the arrow keys.
// Uses canvas Keymap for keyboard messages.
func DefaultUpdateHandler() UpdateHandler {
	return func(m *Model, tm tea.Msg) {
		switch msg := tm.(type) {
		case tea.KeyPressMsg:
			if m.KeyMap == nil {
				return // nil disables keyboard bindings; mouse handling remains available
			}
			switch {
			case key.Matches(msg, m.KeyMap.Up):
				m.MoveUp(1)
			case key.Matches(msg, m.KeyMap.Down):
				m.MoveDown(1)
			case key.Matches(msg, m.KeyMap.Left):
				m.MoveLeft(1)
			case key.Matches(msg, m.KeyMap.Right):
				m.MoveRight(1)
			case key.Matches(msg, m.KeyMap.PgUp):
				m.MoveUp(1)
			case key.Matches(msg, m.KeyMap.PgDown):
				m.MoveDown(1)
			}
		case tea.MouseWheelMsg:
			mouse := msg.Mouse()
			switch mouse.Button {
			case tea.MouseWheelUp:
				m.MoveUp(1)
			case tea.MouseWheelDown:
				m.MoveDown(1)
			case tea.MouseWheelRight:
				m.MoveRight(1)
			case tea.MouseWheelLeft:
				m.MoveLeft(1)
			}
		case tea.MouseReleaseMsg:
			m.dragging = false
		case tea.MouseClickMsg:
			m.dragging = false
			if m.zoneManager == nil {
				return
			}
			zInfo := m.zoneManager.Get(m.zoneID)
			if zInfo.InBounds(msg) {
				x, y := zInfo.Pos(msg)
				m.lastMouse = Point{X: x, Y: y}
				m.dragging = true
			}
		case tea.MouseMotionMsg: // event occurs when mouse is pressed
			if !m.dragging || msg.Mouse().Button == tea.MouseNone {
				return
			}
			if m.zoneManager == nil {
				return
			}
			zInfo := m.zoneManager.Get(m.zoneID)
			if zInfo.InBounds(msg) {
				x, y := zInfo.Pos(msg)
				if x > m.lastMouse.X {
					m.MoveRight(1)
				} else if x < m.lastMouse.X {
					m.MoveLeft(1)
				}
				if y > m.lastMouse.Y {
					m.MoveDown(1)
				} else if y < m.lastMouse.Y {
					m.MoveUp(1)
				}
				m.lastMouse = Point{X: x, Y: y} // update last mouse position
			}
		}
	}
}

// Movement saturates within the surface; nonpositive amounts are ignored.
func (m *Model) move(dx, dy int, forward bool) {
	m.cursor = m.Cursor()
	b := m.Bounds()
	if b.Empty() {
		return
	}
	if forward {
		endX := b.Min.X + max(0, b.Dx()-max(1, m.ViewWidth))
		endY := b.Min.Y + max(0, b.Dy()-max(1, m.ViewHeight))
		if dx > 0 {
			m.cursor.X = min(m.cursor.X, endX)
			m.cursor.X += min(dx, endX-m.cursor.X)
		}
		if dy > 0 {
			m.cursor.Y = min(m.cursor.Y, endY)
			m.cursor.Y += min(dy, endY-m.cursor.Y)
		}
	} else {
		m.cursor.X -= min(max(0, dx), m.cursor.X-b.Min.X)
		m.cursor.Y -= min(max(0, dy), m.cursor.Y-b.Min.Y)
	}
}

// MoveUp pans by up to n cells toward the top edge.
// Nonpositive amounts are ignored.
func (m *Model) MoveUp(n int) {
	if n > 0 {
		m.move(0, n, false)
	}
}

// MoveDown pans by up to n cells, stopping at the last full viewport.
// A partial edge viewport is first moved back to that limit. Nonpositive amounts
// are ignored.
func (m *Model) MoveDown(n int) {
	if n > 0 {
		m.move(0, n, true)
	}
}

// MoveLeft pans by up to n cells toward the left edge.
// Nonpositive amounts are ignored.
func (m *Model) MoveLeft(n int) {
	if n > 0 {
		m.move(n, 0, false)
	}
}

// MoveRight pans by up to n cells, stopping at the last full viewport.
// A partial edge viewport is first moved back to that limit. Nonpositive amounts
// are ignored.
func (m *Model) MoveRight(n int) {
	if n > 0 {
		m.move(n, 0, true)
	}
}
