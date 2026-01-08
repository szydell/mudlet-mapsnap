package maprenderer

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"sort"

	"github.com/szydell/arkadia-mapsnap/pkg/mapparser"
)

// Renderer handles map rendering operations
type Renderer struct {
	config  *Config
	mapData *mapparser.MudletMap
}

// NewRenderer creates a new renderer with the given configuration
func NewRenderer(cfg *Config) *Renderer {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Renderer{
		config: cfg,
	}
}

// SetMap sets the map data to render
func (r *Renderer) SetMap(m *mapparser.MudletMap) {
	r.mapData = m
}

// RenderResult contains the rendered image and metadata
type RenderResult struct {
	Image      *image.RGBA
	CenterRoom int32
	AreaID     int32
	AreaName   string
	ZLevel     int32
	RoomsDrawn int
}

// RenderFragment renders a map fragment centered on the given room
func (r *Renderer) RenderFragment(roomID int32) (*RenderResult, error) {
	if r.mapData == nil {
		return nil, fmt.Errorf("no map data loaded")
	}

	centerRoom := r.mapData.GetRoom(roomID)
	if centerRoom == nil {
		return nil, fmt.Errorf("room %d not found", roomID)
	}

	area := r.mapData.GetArea(centerRoom.Area)
	if area == nil {
		return nil, fmt.Errorf("area %d not found", centerRoom.Area)
	}

	// Create the output image
	img := image.NewRGBA(image.Rect(0, 0, r.config.Width, r.config.Height))

	// Fill background
	draw.Draw(img, img.Bounds(), &image.Uniform{r.config.BackgroundColor}, image.Point{}, draw.Src)

	// Calculate rendering parameters
	centerX := centerRoom.X
	centerY := centerRoom.Y
	centerZ := centerRoom.Z
	areaID := centerRoom.Area

	halfWidth := r.config.Width / 2
	halfHeight := r.config.Height / 2
	spacing := r.config.RoomSpacing

	// Build custom environment colors map from map data
	customEnvColors := make(map[int32]color.RGBA)
	for envID, c := range r.mapData.CustomEnvColors {
		rc, gc, bc, ac := c.ToRGBA()
		customEnvColors[envID] = color.RGBA{R: rc, G: gc, B: bc, A: ac}
	}

	// Collect rooms to render - ONLY from the same area
	roomsToRender := r.collectRoomsInArea(centerX, centerY, centerZ, int32(r.config.Radius), areaID)

	// Build room lookup map
	roomMap := make(map[int32]*mapparser.MudletRoom)
	for _, room := range roomsToRender {
		roomMap[room.ID] = room
	}

	// Optionally draw lower level rooms (same area only)
	if r.config.ShowLowerLevel {
		lowerRooms := r.collectRoomsInArea(centerX, centerY, centerZ-1, int32(r.config.Radius), areaID)
		r.drawOtherLevelRooms(img, lowerRooms, centerX, centerY, halfWidth, halfHeight, spacing, true)
	}

	// Optionally draw upper level rooms (same area only)
	if r.config.ShowUpperLevel {
		upperRooms := r.collectRoomsInArea(centerX, centerY, centerZ+1, int32(r.config.Radius), areaID)
		r.drawOtherLevelRooms(img, upperRooms, centerX, centerY, halfWidth, halfHeight, spacing, false)
	}

	// Draw background labels (under everything)
	r.drawLabels(img, areaID, centerZ, false, centerX, centerY, halfWidth, halfHeight, spacing)

	// Draw exits FIRST (under rooms)
	r.drawExits(img, roomsToRender, roomMap, centerX, centerY, halfWidth, halfHeight, spacing, areaID)

	// Draw rooms on current z-level
	roomsDrawn := 0
	for _, room := range roomsToRender {
		screenX, screenY := r.roomToScreen(room, centerX, centerY, halfWidth, halfHeight, spacing)

		// Check if room is within image bounds
		margin := r.config.RoomSize
		if screenX < -margin || screenX > r.config.Width+margin ||
			screenY < -margin || screenY > r.config.Height+margin {
			continue
		}

		// Get room color based on environment
		envColor := r.getEnvColor(room.Environment, customEnvColors)

		// Draw the room
		r.drawRoom(img, screenX, screenY, envColor, room)
		roomsDrawn++
	}

	// Draw player room highlight (gradient like Mudlet)
	r.drawPlayerHighlight(img, halfWidth, halfHeight)

	// Draw foreground labels (on top of everything)
	r.drawLabels(img, areaID, centerZ, true, centerX, centerY, halfWidth, halfHeight, spacing)

	return &RenderResult{
		Image:      img,
		CenterRoom: roomID,
		AreaID:     centerRoom.Area,
		AreaName:   area.Name,
		ZLevel:     centerZ,
		RoomsDrawn: roomsDrawn,
	}, nil
}

// roomToScreen converts room coordinates to screen coordinates
func (r *Renderer) roomToScreen(room *mapparser.MudletRoom, centerX, centerY int32, halfWidth, halfHeight, spacing int) (int, int) {
	dx := int(room.X - centerX)
	dy := int(room.Y - centerY)
	// Y is flipped: in Mudlet, Y increases upward, but screen Y increases downward
	return halfWidth + dx*spacing, halfHeight - dy*spacing
}

// collectRoomsInArea returns all rooms within radius of center point, filtered by area and z-level
func (r *Renderer) collectRoomsInArea(centerX, centerY, centerZ, radius, areaID int32) []*mapparser.MudletRoom {
	var rooms []*mapparser.MudletRoom

	for _, room := range r.mapData.Rooms {
		// Filter by area - this is the key fix!
		if room.Area != areaID {
			continue
		}

		if room.Z != centerZ {
			continue
		}

		dx := abs32(room.X - centerX)
		dy := abs32(room.Y - centerY)

		// Use Chebyshev distance (max of dx, dy) for square area
		if dx <= radius && dy <= radius {
			rooms = append(rooms, room)
		}
	}

	// Sort by rendering order (Y desc, then X asc for consistent drawing)
	sort.Slice(rooms, func(i, j int) bool {
		if rooms[i].Y != rooms[j].Y {
			return rooms[i].Y > rooms[j].Y
		}
		return rooms[i].X < rooms[j].X
	})

	return rooms
}

