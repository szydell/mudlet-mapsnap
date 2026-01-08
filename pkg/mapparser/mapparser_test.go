package mapparser

import (
	"os"
	"testing"
)

// Test fixtures paths
const (
	smallMapPath = "../../tests/fixtures/2_rooms_map/2lok.dat"
	largeMapPath = "../../tests/fixtures/large_maps/2025-05-27#15-06-15map.dat"
)

// TestNewMudletMap tests MudletMap constructor
func TestNewMudletMap(t *testing.T) {
	m := NewMudletMap()

	if m == nil {
		t.Fatal("NewMudletMap returned nil")
	}
	if m.EnvColors == nil {
		t.Error("EnvColors should be initialized")
	}
	if m.CustomEnvColors == nil {
		t.Error("CustomEnvColors should be initialized")
	}
	if m.Areas == nil {
		t.Error("Areas should be initialized")
	}
	if m.Rooms == nil {
		t.Error("Rooms should be initialized")
	}
	if m.Labels == nil {
		t.Error("Labels should be initialized")
	}
}

// TestNewMudletArea tests MudletArea constructor
func TestNewMudletArea(t *testing.T) {
	area := NewMudletArea(42, "Test Area")

	if area == nil {
		t.Fatal("NewMudletArea returned nil")
	}
	if area.ID != 42 {
		t.Errorf("Expected ID 42, got %d", area.ID)
	}
	if area.Name != "Test Area" {
		t.Errorf("Expected name 'Test Area', got %q", area.Name)
	}
	if area.Rooms == nil {
		t.Error("Rooms should be initialized")
	}
	if area.ZLevels == nil {
		t.Error("ZLevels should be initialized")
	}
	if area.UserData == nil {
		t.Error("UserData should be initialized")
	}
}

// TestNewMudletRoom tests MudletRoom constructor
func TestNewMudletRoom(t *testing.T) {
	room := NewMudletRoom(123)

	if room == nil {
		t.Fatal("NewMudletRoom returned nil")
	}
	if room.ID != 123 {
		t.Errorf("Expected ID 123, got %d", room.ID)
	}
	if room.Weight != 1 {
		t.Errorf("Expected default weight 1, got %d", room.Weight)
	}

	// All exits should be NoExit by default
	for i, exit := range room.Exits {
		if exit != NoExit {
			t.Errorf("Exit %d should be NoExit (-1), got %d", i, exit)
		}
	}

	// Maps should be initialized
	if room.SpecialExits == nil {
		t.Error("SpecialExits should be initialized")
	}
	if room.UserData == nil {
		t.Error("UserData should be initialized")
	}
	if room.CustomLines == nil {
		t.Error("CustomLines should be initialized")
	}
	if room.Doors == nil {
		t.Error("Doors should be initialized")
	}
}

// TestRoomExitMethods tests MudletRoom exit-related methods
func TestRoomExitMethods(t *testing.T) {
	room := NewMudletRoom(1)

	// Initially no exits
	if room.HasExit(ExitNorth) {
		t.Error("Room should not have north exit initially")
	}
	if len(room.ActiveExits()) != 0 {
		t.Error("Room should have no active exits initially")
	}

	// Add north exit
	room.Exits[ExitNorth] = 100

	if !room.HasExit(ExitNorth) {
		t.Error("Room should have north exit after setting it")
	}
	if room.GetExit(ExitNorth) != 100 {
		t.Errorf("North exit should be 100, got %d", room.GetExit(ExitNorth))
	}

	activeExits := room.ActiveExits()
	if len(activeExits) != 1 {
		t.Errorf("Expected 1 active exit, got %d", len(activeExits))
	}
	if activeExits[0] != ExitNorth {
		t.Errorf("Active exit should be north (0), got %d", activeExits[0])
	}

	// Add more exits
	room.Exits[ExitSouth] = 200
	room.Exits[ExitUp] = 300

	activeExits = room.ActiveExits()
	if len(activeExits) != 3 {
		t.Errorf("Expected 3 active exits, got %d", len(activeExits))
	}

	// Test invalid direction
	if room.GetExit(-1) != NoExit {
		t.Error("Invalid direction should return NoExit")
	}
	if room.GetExit(20) != NoExit {
		t.Error("Out of range direction should return NoExit")
	}
}

