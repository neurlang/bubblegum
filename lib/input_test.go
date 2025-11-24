package lib

import (
	"testing"

	"github.com/neurlang/wayland/window"
	"github.com/neurlang/wayland/wl"
	"github.com/neurlang/wayland/xkbcommon"
)

// Test the mapSpecialKey function directly since it doesn't depend on Input
func TestMapSpecialKey(t *testing.T) {
	tests := []struct {
		name      string
		keysym    uint32
		mods      window.ModType
		expected  KeyType
		isSpecial bool
	}{
		{"Enter", xkbcommon.KeyReturn, 0, KeyEnter, true},
		{"Backspace", xkbcommon.KeyBackspace, 0, KeyBackspace, true},
		{"Tab", xkbcommon.KeyTab, 0, KeyTab, true},
		{"Escape", xkbcommon.KeyEscape, 0, KeyEsc, true},
		{"Up Arrow", xkbcommon.KeyUp, 0, KeyUp, true},
		{"Down Arrow", xkbcommon.KeyDown, 0, KeyDown, true},
		{"Left Arrow", xkbcommon.KeyLeft, 0, KeyLeft, true},
		{"Right Arrow", xkbcommon.KeyRight, 0, KeyRight, true},
		{"Home", xkbcommon.KeyHome, 0, KeyHome, true},
		{"End", xkbcommon.KeyEnd, 0, KeyEnd, true},
		{"Page Up", xkbcommon.KeyPageUp, 0, KeyPgUp, true},
		{"Page Down", xkbcommon.KeyPageDown, 0, KeyPgDown, true},
		{"Delete", xkbcommon.KeyDelete, 0, KeyDelete, true},
		{"Insert", xkbcommon.KeyInsert, 0, KeyInsert, true},
		{"F1", xkbcommon.KeyF1, 0, KeyF1, true},
		{"F2", xkbcommon.KeyF2, 0, KeyF2, true},
		{"F12", xkbcommon.KeyF12, 0, KeyF12, true},
		{"Regular key", 'a', 0, KeyRunes, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyType, isSpecial := mapSpecialKey(tt.keysym, tt.mods)
			if isSpecial != tt.isSpecial {
				t.Errorf("Expected isSpecial=%v, got %v", tt.isSpecial, isSpecial)
			}
			if isSpecial && keyType != tt.expected {
				t.Errorf("Expected KeyType %v, got %v", tt.expected, keyType)
			}
		})
	}
}

func TestMapSpecialKey_CtrlCombinations(t *testing.T) {
	tests := []struct {
		name     string
		keysym   uint32
		expected KeyType
	}{
		{"Ctrl+C", 'c', KeyCtrlC},
		{"Ctrl+D", 'd', KeyCtrlD},
		{"Ctrl+L", 'l', KeyCtrlL},
		{"Ctrl+Z", 'z', KeyCtrlZ},
	}

	const ctrlMask window.ModType = 0x04 // ModControlMask

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyType, isSpecial := mapSpecialKey(tt.keysym, ctrlMask)
			if !isSpecial {
				t.Error("Expected Ctrl combination to be recognized as special")
			}
			if keyType != tt.expected {
				t.Errorf("Expected KeyType %v, got %v", tt.expected, keyType)
			}
		})
	}
}

func TestMapMouseButton_LeftClick(t *testing.T) {
	msg := MapMouseButton(100.0, 50.0, 272, wl.PointerButtonStatePressed, 10, 20)
	
	if msg == nil {
		t.Fatal("MapMouseButton returned nil")
	}
	
	if msg.X != 10 || msg.Y != 2 {
		t.Errorf("Expected position (10, 2), got (%d, %d)", msg.X, msg.Y)
	}
	
	if msg.Type != MousePress {
		t.Errorf("Expected MousePress, got %v", msg.Type)
	}
	
	if msg.Button != MouseButtonLeft {
		t.Errorf("Expected MouseButtonLeft, got %v", msg.Button)
	}
}

