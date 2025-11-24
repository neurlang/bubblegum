package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/bubblegum/lib"
)

type clickEvent struct {
	x      int
	y      int
	button lib.MouseButton
}

type model struct {
	mouseX      int
	mouseY      int
	clicks      []clickEvent
	lastButton  lib.MouseButton
	windowWidth int
	windowHeight int
}

func initialModel() model {
	return model{
		clicks: make([]clickEvent, 0),
	}
}

func (m model) Init() lib.Cmd {
	return nil
}

func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
	switch msg := msg.(type) {
	case lib.KeyMsg:
		switch msg.Type {
		case lib.KeyEsc, lib.KeyCtrlC:
			return m, lib.Quit
		case lib.KeyEnter:
			// Clear click history
			m.clicks = make([]clickEvent, 0)
			m.lastButton = lib.MouseButtonNone
		}

	case lib.MouseMsg:
		switch msg.Type {
		case lib.MouseMotion:
			// Update mouse position (throttled by non-blocking send)
			m.mouseX = msg.X
			m.mouseY = msg.Y

		case lib.MousePress:
			// Record click
			m.lastButton = msg.Button
			m.clicks = append(m.clicks, clickEvent{
				x:      msg.X,
				y:      msg.Y,
				button: msg.Button,
			})
			// Keep only last 10 clicks
			if len(m.clicks) > 10 {
				m.clicks = m.clicks[1:]
			}

		case lib.MouseWheel:
			// Record scroll as a click event
			m.clicks = append(m.clicks, clickEvent{
				x:      msg.X,
				y:      msg.Y,
				button: msg.Button,
			})
			// Keep only last 10 clicks
			if len(m.clicks) > 10 {
				m.clicks = m.clicks[1:]
			}
		}

	case lib.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
	}

	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString("Mouse Interactive Example\n")
	b.WriteString("==========================\n\n")

	b.WriteString(fmt.Sprintf("Window Size: %d x %d cells\n", m.windowWidth, m.windowHeight))
	b.WriteString(fmt.Sprintf("Mouse Position: (%d, %d)\n", m.mouseX, m.mouseY))
	
	if m.lastButton != lib.MouseButtonNone {
		b.WriteString(fmt.Sprintf("Last Button: %s\n", buttonName(m.lastButton)))
	}
	
	b.WriteString("\n")

	if len(m.clicks) > 0 {
		b.WriteString("Recent Clicks/Scrolls:\n")
		for i := len(m.clicks) - 1; i >= 0; i-- {
			click := m.clicks[i]
			b.WriteString(fmt.Sprintf("  %d. (%d, %d) - %s\n", 
				len(m.clicks)-i, click.x, click.y, buttonName(click.button)))
		}
	} else {
		b.WriteString("No clicks yet. Click anywhere!\n")
	}

	b.WriteString("\n")
	b.WriteString("Move mouse to see position\n")
	b.WriteString("Click to record position\n")
	b.WriteString("Scroll to record scroll events\n")
	b.WriteString("Enter: Clear history | Esc: Quit")

	return b.String()
}

func buttonName(button lib.MouseButton) string {
	switch button {
	case lib.MouseButtonLeft:
		return "Left Click"
	case lib.MouseButtonMiddle:
		return "Middle Click"
	case lib.MouseButtonRight:
		return "Right Click"
	case lib.MouseButtonWheelUp:
		return "Scroll Up"
	case lib.MouseButtonWheelDown:
		return "Scroll Down"
	case lib.MouseButtonWheelLeft:
		return "Scroll Left"
	case lib.MouseButtonWheelRight:
		return "Scroll Right"
	default:
		return "Unknown"
	}
}

func main() {
	p := lib.NewProgram(
		initialModel(),
		lib.WithWindowTitle("Mouse Interactive"),
		lib.WithInitialSize(800, 600),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
