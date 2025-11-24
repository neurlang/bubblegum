package lib

// Color represents an RGB color value for terminal rendering.
// IsDefault indicates whether to use the default terminal color.
type Color struct {
	R         uint8
	G         uint8
	B         uint8
	IsDefault bool
}

// DefaultColor returns a Color marked as default.
func DefaultColor() Color {
	return Color{IsDefault: true}
}

// NewColor creates a Color from RGB values.
func NewColor(r, g, b uint8) Color {
	return Color{R: r, G: g, B: b, IsDefault: false}
}

// Cell represents a single character cell in the terminal grid.
type Cell struct {
	Rune          rune
	FgColor       Color
	BgColor       Color
	Bold          bool
	Italic        bool
	Underline     bool
	Strikethrough bool
}

// NewCell creates a new Cell with default values.
func NewCell() Cell {
	return Cell{
		Rune:    ' ',
		FgColor: DefaultColor(),
		BgColor: DefaultColor(),
	}
}

// TerminalGrid represents a character-based grid where each cell contains
// a single character with styling information.
type TerminalGrid struct {
	Width  int
	Height int
	Cells  [][]Cell
}

// NewTerminalGrid creates a new TerminalGrid with the specified dimensions.
// All cells are initialized with default values (space character, default colors).
func NewTerminalGrid(width, height int) *TerminalGrid {
	if width <= 0 || height <= 0 {
		return nil
	}

	cells := make([][]Cell, height)
	for y := 0; y < height; y++ {
		cells[y] = make([]Cell, width)
		for x := 0; x < width; x++ {
			cells[y][x] = NewCell()
		}
	}

	return &TerminalGrid{
		Width:  width,
		Height: height,
		Cells:  cells,
	}
}

// GetCell returns the cell at the specified position.
// Returns nil if the position is out of bounds.
func (tg *TerminalGrid) GetCell(x, y int) *Cell {
	if x < 0 || x >= tg.Width || y < 0 || y >= tg.Height {
		return nil
	}
	return &tg.Cells[y][x]
}

// SetCell sets the cell at the specified position.
// Does nothing if the position is out of bounds.
func (tg *TerminalGrid) SetCell(x, y int, cell Cell) {
	if x < 0 || x >= tg.Width || y < 0 || y >= tg.Height {
		return
	}
	tg.Cells[y][x] = cell
}

// Clear resets all cells to their default values.
func (tg *TerminalGrid) Clear() {
	for y := 0; y < tg.Height; y++ {
		for x := 0; x < tg.Width; x++ {
			tg.Cells[y][x] = NewCell()
		}
	}
}

// ClearLine resets all cells in the specified line to default values.
func (tg *TerminalGrid) ClearLine(y int) {
	if y < 0 || y >= tg.Height {
		return
	}
	for x := 0; x < tg.Width; x++ {
		tg.Cells[y][x] = NewCell()
	}
}

// ClearFromCursor clears from the cursor position to the end of the line.
func (tg *TerminalGrid) ClearFromCursor(x, y int) {
	if y < 0 || y >= tg.Height {
		return
	}
	for i := x; i < tg.Width; i++ {
		tg.Cells[y][i] = NewCell()
	}
}

// Region represents a rectangular region in the terminal grid.
type Region struct {
	X      int
	Y      int
	Width  int
	Height int
}

// Diff compares this grid with another and returns regions that differ.
// This is used for differential rendering optimization.
func (tg *TerminalGrid) Diff(other *TerminalGrid) []Region {
	if other == nil || tg.Width != other.Width || tg.Height != other.Height {
		// If dimensions don't match, return the entire grid as changed
		return []Region{{X: 0, Y: 0, Width: tg.Width, Height: tg.Height}}
	}

	var regions []Region
	
	// Simple implementation: check each line for changes
	for y := 0; y < tg.Height; y++ {
		lineChanged := false
		startX := -1
		
		for x := 0; x < tg.Width; x++ {
			cellChanged := !cellsEqual(tg.Cells[y][x], other.Cells[y][x])
			
			if cellChanged && startX == -1 {
				startX = x
				lineChanged = true
			} else if !cellChanged && startX != -1 {
				// End of changed region
				regions = append(regions, Region{
					X:      startX,
					Y:      y,
					Width:  x - startX,
					Height: 1,
				})
				startX = -1
			}
		}
		
		// If we reached the end of the line with changes
		if lineChanged && startX != -1 {
			regions = append(regions, Region{
				X:      startX,
				Y:      y,
				Width:  tg.Width - startX,
				Height: 1,
			})
		}
	}
	
	return regions
}

// cellsEqual compares two cells for equality.
func cellsEqual(a, b Cell) bool {
	return a.Rune == b.Rune &&
		a.FgColor == b.FgColor &&
		a.BgColor == b.BgColor &&
		a.Bold == b.Bold &&
		a.Italic == b.Italic &&
		a.Underline == b.Underline &&
		a.Strikethrough == b.Strikethrough
}
