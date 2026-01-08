package main

import (
	"fmt"
	"os"

	"github.com/szydell/arkadia-mapsnap/pkg/mapparser"
)

// examineContext holds state for examine operations
type examineContext struct {
	r       *mapparser.BinaryReader
	debug   bool
	version int32
}

// ExamineFile examines a binary map file and walks through its Qt/MudletMap structure.
// With debug=false, shows compact summary. With debug=true, shows detailed offsets and values.
func ExamineFile(filename string, debug bool) error {
	info, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}
	fmt.Printf("File size: %d bytes\n\n", info.Size())

	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	ctx := &examineContext{
		r:     mapparser.NewBinaryReader(f),
		debug: debug,
	}

	// version (qint32)
	ctx.logSection("MudletMap.version")
	version, err := ctx.r.ReadInt32()
	if err != nil {
		return fmt.Errorf("version: %w", err)
	}
	ctx.version = version
	fmt.Printf("  version = %d\n", version)

	// envColors: QMap<int,int>
	ctx.logSection("envColors QMap<int,int>")
	envColorsCount, err := ctx.readQMapIntInt()
	if err != nil {
		return fmt.Errorf("envColors: %w", err)
	}
	fmt.Printf("  count = %d\n", envColorsCount)

	// areaNames: QMap<int, QString>
	ctx.logSection("areaNames QMap<int,QString>")
	areaNames, err := ctx.readQMapIntQString()
	if err != nil {
		return fmt.Errorf("areaNames: %w", err)
	}
	fmt.Printf("  count = %d\n", len(areaNames))
	if debug {
		for i, entry := range areaNames {
			fmt.Printf("    [%d] id=%d name='%s'\n", i, entry.key, entry.value)
		}
	}

	// mCustomEnvColors: QMap<int,QColor>
	ctx.logSection("mCustomEnvColors QMap<int,QColor>")
	customEnvCount, err := ctx.readQMapIntQColor()
	if err != nil {
		return fmt.Errorf("mCustomEnvColors: %w", err)
	}
	fmt.Printf("  count = %d\n", customEnvCount)

	// mpRoomDbHashToRoomId: QMap<QString,QUInt>
	ctx.logSection("mpRoomDbHashToRoomId QMap<QString,uint>")
	roomDbHashCount, err := ctx.readQMapQStringUInt()
	if err != nil {
		return fmt.Errorf("mpRoomDbHashToRoomId: %w", err)
	}
	fmt.Printf("  count = %d\n", roomDbHashCount)

	// mUserData: QMap<QString,QString>
	ctx.logSection("mUserData QMap<QString,QString>")
	userDataCount, err := ctx.readQMapQStringQString()
	if err != nil {
		return fmt.Errorf("mUserData: %w", err)
	}
	fmt.Printf("  count = %d\n", userDataCount)

	// mapSymbolFont: QFont
	ctx.logSection("mapSymbolFont QFont")
	if err := ctx.skipQFont(); err != nil {
		return fmt.Errorf("mapSymbolFont: %w", err)
	}
	fmt.Printf("  (parsed)\n")

	// mapFontFudgeFactor: double
	ctx.logSection("mapFontFudgeFactor")
	fudge, err := ctx.r.ReadDouble()
	if err != nil {
		return fmt.Errorf("mapFontFudgeFactor: %w", err)
	}
	fmt.Printf("  value = %f\n", fudge)

	// useOnlyMapFont: bool
	ctx.logSection("useOnlyMapFont")
	useOnly, err := ctx.r.ReadBool()
	if err != nil {
		return fmt.Errorf("useOnlyMapFont: %w", err)
	}
	fmt.Printf("  value = %v\n", useOnly)

	// areas: MudletAreas
	ctx.logSection("areas MudletAreas")
	areasInfo, err := ctx.readMudletAreas()
	if err != nil {
		return fmt.Errorf("areas: %w", err)
	}
	fmt.Printf("  count = %d areas, total rooms = %d\n", areasInfo.count, areasInfo.totalRooms)
	if debug {
		for _, a := range areasInfo.areas {
			fmt.Printf("    area id=%d: rooms=%d, zLevels=%d, userData=%d\n",
				a.id, a.rooms, a.zLevels, a.userData)
		}
	}

	// mRoomIdHash: QMap<QString,QInt>
	ctx.logSection("mRoomIdHash QMap<QString,int>")
	roomIdHashCount, err := ctx.readQMapQStringInt()
	if err != nil {
		return fmt.Errorf("mRoomIdHash: %w", err)
	}
	fmt.Printf("  count = %d\n", roomIdHashCount)

	// labels: MudletLabels
	ctx.logSection("labels MudletLabels")
	labelsInfo, err := ctx.readMudletLabels()
	if err != nil {
		return fmt.Errorf("labels: %w", err)
	}
	fmt.Printf("  areas with labels = %d, total labels = %d\n", labelsInfo.areasCount, labelsInfo.totalLabels)
	if debug {
		for _, area := range labelsInfo.areas {
			fmt.Printf("    area id=%d: %d labels\n", area.areaID, len(area.labels))
			for j, lbl := range area.labels {
				fmt.Printf("      [%d] %s\n", j, lbl)
			}
		}
	}

	// rooms: MudletRooms (until end of file)
	ctx.logSection("rooms MudletRooms")
	roomsInfo, err := ctx.readMudletRooms()
	if err != nil {
		return fmt.Errorf("rooms: %w", err)
	}
	fmt.Printf("  total rooms = %d\n", roomsInfo.count)
	if debug && len(roomsInfo.rooms) > 0 {
		// Show first 5 rooms as sample
		limit := 5
		if len(roomsInfo.rooms) < limit {
			limit = len(roomsInfo.rooms)
		}
		fmt.Printf("  first %d rooms:\n", limit)
		for i := 0; i < limit; i++ {
			r := roomsInfo.rooms[i]
			fmt.Printf("    [%d] %s\n", i, r)
		}
		if len(roomsInfo.rooms) > limit {
			fmt.Printf("    ... and %d more rooms\n", len(roomsInfo.rooms)-limit)
		}
	}

	fmt.Printf("\nEnd of file at offset %d\n", ctx.r.Position())
	return nil
}

