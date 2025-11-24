package main

import (
	"fmt"
	"os"
	"time"

	"github.com/bubblegum/lib"
)

type tickMsg time.Time

type model struct {
	elapsed  time.Duration
	running  bool
	startTime time.Time
}

func initialModel() model {
	return model{
		elapsed: 0,
		running: false,
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
			// Toggle timer
			if m.running {
				// Stop timer
				m.running = false
				return m, nil
			} else {
				// Start timer
				m.running = true
				m.startTime = time.Now()
				// Start tick command
				return m, lib.Tick(100*time.Millisecond, func(t time.Time) lib.Msg {
					return tickMsg(t)
				})
			}

		case lib.KeyBackspace:
			// Reset timer
			m.elapsed = 0
			m.running = false
			return m, nil
		}

	case tickMsg:
		if m.running {
			// Update elapsed time
			m.elapsed = time.Since(m.startTime)
			// Schedule next tick
			return m, lib.Tick(100*time.Millisecond, func(t time.Time) lib.Msg {
				return tickMsg(t)
			})
		}
	}

	return m, nil
}

func (m model) View() string {
	status := "Stopped"
	if m.running {
		status = "Running"
	}

	// Format elapsed time
	hours := int(m.elapsed.Hours())
	minutes := int(m.elapsed.Minutes()) % 60
	seconds := int(m.elapsed.Seconds()) % 60
	millis := int(m.elapsed.Milliseconds()) % 1000

	return fmt.Sprintf(`Timer Example
=============

Status: %s

Elapsed Time: %02d:%02d:%02d.%03d

Controls:
  Enter      - Start/Stop
  Backspace  - Reset
  Esc        - Quit

This example demonstrates the Tick command for
periodic updates and timer functionality.
`, status, hours, minutes, seconds, millis)
}

func main() {
	p := lib.NewProgram(
		initialModel(),
		lib.WithWindowTitle("Timer Example"),
		lib.WithInitialSize(800, 600),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
