// Package spinner provides a spinner component for BubbleGum applications.
package spinner

import (
	"sync/atomic"
	"time"

	"github.com/neurlang/bubblegum/lib"
)

// Internal ID management for routing messages.
var lastID int64

func nextID() int {
	return int(atomic.AddInt64(&lastID, 1))
}

// Spinner is a set of frames used in animating the spinner.
type Spinner struct {
	Frames []string
	FPS    time.Duration
}

// Predefined spinners.
var (
	Line = Spinner{
		Frames: []string{"|", "/", "-", "\\"},
		FPS:    time.Second / 10,
	}
	Dot = Spinner{
		Frames: []string{"⣾ ", "⣽ ", "⣻ ", "⢿ ", "⡿ ", "⣟ ", "⣯ ", "⣷ "},
		FPS:    time.Second / 10,
	}
	MiniDot = Spinner{
		Frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		FPS:    time.Second / 12,
	}
	Jump = Spinner{
		Frames: []string{"⢄", "⢂", "⢁", "⡁", "⡈", "⡐", "⡠"},
		FPS:    time.Second / 10,
	}
	Pulse = Spinner{
		Frames: []string{"█", "▓", "▒", "░"},
		FPS:    time.Second / 8,
	}
	Points = Spinner{
		Frames: []string{"∙∙∙", "●∙∙", "∙●∙", "∙∙●"},
		FPS:    time.Second / 7,
	}
	Globe = Spinner{
		Frames: []string{"🌍", "🌎", "🌏"},
		FPS:    time.Second / 4,
	}
	Moon = Spinner{
		Frames: []string{"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘"},
		FPS:    time.Second / 8,
	}
	Monkey = Spinner{
		Frames: []string{"🙈", "🙉", "🙊"},
		FPS:    time.Second / 3,
	}
	Meter = Spinner{
		Frames: []string{
			"▱▱▱",
			"▰▱▱",
			"▰▰▱",
			"▰▰▰",
			"▰▰▱",
			"▰▱▱",
			"▱▱▱",
		},
		FPS: time.Second / 7,
	}
	Hamburger = Spinner{
		Frames: []string{"☱", "☲", "☴", "☲"},
		FPS:    time.Second / 3,
	}
	Ellipsis = Spinner{
		Frames: []string{"", ".", "..", "..."},
		FPS:    time.Second / 3,
	}
)

// Model contains the state for the spinner.
type Model struct {
	// Spinner settings to use.
	Spinner Spinner

	frame int
	id    int
	tag   int
}

// ID returns the spinner's unique ID.
func (m Model) ID() int {
	return m.id
}

// New returns a model with default values.
func New(opts ...Option) Model {
	m := Model{
		Spinner: Line,
		id:      nextID(),
	}

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

// TickMsg indicates that the timer has ticked and we should render a frame.
type TickMsg struct {
	Time time.Time
	tag  int
	ID   int
}

// Update is the update function for the spinner.
func (m Model) Update(msg lib.Msg) (Model, lib.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		// If an ID is set, and the ID doesn't belong to this spinner, reject
		// the message.
		if msg.ID > 0 && msg.ID != m.id {
			return m, nil
		}

		// If a tag is set, and it's not the one we expect, reject the message.
		if msg.tag > 0 && msg.tag != m.tag {
			return m, nil
		}

		m.frame++
		if m.frame >= len(m.Spinner.Frames) {
			m.frame = 0
		}

		m.tag++
		return m, m.tick(m.id, m.tag)
	default:
		return m, nil
	}
}

// View renders the model's view.
func (m Model) View() string {
	if m.frame >= len(m.Spinner.Frames) {
		return "(error)"
	}

	return m.Spinner.Frames[m.frame]
}

// Tick is the command used to advance the spinner one frame.
func (m Model) Tick() lib.Msg {
	return TickMsg{
		Time: time.Now(),
		ID:   m.id,
		tag:  m.tag,
	}
}

func (m Model) tick(id, tag int) lib.Cmd {
	return func() lib.Msg {
		time.Sleep(m.Spinner.FPS)
		return TickMsg{
			Time: time.Now(),
			ID:   id,
			tag:  tag,
		}
	}
}

// Option is used to set options in New.
type Option func(*Model)

// WithSpinner is an option to set the spinner.
func WithSpinner(spinner Spinner) Option {
	return func(m *Model) {
		m.Spinner = spinner
	}
}
