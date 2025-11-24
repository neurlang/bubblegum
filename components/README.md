# BubbleGum Components

This directory contains ported Bubbles components adapted for BubbleGum applications.

## Available Components

### TextInput

A text input field component for capturing user input.

**Features:**
- Cursor positioning and navigation
- Character insertion and deletion
- Horizontal scrolling for long input
- Character limit support
- Placeholder text
- Focus management

**Example:**
```go
ti := textinput.New()
ti.Placeholder = "Enter your name..."
ti.Focus()
ti.SetValue("Hello")
```

### Spinner

An animated spinner component for indicating loading or processing states.

**Features:**
- Multiple predefined spinner styles (Line, Dot, MiniDot, Jump, Pulse, etc.)
- Customizable animation speed
- Automatic frame animation

**Example:**
```go
sp := spinner.New(spinner.WithSpinner(spinner.Dot))
// In Init(): return sp.Tick()
```

### List

A list component for displaying and navigating through items.

**Features:**
- Scrollable item list
- Cursor navigation (up/down, page up/down, home/end)
- Item filtering (press '/' to filter)
- Selection tracking
- Custom item rendering

**Example:**
```go
items := []list.Item{
    list.NewDefaultItem("Item 1", "Description 1"),
    list.NewDefaultItem("Item 2", "Description 2"),
}
l := list.New(items, 40, 10)
l.Title = "My List"
```

### Viewport

A viewport component for scrolling through large content.

**Features:**
- Vertical scrolling
- Mouse wheel support
- Page up/down navigation
- Scroll position tracking
- Content clipping

**Example:**
```go
vp := viewport.New(40, 10)
vp.SetContent("Line 1\nLine 2\nLine 3\n...")
```

## Usage

Import the components you need:

```go
import (
    "github.com/bubblegum/components/textinput"
    "github.com/bubblegum/components/spinner"
    "github.com/bubblegum/components/list"
    "github.com/bubblegum/components/viewport"
)
```

See `examples/components/main.go` for a complete working example demonstrating all components.

## Differences from Bubbles

These components are simplified versions of the original Bubbles components:

- **Styling**: Removed lipgloss styling in favor of ANSI escape codes
- **Complexity**: Focused on core functionality, removed advanced features
- **Dependencies**: Minimal dependencies, using only BubbleGum's lib package
- **API**: Simplified API while maintaining compatibility with common use cases

## Building

All components are built as part of the main BubbleGum project:

```bash
go build ./components/textinput
go build ./components/spinner
go build ./components/list
go build ./components/viewport
```
