package lib

import (
	"context"
	"sync"

	"github.com/neurlang/wayland/window"
)

// Program manages the application lifecycle, window, and event loop.
type Program struct {
	model   Model
	display *window.Display
	window  *window.Window
	widget  *window.Widget

	msgChan  chan Msg
	cmdChan  chan Cmd
	quitChan chan struct{}

	ctx    context.Context
	cancel context.CancelFunc

	options ProgramOptions
	mu      sync.Mutex
}

// ProgramOptions configures the Program's appearance and behavior.
type ProgramOptions struct {
	// FontFamily specifies the font family to use for rendering text.
	FontFamily string

	// FontSize specifies the font size in points.
	FontSize int

	// InitialWidth specifies the initial window width in pixels.
	InitialWidth int32

	// InitialHeight specifies the initial window height in pixels.
	InitialHeight int32

	// WindowTitle specifies the text displayed in the window's title bar.
	WindowTitle string

	// FPS specifies the maximum frames per second for rendering.
	// A value of 0 means no limit.
	FPS int
}

// ProgramOption is a function that configures a Program.
type ProgramOption func(*ProgramOptions)

// WithFontFamily sets the font family for text rendering.
func WithFontFamily(family string) ProgramOption {
	return func(opts *ProgramOptions) {
		opts.FontFamily = family
	}
}

// WithFontSize sets the font size in points.
func WithFontSize(size int) ProgramOption {
	return func(opts *ProgramOptions) {
		opts.FontSize = size
	}
}

// WithInitialSize sets the initial window dimensions.
func WithInitialSize(width, height int32) ProgramOption {
	return func(opts *ProgramOptions) {
		opts.InitialWidth = width
		opts.InitialHeight = height
	}
}

// WithWindowTitle sets the window title.
func WithWindowTitle(title string) ProgramOption {
	return func(opts *ProgramOptions) {
		opts.WindowTitle = title
	}
}

// WithFPS sets the maximum frames per second for rendering.
func WithFPS(fps int) ProgramOption {
	return func(opts *ProgramOptions) {
		opts.FPS = fps
	}
}

// NewProgram creates a new Program with the given model and options.
// This function matches Bubble Tea's NewProgram API for compatibility.
func NewProgram(model Model, opts ...ProgramOption) *Program {
	options := ProgramOptions{
		FontFamily:    "Monospace",
		FontSize:      12,
		InitialWidth:  800,
		InitialHeight: 600,
		WindowTitle:   "BubbleGum Application",
		FPS:           60,
	}

	for _, opt := range opts {
		opt(&options)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Program{
		model:    model,
		msgChan:  make(chan Msg, 100),
		cmdChan:  make(chan Cmd, 100),
		quitChan: make(chan struct{}),
		ctx:      ctx,
		cancel:   cancel,
		options:  options,
	}
}

// Run starts the program and blocks until it exits.
// It returns the final model state and any error that occurred.
func (p *Program) Run() (Model, error) {
	// TODO: Implementation will be added in later tasks
	// This includes:
	// - Creating the Wayland display and window
	// - Setting up the event loop
	// - Calling Init() and executing the initial command
	// - Processing messages and calling Update()
	// - Calling View() and rendering output
	// - Handling window events
	return p.model, nil
}

// Send sends a message to the program's Update function.
// This is thread-safe and can be called from any goroutine.
func (p *Program) Send(msg Msg) {
	select {
	case p.msgChan <- msg:
	case <-p.ctx.Done():
	}
}

// Quit signals the program to exit gracefully.
func (p *Program) Quit() {
	p.cancel()
	close(p.quitChan)
}
