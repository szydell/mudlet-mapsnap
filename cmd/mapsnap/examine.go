package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/szydell/arkadia-mapsnap/pkg/mapparser"
)

// ExamineFile examines a binary map file and displays its Qt/MudletMap structure.
// With debug=false, shows compact summary. With debug=true, shows detailed values.
func ExamineFile(filename string, debug bool) error {
	info, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}
	fmt.Printf("File size: %d bytes\n\n", info.Size())

	m, err := mapparser.ParseMapFile(filename)
	if err != nil {
		return fmt.Errorf("parsing map: %w", err)
	}

	displayMap(m, debug)

	return nil
}

// displayMap outputs the parsed map structure
func displayMap(m *mapparser.MudletMap, debug bool) {
	// Version
	fmt.Printf("MudletMap.version:\n")
	fmt.Printf("  version = %d\n", m.Version)

	// EnvColors
	fmt.Printf("envColors QMap<int,int>:\n")
	fmt.Printf("  count = %d\n", len(m.EnvColors))

	// Areas (from areaNames)
	fmt.Printf("areaNames QMap<int,QString>:\n")
	fmt.Printf("  count = %d\n", len(m.Areas))
	if debug {
		ids := make([]int, 0, len(m.Areas))
		for id := range m.Areas {
			ids = append(ids, int(id))
		}
		sort.Ints(ids)
		for _, id := range ids {
			area := m.Areas[int32(id)]
			fmt.Printf("    id=%d name='%s'\n", id, area.Name)
		}
	}

	// CustomEnvColors
	fmt.Printf("mCustomEnvColors QMap<int,QColor>:\n")
	fmt.Printf("  count = %d\n", len(m.CustomEnvColors))

	// RoomDbHashToRoomId
	fmt.Printf("mpRoomDbHashToRoomId QMap<QString,uint>:\n")
	fmt.Printf("  count = %d\n", len(m.RoomDbHashToRoomId))

	// UserData
	fmt.Printf("mUserData QMap<QString,QString>:\n")
	fmt.Printf("  count = %d\n", len(m.UserData))

	// MapSymbolFont
	fmt.Printf("mapSymbolFont QFont:\n")
	fmt.Printf("  (parsed)\n")

	// MapFontFudgeFactor
	fmt.Printf("mapFontFudgeFactor:\n")
	fmt.Printf("  value = %f\n", m.MapFontFudgeFactor)

	// UseOnlyMapFont
	fmt.Printf("useOnlyMapFont:\n")
	fmt.Printf("  value = %v\n", m.UseOnlyMapFont)

	// Areas (full data)
	totalAreaRooms := 0
	for _, area := range m.Areas {
		totalAreaRooms += len(area.Rooms)
	}
	fmt.Printf("areas MudletAreas:\n")
	fmt.Printf("  count = %d areas, total rooms = %d\n", len(m.Areas), totalAreaRooms)
	if debug {
		ids := make([]int, 0, len(m.Areas))
		for id := range m.Areas {
			ids = append(ids, int(id))
		}
		sort.Ints(ids)
		for _, id := range ids {
			area := m.Areas[int32(id)]
			fmt.Printf("    area id=%d: rooms=%d, zLevels=%d, userData=%d\n",
				id, len(area.Rooms), len(area.ZLevels), len(area.UserData))
		}
	}

	// RoomIdHash
	fmt.Printf("mRoomIdHash QMap<QString,int>:\n")
	fmt.Printf("  count = %d\n", len(m.RoomIdHash))

	// Labels (version < 21)
	totalLabels := 0
	for _, labels := range m.Labels {
		totalLabels += len(labels)
	}
	fmt.Printf("labels MudletLabels:\n")
	fmt.Printf("  areas with labels = %d, total labels = %d\n", len(m.Labels), totalLabels)
	if debug {
		areaIDs := make([]int, 0, len(m.Labels))
		for areaID := range m.Labels {
			areaIDs = append(areaIDs, int(areaID))
		}
		sort.Ints(areaIDs)
		for _, areaID := range areaIDs {
			labels := m.Labels[int32(areaID)]
			fmt.Printf("    area id=%d: %d labels\n", areaID, len(labels))
			for j, lbl := range labels {
				fmt.Printf("      [%d] %s\n", j, formatLabel(lbl))
			}
		}
	}

	// Rooms
	fmt.Printf("rooms MudletRooms:\n")
	fmt.Printf("  total rooms = %d\n", len(m.Rooms))
	if debug && len(m.Rooms) > 0 {
		limit := 5
		count := 0
		fmt.Printf("  first %d rooms:\n", limit)
		ids := make([]int, 0, len(m.Rooms))
		for id := range m.Rooms {
			ids = append(ids, int(id))
		}
		sort.Ints(ids)
		for _, id := range ids {
			if count >= limit {
				break
			}
			room := m.Rooms[int32(id)]
			fmt.Printf("    [%d] %s\n", count, formatRoom(room))
			count++
		}
		if len(m.Rooms) > limit {
			fmt.Printf("    ... and %d more rooms\n", len(m.Rooms)-limit)
		}
	}

	fmt.Println()
}

// formatRoom returns a compact string representation of a room
func formatRoom(room *mapparser.MudletRoom) string {
	exitNames := []string{"n", "ne", "e", "se", "s", "sw", "w", "nw", "up", "down", "in", "out"}
	var exits []string
	for i, dest := range room.Exits {
		if dest != -1 {
			exits = append(exits, fmt.Sprintf("%s:%d", exitNames[i], dest))
		}
	}
	for cmd, dest := range room.SpecialExits {
		exits = append(exits, fmt.Sprintf("spec(%s):%d", cmd, dest))
	}

	exitsStr := "none"
	if len(exits) > 0 {
		exitsStr = strings.Join(exits, ",")
	}

	return fmt.Sprintf("id=%d area=%d pos=(%d,%d,%d) exits=[%s] name='%s' env=%d",
		room.ID, room.Area, room.X, room.Y, room.Z, exitsStr, room.Name, room.Environment)
}

// formatLabel returns a compact string representation of a label
func formatLabel(lbl *mapparser.MudletLabel) string {
	text := lbl.Text
	if len(text) > 30 {
		text = text[:30] + "..."
	}
	pixBytes := 0
	if lbl.Pixmap != nil {
		pixBytes = len(lbl.Pixmap)
	}
	return fmt.Sprintf("id=%d pos=(%.1f,%.1f,%.1f) size=(%.1f,%.1f) text='%s' pix=%d bytes noScale=%v onTop=%v",
		lbl.ID, lbl.Pos.X, lbl.Pos.Y, lbl.Pos.Z, lbl.Width, lbl.Height, text, pixBytes, lbl.NoScaling, lbl.ShowOnTop)
}
