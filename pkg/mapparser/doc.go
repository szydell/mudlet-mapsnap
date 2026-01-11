// Package mapparser provides functionality for parsing Mudlet map files.
//
// Mudlet is a popular MUD (Multi-User Dungeon) client, and it stores maps in
// a binary format using Qt's QDataStream serialization. This package parses
// that binary format and provides Go structures representing the map data.
//
// # Supported Format
//
// The parser supports Mudlet map format versions 6-20. The binary format uses
// big-endian byte order and Qt's QDataStream serialization conventions,
// including QString (UTF-16BE), QMap, QColor, and other Qt types.
//
// # Basic Usage
//
// Parse a map file:
//
//	m, err := mapparser.ParseMapFile("world.map")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Loaded %d rooms in %d areas\n", m.RoomCount(), m.AreaCount())
//
// Access rooms and areas:
//
//	room := m.GetRoom(1234)
//	if room != nil {
//	    fmt.Printf("Room: %s at (%d, %d, %d)\n", room.Name, room.X, room.Y, room.Z)
//	}
//
//	area := m.GetArea(1)
//	if area != nil {
//	    fmt.Printf("Area: %s\n", area.Name)
//	}
//
// # Map Structure
//
// The main types are:
//   - [MudletMap]: The root structure containing all map data
//   - [MudletArea]: An area/zone containing multiple rooms
//   - [MudletRoom]: A single room with exits, position, and metadata
//   - [MudletLabel]: A text or image label on the map
//
// # Validation and Export
//
// Validate map integrity:
//
//	errors := mapparser.ValidateMap(m)
//	for _, err := range errors {
//	    fmt.Printf("Error: %s\n", err.Message)
//	}
//
// Export to JSON:
//
//	err := mapparser.ExportToJSON(m, "output.json")
//
// # Room Exits
//
// Rooms have 12 standard exit directions, accessed via the Exits array:
//
//	north := room.Exits[mapparser.ExitNorth]
//	if north != mapparser.NoExit {
//	    fmt.Printf("North exit leads to room %d\n", north)
//	}
//
// Special exits (non-standard movement commands) are stored in the SpecialExits map.
package mapparser
