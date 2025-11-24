# BubbleGum Troubleshooting Guide

This guide covers common issues you might encounter when using BubbleGum and how to resolve them.

## Table of Contents

- [Installation Issues](#installation-issues)
- [Runtime Errors](#runtime-errors)
- [Display Issues](#display-issues)
- [Input Issues](#input-issues)
- [Performance Issues](#performance-issues)
- [Build Issues](#build-issues)
- [Platform-Specific Issues](#platform-specific-issues)

---

## Installation Issues

### Missing Wayland Libraries (Linux)

**Error:**
```
package github.com/neurlang/wayland/window: C source files not allowed when not using cgo
```

**Solution:**

Install the required development libraries:

**Ubuntu/Debian:**
```bash
sudo apt-get install libwayland-dev libxkbcommon-dev libcairo2-dev
```

**Fedora:**
```bash
sudo dnf install wayland-devel libxkbcommon-devel cairo-devel
```

**Arch Linux:**
```bash
sudo pacman -S wayland libxkbcommon cairo
```

### CGO Not Enabled

**Error:**
```
cgo: C compiler "gcc" not found
```

**Solution:**

1. Install a C compiler:
   ```bash
   # Ubuntu/Debian
   sudo apt-get install build-essential
   
   # Fedora
   sudo dnf install gcc
   
   # Arch Linux
   sudo pacman -S base-devel
   ```

2. Ensure CGO is enabled:
   ```bash
   export CGO_ENABLED=1
   go build
   ```

### Go Version Too Old

**Error:**
```
go: module requires Go 1.21 or later
```

**Solution:**

Update Go to version 1.21 or later:
```bash
# Download from https://go.dev/dl/
# Or use your package manager
```

---

## Runtime Errors

### Failed to Create Wayland Display

**Error:**
```
failed to create Wayland display: ... (ensure Wayland compositor is running)
```

**Cause:** No Wayland compositor is running, or the application can't connect to it.

**Solutions:**

1. **Check if Wayland is running:**
   ```bash
   echo $WAYLAND_DISPLAY
   # Should output something like "wayland-0"
   ```

2. **If using X11, switch to Wayland:**
   - Log out and select "Wayland" session at login screen
   - Or install a Wayland compositor (GNOME, KDE Plasma, Sway)

3. **If running remotely:**
   - Wayland doesn't support remote displays like X11
   - Run the application locally or use X11 forwarding with a terminal version

4. **Check permissions:**
   ```bash
   ls -la $XDG_RUNTIME_DIR/$WAYLAND_DISPLAY
   # Should be accessible by your user
   ```

### Window Creation Failed

**Error:**
```
failed to create window: window.Create returned nil
```

**Cause:** Window system initialization failed.

**Solutions:**

1. Check compositor logs for errors
2. Try a different Wayland compositor
3. Ensure your graphics drivers are up to date
4. Check system resources (memory, file descriptors)

### Font Loading Failed

**Error:**
```
failed to create renderer: font loading error
```

**Cause:** Font system can't find or load the specified font.

**Solutions:**

1. **Use a common font:**
   ```go
   lib.WithFontFamily("Monospace")  // Usually available
   lib.WithFontFamily("DejaVu Sans Mono")
   ```

2. **Install missing fonts:**
   ```bash
   # Ubuntu/Debian
   sudo apt-get install fonts-dejavu-core
   
   # Fedora
   sudo dnf install dejavu-sans-mono-fonts
   ```

3. **List available fonts:**
   ```bash
   fc-list : family | sort | uniq
   ```

### Panic in Update/View

**Error:**
```
Panic in Update(): runtime error: index out of range
```

**Cause:** Your code panicked in Init, Update, or View.

**What BubbleGum Does:**
- Catches the panic
- Logs the error and stack trace
- Exits gracefully
- Cleans up window resources

**Solution:**

1. Check the logged stack trace to find the issue
2. Add bounds checking in your code
3. Use defensive programming:
   ```go
   func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
       if m.cursor >= len(m.items) {
           m.cursor = len(m.items) - 1
       }
       // ... rest of code
   }
   ```

---

## Display Issues

### Window Appears Blank

**Symptoms:** Window opens but shows nothing or is all black.

**Causes and Solutions:**

1. **View returns empty string:**
   ```go
   func (m model) View() string {
       if m.content == "" {
           return "Loading..."  // Always return something
       }
       return m.content
   }
   ```

2. **ANSI codes are malformed:**
   - Check your ANSI escape sequences
   - Use a library like lipgloss for reliable styling

3. **Window size is zero:**
   ```go
   case lib.WindowSizeMsg:
       if msg.Width == 0 || msg.Height == 0 {
           return m, nil  // Wait for valid size
       }
       m.width = msg.Width
       m.height = msg.Height
   ```

### Text Rendering Issues

**Symptoms:** Characters appear as boxes, missing, or garbled.

**Solutions:**

1. **Unicode characters not supported:**
   - The font may not include the characters you're using
   - Try a different font with better Unicode coverage:
     ```go
     lib.WithFontFamily("DejaVu Sans Mono")
     ```

2. **Font size too small:**
   ```go
   lib.WithFontSize(14)  // Increase from default 12
   ```

3. **Check character encoding:**
   - Ensure your source files are UTF-8 encoded
   - Verify strings are valid UTF-8

### Colors Not Appearing

**Symptoms:** All text is the same color, ANSI colors ignored.

**Solutions:**

1. **Check ANSI escape sequences:**
   ```go
   // Correct
   "\x1b[31mRed text\x1b[0m"
   
   // Wrong
   "\\x1b[31mRed text\\x1b[0m"  // Escaped backslash
   ```

2. **Reset after styling:**
   ```go
   "\x1b[1;32mGreen Bold\x1b[0m Normal"
   //                          ^^^^^ Reset is important
   ```

3. **Use lipgloss for reliable styling:**
   ```go
   import "github.com/charmbracelet/lipgloss"
   
   style := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
   text := style.Render("Styled text")
   ```

### Window Doesn't Resize Properly

**Symptoms:** Content doesn't adapt when window is resized.

**Solution:**

Always handle `WindowSizeMsg`:

```go
func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
    switch msg := msg.(type) {
    case lib.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        
        // Update components
        m.list.SetSize(msg.Width, msg.Height-2)
        m.viewport.SetSize(msg.Width, msg.Height)
        
        return m, nil
    }
    return m, nil
}
```

---

## Input Issues

### Keyboard Input Not Working

**Symptoms:** Key presses don't do anything.

**Solutions:**

1. **Check message handling:**
   ```go
   case lib.KeyMsg:
       switch msg.Type {
       case lib.KeyEsc:
           return m, lib.Quit
       // ... handle other keys
       }
   ```

2. **Component not focused:**
   ```go
   ti := textinput.New()
   ti.Focus()  // Don't forget to focus!
   ```

3. **Wrong key type check:**
   ```go
   // Wrong
   if msg.String() == "q" { ... }
   
   // Correct
   case lib.KeyMsg:
       switch msg.Type {
       case lib.KeyRunes:
           if string(msg.Runes) == "q" { ... }
       }
   ```

### Mouse Input Not Working

**Symptoms:** Mouse clicks and scrolling don't register.

**Solutions:**

1. **Check message handling:**
   ```go
   case lib.MouseMsg:
       if msg.Type == lib.MousePress {
           // Handle click at msg.X, msg.Y
       }
   ```

2. **Mouse coordinates out of bounds:**
   ```go
   case lib.MouseMsg:
       if msg.X < 0 || msg.X >= m.width || msg.Y < 0 || msg.Y >= m.height {
           return m, nil  // Ignore out-of-bounds clicks
       }
   ```

3. **Component doesn't handle mouse:**
   - Not all components respond to mouse events
   - Check component documentation

### Special Keys Not Working

**Symptoms:** Function keys, Ctrl+key combinations don't work.

**Solution:**

Use the correct key constants:

```go
case lib.KeyMsg:
    switch msg.Type {
    case lib.KeyF1:
        // Handle F1
    case lib.KeyCtrlC:
        return m, lib.Quit
    case lib.KeyCtrlD:
        // Handle Ctrl+D
    }
```

---

## Performance Issues

### High CPU Usage

**Symptoms:** Application uses excessive CPU even when idle.

**Causes and Solutions:**

1. **No frame rate limiting:**
   ```go
   p := lib.NewProgram(
       model{},
       lib.WithFPS(30),  // Limit to 30 FPS
   )
   ```

2. **Continuous updates:**
   ```go
   // Bad - returns command that triggers immediate update
   func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
       return m, func() lib.Msg { return updateMsg{} }
   }
   
   // Good - only update when needed
   func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
       return m, nil
   }
   ```

3. **Expensive View function:**
   ```go
   // Cache expensive computations
   func (m model) View() string {
       if m.cachedView != "" && !m.dirty {
           return m.cachedView
       }
       m.cachedView = m.computeView()
       m.dirty = false
       return m.cachedView
   }
   ```

### Slow Rendering

**Symptoms:** UI feels sluggish, low frame rate.

**Solutions:**

1. **Simplify View output:**
   - Reduce the amount of text rendered
   - Avoid complex ANSI sequences

2. **Increase FPS limit:**
   ```go
   lib.WithFPS(60)  // Higher frame rate
   ```

3. **Optimize Update logic:**
   - Avoid expensive operations in Update
   - Use commands for heavy work

### Memory Usage Growing

**Symptoms:** Memory usage increases over time.

**Solutions:**

1. **Clean up commands:**
   ```go
   // Ensure goroutines exit
   func fetchDataCmd(ctx context.Context) lib.Cmd {
       return func() lib.Msg {
           select {
           case <-ctx.Done():
               return nil
           case data := <-fetchData():
               return dataMsg{data}
           }
       }
   }
   ```

2. **Limit stored data:**
   ```go
   // Don't accumulate unbounded data
   if len(m.history) > 1000 {
       m.history = m.history[len(m.history)-1000:]
   }
   ```

---

## Build Issues

### Undefined References

**Error:**
```
undefined: lib.NewProgram
```

**Solution:**

1. Check import path:
   ```go
   import "github.com/bubblegum/lib"
   ```

2. Run `go mod tidy`:
   ```bash
   go mod tidy
   go build
   ```

### Version Conflicts

**Error:**
```
module requires different version
```

**Solution:**

```bash
go get github.com/bubblegum/lib@latest
go mod tidy
```

### Cross-Compilation Issues

**Error:**
```
C compiler not available for cross compilation
```

**Cause:** CGO doesn't support simple cross-compilation.

**Solution:**

Build on the target platform, or use a cross-compilation toolchain:

```bash
# For cross-compiling to Linux from macOS
# This is complex and requires target libraries
# Easier to build on target platform
```

---

## Platform-Specific Issues

### Linux: Permission Denied

**Error:**
```
permission denied accessing /run/user/1000/wayland-0
```

**Solution:**

1. Check user permissions:
   ```bash
   groups  # Should include 'video' or similar
   ```

2. Check XDG_RUNTIME_DIR:
   ```bash
   echo $XDG_RUNTIME_DIR
   ls -la $XDG_RUNTIME_DIR
   ```

3. Restart session if needed

### Linux: No Wayland Compositor

**Error:**
```
failed to create Wayland display
```

**Solution:**

Install and run a Wayland compositor:

```bash
# GNOME (Wayland session)
# KDE Plasma (Wayland session)
# Or install Sway
sudo apt-get install sway
sway
```

### Windows: Not Yet Supported

**Status:** Windows support is planned but not yet implemented.

**Workaround:**

Use WSL2 with a Wayland compositor:

```bash
# In WSL2
sudo apt-get install weston
export DISPLAY=:0
weston &
./your-bubblegum-app
```

---

## Debugging Tips

### Enable Debug Logging

BubbleGum includes debug logging. Check the source for logging functions:

```go
// In your code
lib.Debug("Current state: %+v", m)
lib.Info("Processing message: %T", msg)
lib.Error("Something went wrong: %v", err)
```

### Check Error Messages

Always check errors:

```go
if _, err := p.Run(); err != nil {
    log.Printf("Error: %v", err)
    os.Exit(1)
}
```

### Minimal Reproduction

Create a minimal example that reproduces the issue:

```go
package main

import "github.com/bubblegum/lib"

type model struct{}

func (m model) Init() lib.Cmd { return nil }
func (m model) Update(msg lib.Msg) (lib.Model, lib.Cmd) {
    switch msg.(type) {
    case lib.KeyMsg:
        return m, lib.Quit
    }
    return m, nil
}
func (m model) View() string { return "Test" }

func main() {
    p := lib.NewProgram(model{})
    p.Run()
}
```

### Check System Resources

```bash
# Check memory
free -h

# Check file descriptors
ulimit -n

# Check compositor logs
journalctl -xe | grep -i wayland
```

---

## Getting Help

If you're still stuck:

1. **Check the documentation:**
   - [API Documentation](API.md)
   - [Porting Guide](PORTING.md)
   - [Component Documentation](../components/README.md)

2. **Review examples:**
   - Look at [examples/](../examples/) for working code
   - Compare your code to the examples

3. **Search existing issues:**
   - Check GitHub issues for similar problems
   - Someone may have already solved it

4. **Create a new issue:**
   - Include your OS and Go version
   - Provide a minimal reproduction
   - Include error messages and logs
   - Describe what you expected vs. what happened

5. **Community resources:**
   - Bubble Tea community (for general TUI questions)
   - Wayland community (for compositor issues)

---

## Common Error Messages Reference

| Error Message | Likely Cause | Solution |
|--------------|--------------|----------|
| `failed to create Wayland display` | No Wayland compositor | Start Wayland session |
| `window.Create returned nil` | Window creation failed | Check compositor logs |
| `font loading error` | Font not found | Install font or use different one |
| `invalid configuration` | Bad options | Check option values |
| `Panic in Update()` | Code bug | Check stack trace, add bounds checking |
| `Message channel full` | Too many messages | Increase buffer or reduce message rate |
| `C compiler not found` | CGO disabled or no compiler | Install gcc, enable CGO |
| `undefined: lib.NewProgram` | Wrong import | Use `github.com/bubblegum/lib` |

---

## Prevention Tips

### 1. Always Handle WindowSizeMsg

```go
case lib.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
```

### 2. Validate Input

```go
if index < 0 || index >= len(m.items) {
    return m, nil
}
```

### 3. Use Frame Rate Limiting

```go
lib.WithFPS(30)
```

### 4. Test on Target Platform

Build and test on the platform where you'll deploy.

### 5. Handle Errors

```go
if _, err := p.Run(); err != nil {
    log.Fatal(err)
}
```

### 6. Keep Dependencies Updated

```bash
go get -u ./...
go mod tidy
```

---

## Summary

Most issues fall into these categories:

1. **Installation** - Missing libraries or wrong Go version
2. **Runtime** - Wayland compositor not running
3. **Display** - Font or rendering issues
4. **Input** - Incorrect message handling
5. **Performance** - Missing frame rate limiting

Check the relevant section above for solutions, and don't hesitate to ask for help if you're stuck!