// TestExitDirectionNames tests exit direction name arrays
func TestExitDirectionNames(t *testing.T) {
	if len(ExitDirectionNames) != 12 {
		t.Errorf("Expected 12 exit direction names, got %d", len(ExitDirectionNames))
	}
	if len(ExitDirectionShortNames) != 12 {
		t.Errorf("Expected 12 short exit direction names, got %d", len(ExitDirectionShortNames))
	}

	// Spot check some names
	if ExitDirectionNames[ExitNorth] != "north" {
		t.Errorf("Expected 'north' at index 0, got %q", ExitDirectionNames[ExitNorth])
	}
	if ExitDirectionShortNames[ExitNorth] != "n" {
		t.Errorf("Expected 'n' at index 0, got %q", ExitDirectionShortNames[ExitNorth])
	}
	if ExitDirectionNames[ExitUp] != "up" {
		t.Errorf("Expected 'up' at index 8, got %q", ExitDirectionNames[ExitUp])
	}
}

// TestColorToRGBA tests Color.ToRGBA method
func TestColorToRGBA(t *testing.T) {
	// Full red color (16-bit values)
	c := Color{
		Red:   0xFFFF,
		Green: 0x0000,
		Blue:  0x0000,
		Alpha: 0xFFFF,
	}

	r, g, b, a := c.ToRGBA()

	if r != 255 {
		t.Errorf("Expected red 255, got %d", r)
	}
	if g != 0 {
		t.Errorf("Expected green 0, got %d", g)
	}
	if b != 0 {
		t.Errorf("Expected blue 0, got %d", b)
	}
	if a != 255 {
		t.Errorf("Expected alpha 255, got %d", a)
	}

	// Half intensity
	c2 := Color{
		Red:   0x8000,
		Green: 0x8000,
		Blue:  0x8000,
		Alpha: 0x8000,
	}

	r, g, b, a = c2.ToRGBA()
	if r != 128 {
		t.Errorf("Expected red 128, got %d", r)
	}
}

// TestMudletMapMethods tests MudletMap helper methods
func TestMudletMapMethods(t *testing.T) {
	m := NewMudletMap()

	// Add some test data
	m.Areas[1] = NewMudletArea(1, "Area 1")
	m.Areas[2] = NewMudletArea(2, "Area 2")

	m.Rooms[100] = NewMudletRoom(100)
	m.Rooms[100].Area = 1
	m.Rooms[101] = NewMudletRoom(101)
	m.Rooms[101].Area = 1
	m.Rooms[200] = NewMudletRoom(200)
	m.Rooms[200].Area = 2

	// Test counts
	if m.RoomCount() != 3 {
		t.Errorf("Expected 3 rooms, got %d", m.RoomCount())
	}
	if m.AreaCount() != 2 {
		t.Errorf("Expected 2 areas, got %d", m.AreaCount())
	}

	// Test GetRoom
	room := m.GetRoom(100)
	if room == nil {
		t.Error("GetRoom(100) should return room")
	}
	if m.GetRoom(999) != nil {
		t.Error("GetRoom(999) should return nil for non-existent room")
	}

	// Test GetArea
	area := m.GetArea(1)
	if area == nil {
		t.Error("GetArea(1) should return area")
	}
	if m.GetArea(999) != nil {
		t.Error("GetArea(999) should return nil for non-existent area")
	}

	// Test GetRoomsInArea
	roomsInArea1 := m.GetRoomsInArea(1)
	if len(roomsInArea1) != 2 {
		t.Errorf("Expected 2 rooms in area 1, got %d", len(roomsInArea1))
	}

	roomsInArea2 := m.GetRoomsInArea(2)
	if len(roomsInArea2) != 1 {
		t.Errorf("Expected 1 room in area 2, got %d", len(roomsInArea2))
	}

	roomsInArea3 := m.GetRoomsInArea(3)
	if len(roomsInArea3) != 0 {
		t.Errorf("Expected 0 rooms in non-existent area 3, got %d", len(roomsInArea3))
	}
}

