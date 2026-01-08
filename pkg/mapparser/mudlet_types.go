package mapparser

// MudletMap represents the complete structure of a Mudlet map file (version 6-21+)
// This is the primary data structure used throughout the application.
type MudletMap struct {
	Version int32 `json:"version"`

	// Environment colors: maps environment ID to color value
	EnvColors map[int32]int32 `json:"envColors,omitempty"`

	// Custom environment colors: maps environment ID to RGBA color
	CustomEnvColors map[int32]Color `json:"customEnvColors,omitempty"`

	// Room hash to ID mapping (for quick lookup by hash)
	RoomDbHashToRoomId map[string]uint32 `json:"roomDbHashToRoomId,omitempty"`

	// Room ID hash (reverse lookup)
	RoomIdHash map[string]int32 `json:"roomIdHash,omitempty"`

	// User-defined metadata for the map
	UserData map[string]string `json:"userData,omitempty"`

	// Map symbol font settings
	MapSymbolFont      Font    `json:"mapSymbolFont,omitempty"`
	MapFontFudgeFactor float64 `json:"mapFontFudgeFactor"`
	UseOnlyMapFont     bool    `json:"useOnlyMapFont"`

	// Areas indexed by area ID
	Areas map[int32]*MudletArea `json:"areas"`

	// Rooms indexed by room ID
	Rooms map[int32]*MudletRoom `json:"rooms"`

	// Labels organized by area ID (version < 21)
	// In version 21+, labels are stored inside each area
	Labels map[int32][]*MudletLabel `json:"labels,omitempty"`
}

