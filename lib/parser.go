package lib

import (
	"strconv"
	"strings"
)

// ParseANSI parses a string containing ANSI escape sequences and builds a TerminalGrid.
// The output string from View() is parsed to extract characters and styling information.
func ParseANSI(output string, width, height int) *TerminalGrid {
	grid := NewTerminalGrid(width, height)
	if grid == nil {
		return nil
	}

	parser := &ansiParser{
		grid:    grid,
		cursorX: 0,
		cursorY: 0,
		fgColor: DefaultColor(),
		bgColor: DefaultColor(),
	}

	parser.parse(output)
	return grid
}

// ansiParser maintains state while parsing ANSI sequences.
type ansiParser struct {
	grid          *TerminalGrid
	cursorX       int
	cursorY       int
	fgColor       Color
	bgColor       Color
	bold          bool
	italic        bool
	underline     bool
	strikethrough bool
}

// parse processes the input string and populates the grid.
func (p *ansiParser) parse(input string) {
	runes := []rune(input)
	i := 0
	for i < len(runes) {
		if runes[i] == '\x1b' && i+1 < len(runes) && runes[i+1] == '[' {
			// ANSI escape sequence
			seqEnd := p.findSequenceEnd(runes, i+2)
			if seqEnd >= i+2 {
				// Include the command character in the sequence
				seqStr := string(runes[i+2 : seqEnd+1])
				p.handleEscapeSequence(seqStr)
				i = seqEnd + 1
				continue
			}
		}

		// Regular character
		ch := runes[i]
		
		switch ch {
		case '\n':
			p.cursorX = 0
			p.cursorY++
		case '\r':
			p.cursorX = 0
		case '\t':
			// Tab moves to next multiple of 8
			p.cursorX = ((p.cursorX / 8) + 1) * 8
		default:
			if p.cursorY < p.grid.Height && p.cursorX < p.grid.Width {
				p.grid.Cells[p.cursorY][p.cursorX] = Cell{
					Rune:          ch,
					FgColor:       p.fgColor,
					BgColor:       p.bgColor,
					Bold:          p.bold,
					Italic:        p.italic,
					Underline:     p.underline,
					Strikethrough: p.strikethrough,
				}
			}
			p.cursorX++
		}

		// Handle line wrapping
		if p.cursorX >= p.grid.Width {
			p.cursorX = 0
			p.cursorY++
		}

		i++
	}
}

// findSequenceEnd finds the end of an ANSI escape sequence.
func (p *ansiParser) findSequenceEnd(runes []rune, start int) int {
	for i := start; i < len(runes); i++ {
		ch := runes[i]
		// Sequence ends with a letter
		if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
			return i
		}
	}
	return start
}

// handleEscapeSequence processes an ANSI escape sequence.
func (p *ansiParser) handleEscapeSequence(seq string) {
	if len(seq) == 0 {
		return
	}

	command := seq[len(seq)-1]
	params := ""
	if len(seq) > 1 {
		params = seq[:len(seq)-1]
	}

	switch command {
	case 'm': // SGR - Select Graphic Rendition
		p.handleSGR(params)
	case 'H', 'f': // Cursor position
		p.handleCursorPosition(params)
	case 'A': // Cursor up
		p.handleCursorUp(params)
	case 'B': // Cursor down
		p.handleCursorDown(params)
	case 'C': // Cursor forward
		p.handleCursorForward(params)
	case 'D': // Cursor back
		p.handleCursorBack(params)
	case 'J': // Erase in display
		p.handleEraseDisplay(params)
	case 'K': // Erase in line
		p.handleEraseLine(params)
	}
}

