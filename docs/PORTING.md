# Porting Guide: Bubble Tea to BubbleGum

This guide explains how to port existing Bubble Tea applications to BubbleGum, enabling them to run as native GUI applications.

## Table of Contents

- [Overview](#overview)
- [Quick Migration Checklist](#quick-migration-checklist)
- [Step-by-Step Guide](#step-by-step-guide)
- [Code Changes Required](#code-changes-required)
- [Common Patterns](#common-patterns)
- [Common Pitfalls](#common-pitfalls)
- [Testing Your Port](#testing-your-port)

## Overview

BubbleGum is designed to be highly compatible with Bubble Tea. In most cases, porting requires only changing imports and adding window configuration. Your core application logic (Model, Update, View) typically requires no changes.

**Compatibility Level:**
- ✅ Model/Update/View pattern - 100% compatible
- ✅ Message types - 100% compatible
- ✅ Commands (Quit, Batch, Tick, Every) - 100% compatible
- ✅ ANSI escape sequences - Fully supported
- ⚠️ Mouse support - Always enabled (no need to opt-in)
- ⚠️ Terminal-specific features - Not applicable (e.g., alt screen)

## Quick Migration Checklist

- [ ] Change import from `github.com/charmbracelet/bubbletea` to `github.com/bubblegum/lib`
- [ ] Update component imports (if using Bubbles components)
- [ ] Add window configuration options (title, size, etc.)
- [ ] Remove terminal-specific options (alt screen, mouse mode)
- [ ] Handle `WindowSizeMsg` for responsive layouts
- [ ] Test with different window sizes
- [ ] Build and run as GUI application

## Step-by-Step Guide

### Step 1: Update Imports

**Before (Bubble Tea):**
```go
import (
    tea "github.com/charmbracelet/bubbletea"
)
```

**After (BubbleGum):**
```go
import (
    "github.com/bubblegum/lib"
)
```

If you're using Bubbles components:

**Before:**
```go
import (
    "github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/bubbles/spinner"
)
```

**After:**
```go
import (
    "github.com/bubblegum/components/textinput"
    "github.com/bubblegum/components/spinner"
)
```

### Step 2: Update Type References

Replace all `tea.` prefixes with `lib.`:

**Before:**
```go
func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // ...
    }
    return m, nil
}
```

**After:**
```go
func (m model) Init() lib.Cmd {
    return nil
}

func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
    switch msg := msg.(type) {
    case lib.KeyMsg:
        // ...
    }
    return m, nil
}
```

### Step 3: Update Program Creation

**Before (Bubble Tea):**
```go
func main() {
    p := tea.NewProgram(
        initialModel(),
        tea.WithAltScreen(),
        tea.WithMouseCellMotion(),
    )
    
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
}
```

**After (BubbleGum):**
```go
func main() {
    p := lib.NewProgram(
        initialModel(),
        lib.WithWindowTitle("My Application"),
        lib.WithInitialSize(1024, 768),
        lib.WithFontSize(14),
    )
    
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
}
```

**Note:** Remove terminal-specific options like `WithAltScreen()` and `WithMouseCellMotion()` as they don't apply to GUI windows.

### Step 4: Handle Window Sizing

Add handling for `WindowSizeMsg` to make your UI responsive:

```go
type model struct {
    width  int
    height int
    // ... other fields
}

func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
    switch msg := msg.(type) {
    case lib.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        // Update any components that need to know the size
        return m, nil
    // ... other cases
    }
    return m, nil
}
```

### Step 5: Update Key Handling

Key constants remain the same, but ensure you're using the BubbleGum types:

```go
func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
    switch msg := msg.(type) {
    case lib.KeyMsg:
        switch msg.Type {
        case lib.KeyCtrlC, lib.KeyEsc:
            return m, lib.Quit
        case lib.KeyUp:
            // Handle up arrow
        case lib.KeyDown:
            // Handle down arrow
        case lib.KeyEnter:
            // Handle enter
        case lib.KeyRunes:
            // Handle character input
            text := string(msg.Runes)
        }
    }
    return m, nil
}
```

## Code Changes Required

### Minimal Changes Example

Here's a complete before/after example of a simple application:

**Before (Bubble Tea):**
```go
package main

import (
    "fmt"
    tea "github.com/charmbracelet/bubbletea"
)

type model struct {
    choices  []string
    cursor   int
    selected map[int]struct{}
}

func initialModel() model {
    return model{
        choices:  []string{"Buy carrots", "Buy celery", "Buy kohlrabi"},
        selected: make(map[int]struct{}),
    }
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        case "up", "k":
            if m.cursor > 0 {
                m.cursor--
            }
        case "down", "j":
            if m.cursor < len(m.choices)-1 {
                m.cursor++
            }
        case "enter", " ":
            _, ok := m.selected[m.cursor]
            if ok {
                delete(m.selected, m.cursor)
            } else {
                m.selected[m.cursor] = struct{}{}
            }
        }
    }
    return m, nil
}

func (m model) View() string {
    s := "What should we buy at the market?\n\n"
    for i, choice := range m.choices {
        cursor := " "
        if m.cursor == i {
            cursor = ">"
        }
        checked := " "
        if _, ok := m.selected[i]; ok {
            checked = "x"
        }
        s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
    }
    s += "\nPress q to quit.\n"
    return s
}

func main() {
    p := tea.NewProgram(initialModel())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v\n", err)
    }
}
```

**After (BubbleGum):**
```go
package main

import (
    "fmt"
    "os"
    "github.com/bubblegum/lib"  // Changed import
)

type model struct {
    choices  []string
    cursor   int
    selected map[int]struct{}
}

func initialModel() model {
    return model{
        choices:  []string{"Buy carrots", "Buy celery", "Buy kohlrabi"},
        selected: make(map[int]struct{}),
    }
}

func (m model) Init() lib.Cmd {  // Changed tea.Cmd to lib.Cmd
    return nil
}

func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {  // Changed types
    switch msg := msg.(type) {
    case lib.KeyMsg:  // Changed tea.KeyMsg to lib.KeyMsg
        switch msg.Type {  // Use msg.Type instead of msg.String()
        case lib.KeyCtrlC, lib.KeyEsc:  // Changed constants
            return m, lib.Quit  // Changed tea.Quit to lib.Quit
        case lib.KeyUp:  // Changed "up" to lib.KeyUp
            if m.cursor > 0 {
                m.cursor--
            }
        case lib.KeyDown:  // Changed "down" to lib.KeyDown
            if m.cursor < len(m.choices)-1 {
                m.cursor++
            }
        case lib.KeyEnter:  // Changed "enter" to lib.KeyEnter
            _, ok := m.selected[m.cursor]
            if ok {
                delete(m.selected, m.cursor)
            } else {
                m.selected[m.cursor] = struct{}{}
            }
        }
    case lib.WindowSizeMsg:  // Added window size handling
        // Handle resize if needed
    }
    return m, nil
}

func (m model) View() string {
    s := "What should we buy at the market?\n\n"
    for i, choice := range m.choices {
        cursor := " "
        if m.cursor == i {
            cursor = ">"
        }
        checked := " "
        if _, ok := m.selected[i]; ok {
            checked = "x"
        }
        s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
    }
    s += "\nPress Esc to quit.\n"
    return s
}

func main() {
    p := lib.NewProgram(  // Changed tea.NewProgram to lib.NewProgram
        initialModel(),
        lib.WithWindowTitle("Shopping List"),  // Added window config
        lib.WithInitialSize(600, 400),
    )
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
}
```

**Key Changes:**
1. Import path changed
2. All `tea.` prefixes changed to `lib.`
3. Key handling uses `msg.Type` with constants instead of `msg.String()`
4. Added `WindowSizeMsg` case
5. Added window configuration options

## Common Patterns

### Pattern 1: Handling Keyboard Input

**Bubble Tea:**
```go
case tea.KeyMsg:
    switch msg.String() {
    case "q", "ctrl+c":
        return m, tea.Quit
    }
```

**BubbleGum:**
```go
case lib.KeyMsg:
    switch msg.Type {
    case lib.KeyEsc, lib.KeyCtrlC:
        return m, lib.Quit
    case lib.KeyRunes:
        if string(msg.Runes) == "q" {
            return m, lib.Quit
        }
    }
```

### Pattern 2: Using Commands

Commands work identically in both:

```go
// Tick command
return m, lib.Tick(time.Second, func(t time.Time) lib.Msg {
    return tickMsg(t)
})

// Batch command
return m, lib.Batch(
    cmd1,
    cmd2,
    cmd3,
)

// Custom command
func fetchDataCmd() lib.Cmd {
    return func() lib.Msg {
        data, err := fetchData()
        if err != nil {
            return errMsg{err}
        }
        return dataMsg{data}
    }
}
```

### Pattern 3: Using Components

Component APIs are identical:

**Bubble Tea:**
```go
import "github.com/charmbracelet/bubbles/textinput"

type model struct {
    input textinput.Model
}

func initialModel() model {
    ti := textinput.New()
    ti.Focus()
    return model{input: ti}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    m.input, cmd = m.input.Update(msg)
    return m, cmd
}
```

**BubbleGum:**
```go
import "github.com/bubblegum/components/textinput"

type model struct {
    input textinput.Model
}

func initialModel() model {
    ti := textinput.New()
    ti.Focus()
    return model{input: ti}
}

func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
    var cmd lib.Cmd
    m.input, cmd = m.input.Update(msg)
    return m, cmd
}
```

Only the import path and type references change!

### Pattern 4: Styling with Lipgloss

Lipgloss styles work identically:

```go
import "github.com/charmbracelet/lipgloss"

var titleStyle = lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("#FAFAFA")).
    Background(lipgloss.Color("#7D56F4")).
    Padding(0, 1)

func (m model) View() string {
    return titleStyle.Render("My Title") + "\n\n" + m.content
}
```

No changes needed for styling!

## Common Pitfalls

### Pitfall 1: Using msg.String() for Key Handling

**Problem:**
```go
case lib.KeyMsg:
    switch msg.String() {  // This won't work!
    case "up":
        // ...
    }
```

**Solution:**
```go
case lib.KeyMsg:
    switch msg.Type {
    case lib.KeyUp:
        // ...
    case lib.KeyRunes:
        if string(msg.Runes) == "q" {
            // Handle character 'q'
        }
    }
```

### Pitfall 2: Forgetting Window Size Handling

**Problem:** UI doesn't adapt to window resizes.

**Solution:** Always handle `WindowSizeMsg`:
```go
case lib.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    // Update components that need size info
```

### Pitfall 3: Using Terminal-Specific Features

**Problem:**
```go
p := lib.NewProgram(
    model{},
    lib.WithAltScreen(),  // Doesn't exist!
)
```

**Solution:** Remove terminal-specific options. BubbleGum always uses a window.

### Pitfall 4: Hardcoding Dimensions

**Problem:**
```go
func (m model) View() string {
    // Assumes 80x24 terminal
    return strings.Repeat("=", 80) + "\n" + m.content
}
```

**Solution:** Use dynamic sizing:
```go
func (m model) View() string {
    return strings.Repeat("=", m.width) + "\n" + m.content
}
```

### Pitfall 5: Not Testing Different Window Sizes

**Problem:** UI breaks at small or large window sizes.

**Solution:** Test with various sizes:
```go
lib.WithInitialSize(400, 300)  // Small
lib.WithInitialSize(1920, 1080)  // Large
```

### Pitfall 6: Blocking in Update

**Problem:**
```go
func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
    data := http.Get("...")  // Blocks the UI!
    return m, nil
}
```

**Solution:** Use commands for async operations:
```go
func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
    return m, fetchDataCmd()
}

func fetchDataCmd() lib.Cmd {
    return func() lib.Msg {
        data := http.Get("...")
        return dataMsg{data}
    }
}
```

## Testing Your Port

### 1. Build and Run

```bash
go build -o myapp
./myapp
```

### 2. Test Window Operations

- Resize the window - UI should adapt
- Minimize and restore - Should work correctly
- Close with window button - Should exit cleanly
- Close with Esc/Ctrl+C - Should exit cleanly

### 3. Test Input

- Keyboard input - All keys should work
- Mouse clicks - Should register correctly
- Mouse scrolling - Should work if implemented
- Window focus - Should handle focus/blur

### 4. Test Different Configurations

```go
// Test different sizes
lib.WithInitialSize(400, 300)
lib.WithInitialSize(1920, 1080)

// Test different fonts
lib.WithFontSize(10)
lib.WithFontSize(20)

// Test frame rate limits
lib.WithFPS(30)
lib.WithFPS(60)
```

### 5. Check for Errors

Look for error messages in the console:
- Initialization errors
- Rendering errors
- Font loading errors

## Migration Examples

### Example 1: Simple Counter

See [examples/simple/](../examples/simple/) for a complete working example.

### Example 2: Text Input Form

See [examples/textinput-form/](../examples/textinput-form/) for form handling.

### Example 3: List Browser

See [examples/list-browser/](../examples/list-browser/) for component usage.

## Getting Help

If you encounter issues during porting:

1. Check the [API Documentation](API.md) for correct usage
2. Review the [Troubleshooting Guide](TROUBLESHOOTING.md) for common issues
3. Look at the [examples/](../examples/) directory for working code
4. Open an issue on GitHub with your specific problem

## Summary

Porting from Bubble Tea to BubbleGum is straightforward:

1. Change imports
2. Update type references
3. Add window configuration
4. Handle window sizing
5. Test thoroughly

Most applications can be ported in under 30 minutes with minimal code changes!
