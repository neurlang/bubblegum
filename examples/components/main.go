package main

import (
	"fmt"

	"github.com/bubblegum/components/list"
	"github.com/bubblegum/components/spinner"
	"github.com/bubblegum/components/textinput"
	"github.com/bubblegum/components/viewport"
	"github.com/bubblegum/lib"
)

// Model demonstrates using multiple BubbleGum components.
type Model struct {
	textInput textinput.Model
	spinner   spinner.Model
	list      list.Model
	viewport  viewport.Model
	activeTab int
}

func initialModel() Model {
	// Text input
	ti := textinput.New()
	ti.Placeholder = "Type something..."
	ti.Focus()

	// Spinner
	sp := spinner.New(spinner.WithSpinner(spinner.Dot))

	// List
	items := []list.Item{
		list.NewDefaultItem("Item 1", "First item"),
		list.NewDefaultItem("Item 2", "Second item"),
		list.NewDefaultItem("Item 3", "Third item"),
		list.NewDefaultItem("Item 4", "Fourth item"),
		list.NewDefaultItem("Item 5", "Fifth item"),
	}
	l := list.New(items, 40, 10)
	l.Title = "My List"

	// Viewport
	vp := viewport.New(40, 10)
	vp.SetContent("This is a viewport.\nIt can scroll through content.\n\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10\nLine 11\nLine 12")

	return Model{
		textInput: ti,
		spinner:   sp,
		list:      l,
		viewport:  vp,
		activeTab: 0,
	}
}

func (m Model) Init() lib.Cmd {
	// Start the spinner
	return func() lib.Msg {
		return m.spinner.Tick()
	}
}

func (m Model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
	var cmd lib.Cmd
	var cmds []lib.Cmd

	switch msg := msg.(type) {
	case lib.KeyMsg:
		switch msg.Type {
		case lib.KeyCtrlC:
			return m, lib.Quit

		case lib.KeyTab:
			m.activeTab = (m.activeTab + 1) % 4

		case lib.KeyEsc:
			return m, lib.Quit
		}

	case lib.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height/2)
		m.viewport.SetSize(msg.Width, msg.Height/2)
	}

	// Update active component
	switch m.activeTab {
	case 0:
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
	case 1:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	case 2:
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	case 3:
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, lib.Batch(cmds...)
}

func (m Model) View() string {
	tabs := []string{"TextInput", "Spinner", "List", "Viewport"}
	
	var view string
	view += "BubbleGum Components Demo\n"
	view += "Tab: Switch | Esc/Ctrl+C: Quit\n\n"
	
	// Show tabs
	for i, tab := range tabs {
		if i == m.activeTab {
			view += fmt.Sprintf("[%s] ", tab)
		} else {
			view += fmt.Sprintf(" %s  ", tab)
		}
	}
	view += "\n\n"

	// Show active component
	switch m.activeTab {
	case 0:
		view += "Text Input Component:\n"
		view += m.textInput.View()
		view += fmt.Sprintf("\nValue: %s", m.textInput.Value())
	case 1:
		view += "Spinner Component:\n"
		view += m.spinner.View() + " Loading..."
	case 2:
		view += "List Component:\n"
		view += m.list.View()
	case 3:
		view += "Viewport Component:\n"
		view += m.viewport.View()
	}

	return view
}

func main() {
	p := lib.NewProgram(
		initialModel(),
		lib.WithWindowTitle("BubbleGum Components Demo"),
		lib.WithInitialSize(800, 600),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
