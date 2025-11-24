// Package list provides a list component for BubbleGum applications.
package list

import (
	"fmt"
	"strings"

	"github.com/neurlang/bubblegum/lib"
)

// Item is an item that appears in the list.
type Item interface {
	// FilterValue is the value we use when filtering against this item.
	FilterValue() string
}

// DefaultItem is a simple implementation of Item.
type DefaultItem struct {
	title       string
	description string
}

// NewDefaultItem creates a new default item.
func NewDefaultItem(title, description string) DefaultItem {
	return DefaultItem{
		title:       title,
		description: description,
	}
}

// FilterValue implements Item.
func (d DefaultItem) FilterValue() string {
	return d.title
}

// Title returns the item's title.
func (d DefaultItem) Title() string {
	return d.title
}

// Description returns the item's description.
func (d DefaultItem) Description() string {
	return d.description
}

// Model contains the state of the list component.
type Model struct {
	// Title is displayed at the top of the list.
	Title string

	// Width and Height define the list dimensions.
	Width  int
	Height int

	// items is the master set of items.
	items []Item

	// cursor is the currently selected item index.
	cursor int

	// offset is the scroll offset for viewing items.
	offset int

	// filterValue is the current filter text.
	filterValue string

	// filtering indicates whether the user is currently filtering.
	filtering bool

	// filteredItems contains items matching the filter.
	filteredItems []Item
}

// New returns a new list model.
func New(items []Item, width, height int) Model {
	return Model{
		Title:         "List",
		Width:         width,
		Height:        height,
		items:         items,
		cursor:        0,
		offset:        0,
		filtering:     false,
		filteredItems: nil,
	}
}

// SetItems sets the items in the list.
func (m *Model) SetItems(items []Item) {
	m.items = items
	m.cursor = 0
	m.offset = 0
	if m.filtering {
		m.updateFilter()
	}
}

// Items returns the items in the list.
func (m Model) Items() []Item {
	return m.items
}

// VisibleItems returns the items currently visible (filtered or all).
func (m Model) VisibleItems() []Item {
	if m.filtering && m.filteredItems != nil {
		return m.filteredItems
	}
	return m.items
}

// SelectedItem returns the currently selected item.
func (m Model) SelectedItem() Item {
	items := m.VisibleItems()
	if m.cursor < 0 || m.cursor >= len(items) {
		return nil
	}
	return items[m.cursor]
}

// Index returns the index of the currently selected item.
func (m Model) Index() int {
	return m.cursor
}

// SetCursor sets the cursor position.
func (m *Model) SetCursor(index int) {
	items := m.VisibleItems()
	if index < 0 {
		index = 0
	}
	if index >= len(items) {
		index = len(items) - 1
	}
	m.cursor = index
	m.adjustOffset()
}

// CursorUp moves the cursor up one item.
func (m *Model) CursorUp() {
	if m.cursor > 0 {
		m.cursor--
		m.adjustOffset()
	}
}

// CursorDown moves the cursor down one item.
func (m *Model) CursorDown() {
	items := m.VisibleItems()
	if m.cursor < len(items)-1 {
		m.cursor++
		m.adjustOffset()
	}
}

// adjustOffset adjusts the scroll offset to keep the cursor visible.
func (m *Model) adjustOffset() {
	visibleHeight := m.Height - 3 // Reserve space for title and status

	// Scroll up if cursor is above visible area
	if m.cursor < m.offset {
		m.offset = m.cursor
	}

	// Scroll down if cursor is below visible area
	if m.cursor >= m.offset+visibleHeight {
		m.offset = m.cursor - visibleHeight + 1
	}
}

// SetSize sets the width and height of the list.
func (m *Model) SetSize(width, height int) {
	m.Width = width
	m.Height = height
	m.adjustOffset()
}

// StartFiltering starts the filtering mode.
func (m *Model) StartFiltering() {
	m.filtering = true
	m.filterValue = ""
	m.updateFilter()
}

// StopFiltering stops the filtering mode.
func (m *Model) StopFiltering() {
	m.filtering = false
	m.filterValue = ""
	m.filteredItems = nil
	m.cursor = 0
	m.offset = 0
}

