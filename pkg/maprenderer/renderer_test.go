package maprenderer

import (
	"bytes"
	"image/color"
	"testing"

	"github.com/szydell/arkadia-mapsnap/pkg/mapparser"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Width != 800 {
		t.Errorf("Expected default width 800, got %d", cfg.Width)
	}
	if cfg.Height != 600 {
		t.Errorf("Expected default height 600, got %d", cfg.Height)
	}
	if cfg.RoomSize != 20 {
		t.Errorf("Expected default room size 20, got %d", cfg.RoomSize)
	}
	if len(cfg.DefaultEnvColors) != 16 {
		t.Errorf("Expected 16 default env colors, got %d", len(cfg.DefaultEnvColors))
	}
	// Z-level display should be off by default (Mudlet behavior)
	if cfg.ShowUpperLevel {
		t.Error("ShowUpperLevel should be false by default")
	}
	if cfg.ShowLowerLevel {
		t.Error("ShowLowerLevel should be false by default")
	}
}

func TestEnvToColor(t *testing.T) {
	defaultColors := defaultEnvironmentColors()
	customColors := map[int32]color.RGBA{}

	tests := []struct {
		env      int32
		expected color.RGBA
	}{
		{1, color.RGBA{R: 128, G: 0, B: 0, A: 255}},      // Red
		{2, color.RGBA{R: 0, G: 128, B: 0, A: 255}},      // Green
		{9, color.RGBA{R: 255, G: 0, B: 0, A: 255}},      // Light Red
		{16, color.RGBA{R: 128, G: 128, B: 128, A: 255}}, // Light Black (gray)
	}

	for _, tt := range tests {
		result := envToColor(tt.env, customColors, defaultColors)
		if result != tt.expected {
			t.Errorf("envToColor(%d) = %v, expected %v", tt.env, result, tt.expected)
		}
	}
}

func TestEnvToColorANSI256(t *testing.T) {
	defaultColors := defaultEnvironmentColors()
	customColors := map[int32]color.RGBA{}

	// Test ANSI 256 color cube (17-231) - note: 16 is in default colors
	// Color 17 = rgb(0,0,95), Color 21 = rgb(0,0,255)
	result := envToColor(17, customColors, defaultColors)
	if result.R != 0 || result.G != 0 || result.B != 95 {
		t.Errorf("envToColor(17) should be (0,0,95) in ANSI cube, got %v", result)
	}

	// Test grayscale (232-255)
	result = envToColor(232, customColors, defaultColors)
	expected := uint8(8)
	if result.R != expected || result.G != expected || result.B != expected {
		t.Errorf("envToColor(232) = %v, expected grayscale %d", result, expected)
	}
}

func TestEnvToColorCustom(t *testing.T) {
	defaultColors := defaultEnvironmentColors()
	customColors := map[int32]color.RGBA{
		100: {R: 255, G: 128, B: 64, A: 255},
	}

	result := envToColor(100, customColors, defaultColors)
	expected := color.RGBA{R: 255, G: 128, B: 64, A: 255}
	if result != expected {
		t.Errorf("envToColor(100) with custom = %v, expected %v", result, expected)
	}
}

func TestEnvToColorFallback(t *testing.T) {
	defaultColors := defaultEnvironmentColors()
	customColors := map[int32]color.RGBA{}

	// Test fallback for unknown env (should return gray)
	result := envToColor(-1, customColors, defaultColors)
	expected := color.RGBA{R: 128, G: 128, B: 128, A: 255}
	if result != expected {
		t.Errorf("envToColor(-1) = %v, expected fallback gray %v", result, expected)
	}

	// Test that env 999 (not in defaults, not in custom, not in ANSI) falls back to gray
	result = envToColor(999, customColors, defaultColors)
	if result != expected {
		t.Errorf("envToColor(999) = %v, expected fallback gray %v", result, expected)
	}
}

func TestGetEnvColorWithMudletFallback(t *testing.T) {
	// Test that getEnvColor falls back to env=1 (red) for unknown environments
	// This matches Mudlet behavior where unknown envs default to red
	r := NewRenderer(nil)
	m := mapparser.NewMudletMap()
	m.Areas[1] = mapparser.NewMudletArea(1, "Test")
	room := mapparser.NewMudletRoom(1)
	room.Area = 1
	room.Environment = -1 // Unknown environment
	m.Rooms[1] = room
	r.SetMap(m)

	customColors := map[int32]color.RGBA{}
	result := r.getEnvColor(-1, customColors)

	// Should fall back to env=1 (red = 128,0,0)
	expected := color.RGBA{R: 128, G: 0, B: 0, A: 255}
	if result != expected {
		t.Errorf("getEnvColor(-1) = %v, expected Mudlet fallback red %v", result, expected)
	}
}

