# mudlet-mapsnap

A Go library and CLI tool for parsing and visualizing Mudlet map files.

## Description

**mudlet-mapsnap** parses Mudlet's binary map files (QDataStream format, version 20) and provides:
- Map file parsing and validation
- Room, area, and environment extraction
- JSON export for analysis
- Binary structure examination tools
- Visual map fragment rendering to WEBP/PNG (pure Go, no CGO required)

## Installation

### Prerequisites
- Go 1.25 or higher

### Building from source
```bash
git clone https://github.com/szydell/mudlet-mapsnap.git
cd mudlet-mapsnap
make build
# or: go build -o mapsnap ./cmd/mapsnap
```

## Usage

```bash
# Parse and show statistics
./mapsnap -map world.map -stats

# Validate map integrity
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

### Command-line flags
```
-map string       Path to Mudlet map file (.map/.dat)
-room int         Room ID to center on
-output string    Output file path (supports .webp and .png)
-radius int       Rendering radius in rooms (default 15)
-width int        Output image width (default 800)
-height int       Output image height (default 600)
-room-size int    Room size in pixels (default 20)
-room-spacing int Room spacing in pixels (default 25)
-round            Draw rooms as circles instead of squares
-dump-json string Export map to JSON
-validate         Validate map integrity
-stats            Show map statistics
-debug            Enable debug output (verbose mode for -examine)
-examine          Examine binary structure of map file
-timeout int      Timeout in seconds (default 30)
```

### Environment variables
- `MAPSNAP_DEBUG=1` - Enable parser debug output
- `MAPSNAP_SKIP_LABELS=1` - Skip label parsing

## Project Structure

```
mudlet-mapsnap/
├── cmd/mapsnap/       # CLI application
├── pkg/
│   ├── mapparser/     # Map file parsing
│   ├── maprenderer/   # Image generation (WIP)
│   └── maputils/      # Common utilities
├── docs/sources/      # Reference implementations
└── tests/fixtures/    # Test data
```

## Features

### Current
- Binary map file parsing (Mudlet format v20)
- Map validation and statistics
- JSON export
- Binary structure examination tools
- Visual map rendering to WEBP/PNG (pure Go, no CGO)
- Labels with PNG pixmaps
- Mudlet-compatible environment colors
- Contrast-aware room symbol colors
- Configurable rendering (size, spacing, radius, round rooms)

## Documentation

See [AGENTS.md](AGENTS.md) for detailed technical documentation and development guidelines.

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.