// TestGetLabelsForArea tests label retrieval for areas
func TestGetLabelsForArea(t *testing.T) {
	m := NewMudletMap()
	m.Areas[1] = NewMudletArea(1, "Area 1")

	// Add labels at map level (version < 21 style)
	label1 := &MudletLabel{ID: 1, Text: "Label 1"}
	label2 := &MudletLabel{ID: 2, Text: "Label 2"}
	m.Labels[1] = []*MudletLabel{label1, label2}

	labels := m.GetLabelsForArea(1)
	if len(labels) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(labels))
	}

	// Test version 21+ style (labels in area)
	m.Areas[1].Labels = []*MudletLabel{
		{ID: 3, Text: "Area Label"},
	}

	labels = m.GetLabelsForArea(1)
	if len(labels) != 1 {
		t.Errorf("Expected 1 label from area (v21+ takes precedence), got %d", len(labels))
	}
	if labels[0].Text != "Area Label" {
		t.Errorf("Expected 'Area Label', got %q", labels[0].Text)
	}

	// Non-existent area
	labels = m.GetLabelsForArea(999)
	if labels != nil {
		t.Errorf("Expected nil labels for non-existent area, got %v", labels)
	}
}

// TestParseSmallMap tests parsing the small 2-room map fixture
func TestParseSmallMap(t *testing.T) {
	if _, err := os.Stat(smallMapPath); os.IsNotExist(err) {
		t.Skipf("Test fixture not found: %s", smallMapPath)
	}

	m, err := ParseMapFile(smallMapPath)
	if err != nil {
		t.Fatalf("Failed to parse map: %v", err)
	}

	// Verify version
	if m.Version != 20 {
		t.Errorf("Expected version 20, got %d", m.Version)
	}

	// Verify areas
	if len(m.Areas) != 1 {
		t.Errorf("Expected 1 area, got %d", len(m.Areas))
	}

	// Verify area name
	area, ok := m.Areas[-1]
	if !ok {
		t.Error("Expected area with ID -1")
	} else if area.Name != "Default Area" {
		t.Errorf("Expected area name 'Default Area', got %q", area.Name)
	}

	// Verify rooms
	if len(m.Rooms) != 2 {
		t.Errorf("Expected 2 rooms, got %d", len(m.Rooms))
	}
}

// TestParseSmallMapRoomDetails tests detailed room parsing
func TestParseSmallMapRoomDetails(t *testing.T) {
	if _, err := os.Stat(smallMapPath); os.IsNotExist(err) {
		t.Skipf("Test fixture not found: %s", smallMapPath)
	}

	m, err := ParseMapFile(smallMapPath)
	if err != nil {
		t.Fatalf("Failed to parse map: %v", err)
	}

	// Verify room 1
	room1 := m.GetRoom(1)
	if room1 == nil {
		t.Fatal("Room 1 not found")
	}
	if room1.Name != "Przestronny korytarz." {
		t.Errorf("Expected room name 'Przestronny korytarz.', got %q", room1.Name)
	}
	if room1.Symbol != "K" {
		t.Errorf("Expected room symbol 'K', got %q", room1.Symbol)
	}
	if room1.X != 0 || room1.Y != 0 || room1.Z != 0 {
		t.Errorf("Expected room1 pos (0,0,0), got (%d,%d,%d)", room1.X, room1.Y, room1.Z)
	}
	if len(room1.SpecialExits) != 1 {
		t.Errorf("Expected 1 special exit, got %d", len(room1.SpecialExits))
	}

	// Verify room 2
	room2 := m.GetRoom(2)
	if room2 == nil {
		t.Fatal("Room 2 not found")
	}
	if room2.X != 0 || room2.Y != -1 || room2.Z != 0 {
		t.Errorf("Expected room2 pos (0,-1,0), got (%d,%d,%d)", room2.X, room2.Y, room2.Z)
	}

	// In this test fixture, rooms have no standard exits (all -1)
	// Room 1 has a special exit "rufa" to room 2
	for i, exit := range room1.Exits {
		if exit != NoExit {
			t.Errorf("Room1 exit %d should be NoExit (-1), got %d", i, exit)
		}
	}
	for i, exit := range room2.Exits {
		if exit != NoExit {
			t.Errorf("Room2 exit %d should be NoExit (-1), got %d", i, exit)
		}
	}

	// Verify special exit from room1 to room2
	if dest, ok := room1.SpecialExits["rufa"]; !ok {
		t.Error("Room1 should have special exit 'rufa'")
	} else if dest != 2 {
		t.Errorf("Room1 special exit 'rufa' should lead to room 2, got %d", dest)
	}
}

