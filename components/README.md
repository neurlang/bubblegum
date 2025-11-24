# BubbleGum Components

This directory contains ported Bubbles UI components adapted for BubbleGum applications. These components provide familiar, reusable UI elements that work seamlessly in graphical windows.

## Available Components

- **[textinput](#textinput)** - Single-line text input with cursor and selection
- **[spinner](#spinner)** - Animated loading spinner with multiple styles
- **[list](#list)** - Scrollable list with selection and filtering
- **[viewport](#viewport)** - Scrollable content area with mouse wheel support

## Table of Contents

- [Installation](#installation)
- [TextInput Component](#textinput)
- [Spinner Component](#spinner)
- [List Component](#list)
- [Viewport Component](#viewport)
- [Differences from Bubbles](#differences-from-bubbles)

## Installation

Components are included with BubbleGum. Import them as needed:

```go
import (
    "github.com/bubblegum/components/textinput"
    "github.com/bubblegum/components/spinner"
    "github.com/bubblegum/components/list"
    "github.com/bubblegum/components/viewport"
)
```

---

## TextInput

A single-line text input field with cursor, selection, and scrolling support.

### Features

- Character input with cursor positioning
- Horizontal scrolling for long text
- Character limit support
- Placeholder text
- Focus management
- Keyboard navigation (arrows, home, end, backspace, delete)

### Basic Usage

```go
import "github.com/bubblegum/components/textinput"

type model struct {
    input textinput.Model
}

func initialModel() model {
    ti := textinput.New()
    ti.Placeholder = "Enter your name..."
    ti.Focus()
    ti.CharLimit = 50
    ti.Width = 20
    
    return model{input: ti}
}

func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
    var cmd lib.Cmd
    m.input, cmd = m.input.Update(msg)
    return m, cmd
}

func (m model) View() string {
    return m.input.View()
}
```

### API Reference

#### Creating a TextInput

```go
ti := textinput.New()
```

#### Configuration

```go
ti.Prompt = "> "              // Text before input (default: "> ")
ti.Placeholder = "Type here"  // Shown when empty
ti.Width = 20                 // Max visible characters (0 = unlimited)
ti.CharLimit = 100            // Max total characters (0 = unlimited)
```

#### Methods

**Focus Management:**
```go
ti.Focus()           // Enable input focus
ti.Blur()            // Disable input focus
focused := ti.Focused()  // Check focus state
```

**Value Management:**
```go
ti.SetValue("text")  // Set the input value
value := ti.Value()  // Get the input value
ti.Reset()           // Clear the input
```

**Cursor Management:**
```go
ti.SetCursor(5)      // Move cursor to position
pos := ti.Position() // Get cursor position
ti.CursorStart()     // Move to start
ti.CursorEnd()       // Move to end
```

#### Keyboard Controls

- **Left/Right Arrow** - Move cursor
- **Home** - Move to start
- **End** - Move to end
- **Backspace** - Delete character before cursor
- **Delete** - Delete character at cursor
- **Any character** - Insert at cursor

### Example

See [examples/textinput-form/](../examples/textinput-form/) for a complete form example.

---

## Spinner

An animated loading spinner with multiple predefined styles.

### Features

- Multiple spinner styles (line, dot, globe, moon, etc.)
- Configurable animation speed
- Automatic frame advancement
- Unique ID for message routing

### Basic Usage

```go
import (
    "github.com/bubblegum/components/spinner"
    "time"
)

type model struct {
    spinner spinner.Model
}

func initialModel() model {
    s := spinner.New(spinner.WithSpinner(spinner.Dot))
    return model{spinner: s}
}

func (m model) Init() lib.Cmd {
    return m.spinner.Tick()
}

func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
    var cmd lib.Cmd
    m.spinner, cmd = m.spinner.Update(msg)
    return m, cmd
}

func (m model) View() string {
    return m.spinner.View() + " Loading..."
}
```

### API Reference

#### Creating a Spinner

```go
s := spinner.New()                                    // Default (Line)
s := spinner.New(spinner.WithSpinner(spinner.Dot))  // With specific style
```

#### Predefined Spinners

```go
spinner.Line        // |, /, -, \
spinner.Dot         // ⣾ ⣽ ⣻ ⢿ ⡿ ⣟ ⣯ ⣷
spinner.MiniDot     // ⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏
spinner.Jump        // ⢄ ⢂ ⢁ ⡁ ⡈ ⡐ ⡠
spinner.Pulse       // █ ▓ ▒ ░
spinner.Points      // ∙∙∙ ●∙∙ ∙●∙ ∙∙●
spinner.Globe       // 🌍 🌎 🌏
spinner.Moon        // 🌑 🌒 🌓 🌔 🌕 🌖 🌗 🌘
spinner.Monkey      // 🙈 🙉 🙊
spinner.Meter       // ▱▱▱ ▰▱▱ ▰▰▱ ▰▰▰
spinner.Hamburger   // ☱ ☲ ☴
spinner.Ellipsis    // . .. ...
```

#### Custom Spinner

```go
customSpinner := spinner.Spinner{
    Frames: []string{"◐", "◓", "◑", "◒"},
    FPS:    time.Second / 10,
}
s := spinner.New(spinner.WithSpinner(customSpinner))
```

#### Methods

```go
id := s.ID()         // Get unique spinner ID
frame := s.View()    // Get current frame
cmd := s.Tick()      // Get tick command for Init()
```

### Important Notes

1. **Always call Tick() in Init()** to start the animation:
   ```go
   func (m model) Init() lib.Cmd {
       return m.spinner.Tick()
   }
   ```

2. **Update the spinner** to advance frames:
   ```go
   m.spinner, cmd = m.spinner.Update(msg)
   ```

### Example

See [examples/components/](../examples/components/) for spinner usage.

---

## List

A scrollable list component with selection, filtering, and keyboard navigation.

### Features

- Scrollable item list
- Keyboard navigation
- Item selection
- Built-in filtering (press `/`)
- Custom item types
- Automatic scroll adjustment

### Basic Usage

```go
import "github.com/bubblegum/components/list"

type model struct {
    list list.Model
}

func initialModel() model {
    items := []list.Item{
        list.NewDefaultItem("Item 1", "Description 1"),
        list.NewDefaultItem("Item 2", "Description 2"),
        list.NewDefaultItem("Item 3", "Description 3"),
    }
    
    l := list.New(items, 80, 24)
    l.Title = "My List"
    
    return model{list: l}
}

func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
    var cmd lib.Cmd
    m.list, cmd = m.list.Update(msg)
    
    // Check for selection
    if msg, ok := msg.(lib.KeyMsg); ok && msg.Type == lib.KeyEnter {
        selected := m.list.SelectedItem()
        // Handle selection
    }
    
    return m, cmd
}

func (m model) View() string {
    return m.list.View()
}
```

### API Reference

#### Creating a List

```go
items := []list.Item{...}
l := list.New(items, width, height)
l.Title = "My List"
```

#### Item Interface

Implement the `Item` interface for custom items:

```go
type Item interface {
    FilterValue() string  // Value used for filtering
}
```

#### DefaultItem

Built-in item type with title and description:

```go
item := list.NewDefaultItem("Title", "Description")
title := item.Title()
desc := item.Description()
```

#### Methods

**Item Management:**
```go
l.SetItems(items)           // Set all items
items := l.Items()          // Get all items
visible := l.VisibleItems() // Get filtered items
```

**Selection:**
```go
item := l.SelectedItem()    // Get selected item
index := l.Index()          // Get selected index
l.SetCursor(5)              // Set cursor position
```

**Navigation:**
```go
l.CursorUp()                // Move up
l.CursorDown()              // Move down
```

**Sizing:**
```go
l.SetSize(width, height)    // Set dimensions
```

**Filtering:**
```go
l.StartFiltering()          // Enter filter mode
l.StopFiltering()           // Exit filter mode
l.SetFilter("search")       // Set filter text
```

#### Keyboard Controls

**Browsing Mode:**
- **Up/Down Arrow** - Navigate items
- **Page Up/Down** - Jump by page
- **Home/End** - Jump to start/end
- **/** - Start filtering
- **Enter** - Select item

**Filtering Mode:**
- **Type** - Add to filter
- **Backspace** - Remove from filter
- **Enter** - Accept filter
- **Esc** - Cancel filter

### Custom Items

Create custom item types:

```go
type myItem struct {
    name string
    data interface{}
}

func (i myItem) FilterValue() string {
    return i.name
}

// Use in list
items := []list.Item{
    myItem{name: "Item 1", data: someData},
    myItem{name: "Item 2", data: otherData},
}
```

### Example

See [examples/list-browser/](../examples/list-browser/) for a complete list example.

---

## Viewport

A scrollable content area for displaying large amounts of text.

### Features

- Vertical scrolling
- Mouse wheel support
- Keyboard navigation
- Automatic content wrapping
- Scroll position tracking

### Basic Usage

```go
import "github.com/bubblegum/components/viewport"

type model struct {
    viewport viewport.Model
    content  string
}

func initialModel() model {
    vp := viewport.New(80, 24)
    vp.SetContent("Your long content here...\n" +
                  "Multiple lines...\n" +
                  "...")
    
    return model{viewport: vp}
}

func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
    var cmd lib.Cmd
    m.viewport, cmd = m.viewport.Update(msg)
    return m, cmd
}

func (m model) View() string {
    return m.viewport.View()
}
```

### API Reference

#### Creating a Viewport

```go
vp := viewport.New(width, height)
```

#### Configuration

```go
vp.MouseWheelEnabled = true  // Enable mouse wheel (default: true)
vp.MouseWheelDelta = 3       // Lines per wheel event (default: 3)
```

#### Methods

**Content Management:**
```go
vp.SetContent(text)          // Set content
lines := vp.TotalLineCount() // Get total lines
visible := vp.VisibleLineCount() // Get visible lines
```

**Sizing:**
```go
vp.SetSize(width, height)    // Set dimensions
```

**Scrolling:**
```go
vp.ScrollUp(n)               // Scroll up n lines
vp.ScrollDown(n)             // Scroll down n lines
vp.PageUp()                  // Scroll up one page
vp.PageDown()                // Scroll down one page
vp.HalfPageUp()              // Scroll up half page
vp.HalfPageDown()            // Scroll down half page
vp.GotoTop()                 // Jump to top
vp.GotoBottom()              // Jump to bottom
```

**Position Queries:**
```go
atTop := vp.AtTop()          // Check if at top
atBottom := vp.AtBottom()    // Check if at bottom
percent := vp.ScrollPercent() // Get scroll position (0.0-1.0)
offset := vp.YOffset         // Get current offset
```

#### Keyboard Controls

- **Up/Down Arrow** - Scroll one line
- **Page Up/Down** - Scroll one page
- **Home** - Jump to top
- **End** - Jump to bottom

#### Mouse Controls

- **Wheel Up** - Scroll up
- **Wheel Down** - Scroll down

### Dynamic Content

Update content dynamically:

```go
func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
    switch msg := msg.(type) {
    case dataMsg:
        m.viewport.SetContent(msg.text)
    }
    
    m.viewport, cmd = m.viewport.Update(msg)
    return m, cmd
}
```

### Example

See [examples/components/](../examples/components/) for viewport usage.

---

## Differences from Bubbles

BubbleGum components maintain API compatibility with Bubbles components, with these differences:

### Import Paths

**Bubbles:**
```go
import "github.com/charmbracelet/bubbles/textinput"
```

**BubbleGum:**
```go
import "github.com/bubblegum/components/textinput"
```

### Message Types

Use BubbleGum message types:

**Bubbles:**
```go
case tea.KeyMsg:
case tea.MouseMsg:
```

**BubbleGum:**
```go
case lib.KeyMsg:
case lib.MouseMsg:
```

### Mouse Support

Mouse events are always enabled in BubbleGum - no need to enable mouse mode.

### Rendering

Components render to graphical windows instead of terminals, but the visual appearance is preserved through ANSI escape sequence support.

### Window Sizing

Handle `WindowSizeMsg` to make components responsive:

```go
case lib.WindowSizeMsg:
    m.list.SetSize(msg.Width, msg.Height)
    m.viewport.SetSize(msg.Width, msg.Height)
```

## Best Practices

### 1. Always Update Components

Pass messages to components and use returned commands:

```go
var cmd lib.Cmd
m.input, cmd = m.input.Update(msg)
return m, cmd
```

### 2. Handle Window Resizing

Update component sizes on window resize:

```go
case lib.WindowSizeMsg:
    m.list.SetSize(msg.Width, msg.Height-2) // Reserve space for header
```

### 3. Focus Management

Only focused components should receive input:

```go
if m.inputFocused {
    m.input, cmd = m.input.Update(msg)
} else {
    m.list, cmd = m.list.Update(msg)
}
```

### 4. Start Spinner Animation

Always call `Tick()` in `Init()` for spinners:

```go
func (m model) Init() lib.Cmd {
    return m.spinner.Tick()
}
```

### 5. Combine Components

Components work well together:

```go
type model struct {
    list    list.Model
    input   textinput.Model
    spinner spinner.Model
}
```

## See Also

- [API Documentation](../docs/API.md) - Complete API reference
- [Examples](../examples/) - Working example applications
- [Porting Guide](../docs/PORTING.md) - Porting from Bubble Tea