// drawRoom draws a single room at the given screen coordinates
func (r *Renderer) drawRoom(img *image.RGBA, x, y int, roomColor color.RGBA, room *mapparser.MudletRoom) {
	halfSize := r.config.RoomSize / 2

	if r.config.RoomRound {
		r.drawFilledCircle(img, x, y, halfSize, roomColor)
		if r.config.RoomBorder {
			r.drawCircleOutline(img, x, y, halfSize, r.config.BorderColor)
		}
	} else {
		r.drawFilledRect(img, x-halfSize, y-halfSize, r.config.RoomSize, r.config.RoomSize, roomColor)
		if r.config.RoomBorder {
			r.drawRectOutline(img, x-halfSize, y-halfSize, r.config.RoomSize, r.config.RoomSize, r.config.BorderColor)
		}
	}

	// Draw up/down indicators
	r.drawUpDownIndicators(img, x, y, room, roomColor)

	// Draw room symbol if present
	if r.config.ShowSymbol && room.Symbol != "" {
		r.drawRoomSymbol(img, x, y, room.Symbol, room, roomColor)
	}
}

// drawRoomSymbol draws the room symbol text
func (r *Renderer) drawRoomSymbol(img *image.RGBA, cx, cy int, symbol string, room *mapparser.MudletRoom, roomColor color.RGBA) {
	if len(symbol) == 0 {
		return
	}

	// Mudlet logic: use room's symbolColor if set, otherwise contrast with room color
	var symbolColor color.RGBA
	if room.SymbolColor != nil {
		r, g, b, a := room.SymbolColor.ToRGBA()
		symbolColor = color.RGBA{R: r, G: g, B: b, A: a}
	} else {
		// Calculate lightness of room color (simple average)
		lightness := (int(roomColor.R) + int(roomColor.G) + int(roomColor.B)) / 3
		if lightness > 127 {
			symbolColor = color.RGBA{R: 0, G: 0, B: 0, A: 255} // Black on light
		} else {
			symbolColor = color.RGBA{R: 255, G: 255, B: 255, A: 255} // White on dark
		}
	}
	size := max(3, r.config.RoomSize/4)

	// Get first character
	ch := rune(symbol[0])

	// Try to draw as bitmap letter first
	if r.drawBitmapChar(img, cx, cy, ch, symbolColor) {
		return
	}

	// Fallback for special symbols
	switch symbol {
	case "X", "x":
		r.drawLine(img, cx-size, cy-size, cx+size, cy+size, symbolColor)
		r.drawLine(img, cx+size, cy-size, cx-size, cy+size, symbolColor)
	case "+":
		r.drawLine(img, cx-size, cy, cx+size, cy, symbolColor)
		r.drawLine(img, cx, cy-size, cx, cy+size, symbolColor)
	case "O", "o", "0":
		r.drawCircleOutline(img, cx, cy, size, symbolColor)
	default:
		// Draw a small filled square as generic indicator
		halfS := size / 2
		r.drawFilledRect(img, cx-halfS, cy-halfS, size, size, symbolColor)
	}
}

// drawUpDownIndicators draws Mudlet-like up/down markers.
// In Mudlet these are small triangles centered horizontally, offset from room center,
// filled with hatch patterns (Dense4 for real exits, DiagCross for stubs), and optionally
// highlighted in door color.
func (r *Renderer) drawUpDownIndicators(img *image.RGBA, cx, cy int, room *mapparser.MudletRoom, roomColor color.RGBA) {
	// Mudlet constants:
	// allInsideTipOffsetFactor = 1/20, upDownXOrYFactor = 1/3.1
	tipOffset := float64(r.config.RoomSize) * (1.0 / 20.0)
	baseOffset := float64(r.config.RoomSize) * (1.0 / 3.1)

	// Pick black/white based on room color lightness
	lc := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	if rgbaLightness(roomColor) > 127 {
		lc = color.RGBA{R: 0, G: 0, B: 0, A: 255}
	}

	// Door colors from Mudlet
	openDoor := color.RGBA{R: 10, G: 155, B: 10, A: 255}
	closedDoor := color.RGBA{R: 155, G: 155, B: 10, A: 255}
	lockedDoor := color.RGBA{R: 155, G: 10, B: 10, A: 255}

	// Helpers
	getDoorColor := func(key string) (c color.RGBA, isDoor bool) {
		status, ok := room.Doors[key]
		if !ok {
			return lc, false
		}
		switch status {
		case 1:
			return openDoor, true
		case 2:
			return closedDoor, true
		case 3:
			return lockedDoor, true
		default:
			return lc, false
		}
	}
	hasStub := func(dir int32) bool {
		for _, d := range room.ExitStubs {
			if d == dir {
				return true
			}
		}
		return false
	}

	// UP marker (triangle pointing up) shown when there is a real up exit OR an up stub
	if room.HasExit(mapparser.ExitUp) || hasStub(mapparser.ExitUp) {
		isReal := room.HasExit(mapparser.ExitUp)
		fill, isDoor := getDoorColor("up")
		if !isDoor {
			fill = lc
		}
		p0 := fPoint{X: float64(cx), Y: float64(cy) + tipOffset}
		p1 := fPoint{X: float64(cx) - baseOffset, Y: float64(cy) + baseOffset}
		p2 := fPoint{X: float64(cx) + baseOffset, Y: float64(cy) + baseOffset}
		pattern := hatchDense
		if !isReal {
			pattern = hatchDiagCross
		}
		r.fillTriangleHatch(img, p0, p1, p2, fill, pattern)
		r.strokeTriangle(img, p0, p1, p2, lc)
		if isDoor {
			r.strokeTriangle(img, p0, p1, p2, fill)
		}
	}

	// DOWN marker (triangle pointing down) shown when there is a real down exit OR a down stub
	if room.HasExit(mapparser.ExitDown) || hasStub(mapparser.ExitDown) {
		isReal := room.HasExit(mapparser.ExitDown)
		fill, isDoor := getDoorColor("down")
		if !isDoor {
			fill = lc
		}
		p0 := fPoint{X: float64(cx), Y: float64(cy) - tipOffset}
		p1 := fPoint{X: float64(cx) - baseOffset, Y: float64(cy) - baseOffset}
		p2 := fPoint{X: float64(cx) + baseOffset, Y: float64(cy) - baseOffset}
		pattern := hatchDense
		if !isReal {
			pattern = hatchDiagCross
		}
		r.fillTriangleHatch(img, p0, p1, p2, fill, pattern)
		r.strokeTriangle(img, p0, p1, p2, lc)
		if isDoor {
			r.strokeTriangle(img, p0, p1, p2, fill)
		}
	}
}

