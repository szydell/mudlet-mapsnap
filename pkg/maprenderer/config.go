package maprenderer

import (
	"image/color"
)

// Config holds all rendering configuration options for the map renderer.
// Use [DefaultConfig] to create a configuration with sensible defaults.
type Config struct {
	// Image dimensions in pixels.
	Width  int
	Height int

	// Room appearance
	RoomSize     int  // Size of room square in pixels
	RoomSpacing  int  // Space between rooms
	RoomRound    bool // Draw rooms as circles instead of squares
	RoomBorder   bool // Draw border around rooms
	ShowRoomID   bool // Show room ID numbers
	ShowSymbol   bool // Show room symbols
	GridMode     bool // Use grid mode (smaller, no spacing)
	Antialiasing bool // Enable antialiasing

	// Exit appearance
	ExitWidth  float64 // Width of exit lines
	ExitColor  color.RGBA
	StubLength float64 // Length of stub exits

	// Colors
	BackgroundColor color.RGBA
	BorderColor     color.RGBA
	PlayerRoomColor color.RGBA
	TextColor       color.RGBA

	// Environment colors (fallback if not in map)
	DefaultEnvColors map[int32]color.RGBA

	// Z-level display
	ShowUpperLevel  bool
	ShowLowerLevel  bool
	UpperLevelAlpha uint8
	LowerLevelAlpha uint8
}

// DefaultConfig returns a configuration with sensible default values.
//
// Default settings:
//   - Image: 800x600 pixels
//   - Room size: 20 pixels with 25 pixel spacing
//   - Square rooms with borders
//   - Dark background (#1E1E1E)
//   - Antialiasing enabled
func DefaultConfig() *Config {
	return &Config{
		Width:  800,
		Height: 600,

		RoomSize:     20,
		RoomSpacing:  25,
		RoomRound:    false,
		RoomBorder:   true,
		ShowRoomID:   false,
		ShowSymbol:   true,
		GridMode:     false,
		Antialiasing: true,

		ExitWidth:  2.0,
		ExitColor:  color.RGBA{R: 180, G: 180, B: 180, A: 255},
		StubLength: 5.0,

		BackgroundColor: color.RGBA{R: 30, G: 30, B: 30, A: 255},
		BorderColor:     color.RGBA{R: 100, G: 100, B: 100, A: 255},
		PlayerRoomColor: color.RGBA{R: 255, G: 100, B: 100, A: 200},
		TextColor:       color.RGBA{R: 255, G: 255, B: 255, A: 255},

		DefaultEnvColors: defaultEnvironmentColors(),

		ShowUpperLevel:  false,
		ShowLowerLevel:  false,
		UpperLevelAlpha: 80,
		LowerLevelAlpha: 80,
	}
}

// CalculateVisibleRooms calculates how many rooms fit from center to edge
// in both horizontal and vertical directions.
//
// The calculation is based on image dimensions, room size, and room spacing.
// Returns (roomsX, roomsY) representing the number of rooms visible from
// the center to the edge in each direction.
func (c *Config) CalculateVisibleRooms() (int, int) {
	// Effective spacing is the distance between room centers
	// If spacing is less than room size, use room size as minimum
	effectiveSpacing := c.RoomSpacing
	if effectiveSpacing < c.RoomSize {
		effectiveSpacing = c.RoomSize
	}
	if effectiveSpacing < 1 {
		effectiveSpacing = 1
	}

	// Calculate available space from center to edge in each direction
	// Subtract half of room size as margin so edge rooms are fully visible
	margin := c.RoomSize / 2

	horizontalSpace := (c.Width / 2) - margin
	verticalSpace := (c.Height / 2) - margin

	if horizontalSpace < 0 {
		horizontalSpace = 0
	}
	if verticalSpace < 0 {
		verticalSpace = 0
	}

	// Calculate how many rooms fit in each direction
	roomsX := horizontalSpace / effectiveSpacing
	roomsY := verticalSpace / effectiveSpacing

	// Ensure at least 1 room in each direction
	if roomsX < 1 {
		roomsX = 1
	}
	if roomsY < 1 {
		roomsY = 1
	}

	return roomsX, roomsY
}

// defaultEnvironmentColors returns Mudlet's default 16 environment colors
func defaultEnvironmentColors() map[int32]color.RGBA {
	return map[int32]color.RGBA{
		1:  {R: 128, G: 0, B: 0, A: 255},     // Red
		2:  {R: 0, G: 128, B: 0, A: 255},     // Green
		3:  {R: 128, G: 128, B: 0, A: 255},   // Yellow
		4:  {R: 0, G: 0, B: 128, A: 255},     // Blue
		5:  {R: 128, G: 0, B: 128, A: 255},   // Magenta
		6:  {R: 0, G: 128, B: 128, A: 255},   // Cyan
		7:  {R: 192, G: 192, B: 192, A: 255}, // White (light gray)
		8:  {R: 64, G: 64, B: 64, A: 255},    // Black (dark gray)
		9:  {R: 255, G: 0, B: 0, A: 255},     // Light Red
		10: {R: 0, G: 255, B: 0, A: 255},     // Light Green
		11: {R: 255, G: 255, B: 0, A: 255},   // Light Yellow
		12: {R: 0, G: 0, B: 255, A: 255},     // Light Blue
		13: {R: 255, G: 0, B: 255, A: 255},   // Light Magenta
		14: {R: 0, G: 255, B: 255, A: 255},   // Light Cyan
		15: {R: 255, G: 255, B: 255, A: 255}, // Light White
		16: {R: 128, G: 128, B: 128, A: 255}, // Light Black (gray)
	}
}

// Mudlet uses ANSI 256-color palette for environments 17-255
// This function converts environment ID to color
func envToColor(env int32, customColors map[int32]color.RGBA, defaultColors map[int32]color.RGBA) color.RGBA {
	// Check default colors (1-16) FIRST (Mudlet behavior)
	if c, ok := defaultColors[env]; ok {
		return c
	}

	// Check custom colors
	if c, ok := customColors[env]; ok {
		return c
	}

	// ANSI 256-color palette (16-255)
	if env >= 16 && env < 232 {
		// 6x6x6 color cube (16-231)
		base := env - 16
		r := base / 36
		g := (base - (r * 36)) / 6
		b := base - (r * 36) - (g * 6)

		var rv, gv, bv uint8
		if r == 0 {
			rv = 0
		} else {
			rv = uint8((r-1)*40 + 95)
		}
		if g == 0 {
			gv = 0
		} else {
			gv = uint8((g-1)*40 + 95)
		}
		if b == 0 {
			bv = 0
		} else {
			bv = uint8((b-1)*40 + 95)
		}
		return color.RGBA{R: rv, G: gv, B: bv, A: 255}
	} else if env >= 232 && env < 256 {
		// Grayscale (232-255)
		k := uint8(((env - 232) * 10) + 8)
		return color.RGBA{R: k, G: k, B: k, A: 255}
	}

	// Fallback to gray
	return color.RGBA{R: 128, G: 128, B: 128, A: 255}
}
