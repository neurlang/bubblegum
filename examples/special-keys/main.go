package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/neurlang/bubblegum/lib"
)

type model struct {
	lastKey     string
	keyHistory  []string
	maxHistory  int
}

func (m model) Init() lib.Cmd {
	return nil
}

func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
	switch msg := msg.(type) {
	case lib.KeyMsg:
		keyName := getKeyName(msg.Type)
		
		// Handle quit keys
		if msg.Type == lib.KeyEsc || msg.Type == lib.KeyCtrlC {
			return m, lib.Quit
		}
		
		// Update last key and history
		m.lastKey = keyName
		m.keyHistory = append([]string{keyName}, m.keyHistory...)
		if len(m.keyHistory) > m.maxHistory {
			m.keyHistory = m.keyHistory[:m.maxHistory]
		}
	}
	return m, nil
}

func (m model) View() string {
	var sb strings.Builder
	
	sb.WriteString("Special Keys Demo\n")
	sb.WriteString("=================\n\n")
	
	if m.lastKey != "" {
		sb.WriteString(fmt.Sprintf("Last Key Pressed: %s\n\n", m.lastKey))
	} else {
		sb.WriteString("Press any special key...\n\n")
	}
	
	sb.WriteString("Key History:\n")
	sb.WriteString("------------\n")
	for i, key := range m.keyHistory {
		sb.WriteString(fmt.Sprintf("%2d. %s\n", i+1, key))
	}
	
	sb.WriteString("\n\nSupported Keys:\n")
	sb.WriteString("  Arrow Keys: Up, Down, Left, Right\n")
	sb.WriteString("  Function Keys: F1-F12\n")
	sb.WriteString("  Navigation: Home, End, PageUp, PageDown\n")
	sb.WriteString("  Editing: Enter, Backspace, Tab, Delete, Insert\n")
	sb.WriteString("  Control: Ctrl+C, Ctrl+D, Ctrl+L, Ctrl+Z\n")
	sb.WriteString("\nPress Esc or Ctrl+C to quit")
	
	return sb.String()
}

func getKeyName(keyType lib.KeyType) string {
	switch keyType {
	case lib.KeyEnter:
		return "Enter"
	case lib.KeyBackspace:
		return "Backspace"
	case lib.KeyTab:
		return "Tab"
	case lib.KeyEsc:
		return "Escape"
	case lib.KeyUp:
		return "Up Arrow"
	case lib.KeyDown:
		return "Down Arrow"
	case lib.KeyLeft:
		return "Left Arrow"
	case lib.KeyRight:
		return "Right Arrow"
	case lib.KeyHome:
		return "Home"
	case lib.KeyEnd:
		return "End"
	case lib.KeyPgUp:
		return "Page Up"
	case lib.KeyPgDown:
		return "Page Down"
	case lib.KeyDelete:
		return "Delete"
	case lib.KeyInsert:
		return "Insert"
	case lib.KeyF1:
		return "F1"
	case lib.KeyF2:
		return "F2"
	case lib.KeyF3:
		return "F3"
	case lib.KeyF4:
		return "F4"
	case lib.KeyF5:
		return "F5"
	case lib.KeyF6:
		return "F6"
	case lib.KeyF7:
		return "F7"
	case lib.KeyF8:
		return "F8"
	case lib.KeyF9:
		return "F9"
	case lib.KeyF10:
		return "F10"
	case lib.KeyF11:
		return "F11"
	case lib.KeyF12:
		return "F12"
	case lib.KeyCtrlC:
		return "Ctrl+C"
	case lib.KeyCtrlD:
		return "Ctrl+D"
	case lib.KeyCtrlL:
		return "Ctrl+L"
	case lib.KeyCtrlZ:
		return "Ctrl+Z"
	default:
		return fmt.Sprintf("Unknown (%d)", keyType)
	}
}

func main() {
	p := lib.NewProgram(
		model{
			maxHistory: 10,
		},
		lib.WithWindowTitle("Special Keys Demo"),
		lib.WithInitialSize(800, 600),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
