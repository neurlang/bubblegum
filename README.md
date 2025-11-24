# BubbleGum

BubbleGum is a port of the Bubble Tea terminal UI framework that renders user interfaces using native windowing. It enables terminal UI applications built with Bubble Tea to be compiled into full GUI applications with minimal code changes.

## Project Structure

```
.
├── lib/                    # BubbleGum library (Bubble Tea-compatible API)
│   ├── types.go           # Core interfaces (Model, Msg, Cmd)
│   ├── program.go         # Program runner and lifecycle management
│   ├── commands.go        # Command implementations (Quit, Batch, etc.)
│   └── messages.go        # Message types (KeyMsg, MouseMsg, etc.)
├── components/            # Ported Bubbles UI components
│   └── README.md         # Component documentation
└── wayland/              # Wayland window library (submodule/dependency)
```

## Core Types

### Model
The application state following The Elm Architecture pattern:
- `Init()` - Initialize the model and return an optional command
- `Update(Msg)` - Handle messages and return updated model with optional command
- `View()` - Render the model as a string (with ANSI escape sequences)

### Msg
Any type can be a message representing events (keyboard, mouse, timers, etc.)

### Cmd
Asynchronous operations that produce messages

## Getting Started

```go
package main

import "github.com/bubblegum/lib"

type model struct {
    // Your application state
}

func (m model) Init() lib.Cmd {
    return nil
}

func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
    return m, nil
}

func (m model) View() string {
    return "Hello, BubbleGum!"
}

func main() {
    p := lib.NewProgram(model{})
    if _, err := p.Run(); err != nil {
        panic(err)
    }
}
```

## Status

This project is currently under development. Core types and project structure are in place. Implementation of rendering, input handling, and components is in progress.

## Requirements

- Go 1.21 or later
- Wayland window library dependencies (see wayland/README.md)

## License

See LICENSE file for details.
