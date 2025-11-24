// Package textinput provides a text input component for BubbleGum applications.
package textinput

import (
	"strings"
	"unicode"

	"github.com/neurlang/bubblegum/lib"
)

// Model is the text input model for BubbleGum.
type Model struct {
	// Prompt is the text displayed before the input field.
	Prompt string

	// Placeholder is shown when the input is empty.
	Placeholder string

	// Width is the maximum number of characters that can be displayed at once.
	// If 0 or less, there's no limit.
	Width int

	// CharLimit is the maximum amount of characters this input element will
	// accept. If 0 or less, there's no limit.
	CharLimit int

	// value is the underlying text value.
	value []rune

	// focus indicates whether user input focus should be on this input.
	focus bool

	// pos is the cursor position.
	pos int

	// offset is used to emulate a viewport when width is set.
	offset      int
	offsetRight int

	// showCursor tracks whether to show the cursor (for blinking effect).
	showCursor bool
}

// New creates a new text input model with default settings.
func New() Model {
	return Model{
		Prompt:      "> ",
		Placeholder: "",
		Width:       0,
		CharLimit:   0,
		value:       nil,
		focus:       false,
		pos:         0,
		showCursor:  true,
	}
}

// SetValue sets the value of the text input.
func (m *Model) SetValue(s string) {
	runes := []rune(s)
	if m.CharLimit > 0 && len(runes) > m.CharLimit {
		m.value = runes[:m.CharLimit]
	} else {
		m.value = runes
	}
	if m.pos > len(m.value) {
		m.SetCursor(len(m.value))
	}
	m.handleOverflow()
}

// Value returns the value of the text input.
func (m Model) Value() string {
	return string(m.value)
}

// Position returns the cursor position.
func (m Model) Position() int {
	return m.pos
}

// SetCursor moves the cursor to the given position.
func (m *Model) SetCursor(pos int) {
	if pos < 0 {
		pos = 0
	}
	if pos > len(m.value) {
		pos = len(m.value)
	}
	m.pos = pos
	m.handleOverflow()
}

// CursorStart moves the cursor to the start of the input field.
func (m *Model) CursorStart() {
	m.SetCursor(0)
}

// CursorEnd moves the cursor to the end of the input field.
func (m *Model) CursorEnd() {
	m.SetCursor(len(m.value))
}

// Focused returns the focus state on the model.
func (m Model) Focused() bool {
	return m.focus
}

// Focus sets the focus state on the model.
func (m *Model) Focus() lib.Cmd {
	m.focus = true
	return nil
}

// Blur removes the focus state on the model.
func (m *Model) Blur() {
	m.focus = false
}

// Reset sets the input to its default state with no input.
func (m *Model) Reset() {
	m.value = nil
	m.SetCursor(0)
}

// Update is the update loop for the text input.
func (m Model) Update(msg lib.Msg) (Model, lib.Cmd) {
	if !m.focus {
		return m, nil
	}

	switch msg := msg.(type) {
	case lib.KeyMsg:
		switch msg.Type {
		case lib.KeyBackspace:
			if len(m.value) > 0 && m.pos > 0 {
				m.value = append(m.value[:m.pos-1], m.value[m.pos:]...)
				m.SetCursor(m.pos - 1)
			}

		case lib.KeyDelete:
			if len(m.value) > 0 && m.pos < len(m.value) {
				m.value = append(m.value[:m.pos], m.value[m.pos+1:]...)
			}

		case lib.KeyLeft:
			if m.pos > 0 {
				m.SetCursor(m.pos - 1)
			}

		case lib.KeyRight:
			if m.pos < len(m.value) {
				m.SetCursor(m.pos + 1)
			}

		case lib.KeyHome:
			m.CursorStart()

		case lib.KeyEnd:
			m.CursorEnd()

		case lib.KeyCtrlC:
			// Let Ctrl+C pass through for quit handling
			return m, nil

		case lib.KeyRunes:
			// Insert runes at cursor position
			m.insertRunes(msg.Runes)

		default:
			// Ignore other keys
		}
	}

	m.handleOverflow()
	return m, nil
}

// insertRunes inserts runes at the cursor position.
func (m *Model) insertRunes(runes []rune) {
	// Check char limit
	if m.CharLimit > 0 {
		availSpace := m.CharLimit - len(m.value)
		if availSpace <= 0 {
			return
		}
		if len(runes) > availSpace {
			runes = runes[:availSpace]
		}
	}

	// Insert runes at cursor position
	head := m.value[:m.pos]
	tail := m.value[m.pos:]
	m.value = append(append(head, runes...), tail...)
	m.pos += len(runes)
}