func TestNewRenderer(t *testing.T) {
	// Test with nil config
	r := NewRenderer(nil)
	if r.config == nil {
		t.Error("NewRenderer(nil) should create default config")
	}

	// Test with custom config
	cfg := &Config{Width: 100, Height: 100}
	r = NewRenderer(cfg)
	if r.config.Width != 100 {
		t.Error("NewRenderer should use provided config")
	}
}

func TestRenderFragmentNoMap(t *testing.T) {
	r := NewRenderer(nil)
	_, err := r.RenderFragment(1)
	if err == nil {
		t.Error("RenderFragment without map should return error")
	}
}

func TestRenderFragmentRoomNotFound(t *testing.T) {
	r := NewRenderer(nil)
	m := mapparser.NewMudletMap()
	r.SetMap(m)

	_, err := r.RenderFragment(999)
	if err == nil {
		t.Error("RenderFragment with non-existent room should return error")
	}
}

func TestRenderFragmentBasic(t *testing.T) {
	r := NewRenderer(&Config{
		Width:            200,
		Height:           200,
		Radius:           5,
		RoomSize:         10,
		RoomSpacing:      15,
		DefaultEnvColors: defaultEnvironmentColors(),
		BackgroundColor:  color.RGBA{R: 30, G: 30, B: 30, A: 255},
		BorderColor:      color.RGBA{R: 100, G: 100, B: 100, A: 255},
		PlayerRoomColor:  color.RGBA{R: 255, G: 100, B: 100, A: 200},
		ExitColor:        color.RGBA{R: 180, G: 180, B: 180, A: 255},
	})

	m := mapparser.NewMudletMap()
	m.Areas[1] = mapparser.NewMudletArea(1, "Test Area")

	// Create a simple 3x3 grid of rooms
	for i := int32(0); i < 9; i++ {
		room := mapparser.NewMudletRoom(i + 1)
		room.Area = 1
		room.X = i % 3
		room.Y = i / 3
		room.Z = 0
		room.Environment = 1
		m.Rooms[i+1] = room
	}

	// Add some exits
	m.Rooms[1].Exits[mapparser.ExitEast] = 2
	m.Rooms[2].Exits[mapparser.ExitWest] = 1
	m.Rooms[2].Exits[mapparser.ExitEast] = 3
	m.Rooms[3].Exits[mapparser.ExitWest] = 2

	r.SetMap(m)

	result, err := r.RenderFragment(5) // Center room
	if err != nil {
		t.Fatalf("RenderFragment failed: %v", err)
	}

	if result.Image == nil {
		t.Error("RenderFragment should return an image")
	}
	if result.CenterRoom != 5 {
		t.Errorf("CenterRoom = %d, expected 5", result.CenterRoom)
	}
	if result.AreaID != 1 {
		t.Errorf("AreaID = %d, expected 1", result.AreaID)
	}
	if result.RoomsDrawn != 9 {
		t.Errorf("RoomsDrawn = %d, expected 9", result.RoomsDrawn)
	}

	// Check image dimensions
	bounds := result.Image.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 200 {
		t.Errorf("Image size = %dx%d, expected 200x200", bounds.Dx(), bounds.Dy())
	}
}

func TestOutputFormatFromPath(t *testing.T) {
	tests := []struct {
		path     string
		expected OutputFormat
	}{
		{"output.webp", FormatWEBP},
		{"output.WEBP", FormatWEBP},
		{"output.png", FormatPNG},
		{"output.PNG", FormatPNG},
		{"output.jpg", FormatWEBP}, // Default to WEBP
		{"output", FormatWEBP},     // No extension
	}

	for _, tt := range tests {
		result := FormatFromPath(tt.path)
		if result != tt.expected {
			t.Errorf("FormatFromPath(%q) = %d, expected %d", tt.path, result, tt.expected)
		}
	}
}