// drawPlayerHighlight draws the player room highlight with gradient effect
func (r *Renderer) drawPlayerHighlight(img *image.RGBA, x, y int) {
	// Draw a radial gradient highlight like Mudlet does
	outerRadius := r.config.RoomSize/2 + 8
	innerRadius := r.config.RoomSize/2 + 2

	playerColor := r.config.PlayerRoomColor

	// Draw gradient rings from outer to inner
	for radius := outerRadius; radius >= innerRadius; radius-- {
		// Calculate alpha based on position in gradient
		t := float64(radius-innerRadius) / float64(outerRadius-innerRadius)
		alpha := uint8(float64(playerColor.A) * (1.0 - t*0.7))

		ringColor := color.RGBA{R: playerColor.R, G: playerColor.G, B: playerColor.B, A: alpha}
		r.drawCircleOutline(img, x, y, radius, ringColor)
	}

	// Draw solid inner ring
	r.drawCircleOutline(img, x, y, innerRadius, playerColor)
	r.drawCircleOutline(img, x, y, innerRadius+1, playerColor)
}

// drawExits draws exit lines between rooms
func (r *Renderer) drawExits(img *image.RGBA, rooms []*mapparser.MudletRoom, roomMap map[int32]*mapparser.MudletRoom,
	centerX, centerY int32, halfWidth, halfHeight, spacing int, currentAreaID int32) {

	// Direction unit vectors (for exit line direction from room center)
	// Note: Y is inverted for screen coordinates
	dirVectors := [][2]float64{
		{0, -1},          // North (up on screen)
		{0.707, -0.707},  // Northeast
		{1, 0},           // East
		{0.707, 0.707},   // Southeast
		{0, 1},           // South (down on screen)
		{-0.707, 0.707},  // Southwest
		{-1, 0},          // West
		{-0.707, -0.707}, // Northwest
	}

	drawnExits := make(map[string]bool)
	halfRoom := float64(r.config.RoomSize) / 2.0

	for _, room := range rooms {
		fromX, fromY := r.roomToScreen(room, centerX, centerY, halfWidth, halfHeight, spacing)

		// Draw standard exits (first 8 directions - horizontal plane)
		for dir := 0; dir < 8; dir++ {
			destID := room.Exits[dir]
			if destID == mapparser.NoExit {
				continue
			}

			// Get destination room
			destRoom := r.mapData.GetRoom(destID)
			if destRoom == nil {
				continue
			}

			// Check if destination is in same area
			if destRoom.Area != currentAreaID {
				// Area exit - draw stub with arrow pointing outward
				r.drawAreaExitStub(img, fromX, fromY, dir, dirVectors[dir], halfRoom)
				continue
			}

			// Check if destination is on different Z level
			if destRoom.Z != room.Z {
				// Different Z level - draw stub
				r.drawExitStub(img, fromX, fromY, dir, dirVectors[dir], halfRoom)
				continue
			}

			// Check if destination is in current view
			destInView := roomMap[destID] != nil

			if !destInView {
				// Not in view - draw stub
				r.drawExitStub(img, fromX, fromY, dir, dirVectors[dir], halfRoom)
				continue
			}

			// Avoid drawing the same exit twice
			minID := min32(room.ID, destID)
			maxID := max32(room.ID, destID)
			key := fmt.Sprintf("%d-%d", minID, maxID)

			if drawnExits[key] {
				continue
			}
			drawnExits[key] = true

			toX, toY := r.roomToScreen(destRoom, centerX, centerY, halfWidth, halfHeight, spacing)

			// Calculate exit line start and end points (from room edges, not centers)
			// Line goes from edge of source room towards edge of destination room
			dx := float64(toX - fromX)
			dy := float64(toY - fromY)
			length := math.Sqrt(dx*dx + dy*dy)

			if length < 1 {
				continue
			}

			// Normalize
			nx := dx / length
			ny := dy / length

			// Start from edge of source room, end at edge of dest room
			startX := float64(fromX) + nx*halfRoom
			startY := float64(fromY) + ny*halfRoom
			endX := float64(toX) - nx*halfRoom
			endY := float64(toY) - ny*halfRoom

			// Check if it's a one-way exit
			isOneWay := !r.hasReturnExit(room.ID, destRoom, dir)

			exitColor := r.config.ExitColor
			if isOneWay {
				// Dotted line for one-way (we'll use a different color)
				exitColor = color.RGBA{R: 180, G: 180, B: 180, A: 180}
				r.drawDottedLine(img, int(startX), int(startY), int(endX), int(endY), exitColor)
				// Draw arrow
				r.drawArrowHead(img, int(endX), int(endY), nx, ny, exitColor)
			} else {
				r.drawLine(img, int(startX), int(startY), int(endX), int(endY), exitColor)
			}

			// Draw doors if present
			r.drawDoor(img, room, dir, int(startX), int(startY), int(endX), int(endY))
		}

		// Draw stub exits
		for _, stubDir := range room.ExitStubs {
			if stubDir < 0 || stubDir >= 8 {
				continue
			}
			// Check if there's already a real exit in this direction
			if room.Exits[stubDir] != mapparser.NoExit {
				continue
			}
			r.drawExitStub(img, fromX, fromY, int(stubDir), dirVectors[stubDir], halfRoom)
		}

		// Draw custom lines (used for special exits like "drzwi", "dziob" etc.)
		r.drawCustomLines(img, room, centerX, centerY, halfWidth, halfHeight, spacing)
	}
}

