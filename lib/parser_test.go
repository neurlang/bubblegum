package lib

import (
	"testing"
)

func TestParseANSI_BasicText(t *testing.T) {
	output := "Hello, World!"
	grid := ParseANSI(output, 20, 5)
	
	if grid == nil {
		t.Fatal("ParseANSI returned nil")
	}
	
	if grid.Width != 20 || grid.Height != 5 {
		t.Errorf("Expected grid dimensions 20x5, got %dx%d", grid.Width, grid.Height)
	}
	
	// Check first few characters
	expected := "Hello, World!"
	for i, ch := range expected {
		cell := grid.GetCell(i, 0)
		if cell == nil {
			t.Fatalf("GetCell(%d, 0) returned nil", i)
		}
		if cell.Rune != ch {
			t.Errorf("Cell[0][%d]: expected '%c', got '%c'", i, ch, cell.Rune)
		}
	}
}

func TestParseANSI_WithNewlines(t *testing.T) {
	output := "Line 1\nLine 2\nLine 3"
	grid := ParseANSI(output, 10, 5)
	
	if grid == nil {
		t.Fatal("ParseANSI returned nil")
	}
	
	// Check first line
	line1 := "Line 1"
	for i, ch := range line1 {
		cell := grid.GetCell(i, 0)
		if cell == nil {
			t.Fatalf("GetCell(%d, 0) returned nil", i)
		}
		if cell.Rune != ch {
			t.Errorf("Line 1, Cell[%d]: expected '%c', got '%c'", i, ch, cell.Rune)
		}
	}
	
	// Check second line
	line2 := "Line 2"
	for i, ch := range line2 {
		cell := grid.GetCell(i, 1)
		if cell == nil {
			t.Fatalf("GetCell(%d, 1) returned nil", i)
		}
		if cell.Rune != ch {
			t.Errorf("Line 2, Cell[%d]: expected '%c', got '%c'", i, ch, cell.Rune)
		}
	}
}

func TestParseANSI_WithColors(t *testing.T) {
	// Red text
	output := "\x1b[31mRed Text\x1b[0m"
	grid := ParseANSI(output, 20, 5)
	
	if grid == nil {
		t.Fatal("ParseANSI returned nil")
	}
	
	// Check that the first character has red foreground
	cell := grid.GetCell(0, 0)
	if cell == nil {
		t.Fatal("GetCell(0, 0) returned nil")
	}
	
	if cell.Rune != 'R' {
		t.Errorf("Expected 'R', got '%c'", cell.Rune)
	}
	
	if cell.FgColor.IsDefault {
		t.Error("Expected non-default foreground color")
	}
	
	// Red color should be (128, 0, 0) for standard ANSI red
	if cell.FgColor.R != 128 || cell.FgColor.G != 0 || cell.FgColor.B != 0 {
		t.Errorf("Expected red color (128, 0, 0), got (%d, %d, %d)",
			cell.FgColor.R, cell.FgColor.G, cell.FgColor.B)
	}
}

func TestParseANSI_WithBold(t *testing.T) {
	output := "\x1b[1mBold\x1b[0m"
	grid := ParseANSI(output, 20, 5)
	
	if grid == nil {
		t.Fatal("ParseANSI returned nil")
	}
	
	cell := grid.GetCell(0, 0)
	if cell == nil {
		t.Fatal("GetCell(0, 0) returned nil")
	}
	
	if !cell.Bold {
		t.Error("Expected bold text")
	}
}

func TestParseANSI_WithUnderline(t *testing.T) {
	output := "\x1b[4mUnderlined\x1b[0m"
	grid := ParseANSI(output, 20, 5)
	
	if grid == nil {
		t.Fatal("ParseANSI returned nil")
	}
	
	cell := grid.GetCell(0, 0)
	if cell == nil {
		t.Fatal("GetCell(0, 0) returned nil")
	}
	
	if !cell.Underline {
		t.Error("Expected underlined text")
	}
}

func TestTerminalGrid_Clear(t *testing.T) {
	grid := NewTerminalGrid(10, 5)
	if grid == nil {
		t.Fatal("NewTerminalGrid returned nil")
	}
	
	// Set some cells
	grid.SetCell(0, 0, Cell{Rune: 'A'})
	grid.SetCell(5, 2, Cell{Rune: 'B'})
	
	// Clear the grid
	grid.Clear()
	
	// Check that cells are reset
	cell1 := grid.GetCell(0, 0)
	if cell1.Rune != ' ' {
		t.Errorf("Expected space after clear, got '%c'", cell1.Rune)
	}
	
	cell2 := grid.GetCell(5, 2)
	if cell2.Rune != ' ' {
		t.Errorf("Expected space after clear, got '%c'", cell2.Rune)
	}
}

func TestTerminalGrid_Diff(t *testing.T) {
	grid1 := NewTerminalGrid(10, 5)
	grid2 := NewTerminalGrid(10, 5)
	
	// Initially identical
	regions := grid1.Diff(grid2)
	if len(regions) != 0 {
		t.Errorf("Expected no differences, got %d regions", len(regions))
	}
	
	// Change one cell
	grid2.SetCell(5, 2, Cell{Rune: 'X'})
	regions = grid1.Diff(grid2)
	
	if len(regions) == 0 {
		t.Error("Expected differences after changing a cell")
	}
}
