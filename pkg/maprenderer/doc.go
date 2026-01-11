// Package maprenderer provides functionality for rendering Mudlet maps to images.
//
// This package generates visual map fragments from parsed Mudlet map data,
// supporting output to WEBP and PNG formats. It is implemented in pure Go
// with no CGO dependencies.
//
// # Basic Usage
//
// Render a map fragment centered on a specific room:
//
//	// Parse the map first
//	m, err := mapparser.ParseMapFile("world.map")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Create renderer with default configuration
//	cfg := maprenderer.DefaultConfig()
//	cfg.Width = 1024
//	cfg.Height = 768
//
//	renderer := maprenderer.NewRenderer(cfg)
//	renderer.SetMap(m)
//
//	// Render fragment centered on room 1234
//	result, err := renderer.RenderFragment(1234)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Save to file
//	err = maprenderer.SaveImage(result.Image, "map.webp", nil)
//
// # Configuration
//
// The [Config] struct controls rendering behavior:
//   - Image dimensions (Width, Height)
//   - Room appearance (RoomSize, RoomSpacing, RoomRound)
//   - Exit lines (ExitWidth, ExitColor)
//   - Colors (BackgroundColor, BorderColor, PlayerRoomColor)
//   - Z-level display (ShowUpperLevel, ShowLowerLevel)
//
// # Output Formats
//
// Supported output formats:
//   - WEBP: Lossless compression using pure Go encoder (default)
//   - PNG: Standard PNG with best compression
//
// The format is auto-detected from the file extension, or can be specified
// explicitly via [OutputOptions].
//
// # Environment Colors
//
// Room colors are determined by their environment ID. The renderer uses:
//  1. Mudlet's default 16 ANSI colors (environments 1-16)
//  2. Custom environment colors defined in the map file
//  3. ANSI 256-color palette for environments 17-255
//  4. Fallback gray for undefined environments
//
// # Labels
//
// Map labels (text and images) are rendered according to their ShowOnTop flag:
//   - Background labels: rendered under rooms and exits
//   - Foreground labels: rendered on top of everything
package maprenderer