func (ctx *examineContext) logSection(name string) {
	if ctx.debug {
		fmt.Printf("\n@%d: %s\n", ctx.r.Position(), name)
	} else {
		fmt.Printf("%s:\n", name)
	}
}

// --- QMap readers ---

func (ctx *examineContext) readQMapIntInt() (int32, error) {
	sz, err := ctx.r.ReadInt32()
	if err != nil {
		return 0, err
	}
	for i := 0; i < int(sz); i++ {
		if _, err := ctx.r.ReadInt32(); err != nil {
			return 0, err
		}
		if _, err := ctx.r.ReadInt32(); err != nil {
			return 0, err
		}
	}
	return sz, nil
}

type intQStringEntry struct {
	key   int32
	value string
}

func (ctx *examineContext) readQMapIntQString() ([]intQStringEntry, error) {
	sz, err := ctx.r.ReadInt32()
	if err != nil {
		return nil, err
	}
	entries := make([]intQStringEntry, 0, sz)
	for i := 0; i < int(sz); i++ {
		key, err := ctx.r.ReadInt32()
		if err != nil {
			return nil, err
		}
		value, err := ctx.r.ReadQString()
		if err != nil {
			return nil, err
		}
		entries = append(entries, intQStringEntry{key: key, value: value})
	}
	return entries, nil
}

func (ctx *examineContext) readQMapIntQColor() (int32, error) {
	sz, err := ctx.r.ReadInt32()
	if err != nil {
		return 0, err
	}
	for i := 0; i < int(sz); i++ {
		if _, err := ctx.r.ReadInt32(); err != nil {
			return 0, err
		}
		if err := ctx.skipQColor(); err != nil {
			return 0, err
		}
	}
	return sz, nil
}

func (ctx *examineContext) readQMapQStringUInt() (int32, error) {
	sz, err := ctx.r.ReadInt32()
	if err != nil {
		return 0, err
	}
	for i := 0; i < int(sz); i++ {
		if _, err := ctx.r.ReadQString(); err != nil {
			return 0, err
		}
		if _, err := ctx.r.ReadUInt32(); err != nil {
			return 0, err
		}
	}
	return sz, nil
}

func (ctx *examineContext) readQMapQStringQString() (int32, error) {
	sz, err := ctx.r.ReadInt32()
	if err != nil {
		return 0, err
	}
	for i := 0; i < int(sz); i++ {
		if _, err := ctx.r.ReadQString(); err != nil {
			return 0, err
		}
		if _, err := ctx.r.ReadQString(); err != nil {
			return 0, err
		}
	}
	return sz, nil
}