// handleSGR processes Select Graphic Rendition sequences (colors and styles).
func (p *ansiParser) handleSGR(params string) {
	if params == "" {
		params = "0"
	}

	codes := parseSGRParams(params)
	
	for i := 0; i < len(codes); i++ {
		code := codes[i]
		
		switch code {
		case 0: // Reset
			p.fgColor = DefaultColor()
			p.bgColor = DefaultColor()
			p.bold = false
			p.italic = false
			p.underline = false
			p.strikethrough = false
		case 1: // Bold
			p.bold = true
		case 3: // Italic
			p.italic = true
		case 4: // Underline
			p.underline = true
		case 9: // Strikethrough
			p.strikethrough = true
		case 22: // Normal intensity (not bold)
			p.bold = false
		case 23: // Not italic
			p.italic = false
		case 24: // Not underlined
			p.underline = false
		case 29: // Not strikethrough
			p.strikethrough = false
		case 30, 31, 32, 33, 34, 35, 36, 37: // Foreground colors (8 colors)
			p.fgColor = ansi16Color(code - 30)
		case 38: // Extended foreground color
			if i+1 < len(codes) {
				if codes[i+1] == 5 && i+2 < len(codes) {
					// 256-color mode
					p.fgColor = ansi256Color(codes[i+2])
					i += 2
				} else if codes[i+1] == 2 && i+4 < len(codes) {
					// RGB mode
					p.fgColor = NewColor(uint8(codes[i+2]), uint8(codes[i+3]), uint8(codes[i+4]))
					i += 4
				}
			}
		case 39: // Default foreground color
			p.fgColor = DefaultColor()
		case 40, 41, 42, 43, 44, 45, 46, 47: // Background colors (8 colors)
			p.bgColor = ansi16Color(code - 40)
		case 48: // Extended background color
			if i+1 < len(codes) {
				if codes[i+1] == 5 && i+2 < len(codes) {
					// 256-color mode
					p.bgColor = ansi256Color(codes[i+2])
					i += 2
				} else if codes[i+1] == 2 && i+4 < len(codes) {
					// RGB mode
					p.bgColor = NewColor(uint8(codes[i+2]), uint8(codes[i+3]), uint8(codes[i+4]))
					i += 4
				}
			}
		case 49: // Default background color
			p.bgColor = DefaultColor()
		case 90, 91, 92, 93, 94, 95, 96, 97: // Bright foreground colors
			p.fgColor = ansi16Color(code - 90 + 8)
		case 100, 101, 102, 103, 104, 105, 106, 107: // Bright background colors
			p.bgColor = ansi16Color(code - 100 + 8)
		}
	}
}

// parseSGRParams parses semicolon-separated SGR parameters.
func parseSGRParams(params string) []int {
	if params == "" {
		return []int{0}
	}
	
	parts := strings.Split(params, ";")
	codes := make([]int, 0, len(parts))
	
	for _, part := range parts {
		if part == "" {
			codes = append(codes, 0)
			continue
		}
		if num, err := strconv.Atoi(part); err == nil {
			codes = append(codes, num)
		}
	}
	
	return codes
}

// ansi16Color returns the RGB color for a 16-color ANSI code (0-15).
func ansi16Color(code int) Color {
	// Standard ANSI color palette
	colors := []Color{
		NewColor(0, 0, 0),       // 0: Black
		NewColor(128, 0, 0),     // 1: Red
		NewColor(0, 128, 0),     // 2: Green
		NewColor(128, 128, 0),   // 3: Yellow
		NewColor(0, 0, 128),     // 4: Blue
		NewColor(128, 0, 128),   // 5: Magenta
		NewColor(0, 128, 128),   // 6: Cyan
		NewColor(192, 192, 192), // 7: White
		NewColor(128, 128, 128), // 8: Bright Black (Gray)
		NewColor(255, 0, 0),     // 9: Bright Red
		NewColor(0, 255, 0),     // 10: Bright Green
		NewColor(255, 255, 0),   // 11: Bright Yellow
		NewColor(0, 0, 255),     // 12: Bright Blue
		NewColor(255, 0, 255),   // 13: Bright Magenta
		NewColor(0, 255, 255),   // 14: Bright Cyan
		NewColor(255, 255, 255), // 15: Bright White
	}
	
	if code >= 0 && code < len(colors) {
		return colors[code]
	}
	return DefaultColor()
}

