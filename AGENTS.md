# AI Assistant instructions for arkadia-mapsnap

## Project overview

**arkadia-mapsnap** is a Go library and CLI tool for parsing and visualizing Mudlet map files from the Polish MUD game "Arkadia". The project enables generating visual map fragments centered on selected locations.

### Problem solved
- Players use Mudlet as their MUD client
- Mudlet stores maps in a binary format (QDataStream) that's difficult to process
- Need for quick generation of visual map fragments for navigation and sharing

### Target functionality
1. **Library (`pkg/`)**: Parse Mudlet map files, search rooms, render images
2. **CLI (`cmd/mapsnap`)**: Command-line tool for map operations

## Core technologies

- **Go 1.24+**
- **QDataStream binary format** (Qt serialization)
- **Image output**: WEBP format (planned)

## Project structure

```
arkadia-mapsnap/
├── cmd/mapsnap/           # CLI application
│   ├── main.go           # Entry point and flags
│   ├── examine.go        # Binary examination tools
│   └── examine_qt.go     # Qt-specific examination
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
1. **MudletMap** - version → envColors → areaNames → customEnvColors → areas → rooms → labels
2. **QString** - quint32 length (BYTES, not chars) + UTF-16BE data. 0xFFFFFFFF = null string
3. **QMap<K,V>** - qint32 count + key-value pairs
4. **MudletRoom** - 16 standard exits + special exits, environment, weight, name, userData
5. **MudletLabel** - id, pos(3×double), 2×dummy(double), size(2×double), text, colors, pixmap, flags

### Critical pitfalls
- QString length is in BYTES (must be even for UTF-16)
- QPixmap often contains inline PNG - scan for IEND + skip 4-byte CRC
- MudletLabel has 7 doubles before QString (not 5 or 6)
- Always use `bufio.Reader` for performance

## CLI usage

```bash
# Parse and show stats
./mapsnap -map arkadia.map -stats

# Validate map
./mapsnap -map arkadia.map -validate

# Export to JSON
./mapsnap -map arkadia.map -dump-json output.json

# Examine binary structure
./mapsnap -map arkadia.map -examine
./mapsnap -map arkadia.map -examine-qt

# Generate map fragment (target functionality)
./mapsnap -map arkadia.map -room 1234 -output fragment.webp
```

### Flags
```
-map string       Path to Mudlet map file (.map/.dat)
-room int         Room ID to center on
-output string    Output file path
-dump-json string Export to JSON
-validate         Validate map integrity
-stats            Show statistics
-debug            Enable debug output
-examine          Examine binary structure
-examine-qt       Examine Qt/MudletMap sections
-timeout int      Timeout in seconds (default 30)
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
- Test with real Arkadia map files
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
- [ ] Basic WEBP renderer
- [ ] Unit tests

### Phase 2: Features
- [ ] Configurable rendering styles
- [ ] Batch processing
- [ ] YAML configuration files

### Phase 3: Extended
- [ ] HTTP API server
- [ ] Multiple output formats (PNG, SVG)
- [ ] Docker images