func (ctx *examineContext) readQMapQStringInt() (int32, error) {
	sz, err := ctx.r.ReadInt32()
	if err != nil {
		return 0, err
	}
	for i := 0; i < int(sz); i++ {
		if _, err := ctx.r.ReadQString(); err != nil {
			return 0, err
		}
		if _, err := ctx.r.ReadInt32(); err != nil {
			return 0, err
		}
	}
	return sz, nil
}

func (ctx *examineContext) skipQMapIntInt() (int32, error) {
	sz, err := ctx.r.ReadInt32()
	if err != nil {
		return 0, err
	}
	for i := 0; i < int(sz); i++ {
		if _, err := ctx.r.ReadInt32(); err != nil {
			return 0, err
		}
		if _, err := ctx.r.ReadInt32(); err != nil {
			return 0, err
		}
	}
	return sz, nil
}

// --- Qt type readers ---

func (ctx *examineContext) skipQColor() error {
	if _, err := ctx.r.ReadInt8(); err != nil {
		return err
	}
	for i := 0; i < 5; i++ {
		if _, err := ctx.r.ReadUInt16(); err != nil {
			return err
		}
	}
	return nil
}

func (ctx *examineContext) skipQFont() error {
	if _, err := ctx.r.ReadQString(); err != nil {
		return err
	}
	if _, err := ctx.r.ReadQString(); err != nil {
		return err
	}
	if _, err := ctx.r.ReadDouble(); err != nil {
		return err
	}
	if _, err := ctx.r.ReadInt32(); err != nil {
		return err
	}
	if _, err := ctx.r.ReadInt8(); err != nil {
		return err
	}
	if _, err := ctx.r.ReadUInt16(); err != nil {
		return err
	}
	if _, err := ctx.r.ReadByte(); err != nil {
		return err
	}
	if _, err := ctx.r.ReadInt8(); err != nil {
		return err
	}
	if _, err := ctx.r.ReadInt8(); err != nil {
		return err
	}
	if _, err := ctx.r.ReadUInt16(); err != nil {
		return err
	}
	if _, err := ctx.r.ReadInt8(); err != nil {
		return err
	}
	if _, err := ctx.r.ReadInt32(); err != nil {
		return err
	}
	if _, err := ctx.r.ReadInt32(); err != nil {
		return err
	}
	if _, err := ctx.r.ReadInt8(); err != nil {
		return err
	}
	if _, err := ctx.r.ReadInt8(); err != nil {
		return err
	}
	return nil
}

func (ctx *examineContext) skipQVector3D() error {
	for i := 0; i < 3; i++ {
		if _, err := ctx.r.ReadDouble(); err != nil {
			return err
		}
	}
	return nil
}

// --- MudletAreas ---

type areaInfo struct {
	id       int32
	rooms    int32
	zLevels  int32
	userData int32
}

type areasResult struct {
	count      int32
	totalRooms int32
	areas      []areaInfo
}

func (ctx *examineContext) readMudletAreas() (*areasResult, error) {
	sz, err := ctx.r.ReadInt32()
	if err != nil {
		return nil, err
	}

	result := &areasResult{
		count: sz,
		areas: make([]areaInfo, 0, sz),
	}

	for i := 0; i < int(sz); i++ {
		areaID, err := ctx.r.ReadInt32()
		if err != nil {
			return nil, err
		}

		info, err := ctx.readMudletArea()
		if err != nil {
			return nil, err
		}
		info.id = areaID

		result.totalRooms += info.rooms
		result.areas = append(result.areas, *info)
	}

	return result, nil
}

