# Mouse Interactive Example

This example demonstrates mouse event handling in BubbleGum applications.

## Features

- Mouse motion tracking (displays current cell position)
- Click recording (left, middle, right buttons)
- Scroll event handling (wheel up/down/left/right)
- Event history (last 10 events)

## Known Issues

- Very fast horizontal mouse movement may cause the UI to temporarily freeze
- This is due to the interaction between the Wayland event loop and redraw scheduling
- Slow to moderate mouse movement works correctly

## Usage

```bash
go run examples/mouse/main.go
```

Or with debug logging:

```bash
BUBBLEGUM_DEBUG=1 go run examples/mouse/main.go
```

## Controls

- **Mouse Movement**: Updates position display
- **Click**: Records click event
- **Scroll**: Records scroll event
- **Enter**: Clear event history
- **Esc**: Quit