func TestWriteImageWEBP(t *testing.T) {
	r := NewRenderer(&Config{
		Width:            100,
		Height:           100,
		Radius:           2,
		RoomSize:         10,
		RoomSpacing:      15,
		DefaultEnvColors: defaultEnvironmentColors(),
		BackgroundColor:  color.RGBA{R: 30, G: 30, B: 30, A: 255},
		BorderColor:      color.RGBA{R: 100, G: 100, B: 100, A: 255},
		PlayerRoomColor:  color.RGBA{R: 255, G: 100, B: 100, A: 200},
		ExitColor:        color.RGBA{R: 180, G: 180, B: 180, A: 255},
	})

	m := mapparser.NewMudletMap()
	m.Areas[1] = mapparser.NewMudletArea(1, "Test")
	room := mapparser.NewMudletRoom(1)
	room.Area = 1
	m.Rooms[1] = room
	r.SetMap(m)

	result, err := r.RenderFragment(1)
	if err != nil {
		t.Fatalf("RenderFragment failed: %v", err)
	}

	var buf bytes.Buffer
	opts := &OutputOptions{Format: FormatWEBP, Quality: 80}
	err = WriteImage(result.Image, &buf, opts)
	if err != nil {
		t.Fatalf("WriteImage WEBP failed: %v", err)
	}

	// Check WEBP magic bytes (RIFF....WEBP)
	data := buf.Bytes()
	if len(data) < 12 {
		t.Fatal("WEBP output too small")
	}
	if string(data[0:4]) != "RIFF" || string(data[8:12]) != "WEBP" {
		t.Error("Invalid WEBP header")
	}
}

func TestWriteImagePNG(t *testing.T) {
	r := NewRenderer(&Config{
		Width:            100,
		Height:           100,
		Radius:           2,
		RoomSize:         10,
		RoomSpacing:      15,
		DefaultEnvColors: defaultEnvironmentColors(),
		BackgroundColor:  color.RGBA{R: 30, G: 30, B: 30, A: 255},
		BorderColor:      color.RGBA{R: 100, G: 100, B: 100, A: 255},
		PlayerRoomColor:  color.RGBA{R: 255, G: 100, B: 100, A: 200},
		ExitColor:        color.RGBA{R: 180, G: 180, B: 180, A: 255},
	})

	m := mapparser.NewMudletMap()
	m.Areas[1] = mapparser.NewMudletArea(1, "Test")
	room := mapparser.NewMudletRoom(1)
	room.Area = 1
	m.Rooms[1] = room
	r.SetMap(m)

	result, err := r.RenderFragment(1)
	if err != nil {
		t.Fatalf("RenderFragment failed: %v", err)
	}

	var buf bytes.Buffer
	opts := &OutputOptions{Format: FormatPNG}
	err = WriteImage(result.Image, &buf, opts)
	if err != nil {
		t.Fatalf("WriteImage PNG failed: %v", err)
	}

	// Check PNG magic bytes
	data := buf.Bytes()
	pngMagic := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if len(data) < 8 {
		t.Fatal("PNG output too small")
	}
	for i, b := range pngMagic {
		if data[i] != b {
			t.Error("Invalid PNG header")
			break
		}
	}
}

func TestDrawingPrimitives(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Width = 100
	cfg.Height = 100
	r := NewRenderer(cfg)
	m := mapparser.NewMudletMap()
	m.Areas[1] = mapparser.NewMudletArea(1, "Test")
	room := mapparser.NewMudletRoom(1)
	room.Area = 1
	m.Rooms[1] = room
	r.SetMap(m)

	result, _ := r.RenderFragment(1)
	img := result.Image

	// Test that background color is applied
	bgColor := cfg.BackgroundColor
	c := img.RGBAAt(0, 0)
	if c.R != bgColor.R || c.G != bgColor.G || c.B != bgColor.B {
		t.Errorf("Background color = %v, expected %v", c, bgColor)
	}
}

func TestCollectRoomsInArea(t *testing.T) {
	r := NewRenderer(nil)
	m := mapparser.NewMudletMap()

	// Create rooms in a 10x10 grid, all in area 1
	for x := int32(0); x < 10; x++ {
		for y := int32(0); y < 10; y++ {
			id := x*10 + y + 1
			room := mapparser.NewMudletRoom(id)
			room.X = x
			room.Y = y
			room.Z = 0
			room.Area = 1 // Set area ID
			m.Rooms[id] = room
		}
	}
	r.SetMap(m)

	// Collect rooms within radius 2 of center (5,5), areaID 1
	rooms := r.collectRoomsInArea(5, 5, 0, 2, 1)

	// Should be 5x5 = 25 rooms (from 3,3 to 7,7)
	if len(rooms) != 25 {
		t.Errorf("collectRoomsInArea returned %d rooms, expected 25", len(rooms))
	}

	// Test area filtering - rooms in area 2 should not be collected
	roomsWrongArea := r.collectRoomsInArea(5, 5, 0, 2, 2)
	if len(roomsWrongArea) != 0 {
		t.Errorf("collectRoomsInArea with wrong area returned %d rooms, expected 0", len(roomsWrongArea))
	}
}