// drawExitStub draws a stub exit line with a small circle at the end
func (r *Renderer) drawExitStub(img *image.RGBA, fromX, fromY, dir int, dirVec [2]float64, halfRoom float64) {
	stubLen := halfRoom * 0.8
	startX := float64(fromX) + dirVec[0]*halfRoom
	startY := float64(fromY) + dirVec[1]*halfRoom
	endX := startX + dirVec[0]*stubLen
	endY := startY + dirVec[1]*stubLen

	stubColor := r.config.ExitColor
	r.drawLine(img, int(startX), int(startY), int(endX), int(endY), stubColor)

	// Draw small filled circle at stub end
	dotRadius := max(2, r.config.RoomSize/10)
	r.drawFilledCircle(img, int(endX), int(endY), dotRadius, stubColor)
}

// drawCustomLines draws custom lines for special exits
// CustomLines are used in Mudlet for non-standard directions like "drzwi", "dziob", etc.
// Points in customLines are in absolute map coordinates.
// Qt::PenStyle: 0=NoPen, 1=SolidLine, 2=DashLine, 3=DotLine, 4=DashDotLine, 5=DashDotDotLine
func (r *Renderer) drawCustomLines(img *image.RGBA, room *mapparser.MudletRoom,
	centerX, centerY int32, halfWidth, halfHeight, spacing int) {

	if len(room.CustomLines) == 0 {
		return
	}

	for exitName, points := range room.CustomLines {
		if len(points) == 0 {
			continue
		}

		// Get line color (default to exit color if not specified)
		lineColor := r.config.ExitColor
		if c, ok := room.CustomLinesColor[exitName]; ok {
			rc, gc, bc, ac := c.ToRGBA()
			lineColor = color.RGBA{R: rc, G: gc, B: bc, A: ac}
		}

		// Get line style - Qt::PenStyle enum
		// 0=NoPen, 1=SolidLine, 2=DashLine, 3=DotLine, 4=DashDotLine, 5=DashDotDotLine
		lineStyle := int32(1) // Default to solid
		if style, ok := room.CustomLinesStyle[exitName]; ok {
			lineStyle = style
		}

		// Draw arrow at end?
		hasArrow := false
		if arrow, ok := room.CustomLinesArrow[exitName]; ok {
			hasArrow = arrow
		}

		// Start from room center (in screen coordinates)
		roomScreenX := halfWidth + int(room.X-centerX)*spacing
		roomScreenY := halfHeight - int(room.Y-centerY)*spacing

		prevX := roomScreenX
		prevY := roomScreenY

		// Draw line segments through all points
		// Points are in absolute map coordinates
		for _, pt := range points {
			// Convert absolute map coordinates to screen coordinates
			ptScreenX := halfWidth + int(math.Round(pt.X)-float64(centerX))*spacing
			ptScreenY := halfHeight - int(math.Round(pt.Y)-float64(centerY))*spacing

			// Draw line segment based on style
			switch lineStyle {
			case 0: // NoPen - don't draw
				// skip
			case 2: // DashLine
				r.drawDashedLine(img, prevX, prevY, ptScreenX, ptScreenY, lineColor)
			case 3: // DotLine
				r.drawDottedLine(img, prevX, prevY, ptScreenX, ptScreenY, lineColor)
			case 4, 5: // DashDotLine, DashDotDotLine - use dashed for simplicity
				r.drawDashedLine(img, prevX, prevY, ptScreenX, ptScreenY, lineColor)
			default: // 1 = SolidLine (default)
				r.drawLine(img, prevX, prevY, ptScreenX, ptScreenY, lineColor)
			}

			prevX = ptScreenX
			prevY = ptScreenY
		}

		// Draw arrow at last point if requested
		if hasArrow && len(points) > 0 {
			lastPt := points[len(points)-1]
			lastX := halfWidth + int(math.Round(lastPt.X)-float64(centerX))*spacing
			lastY := halfHeight - int(math.Round(lastPt.Y)-float64(centerY))*spacing

			// Calculate direction for arrow
			var dx, dy float64
			if len(points) >= 2 {
				prevPt := points[len(points)-2]
				prevPtX := halfWidth + int(math.Round(prevPt.X)-float64(centerX))*spacing
				prevPtY := halfHeight - int(math.Round(prevPt.Y)-float64(centerY))*spacing
				dx = float64(lastX - prevPtX)
				dy = float64(lastY - prevPtY)
			} else {
				dx = float64(lastX - roomScreenX)
				dy = float64(lastY - roomScreenY)
			}

			length := math.Sqrt(dx*dx + dy*dy)
			if length > 0 {
				dx /= length
				dy /= length
				r.drawArrowHead(img, lastX, lastY, dx, dy, lineColor)
			}
		}
	}
}

// drawAreaExitStub draws a stub for exits leading to other areas (with arrow)
func (r *Renderer) drawAreaExitStub(img *image.RGBA, fromX, fromY, dir int, dirVec [2]float64, halfRoom float64) {
	stubLen := halfRoom * 1.2
	startX := float64(fromX) + dirVec[0]*halfRoom
	startY := float64(fromY) + dirVec[1]*halfRoom
	endX := startX + dirVec[0]*stubLen
	endY := startY + dirVec[1]*stubLen

	// Use a distinct color for area exits
	areaExitColor := color.RGBA{R: 200, G: 100, B: 100, A: 255}
	r.drawLine(img, int(startX), int(startY), int(endX), int(endY), areaExitColor)

	// Draw arrow head
	r.drawArrowHead(img, int(endX), int(endY), dirVec[0], dirVec[1], areaExitColor)
}