// ansi256Color returns the RGB color for a 256-color ANSI code.
func ansi256Color(code int) Color {
	if code < 0 || code > 255 {
		return DefaultColor()
	}
	
	// First 16 colors are the standard ANSI colors
	if code < 16 {
		return ansi16Color(code)
	}
	
	// Colors 16-231 are a 6x6x6 RGB cube
	if code >= 16 && code <= 231 {
		code -= 16
		r := (code / 36) * 51
		g := ((code % 36) / 6) * 51
		b := (code % 6) * 51
		return NewColor(uint8(r), uint8(g), uint8(b))
	}
	
	// Colors 232-255 are grayscale
	if code >= 232 {
		gray := uint8((code-232)*10 + 8)
		return NewColor(gray, gray, gray)
	}
	
	return DefaultColor()
}

// handleCursorPosition moves the cursor to the specified position.
func (p *ansiParser) handleCursorPosition(params string) {
	coords := parseSGRParams(params)
	if len(coords) == 0 {
		p.cursorX = 0
		p.cursorY = 0
		return
	}
	
	y := 0
	x := 0
	if len(coords) >= 1 {
		y = coords[0] - 1 // ANSI uses 1-based indexing
	}
	if len(coords) >= 2 {
		x = coords[1] - 1
	}
	
	if y < 0 {
		y = 0
	}
	if x < 0 {
		x = 0
	}
	
	p.cursorY = y
	p.cursorX = x
}

// handleCursorUp moves the cursor up by n lines.
func (p *ansiParser) handleCursorUp(params string) {
	n := 1
	if params != "" {
		if num, err := strconv.Atoi(params); err == nil && num > 0 {
			n = num
		}
	}
	p.cursorY -= n
	if p.cursorY < 0 {
		p.cursorY = 0
	}
}

// handleCursorDown moves the cursor down by n lines.
func (p *ansiParser) handleCursorDown(params string) {
	n := 1
	if params != "" {
		if num, err := strconv.Atoi(params); err == nil && num > 0 {
			n = num
		}
	}
	p.cursorY += n
}

// handleCursorForward moves the cursor forward by n columns.
func (p *ansiParser) handleCursorForward(params string) {
	n := 1
	if params != "" {
		if num, err := strconv.Atoi(params); err == nil && num > 0 {
			n = num
		}
	}
	p.cursorX += n
}

// handleCursorBack moves the cursor back by n columns.
func (p *ansiParser) handleCursorBack(params string) {
	n := 1
	if params != "" {
		if num, err := strconv.Atoi(params); err == nil && num > 0 {
			n = num
		}
	}
	p.cursorX -= n
	if p.cursorX < 0 {
		p.cursorX = 0
	}
}

// handleEraseDisplay clears parts of the display.
func (p *ansiParser) handleEraseDisplay(params string) {
	mode := 0
	if params != "" {
		if num, err := strconv.Atoi(params); err == nil {
			mode = num
		}
	}
	
	switch mode {
	case 0: // Clear from cursor to end of screen
		p.grid.ClearFromCursor(p.cursorX, p.cursorY)
		for y := p.cursorY + 1; y < p.grid.Height; y++ {
			p.grid.ClearLine(y)
		}
	case 1: // Clear from cursor to beginning of screen
		for y := 0; y < p.cursorY; y++ {
			p.grid.ClearLine(y)
		}
		for x := 0; x <= p.cursorX && x < p.grid.Width; x++ {
			p.grid.Cells[p.cursorY][x] = NewCell()
		}
	case 2, 3: // Clear entire screen
		p.grid.Clear()
	}
}

// handleEraseLine clears parts of the current line.
func (p *ansiParser) handleEraseLine(params string) {
	mode := 0
	if params != "" {
		if num, err := strconv.Atoi(params); err == nil {
			mode = num
		}
	}
	
	if p.cursorY < 0 || p.cursorY >= p.grid.Height {
		return
	}
	
	switch mode {
	case 0: // Clear from cursor to end of line
		p.grid.ClearFromCursor(p.cursorX, p.cursorY)
	case 1: // Clear from beginning of line to cursor
		for x := 0; x <= p.cursorX && x < p.grid.Width; x++ {
			p.grid.Cells[p.cursorY][x] = NewCell()
		}
	case 2: // Clear entire line
		p.grid.ClearLine(p.cursorY)
	}
}
