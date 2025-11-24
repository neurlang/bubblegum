// Package lib provides a Bubble Tea-compatible API for building GUI applications
// using the Wayland window library.
package lib

// Model represents the application state in The Elm Architecture pattern.
// This interface matches Bubble Tea's Model interface for compatibility.
type Model interface {
	// Init is called when the program starts and returns an optional initial command.
	Init() Cmd

	// Update is called when a message is received and returns the updated model
	// and an optional command to execute.
	Update(Msg) (Model, Cmd)

	// View renders the model as a string. The string may contain ANSI escape
	// sequences for styling (colors, bold, italic, etc.).
	View() string
}

// Msg represents an event in the system (keyboard input, mouse click, timer tick, etc.).
// Any type can be a message.
type Msg interface{}

// Cmd represents an asynchronous operation that produces messages.
// Commands are returned by Init and Update and executed by the runtime.
type Cmd func() Msg