func TestMapMouseButton_RightClick(t *testing.T) {
	msg := MapMouseButton(50.0, 100.0, 273, wl.PointerButtonStatePressed, 10, 20)
	
	if msg == nil {
		t.Fatal("MapMouseButton returned nil")
	}
	
	if msg.Button != MouseButtonRight {
		t.Errorf("Expected MouseButtonRight, got %v", msg.Button)
	}
}

func TestMapMouseButton_MiddleClick(t *testing.T) {
	msg := MapMouseButton(50.0, 100.0, 274, wl.PointerButtonStatePressed, 10, 20)
	
	if msg == nil {
		t.Fatal("MapMouseButton returned nil")
	}
	
	if msg.Button != MouseButtonMiddle {
		t.Errorf("Expected MouseButtonMiddle, got %v", msg.Button)
	}
}

func TestMapMouseButton_Release(t *testing.T) {
	msg := MapMouseButton(100.0, 50.0, 272, wl.PointerButtonStateReleased, 10, 20)
	
	if msg == nil {
		t.Fatal("MapMouseButton returned nil")
	}
	
	if msg.Type != MouseRelease {
		t.Errorf("Expected MouseRelease, got %v", msg.Type)
	}
}

func TestMapMouseMotion(t *testing.T) {
	msg := MapMouseMotion(150.0, 80.0, 10, 20)
	
	if msg == nil {
		t.Fatal("MapMouseMotion returned nil")
	}
	
	if msg.X != 15 || msg.Y != 4 {
		t.Errorf("Expected position (15, 4), got (%d, %d)", msg.X, msg.Y)
	}
	
	if msg.Type != MouseMotion {
		t.Errorf("Expected MouseMotion, got %v", msg.Type)
	}
	
	if msg.Button != MouseButtonNone {
		t.Errorf("Expected MouseButtonNone, got %v", msg.Button)
	}
}

func TestMapMouseScroll_Vertical(t *testing.T) {
	// Scroll up (negative value)
	msgUp := MapMouseScroll(100.0, 50.0, 0, -1.0, 10, 20)
	if msgUp == nil {
		t.Fatal("MapMouseScroll returned nil for scroll up")
	}
	if msgUp.Type != MouseWheel {
		t.Errorf("Expected MouseWheel, got %v", msgUp.Type)
	}
	if msgUp.Button != MouseButtonWheelUp {
		t.Errorf("Expected MouseButtonWheelUp, got %v", msgUp.Button)
	}
	
	// Scroll down (positive value)
	msgDown := MapMouseScroll(100.0, 50.0, 0, 1.0, 10, 20)
	if msgDown == nil {
		t.Fatal("MapMouseScroll returned nil for scroll down")
	}
	if msgDown.Button != MouseButtonWheelDown {
		t.Errorf("Expected MouseButtonWheelDown, got %v", msgDown.Button)
	}
}

func TestMapMouseScroll_Horizontal(t *testing.T) {
	// Scroll left (negative value)
	msgLeft := MapMouseScroll(100.0, 50.0, 1, -1.0, 10, 20)
	if msgLeft == nil {
		t.Fatal("MapMouseScroll returned nil for scroll left")
	}
	if msgLeft.Button != MouseButtonWheelLeft {
		t.Errorf("Expected MouseButtonWheelLeft, got %v", msgLeft.Button)
	}
	
	// Scroll right (positive value)
	msgRight := MapMouseScroll(100.0, 50.0, 1, 1.0, 10, 20)
	if msgRight == nil {
		t.Fatal("MapMouseScroll returned nil for scroll right")
	}
	if msgRight.Button != MouseButtonWheelRight {
		t.Errorf("Expected MouseButtonWheelRight, got %v", msgRight.Button)
	}
}

func TestMapMouseScroll_CoordinateConversion(t *testing.T) {
	msg := MapMouseScroll(125.0, 65.0, 0, -1.0, 10, 20)
	
	if msg == nil {
		t.Fatal("MapMouseScroll returned nil")
	}
	
	if msg.X != 12 || msg.Y != 3 {
		t.Errorf("Expected position (12, 3), got (%d, %d)", msg.X, msg.Y)
	}
}
