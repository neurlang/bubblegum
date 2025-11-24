package main

import (
	"fmt"
	"os"

	"github.com/bubblegum/lib"
)

type model struct {
	eventCount int
	lastEvent  string
}

func (m model) Init() lib.Cmd {
	return nil
}

func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
	switch msg := msg.(type) {
	case lib.KeyMsg:
		if msg.Type == lib.KeyEsc || msg.Type == lib.KeyCtrlC {
			return m, lib.Quit
		}
		m.eventCount++
		m.lastEvent = fmt.Sprintf("Key: %v", msg.Type)

	case lib.MouseMsg:
		m.eventCount++
		m.lastEvent = fmt.Sprintf("Mouse: type=%v, pos=(%d,%d), button=%v", 
			msg.Type, msg.X, msg.Y, msg.Button)
	}

	return m, nil
}

func (m model) View() string {
	return fmt.Sprintf(`Simple Mouse Test
==================

Event Count: %d
Last Event: %s

Move mouse or click to test
Press Esc to quit
`, m.eventCount, m.lastEvent)
}

func main() {
	p := lib.NewProgram(
		model{lastEvent: "None"},
		lib.WithWindowTitle("Simple Mouse Test"),
		lib.WithInitialSize(600, 400),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
