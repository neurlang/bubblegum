package lib

import (
	"fmt"

	cairo "github.com/neurlang/wayland/cairoshim"
)

// Renderer handles rendering a TerminalGrid to a Cairo surface.
type Renderer struct {
	font      *Font
	defaultFg Color
	defaultBg Color
	lastGrid  *TerminalGrid
}

// RendererOptions configures the renderer.
type RendererOptions struct {
	DefaultFg Color
	DefaultBg Color
}

// NewRenderer creates a new Renderer with the specified options.
func NewRenderer(opts RendererOptions) (*Renderer, error) {
	font, err := NewFont()
	if err != nil {
		return nil, fmt.Errorf("failed to load base font (ascii.png): %w (ensure font files are embedded)", err)
	}

	// Try to load extended fonts (optional, failures are ignored)
	_ = font.LoadExtendedFonts()

	return &Renderer{
		font:      font,
		defaultFg: opts.DefaultFg,
		defaultBg: opts.DefaultBg,
	}, nil
}

// CellWidth returns the width of a character cell in pixels.
func (r *Renderer) CellWidth() int32 {
	return int32(r.font.CellWidth())
}

// CellHeight returns the height of a character cell in pixels.
func (r *Renderer) CellHeight() int32 {
	return int32(r.font.CellHeight())
}

// Render renders the entire terminal grid to the Cairo surface.
func (r *Renderer) Render(grid *TerminalGrid, surface cairo.Surface) error {
	if grid == nil {
		err := fmt.Errorf("grid is nil")
		Error("Render failed: %v", err)
		return err
	}

	// Get surface dimensions
	width := surface.ImageSurfaceGetWidth()
	height := surface.ImageSurfaceGetHeight()
	
	if width <= 0 || height <= 0 {
		err := fmt.Errorf("invalid surface dimensions: %dx%d", width, height)
		Error("Render failed: %v", err)
		return err
	}

	Debug("Rendering grid: %dx%d cells to surface: %dx%d pixels", grid.Width, grid.Height, width, height)

	// Render all cells - continue even if individual cells fail
	for y := 0; y < grid.Height; y++ {
		for x := 0; x < grid.Width; x++ {
			cell := grid.Cells[y][x]
			r.renderCell(surface, x, y, cell)
		}
	}

	// Store this grid for future diff operations
	r.lastGrid = grid

	return nil
}

// RenderDiff renders only the changed regions of the terminal grid.
func (r *Renderer) RenderDiff(regions []Region, grid *TerminalGrid, surface cairo.Surface) error {
	if grid == nil {
		return fmt.Errorf("grid is nil")
	}

	for _, region := range regions {
		for y := region.Y; y < region.Y+region.Height && y < grid.Height; y++ {
			for x := region.X; x < region.X+region.Width && x < grid.Width; x++ {
				cell := grid.Cells[y][x]
				r.renderCell(surface, x, y, cell)
			}
		}
	}

	r.lastGrid = grid
	return nil
}

// renderCell renders a single cell at the specified grid position.
func (r *Renderer) renderCell(surface cairo.Surface, gridX, gridY int, cell Cell) {
	cellWidth := r.font.CellWidth()
	cellHeight := r.font.CellHeight()

	// Calculate pixel position
	pixelX := int32(gridX * cellWidth)
	pixelY := int32(gridY * cellHeight)

	// Get foreground and background colors
	fg := cell.FgColor
	if fg.IsDefault {
		fg = r.defaultFg
	}
	bg := cell.BgColor
	if bg.IsDefault {
		bg = r.defaultBg
	}

	// Get the character texture
	charStr := string(cell.Rune)
	texture := r.font.GetRGBTexture(charStr)

	// Handle missing glyph - texture will be nil or a placeholder
	if texture == nil {
		Debug("Missing glyph for character: %q (U+%04X), using space", charStr, cell.Rune)
		// Use space character as fallback
		texture = r.font.GetRGBTexture(" ")
		if texture == nil {
			// If even space is missing, skip rendering this cell
			Warn("Font missing space character, skipping cell at (%d, %d)", gridX, gridY)
			return
		}
	}

	// Render using the PutRGB method similar to the texteditor
	r.putRGB(surface, pixelX, pixelY, texture, cellWidth, cellHeight, 
		[3]byte{bg.R, bg.G, bg.B}, [3]byte{fg.R, fg.G, fg.B})
}

// putRGB renders an RGB texture to the Cairo surface at the specified position.
// This is adapted from wayland/go-wayland-texteditor/main.go
func (r *Renderer) putRGB(surface cairo.Surface, posX, posY int32, 
	textureRGB [][3]byte, textureWidth, textureHeight int, bg, fg [3]byte) {
	
	if textureRGB == nil {
		return
	}

	dst8 := surface.ImageSurfaceGetData()
	width := surface.ImageSurfaceGetWidth()
	height := surface.ImageSurfaceGetHeight()
	stride := surface.ImageSurfaceGetStride()

	// Render the texture
	for j := 0; j < textureWidth && posX+int32(j) < int32(width); j++ {
		for i := 0; i < textureHeight && posY+int32(i) < int32(height); i++ {
			dstPos := int(posY+int32(i))*stride + int(posX+int32(j))*4
			srcPos := i*textureWidth + j

			if srcPos >= len(textureRGB) {
				continue
			}

			// Cairo uses BGRA format
			dst8[dstPos] = textureRGB[srcPos][2]     // B
			dst8[dstPos+1] = textureRGB[srcPos][1]   // G
			dst8[dstPos+2] = textureRGB[srcPos][0]   // R
			dst8[dstPos+3] = 255                      // A

			// Apply background color (minimum values)
			if dst8[dstPos] < bg[2] {
				dst8[dstPos] = bg[2]
			}
			if dst8[dstPos+1] < bg[1] {
				dst8[dstPos+1] = bg[1]
			}
			if dst8[dstPos+2] < bg[0] {
				dst8[dstPos+2] = bg[0]
			}

			// Apply foreground color (maximum values)
			if dst8[dstPos] > fg[2] {
				dst8[dstPos] = fg[2]
			}
			if dst8[dstPos+1] > fg[1] {
				dst8[dstPos+1] = fg[1]
			}
			if dst8[dstPos+2] > fg[0] {
				dst8[dstPos+2] = fg[0]
			}
		}
	}
}
