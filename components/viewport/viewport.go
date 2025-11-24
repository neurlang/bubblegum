// Package viewport provides a viewport component for BubbleGum applications.
package viewport

import (
	"math"
	"strings"

	"github.com/bubblegum/lib"
)

// Model is the viewport model for BubbleGum.
type Model struct {
	// Width and Height define the viewport dimensions.
	Width  int
	Height int

	// MouseWheelEnabled enables mouse wheel scrolling.
	MouseWheelEnabled bool

	// MouseWheelDelta is the number of lines to scroll per wheel event.
	MouseWheelDelta int

	// YOffset is the vertical scroll position.
	YOffset int

	// lines contains the content split into lines.
	lines []string
}

// New returns a new viewport model with the given width and height.
func New(width, height int) Model {
	return Model{
		Width:             width,
		Height:            height,
		MouseWheelEnabled: true,
		MouseWheelDelta:   3,
		YOffset:           0,
		lines:             []string{},
	}
}

// SetContent sets the viewport's text content.
func (m *Model) SetContent(s string) {
	s = strings.ReplaceAll(s, "\r\n", "\n") // normalize line endings
	m.lines = strings.Split(s, "\n")

	// Adjust offset if we're past the bottom
	if m.YOffset > m.maxYOffset() {
		m.GotoBottom()
	}
}

// SetSize sets the width and height of the viewport.
func (m *Model) SetSize(width, height int) {
	m.Width = width
	m.Height = height

	// Adjust offset if needed
	if m.YOffset > m.maxYOffset() {
		m.SetYOffset(m.maxYOffset())
	}
}

// maxYOffset returns the maximum possible Y offset.
func (m Model) maxYOffset() int {
	if len(m.lines) <= m.Height {
		return 0
	}
	return len(m.lines) - m.Height
}

// SetYOffset sets the Y offset, clamping to valid range.
func (m *Model) SetYOffset(n int) {
	if n < 0 {
		n = 0
	}
	max := m.maxYOffset()
	if n > max {
		n = max
	}
	m.YOffset = n
}

// AtTop returns whether the viewport is at the top.
func (m Model) AtTop() bool {
	return m.YOffset <= 0
}

// AtBottom returns whether the viewport is at the bottom.
func (m Model) AtBottom() bool {
	return m.YOffset >= m.maxYOffset()
}

// ScrollPercent returns the scroll position as a percentage (0.0 to 1.0).
func (m Model) ScrollPercent() float64 {
	if m.Height >= len(m.lines) {
		return 1.0
	}
	y := float64(m.YOffset)
	h := float64(m.Height)
	t := float64(len(m.lines))
	v := y / (t - h)
	return math.Max(0.0, math.Min(1.0, v))
}

// TotalLineCount returns the total number of lines.
func (m Model) TotalLineCount() int {
	return len(m.lines)
}

// VisibleLineCount returns the number of visible lines.
func (m Model) VisibleLineCount() int {
	if len(m.lines) == 0 {
		return 0
	}
	top := m.YOffset
	if top < 0 {
		top = 0
	}
	bottom := m.YOffset + m.Height
	if bottom > len(m.lines) {
		bottom = len(m.lines)
	}
	return bottom - top
}

// ScrollDown scrolls down by n lines.
func (m *Model) ScrollDown(n int) {
	if m.AtBottom() || n == 0 {
		return
	}
	m.SetYOffset(m.YOffset + n)
}

// ScrollUp scrolls up by n lines.
func (m *Model) ScrollUp(n int) {
	if m.AtTop() || n == 0 {
		return
	}
	m.SetYOffset(m.YOffset - n)
}

// PageDown scrolls down by one page (viewport height).
func (m *Model) PageDown() {
	m.ScrollDown(m.Height)
}

// PageUp scrolls up by one page (viewport height).
func (m *Model) PageUp() {
	m.ScrollUp(m.Height)
}

// HalfPageDown scrolls down by half a page.
func (m *Model) HalfPageDown() {
	m.ScrollDown(m.Height / 2)
}

// HalfPageUp scrolls up by half a page.
func (m *Model) HalfPageUp() {
	m.ScrollUp(m.Height / 2)
}

// GotoTop scrolls to the top.
func (m *Model) GotoTop() {
	m.SetYOffset(0)
}

// GotoBottom scrolls to the bottom.
func (m *Model) GotoBottom() {
	m.SetYOffset(m.maxYOffset())
}

// Update handles viewport updates.
func (m Model) Update(msg lib.Msg) (Model, lib.Cmd) {
	switch msg := msg.(type) {
	case lib.KeyMsg:
		switch msg.Type {
		case lib.KeyPgDown:
			m.PageDown()

		case lib.KeyPgUp:
			m.PageUp()

		case lib.KeyDown:
			m.ScrollDown(1)

		case lib.KeyUp:
			m.ScrollUp(1)

		case lib.KeyHome:
			m.GotoTop()

		case lib.KeyEnd:
			m.GotoBottom()
		}

	case lib.MouseMsg:
		if !m.MouseWheelEnabled {
			break
		}

		switch msg.Button {
		case lib.MouseButtonWheelUp:
			m.ScrollUp(m.MouseWheelDelta)

		case lib.MouseButtonWheelDown:
			m.ScrollDown(m.MouseWheelDelta)
		}

	case lib.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
	}

	return m, nil
}

// View renders the viewport.
func (m Model) View() string {
	if len(m.lines) == 0 {
		return strings.Repeat("\n", m.Height-1)
	}

	// Calculate visible range
	top := m.YOffset
	if top < 0 {
		top = 0
	}
	bottom := m.YOffset + m.Height
	if bottom > len(m.lines) {
		bottom = len(m.lines)
	}

	// Get visible lines
	visibleLines := m.lines[top:bottom]

	// Pad with empty lines if needed
	if len(visibleLines) < m.Height {
		padding := m.Height - len(visibleLines)
		for i := 0; i < padding; i++ {
			visibleLines = append(visibleLines, "")
		}
	}

	return strings.Join(visibleLines, "\n")
}
