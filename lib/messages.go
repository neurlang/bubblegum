package lib

import "fmt"

// KeyType represents the type of key that was pressed.
type KeyType int

const (
	KeyRunes KeyType = iota
	KeyEnter
	KeyBackspace
	KeyTab
	KeyEsc
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyHome
	KeyEnd
	KeyPgUp
	KeyPgDown
	KeyDelete
	KeyInsert
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
	KeyCtrlC
	KeyCtrlD
	KeyCtrlL
	KeyCtrlZ
)

// KeyMsg represents a keyboard input event.
type KeyMsg struct {
	Type  KeyType
	Runes []rune
	Alt   bool
}

// String returns a string representation of the key message for debugging.
func (k KeyMsg) String() string {
	if k.Type == KeyRunes {
		return fmt.Sprintf("KeyMsg{Runes: %q, Alt: %v}", string(k.Runes), k.Alt)
	}
	return fmt.Sprintf("KeyMsg{Type: %v, Alt: %v}", k.Type, k.Alt)
}

// MouseEventType represents the type of mouse event.
type MouseEventType int

const (
	MousePress MouseEventType = iota
	MouseRelease
	MouseMotion
	MouseWheel
)

// MouseButton represents a mouse button.
type MouseButton int

const (
	MouseButtonNone MouseButton = iota
	MouseButtonLeft
	MouseButtonMiddle
	MouseButtonRight
	MouseButtonWheelUp
	MouseButtonWheelDown
	MouseButtonWheelLeft
	MouseButtonWheelRight
)

// MouseMsg represents a mouse input event.
type MouseMsg struct {
	X      int
	Y      int
	Type   MouseEventType
	Button MouseButton
}

// String returns a string representation of the mouse message for debugging.
func (m MouseMsg) String() string {
	return fmt.Sprintf("MouseMsg{X: %d, Y: %d, Type: %v, Button: %v}", m.X, m.Y, m.Type, m.Button)
}

// WindowSizeMsg represents a window resize event.
type WindowSizeMsg struct {
	Width  int
	Height int
}

// String returns a string representation of the window size message for debugging.
func (w WindowSizeMsg) String() string {
	return fmt.Sprintf("WindowSizeMsg{Width: %d, Height: %d}", w.Width, w.Height)
}

// QuitMsg represents a termination signal for the application.
type QuitMsg struct{}

// String returns a string representation of the quit message for debugging.
func (q QuitMsg) String() string {
	return "QuitMsg{}"
}
