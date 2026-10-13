package ntcanvas

import (
	"charm.land/lipgloss/v2"
	"errors"
	"image"
	"strings"
	"testing"
	"unicode"
)

func TestCheckedSizeAndResizeTransaction(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	for _, tc := range []struct {
		w, h   int
		limits SizeLimits
		err    error
	}{
		{-1, 1, SizeLimits{}, ErrInvalidSize}, {1, -1, SizeLimits{}, ErrInvalidSize},
		{1, 1, SizeLimits{MaxCells: -1}, ErrInvalidSize},
		{maxInt, maxInt, SizeLimits{}, ErrInvalidSize},
		{0, maxInt, SizeLimits{}, ErrInvalidSize},
		{maxInt, 0, SizeLimits{}, ErrSizeLimit},
		{3, 2, SizeLimits{MaxCells: 5}, ErrSizeLimit},
		{3, 2, SizeLimits{MaxWidth: 2}, ErrSizeLimit},
		{3, 2, SizeLimits{MaxHeight: 1}, ErrSizeLimit},
		{3, 2, SizeLimits{MaxCells: 6}, nil}, {0, 0, SizeLimits{}, nil},
	} {
		if err := ValidateSize(tc.w, tc.h, tc.limits); !errors.Is(err, tc.err) {
			t.Fatalf("%+v: %v", tc, err)
		}
	}
	m, err := NewStyledRuneGridChecked(3, 2, SizeLimits{MaxCells: 6}, WithContent([]string{"abc", "def"}))
	if err != nil {
		t.Fatal(err)
	}
	before := m.Render(m.Bounds())
	if err = m.ResizeChecked(100, 100, SizeLimits{MaxCells: 6}); !errors.Is(err, ErrSizeLimit) || m.Render(m.Bounds()) != before {
		t.Fatal("rejected resize mutated model", err)
	}
	if err = m.ResizeChecked(2, 1, SizeLimits{}); err != nil || m.Render(m.Bounds()) != "ab" {
		t.Fatal("checked resize", err, m.Render(m.Bounds()))
	}
}

func TestEmptyAndExtremeMovement(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	minInt := -maxInt - 1
	for _, size := range []Point{{}, {X: -1, Y: 1}, {X: 1, Y: -1}, {X: 0, Y: 2}} {
		m := NewStyledRuneGrid(size.X, size.Y)
		model := NewModel(m)
		model.SetCursor(Point{X: maxInt, Y: maxInt})
		m.ShiftUp()
		m.ShiftDown()
		m.ShiftLeft()
		m.ShiftRight()
		model.ViewWidth, model.ViewHeight = maxInt, maxInt
		if model.View() != "" {
			t.Fatal("empty grid output")
		}
		if model.Cursor().X < 0 || model.Cursor().Y < 0 {
			t.Fatal("negative empty cursor")
		}
	}
	m := NewModel(NewStyledRuneGrid(5, 4), WithViewWidth(2), WithViewHeight(2))
	m.MoveRight(maxInt)
	m.MoveDown(maxInt)
	if m.Cursor() != (Point{X: 3, Y: 2}) {
		t.Fatal("oversized movement", m.Cursor())
	}
	m.MoveRight(minInt)
	m.MoveDown(-1)
	m.MoveUp(-1)
	m.MoveLeft(minInt)
	if m.Cursor() != (Point{X: 3, Y: 2}) {
		t.Fatal("negative movement changed cursor")
	}
	m.MoveLeft(maxInt)
	m.MoveUp(maxInt)
	if m.Cursor() != (Point{}) {
		t.Fatal("movement overflow")
	}
	m.SetCursor(Point{X: 4, Y: 3})
	m.ViewWidth, m.ViewHeight = maxInt, maxInt
	if m.View() != " " {
		t.Fatal("overflowing viewport", m.View())
	}
	m.ViewWidth = -1
	if m.View() != "" {
		t.Fatal("negative viewport")
	}
	var zero Model
	zero.Focus()
	zero.Update(nil) // nil update handler is safe
}