func (ctx *examineContext) readMudletArea() (*areaInfo, error) {
	info := &areaInfo{}

	// rooms: QList<quint32>
	roomCount, err := ctx.r.ReadInt32()
	if err != nil {
		return nil, err
	}
	info.rooms = roomCount
	for i := 0; i < int(roomCount); i++ {
		if _, err := ctx.r.ReadUInt32(); err != nil {
			return nil, err
		}
	}

	// zLevels: QList<int>
	zLevelCount, err := ctx.r.ReadInt32()
	if err != nil {
		return nil, err
	}
	info.zLevels = zLevelCount
	for i := 0; i < int(zLevelCount); i++ {
		if _, err := ctx.r.ReadInt32(); err != nil {
			return nil, err
		}
	}

	// coordinates: count + 3 ints per entry
	coordCount, err := ctx.r.ReadInt32()
	if err != nil {
		return nil, err
	}
	for i := 0; i < int(coordCount); i++ {
		if _, err := ctx.r.ReadInt32(); err != nil {
			return nil, err
		}
		if _, err := ctx.r.ReadInt32(); err != nil {
			return nil, err
		}
		if _, err := ctx.r.ReadInt32(); err != nil {
			return nil, err
		}
	}

	// gridMode: bool
	if _, err := ctx.r.ReadBool(); err != nil {
		return nil, err
	}

	// bounds: 6 x int32
	for i := 0; i < 6; i++ {
		if _, err := ctx.r.ReadInt32(); err != nil {
			return nil, err
		}
	}

	// span: QVector3D
	if err := ctx.skipQVector3D(); err != nil {
		return nil, err
	}

	// 4 grid maps
	for i := 0; i < 4; i++ {
		if _, err := ctx.skipQMapIntInt(); err != nil {
			return nil, err
		}
	}

	// offset: QVector3D
	if err := ctx.skipQVector3D(); err != nil {
		return nil, err
	}

	// isZLocked: bool
	if _, err := ctx.r.ReadBool(); err != nil {
		return nil, err
	}

	// min_z: int32
	if _, err := ctx.r.ReadInt32(); err != nil {
		return nil, err
	}

	// userData: QMap<QString,QString>
	userDataCount, err := ctx.r.ReadInt32()
	if err != nil {
		return nil, err
	}
	info.userData = userDataCount
	for i := 0; i < int(userDataCount); i++ {
		if _, err := ctx.r.ReadQString(); err != nil {
			return nil, err
		}
		if _, err := ctx.r.ReadQString(); err != nil {
			return nil, err
		}
	}

	return info, nil
}

// --- MudletLabels ---

type labelAreaInfo struct {
	areaID int32
	labels []string
}

type labelsResult struct {
	areasCount  int32
	totalLabels int
	areas       []labelAreaInfo
}

func (ctx *examineContext) readMudletLabels() (*labelsResult, error) {
	sz, err := ctx.r.ReadInt32()
	if err != nil {
		return nil, err
	}

	result := &labelsResult{
		areasCount: sz,
		areas:      make([]labelAreaInfo, 0, sz),
	}

	for i := 0; i < int(sz); i++ {
		total, err := ctx.r.ReadInt32()
		if err != nil {
			return nil, err
		}
		areaID, err := ctx.r.ReadInt32()
		if err != nil {
			return nil, err
		}

		areaInfo := labelAreaInfo{
			areaID: areaID,
			labels: make([]string, 0, total),
		}

		for j := 0; j < int(total); j++ {
			summary, err := ctx.readMudletLabel()
			if err != nil {
				return nil, err
			}
			areaInfo.labels = append(areaInfo.labels, summary)
			result.totalLabels++
		}

		result.areas = append(result.areas, areaInfo)
	}

	return result, nil
}

func (ctx *examineContext) readMudletLabel() (string, error) {
	// id: int32
	labelID, err := ctx.r.ReadInt32()
	if err != nil {
		return "", err
	}

	// pos: QVector3D (x, y, z)
	posX, err := ctx.r.ReadDouble()
	if err != nil {
		return "", err
	}
	posY, err := ctx.r.ReadDouble()
	if err != nil {
		return "", err
	}
	posZ, err := ctx.r.ReadDouble()
	if err != nil {
		return "", err
	}

	// dummy1, dummy2
	for i := 0; i < 2; i++ {
		if _, err := ctx.r.ReadDouble(); err != nil {
			return "", err
		}
	}

	// size: QPair<double,double>
	sizeW, err := ctx.r.ReadDouble()
	if err != nil {
		return "", err
	}
	sizeH, err := ctx.r.ReadDouble()
	if err != nil {
		return "", err
	}

	// text: QString
	text, err := ctx.r.ReadQString()
	if err != nil {
		return "", err
	}

	// fgColor, bgColor
	if err := ctx.skipQColor(); err != nil {
		return "", err
	}
	if err := ctx.skipQColor(); err != nil {
		return "", err
	}

	// QPixMap
	hasPNG := false
	pngSize := 0
	startPos := ctx.r.Position()
	_, _ = ctx.r.ReadUInt32()
	if sig, _ := ctx.r.Peek(4); len(sig) == 4 {
		if uint32(sig[0])<<24|uint32(sig[1])<<16|uint32(sig[2])<<8|uint32(sig[3]) == 0x89504e47 {
			hasPNG = true
			if err := ctx.skipPNG(); err != nil {
				return "", err
			}
			pngSize = int(ctx.r.Position() - startPos)
		}
	}

	// noScaling, showOnTop
	noScaling, err := ctx.r.ReadBool()
	if err != nil {
		return "", err
	}
	showOnTop, err := ctx.r.ReadBool()
	if err != nil {
		return "", err
	}

	// Build summary
	summary := fmt.Sprintf("id=%d pos=(%.0f,%.0f,%.0f) size=(%.0f,%.0f)",
		labelID, posX, posY, posZ, sizeW, sizeH)
	if text != "" {
		summary += fmt.Sprintf(" text='%s'", text)
	}
	if hasPNG {
		summary += fmt.Sprintf(" PNG=%d bytes", pngSize)
	}
	if noScaling {
		summary += " noScaling"
	}
	if showOnTop {
		summary += " showOnTop"
	}

	return summary, nil
}