// drawArrowHead draws an arrow head at the given position
func (r *Renderer) drawArrowHead(img *image.RGBA, x, y int, dx, dy float64, c color.RGBA) {
	arrowLen := float64(max(4, r.config.RoomSize/4))
	arrowAngle := math.Pi / 6 // 30 degrees

	sin1 := math.Sin(arrowAngle)
	cos1 := math.Cos(arrowAngle)

	// Arrow points
	ax1 := float64(x) - arrowLen*(dx*cos1-dy*sin1)
	ay1 := float64(y) - arrowLen*(dy*cos1+dx*sin1)
	ax2 := float64(x) - arrowLen*(dx*cos1+dy*sin1)
	ay2 := float64(y) - arrowLen*(dy*cos1-dx*sin1)

	r.drawLine(img, x, y, int(ax1), int(ay1), c)
	r.drawLine(img, x, y, int(ax2), int(ay2), c)
}

// drawDoor draws door indicators on an exit
func (r *Renderer) drawDoor(img *image.RGBA, room *mapparser.MudletRoom, dir int, x1, y1, x2, y2 int) {
	dirName := mapparser.ExitDirectionShortNames[dir]
	doorStatus, hasDoor := room.Doors[dirName]
	if !hasDoor || doorStatus == 0 {
		return
	}

	// Calculate door position (middle of the exit line, closer to source room)
	midX := (x1 + x2) / 2
	midY := (y1 + y2) / 2

	// Door colors from Mudlet
	var doorColor color.RGBA
	switch doorStatus {
	case 1: // Open
		doorColor = color.RGBA{R: 10, G: 155, B: 10, A: 255}
	case 2: // Closed
		doorColor = color.RGBA{R: 155, G: 155, B: 10, A: 255}
	case 3: // Locked
		doorColor = color.RGBA{R: 155, G: 10, B: 10, A: 255}
	default:
		return
	}

	// Draw X shape for door
	doorSize := max(3, r.config.RoomSize/6)
	r.drawLine(img, midX-doorSize, midY-doorSize, midX+doorSize, midY+doorSize, doorColor)
	r.drawLine(img, midX+doorSize, midY-doorSize, midX-doorSize, midY+doorSize, doorColor)
}

// hasReturnExit checks if destRoom has an exit back to srcRoomID in the opposite direction
func (r *Renderer) hasReturnExit(srcRoomID int32, destRoom *mapparser.MudletRoom, direction int) bool {
	opposite := []int{4, 5, 6, 7, 0, 1, 2, 3} // N<->S, NE<->SW, etc.
	if direction >= len(opposite) {
		return false
	}
	return destRoom.Exits[opposite[direction]] == srcRoomID
}

// drawOtherLevelRooms draws rooms from other z-levels with transparency
func (r *Renderer) drawOtherLevelRooms(img *image.RGBA, rooms []*mapparser.MudletRoom,
	centerX, centerY int32, halfWidth, halfHeight, spacing int, isLower bool) {

	var levelColor color.RGBA
	var offsetX, offsetY int

	if isLower {
		levelColor = color.RGBA{R: 50, G: 50, B: 70, A: r.config.LowerLevelAlpha}
		offsetX, offsetY = -2, 2 // Offset down-left
	} else {
		levelColor = color.RGBA{R: 70, G: 70, B: 50, A: r.config.UpperLevelAlpha}
		offsetX, offsetY = 2, -2 // Offset up-right
	}

	halfSize := r.config.RoomSize / 2

	for _, room := range rooms {
		screenX, screenY := r.roomToScreen(room, centerX, centerY, halfWidth, halfHeight, spacing)
		screenX += offsetX
		screenY += offsetY

		if isLower {
			r.drawFilledRect(img, screenX-halfSize, screenY-halfSize, r.config.RoomSize, r.config.RoomSize, levelColor)
		} else {
			r.drawRectOutline(img, screenX-halfSize, screenY-halfSize, r.config.RoomSize, r.config.RoomSize, levelColor)
		}
	}
}

// getEnvColor returns the color for an environment ID
// Mudlet behavior: if env is not in mEnvColors AND not in mCustomEnvColors,
// it defaults to env=1 (red). We replicate this behavior.
func (r *Renderer) getEnvColor(env int32, customColors map[int32]color.RGBA) color.RGBA {
	// First check mEnvColors mapping
	if mappedEnv, ok := r.mapData.EnvColors[env]; ok {
		env = mappedEnv
	}

	// If env is NOT a default color (1-16) and NOT in customColors,
	// fall back to env=1 (red) like Mudlet does
	_, isDefault := r.config.DefaultEnvColors[env]
	_, isCustom := customColors[env]
	if !isDefault && !isCustom {
		env = 1 // Default to red
	}

	return envToColor(env, customColors, r.config.DefaultEnvColors)
}

// Drawing primitives

func (r *Renderer) drawFilledRect(img *image.RGBA, x, y, w, h int, c color.RGBA) {
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			blendPixel(img, x+dx, y+dy, c)
		}
	}
}

func (r *Renderer) drawRectOutline(img *image.RGBA, x, y, w, h int, c color.RGBA) {
	for dx := 0; dx < w; dx++ {
		setPixelSafe(img, x+dx, y, c)
		setPixelSafe(img, x+dx, y+h-1, c)
	}
	for dy := 0; dy < h; dy++ {
		setPixelSafe(img, x, y+dy, c)
		setPixelSafe(img, x+w-1, y+dy, c)
	}
}

func (r *Renderer) drawFilledCircle(img *image.RGBA, cx, cy, radius int, c color.RGBA) {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy <= radius*radius {
				blendPixel(img, cx+dx, cy+dy, c)
			}
		}
	}
}

