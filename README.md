# arkadia-mapsnap

A Go library and CLI tool for parsing and visualizing Mudlet map files from the Polish MUD game "Arkadia".

## Description

**arkadia-mapsnap** parses Mudlet's binary map files (QDataStream format, version 20) and provides:
- Map file parsing and validation
- Room, area, and environment extraction
- JSON export for analysis
- Binary structure examination tools
- (In progress) Visual map fragment rendering to WEBP

## Installation

### Prerequisites
- Go 1.25 or higher

### Building from source
```bash
git clone https://github.com/szydell/arkadia-mapsnap.git
cd arkadia-mapsnap
make build
# or: go build -o mapsnap ./cmd/mapsnap
```

## Usage

```bash
# Parse and show statistics
./mapsnap -map arkadia.map -stats

# Validate map integrity
./mapsnap -map arkadia.map -validate

# Export to JSON
./mapsnap -map arkadia.map -dump-json output.json

# Examine binary structure
./mapsnap -map arkadia.map -examine
./mapsnap -map arkadia.map -examine-qt

# Generate map fragment (target functionality)
./mapsnap -map arkadia.map -room 1234 -output fragment.webp
```

### Command-line flags
```
-map string       Path to Mudlet map file (.map/.dat)
-room int         Room ID to center on
-output string    Output file path
-dump-json string Export map to JSON
-validate         Validate map integrity
-stats            Show map statistics
-debug            Enable debug output
-examine          Examine binary structure
-examine-qt       Examine Qt/MudletMap sections
-timeout int      Timeout in seconds (default 30)
```

### Environment variables
- `MAPSNAP_DEBUG=1` - Enable parser debug output
- `MAPSNAP_SKIP_LABELS=1` - Skip label parsing

## Project Structure

```
arkadia-mapsnap/
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

### In Progress
- Visual map rendering to WEBP
- Configurable rendering styles

### Planned
- HTTP API server
- Multiple output formats (PNG, SVG)
- Batch processing

## Documentation

See [AGENTS.md](AGENTS.md) for detailed technical documentation and development guidelines.

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.