func (ctx *examineContext) skipPNG() error {
	needle := []byte{0x49, 0x45, 0x4e, 0x44} // 'I','E','N','D'
	for {
		peek, err := ctx.r.Peek(4)
		if err != nil || len(peek) < 4 {
			return err
		}
		if peek[0] == needle[0] && peek[1] == needle[1] && peek[2] == needle[2] && peek[3] == needle[3] {
			if err := ctx.r.Skip(8); err != nil {
				return err
			}
			return nil
		}
		if _, err := ctx.r.ReadByte(); err != nil {
			return err
		}
	}
}

// --- MudletRooms ---

type roomsResult struct {
	count int
	rooms []string // room summaries
}

func (ctx *examineContext) readMudletRooms() (*roomsResult, error) {
	result := &roomsResult{
		rooms: make([]string, 0),
	}

	for {
		// Check if we're at EOF
		peek, err := ctx.r.Peek(4)
		if err != nil || len(peek) < 4 {
			break
		}

		// Read room ID
		roomID, err := ctx.r.ReadInt32()
		if err != nil {
			break
		}

		// Read room data
		summary, err := ctx.readMudletRoom(roomID)
		if err != nil {
			return nil, fmt.Errorf("room %d: %w", roomID, err)
		}

		result.rooms = append(result.rooms, summary)
		result.count++
	}

	return result, nil
}

