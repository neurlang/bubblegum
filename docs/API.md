# BubbleGum API Documentation

This document provides comprehensive documentation for all public types and functions in the BubbleGum library.

## Table of Contents

- [Core Interfaces](#core-interfaces)
- [Program](#program)
- [Messages](#messages)
- [Commands](#commands)
- [Configuration](#configuration)
- [Differences from Bubble Tea](#differences-from-bubble-tea)

## Core Interfaces

### Model

The `Model` interface represents your application state following The Elm Architecture pattern.

```go
type Model interface {
    Init() Cmd
    Update(Msg) (Model, Cmd)
    View() string
}
```

**Methods:**

- `Init() Cmd` - Called when the program starts. Returns an optional initial command to execute.
- `Update(Msg) (Model, Cmd)` - Called when a message is received. Returns the updated model and an optional command to execute.
- `View() string` - Renders the model as a string. The string may contain ANSI escape sequences for styling.

**Example:**

```go
type myModel struct {
    counter int
}

func (m myModel) Init() lib.Cmd {
    return nil // No initial command
}

func (m myModel) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
    switch msg := msg.(type) {
    case lib.KeyMsg:
        if msg.Type == lib.KeyUp {
            m.counter++
        }
    }
    return m, nil
}

func (m myModel) View() string {
    return fmt.Sprintf("Count: %d", m.counter)
}
```

### Msg

The `Msg` interface represents an event in the system. Any type can be a message.

```go
type Msg interface{}
```

Messages are passed to the `Update` function to handle events like keyboard input, mouse clicks, timer ticks, or custom application events.

### Cmd

The `Cmd` type represents an asynchronous operation that produces messages.

```go
type Cmd func() Msg
```

Commands are returned by `Init` and `Update` and executed by the runtime. When a command completes, its resulting message is sent to `Update`.

## Program

### NewProgram

Creates a new Program with the given model and options.

```go
func NewProgram(model Model, opts ...ProgramOption) *Program
```

**Parameters:**
- `model` - Your application's initial model implementing the Model interface
- `opts` - Optional configuration options (see [Configuration](#configuration))

**Returns:**
- `*Program` - A new Program instance ready to run

**Example:**

```go
p := lib.NewProgram(
    myModel{},
    lib.WithWindowTitle("My App"),
    lib.WithInitialSize(800, 600),
)
```

### Program.Run

Starts the program and blocks until it exits.

```go
func (p *Program) Run() (Model, error)
```

**Returns:**
- `Model` - The final model state when the program exits
- `error` - Any error that occurred during initialization or execution

**Example:**

```go
finalModel, err := p.Run()
if err != nil {
    log.Fatal(err)
}
```

### Program.Send

Sends a message to the program's Update function. This is thread-safe and can be called from any goroutine.

```go
func (p *Program) Send(msg Msg)
```

**Parameters:**
- `msg` - The message to send

**Example:**

```go
// From a goroutine
go func() {
    time.Sleep(time.Second)
    p.Send(MyCustomMsg{})
}()
```

### Program.Quit

Signals the program to exit gracefully.

```go
func (p *Program) Quit()
```

**Example:**

```go
p.Quit()
```

## Messages

### KeyMsg

Represents a keyboard input event.

```go
type KeyMsg struct {
    Type  KeyType
    Runes []rune
    Alt   bool
}
```

**Fields:**
- `Type` - The type of key pressed (see KeyType constants)
- `Runes` - The character(s) typed (for KeyRunes type)
- `Alt` - Whether the Alt modifier was held

**KeyType Constants:**

```go
const (
    KeyRunes      // Regular character input
    KeyEnter      // Enter/Return key
    KeyBackspace  // Backspace key
    KeyTab        // Tab key
    KeyEsc        // Escape key
    KeyUp         // Up arrow
    KeyDown       // Down arrow
    KeyLeft       // Left arrow
    KeyRight      // Right arrow
    KeyHome       // Home key
    KeyEnd        // End key
    KeyPgUp       // Page Up
    KeyPgDown     // Page Down
    KeyDelete     // Delete key
    KeyInsert     // Insert key
    KeyF1-KeyF12  // Function keys
    KeyCtrlC      // Ctrl+C
    KeyCtrlD      // Ctrl+D
    KeyCtrlL      // Ctrl+L
    KeyCtrlZ      // Ctrl+Z
)
```

**Example:**

```go
func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
    switch msg := msg.(type) {
    case lib.KeyMsg:
        switch msg.Type {
        case lib.KeyEsc:
            return m, lib.Quit
        case lib.KeyRunes:
            // Handle character input
            text := string(msg.Runes)
        }
    }
    return m, nil
}
```

### MouseMsg

Represents a mouse input event.

```go
type MouseMsg struct {
    X      int
    Y      int
    Type   MouseEventType
    Button MouseButton
}
```

**Fields:**
- `X` - The column position in the terminal grid
- `Y` - The row position in the terminal grid
- `Type` - The type of mouse event (see MouseEventType constants)
- `Button` - The mouse button involved (see MouseButton constants)

**MouseEventType Constants:**

```go
const (
    MousePress    // Mouse button pressed
    MouseRelease  // Mouse button released
    MouseMotion   // Mouse moved
    MouseWheel    // Mouse wheel scrolled
)
```

**MouseButton Constants:**

```go
const (
    MouseButtonNone        // No button (for motion events)
    MouseButtonLeft        // Left mouse button
    MouseButtonMiddle      // Middle mouse button
    MouseButtonRight       // Right mouse button
    MouseButtonWheelUp     // Scroll wheel up
    MouseButtonWheelDown   // Scroll wheel down
    MouseButtonWheelLeft   // Scroll wheel left
    MouseButtonWheelRight  // Scroll wheel right
)
```

**Example:**

```go
func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
    switch msg := msg.(type) {
    case lib.MouseMsg:
        if msg.Type == lib.MousePress && msg.Button == lib.MouseButtonLeft {
            m.clickX = msg.X
            m.clickY = msg.Y
        }
    }
    return m, nil
}
```

### WindowSizeMsg

Represents a window resize event.

```go
type WindowSizeMsg struct {
    Width  int
    Height int
}
```

**Fields:**
- `Width` - The new window width in terminal grid columns
- `Height` - The new window height in terminal grid rows

**Example:**

```go
func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
    switch msg := msg.(type) {
    case lib.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
    }
    return m, nil
}
```

## Commands

### Quit

Signals the program to exit.

```go
func Quit() Msg
```

**Returns:**
- A quit message that will cause the program to exit

**Example:**

```go
func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
    switch msg := msg.(type) {
    case lib.KeyMsg:
        if msg.Type == lib.KeyEsc {
            return m, lib.Quit
        }
    }
    return m, nil
}
```

### Batch

Executes multiple commands concurrently.

```go
func Batch(cmds ...Cmd) Cmd
```

**Parameters:**
- `cmds` - Variable number of commands to execute

**Returns:**
- A command that executes all provided commands

**Example:**

```go
return m, lib.Batch(
    lib.Tick(time.Second, func(t time.Time) lib.Msg {
        return tickMsg(t)
    }),
    fetchDataCmd(),
    saveStateCmd(),
)
```

### Tick

Creates a timer command that fires once after the specified duration.

```go
func Tick(d time.Duration, fn func(time.Time) Msg) Cmd
```

**Parameters:**
- `d` - Duration to wait before firing
- `fn` - Function to call with the current time, returns a message

**Returns:**
- A command that will send the message after the duration

**Example:**

```go
func (m model) Init() lib.Cmd {
    return lib.Tick(time.Second, func(t time.Time) lib.Msg {
        return tickMsg{time: t}
    })
}
```

### Every

Creates a recurring timer command that fires repeatedly at the specified interval.

```go
func Every(d time.Duration, fn func(time.Time) Msg) Cmd
```

**Parameters:**
- `d` - Duration between each tick
- `fn` - Function to call with the current time, returns a message

**Returns:**
- A command that will send messages repeatedly

**Example:**

```go
func (m model) Init() lib.Cmd {
    return lib.Every(time.Second, func(t time.Time) lib.Msg {
        return tickMsg{time: t}
    })
}
```

**Note:** To stop a recurring timer, return `lib.Quit` or don't return the command from Update.

## Configuration

### ProgramOptions

Configuration options for customizing the program's appearance and behavior.

```go
type ProgramOptions struct {
    FontFamily    string
    FontSize      int
    InitialWidth  int32
    InitialHeight int32
    WindowTitle   string
    FPS           int
}
```

**Fields:**
- `FontFamily` - Font family to use for rendering text (default: "Monospace")
- `FontSize` - Font size in points (default: 12)
- `InitialWidth` - Initial window width in pixels (default: 800)
- `InitialHeight` - Initial window height in pixels (default: 600)
- `WindowTitle` - Text displayed in the window's title bar (default: "BubbleGum Application")
- `FPS` - Maximum frames per second for rendering, 0 means no limit (default: 60)

### Configuration Functions

#### WithFontFamily

Sets the font family for text rendering.

```go
func WithFontFamily(family string) ProgramOption
```

**Example:**
```go
lib.WithFontFamily("Courier New")
```

#### WithFontSize

Sets the font size in points.

```go
func WithFontSize(size int) ProgramOption
```

**Example:**
```go
lib.WithFontSize(14)
```

#### WithInitialSize

Sets the initial window dimensions in pixels.

```go
func WithInitialSize(width, height int32) ProgramOption
```

**Example:**
```go
lib.WithInitialSize(1024, 768)
```

#### WithWindowTitle

Sets the window title.

```go
func WithWindowTitle(title string) ProgramOption
```

**Example:**
```go
lib.WithWindowTitle("My Application")
```

#### WithFPS

Sets the maximum frames per second for rendering.

```go
func WithFPS(fps int) ProgramOption
```

**Example:**
```go
lib.WithFPS(30) // Limit to 30 FPS
```

## Differences from Bubble Tea

While BubbleGum maintains API compatibility with Bubble Tea, there are some key differences:

### Import Path

**Bubble Tea:**
```go
import tea "github.com/charmbracelet/bubbletea"
```

**BubbleGum:**
```go
import "github.com/neurlang/bubblegum/lib"
```

### Program Creation

**Bubble Tea:**
```go
p := tea.NewProgram(model{})
```

**BubbleGum:**
```go
p := lib.NewProgram(
    model{},
    lib.WithWindowTitle("My App"),
    lib.WithInitialSize(800, 600),
)
```

BubbleGum provides additional configuration options for window customization.

### Mouse Support

**Bubble Tea:** Mouse support must be explicitly enabled with `tea.WithMouseCellMotion()` or `tea.WithMouseAllMotion()`.

**BubbleGum:** Mouse events are always enabled and delivered as `MouseMsg`.

### Terminal vs Window

**Bubble Tea:** Runs in a terminal and uses terminal control sequences.

**BubbleGum:** Creates a native GUI window and renders using Cairo graphics.

### Performance Considerations

**Bubble Tea:** Optimized for terminal rendering with minimal overhead.

**BubbleGum:** Includes frame rate limiting (default 60 FPS) to prevent excessive rendering in GUI windows.

### ANSI Escape Sequences

Both support ANSI escape sequences for styling, but BubbleGum renders them graphically:

- Colors (16-color, 256-color, RGB)
- Text attributes (bold, italic, underline, strikethrough)
- Cursor positioning and clearing

### Platform Support

**Bubble Tea:** Cross-platform terminal support (Linux, macOS, Windows, BSD).

**BubbleGum:** Currently supports Linux (Wayland) and Windows with native windowing.

### Error Handling

**BubbleGum** provides more detailed error messages for initialization failures:

```go
if _, err := p.Run(); err != nil {
    // Errors include context about what failed
    // e.g., "failed to create Wayland display: ..."
    log.Fatal(err)
}
```

### Panic Recovery

BubbleGum includes panic recovery in `Init`, `Update`, and `View` functions:

- Panics are caught and logged
- Stack traces are recorded
- Application exits gracefully
- Window resources are cleaned up

This prevents the window from being left in an invalid state if your code panics.

## Best Practices

### 1. Handle WindowSizeMsg

Always handle window resize events to ensure your UI adapts properly:

```go
case lib.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
```

### 2. Use Frame Rate Limiting

For applications with frequent updates, use FPS limiting to reduce CPU usage:

```go
lib.WithFPS(30) // 30 FPS is often sufficient
```

### 3. Avoid Blocking Operations in Update

Never perform blocking operations in `Update`. Use commands instead:

```go
// Bad
func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
    data := fetchData() // Blocks!
    return m, nil
}

// Good
func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
    return m, fetchDataCmd() // Returns immediately
}

func fetchDataCmd() lib.Cmd {
    return func() lib.Msg {
        data := fetchData() // Runs in background
        return dataMsg{data}
    }
}
```

### 4. Test with Different Window Sizes

Your UI should work at various window sizes. Test with small and large windows.

### 5. Use Components

Leverage the ported Bubbles components instead of building from scratch:

```go
import "github.com/neurlang/bubblegum/components/textinput"

type model struct {
    input textinput.Model
}
```

## See Also

- [Porting Guide](PORTING.md) - How to port existing Bubble Tea applications
- [Component Documentation](../components/README.md) - Documentation for ported Bubbles components
- [Troubleshooting](TROUBLESHOOTING.md) - Common issues and solutions
