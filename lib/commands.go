package lib

import (
	"context"
	"fmt"
	"sync"
	"time"
)

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

// Tick creates a command that waits for the specified duration and then sends a message.
// This matches Bubble Tea's Tick command for compatibility.
func Tick(d time.Duration, fn func(time.Time) Msg) Cmd {
	return func() Msg {
		time.Sleep(d)
		return fn(time.Now())
	}
}

// Every creates a command that sends messages at regular intervals.
// The returned function can be called to cancel the timer.
func Every(d time.Duration, fn func(time.Time) Msg) Cmd {
	return func() Msg {
		return everyMsg{
			duration: d,
			fn:       fn,
		}
	}
}

// everyMsg is the internal message type for recurring timer commands.
type everyMsg struct {
	duration time.Duration
	fn       func(time.Time) Msg
}

// CommandExecutor manages asynchronous command execution.
// It executes commands in separate goroutines and delivers their messages
// to the program's message channel in a thread-safe manner.
type CommandExecutor struct {
	msgChan chan Msg
	ctx     context.Context
	wg      sync.WaitGroup
	mu      sync.Mutex
	timers  map[*time.Ticker]context.CancelFunc
}

// NewCommandExecutor creates a new CommandExecutor that delivers messages to the given channel.
func NewCommandExecutor(ctx context.Context, msgChan chan Msg) *CommandExecutor {
	return &CommandExecutor{
		msgChan: msgChan,
		ctx:     ctx,
		timers:  make(map[*time.Ticker]context.CancelFunc),
	}
}

// Execute runs a command asynchronously and delivers its message to the message channel.
// If the command is nil, this is a no-op.
func (ce *CommandExecutor) Execute(cmd Cmd) {
	if cmd == nil {
		return
	}

	ce.wg.Add(1)
	go func() {
		defer ce.wg.Done()
		defer func() {
			// Recover from panics in command execution
			if r := recover(); r != nil {
				Error("Command panicked: %v", r)
				// Deliver error message to Update
				ce.deliverMessage(ErrorMsg{Err: fmt.Errorf("command panic: %v", r)})
			}
		}()

		Debug("Executing command")

		// Execute the command and get the resulting message
		msg := cmd()

		// Handle special message types
		switch m := msg.(type) {
		case batchMsg:
			// Execute all batched commands
			ce.ExecuteBatch(m.cmds)
		case everyMsg:
			// Start a recurring timer
			ce.startTimer(m.duration, m.fn)
		default:
			// Deliver the message to the channel
			ce.deliverMessage(msg)
		}
	}()
}

// ExecuteBatch executes multiple commands concurrently and delivers all their messages.
func (ce *CommandExecutor) ExecuteBatch(cmds []Cmd) {
	for _, cmd := range cmds {
		if cmd != nil {
			ce.Execute(cmd)
		}
	}
}

// deliverMessage sends a message to the message channel in a thread-safe manner.
// It respects the context cancellation to avoid blocking on a closed channel.
func (ce *CommandExecutor) deliverMessage(msg Msg) {
	if msg == nil {
		Debug("Skipping nil message delivery")
		return
	}

	Debug("Delivering message: %T", msg)

	select {
	case ce.msgChan <- msg:
		Debug("Message delivered successfully")
	case <-ce.ctx.Done():
		Debug("Context cancelled, message not delivered")
	}
}

// startTimer creates a recurring timer that sends messages at regular intervals.
func (ce *CommandExecutor) startTimer(d time.Duration, fn func(time.Time) Msg) {
	Debug("Starting timer with duration: %v", d)
	ticker := time.NewTicker(d)
	timerCtx, cancel := context.WithCancel(ce.ctx)

	ce.mu.Lock()
	ce.timers[ticker] = cancel
	ce.mu.Unlock()

	ce.wg.Add(1)
	go func() {
		defer ce.wg.Done()
		defer ticker.Stop()
		defer func() {
			ce.mu.Lock()
			delete(ce.timers, ticker)
			ce.mu.Unlock()
			Debug("Timer stopped")
		}()

		for {
			select {
			case t := <-ticker.C:
				Debug("Timer tick at %v", t)
				msg := fn(t)
				ce.deliverMessage(msg)
			case <-timerCtx.Done():
				Debug("Timer context cancelled")
				return
			case <-ce.ctx.Done():
				Debug("Command executor context cancelled")
				return
			}
		}
	}()
}

// Shutdown stops all running commands and waits for them to complete.
// It cancels all recurring timers and waits for all goroutines to finish.
func (ce *CommandExecutor) Shutdown() {
	Debug("Shutting down command executor")
	
	// Cancel all timers
	ce.mu.Lock()
	timerCount := len(ce.timers)
	for _, cancel := range ce.timers {
		cancel()
	}
	ce.mu.Unlock()

	Debug("Cancelled %d timers", timerCount)

	// Wait for all goroutines to finish
	ce.wg.Wait()
	
	Debug("Command executor shutdown complete")
}
