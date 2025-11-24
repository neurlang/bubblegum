package lib

// Quit is a command that signals the program to exit.
// This matches Bubble Tea's Quit command for compatibility.
func Quit() Msg {
	return quitMsg{}
}

// quitMsg is the internal message type for quit signals.
type quitMsg struct{}

// Batch executes multiple commands concurrently and collects their messages.
// This matches Bubble Tea's Batch command for compatibility.
func Batch(cmds ...Cmd) Cmd {
	return func() Msg {
		return batchMsg{cmds: cmds}
	}
}

// batchMsg is the internal message type for batch command execution.
type batchMsg struct {
	cmds []Cmd
}
