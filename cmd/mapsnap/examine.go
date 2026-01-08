package main

import (
	"fmt"
	"os"

	"github.com/szydell/arkadia-mapsnap/pkg/mapparser"
)

// examineContext holds state for examine operations
type examineContext struct {
	r     *mapparser.BinaryReader
	debug bool
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

	fmt.Printf("\nRooms section starts at offset %d\n", ctx.r.Position())
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