func (ctx *examineContext) readMudletRoom(roomID int32) (string, error) {
	// area: int32
	area, err := ctx.r.ReadInt32()
	if err != nil {
		return "", err
	}

	// x, y, z: int32
	x, err := ctx.r.ReadInt32()
	if err != nil {
		return "", err
	}
	y, err := ctx.r.ReadInt32()
	if err != nil {
		return "", err
	}
	z, err := ctx.r.ReadInt32()
	if err != nil {
		return "", err
	}

	// 12 standard exits: north, northeast, east, southeast, south, southwest, west, northwest, up, down, in, out
	exits := make([]int32, 12)
	exitNames := []string{"n", "ne", "e", "se", "s", "sw", "w", "nw", "up", "down", "in", "out"}
	for i := 0; i < 12; i++ {
		exits[i], err = ctx.r.ReadInt32()
		if err != nil {
			return "", err
		}
	}

	// environment: int32
	environment, err := ctx.r.ReadInt32()
	if err != nil {
		return "", err
	}

	// weight: int32
	weight, err := ctx.r.ReadInt32()
	if err != nil {
		return "", err
	}

	// name: QString
	name, err := ctx.r.ReadQString()
	if err != nil {
		return "", err
	}

	// isLocked: bool
	isLocked, err := ctx.r.ReadBool()
	if err != nil {
		return "", err
	}

	// mSpecialExits: depends on version
	var specialExitCount int32
	if ctx.version >= 21 {
		// Version 21+: QMultiMap<QString, int>
		specialExitCount, err = ctx.r.ReadInt32()
		if err != nil {
			return "", err
		}
		for i := 0; i < int(specialExitCount); i++ {
			if _, err := ctx.r.ReadQString(); err != nil {
				return "", err
			}
			if _, err := ctx.r.ReadInt32(); err != nil {
				return "", err
			}
		}
	} else if ctx.version >= 6 {
		// Version 6-20: QMultiMap<int, QString> (key=dest room, value=command with lock prefix)
		specialExitCount, err = ctx.r.ReadInt32()
		if err != nil {
			return "", err
		}
		for i := 0; i < int(specialExitCount); i++ {
			if _, err := ctx.r.ReadInt32(); err != nil { // destination room ID
				return "", err
			}
			if _, err := ctx.r.ReadQString(); err != nil { // command with "0"/"1" prefix
				return "", err
			}
		}
	}

	// symbol: QString (version >= 19)
	var symbol string
	if ctx.version >= 19 {
		symbol, err = ctx.r.ReadQString()
		if err != nil {
			return "", err
		}
	} else if ctx.version >= 9 {
		// old format: qint8
		_, err = ctx.r.ReadByte()
		if err != nil {
			return "", err
		}
	}

	// symbolColor: QColor (version >= 21 only)
	if ctx.version >= 21 {
		if err := ctx.skipQColor(); err != nil {
			return "", err
		}
	}

	// userData: QMap<QString, QString> (version >= 10)
	if ctx.version >= 10 {
		userDataCount, err := ctx.r.ReadInt32()
		if err != nil {
			return "", err
		}
		for i := 0; i < int(userDataCount); i++ {
			if _, err := ctx.r.ReadQString(); err != nil {
				return "", err
			}
			if _, err := ctx.r.ReadQString(); err != nil {
				return "", err
			}
		}
	}

	// customLines and related fields (version >= 11)
	if ctx.version >= 11 {
		if ctx.version >= 20 {
			// Version 20+: QMap<QString, QList<QPointF>>
			customLinesCount, err := ctx.r.ReadInt32()
			if err != nil {
				return "", err
			}
			for i := 0; i < int(customLinesCount); i++ {
				if _, err := ctx.r.ReadQString(); err != nil {
					return "", err
				}
				// QList<QPointF>
				pointCount, err := ctx.r.ReadInt32()
				if err != nil {
					return "", err
				}
				for j := 0; j < int(pointCount); j++ {
					if _, err := ctx.r.ReadDouble(); err != nil { // x
						return "", err
					}
					if _, err := ctx.r.ReadDouble(); err != nil { // y
						return "", err
					}
				}
			}

			// customLinesArrow: QMap<QString, bool>
			customLinesArrowCount, err := ctx.r.ReadInt32()
			if err != nil {
				return "", err
			}
			for i := 0; i < int(customLinesArrowCount); i++ {
				if _, err := ctx.r.ReadQString(); err != nil {
					return "", err
				}
				if _, err := ctx.r.ReadBool(); err != nil {
					return "", err
				}
			}

			// customLinesColor: QMap<QString, QColor> (version 20+)
			customLinesColorCount, err := ctx.r.ReadInt32()
			if err != nil {
				return "", err
			}
			for i := 0; i < int(customLinesColorCount); i++ {
				if _, err := ctx.r.ReadQString(); err != nil {
					return "", err
				}
				if err := ctx.skipQColor(); err != nil {
					return "", err
				}
			}

			// customLinesStyle: QMap<QString, Qt::PenStyle(int)> (version 20+)
			customLinesStyleCount, err := ctx.r.ReadInt32()
			if err != nil {
				return "", err
			}
			for i := 0; i < int(customLinesStyleCount); i++ {
				if _, err := ctx.r.ReadQString(); err != nil {
					return "", err
				}
				if _, err := ctx.r.ReadInt32(); err != nil {
					return "", err
				}
			}
		} else {
			// Version 11-19: old format
			// customLines: QMap<QString, QList<QPointF>>
			customLinesCount, err := ctx.r.ReadInt32()
			if err != nil {
				return "", err
			}
			for i := 0; i < int(customLinesCount); i++ {
				if _, err := ctx.r.ReadQString(); err != nil {
					return "", err
				}
				pointCount, err := ctx.r.ReadInt32()
				if err != nil {
					return "", err
				}
				for j := 0; j < int(pointCount); j++ {
					if _, err := ctx.r.ReadDouble(); err != nil {
						return "", err
					}
					if _, err := ctx.r.ReadDouble(); err != nil {
						return "", err
					}
				}
			}

			// customLinesArrow: QMap<QString, bool>
			customLinesArrowCount, err := ctx.r.ReadInt32()
			if err != nil {
				return "", err
			}
			for i := 0; i < int(customLinesArrowCount); i++ {
				if _, err := ctx.r.ReadQString(); err != nil {
					return "", err
				}
				if _, err := ctx.r.ReadBool(); err != nil {
					return "", err
				}
			}

			// customLinesColor: QMap<QString, QList<int>> (3 ints for RGB)
			customLinesColorCount, err := ctx.r.ReadInt32()
			if err != nil {
				return "", err
			}
			for i := 0; i < int(customLinesColorCount); i++ {
				if _, err := ctx.r.ReadQString(); err != nil {
					return "", err
				}
				// QList<int>
				rgbCount, err := ctx.r.ReadInt32()
				if err != nil {
					return "", err
				}
				for j := 0; j < int(rgbCount); j++ {
					if _, err := ctx.r.ReadInt32(); err != nil {
						return "", err
					}
				}
			}

			// customLinesStyle: QMap<QString, QString>
			customLinesStyleCount, err := ctx.r.ReadInt32()
			if err != nil {
				return "", err
			}
			for i := 0; i < int(customLinesStyleCount); i++ {
				if _, err := ctx.r.ReadQString(); err != nil {
					return "", err
				}
				if _, err := ctx.r.ReadQString(); err != nil {
					return "", err
				}
			}
		}

		// mSpecialExitLocks: QSet<QString> (version >= 21 only)
		if ctx.version >= 21 {
			specialExitLocksCount, err := ctx.r.ReadInt32()
			if err != nil {
				return "", err
			}
			for i := 0; i < int(specialExitLocksCount); i++ {
				if _, err := ctx.r.ReadQString(); err != nil {
					return "", err
				}
			}
		}

		// exitLocks: QList<int>
		exitLocksCount, err := ctx.r.ReadInt32()
		if err != nil {
			return "", err
		}
		for i := 0; i < int(exitLocksCount); i++ {
			if _, err := ctx.r.ReadInt32(); err != nil {
				return "", err
			}
		}
	}

	// exitStubs: QList<int> (version >= 13)
	if ctx.version >= 13 {
		exitStubsCount, err := ctx.r.ReadInt32()
		if err != nil {
			return "", err
		}
		for i := 0; i < int(exitStubsCount); i++ {
			if _, err := ctx.r.ReadInt32(); err != nil {
				return "", err
			}
		}
	}

	// exitWeights and doors (version >= 16)
	if ctx.version >= 16 {
		// exitWeights: QMap<QString, int>
		exitWeightsCount, err := ctx.r.ReadInt32()
		if err != nil {
			return "", err
		}
		for i := 0; i < int(exitWeightsCount); i++ {
			if _, err := ctx.r.ReadQString(); err != nil {
				return "", err
			}
			if _, err := ctx.r.ReadInt32(); err != nil {
				return "", err
			}
		}

		// doors: QMap<QString, int>
		doorsCount, err := ctx.r.ReadInt32()
		if err != nil {
			return "", err
		}
		for i := 0; i < int(doorsCount); i++ {
			if _, err := ctx.r.ReadQString(); err != nil {
				return "", err
			}
			if _, err := ctx.r.ReadInt32(); err != nil {
				return "", err
			}
		}
	}

	// Build summary
	summary := fmt.Sprintf("id=%d area=%d pos=(%d,%d,%d)", roomID, area, x, y, z)

	// Add exits
	var activeExits []string
	for i, exitID := range exits {
		if exitID != -1 {
			activeExits = append(activeExits, fmt.Sprintf("%s:%d", exitNames[i], exitID))
		}
	}
	if len(activeExits) > 0 {
		summary += fmt.Sprintf(" exits=[%s]", joinStrings(activeExits, ","))
	}

	if name != "" {
		summary += fmt.Sprintf(" name='%s'", name)
	}
	if symbol != "" {
		summary += fmt.Sprintf(" symbol='%s'", symbol)
	}
	if environment != -1 {
		summary += fmt.Sprintf(" env=%d", environment)
	}
	if weight != 1 {
		summary += fmt.Sprintf(" weight=%d", weight)
	}
	if isLocked {
		summary += " locked"
	}
	if specialExitCount > 0 {
		summary += fmt.Sprintf(" specialExits=%d", specialExitCount)
	}

	return summary, nil
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
