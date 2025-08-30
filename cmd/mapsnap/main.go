package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/szydell/arkadia-mapsnap/pkg/mapparser"
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
	examine := flag.Bool("examine", false, "Examine the binary structure of the map file")
	examineQt := flag.Bool("examine-qt", false, "Examine Qt/MudletMap sections and offsets")
	timeout := flag.Int("timeout", 30, "Timeout in seconds for parsing operations")

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

	// Examine a file if requested
	if *examine {
		fmt.Printf("Examining map file: %s\n", *mapFile)
		if err := ExamineFile(*mapFile); err != nil {
			fmt.Printf("Error examining file: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	// Examine Qt sections if requested
	if *examineQt {
		fmt.Printf("Examining Qt/MudletMap sections: %s\n", *mapFile)
		if err := ExamineQt(*mapFile); err != nil {
			fmt.Printf("Error examining Qt sections: %v\n", err)
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
		len(m.Rooms), len(m.Areas), len(m.Environments))

	// Print debug information if requested
	if *debug {
		fmt.Println("\nDebug Information:")
		fmt.Printf("Map Version: %d\n", m.Header.Version)
		fmt.Printf("Magic String: %s\n", m.Header.Magic)

		// Print first 5 rooms for debugging
		fmt.Println("\nSample Rooms:")
		count := 0
		for id, room := range m.Rooms {
			fmt.Printf("Room %d: %s at (%d,%d,%d) with %d exits\n",
				id, room.Name, room.X, room.Y, room.Z, len(room.Exits))
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

	// If room ID is provided, we would render the map (not implemented yet)
	if *roomID > 0 && *outputFile != "" {
		fmt.Printf("Map rendering not implemented yet. Would render room %d to %s\n", *roomID, *outputFile)
	}
}

func printUsage() {
	fmt.Printf("arkadia-mapsnap %s - Mudlet map snapshot tool\n\n", version)
	fmt.Println("Usage:")
	fmt.Println("  mapsnap -map <file.map> [options]")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nExamples:")
	fmt.Println("  mapsnap -map arkadia.map -dump-json map.json")
	fmt.Println("  mapsnap -map arkadia.map -validate -stats")
	fmt.Println("  mapsnap -map arkadia.map -room 1234 -output map.webp")
	fmt.Println("  mapsnap -map arkadia.map -timeout 60 -stats")
}
