# BubbleGum

BubbleGum is a port of the [Bubble Tea](https://github.com/charmbracelet/bubbletea) terminal UI framework that renders user interfaces using native windowing. It enables terminal UI applications built with Bubble Tea to be compiled into full GUI applications with minimal code changes, providing a portable way to create both terminal and graphical user interfaces from the same codebase.

## Features

- **Bubble Tea Compatible API** - Use the same Model/Update/View pattern you know from Bubble Tea
- **Native GUI Windows** - Render your TUI apps in native windows on Linux (Wayland) and Windows
- **ANSI Escape Sequence Support** - Full support for colors, bold, italic, underline, and other text styling
- **Mouse and Keyboard Input** - Complete input handling including mouse clicks, scrolling, and keyboard shortcuts
- **Ported Bubbles Components** - Familiar UI components like text inputs, spinners, lists, and viewports
- **Asynchronous Commands** - Execute I/O operations, timers, and custom commands just like in Bubble Tea

## Installation

### Prerequisites

**Linux (Wayland):**
- Go 1.21 or later
- Wayland compositor (GNOME, KDE Plasma, Sway, etc.)
- Development libraries:
  ```bash
  # Ubuntu/Debian
  sudo apt-get install libwayland-dev libxkbcommon-dev libcairo2-dev
  
  # Fedora
  sudo dnf install wayland-devel libxkbcommon-devel cairo-devel
  
  # Arch Linux
  sudo pacman -S wayland libxkbcommon cairo
  ```

**Windows:**
- Go 1.21 or later
- No additional dependencies required

### Install BubbleGum

```bash
go get github.com/neurlang/bubblegum/lib
```

## Quick Start

Here's a simple counter application that demonstrates the core concepts:

```go
package main

import (
	"fmt"
	"os"

	"github.com/neurlang/bubblegum/lib"
)

type model struct {
	count int
}

func (m model) Init() lib.Cmd {
	return nil
}

func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
	switch msg := msg.(type) {
	case lib.KeyMsg:
		switch msg.Type {
		case lib.KeyEsc, lib.KeyCtrlC:
			return m, lib.Quit
		case lib.KeyUp:
			m.count++
		case lib.KeyDown:
			m.count--
		}
	case lib.WindowSizeMsg:
		// Handle window resize if needed
	}
	return m, nil
}

func (m model) View() string {
	return fmt.Sprintf("Counter: %d\n\nPress Up/Down to change, Esc to quit", m.count)
}

func main() {
	p := lib.NewProgram(
		model{},
		lib.WithWindowTitle("Simple Counter"),
		lib.WithInitialSize(800, 600),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
```

Build and run:
```bash
go build -o counter
./counter
```

## Core Concepts

### The Elm Architecture

BubbleGum follows The Elm Architecture pattern with three core components:

**Model** - Your application state
```go
type model struct {
    // Your data here
}
```

**Update** - Handle events and update state
```go
func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
    // Process messages and return updated model
    return m, nil
}
```

**View** - Render your UI as a string
```go
func (m model) View() string {
    return "Hello, World!"
}
```

### Messages

Messages represent events in your application. BubbleGum provides several built-in message types:

- `KeyMsg` - Keyboard input
- `MouseMsg` - Mouse clicks, movement, and scrolling
- `WindowSizeMsg` - Window resize events

You can also create custom message types for your application logic.

### Commands

Commands are asynchronous operations that produce messages. Common commands include:

- `lib.Quit` - Exit the application
- `lib.Batch(...)` - Execute multiple commands
- `lib.Tick(duration, func)` - Timer that fires once
- `lib.Every(duration, func)` - Recurring timer

### Styling with ANSI

Use ANSI escape sequences in your View output for styling:

```go
func (m model) View() string {
    return "\x1b[1;32mGreen Bold Text\x1b[0m\n" +
           "\x1b[4mUnderlined\x1b[0m\n" +
           "\x1b[7mInverted\x1b[0m"
}
```

Or use a styling library like [lipgloss](https://github.com/charmbracelet/lipgloss) for easier styling.

## Configuration Options

Customize your application window with these options:

```go
p := lib.NewProgram(
    model{},
    lib.WithWindowTitle("My App"),           // Set window title
    lib.WithInitialSize(1024, 768),          // Set initial dimensions
    lib.WithFontFamily("Monospace"),         // Set font family
    lib.WithFontSize(14),                    // Set font size
    lib.WithFPS(60),                         // Set frame rate limit
)
```

## Examples

The `examples/` directory contains several complete applications:

- **[simple](examples/simple/)** - Basic counter demonstrating Update/View cycle
- **[textinput-form](examples/textinput-form/)** - Text input component usage
- **[mouse](examples/mouse/)** - Mouse event handling
- **[mouse-simple](examples/mouse-simple/)** - Simple mouse interaction
- **[timer](examples/timer/)** - Timer command usage
- **[list-browser](examples/list-browser/)** - List component with scrolling
- **[components](examples/components/)** - Showcase of all ported components

Run any example:
```bash
cd examples/simple
go run main.go
```

## Components

BubbleGum includes ported versions of popular Bubbles components. See [components/README.md](components/README.md) for detailed documentation.

Available components:
- **textinput** - Single-line text input with cursor
- **spinner** - Animated loading spinner
- **list** - Scrollable list with selection
- **viewport** - Scrollable content area

## Project Structure

```
.
├── lib/                    # BubbleGum library (Bubble Tea-compatible API)
│   ├── types.go           # Core interfaces (Model, Msg, Cmd)
│   ├── program.go         # Program runner and lifecycle management
│   ├── commands.go        # Command implementations (Quit, Batch, Tick, etc.)
│   ├── messages.go        # Message types (KeyMsg, MouseMsg, WindowSizeMsg)
│   ├── input.go           # Input event mapping (keyboard and mouse)
│   ├── parser.go          # ANSI escape sequence parser
│   ├── renderer.go        # Cairo-based graphical renderer
│   ├── grid.go            # Terminal grid data structures
│   └── font.go            # Font loading and rendering
├── components/            # Ported Bubbles UI components
│   ├── textinput/        # Text input component
│   ├── spinner/          # Spinner component
│   ├── list/             # List component
│   └── viewport/         # Viewport component
├── examples/             # Example applications
└── wayland/              # Wayland window library (dependency)
```

## Differences from Bubble Tea

While BubbleGum maintains API compatibility with Bubble Tea, there are some differences:

1. **Import Path** - Use `github.com/neurlang/bubblegum/lib` instead of `github.com/charmbracelet/bubbletea`
2. **Window Configuration** - Additional options for window title, size, and font
3. **Mouse Support** - Mouse events are always enabled (no need to enable mouse mode)
4. **Performance** - Rendering is optimized for GUI windows with frame rate limiting

See [docs/PORTING.md](docs/PORTING.md) for a complete porting guide.

## Troubleshooting

See [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) for common issues and solutions.

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

See [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - The original terminal UI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - UI components for Bubble Tea
- [Wayland Window Library](https://github.com/neurlang/wayland) - Cross-platform windowing
