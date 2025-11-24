package lib

import (
	"context"
	"fmt"
	"runtime/debug"
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

	lastView          string
	lastRender        time.Time
	windowWidth       int
	windowHeight      int
	input             *window.Input
	pointerX          float32
	pointerY          float32
	lastCellX         int
	lastCellY         int
	cellPosValid      bool
	redrawScheduled   bool
	motionPending     bool
	pendingMotionX    int
	pendingMotionY    int
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
	Info("Starting BubbleGum application")
	Debug("Configuration: %+v", p.options)
	
	// Validate configuration options
	if err := p.validateOptions(); err != nil {
		return p.model, fmt.Errorf("invalid configuration: %w", err)
	}

	Debug("Creating Wayland display")
	// Create Wayland display
	display, err := window.DisplayCreate([]string{})
	if err != nil {
		return p.model, fmt.Errorf("failed to create Wayland display: %w (ensure Wayland compositor is running)", err)
	}
	p.display = display
	defer p.display.Destroy()

	Debug("Creating window")
	// Create window
	p.window = window.Create(display)
	if p.window == nil {
		return p.model, fmt.Errorf("failed to create window: window.Create returned nil")
	}
	defer p.window.Destroy()

	// Set window title
	Debug("Setting window title: %s", p.options.WindowTitle)
	p.window.SetTitle(p.options.WindowTitle)

	// Set buffer type
	Debug("Setting buffer type to SHM")
	p.window.SetBufferType(window.BufferTypeShm)

	Debug("Creating widget")
	// Create widget
	p.widget = p.window.AddWidget(p)
	if p.widget == nil {
		return p.model, fmt.Errorf("failed to create widget: AddWidget returned nil")
	}
	defer p.widget.Destroy()

	// Set up keyboard handler
	Debug("Setting up keyboard handler")
	p.window.SetKeyboardHandler(p)

	Debug("Creating renderer")
	// Create renderer
	p.renderer, err = NewRenderer(RendererOptions{
		DefaultFg: NewColor(255, 255, 255),
		DefaultBg: NewColor(0, 0, 0),
	})
	if err != nil {
		return p.model, fmt.Errorf("failed to create renderer: %w", err)
	}

	Debug("Creating command executor")
	// Create command executor
	p.cmdExec = NewCommandExecutor(p.ctx, p.msgChan)
	defer p.cmdExec.Shutdown()

	// Create input handler (note: Input is created by the window system, not by us)
	// We'll get it from event handlers

	// Schedule initial resize
	Debug("Scheduling initial resize: %dx%d", p.options.InitialWidth, p.options.InitialHeight)
	p.widget.ScheduleResize(p.options.InitialWidth, p.options.InitialHeight)

	Debug("Calling model Init()")
	// Call model's Init() with panic recovery
	var initialCmd Cmd
	func() {
		defer func() {
			if r := recover(); r != nil {
				Error("Panic in Init(): %v", r)
				// Log stack trace
				Error("Stack trace: %v", getStackTrace())
				// Don't execute any command if Init panicked
				initialCmd = nil
			}
		}()
		initialCmd = p.model.Init()
	}()
	
	if initialCmd != nil {
		Debug("Executing initial command")
		p.cmdExec.Execute(initialCmd)
	}

	Info("Starting event loop")
	// Run the display event loop (blocks until quit)
	window.DisplayRun(display)

	Info("Application exited")
	return p.model, nil
}

// validateOptions validates the program configuration options.
func (p *Program) validateOptions() error {
	if p.options.InitialWidth <= 0 {
		return fmt.Errorf("initial width must be positive, got %d", p.options.InitialWidth)
	}
	if p.options.InitialHeight <= 0 {
		return fmt.Errorf("initial height must be positive, got %d", p.options.InitialHeight)
	}
	if p.options.FontSize <= 0 {
		return fmt.Errorf("font size must be positive, got %d", p.options.FontSize)
	}
	if p.options.FPS < 0 {
		return fmt.Errorf("FPS must be non-negative, got %d", p.options.FPS)
	}
	if p.options.WindowTitle == "" {
		return fmt.Errorf("window title cannot be empty")
	}
	if p.options.FontFamily == "" {
		return fmt.Errorf("font family cannot be empty")
	}
	return nil
}