// handleOverflow manages horizontal scrolling when width is set.
func (m *Model) handleOverflow() {
	if m.Width <= 0 || len(m.value) <= m.Width {
		m.offset = 0
		m.offsetRight = len(m.value)
		return
	}

	// Correct right offset if we've deleted characters
	if m.offsetRight > len(m.value) {
		m.offsetRight = len(m.value)
	}

	// Scroll left if cursor moved before visible area
	if m.pos < m.offset {
		m.offset = m.pos
		m.offsetRight = m.offset + m.Width
		if m.offsetRight > len(m.value) {
			m.offsetRight = len(m.value)
		}
	}

	// Scroll right if cursor moved after visible area
	if m.pos >= m.offsetRight {
		m.offsetRight = m.pos + 1
		m.offset = m.offsetRight - m.Width
		if m.offset < 0 {
			m.offset = 0
		}
	}
}

// View renders the text input in its current state.
func (m Model) View() string {
	// Show placeholder if empty
	if len(m.value) == 0 && m.Placeholder != "" {
		if m.focus && m.showCursor {
			return m.Prompt + "\x1b[7m \x1b[0m" + m.Placeholder[1:]
		}
		return m.Prompt + m.Placeholder
	}

	// Get visible portion of value
	value := m.value[m.offset:m.offsetRight]
	pos := m.pos - m.offset

	var result strings.Builder
	result.WriteString(m.Prompt)

	// Text before cursor
	if pos > 0 {
		result.WriteString(string(value[:pos]))
	}

	// Cursor and character under it
	if m.focus && m.showCursor {
		if pos < len(value) {
			// Cursor on a character - show it inverted
			result.WriteString("\x1b[7m")
			result.WriteRune(value[pos])
			result.WriteString("\x1b[0m")
			// Text after cursor
			if pos+1 < len(value) {
				result.WriteString(string(value[pos+1:]))
			}
		} else {
			// Cursor at end - show space inverted
			result.WriteString("\x1b[7m \x1b[0m")
		}
	} else {
		// No cursor - just show remaining text
		if pos < len(value) {
			result.WriteString(string(value[pos:]))
		}
	}

	// Padding if width is set
	if m.Width > 0 {
		currentWidth := len(value)
		if m.focus && m.showCursor && pos >= len(value) {
			currentWidth++ // Account for cursor space
		}
		if currentWidth < m.Width {
			result.WriteString(strings.Repeat(" ", m.Width-currentWidth))
		}
	}

	return result.String()
}

// deleteWordBackward deletes the word left to the cursor.
func (m *Model) deleteWordBackward() {
	if m.pos == 0 || len(m.value) == 0 {
		return
	}

	oldPos := m.pos

	// Move back past whitespace
	m.SetCursor(m.pos - 1)
	for m.pos > 0 && unicode.IsSpace(m.value[m.pos]) {
		m.SetCursor(m.pos - 1)
	}

	// Move back past word
	for m.pos > 0 && !unicode.IsSpace(m.value[m.pos]) {
		m.SetCursor(m.pos - 1)
	}

	// Keep one space if we're not at the start
	if m.pos > 0 {
		m.SetCursor(m.pos + 1)
	}

	// Delete from current position to old position
	if oldPos > len(m.value) {
		m.value = m.value[:m.pos]
	} else {
		m.value = append(m.value[:m.pos], m.value[oldPos:]...)
	}
}

// deleteWordForward deletes the word right to the cursor.
func (m *Model) deleteWordForward() {
	if m.pos >= len(m.value) || len(m.value) == 0 {
		return
	}

	oldPos := m.pos

	// Move forward past whitespace
	newPos := m.pos + 1
	for newPos < len(m.value) && unicode.IsSpace(m.value[newPos]) {
		newPos++
	}

	// Move forward past word
	for newPos < len(m.value) && !unicode.IsSpace(m.value[newPos]) {
		newPos++
	}

	// Delete from old position to new position
	if newPos > len(m.value) {
		m.value = m.value[:oldPos]
	} else {
		m.value = append(m.value[:oldPos], m.value[newPos:]...)
	}

	m.SetCursor(oldPos)
}
