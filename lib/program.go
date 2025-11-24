package lib

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/neurlang/wayland/window"
	"github.com/neurlang/wayland/wl"
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

	options  ProgramOptions
	mu       sync.Mutex
	renderer *Renderer
	cmdExec  *CommandExecutor

	lastView     string
	lastRender   time.Time
	windowWidth  int
	windowHeight int
	input        *window.Input
	pointerX     float32
	pointerY     float32
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
	// Create Wayland display
	display, err := window.DisplayCreate([]string{})
	if err != nil {
		return p.model, fmt.Errorf("failed to create display: %w", err)
	}
	p.display = display
	defer p.display.Destroy()

	// Create window
	p.window = window.Create(display)
	defer p.window.Destroy()

	// Set window title
	p.window.SetTitle(p.options.WindowTitle)

	// Set buffer type
	p.window.SetBufferType(window.BufferTypeShm)

	// Create widget
	p.widget = p.window.AddWidget(p)
	defer p.widget.Destroy()

	// Set up keyboard handler
	p.window.SetKeyboardHandler(p)

	// Create renderer
	p.renderer, err = NewRenderer(RendererOptions{
		DefaultFg: NewColor(255, 255, 255),
		DefaultBg: NewColor(0, 0, 0),
	})
	if err != nil {
		return p.model, fmt.Errorf("failed to create renderer: %w", err)
	}

	// Create command executor
	p.cmdExec = NewCommandExecutor(p.ctx, p.msgChan)
	defer p.cmdExec.Shutdown()

	// Create input handler (note: Input is created by the window system, not by us)
	// We'll get it from event handlers

	// Schedule initial resize
	p.widget.ScheduleResize(p.options.InitialWidth, p.options.InitialHeight)

	// Call model's Init() and execute initial command
	initialCmd := p.model.Init()
	if initialCmd != nil {
		p.cmdExec.Execute(initialCmd)
	}

	// Start message processing goroutine
	go p.processMessages()

	// Run the display event loop (blocks until quit)
	window.DisplayRun(display)

	return p.model, nil
}

// processMessages handles messages from the message channel.
func (p *Program) processMessages() {
	for {
		select {
		case msg := <-p.msgChan:
			p.handleMessage(msg)
		case <-p.ctx.Done():
			return
		case <-p.quitChan:
			return
		}
	}
}

// handleMessage processes a single message by calling Update and rendering.
func (p *Program) handleMessage(msg Msg) {
	// Check if this is a quit message
	if _, isQuit := msg.(quitMsg); isQuit {
		p.quit()
		return
	}

	// Call Update
	p.mu.Lock()
	var cmd Cmd
	p.model, cmd = p.model.Update(msg)
	p.mu.Unlock()

	// Execute the returned command
	if cmd != nil {
		p.cmdExec.Execute(cmd)
	}

	// Trigger a redraw
	if p.widget != nil {
		p.widget.ScheduleRedraw()
	}
}

// quit handles the quit process.
func (p *Program) quit() {
	p.cancel()
	if p.display != nil {
		p.display.Exit()
	}
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

// Resize implements window.WidgetHandler interface.
// It handles window resize events and sends WindowSizeMsg.
func (p *Program) Resize(widget *window.Widget, width int32, height int32, pwidth int32, pheight int32) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Update widget allocation
	if width != pwidth || height != pheight {
		widget.SetAllocation(0, 0, pwidth, pheight)
	}

	// Calculate grid dimensions based on cell size
	cellWidth := p.renderer.CellWidth()
	cellHeight := p.renderer.CellHeight()

	gridWidth := int(pwidth / cellWidth)
	gridHeight := int(pheight / cellHeight)

	// Store dimensions
	p.windowWidth = gridWidth
	p.windowHeight = gridHeight

	// Send WindowSizeMsg
	p.Send(WindowSizeMsg{
		Width:  gridWidth,
		Height: gridHeight,
	})
}

// Redraw implements window.WidgetHandler interface.
// It renders the current view to the window.
func (p *Program) Redraw(widget *window.Widget) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check frame rate limiting
	if p.options.FPS > 0 {
		minFrameTime := time.Second / time.Duration(p.options.FPS)
		elapsed := time.Since(p.lastRender)
		if elapsed < minFrameTime {
			// Skip this frame
			return
		}
	}

	// Get the current view
	view := p.model.View()

	// Skip rendering if view hasn't changed
	if view == p.lastView && p.lastView != "" {
		return
	}

	// Get the window surface
	surface := p.window.WindowGetSurface()
	if surface == nil {
		return
	}

	// Parse the view into a terminal grid
	grid := ParseANSI(view, p.windowWidth, p.windowHeight)
	if grid == nil {
		return
	}

	// Render the grid
	err := p.renderer.Render(grid, surface)
	if err != nil {
		// Log error but continue
		fmt.Printf("render error: %v\n", err)
		return
	}

	// Update state
	p.lastView = view
	p.lastRender = time.Now()
}