// handleMessage processes a single message by calling Update and rendering.
func (p *Program) handleMessage(msg Msg) {
	Debug("handleMessage received: %T", msg)
	
	// Check if this is a quit message
	if _, isQuit := msg.(quitMsg); isQuit {
		Info("Quit message received, exiting")
		p.quit()
		return
	}

	// Call Update
	p.mu.Lock()
	var cmd Cmd
	p.model, cmd = p.model.Update(msg)
	p.mu.Unlock()

	Debug("Update completed, returned command: %v", cmd != nil)

	// Execute the returned command
	if cmd != nil {
		p.cmdExec.Execute(cmd)
	}

	// Trigger a redraw
	if p.window != nil {
		p.window.UninhibitRedraw()
		p.window.ScheduleRedraw()
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

	Debug("Window resized: %dx%d pixels -> %dx%d cells", pwidth, pheight, gridWidth, gridHeight)

	// Store dimensions
	p.windowWidth = gridWidth
	p.windowHeight = gridHeight

	// Send WindowSizeMsg (non-blocking)
	select {
	case p.msgChan <- WindowSizeMsg{
		Width:  gridWidth,
		Height: gridHeight,
	}:
	default:
		Warn("Message channel full, dropping WindowSizeMsg")
	}
}

// Redraw implements window.WidgetHandler interface.
// It renders the current view to the window.
func (p *Program) Redraw(widget *window.Widget) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Clear the redraw scheduled flag
	p.redrawScheduled = false
	
	// Check for pending motion and create a message for it
	// This avoids flooding the message channel with motion events
	var processedMotion bool
	if p.motionPending {
		mouseMsg := &MouseMsg{
			X:      p.pendingMotionX,
			Y:      p.pendingMotionY,
			Type:   MouseMotion,
			Button: MouseButtonNone,
		}
		p.motionPending = false
		processedMotion = true
		
		// Process the motion message directly
		func() {
			defer func() {
				if r := recover(); r != nil {
					Error("Panic in Update(): %v", r)
					Error("Stack trace: %v", getStackTrace())
					p.quit()
				}
			}()
			
			var cmd Cmd
			p.model, cmd = p.model.Update(*mouseMsg)
			if cmd != nil {
				p.cmdExec.Execute(cmd)
			}
		}()
	}

	// Process pending messages (non-blocking loop)
	hadMessages := false
	var messagesToProcess []Msg
	
	// Collect all pending messages (no more motion coalescing needed)
	for {
		select {
		case msg := <-p.msgChan:
			hadMessages = true
			
			// Check if this is a quit message
			if _, isQuit := msg.(quitMsg); isQuit {
				p.quit()
				return
			}
			
			messagesToProcess = append(messagesToProcess, msg)
		default:
			// No more messages to process
			goto done
		}
	}
done:
	
	// Process all collected messages
	for _, msg := range messagesToProcess {
		// Call Update with panic recovery
		func() {
			defer func() {
				if r := recover(); r != nil {
					Error("Panic in Update(): %v", r)
					Error("Stack trace: %v", getStackTrace())
					// Exit gracefully on panic
					p.quit()
				}
			}()
			
			var cmd Cmd
			p.model, cmd = p.model.Update(msg)

			// Execute the returned command
			if cmd != nil {
				p.cmdExec.Execute(cmd)
			}
		}()
	}

	// Check frame rate limiting
	if p.options.FPS > 0 {
		minFrameTime := time.Second / time.Duration(p.options.FPS)
		elapsed := time.Since(p.lastRender)
		if elapsed < minFrameTime {
			// Skip this frame but schedule another redraw if we had messages
			if hadMessages && p.window != nil {
				p.window.UninhibitRedraw()
			}
			return
		}
	}

	// Get the current view with panic recovery
	var view string
	func() {
		defer func() {
			if r := recover(); r != nil {
				Error("Panic in View(): %v", r)
				Error("Stack trace: %v", getStackTrace())
				// Use empty view on panic
				view = ""
				// Exit gracefully on panic
				p.quit()
			}
		}()
		view = p.model.View()
	}()

	// Skip rendering if view hasn't changed (unless we processed messages)
	if view == p.lastView && p.lastView != "" && !hadMessages {
		return
	}

	// Get the window surface
	surface := p.window.WindowGetSurface()
	if surface == nil {
		Warn("WindowGetSurface returned nil, skipping render")
		return
	}

	// Parse the view into a terminal grid
	grid := ParseANSI(view, p.windowWidth, p.windowHeight)
	if grid == nil {
		Error("ParseANSI returned nil grid, skipping render")
		return
	}

	// Render the grid
	err := p.renderer.Render(grid, surface)
	if err != nil {
		// Log error but continue - don't crash the application
		Error("Render failed: %v", err)
		return
	}

	Debug("Rendered frame successfully")

	// Update state
	p.lastView = view
	p.lastRender = time.Now()

	// Uninhibit redraw to allow future redraws
	if p.window != nil {
		p.window.UninhibitRedraw()
		
		// If we processed motion and there's STILL motion pending
		// (because more motion events came in during this redraw),
		// schedule another redraw to process it
		if processedMotion && p.motionPending && p.widget != nil {
			p.widget.ScheduleRedraw()
		}
	}
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
	// Uninhibit redraw to allow the window to update
	win.UninhibitRedraw()

	// Store input reference
	if p.input == nil {
		p.input = input
	}

	// The notUnicode parameter contains the keysym
	// GetRune will modify it, so we need to save it first
	keysym := notUnicode

	// Map the keyboard event to a KeyMsg
	keyMsg := MapKeyboardEvent(input, keysym, key, input.GetModifiers(), state)
	if keyMsg != nil {
		Debug("Keyboard event: key=%d, keysym=%d, state=%d", key, keysym, state)
		// Send to channel (non-blocking)
		select {
		case p.msgChan <- *keyMsg:
			// Schedule a redraw to process the message
			if p.widget != nil {
				p.widget.ScheduleRedraw()
			}
		default:
			Warn("Message channel full, dropping keyboard event")
		}
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

	// Calculate cell position
	cellX := int(x / float32(cellWidth))
	cellY := int(y / float32(cellHeight))

	// Only mark motion as pending if the cell position has changed
	if !p.cellPosValid || cellX != p.lastCellX || cellY != p.lastCellY {
		p.mu.Lock()
		wasAlreadyPending := p.motionPending
		p.lastCellX = cellX
		p.lastCellY = cellY
		p.cellPosValid = true
		p.motionPending = true
		p.pendingMotionX = cellX
		p.pendingMotionY = cellY
		p.mu.Unlock()
		
		Debug("Mouse motion: cell (%d, %d), already pending: %v", cellX, cellY, wasAlreadyPending)
		
		// Only schedule a redraw if motion wasn't already pending
		// This prevents spamming ScheduleRedraw calls
		if !wasAlreadyPending && p.window != nil {
			Debug("Scheduling redraw for motion")
			p.window.UninhibitRedraw()
			widget.ScheduleRedraw()
		}
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

	Debug("Mouse button: button=%d, state=%d", button, state)

	// Use stored pointer position
	mouseMsg := MapMouseButton(p.pointerX, p.pointerY, button, state, cellWidth, cellHeight)
	if mouseMsg != nil {
		p.Send(*mouseMsg)
		// Schedule a redraw to process the message
		if !p.redrawScheduled && p.window != nil {
			p.redrawScheduled = true
			p.window.UninhibitRedraw()
			p.widget.ScheduleRedraw()
		}
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
		// Schedule a redraw to process the message
		if !p.redrawScheduled && p.window != nil {
			p.redrawScheduled = true
			p.window.UninhibitRedraw()
			p.widget.ScheduleRedraw()
		}
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

// getStackTrace returns the current stack trace as a string.
func getStackTrace() string {
	return string(debug.Stack())
}
