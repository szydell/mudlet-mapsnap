package mapparser

// Map represents the entire map structure from a Mudlet map file
type Map struct {
	Header       Header                `json:"header"`
	Rooms        map[int32]*Room       `json:"rooms"`
	Areas        map[int32]*Area       `json:"areas"`
	Environments []Environment         `json:"environments"`
	CustomLines  []CustomLine          `json:"customLines,omitempty"`
	Labels       []Label               `json:"labels,omitempty"`
}

// Header contains the map file header information
type Header struct {
	Magic   string `json:"magic"`   // "ATADNOOM"
	Version int8   `json:"version"` // 1, 2, or 3
}

// Room represents a single room in the map
type Room struct {
	ID          int32  `json:"id"`
	X           int32  `json:"x"`
	Y           int32  `json:"y"`
	Z           int32  `json:"z"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Environment int32  `json:"environment"`
	Exits       []Exit `json:"exits"`
}

// Exit represents a connection between rooms
type Exit struct {
	Direction string `json:"direction"`  // "north", "south", etc.
	TargetID  int32  `json:"targetId"`   // ID of the target room
	Lock      bool   `json:"lock"`       // locked exit (v3+)
	Weight    int32  `json:"weight"`     // path weight (v3+)
}

// Area represents a map area
type Area struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

// Environment represents a room environment type
type Environment struct {
	Name  string `json:"name"`    // "forest", "city", etc.
	Color int32  `json:"color"`   // RGB color as int32
}

// CustomLine represents a custom line drawn on the map
type CustomLine struct {
	X1, Y1, Z1 int32 `json:"x1,y1,z1"`
	X2, Y2, Z2 int32 `json:"x2,y2,z2"`
	Color      int32 `json:"color"`
	Width      int8  `json:"width"`
	Style      int8  `json:"style"`
}

// Label represents a text label on the map
type Label struct {
	X, Y, Z        int32  `json:"x,y,z"`
	Text           string `json:"text"`
	Color          int32  `json:"color"`
	Size           int8   `json:"size"`
	ShowBackground bool   `json:"showBackground"`
}

// ValidationError represents an error found during map validation
type ValidationError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	RoomID  int32  `json:"roomId,omitempty"`
}

// MapStats contains statistics about the map
type MapStats struct {
	TotalRooms       int         `json:"totalRooms"`
	TotalAreas       int         `json:"totalAreas"`
	TotalEnvironments int        `json:"totalEnvironments"`
	BoundingBox      BoundingBox `json:"boundingBox"`
	ZLevels          []int32     `json:"zLevels"`
}

// BoundingBox represents the minimum and maximum coordinates of the map
type BoundingBox struct {
	MinX, MinY, MinZ int32 `json:"minX,minY,minZ"`
	MaxX, MaxY, MaxZ int32 `json:"maxX,maxY,maxZ"`
}
