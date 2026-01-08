package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/szydell/arkadia-mapsnap/pkg/mapparser"
	"github.com/szydell/arkadia-mapsnap/pkg/maprenderer"
)

var (
	version = "dev"
)

func main() {
	// Define command line flags
	mapFile := flag.String("map", "", "Path to the Mudlet map file (.map)")
	roomID := flag.Int("room", 0, "Room ID to center the map on")
	outputFile := flag.String("output", "", "Output file path")
	dumpJSON := flag.String("dump-json", "", "Dump map to JSON file")
	validate := flag.Bool("validate", false, "Validate map integrity")
	showStats := flag.Bool("stats", false, "Show map statistics")
	debug := flag.Bool("debug", false, "Enable debug output")
	examine := flag.Bool("examine", false, "Examine Qt/MudletMap binary structure with offsets")
	timeout := flag.Int("timeout", 30, "Timeout in seconds for parsing operations")

	// Rendering options
	imgWidth := flag.Int("width", 800, "Output image width")
	imgHeight := flag.Int("height", 600, "Output image height")
	radius := flag.Int("radius", 15, "Rendering radius (rooms from center)")
	roomSize := flag.Int("room-size", 20, "Room size in pixels")
	roomSpacing := flag.Int("room-spacing", 25, "Room spacing in pixels")
	quality := flag.Float64("quality", 85, "WEBP output quality (0-100)")
	roundRooms := flag.Bool("round", false, "Draw rooms as circles")

	// Parse flags
	flag.Parse()

	// Show usage if no arguments provided
	if len(os.Args) == 1 {
		printUsage()
		os.Exit(0)
	}

	// Validate required arguments
	if *mapFile == "" {
		fmt.Println("Error: Map file is required")
		printUsage()
		os.Exit(1)
	}

	// Check if a map file exists
	if _, err := os.Stat(*mapFile); os.IsNotExist(err) {
		fmt.Printf("Error: Map file not found: %s\n", *mapFile)
		os.Exit(1)
	}

	// Examine file if requested
	if *examine {
		fmt.Printf("Examining map file: %s\n", *mapFile)
		if *debug {
			fmt.Println("(debug mode - showing detailed output)")
		}
		if err := ExamineFile(*mapFile, *debug); err != nil {
			fmt.Printf("Error examining file: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeout)*time.Second)
	defer cancel()

	// Create a channel to receive the parsing result
	resultCh := make(chan struct {
		m   *mapparser.Map
		err error
	})

	// Parse map file in a goroutine
	go func() {
		fmt.Printf("Parsing map file: %s (timeout: %d seconds)\n", *mapFile, *timeout)
		m, err := mapparser.ParseMapFile(*mapFile)
		resultCh <- struct {
			m   *mapparser.Map
			err error
		}{m, err}
	}()

	// Wait for either the parsing to complete or the timeout to expire
	var m *mapparser.Map
	var err error
	select {
	case result := <-resultCh:
		m = result.m
		err = result.err
	case <-ctx.Done():
		fmt.Println("Error: Parsing operation timed out. The map file may be too large or corrupted.")
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error parsing map file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Map parsed successfully. Found %d rooms, %d areas, %d environments.\n",
		len(m.Rooms), len(m.Areas), len(m.EnvColors)+len(m.CustomEnvColors))

	// Print debug information if requested
	if *debug {
		fmt.Println("\nDebug Information:")
		fmt.Printf("Map Version: %d\n", m.Version)

		// Print first 5 rooms for debugging
		fmt.Println("\nSample Rooms:")
		count := 0
		for id, room := range m.Rooms {
			activeExits := room.ActiveExits()
			fmt.Printf("Room %d: %s at (%d,%d,%d) with %d exits\n",
				id, room.Name, room.X, room.Y, room.Z, len(activeExits))
			count++
			if count >= 5 {
				break
			}
		}
	}

	// Validate map if requested
	if *validate {
		fmt.Println("Validating map...")
		errors := mapparser.ValidateMap(m)
		if len(errors) > 0 {
			fmt.Printf("Found %d validation errors:\n", len(errors))
			for i, err := range errors {
				fmt.Printf("%d. %s: %s\n", i+1, err.Type, err.Message)
			}
		} else {
			fmt.Println("Map validation passed. No errors found.")
		}
	}

	// Show map statistics if requested
	if *showStats {
		stats := mapparser.GetMapStats(m)
		fmt.Println("\nMap Statistics:")
		fmt.Printf("Total Rooms: %d\n", stats.TotalRooms)
		fmt.Printf("Total Areas: %d\n", stats.TotalAreas)
		fmt.Printf("Total Environments: %d\n", stats.TotalEnvironments)
		fmt.Printf("Z Levels: %v\n", stats.ZLevels)
		fmt.Printf("Bounding Box: X(%d,%d) Y(%d,%d) Z(%d,%d)\n",
			stats.BoundingBox.MinX, stats.BoundingBox.MaxX,
			stats.BoundingBox.MinY, stats.BoundingBox.MaxY,
			stats.BoundingBox.MinZ, stats.BoundingBox.MaxZ)

		// Display a list of all areas
		if stats.TotalAreas > 0 {
			fmt.Println("\nAreas:")
			// Get a sorted list of area IDs
			var areaIDs []int
			for id := range m.Areas {
				areaIDs = append(areaIDs, int(id))
			}
			sort.Ints(areaIDs)

			// Display each area
			for _, id := range areaIDs {
				area := m.Areas[int32(id)]
				fmt.Printf("  %3d: %s\n", id, area.Name)
			}
		}
	}

	// Dump to JSON if requested
	if *dumpJSON != "" {
		fmt.Printf("Exporting map to JSON: %s\n", *dumpJSON)
		if err := mapparser.ExportToJSON(m, *dumpJSON); err != nil {
			fmt.Printf("Error exporting to JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("JSON export completed successfully.")
	}

	// Render map fragment if room ID and output file provided
	if *roomID > 0 && *outputFile != "" {
		fmt.Printf("Rendering map fragment centered on room %d...\n", *roomID)

		// Configure renderer
		cfg := maprenderer.DefaultConfig()
		cfg.Width = *imgWidth
		cfg.Height = *imgHeight
		cfg.Radius = *radius
		cfg.RoomSize = *roomSize
		cfg.RoomSpacing = *roomSpacing
		cfg.RoomRound = *roundRooms

		// Create renderer
		renderer := maprenderer.NewRenderer(cfg)
		renderer.SetMap(m)

		// Render the fragment
		result, err := renderer.RenderFragment(int32(*roomID))
		if err != nil {
			fmt.Printf("Error rendering map: %v\n", err)
			os.Exit(1)
		}

		// Save the output
		outputOpts := &maprenderer.OutputOptions{
			Format:  maprenderer.FormatFromPath(*outputFile),
			Quality: float32(*quality),
		}

		if err := maprenderer.SaveImage(result.Image, *outputFile, outputOpts); err != nil {
			fmt.Printf("Error saving image: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Map fragment saved to: %s\n", *outputFile)
		fmt.Printf("  Center room: %d\n", result.CenterRoom)
		fmt.Printf("  Area: %s (ID: %d)\n", result.AreaName, result.AreaID)
		fmt.Printf("  Z-level: %d\n", result.ZLevel)
		fmt.Printf("  Rooms rendered: %d\n", result.RoomsDrawn)
		fmt.Printf("  Image size: %dx%d\n", result.Image.Bounds().Dx(), result.Image.Bounds().Dy())
	}
}

func printUsage() {
	fmt.Printf("arkadia-mapsnap %s - Mudlet map snapshot tool\n\n", version)
	fmt.Println("Usage:")
	fmt.Println("  mapsnap -map <file.map> [options]")
	fmt.Println("\nGeneral Options:")
	fmt.Println("  -map string       Path to Mudlet map file (.map)")
	fmt.Println("  -validate         Validate map integrity")
	fmt.Println("  -stats            Show map statistics")
	fmt.Println("  -dump-json string Export map to JSON")
	fmt.Println("  -examine          Examine binary structure")
	fmt.Println("  -debug            Enable debug output")
	fmt.Println("  -timeout int      Timeout in seconds (default 30)")
	fmt.Println("\nRendering Options:")
	fmt.Println("  -room int         Room ID to center the map on")
	fmt.Println("  -output string    Output file path (.webp or .png)")
	fmt.Println("  -width int        Output image width (default 800)")
	fmt.Println("  -height int       Output image height (default 600)")
	fmt.Println("  -radius int       Rendering radius in rooms (default 15)")
	fmt.Println("  -room-size int    Room size in pixels (default 20)")
	fmt.Println("  -room-spacing int Room spacing in pixels (default 25)")
	fmt.Println("  -quality float    WEBP quality 0-100 (default 85)")
	fmt.Println("  -round            Draw rooms as circles")
	fmt.Println("\nExamples:")
	fmt.Println("  mapsnap -map arkadia.map -stats")
	fmt.Println("  mapsnap -map arkadia.map -validate")
	fmt.Println("  mapsnap -map arkadia.map -dump-json map.json")
	fmt.Println("  mapsnap -map arkadia.map -room 1234 -output map.webp")
	fmt.Println("  mapsnap -map arkadia.map -room 1234 -output map.png -width 1200 -height 900")
	fmt.Println("  mapsnap -map arkadia.map -room 1234 -output map.webp -radius 20 -round")
}