// TestParseLargeMap tests parsing the large map fixture
func TestParseLargeMap(t *testing.T) {
	if _, err := os.Stat(largeMapPath); os.IsNotExist(err) {
		t.Skipf("Test fixture not found: %s", largeMapPath)
	}

	m, err := ParseMapFile(largeMapPath)
	if err != nil {
		t.Fatalf("Failed to parse map: %v", err)
	}

	// Verify version
	if m.Version != 20 {
		t.Errorf("Expected version 20, got %d", m.Version)
	}

	// Verify areas count (61 from areaNames, but areas structure has 64)
	if len(m.Areas) < 60 {
		t.Errorf("Expected at least 60 areas, got %d", len(m.Areas))
	}

	// Verify rooms count
	if len(m.Rooms) != 26758 {
		t.Errorf("Expected 26758 rooms, got %d", len(m.Rooms))
	}

	// Verify user data
	if len(m.UserData) != 6 {
		t.Errorf("Expected 6 user data entries, got %d", len(m.UserData))
	}

	// Verify labels
	totalLabels := 0
	for _, labels := range m.Labels {
		totalLabels += len(labels)
	}
	if totalLabels != 397 {
		t.Errorf("Expected 397 labels, got %d", totalLabels)
	}
}

// TestParseMapFileError tests error handling for invalid file
func TestParseMapFileError(t *testing.T) {
	_, err := ParseMapFile("/nonexistent/path/to/file.dat")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

// TestValidateMap tests map validation
func TestValidateMap(t *testing.T) {
	// Test nil map
	errs := ValidateMap(nil)
	if len(errs) != 1 || errs[0].Type != "nil_map" {
		t.Error("Expected nil_map error for nil map")
	}

	// Test valid map
	m := NewMudletMap()
	m.Version = 20
	errs = ValidateMap(m)
	if len(errs) != 0 {
		t.Errorf("Expected no errors for valid empty map, got %d", len(errs))
	}

	// Test invalid version
	m.Version = 0
	errs = ValidateMap(m)
	if len(errs) != 1 || errs[0].Type != "invalid_version" {
		t.Error("Expected invalid_version error for version 0")
	}

	// Test broken exit
	m.Version = 20
	room := NewMudletRoom(1)
	room.Exits[ExitNorth] = 999 // points to non-existent room
	m.Rooms[1] = room

	errs = ValidateMap(m)
	if len(errs) != 1 || errs[0].Type != "broken_exit" {
		t.Error("Expected broken_exit error for exit to non-existent room")
	}

	// Add target room - should now be valid
	m.Rooms[999] = NewMudletRoom(999)
	errs = ValidateMap(m)
	if len(errs) != 0 {
		t.Errorf("Expected no errors after adding target room, got %d", len(errs))
	}
}

// TestGetMapStats tests statistics computation
func TestGetMapStats(t *testing.T) {
	m := NewMudletMap()

	// Empty map stats
	stats := GetMapStats(m)
	if stats.TotalRooms != 0 {
		t.Errorf("Expected 0 rooms, got %d", stats.TotalRooms)
	}

	// Add rooms at various positions
	for i := int32(1); i <= 5; i++ {
		room := NewMudletRoom(i)
		room.X = i * 10
		room.Y = i * 20
		room.Z = i % 3
		m.Rooms[i] = room
	}
	m.Areas[1] = NewMudletArea(1, "Test")

	stats = GetMapStats(m)

	if stats.TotalRooms != 5 {
		t.Errorf("Expected 5 rooms, got %d", stats.TotalRooms)
	}
	if stats.TotalAreas != 1 {
		t.Errorf("Expected 1 area, got %d", stats.TotalAreas)
	}

	// Check bounding box
	if stats.BoundingBox.MinX != 10 {
		t.Errorf("Expected MinX 10, got %d", stats.BoundingBox.MinX)
	}
	if stats.BoundingBox.MaxX != 50 {
		t.Errorf("Expected MaxX 50, got %d", stats.BoundingBox.MaxX)
	}
}

// BenchmarkParseSmallMap benchmarks parsing small map
func BenchmarkParseSmallMap(b *testing.B) {
	if _, err := os.Stat(smallMapPath); os.IsNotExist(err) {
		b.Skipf("Test fixture not found: %s", smallMapPath)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseMapFile(smallMapPath)
	}
}

// BenchmarkParseLargeMap benchmarks parsing large map
func BenchmarkParseLargeMap(b *testing.B) {
	if _, err := os.Stat(largeMapPath); os.IsNotExist(err) {
		b.Skipf("Test fixture not found: %s", largeMapPath)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseMapFile(largeMapPath)
	}
}
