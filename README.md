# mudlet-mapsnap

[![Go Reference](https://pkg.go.dev/badge/github.com/szydell/mudlet-mapsnap.svg)](https://pkg.go.dev/github.com/szydell/mudlet-mapsnap)
[![Go Report Card](https://goreportcard.com/badge/github.com/szydell/mudlet-mapsnap)](https://goreportcard.com/report/github.com/szydell/mudlet-mapsnap)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

A Go library and CLI tool for parsing and visualizing Mudlet map files.

## Overview

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
│   ├── mapparser/     # Map file parsing library
│   └── maprenderer/   # Image rendering library
├── docs/              # Documentation and references
└── tests/fixtures/    # Test data
```

## Features

- Binary map file parsing (Mudlet format v6-20)
- Map validation and statistics
- JSON export for external tools
- Binary structure examination tools
- Visual map rendering to WEBP/PNG (pure Go, no CGO)
- Labels with PNG pixmaps
- Mudlet-compatible environment colors (ANSI 256-color palette)
- Contrast-aware room symbol colors
- Configurable rendering (dimensions, room size, spacing, shape)
- Auto-calculated room visibility based on image dimensions

## Library Usage

```go
package main

import (
    "fmt"
    "log"

    "github.com/szydell/mudlet-mapsnap/pkg/mapparser"
    "github.com/szydell/mudlet-mapsnap/pkg/maprenderer"
)

func main() {
    // Parse a Mudlet map file
    m, err := mapparser.ParseMapFile("world.map")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Loaded %d rooms in %d areas\n", m.RoomCount(), m.AreaCount())

    // Render a map fragment centered on room 1234
    cfg := maprenderer.DefaultConfig()
    cfg.Width = 1024
    cfg.Height = 768

    renderer := maprenderer.NewRenderer(cfg)
    renderer.SetMap(m)

    result, err := renderer.RenderFragment(1234)
    if err != nil {
        log.Fatal(err)
    }

    // Save to file
    err = maprenderer.SaveImage(result.Image, "map.webp", nil)
    if err != nil {
        log.Fatal(err)
    }
}
```

## API Documentation

Full API documentation is available at [pkg.go.dev](https://pkg.go.dev/github.com/szydell/mudlet-mapsnap).

### Key Packages

- **[mapparser](https://pkg.go.dev/github.com/szydell/mudlet-mapsnap/pkg/mapparser)** - Parse Mudlet map files and access room/area data
- **[maprenderer](https://pkg.go.dev/github.com/szydell/mudlet-mapsnap/pkg/maprenderer)** - Render map fragments to WEBP/PNG images

## Technical Documentation

See [AGENTS.md](AGENTS.md) for detailed technical documentation including:
- Binary format specification (QDataStream)
- Data structures (MudletMap, MudletRoom, MudletArea, MudletLabel)
- Development guidelines

## Contributing

Contributions are welcome! Please ensure all code follows Go conventions and includes appropriate documentation.

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.
