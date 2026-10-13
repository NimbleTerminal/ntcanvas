// Package parity defines deterministic public-API scenarios shared by the
// upstream fixture generator and ntcanvas's independent regression tests.
package parity

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"encoding/json"
	"image"
	"nimbleterminal.dev/ntcanvas"
)

func Snapshot() ([]byte, error) {
	out := map[string]any{}
	origin := ntcanvas.Point{X: 4, Y: 8}
	out["coordinates"] = []any{
		ntcanvas.NewFloat64PointFromPoint(ntcanvas.Point{X: -3, Y: 5}),
		ntcanvas.NewPointFromFloat64Point(ntcanvas.Float64Point{X: -1.5, Y: 2.5}),
		ntcanvas.CanvasYCoordinates(8, []int{-2, 0, 3}),
		ntcanvas.CanvasPoints(origin, []ntcanvas.Point{{X: -2, Y: 1}, {X: 3, Y: -4}}),
		ntcanvas.CanvasPointFromFloat64Point(origin, ntcanvas.Float64Point{X: -1.5, Y: 2.5}),
		ntcanvas.CanvasFloat64Point(origin, ntcanvas.Float64Point{X: 1.25, Y: -2.5}),
	}
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Background(lipgloss.Color("4")).Bold(true)
	m := ntcanvas.NewStyledRuneGrid(4, 3, ntcanvas.WithStyle(style))
	m.SetCell(ntcanvas.Point{}, ntcanvas.NewCellWithStyle('A', style))
	m.SetRunes(ntcanvas.Point{X: 1}, []rune{'B', 'C', 'D', 'E'})
	m.SetTextWithStyle(ntcanvas.Point{X: -1, Y: 1}, "012345", lipgloss.NewStyle().Underline(true))
	m.SetRuneWithStyle(ntcanvas.Point{X: 2, Y: 2}, '!', style)
	m.SetCellStyle(ntcanvas.Point{X: 3, Y: 2}, style)
	out["styled_cells"] = m.Render(m.Bounds())
	out["clipped_returns"] = []bool{m.SetCell(ntcanvas.Point{X: 4}, ntcanvas.NewCell('x')), m.SetCellStyle(ntcanvas.Point{Y: -1}, style), m.SetText(ntcanvas.Point{X: 9}, "x"), m.SetText(ntcanvas.Point{Y: 3}, "x")}
	out["style_lookup"] = []any{m.GetCellStyle(ntcanvas.Point{X: -1}) == nil, m.GetCellStyle(ntcanvas.Point{}).Render("A")}
	m.FillLine(1, ntcanvas.NewCell('z'))
	m.FillLine(-1, ntcanvas.NewCell('x'))
	out["fill_line"] = m.Render(m.Bounds())
	m.Resize(3, 4)
	out["resize"] = m.Render(image.Rect(0, 0, 4, 3))
	m.ShiftDown()
	m.ShiftRight()
	out["shift_down_right"] = m.Render(image.Rect(0, 0, 4, 3))
	m.ShiftUp()
	m.ShiftLeft()
	out["shift_up_left"] = m.Render(image.Rect(0, 0, 4, 3))
	// The historical fixture includes an oversized-viewport trailing newline.
	// Surface.Render now deliberately omits it; normalize only this legacy case.
	out["edge_view"] = m.Render(image.Rect(2, 3, 4, 5)) + "\n"
	m = ntcanvas.NewStyledRuneGrid(4, 1)
	m.SetText(ntcanvas.Point{}, "界e\u0301xZ")
	out["unicode"] = m.Render(m.Bounds())
	model := ntcanvas.NewModel(ntcanvas.NewStyledRuneGrid(8, 7), ntcanvas.WithFocus(), ntcanvas.WithViewWidth(3), ntcanvas.WithViewHeight(2))
	positions := []ntcanvas.Point{}
	for _, msg := range []tea.Msg{tea.KeyPressMsg{Code: tea.KeyRight}, tea.KeyPressMsg{Code: 'j', Text: "j"}, tea.KeyPressMsg{Code: tea.KeyPgDown}, tea.MouseWheelMsg{Button: tea.MouseWheelLeft}, tea.MouseWheelMsg{Button: tea.MouseWheelUp}, tea.KeyPressMsg{Code: tea.KeyPgUp}, tea.KeyPressMsg{Code: 'h', Text: "h"}} {
		model, _ = model.Update(msg)
		positions = append(positions, model.Cursor())
	}
	out["input_positions"] = positions
	return json.MarshalIndent(out, "", "  ")
}
