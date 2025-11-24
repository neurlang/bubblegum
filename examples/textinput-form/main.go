package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/bubblegum/components/textinput"
	"github.com/bubblegum/lib"
)

type model struct {
	nameInput  textinput.Model
	emailInput textinput.Model
	focused    int
	submitted  bool
	name       string
	email      string
}

func initialModel() model {
	// Create name input
	nameInput := textinput.New()
	nameInput.Placeholder = "Enter your name"
	nameInput.Focus()
	nameInput.CharLimit = 50

	// Create email input
	emailInput := textinput.New()
	emailInput.Placeholder = "Enter your email"
	emailInput.CharLimit = 100

	return model{
		nameInput:  nameInput,
		emailInput: emailInput,
		focused:    0,
	}
}

func (m model) Init() lib.Cmd {
	return nil
}

func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
	var cmd lib.Cmd

	switch msg := msg.(type) {
	case lib.KeyMsg:
		switch msg.Type {
		case lib.KeyEsc, lib.KeyCtrlC:
			return m, lib.Quit

		case lib.KeyEnter:
			if m.submitted {
				// Reset form
				return initialModel(), nil
			}
			// Submit form
			m.submitted = true
			m.name = m.nameInput.Value()
			m.email = m.emailInput.Value()
			return m, nil

		case lib.KeyTab:
			// Switch focus between inputs
			if m.submitted {
				return m, nil
			}
			m.focused = (m.focused + 1) % 2
			if m.focused == 0 {
				m.nameInput.Focus()
				m.emailInput.Blur()
			} else {
				m.nameInput.Blur()
				m.emailInput.Focus()
			}
			return m, nil
		}
	}

	// Update the focused input
	if !m.submitted {
		if m.focused == 0 {
			m.nameInput, cmd = m.nameInput.Update(msg)
		} else {
			m.emailInput, cmd = m.emailInput.Update(msg)
		}
	}

	return m, cmd
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString("Text Input Form Example\n")
	b.WriteString("========================\n\n")

	if m.submitted {
		b.WriteString("Form Submitted!\n\n")
		b.WriteString(fmt.Sprintf("Name:  %s\n", m.name))
		b.WriteString(fmt.Sprintf("Email: %s\n", m.email))
		b.WriteString("\nPress Enter to reset, Esc to quit")
	} else {
		b.WriteString("Name:\n")
		b.WriteString(m.nameInput.View())
		b.WriteString("\n\n")

		b.WriteString("Email:\n")
		b.WriteString(m.emailInput.View())
		b.WriteString("\n\n")

		b.WriteString("Tab: Switch field | Enter: Submit | Esc: Quit")
	}

	return b.String()
}

func main() {
	p := lib.NewProgram(
		initialModel(),
		lib.WithWindowTitle("Text Input Form"),
		lib.WithInitialSize(800, 600),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
