package lib

import (
	"github.com/neurlang/wayland/window"
	"github.com/neurlang/wayland/wl"
	"github.com/neurlang/wayland/xkbcommon"
)

// MapKeyboardEvent converts a Wayland keyboard event to a Bubble Tea KeyMsg.
// It handles special keys, modifiers, and character input.
func MapKeyboardEvent(input *window.Input, keysym uint32, key uint32, mods window.ModType, state wl.KeyboardKeyState) *KeyMsg {
	// Only process key press events
	if state != wl.KeyboardKeyStatePressed {
		return nil
	}

	// Check for Alt modifier
	hasAlt := (mods & window.ModAltMask) != 0

	// Map special keys first
	keyType, isSpecial := mapSpecialKey(keysym, mods)
	if isSpecial {
		return &KeyMsg{
			Type: keyType,
			Alt:  hasAlt,
		}
	}

	// Try to get a rune from the key
	r := input.GetRune(&keysym, key)
	if r != 0 {
		return &KeyMsg{
			Type:  KeyRunes,
			Runes: []rune{r},
			Alt:   hasAlt,
		}
	}

	// If we couldn't map the key, return nil
	return nil
}

// mapSpecialKey maps Wayland keysyms to Bubble Tea KeyType values.
// It returns the KeyType and a boolean indicating if the key was mapped.
func mapSpecialKey(keysym uint32, mods window.ModType) (KeyType, bool) {
	hasCtrl := (mods & window.ModControlMask) != 0

	// Handle Ctrl+key combinations
	if hasCtrl {
		switch keysym {
		case 'c', 'C':
			return KeyCtrlC, true
		case 'd', 'D':
			return KeyCtrlD, true
		case 'l', 'L':
			return KeyCtrlL, true
		case 'z', 'Z':
			return KeyCtrlZ, true
		}
	}

	// Map special keys
	switch keysym {
	case xkbcommon.KeyReturn, xkbcommon.KeyKpEnter:
		return KeyEnter, true
	case xkbcommon.KeyBackspace:
		return KeyBackspace, true
	case xkbcommon.KeyTab:
		return KeyTab, true
	case xkbcommon.KeyEscape:
		return KeyEsc, true
	case xkbcommon.KeyUp:
		return KeyUp, true
	case xkbcommon.KeyDown:
		return KeyDown, true
	case xkbcommon.KeyLeft:
		return KeyLeft, true
	case xkbcommon.KeyRight:
		return KeyRight, true
	case xkbcommon.KeyHome:
		return KeyHome, true
	case xkbcommon.KeyEnd:
		return KeyEnd, true
	case xkbcommon.KeyPageUp:
		return KeyPgUp, true
	case xkbcommon.KeyPageDown:
		return KeyPgDown, true
	case xkbcommon.KeyDelete:
		return KeyDelete, true
	case xkbcommon.KeyInsert:
		return KeyInsert, true
	case xkbcommon.KeyF1:
		return KeyF1, true
	case xkbcommon.KeyF2:
		return KeyF2, true
	case xkbcommon.KeyF3:
		return KeyF3, true
	case xkbcommon.KeyF4:
		return KeyF4, true
	case xkbcommon.KeyF5:
		return KeyF5, true
	case xkbcommon.KeyF6:
		return KeyF6, true
	case xkbcommon.KeyF7:
		return KeyF7, true
	case xkbcommon.KeyF8:
		return KeyF8, true
	case xkbcommon.KeyF9:
		return KeyF9, true
	case xkbcommon.KeyF10:
		return KeyF10, true
	case xkbcommon.KeyF11:
		return KeyF11, true
	case xkbcommon.KeyF12:
		return KeyF12, true
	}

	return KeyRunes, false
}

// MapMouseButton converts a Wayland pointer button event to a Bubble Tea MouseMsg.
// It handles button presses and releases.
func MapMouseButton(x, y float32, button uint32, state wl.PointerButtonState, cellWidth, cellHeight int32) *MouseMsg {
	// Convert pixel coordinates to cell positions
	cellX := int(x / float32(cellWidth))
	cellY := int(y / float32(cellHeight))

	// Determine event type
	var eventType MouseEventType
	if state == wl.PointerButtonStatePressed {
		eventType = MousePress
	} else {
		eventType = MouseRelease
	}

	// Map button codes (Linux input event codes)
	var mouseButton MouseButton
	switch button {
	case 272: // BTN_LEFT
		mouseButton = MouseButtonLeft
	case 273: // BTN_RIGHT
		mouseButton = MouseButtonRight
	case 274: // BTN_MIDDLE
		mouseButton = MouseButtonMiddle
	default:
		mouseButton = MouseButtonNone
	}

	return &MouseMsg{
		X:      cellX,
		Y:      cellY,
		Type:   eventType,
		Button: mouseButton,
	}
}

// MapMouseMotion converts a Wayland pointer motion event to a Bubble Tea MouseMsg.
func MapMouseMotion(x, y float32, cellWidth, cellHeight int32) *MouseMsg {
	// Convert pixel coordinates to cell positions
	cellX := int(x / float32(cellWidth))
	cellY := int(y / float32(cellHeight))

	return &MouseMsg{
		X:      cellX,
		Y:      cellY,
		Type:   MouseMotion,
		Button: MouseButtonNone,
	}
}

// MapMouseScroll converts a Wayland pointer axis (scroll) event to a Bubble Tea MouseMsg.
// The axis parameter indicates the scroll direction (vertical or horizontal).
// The value parameter indicates the scroll amount (positive or negative).
func MapMouseScroll(x, y float32, axis uint32, value float32, cellWidth, cellHeight int32) *MouseMsg {
	// Convert pixel coordinates to cell positions
	cellX := int(x / float32(cellWidth))
	cellY := int(y / float32(cellHeight))

	// Determine scroll direction
	// axis 0 = vertical, axis 1 = horizontal
	var mouseButton MouseButton
	if axis == 0 { // Vertical scroll
		if value < 0 {
			mouseButton = MouseButtonWheelUp
		} else {
			mouseButton = MouseButtonWheelDown
		}
	} else { // Horizontal scroll
		if value < 0 {
			mouseButton = MouseButtonWheelLeft
		} else {
			mouseButton = MouseButtonWheelRight
		}
	}

	return &MouseMsg{
		X:      cellX,
		Y:      cellY,
		Type:   MouseWheel,
		Button: mouseButton,
	}
}