// SetFilter sets the filter value and updates filtered items.
func (m *Model) SetFilter(value string) {
	m.filterValue = value
	m.updateFilter()
}

// updateFilter updates the filtered items based on the current filter value.
func (m *Model) updateFilter() {
	if m.filterValue == "" {
		m.filteredItems = m.items
		return
	}

	filtered := make([]Item, 0)
	filterLower := strings.ToLower(m.filterValue)

	for _, item := range m.items {
		if strings.Contains(strings.ToLower(item.FilterValue()), filterLower) {
			filtered = append(filtered, item)
		}
	}

	m.filteredItems = filtered
	m.cursor = 0
	m.offset = 0
}

// Update is the update loop for the list.
func (m Model) Update(msg lib.Msg) (Model, lib.Cmd) {
	switch msg := msg.(type) {
	case lib.KeyMsg:
		if m.filtering {
			return m.handleFilteringKeys(msg)
		}
		return m.handleBrowsingKeys(msg)

	case lib.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
	}

	return m, nil
}

// handleBrowsingKeys handles keys when browsing the list.
func (m Model) handleBrowsingKeys(msg lib.KeyMsg) (Model, lib.Cmd) {
	switch msg.Type {
	case lib.KeyUp:
		m.CursorUp()

	case lib.KeyDown:
		m.CursorDown()

	case lib.KeyPgUp:
		// Move up by visible height
		visibleHeight := m.Height - 3
		for i := 0; i < visibleHeight && m.cursor > 0; i++ {
			m.cursor--
		}
		m.adjustOffset()

	case lib.KeyPgDown:
		// Move down by visible height
		visibleHeight := m.Height - 3
		items := m.VisibleItems()
		for i := 0; i < visibleHeight && m.cursor < len(items)-1; i++ {
			m.cursor++
		}
		m.adjustOffset()

	case lib.KeyHome:
		m.cursor = 0
		m.offset = 0

	case lib.KeyEnd:
		items := m.VisibleItems()
		m.cursor = len(items) - 1
		m.adjustOffset()

	case lib.KeyRunes:
		// Start filtering if '/' is pressed
		if len(msg.Runes) == 1 && msg.Runes[0] == '/' {
			m.StartFiltering()
		}
	}

	return m, nil
}

// handleFilteringKeys handles keys when in filtering mode.
func (m Model) handleFilteringKeys(msg lib.KeyMsg) (Model, lib.Cmd) {
	switch msg.Type {
	case lib.KeyEsc:
		m.StopFiltering()

	case lib.KeyEnter:
		m.filtering = false

	case lib.KeyBackspace:
		if len(m.filterValue) > 0 {
			m.filterValue = m.filterValue[:len(m.filterValue)-1]
			m.updateFilter()
		}

	case lib.KeyRunes:
		m.filterValue += string(msg.Runes)
		m.updateFilter()
	}

	return m, nil
}

// View renders the list.
func (m Model) View() string {
	var b strings.Builder

	// Title
	b.WriteString(m.Title)
	b.WriteString("\n")

	// Filter indicator
	if m.filtering {
		b.WriteString("Filter: ")
		b.WriteString(m.filterValue)
		b.WriteString("_\n")
	} else {
		b.WriteString("\n")
	}

	// Items
	items := m.VisibleItems()
	visibleHeight := m.Height - 3

	for i := m.offset; i < m.offset+visibleHeight && i < len(items); i++ {
		if i == m.cursor {
			b.WriteString("\x1b[7m> ") // Inverted
		} else {
			b.WriteString("  ")
		}

		// Render item
		if defaultItem, ok := items[i].(DefaultItem); ok {
			b.WriteString(defaultItem.Title())
			if defaultItem.Description() != "" {
				b.WriteString(" - ")
				b.WriteString(defaultItem.Description())
			}
		} else {
			b.WriteString(items[i].FilterValue())
		}

		if i == m.cursor {
			b.WriteString("\x1b[0m") // Reset
		}
		b.WriteString("\n")
	}

	// Status bar
	b.WriteString(fmt.Sprintf("\n%d/%d items", m.cursor+1, len(items)))

	return b.String()
}