func (r *Renderer) drawCircleOutline(img *image.RGBA, cx, cy, radius int, c color.RGBA) {
	x := radius
	y := 0
	err := 0

	for x >= y {
		setPixelSafe(img, cx+x, cy+y, c)
		setPixelSafe(img, cx+y, cy+x, c)
		setPixelSafe(img, cx-y, cy+x, c)
		setPixelSafe(img, cx-x, cy+y, c)
		setPixelSafe(img, cx-x, cy-y, c)
		setPixelSafe(img, cx-y, cy-x, c)
		setPixelSafe(img, cx+y, cy-x, c)
		setPixelSafe(img, cx+x, cy-y, c)

		y++
		if err <= 0 {
			err += 2*y + 1
		}
		if err > 0 {
			x--
			err -= 2*x + 1
		}
	}
}

func (r *Renderer) drawLine(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := 1
	if x1 >= x2 {
		sx = -1
	}
	sy := 1
	if y1 >= y2 {
		sy = -1
	}
	err := dx - dy

	for {
		setPixelSafe(img, x1, y1, c)

		if x1 == x2 && y1 == y2 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func (r *Renderer) drawDottedLine(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := 1
	if x1 >= x2 {
		sx = -1
	}
	sy := 1
	if y1 >= y2 {
		sy = -1
	}
	err := dx - dy
	step := 0

	for {
		// Draw every 4th pixel for dotted effect (dot on, 3 off)
		if step%4 == 0 {
			setPixelSafe(img, x1, y1, c)
		}
		step++

		if x1 == x2 && y1 == y2 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func (r *Renderer) drawDashedLine(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := 1
	if x1 >= x2 {
		sx = -1
	}
	sy := 1
	if y1 >= y2 {
		sy = -1
	}
	err := dx - dy
	step := 0

	for {
		// Draw 6 pixels on, 4 pixels off for dashed effect
		if step%10 < 6 {
			setPixelSafe(img, x1, y1, c)
		}
		step++

		if x1 == x2 && y1 == y2 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func (r *Renderer) drawTriangleUp(img *image.RGBA, cx, cy, size int, c color.RGBA) {
	halfSize := size / 2
	for row := 0; row < size; row++ {
		width := row
		startX := cx - width/2
		for dx := 0; dx <= width; dx++ {
			setPixelSafe(img, startX+dx, cy+halfSize-row, c)
		}
	}
}

func (r *Renderer) drawTriangleDown(img *image.RGBA, cx, cy, size int, c color.RGBA) {
	halfSize := size / 2
	for row := 0; row < size; row++ {
		width := row
		startX := cx - width/2
		for dx := 0; dx <= width; dx++ {
			setPixelSafe(img, startX+dx, cy-halfSize+row, c)
		}
	}
}

// drawFilledTriangleUp draws a filled triangle pointing up (apex at top)
func (r *Renderer) drawFilledTriangleUp(img *image.RGBA, cx, cy, size int, c color.RGBA) {
	// Triangle with apex at top, base at bottom
	// Row 0 is at top (apex), row size-1 is at bottom (widest)
	for row := 0; row < size; row++ {
		// Width increases as we go down
		width := row + 1
		startX := cx - row/2
		y := cy - size/2 + row
		for dx := 0; dx < width; dx++ {
			setPixelSafe(img, startX+dx, y, c)
		}
	}
}

// drawFilledTriangleDown draws a filled triangle pointing down (apex at bottom)
func (r *Renderer) drawFilledTriangleDown(img *image.RGBA, cx, cy, size int, c color.RGBA) {
	// Triangle with apex at bottom, base at top
	// Row 0 is at top (widest), row size-1 is at bottom (apex)
	for row := 0; row < size; row++ {
		// Width decreases as we go down
		width := size - row
		startX := cx - (size-row-1)/2
		y := cy - size/2 + row
		for dx := 0; dx < width; dx++ {
			setPixelSafe(img, startX+dx, y, c)
		}
	}
}

type fPoint struct {
	X float64
	Y float64
}

const (
	hatchDense     = "dense"
	hatchDiagCross = "diagcross"
)

func rgbaLightness(c color.RGBA) uint8 {
	// Approximate perceived lightness (0..255)
	return uint8((299*int(c.R) + 587*int(c.G) + 114*int(c.B)) / 1000)
}

func (r *Renderer) strokeTriangle(img *image.RGBA, a, b, c fPoint, col color.RGBA) {
	r.drawLine(img, int(math.Round(a.X)), int(math.Round(a.Y)), int(math.Round(b.X)), int(math.Round(b.Y)), col)
	r.drawLine(img, int(math.Round(b.X)), int(math.Round(b.Y)), int(math.Round(c.X)), int(math.Round(c.Y)), col)
	r.drawLine(img, int(math.Round(c.X)), int(math.Round(c.Y)), int(math.Round(a.X)), int(math.Round(a.Y)), col)
}

func (r *Renderer) fillTriangleHatch(img *image.RGBA, a, b, c fPoint, col color.RGBA, hatch string) {
	minX := int(math.Floor(min3(a.X, b.X, c.X)))
	maxX := int(math.Ceil(max3(a.X, b.X, c.X)))
	minY := int(math.Floor(min3(a.Y, b.Y, c.Y)))
	maxY := int(math.Ceil(max3(a.Y, b.Y, c.Y)))

	// Clamp to image bounds
	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX > img.Bounds().Max.X-1 {
		maxX = img.Bounds().Max.X - 1
	}
	if maxY > img.Bounds().Max.Y-1 {
		maxY = img.Bounds().Max.Y - 1
	}

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			px := float64(x) + 0.5
			py := float64(y) + 0.5
			if !pointInTriangle(px, py, a, b, c) {
				continue
			}

			// Hatch patterns: mimic Qt Dense4Pattern / DiagCrossPattern
			switch hatch {
			case hatchDiagCross:
				// two diagonals, wider spacing
				if ((x+y)%8 != 0) && ((x-y)%8 != 0) {
					continue
				}
			case hatchDense:
				// denser diagonal hatch
				if (x+y)%4 != 0 {
					continue
				}
			default:
				// solid fallback
			}

			setPixelSafe(img, x, y, col)
		}
	}
}

func min3(a, b, c float64) float64 {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func max3(a, b, c float64) float64 {
	if a > b {
		if a > c {
			return a
		}
		return c
	}
	if b > c {
		return b
	}
	return c
}

func pointInTriangle(px, py float64, a, b, c fPoint) bool {
	// Barycentric technique with sign checks
	b1 := sign(px, py, a, b) < 0.0
	b2 := sign(px, py, b, c) < 0.0
	b3 := sign(px, py, c, a) < 0.0
	return (b1 == b2) && (b2 == b3)
}

func sign(px, py float64, a, b fPoint) float64 {
	return (px-b.X)*(a.Y-b.Y) - (a.X-b.X)*(py-b.Y)
}

// drawTriangleUpOutline draws outline of triangle pointing up
func (r *Renderer) drawTriangleUpOutline(img *image.RGBA, cx, cy, size int, c color.RGBA) {
	halfSize := size / 2
	// Three points: apex at top, two corners at bottom
	topX, topY := cx, cy-halfSize
	leftX, leftY := cx-halfSize, cy+halfSize
	rightX, rightY := cx+halfSize, cy+halfSize

	r.drawLine(img, topX, topY, leftX, leftY, c)
	r.drawLine(img, topX, topY, rightX, rightY, c)
	r.drawLine(img, leftX, leftY, rightX, rightY, c)
}

// drawTriangleDownOutline draws outline of triangle pointing down
func (r *Renderer) drawTriangleDownOutline(img *image.RGBA, cx, cy, size int, c color.RGBA) {
	halfSize := size / 2
	// Three points: apex at bottom, two corners at top
	bottomX, bottomY := cx, cy+halfSize
	leftX, leftY := cx-halfSize, cy-halfSize
	rightX, rightY := cx+halfSize, cy-halfSize

	r.drawLine(img, bottomX, bottomY, leftX, leftY, c)
	r.drawLine(img, bottomX, bottomY, rightX, rightY, c)
	r.drawLine(img, leftX, leftY, rightX, rightY, c)
}

// Bitmap font for common characters (5x7 pixels)
var bitmapFont = map[rune][]uint8{
	'A': {0x0E, 0x11, 0x11, 0x1F, 0x11, 0x11, 0x11},
	'B': {0x1E, 0x11, 0x11, 0x1E, 0x11, 0x11, 0x1E},
	'C': {0x0E, 0x11, 0x10, 0x10, 0x10, 0x11, 0x0E},
	'D': {0x1C, 0x12, 0x11, 0x11, 0x11, 0x12, 0x1C},
	'E': {0x1F, 0x10, 0x10, 0x1E, 0x10, 0x10, 0x1F},
	'F': {0x1F, 0x10, 0x10, 0x1E, 0x10, 0x10, 0x10},
	'G': {0x0E, 0x11, 0x10, 0x17, 0x11, 0x11, 0x0F},
	'H': {0x11, 0x11, 0x11, 0x1F, 0x11, 0x11, 0x11},
	'I': {0x0E, 0x04, 0x04, 0x04, 0x04, 0x04, 0x0E},
	'J': {0x07, 0x02, 0x02, 0x02, 0x02, 0x12, 0x0C},
	'K': {0x11, 0x12, 0x14, 0x18, 0x14, 0x12, 0x11},
	'L': {0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x1F},
	'M': {0x11, 0x1B, 0x15, 0x15, 0x11, 0x11, 0x11},
	'N': {0x11, 0x11, 0x19, 0x15, 0x13, 0x11, 0x11},
	'O': {0x0E, 0x11, 0x11, 0x11, 0x11, 0x11, 0x0E},
	'P': {0x1E, 0x11, 0x11, 0x1E, 0x10, 0x10, 0x10},
	'Q': {0x0E, 0x11, 0x11, 0x11, 0x15, 0x12, 0x0D},
	'R': {0x1E, 0x11, 0x11, 0x1E, 0x14, 0x12, 0x11},
	'S': {0x0E, 0x11, 0x10, 0x0E, 0x01, 0x11, 0x0E},
	'T': {0x1F, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04},
	'U': {0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x0E},
	'V': {0x11, 0x11, 0x11, 0x11, 0x11, 0x0A, 0x04},
	'W': {0x11, 0x11, 0x11, 0x15, 0x15, 0x15, 0x0A},
	'X': {0x11, 0x11, 0x0A, 0x04, 0x0A, 0x11, 0x11},
	'Y': {0x11, 0x11, 0x0A, 0x04, 0x04, 0x04, 0x04},
	'Z': {0x1F, 0x01, 0x02, 0x04, 0x08, 0x10, 0x1F},
	'0': {0x0E, 0x11, 0x13, 0x15, 0x19, 0x11, 0x0E},
	'1': {0x04, 0x0C, 0x04, 0x04, 0x04, 0x04, 0x0E},
	'2': {0x0E, 0x11, 0x01, 0x0E, 0x10, 0x10, 0x1F},
	'3': {0x0E, 0x11, 0x01, 0x06, 0x01, 0x11, 0x0E},
	'4': {0x02, 0x06, 0x0A, 0x12, 0x1F, 0x02, 0x02},
	'5': {0x1F, 0x10, 0x1E, 0x01, 0x01, 0x11, 0x0E},
	'6': {0x06, 0x08, 0x10, 0x1E, 0x11, 0x11, 0x0E},
	'7': {0x1F, 0x01, 0x02, 0x04, 0x08, 0x08, 0x08},
	'8': {0x0E, 0x11, 0x11, 0x0E, 0x11, 0x11, 0x0E},
	'9': {0x0E, 0x11, 0x11, 0x0F, 0x01, 0x02, 0x0C},
}

// drawBitmapChar draws a character from bitmap font, returns true if character was found
func (r *Renderer) drawBitmapChar(img *image.RGBA, cx, cy int, ch rune, c color.RGBA) bool {
	// Convert lowercase to uppercase
	if ch >= 'a' && ch <= 'z' {
		ch = ch - 'a' + 'A'
	}

	bitmap, ok := bitmapFont[ch]
	if !ok {
		return false
	}

	// Font is 5x7, draw centered at cx, cy
	startX := cx - 2
	startY := cy - 3

	for row, rowData := range bitmap {
		for col := 0; col < 5; col++ {
			if (rowData & (0x10 >> col)) != 0 {
				setPixelSafe(img, startX+col, startY+row, c)
			}
		}
	}

	return true
}

// Helper functions

func setPixelSafe(img *image.RGBA, x, y int, c color.RGBA) {
	if x >= 0 && x < img.Bounds().Max.X && y >= 0 && y < img.Bounds().Max.Y {
		img.Set(x, y, c)
	}
}

func blendPixel(img *image.RGBA, x, y int, c color.RGBA) {
	if x < 0 || x >= img.Bounds().Max.X || y < 0 || y >= img.Bounds().Max.Y {
		return
	}
	if c.A == 255 {
		img.Set(x, y, c)
		return
	}

	existing := img.RGBAAt(x, y)
	alpha := float64(c.A) / 255.0
	invAlpha := 1.0 - alpha

	nr := uint8(float64(c.R)*alpha + float64(existing.R)*invAlpha)
	ng := uint8(float64(c.G)*alpha + float64(existing.G)*invAlpha)
	nb := uint8(float64(c.B)*alpha + float64(existing.B)*invAlpha)
	na := uint8(float64(c.A) + float64(existing.A)*invAlpha)

	img.Set(x, y, color.RGBA{R: nr, G: ng, B: nb, A: na})
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func abs32(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}

func min32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func max32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// drawLabels draws all labels for the given area and Z level
func (r *Renderer) drawLabels(img *image.RGBA, areaID, centerZ int32, showOnTop bool, centerX, centerY int32, halfWidth, halfHeight, spacing int) {
	labels := r.mapData.GetLabelsForArea(areaID)

	for _, lbl := range labels {
		// Filter by showOnTop
		if lbl.ShowOnTop != showOnTop {
			continue
		}

		// Filter by Z level (Mudlet logic: only show labels on current Z level)
		if int32(lbl.Pos.Z) != centerZ {
			continue
		}

		// Calculate position
		// lbl.Pos is in default map units (same as room coordinates, but float)
		// We calculate offset relative to map center room
		dx := lbl.Pos.X - float64(centerX)
		dy := lbl.Pos.Y - float64(centerY)

		// Calculate screen coordinates
		// Note: Y is flipped (up is negative Y on screen)
		screenX := halfWidth + int(dx*float64(spacing))
		screenY := halfHeight - int(dy*float64(spacing))

		// Calculate scaled size
		width := int(lbl.Width * float64(spacing))
		height := int(lbl.Height * float64(spacing))

		if width <= 0 || height <= 0 {
			continue
		}

		// Skip if completely off-screen
		if screenX+width < 0 || screenX > r.config.Width ||
			screenY+height < 0 || screenY > r.config.Height {
			continue
		}

		// Draw image if available
		if len(lbl.Pixmap) > 0 {
			// Decode PNG data
			lblImg, err := png.Decode(bytes.NewReader(lbl.Pixmap))
			if err == nil {
				destRect := image.Rect(screenX, screenY, screenX+width, screenY+height)

				if !lbl.NoScaling {
					// Scale to fit width/height
					r.drawScaled(img, destRect, lblImg)
				} else {
					// Draw unscaled at position
					// In Mudlet, NoScaling means it ignores lbl.Width/Height for rendering size,
					// and uses the original image size.
					bounds := lblImg.Bounds()
					targetRect := image.Rect(screenX, screenY, screenX+bounds.Dx(), screenY+bounds.Dy())
					draw.Draw(img, targetRect, lblImg, bounds.Min, draw.Over)
				}
			}
		}
		// TODO: Handle text-only labels if Pixmap is missing?
		// Mudlet usually includes rendered text in Pixmap.
	}
}

// drawScaled performs simple nearest-neighbor scaling of src to dst rect
func (r *Renderer) drawScaled(dst *image.RGBA, rect image.Rectangle, src image.Image) {
	if rect.Empty() {
		return
	}
	srcBounds := src.Bounds()
	sw := srcBounds.Dx()
	sh := srcBounds.Dy()
	if sw == 0 || sh == 0 {
		return
	}

	w := rect.Dx()
	h := rect.Dy()
	x0 := rect.Min.X
	y0 := rect.Min.Y

	// Clip against destination bounds
	if x0 < 0 {
		// Optimization needed but for now simple loop check
		// or advanced clipping logic
	}

	dstBounds := dst.Bounds()

	for y := 0; y < h; y++ {
		dy := y0 + y
		if dy < dstBounds.Min.Y || dy >= dstBounds.Max.Y {
			continue
		}

		sy := (y * sh) / h
		for x := 0; x < w; x++ {
			dx := x0 + x
			if dx < dstBounds.Min.X || dx >= dstBounds.Max.X {
				continue
			}

			sx := (x * sw) / w

			// Get source color
			c := src.At(srcBounds.Min.X+sx, srcBounds.Min.Y+sy)

			// Blend pixel
			blendPixel(dst, dx, dy, colorToRGBA(c))
		}
	}
}

// colorToRGBA converts any color.Color to color.RGBA
func colorToRGBA(c color.Color) color.RGBA {
	if rgba, ok := c.(color.RGBA); ok {
		return rgba
	}
	r, g, b, a := c.RGBA()
	return color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
}
