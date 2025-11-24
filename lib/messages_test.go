package lib

import (
	"strings"
	"testing"
)

func TestKeyMsg_String(t *testing.T) {
	tests := []struct {
		name     string
		msg      KeyMsg
		contains []string
	}{
		{
			name: "runes message",
			msg: KeyMsg{
				Type:  KeyRunes,
				Runes: []rune("hello"),
				Alt:   false,
			},
			contains: []string{"Runes", "hello", "Alt: false"},
		},
		{
			name: "special key message",
			msg: KeyMsg{
				Type: KeyEnter,
				Alt:  true,
			},
			contains: []string{"Type", "Alt: true"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.msg.String()
			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("KeyMsg.String() = %q, should contain %q", result, substr)
				}
			}
		})
	}
}

func TestMouseMsg_String(t *testing.T) {
	msg := MouseMsg{
		X:      10,
		Y:      20,
		Type:   MousePress,
		Button: MouseButtonLeft,
	}

	result := msg.String()
	expected := []string{"X: 10", "Y: 20", "Type", "Button"}

	for _, substr := range expected {
		if !strings.Contains(result, substr) {
			t.Errorf("MouseMsg.String() = %q, should contain %q", result, substr)
		}
	}
}

func TestWindowSizeMsg_String(t *testing.T) {
	msg := WindowSizeMsg{
		Width:  800,
		Height: 600,
	}

	result := msg.String()
	expected := []string{"Width: 800", "Height: 600"}

	for _, substr := range expected {
		if !strings.Contains(result, substr) {
			t.Errorf("WindowSizeMsg.String() = %q, should contain %q", result, substr)
		}
	}
}

func TestQuitMsg_String(t *testing.T) {
	msg := QuitMsg{}
	result := msg.String()

	if result != "QuitMsg{}" {
		t.Errorf("QuitMsg.String() = %q, want %q", result, "QuitMsg{}")
	}
}
