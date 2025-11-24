package main

import (
	"fmt"
	"os"

	"github.com/bubblegum/components/list"
	"github.com/bubblegum/lib"
)

type model struct {
	list         list.Model
	selectedItem string
	quitting     bool
}

func initialModel() model {
	// Create a list of programming languages
	items := []list.Item{
		list.NewDefaultItem("Go", "A statically typed, compiled language designed at Google"),
		list.NewDefaultItem("Python", "A high-level, interpreted programming language"),
		list.NewDefaultItem("JavaScript", "The programming language of the web"),
		list.NewDefaultItem("Rust", "A systems programming language focused on safety"),
		list.NewDefaultItem("TypeScript", "JavaScript with syntax for types"),
		list.NewDefaultItem("Java", "A class-based, object-oriented programming language"),
		list.NewDefaultItem("C++", "A general-purpose programming language"),
		list.NewDefaultItem("C#", "A modern, object-oriented language from Microsoft"),
		list.NewDefaultItem("Ruby", "A dynamic, open source programming language"),
		list.NewDefaultItem("Swift", "A powerful language for iOS, macOS, and more"),
		list.NewDefaultItem("Kotlin", "A modern programming language for Android"),
		list.NewDefaultItem("PHP", "A popular general-purpose scripting language"),
		list.NewDefaultItem("Scala", "A language that combines object-oriented and functional programming"),
		list.NewDefaultItem("Haskell", "A purely functional programming language"),
		list.NewDefaultItem("Elixir", "A dynamic, functional language for building scalable applications"),
		list.NewDefaultItem("Clojure", "A modern, functional dialect of Lisp"),
		list.NewDefaultItem("Dart", "A client-optimized language for fast apps"),
		list.NewDefaultItem("Lua", "A lightweight, embeddable scripting language"),
		list.NewDefaultItem("Perl", "A highly capable, feature-rich programming language"),
		list.NewDefaultItem("R", "A language for statistical computing and graphics"),
	}

	l := list.New(items, 80, 20)
	l.Title = "Programming Languages Browser"

	return model{
		list: l,
	}
}

func (m model) Init() lib.Cmd {
	return nil
}

func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
	switch msg := msg.(type) {
	case lib.KeyMsg:
		// Handle global keys
		switch msg.Type {
		case lib.KeyEsc, lib.KeyCtrlC:
			m.quitting = true
			return m, lib.Quit

		case lib.KeyEnter:
			// Select the current item
			item := m.list.SelectedItem()
			if item != nil {
				if defaultItem, ok := item.(list.DefaultItem); ok {
					m.selectedItem = defaultItem.Title()
				}
			}
		}

	case lib.WindowSizeMsg:
		// Update list size
		m.list.SetSize(msg.Width, msg.Height-5) // Reserve space for header and footer
	}

	// Update the list
	var cmd lib.Cmd
	m.list, cmd = m.list.Update(msg)

	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return "Thanks for browsing!\n"
	}

	s := "List Browser Example\n"
	s += "====================\n\n"

	// Show the list
	s += m.list.View()

	s += "\n\n"

	// Show selected item
	if m.selectedItem != "" {
		s += fmt.Sprintf("Selected: %s\n", m.selectedItem)
	}

	// Show controls
	s += "\nControls:\n"
	s += "  ↑/↓       - Navigate\n"
	s += "  PgUp/PgDn - Page up/down\n"
	s += "  Home/End  - Jump to start/end\n"
	s += "  /         - Start filtering\n"
	s += "  Enter     - Select item\n"
	s += "  Esc       - Quit\n"

	return s
}

func main() {
	p := lib.NewProgram(
		initialModel(),
		lib.WithWindowTitle("List Browser"),
		lib.WithInitialSize(900, 700),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
