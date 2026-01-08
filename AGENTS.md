# AI Assistant instructions for mudlet-mapsnap

## Project overview

**mudlet-mapsnap** is a Go library and CLI tool for parsing and visualizing Mudlet map files. The project enables generating visual map fragments centered on selected locations.

### Problem solved
- Mudlet is a popular MUD client used by players of various text-based games
- Mudlet stores maps in a binary format (QDataStream) that's difficult to process
- Need for quick generation of visual map fragments for navigation and sharing

### Target functionality
1. **Library (`pkg/`)**: Parse Mudlet map files, search rooms, render images
2. **CLI (`cmd/mapsnap`)**: Command-line tool for map operations

## Core technologies

- **Go 1.25+**
- **QDataStream binary format** (Qt serialization)
- **Image output**: WEBP (lossless, via `nativewebp`) and PNG
- **Pure Go**: No CGO required, fully static binary

## Project structure

```
mudlet-mapsnap/
├── cmd/mapsnap/           # CLI application
│   ├── main.go           # Entry point and flags
│   └── examine.go        # Binary examination with Qt/MudletMap parsing
├── pkg/
│   ├── mapparser/        # Map file parsing
│   │   ├── parser.go     # Main parser
│   │   ├── types.go      # Data structures
│   │   ├── reader.go     # Binary reading helpers
│   │   └── utils.go      # Utilities
│   ├── maprenderer/      # Image generation (WIP)
│   └── maputils/         # Common utilities
├── docs/
│   └── sources/          # Reference implementations
│       ├── Mudlet/       # Mudlet C++ source excerpts
│       └── node-mudlet-map-binary-reader/  # Node.js parser reference
├── tests/fixtures/       # Test data
├── go.mod
├── Makefile
└── README.md
```

## Binary format reference

### Mudlet map format (version 20)
The map file uses Qt's QDataStream serialization (big-endian).

**Key structures:**
1. **MudletMap** - version → envColors → areaNames → customEnvColors → mpRoomDbHashToRoomId → mUserData → mapSymbolFont → areas → mRoomIdHash → labels → rooms
2. **QString** - quint32 length (BYTES, not chars) + UTF-16BE data. 0xFFFFFFFF = null string
3. **QMap<K,V>** - qint32 count + key-value pairs
4. **MudletRoom** - 12 standard exits + special exits, environment, weight, name, userData (see detailed structure below)
5. **MudletLabel** - id, pos(3×double), dummy(2×double), size(2×double), text, colors, pixmap, flags
6. **MudletArea** - rooms, zLevels, mAreaExits, gridMode, bounds, span, grid maps, pos, isZone, zoneAreaRef, userData

### MudletArea structure (version 20)
```
MudletArea:
  QSet<uint32> rooms           # Room IDs in this area (QList<int> in v<18)
  QList<int>   zLevels         # Z-levels used in this area
  QMultiMap<int, QPair<int,int>> mAreaExits  # Area border exits
                               # key=in_area room, pair=(out_area room, direction)
  bool         gridMode        # Grid display mode
  int32        max_x, max_y, max_z, min_x, min_y, min_z  # Bounding box
  QVector3D    span            # 3 x double
  QMap<int,int> xmaxForZ, ymaxForZ, xminForZ, yminForZ  # Per-Z bounds
  QVector3D    pos             # Area position (3 x double)
  bool         isZone          # Is this area a zone?
  int32        zoneAreaRef     # Reference to zone area
  # double     mLast2DMapZoom  # Only version >= 21
  QMap<QString,QString> mUserData
  # mMapLabels                 # Only version >= 21 (labels inside area)
```

### MudletRoom structure (version 20)
```
MudletRoom:
  int32   area           # ID of parent area
  int32   x, y, z        # Position on map grid
  int32   north          # -1 = no exit, otherwise destination room ID
  int32   northeast
  int32   east
  int32   southeast
  int32   south
  int32   southwest
  int32   west
  int32   northwest
  int32   up
  int32   down
  int32   in
  int32   out
  int32   environment    # Environment type (for coloring)
  int32   weight         # Pathfinding weight (min 1)
  QString name           # Room name/label
  bool    isLocked       # Whether room is locked
  
  # Special exits (version 6-20): QMultiMap<int, QString>
  # Key = destination room ID, Value = command with "0"/"1" lock prefix
  # Version 21+: QMultiMap<QString, int> (reversed)
  
  QString symbol         # Map symbol (version >= 19)
  # QColor symbolColor   # Only in version >= 21
  
  QMap<QString, QString> userData    # version >= 10
  
  # Custom lines (version >= 11, format differs v20 vs older):
  QMap<QString, QList<QPointF>> customLines
  QMap<QString, bool> customLinesArrow
  QMap<QString, QColor> customLinesColor     # v20+: QColor, older: QList<int>
  QMap<QString, int> customLinesStyle        # v20+: int, older: QString
  
  # QSet<QString> mSpecialExitLocks  # Only version >= 21
  QList<int> exitLocks               # version >= 11
  QList<int> exitStubs               # version >= 13
  QMap<QString, int> exitWeights     # version >= 16
  QMap<QString, int> doors           # version >= 16
```

