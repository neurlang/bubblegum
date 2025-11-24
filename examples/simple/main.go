package main

import (
	"fmt"
	"os"

	"github.com/neurlang/bubblegum/lib"
)

type model struct {
	count int
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
		case lib.KeyUp:
			m.count++
		case lib.KeyDown:
			m.count--
		}
	case lib.WindowSizeMsg:
		// Window was resized
	}
	return m, nil
}

func (m model) View() string {
	return fmt.Sprintf("Counter: %d\n\nPress Up/Down to change, Esc to quit", m.count)
}

func main() {
	p := lib.NewProgram(
		model{},
		lib.WithWindowTitle("Simple Counter"),
		lib.WithInitialSize(800, 600),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
