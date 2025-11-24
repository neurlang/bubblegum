package lib

import (
	"testing"
)

func TestNewRenderer(t *testing.T) {
	opts := RendererOptions{
		DefaultFg: NewColor(255, 255, 255),
		DefaultBg: NewColor(0, 0, 0),
	}

	renderer, err := NewRenderer(opts)
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}

	if renderer == nil {
		t.Fatal("Renderer is nil")
	}

	if renderer.font == nil {
		t.Fatal("Font is nil")
	}

	// Check that cell dimensions are reasonable
	cellWidth := renderer.CellWidth()
	cellHeight := renderer.CellHeight()

	if cellWidth <= 0 {
		t.Errorf("Cell width should be positive, got %d", cellWidth)
	}

	if cellHeight <= 0 {
		t.Errorf("Cell height should be positive, got %d", cellHeight)
	}

	t.Logf("Cell dimensions: %dx%d", cellWidth, cellHeight)
}

func TestFontGetRGBTexture(t *testing.T) {
	font, err := NewFont()
	if err != nil {
		t.Fatalf("Failed to create font: %v", err)
	}

	// Test ASCII character
	texture := font.GetRGBTexture("A")
	if texture == nil {
		t.Error("Texture for 'A' should not be nil")
	}

	expectedSize := font.CellWidth() * font.CellHeight()
	if len(texture) != expectedSize {
		t.Errorf("Texture size mismatch: expected %d, got %d", expectedSize, len(texture))
	}

	// Test unsupported character (may generate placeholder if cell size is large enough)
	texture = font.GetRGBTexture("🎉")
	// Placeholder generation requires cell size >= 12x24
	if font.CellWidth() >= 12 && font.CellHeight() >= 24 {
		if texture == nil {
			t.Error("Texture for unsupported character should generate placeholder when cell size is sufficient")
		}
		if len(texture) != expectedSize {
			t.Errorf("Placeholder texture size mismatch: expected %d, got %d", expectedSize, len(texture))
		}
	} else {
		// Small fonts can't generate placeholders
		if texture != nil {
			t.Logf("Placeholder generated despite small cell size: %dx%d", font.CellWidth(), font.CellHeight())
		}
	}
}

func TestFontLoadExtended(t *testing.T) {
	font, err := NewFont()
	if err != nil {
		t.Fatalf("Failed to create font: %v", err)
	}

	// Try to load extended fonts (may fail if files don't exist)
	err = font.LoadExtendedFonts()
	if err != nil {
		t.Logf("Extended fonts not loaded (this is OK): %v", err)
	}

	// Test that basic ASCII still works
	texture := font.GetRGBTexture("a")
	if texture == nil {
		t.Error("Basic ASCII should still work after attempting to load extended fonts")
	}
}

func TestRendererCellDimensions(t *testing.T) {
	opts := RendererOptions{
		DefaultFg: NewColor(255, 255, 255),
		DefaultBg: NewColor(0, 0, 0),
	}

	renderer, err := NewRenderer(opts)
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}

	cellWidth := renderer.CellWidth()
	cellHeight := renderer.CellHeight()

	// Verify dimensions match font dimensions
	if cellWidth != int32(renderer.font.CellWidth()) {
		t.Errorf("Cell width mismatch: renderer=%d, font=%d", cellWidth, renderer.font.CellWidth())
	}

	if cellHeight != int32(renderer.font.CellHeight()) {
		t.Errorf("Cell height mismatch: renderer=%d, font=%d", cellHeight, renderer.font.CellHeight())
	}
}
