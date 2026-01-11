package mapparser

// ============================================================================
// Type Aliases (Backward Compatibility)
// ============================================================================

// Map is a type alias for [MudletMap].
//
// Deprecated: Use [MudletMap] directly in new code.
type Map = MudletMap

// Room is a type alias for [MudletRoom].
//
// Deprecated: Use [MudletRoom] directly in new code.
type Room = MudletRoom

// Area is a type alias for [MudletArea].
//
// Deprecated: Use [MudletArea] directly in new code.
type Area = MudletArea

// Label is a type alias for [MudletLabel].
//
// Deprecated: Use [MudletLabel] directly in new code.
type Label = MudletLabel

// ============================================================================
// Validation and Statistics Types
// ============================================================================

// ValidationError represents an error found during map validation.
type ValidationError struct {
	// Type categorizes the error (e.g., "broken_exit", "invalid_version").
	Type string `json:"type"`
	// Message provides a human-readable description of the error.
	Message string `json:"message"`
	// RoomID identifies the room where the error occurred (if applicable).
	RoomID int32 `json:"roomId,omitempty"`
}

// MapStats contains aggregate statistics about a map.
type MapStats struct {
	// TotalRooms is the number of rooms in the map.
	TotalRooms int `json:"totalRooms"`
	// TotalAreas is the number of areas in the map.
	TotalAreas int `json:"totalAreas"`
	// TotalEnvironments is the count of unique environment types.
	TotalEnvironments int `json:"totalEnvironments"`
	// BoundingBox defines the spatial extent of all rooms.
	BoundingBox BoundingBox `json:"boundingBox"`
	// ZLevels is a sorted list of all Z-coordinates used.
	ZLevels []int32 `json:"zLevels"`
}

// BoundingBox represents the minimum and maximum coordinates of the map.
type BoundingBox struct {
	MinX int32 `json:"minX"`
	MinY int32 `json:"minY"`
	MinZ int32 `json:"minZ"`
	MaxX int32 `json:"maxX"`
	MaxY int32 `json:"maxY"`
	MaxZ int32 `json:"maxZ"`
}
