package mapparser

// ============================================================================
// Legacy type aliases for backward compatibility
// New code should use MudletMap, MudletRoom, MudletArea, MudletLabel directly
// ============================================================================

// Map is an alias for MudletMap (for backward compatibility)
// Deprecated: Use MudletMap instead
type Map = MudletMap

// Room is an alias for MudletRoom (for backward compatibility)
// Deprecated: Use MudletRoom instead
type Room = MudletRoom

// Area is an alias for MudletArea (for backward compatibility)
// Deprecated: Use MudletArea instead
type Area = MudletArea

// Label is an alias for MudletLabel (for backward compatibility)
// Deprecated: Use MudletLabel instead
type Label = MudletLabel

// ValidationError represents an error found during map validation
type ValidationError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	RoomID  int32  `json:"roomId,omitempty"`
}

// MapStats contains statistics about the map
type MapStats struct {
	TotalRooms        int         `json:"totalRooms"`
	TotalAreas        int         `json:"totalAreas"`
	TotalEnvironments int         `json:"totalEnvironments"`
	BoundingBox       BoundingBox `json:"boundingBox"`
	ZLevels           []int32     `json:"zLevels"`
}

// BoundingBox represents the minimum and maximum coordinates of the map
type BoundingBox struct {
	MinX, MinY, MinZ int32 `json:"minX,minY,minZ"`
	MaxX, MaxY, MaxZ int32 `json:"maxX,maxY,maxZ"`
}
