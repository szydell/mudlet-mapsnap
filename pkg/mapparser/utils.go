package mapparser

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// ValidateMap performs a minimal validation of the parsed map structure.
// This is a stub implementation intended to unblock compilation and basic CLI flows.
// It checks for a valid version and that exit targets (if not NoExit) refer to existing rooms.
func ValidateMap(m *Map) []ValidationError {
	var errs []ValidationError
	if m == nil {
		errs = append(errs, ValidationError{Type: "nil_map", Message: "map is nil"})
		return errs
	}
	// Mudlet QDataStream version is typically >= 6; just ensure positive
	if m.Version <= 0 {
		errs = append(errs, ValidationError{Type: "invalid_version", Message: fmt.Sprintf("non-positive version: %d", m.Version)})
	}
	// Check that exits point to existing rooms when not NoExit
	for _, room := range m.Rooms {
		for i, exitTarget := range room.Exits {
			if exitTarget != NoExit {
				if _, ok := m.Rooms[exitTarget]; !ok {
					errs = append(errs, ValidationError{
						Type:    "broken_exit",
						Message: fmt.Sprintf("room %d has %s exit to missing room %d", room.ID, ExitDirectionNames[i], exitTarget),
						RoomID:  room.ID,
					})
				}
			}
		}
	}
	return errs
}

// GetMapStats returns basic statistics computed from the map structure.
func GetMapStats(m *Map) MapStats {
	stats := MapStats{}
	if m == nil {
		return stats
	}
	stats.TotalRooms = len(m.Rooms)
	stats.TotalAreas = len(m.Areas)
	stats.TotalEnvironments = len(m.EnvColors) + len(m.CustomEnvColors)
	if len(m.Rooms) == 0 {
		return stats
	}
	// Compute bounding box and Z levels
	first := true
	zset := make(map[int32]struct{})
	for _, r := range m.Rooms {
		if first {
			stats.BoundingBox.MinX, stats.BoundingBox.MaxX = r.X, r.X
			stats.BoundingBox.MinY, stats.BoundingBox.MaxY = r.Y, r.Y
			stats.BoundingBox.MinZ, stats.BoundingBox.MaxZ = r.Z, r.Z
			first = false
		} else {
			if r.X < stats.BoundingBox.MinX {
				stats.BoundingBox.MinX = r.X
			}
			if r.X > stats.BoundingBox.MaxX {
				stats.BoundingBox.MaxX = r.X
			}
			if r.Y < stats.BoundingBox.MinY {
				stats.BoundingBox.MinY = r.Y
			}
			if r.Y > stats.BoundingBox.MaxY {
				stats.BoundingBox.MaxY = r.Y
			}
			if r.Z < stats.BoundingBox.MinZ {
				stats.BoundingBox.MinZ = r.Z
			}
			if r.Z > stats.BoundingBox.MaxZ {
				stats.BoundingBox.MaxZ = r.Z
			}
		}
		zset[r.Z] = struct{}{}
	}
	// Sorted z-levels
	for z := range zset {
		stats.ZLevels = append(stats.ZLevels, z)
	}
	sort.Slice(stats.ZLevels, func(i, j int) bool { return stats.ZLevels[i] < stats.ZLevels[j] })
	return stats
}

// ExportToJSON writes the map structure to a JSON file with indentation.
func ExportToJSON(m *Map, filename string) error {
	if m == nil {
		return fmt.Errorf("nil map provided")
	}
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("creating json file: %w", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(m); err != nil {
		return fmt.Errorf("encoding json: %w", err)
	}
	return nil
}