// Key implements window.KeyboardHandler interface.
// It handles keyboard input events.
func (p *Program) Key(
	win *window.Window,
	input *window.Input,
	time uint32,
	key uint32,
	notUnicode uint32,
	state wl.KeyboardKeyState,
	data window.WidgetHandler,
) {
	// Store input reference
	if p.input == nil {
		p.input = input
	}

	// Map the keyboard event to a KeyMsg
	keyMsg := MapKeyboardEvent(input, notUnicode, key, input.GetModifiers(), state)
	if keyMsg != nil {
		p.Send(*keyMsg)
	}
}

// Focus implements window.KeyboardHandler interface.
func (p *Program) Focus(win *window.Window, device *window.Input) {
	// Send focus event (could be extended to send a FocusMsg)
}

// Enter implements window.WidgetHandler interface for pointer enter events.
func (p *Program) Enter(widget *window.Widget, input *window.Input, x float32, y float32) {
	// Store pointer position
	p.pointerX = x
	p.pointerY = y
}

// Leave implements window.WidgetHandler interface for pointer leave events.
func (p *Program) Leave(widget *window.Widget, input *window.Input) {
	// Mouse left the window
}

// Motion implements window.WidgetHandler interface for pointer motion events.
func (p *Program) Motion(widget *window.Widget, input *window.Input, time uint32, x float32, y float32) int {
	// Store pointer position
	p.pointerX = x
	p.pointerY = y

	cellWidth := p.renderer.CellWidth()
	cellHeight := p.renderer.CellHeight()

	mouseMsg := MapMouseMotion(x, y, cellWidth, cellHeight)
	if mouseMsg != nil {
		p.Send(*mouseMsg)
	}

	return window.CursorLeft
}

// Button implements window.WidgetHandler interface for pointer button events.
func (p *Program) Button(
	widget *window.Widget,
	input *window.Input,
	time uint32,
	button uint32,
	state wl.PointerButtonState,
	data window.WidgetHandler,
) {
	cellWidth := p.renderer.CellWidth()
	cellHeight := p.renderer.CellHeight()

	// Use stored pointer position
	mouseMsg := MapMouseButton(p.pointerX, p.pointerY, button, state, cellWidth, cellHeight)
	if mouseMsg != nil {
		p.Send(*mouseMsg)
	}
}

// Axis implements window.WidgetHandler interface for pointer axis (scroll) events.
func (p *Program) Axis(widget *window.Widget, input *window.Input, time uint32, axis uint32, value float32) {
	cellWidth := p.renderer.CellWidth()
	cellHeight := p.renderer.CellHeight()

	// Use stored pointer position
	mouseMsg := MapMouseScroll(p.pointerX, p.pointerY, axis, value, cellWidth, cellHeight)
	if mouseMsg != nil {
		p.Send(*mouseMsg)
	}
}

// TouchUp implements window.WidgetHandler interface.
func (p *Program) TouchUp(widget *window.Widget, input *window.Input, serial uint32, time uint32, id int32) {
	// Touch events not implemented yet
}

// TouchDown implements window.WidgetHandler interface.
func (p *Program) TouchDown(
	widget *window.Widget,
	input *window.Input,
	serial uint32,
	time uint32,
	id int32,
	x float32,
	y float32,
) {
	// Touch events not implemented yet
}

// TouchMotion implements window.WidgetHandler interface.
func (p *Program) TouchMotion(widget *window.Widget, input *window.Input, time uint32, id int32, x float32, y float32) {
	// Touch events not implemented yet
}

// TouchFrame implements window.WidgetHandler interface.
func (p *Program) TouchFrame(widget *window.Widget, input *window.Input) {
	// Touch events not implemented yet
}

// TouchCancel implements window.WidgetHandler interface.
func (p *Program) TouchCancel(widget *window.Widget, width int32, height int32) {
	// Touch events not implemented yet
}

// AxisSource implements window.WidgetHandler interface.
func (p *Program) AxisSource(widget *window.Widget, input *window.Input, source uint32) {
	// Axis source events not needed for basic functionality
}

// AxisStop implements window.WidgetHandler interface.
func (p *Program) AxisStop(widget *window.Widget, input *window.Input, time uint32, axis uint32) {
	// Axis stop events not needed for basic functionality
}

// AxisDiscrete implements window.WidgetHandler interface.
func (p *Program) AxisDiscrete(widget *window.Widget, input *window.Input, axis uint32, discrete int32) {
	// Axis discrete events not needed for basic functionality
}

// PointerFrame implements window.WidgetHandler interface.
func (p *Program) PointerFrame(widget *window.Widget, input *window.Input) {
	// Pointer frame events not needed for basic functionality
}
