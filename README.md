# arkadia-mapsnap

A Go library and CLI tool for parsing and visualizing Mudlet map files from the MUD game "Arkadia".

## Description

**arkadia-mapsnap** is designed to work with Mudlet's binary map files (.map) and provide tools for:
- Parsing the map file format
- Extracting map data (rooms, areas, environments)
- Validating map integrity
- Exporting map data to JSON
- (Future) Generating visual snapshots of map areas

## Installation

### Prerequisites
- Go 1.18 or higher

### Building from source
```bash
# Clone the repository
git clone https://github.com/szydell/arkadia-mapsnap.git
cd arkadia-mapsnap

# Build the CLI tool
go build -o mapsnap ./cmd/mapsnap
```

## Usage

### Basic usage
```bash
# Parse a map file and show statistics
./mapsnap -map path/to/map.dat -stats

# Validate map integrity
./mapsnap -map path/to/map.dat -validate

# Export map to JSON
./mapsnap -map path/to/map.dat -dump-json output.json

# Show debug information
./mapsnap -map path/to/map.dat -debug
```

### Command-line options
```
  -debug
        Enable debug output
  -dump-json string
        Dump map to JSON file
  -map string
        Path to the Mudlet map file (.map)
  -output string
        Output file path
  -room int
        Room ID to center the map on
  -stats
        Show map statistics
  -validate
        Validate map integrity
```

## Project Structure

```
arkadia-mapsnap/
├── cmd/
│   └── mapsnap/               # CLI application
├── pkg/
│   ├── mapparser/             # Map file parsing
│   ├── maprenderer/           # Image generation (future)
│   └── maputils/              # Common utilities (future)
└── tests/
    └── fixtures/              # Test data
        └── sample_maps/       # Sample map files
```

## Features

### Current
- Binary map file parsing (supports Mudlet map format v1-v3)
- Map validation
- Map statistics
- JSON export

### Planned
- Visual map rendering
- Map navigation
- Path finding
- Custom styling options

## License

This project is licensed under the Apache License - see the LICENSE file for details.