// MudletArea represents a map area containing rooms
type MudletArea struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`

	// Room IDs belonging to this area
	Rooms []uint32 `json:"rooms,omitempty"`

	// Z-levels used in this area
	ZLevels []int32 `json:"zLevels,omitempty"`

	// Area exits: rooms that connect to other areas
	// Key is the room ID in this area, value contains destination room and direction
	AreaExits []AreaExit `json:"areaExits,omitempty"`

	// Grid display mode
	GridMode bool `json:"gridMode"`

	// Bounding box
	Bounds BoundingBox3D `json:"bounds"`

	// Span vector
	Span Vector3D `json:"span"`

	// Per-Z-level bounds
	XMaxForZ map[int32]int32 `json:"xMaxForZ,omitempty"`
	YMaxForZ map[int32]int32 `json:"yMaxForZ,omitempty"`
	XMinForZ map[int32]int32 `json:"xMinForZ,omitempty"`
	YMinForZ map[int32]int32 `json:"yMinForZ,omitempty"`

	// Position vector
	Pos Vector3D `json:"pos"`

	// Zone settings
	IsZone      bool  `json:"isZone"`
	ZoneAreaRef int32 `json:"zoneAreaRef"`

	// Last 2D map zoom level (version >= 21)
	Last2DMapZoom float64 `json:"last2DMapZoom,omitempty"`

	// User-defined metadata
	UserData map[string]string `json:"userData,omitempty"`

	// Labels (version >= 21 stores labels inside area)
	Labels []*MudletLabel `json:"labels,omitempty"`
}

// AreaExit represents an exit from this area to another area
type AreaExit struct {
	RoomID     int32 `json:"roomId"`     // Room ID in this area
	DestRoomID int32 `json:"destRoomId"` // Room ID in other area
	Direction  int32 `json:"direction"`  // Exit direction
}

// MudletRoom represents a single room in the map
type MudletRoom struct {
	ID   int32 `json:"id"`
	Area int32 `json:"area"`

	// Position on map grid
	X int32 `json:"x"`
	Y int32 `json:"y"`
	Z int32 `json:"z"`

	// Standard exits (12 directions): -1 means no exit
	// Index: 0=north, 1=northeast, 2=east, 3=southeast, 4=south, 5=southwest,
	//        6=west, 7=northwest, 8=up, 9=down, 10=in, 11=out
	Exits [12]int32 `json:"exits"`

	// Environment type (for coloring)
	Environment int32 `json:"environment"`

	// Pathfinding weight (minimum 1)
	Weight int32 `json:"weight"`

	// Room name/label
	Name string `json:"name"`

	// Whether the room is locked for pathfinding
	IsLocked bool `json:"isLocked"`

	// Special exits: custom movement commands
	// Version 6-20: key=destination room ID, value=command (with "0"/"1" lock prefix)
	// Version 21+: key=command, value=destination room ID
	SpecialExits map[string]int32 `json:"specialExits,omitempty"`

	// Map symbol displayed on the room (version >= 19)
	Symbol string `json:"symbol,omitempty"`

	// Symbol color (version >= 21)
	SymbolColor *Color `json:"symbolColor,omitempty"`

	// User-defined metadata (version >= 10)
	UserData map[string]string `json:"userData,omitempty"`

	// Custom lines drawn from this room (version >= 11)
	CustomLines      map[string][]Point2D `json:"customLines,omitempty"`
	CustomLinesArrow map[string]bool      `json:"customLinesArrow,omitempty"`
	CustomLinesColor map[string]Color     `json:"customLinesColor,omitempty"`
	CustomLinesStyle map[string]int32     `json:"customLinesStyle,omitempty"`

	// Special exit locks (version >= 21)
	SpecialExitLocks []string `json:"specialExitLocks,omitempty"`

	// Exit locks: list of locked standard exit directions (version >= 11)
	ExitLocks []int32 `json:"exitLocks,omitempty"`

	// Exit stubs: directions with stub exits (version >= 13)
	ExitStubs []int32 `json:"exitStubs,omitempty"`

	// Exit weights: custom weights per direction (version >= 16)
	ExitWeights map[string]int32 `json:"exitWeights,omitempty"`

	// Doors: door type per direction (version >= 16)
	// 0=none, 1=open, 2=closed, 3=locked
	Doors map[string]int32 `json:"doors,omitempty"`
}

// MudletLabel represents a text or image label on the map
type MudletLabel struct {
	ID int32 `json:"id"`

	// Position (version >= 12 uses 3D, earlier used 2D)
	Pos Vector3D `json:"pos"`

	// Size
	Width  float64 `json:"width"`
	Height float64 `json:"height"`

	// Label text
	Text string `json:"text,omitempty"`

	// Colors
	FgColor Color `json:"fgColor"`
	BgColor Color `json:"bgColor"`

	// Image data (PNG bytes)
	Pixmap []byte `json:"pixmap,omitempty"`

	// Display flags (version >= 15)
	NoScaling bool `json:"noScaling"`
	ShowOnTop bool `json:"showOnTop"`
}

// Color represents an RGBA color (Qt QColor)
type Color struct {
	Spec  int8   `json:"spec"` // Color specification type
	Red   uint16 `json:"r"`
	Green uint16 `json:"g"`
	Blue  uint16 `json:"b"`
	Alpha uint16 `json:"a"`
	Pad   uint16 `json:"-"` // Padding field in QColor
}

// ToRGBA returns the color as 8-bit RGBA values
func (c Color) ToRGBA() (r, g, b, a uint8) {
	return uint8(c.Red >> 8), uint8(c.Green >> 8), uint8(c.Blue >> 8), uint8(c.Alpha >> 8)
}

// Font represents Qt QFont structure
type Font struct {
	Family            string  `json:"family"`
	StyleHint         string  `json:"styleHint,omitempty"`
	PointSizeF        float64 `json:"pointSizeF"`
	PixelSize         int32   `json:"pixelSize"`
	StyleStrategy     int8    `json:"styleStrategy"`
	Weight            uint16  `json:"weight"`
	Style             uint8   `json:"style"`
	Underline         bool    `json:"underline"`
	StrikeOut         bool    `json:"strikeOut"`
	FixedPitch        bool    `json:"fixedPitch"`
	Capitalization    int8    `json:"capitalization"`
	LetterSpacing     int32   `json:"letterSpacing"`
	WordSpacing       int32   `json:"wordSpacing"`
	Stretch           int8    `json:"stretch"`
	HintingPreference int8    `json:"hintingPreference"`
}

// Vector3D represents a 3D vector (Qt QVector3D stored as 3 doubles)
type Vector3D struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// Point2D represents a 2D point (Qt QPointF)
type Point2D struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// BoundingBox3D represents 3D bounding box
type BoundingBox3D struct {
	MinX int32 `json:"minX"`
	MinY int32 `json:"minY"`
	MinZ int32 `json:"minZ"`
	MaxX int32 `json:"maxX"`
	MaxY int32 `json:"maxY"`
	MaxZ int32 `json:"maxZ"`
}

// ExitDirection constants for standard exits
const (
	ExitNorth     = 0
	ExitNortheast = 1
	ExitEast      = 2
	ExitSoutheast = 3
	ExitSouth     = 4
	ExitSouthwest = 5
	ExitWest      = 6
	ExitNorthwest = 7
	ExitUp        = 8
	ExitDown      = 9
	ExitIn        = 10
	ExitOut       = 11
)

// ExitDirectionNames maps exit index to direction name
var ExitDirectionNames = [12]string{
	"north", "northeast", "east", "southeast",
	"south", "southwest", "west", "northwest",
	"up", "down", "in", "out",
}

// ExitDirectionShortNames maps exit index to short direction name
var ExitDirectionShortNames = [12]string{
	"n", "ne", "e", "se", "s", "sw", "w", "nw", "up", "down", "in", "out",
}

// NoExit is the value indicating no exit in that direction
const NoExit int32 = -1

// DoorType constants
const (
	DoorNone   = 0
	DoorOpen   = 1
	DoorClosed = 2
	DoorLocked = 3
)

// NewMudletMap creates a new empty MudletMap
func NewMudletMap() *MudletMap {
	return &MudletMap{
		EnvColors:          make(map[int32]int32),
		CustomEnvColors:    make(map[int32]Color),
		RoomDbHashToRoomId: make(map[string]uint32),
		RoomIdHash:         make(map[string]int32),
		UserData:           make(map[string]string),
		Areas:              make(map[int32]*MudletArea),
		Rooms:              make(map[int32]*MudletRoom),
		Labels:             make(map[int32][]*MudletLabel),
	}
}

// NewMudletArea creates a new empty MudletArea
func NewMudletArea(id int32, name string) *MudletArea {
	return &MudletArea{
		ID:       id,
		Name:     name,
		Rooms:    make([]uint32, 0),
		ZLevels:  make([]int32, 0),
		XMaxForZ: make(map[int32]int32),
		YMaxForZ: make(map[int32]int32),
		XMinForZ: make(map[int32]int32),
		YMinForZ: make(map[int32]int32),
		UserData: make(map[string]string),
		Labels:   make([]*MudletLabel, 0),
	}
}

// NewMudletRoom creates a new MudletRoom with default values
func NewMudletRoom(id int32) *MudletRoom {
	r := &MudletRoom{
		ID:               id,
		Weight:           1,
		SpecialExits:     make(map[string]int32),
		UserData:         make(map[string]string),
		CustomLines:      make(map[string][]Point2D),
		CustomLinesArrow: make(map[string]bool),
		CustomLinesColor: make(map[string]Color),
		CustomLinesStyle: make(map[string]int32),
		ExitWeights:      make(map[string]int32),
		Doors:            make(map[string]int32),
	}
	// Initialize all exits to NoExit
	for i := range r.Exits {
		r.Exits[i] = NoExit
	}
	return r
}

// GetExit returns the destination room ID for a given direction, or NoExit
func (r *MudletRoom) GetExit(direction int) int32 {
	if direction < 0 || direction >= 12 {
		return NoExit
	}
	return r.Exits[direction]
}

// HasExit checks if the room has an exit in the given direction
func (r *MudletRoom) HasExit(direction int) bool {
	return r.GetExit(direction) != NoExit
}

// ActiveExits returns a slice of directions that have exits
func (r *MudletRoom) ActiveExits() []int {
	var result []int
	for i, exit := range r.Exits {
		if exit != NoExit {
			result = append(result, i)
		}
	}
	return result
}

// GetRoom returns a room by ID, or nil if not found
func (m *MudletMap) GetRoom(id int32) *MudletRoom {
	return m.Rooms[id]
}

// GetArea returns an area by ID, or nil if not found
func (m *MudletMap) GetArea(id int32) *MudletArea {
	return m.Areas[id]
}

// RoomCount returns the total number of rooms
func (m *MudletMap) RoomCount() int {
	return len(m.Rooms)
}

// AreaCount returns the total number of areas
func (m *MudletMap) AreaCount() int {
	return len(m.Areas)
}

// GetRoomsInArea returns all rooms belonging to an area
func (m *MudletMap) GetRoomsInArea(areaID int32) []*MudletRoom {
	var rooms []*MudletRoom
	for _, room := range m.Rooms {
		if room.Area == areaID {
			rooms = append(rooms, room)
		}
	}
	return rooms
}

// GetLabelsForArea returns labels for a specific area
func (m *MudletMap) GetLabelsForArea(areaID int32) []*MudletLabel {
	// First check if labels are stored in the area itself (version 21+)
	if area, ok := m.Areas[areaID]; ok && len(area.Labels) > 0 {
		return area.Labels
	}
	// Fall back to map-level labels (version < 21)
	return m.Labels[areaID]
}