func TestRectangles(t *testing.T) {
	m := NewStyledRuneGrid(4, 3)
	m.FillRect(image.Rect(-1, -1, 2, 2), NewCell('x'))
	if m.Render(m.Bounds()) != "xx  \nxx  \n    " {
		t.Fatal(m.Render(m.Bounds()))
	}
	m.ClearRect(image.Rect(1, 0, 9, 1))
	if m.Render(m.Bounds()) != "x   \nxx  \n    " {
		t.Fatal(m.Render(m.Bounds()))
	}
	m.FillRect(image.Rectangle{Min: Point{X: 3, Y: 2}, Max: Point{X: 1, Y: 1}}, NewCell('!'))
	model := NewModel(m)
	model.SetViewport(image.Rect(1, 1, 9, 9))
	if model.Bounds() != image.Rect(0, 0, 4, 3) || model.Viewport() != image.Rect(1, 1, 4, 3) || model.View() != "x  \n   " {
		t.Fatal("viewport", model.Viewport(), model.View())
	}
	model.SetViewport(image.Rect(-3, -3, -1, -1))
	if !model.Viewport().Empty() || model.View() != "" {
		t.Fatal("empty intersection")
	}
}

func TestPlainTextAndClipping(t *testing.T) {
	input := "\x1b[2J\x1b]0;title\a\n\r\t\x7f\u009b\u202e\u2028\u200d\xff"
	clean := SanitizeText(input)
	for _, r := range clean {
		if unicode.IsControl(r) || unicode.In(r, unicode.Cf, unicode.Zl, unicode.Zp) {
			t.Fatalf("control survived: %U", r)
		}
	}
	m := NewStyledRuneGrid(len([]rune(clean)), 1)
	m.SetText(Point{}, input)
	if m.Render(m.Bounds()) != clean {
		t.Fatalf("plain %q != %q", m.Render(m.Bounds()), clean)
	}
	m.SetRawText(Point{}, input)
	if !strings.Contains(m.Render(m.Bounds()), "\x1b[2J") {
		t.Fatal("raw API must remain raw")
	}
	m = NewStyledRuneGrid(3, 1)
	m.SetTextWithStyle(Point{X: -1}, "a\nbc", lipgloss.NewStyle().Bold(true))
	if m.Cell(Point{}).Rune != '�' || !m.Cell(Point{}).Style.GetBold() || m.Cell(Point{X: 2}).Rune != 'c' {
		t.Fatal("safe styled left clip")
	}
	minInt := -int(^uint(0)>>1) - 1
	for _, x := range []int{minInt, 3, 100} {
		if !m.SetText(Point{X: x}, "界x") || !m.SetRunes(Point{X: x}, []rune("界x")) {
			t.Fatal("valid row clipped write")
		}
	}
	if m.SetText(Point{Y: -1}, "x") {
		t.Fatal("invalid row accepted")
	}
}

func TestSharedStorageOperations(t *testing.T) {
	m := NewStyledRuneGrid(2, 2, WithContent([]string{"ab", "cd"}))
	other := m
	m.ShiftUp()
	if other.Render(other.Bounds()) != "cd\n  " {
		t.Fatal("shift must remain shared")
	}
	m.Clear()
	if other.Render(other.Bounds()) != "  \n  " {
		t.Fatal("clear must remain shared")
	}
	m.SetRune(Point{}, 'x')
	other = m
	m.Resize(2, 2)
	m.Clear()
	if other.Cell(Point{}).Rune != 0 {
		t.Fatal("grid pointers must share resized storage")
	}
}

func FuzzPlainText(f *testing.F) {
	for _, s := range []string{"hello", "\x1b[2J", "界e\u0301", "\xff\x00\u202e"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		m := NewStyledRuneGrid(8, 1)
		m.SetText(Point{}, s)
		for x := 0; x < 8; x++ {
			r := m.Cell(Point{X: x}).Rune
			if r != 0 && (unicode.IsControl(r) || unicode.In(r, unicode.Cf, unicode.Zl, unicode.Zp)) {
				t.Fatalf("control %U", r)
			}
		}
	})
}

func TestImpossibleAllocationPanicsBeforeMake(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	for _, f := range []func(){func() { NewStyledRuneGrid(maxInt, 2) }, func() { m := NewStyledRuneGrid(1, 1); m.Resize(maxInt, 2) }} {
		func() {
			defer func() {
				err, ok := recover().(error)
				if !ok || !errors.Is(err, ErrInvalidSize) {
					t.Errorf("expected ErrInvalidSize, got %v", err)
				}
			}()
			f()
		}()
	}
}

func TestClippedWritesAndClearReuseStorage(t *testing.T) {
	m := NewStyledRuneGrid(40, 2)
	large := strings.Repeat("x", 1000000)
	if n := testing.AllocsPerRun(20, func() { m.SetText(Point{}, large); m.Clear() }); n != 0 {
		t.Fatal("clipped text/clear allocated", n)
	}
}