### MudletLabel structure (version 11-20)
```
MudletLabel:
  int32     labelID
  QVector3D pos            # 3 x double (v12+, earlier: QPointF = 2 x double)
  QPointF   dummy          # 2 x double (unused, removed in v21)
  QSizeF    size           # 2 x double (width, height)
  QString   text           # Label text
  QColor    fgColor        # Foreground color
  QColor    bgColor        # Background color  
  QPixmap   pix            # Image data (often PNG)
  bool      noScaling      # version >= 15
  bool      showOnTop      # version >= 15
```

### Room connections
Rooms are linked through:
1. **12 standard exits** - each points to destination room ID (-1 = no exit)
2. **Special exits** - custom commands for non-standard movement

### Critical pitfalls
- QString length is in BYTES (must be even for UTF-16)
- QPixmap often contains inline PNG - scan for IEND + skip 4-byte CRC
- MudletLabel has 7 doubles before QString (not 5 or 6)
- Always use `bufio.Reader` for performance
- Version-dependent fields: symbolColor (v21+), specialExits format changes at v21
- Area structure differs significantly between v20 and v21 (labels moved inside area)

## CLI usage

```bash
# Parse and show stats
./mapsnap -map world.map -stats

# Validate map
./mapsnap -map world.map -validate

# Export to JSON
./mapsnap -map world.map -dump-json output.json

# Examine binary structure (compact summary)
./mapsnap -map world.map -examine

# Examine with detailed output (offsets, all values)
./mapsnap -map world.map -examine -debug

# Generate map fragment (target functionality)
./mapsnap -map world.map -room 1234 -output fragment.webp
```

### Flags
```
-map string       Path to Mudlet map file (.map/.dat)
-room int         Room ID to center on
-output string    Output file path
-dump-json string Export to JSON
-validate         Validate map integrity
-stats            Show statistics
-debug            Enable debug output (verbose mode for -examine)
-examine          Examine binary structure of map file
-timeout int      Timeout in seconds (default 30)
```

### The -examine command

The `-examine` flag walks through the binary map file and displays its structure.

**Compact mode** (`-examine`):
```
MudletMap.version:
  version = 20
areaNames QMap<int,QString>:
  count = 61
areas MudletAreas:
  count = 64 areas, total rooms = 26758
labels MudletLabels:
  areas with labels = 51, total labels = 397
rooms MudletRooms:
  total rooms = 26758
```

**Debug mode** (`-examine -debug`):
- Shows byte offsets for each section (e.g., `@1058553: rooms MudletRooms`)
- Lists all area names with IDs
- Shows detailed area info (room counts, z-levels, bounding box)
- Lists all labels with position, size, text, PNG bytes, and flags
- Shows first 5 rooms with full details (exits, name, environment, etc.)

Example room output:
```
id=15951 area=30 pos=(82,-9,0) exits=[ne:15950,e:15949,sw:15952,nw:15966] name='15951' env=303
```

### Environment variables
- `MAPSNAP_DEBUG=1` - Parser debug output
- `MAPSNAP_SKIP_LABELS=1` - Skip label parsing (performance/debug)

## Development guidelines

### Error handling
```go
// Use wrapped errors for context
if err := parseRoom(reader, version); err != nil {
    return fmt.Errorf("parsing room at offset %d: %w", offset, err)
}

// Handle file closing properly with errors.Join (Go 1.20+)
func example(path string) (err error) {
    f, err := os.Open(path)
    if err != nil {
        return fmt.Errorf("open: %w", err)
    }
    defer func() {
        if cerr := f.Close(); cerr != nil {
            err = errors.Join(err, fmt.Errorf("close: %w", cerr))
        }
    }()
    // ... work with file ...
    return nil
}
```

### Performance
- Use `bufio.Reader` for large files
- Validate reasonable bounds before loops (e.g., QMap count < threshold)
- Avoid redundant scanning - move forward systematically

### Testing
- Test with real Mudlet map files
- Compare results with Node.js parser reference
- Use fixtures in `tests/fixtures/`

## Reference documentation

### docs/sources/Mudlet/
C++ source excerpts from Mudlet client:
- `TRoom.cpp/h` - Room serialization
- `TArea.cpp/h` - Area handling
- `TRoomDB.cpp/h` - Room database I/O
- `TMap.h` - Main map class
- `TMapLabel.cpp/h` - Label serialization

### docs/sources/node-mudlet-map-binary-reader/
Working Node.js implementation:
- `index.js` - API entry point
- `map-operations.js` - Read/write logic
- `models/mudlet-models.js` - MudletMap, MudletRoom, MudletArea, MudletLabel
- `models/qstream-types.js` - QString, QColor, QPoint, QFont

## Roadmap

### Phase 1: MVP
- [x] Binary parser for Mudlet format v20
- [x] Map validation and stats
- [x] JSON export
- [x] Debug/examine tools
- [x] WEBP/PNG renderer (pure Go)
- [x] Labels with PNG pixmaps
- [x] Mudlet-compatible colors
- [x] Unit tests

### Phase 2: Features
- [ ] Custom line rendering
