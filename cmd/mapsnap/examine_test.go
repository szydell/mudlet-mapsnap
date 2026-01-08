package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/szydell/arkadia-mapsnap/pkg/mapparser"
)

// Test fixtures paths
const (
	smallMapPath = "../../tests/fixtures/2_rooms_map/2lok.dat"
	largeMapPath = "../../tests/fixtures/large_maps/2025-05-27#15-06-15map.dat"
)

// TestExamineSmallMap tests examine output on a small 2-room map
func TestExamineSmallMap(t *testing.T) {
	if _, err := os.Stat(smallMapPath); os.IsNotExist(err) {
		t.Skipf("Test fixture not found: %s", smallMapPath)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := ExamineFile(smallMapPath, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if err != nil {
		t.Fatalf("ExamineFile failed: %v", err)
	}

	expectedStrings := []string{
		"version = 20",
		"areaNames QMap<int,QString>:",
		"count = 1",
		"areas MudletAreas:",
		"total rooms = 2",
		"labels MudletLabels:",
		"rooms MudletRooms:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", expected, output)
		}
	}
}

// TestExamineSmallMapDebug tests examine with debug mode
func TestExamineSmallMapDebug(t *testing.T) {
	if _, err := os.Stat(smallMapPath); os.IsNotExist(err) {
		t.Skipf("Test fixture not found: %s", smallMapPath)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := ExamineFile(smallMapPath, true)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if err != nil {
		t.Fatalf("ExamineFile failed: %v", err)
	}

	// Debug mode shows area names and room details
	expectedStrings := []string{
		"id=-1 name='Default Area'",
		"area=-1 pos=(0,-1,0)",
		"area=-1 pos=(0,0,0)",
		"name='Przestronny korytarz.'",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected debug output to contain %q, but it didn't.\nOutput:\n%s", expected, output)
		}
	}
}

// TestExamineLargeMap tests examine on a large map
func TestExamineLargeMap(t *testing.T) {
	if _, err := os.Stat(largeMapPath); os.IsNotExist(err) {
		t.Skipf("Test fixture not found: %s", largeMapPath)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := ExamineFile(largeMapPath, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if err != nil {
		t.Fatalf("ExamineFile failed: %v", err)
	}

	expectedStrings := []string{
		"version = 20",
		"count = 64",                            // areaNames
		"count = 64 areas, total rooms = 26758", // areas summary
		"areas with labels = 51, total labels = 397",
		"total rooms = 26758",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", expected, output)
		}
	}
}

// TestExamineNonExistentFile tests error handling for missing file
func TestExamineNonExistentFile(t *testing.T) {
	err := ExamineFile("/nonexistent/path/to/file.dat", false)
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

// TestFormatRoom tests the formatRoom helper function
func TestFormatRoom(t *testing.T) {
	room := &mapparser.MudletRoom{
		ID:          1,
		Area:        5,
		X:           10,
		Y:           20,
		Z:           0,
		Exits:       [12]int32{2, -1, 3, -1, 4, -1, -1, -1, -1, -1, -1, -1},
		Name:        "Test Room",
		Environment: 100,
	}

	output := formatRoom(room)

	expectedParts := []string{
		"id=1",
		"area=5",
		"pos=(10,20,0)",
		"n:2",
		"e:3",
		"s:4",
		"name='Test Room'",
		"env=100",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("Expected formatRoom output to contain %q, got: %s", part, output)
		}
	}
}

// TestFormatRoomNoExits tests formatRoom with no exits
func TestFormatRoomNoExits(t *testing.T) {
	room := mapparser.NewMudletRoom(42)
	room.Name = "Isolated Room"

	output := formatRoom(room)

	if !strings.Contains(output, "exits=[none]") {
		t.Errorf("Expected 'exits=[none]' for room with no exits, got: %s", output)
	}
}

// TestFormatRoomSpecialExits tests formatRoom with special exits
func TestFormatRoomSpecialExits(t *testing.T) {
	room := mapparser.NewMudletRoom(1)
	room.SpecialExits["wejdz do portalu"] = 100

	output := formatRoom(room)

	if !strings.Contains(output, "spec(wejdz do portalu):100") {
		t.Errorf("Expected special exit in output, got: %s", output)
	}
}

// TestFormatLabel tests the formatLabel helper function
func TestFormatLabel(t *testing.T) {
	label := &mapparser.MudletLabel{
		ID:        1,
		Pos:       mapparser.Vector3D{X: 10.5, Y: 20.5, Z: 0},
		Width:     100,
		Height:    50,
		Text:      "Test Label",
		NoScaling: true,
		ShowOnTop: false,
	}

	output := formatLabel(label)

	expectedParts := []string{
		"id=1",
		"pos=(10.5,20.5,0.0)",
		"size=(100.0,50.0)",
		"text='Test Label'",
		"noScale=true",
		"onTop=false",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("Expected formatLabel output to contain %q, got: %s", part, output)
		}
	}
}

// TestFormatLabelLongText tests formatLabel truncates long text
func TestFormatLabelLongText(t *testing.T) {
	label := &mapparser.MudletLabel{
		ID:   1,
		Text: "This is a very long label text that should be truncated for display",
	}

	output := formatLabel(label)

	if !strings.Contains(output, "...") {
		t.Errorf("Expected long text to be truncated with '...', got: %s", output)
	}
	if strings.Contains(output, "truncated for display") {
		t.Errorf("Expected text to be truncated, but full text present: %s", output)
	}
}

// --- Benchmarks ---

// BenchmarkExamineLargeMap benchmarks ExamineFile display
func BenchmarkExamineLargeMap(b *testing.B) {
	if _, err := os.Stat(largeMapPath); os.IsNotExist(err) {
		b.Skipf("Test fixture not found: %s", largeMapPath)
	}

	oldStdout := os.Stdout
	devNull, _ := os.Open(os.DevNull)
	defer devNull.Close()
	os.Stdout = devNull

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExamineFile(largeMapPath, false)
	}

	os.Stdout = oldStdout
}
