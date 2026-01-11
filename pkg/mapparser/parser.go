package mapparser

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// ParseMapFile parses a Mudlet map file and returns a [MudletMap] structure.
//
// This is the primary entry point for parsing map files. It opens the file,
// parses its contents, and properly closes the file handle.
//
// Example:
//
//	m, err := mapparser.ParseMapFile("world.map")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Loaded %d rooms\n", m.RoomCount())
func ParseMapFile(filename string) (m *MudletMap, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening map file: %w", err)
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			if err != nil {
				err = errors.Join(err, closeErr)
			} else {
				err = fmt.Errorf("closing map file: %w", closeErr)
			}
		}
	}()

	return ParseMap(file)
}

// ParseMap parses a Mudlet map from an [io.Reader].
//
// Use this function when you have an already-open reader, such as an embedded
// file or network stream. For parsing files, prefer [ParseMapFile].
func ParseMap(reader io.Reader) (*MudletMap, error) {
	p := &parser{
		r: NewBinaryReader(reader),
		m: NewMudletMap(),
	}

	if err := p.parse(); err != nil {
		return nil, err
	}

	return p.m, nil
}

// parser holds internal state for map parsing operations.
type parser struct {
	r *BinaryReader
	m *MudletMap
}

// parse processes the entire map file structure.
func (p *parser) parse() error {
	// version (qint32)
	version, err := p.r.ReadInt32()
	if err != nil {
		return fmt.Errorf("version: %w", err)
	}
	p.m.Version = version

	// envColors: QMap<int,int>
	if err := p.readEnvColors(); err != nil {
		return fmt.Errorf("envColors: %w", err)
	}

	// areaNames: QMap<int, QString>
	if err := p.readAreaNames(); err != nil {
		return fmt.Errorf("areaNames: %w", err)
	}

	// mCustomEnvColors: QMap<int,QColor>
	if err := p.readCustomEnvColors(); err != nil {
		return fmt.Errorf("mCustomEnvColors: %w", err)
	}

	// mpRoomDbHashToRoomId: QMap<QString,uint>
	if err := p.readRoomDbHashToRoomId(); err != nil {
		return fmt.Errorf("mpRoomDbHashToRoomId: %w", err)
	}

	// mUserData: QMap<QString,QString>
	if err := p.readUserData(); err != nil {
		return fmt.Errorf("mUserData: %w", err)
	}

	// mapSymbolFont: QFont
	font, err := p.readQFont()
	if err != nil {
		return fmt.Errorf("mapSymbolFont: %w", err)
	}
	p.m.MapSymbolFont = font

	// mapFontFudgeFactor: double
	fudge, err := p.r.ReadDouble()
	if err != nil {
		return fmt.Errorf("mapFontFudgeFactor: %w", err)
	}
	p.m.MapFontFudgeFactor = fudge

	// useOnlyMapFont: bool
	useOnly, err := p.r.ReadBool()
	if err != nil {
		return fmt.Errorf("useOnlyMapFont: %w", err)
	}
	p.m.UseOnlyMapFont = useOnly

	// areas: MudletAreas
	if err := p.readAreas(); err != nil {
		return fmt.Errorf("areas: %w", err)
	}

	// mRoomIdHash: QMap<QString,int>
	if err := p.readRoomIdHash(); err != nil {
		return fmt.Errorf("mRoomIdHash: %w", err)
	}

	// labels: MudletLabels (version < 21)
	if err := p.readLabels(); err != nil {
		return fmt.Errorf("labels: %w", err)
	}

	// rooms: MudletRooms (until end of file)
	if err := p.readRooms(); err != nil {
		return fmt.Errorf("rooms: %w", err)
	}

	return nil
}

// --- Map-level field readers ---

func (p *parser) readEnvColors() error {
	count, err := p.r.ReadInt32()
	if err != nil {
		return err
	}
	for i := int32(0); i < count; i++ {
		key, err := p.r.ReadInt32()
		if err != nil {
			return err
		}
		value, err := p.r.ReadInt32()
		if err != nil {
			return err
		}
		p.m.EnvColors[key] = value
	}
	return nil
}

func (p *parser) readAreaNames() error {
	count, err := p.r.ReadInt32()
	if err != nil {
		return err
	}
	for i := int32(0); i < count; i++ {
		id, err := p.r.ReadInt32()
		if err != nil {
			return err
		}
		name, err := p.r.ReadQString()
		if err != nil {
			return err
		}
		p.m.Areas[id] = NewMudletArea(id, name)
	}
	return nil
}

func (p *parser) readCustomEnvColors() error {
	count, err := p.r.ReadInt32()
	if err != nil {
		return err
	}
	for i := int32(0); i < count; i++ {
		key, err := p.r.ReadInt32()
		if err != nil {
			return err
		}
		color, err := p.readQColor()
		if err != nil {
			return err
		}
		p.m.CustomEnvColors[key] = color
	}
	return nil
}

func (p *parser) readRoomDbHashToRoomId() error {
	count, err := p.r.ReadInt32()
	if err != nil {
		return err
	}
	for i := int32(0); i < count; i++ {
		key, err := p.r.ReadQString()
		if err != nil {
			return err
		}
		value, err := p.r.ReadUInt32()
		if err != nil {
			return err
		}
		p.m.RoomDbHashToRoomId[key] = value
	}
	return nil
}

func (p *parser) readUserData() error {
	count, err := p.r.ReadInt32()
	if err != nil {
		return err
	}
	for i := int32(0); i < count; i++ {
		key, err := p.r.ReadQString()
		if err != nil {
			return err
		}
		value, err := p.r.ReadQString()
		if err != nil {
			return err
		}
		p.m.UserData[key] = value
	}
	return nil
}

func (p *parser) readRoomIdHash() error {
	count, err := p.r.ReadInt32()
	if err != nil {
		return err
	}
	for i := int32(0); i < count; i++ {
		key, err := p.r.ReadQString()
		if err != nil {
			return err
		}
		value, err := p.r.ReadInt32()
		if err != nil {
			return err
		}
		p.m.RoomIdHash[key] = value
	}
	return nil
}

// --- Qt type readers ---

func (p *parser) readQColor() (Color, error) {
	var c Color
	spec, err := p.r.ReadInt8()
	if err != nil {
		return c, err
	}
	c.Spec = spec

	c.Alpha, err = p.r.ReadUInt16()
	if err != nil {
		return c, err
	}
	c.Red, err = p.r.ReadUInt16()
	if err != nil {
		return c, err
	}
	c.Green, err = p.r.ReadUInt16()
	if err != nil {
		return c, err
	}
	c.Blue, err = p.r.ReadUInt16()
	if err != nil {
		return c, err
	}
	c.Pad, err = p.r.ReadUInt16()
	if err != nil {
		return c, err
	}
	return c, nil
}

func (p *parser) readQFont() (Font, error) {
	var f Font
	var err error

	f.Family, err = p.r.ReadQString()
	if err != nil {
		return f, err
	}
	f.StyleHint, err = p.r.ReadQString()
	if err != nil {
		return f, err
	}
	f.PointSizeF, err = p.r.ReadDouble()
	if err != nil {
		return f, err
	}
	f.PixelSize, err = p.r.ReadInt32()
	if err != nil {
		return f, err
	}
	f.StyleStrategy, err = p.r.ReadInt8()
	if err != nil {
		return f, err
	}
	f.Weight, err = p.r.ReadUInt16()
	if err != nil {
		return f, err
	}
	style, err := p.r.ReadByte()
	if err != nil {
		return f, err
	}
	f.Style = style

	underline, err := p.r.ReadInt8()
	if err != nil {
		return f, err
	}
	f.Underline = underline != 0

	strikeOut, err := p.r.ReadInt8()
	if err != nil {
		return f, err
	}
	f.StrikeOut = strikeOut != 0

	// Skip fixedPitch (uint16 in this version)
	_, err = p.r.ReadUInt16()
	if err != nil {
		return f, err
	}

	f.Capitalization, err = p.r.ReadInt8()
	if err != nil {
		return f, err
	}
	f.LetterSpacing, err = p.r.ReadInt32()
	if err != nil {
		return f, err
	}
	f.WordSpacing, err = p.r.ReadInt32()
	if err != nil {
		return f, err
	}
	f.Stretch, err = p.r.ReadInt8()
	if err != nil {
		return f, err
	}
	f.HintingPreference, err = p.r.ReadInt8()
	if err != nil {
		return f, err
	}

	return f, nil
}

func (p *parser) readQVector3D() (Vector3D, error) {
	var v Vector3D
	var err error
	v.X, err = p.r.ReadDouble()
	if err != nil {
		return v, err
	}
	v.Y, err = p.r.ReadDouble()
	if err != nil {
		return v, err
	}
	v.Z, err = p.r.ReadDouble()
	if err != nil {
		return v, err
	}
	return v, nil
}

func (p *parser) readQMapIntInt() (map[int32]int32, error) {
	count, err := p.r.ReadInt32()
	if err != nil {
		return nil, err
	}
	result := make(map[int32]int32, count)
	for i := int32(0); i < count; i++ {
		key, err := p.r.ReadInt32()
		if err != nil {
			return nil, err
		}
		value, err := p.r.ReadInt32()
		if err != nil {
			return nil, err
		}
		result[key] = value
	}
	return result, nil
}

// --- Area readers ---

func (p *parser) readAreas() error {
	count, err := p.r.ReadInt32()
	if err != nil {
		return err
	}

	for i := int32(0); i < count; i++ {
		areaID, err := p.r.ReadInt32()
		if err != nil {
			return err
		}

		area := p.m.Areas[areaID]
		if area == nil {
			area = NewMudletArea(areaID, "")
			p.m.Areas[areaID] = area
		}

		if err := p.readAreaData(area); err != nil {
			return fmt.Errorf("area %d: %w", areaID, err)
		}
	}

	return nil
}

func (p *parser) readAreaData(area *MudletArea) error {
	var err error

	// rooms: QSet<quint32>
	roomCount, err := p.r.ReadInt32()
	if err != nil {
		return err
	}
	area.Rooms = make([]uint32, 0, roomCount)
	for i := int32(0); i < roomCount; i++ {
		roomID, err := p.r.ReadUInt32()
		if err != nil {
			return err
		}
		area.Rooms = append(area.Rooms, roomID)
	}

	// zLevels: QList<int>
	zLevelCount, err := p.r.ReadInt32()
	if err != nil {
		return err
	}
	area.ZLevels = make([]int32, 0, zLevelCount)
	for i := int32(0); i < zLevelCount; i++ {
		z, err := p.r.ReadInt32()
		if err != nil {
			return err
		}
		area.ZLevels = append(area.ZLevels, z)
	}

	// mAreaExits: QMultiMap<int, QPair<int, int>>
	areaExitsCount, err := p.r.ReadInt32()
	if err != nil {
		return err
	}
	area.AreaExits = make([]AreaExit, 0, areaExitsCount)
	for i := int32(0); i < areaExitsCount; i++ {
		roomID, err := p.r.ReadInt32()
		if err != nil {
			return err
		}
		destRoomID, err := p.r.ReadInt32()
		if err != nil {
			return err
		}
		direction, err := p.r.ReadInt32()
		if err != nil {
			return err
		}
		area.AreaExits = append(area.AreaExits, AreaExit{
			RoomID:     roomID,
			DestRoomID: destRoomID,
			Direction:  direction,
		})
	}

	// gridMode: bool
	area.GridMode, err = p.r.ReadBool()
	if err != nil {
		return err
	}

	// bounds: max_x, max_y, max_z, min_x, min_y, min_z
	area.Bounds.MaxX, err = p.r.ReadInt32()
	if err != nil {
		return err
	}
	area.Bounds.MaxY, err = p.r.ReadInt32()
	if err != nil {
		return err
	}
	area.Bounds.MaxZ, err = p.r.ReadInt32()
	if err != nil {
		return err
	}
	area.Bounds.MinX, err = p.r.ReadInt32()
	if err != nil {
		return err
	}
	area.Bounds.MinY, err = p.r.ReadInt32()
	if err != nil {
		return err
	}
	area.Bounds.MinZ, err = p.r.ReadInt32()
	if err != nil {
		return err
	}

	// span: QVector3D
	area.Span, err = p.readQVector3D()
	if err != nil {
		return err
	}

	// xmaxForZ, ymaxForZ, xminForZ, yminForZ: 4 x QMap<int,int>
	area.XMaxForZ, err = p.readQMapIntInt()
	if err != nil {
		return err
	}
	area.YMaxForZ, err = p.readQMapIntInt()
	if err != nil {
		return err
	}
	area.XMinForZ, err = p.readQMapIntInt()
	if err != nil {
		return err
	}
	area.YMinForZ, err = p.readQMapIntInt()
	if err != nil {
		return err
	}

	// pos: QVector3D
	area.Pos, err = p.readQVector3D()
	if err != nil {
		return err
	}

	// isZone: bool
	area.IsZone, err = p.r.ReadBool()
	if err != nil {
		return err
	}

	// zoneAreaRef: int32
	area.ZoneAreaRef, err = p.r.ReadInt32()
	if err != nil {
		return err
	}

	// mLast2DMapZoom: double (version >= 21 only)
	if p.m.Version >= 21 {
		area.Last2DMapZoom, err = p.r.ReadDouble()
		if err != nil {
			return err
		}
	}

	// mUserData: QMap<QString,QString>
	userDataCount, err := p.r.ReadInt32()
	if err != nil {
		return err
	}
	for i := int32(0); i < userDataCount; i++ {
		key, err := p.r.ReadQString()
		if err != nil {
			return err
		}
		value, err := p.r.ReadQString()
		if err != nil {
			return err
		}
		area.UserData[key] = value
	}

	// mMapLabels (version >= 21 only)
	if p.m.Version >= 21 {
		if err := p.readAreaLabels(area); err != nil {
			return err
		}
	}

	return nil
}

func (p *parser) readAreaLabels(area *MudletArea) error {
	count, err := p.r.ReadInt32()
	if err != nil {
		return err
	}
	area.Labels = make([]*MudletLabel, 0, count)
	for i := int32(0); i < count; i++ {
		labelID, err := p.r.ReadInt32()
		if err != nil {
			return err
		}
		label, err := p.readLabelV21(labelID)
		if err != nil {
			return err
		}
		area.Labels = append(area.Labels, label)
	}
	return nil
}

// --- Label readers ---

func (p *parser) readLabels() error {
	count, err := p.r.ReadInt32()
	if err != nil {
		return err
	}

	for i := int32(0); i < count; i++ {
		labelCount, err := p.r.ReadInt32()
		if err != nil {
			return err
		}
		areaID, err := p.r.ReadInt32()
		if err != nil {
			return err
		}

		labels := make([]*MudletLabel, 0, labelCount)
		for j := int32(0); j < labelCount; j++ {
			label, err := p.readLabel()
			if err != nil {
				return fmt.Errorf("label %d in area %d: %w", j, areaID, err)
			}
			labels = append(labels, label)
		}
		p.m.Labels[areaID] = labels
	}

	return nil
}

func (p *parser) readLabel() (*MudletLabel, error) {
	label := &MudletLabel{}
	var err error

	label.ID, err = p.r.ReadInt32()
	if err != nil {
		return nil, err
	}

	// pos: QVector3D
	label.Pos, err = p.readQVector3D()
	if err != nil {
		return nil, err
	}

	// dummy1, dummy2 (unused in v20)
	for i := 0; i < 2; i++ {
		if _, err := p.r.ReadDouble(); err != nil {
			return nil, err
		}
	}

	// size: QSizeF
	label.Width, err = p.r.ReadDouble()
	if err != nil {
		return nil, err
	}
	label.Height, err = p.r.ReadDouble()
	if err != nil {
		return nil, err
	}

	// text: QString
	label.Text, err = p.r.ReadQString()
	if err != nil {
		return nil, err
	}

	// fgColor, bgColor
	label.FgColor, err = p.readQColor()
	if err != nil {
		return nil, err
	}
	label.BgColor, err = p.readQColor()
	if err != nil {
		return nil, err
	}

	// QPixmap
	label.Pixmap, err = p.readQPixmap()
	if err != nil {
		return nil, err
	}

	// noScaling, showOnTop
	label.NoScaling, err = p.r.ReadBool()
	if err != nil {
		return nil, err
	}
	label.ShowOnTop, err = p.r.ReadBool()
	if err != nil {
		return nil, err
	}

	return label, nil
}

func (p *parser) readLabelV21(labelID int32) (*MudletLabel, error) {
	label := &MudletLabel{ID: labelID}
	var err error

	// pos: QVector3D
	label.Pos, err = p.readQVector3D()
	if err != nil {
		return nil, err
	}

	// size: QSizeF
	label.Width, err = p.r.ReadDouble()
	if err != nil {
		return nil, err
	}
	label.Height, err = p.r.ReadDouble()
	if err != nil {
		return nil, err
	}

	// text: QString
	label.Text, err = p.r.ReadQString()
	if err != nil {
		return nil, err
	}

	// fgColor, bgColor
	label.FgColor, err = p.readQColor()
	if err != nil {
		return nil, err
	}
	label.BgColor, err = p.readQColor()
	if err != nil {
		return nil, err
	}

	// QPixmap
	label.Pixmap, err = p.readQPixmap()
	if err != nil {
		return nil, err
	}

	// noScaling, showOnTop
	label.NoScaling, err = p.r.ReadBool()
	if err != nil {
		return nil, err
	}
	label.ShowOnTop, err = p.r.ReadBool()
	if err != nil {
		return nil, err
	}

	return label, nil
}

func (p *parser) readQPixmap() ([]byte, error) {
	// QPixmap marker
	_, err := p.r.ReadUInt32()
	if err != nil {
		return nil, err
	}

	// Check for PNG signature
	sig, err := p.r.Peek(4)
	if err != nil || len(sig) < 4 {
		return nil, nil
	}

	// PNG signature: 0x89 'P' 'N' 'G'
	if uint32(sig[0])<<24|uint32(sig[1])<<16|uint32(sig[2])<<8|uint32(sig[3]) != 0x89504e47 {
		return nil, nil
	}

	// Read PNG data until IEND + CRC
	return p.readPNG()
}

func (p *parser) readPNG() ([]byte, error) {
	var buf []byte
	needle := []byte{0x49, 0x45, 0x4e, 0x44} // 'I','E','N','D'
	for {
		peek, err := p.r.Peek(4)
		if err != nil || len(peek) < 4 {
			return buf, err
		}
		if peek[0] == needle[0] && peek[1] == needle[1] && peek[2] == needle[2] && peek[3] == needle[3] {
			// Read IEND + 4-byte CRC (8 bytes total)
			for i := 0; i < 8; i++ {
				b, err := p.r.ReadByte()
				if err != nil {
					return buf, err
				}
				buf = append(buf, b)
			}
			return buf, nil
		}
		b, err := p.r.ReadByte()
		if err != nil {
			return buf, err
		}
		buf = append(buf, b)
	}
}

func (p *parser) skipPNG() error {
	needle := []byte{0x49, 0x45, 0x4e, 0x44} // 'I','E','N','D'
	for {
		peek, err := p.r.Peek(4)
		if err != nil || len(peek) < 4 {
			return err
		}
		if peek[0] == needle[0] && peek[1] == needle[1] && peek[2] == needle[2] && peek[3] == needle[3] {
			// Skip IEND + 4-byte CRC
			return p.r.Skip(8)
		}
		if _, err := p.r.ReadByte(); err != nil {
			return err
		}
	}
}

// --- Room readers ---

func (p *parser) readRooms() error {
	for {
		peek, err := p.r.Peek(4)
		if err != nil || len(peek) < 4 {
			break
		}

		roomID, err := p.r.ReadInt32()
		if err != nil {
			break
		}

		room, err := p.readRoom(roomID)
		if err != nil {
			return fmt.Errorf("room %d: %w", roomID, err)
		}

		p.m.Rooms[roomID] = room
	}

	return nil
}

func (p *parser) readRoom(roomID int32) (*MudletRoom, error) {
	room := NewMudletRoom(roomID)
	var err error

	room.Area, err = p.r.ReadInt32()
	if err != nil {
		return nil, err
	}

	room.X, err = p.r.ReadInt32()
	if err != nil {
		return nil, err
	}
	room.Y, err = p.r.ReadInt32()
	if err != nil {
		return nil, err
	}
	room.Z, err = p.r.ReadInt32()
	if err != nil {
		return nil, err
	}

	// 12 standard exits
	for i := 0; i < 12; i++ {
		room.Exits[i], err = p.r.ReadInt32()
		if err != nil {
			return nil, err
		}
	}

	room.Environment, err = p.r.ReadInt32()
	if err != nil {
		return nil, err
	}

	room.Weight, err = p.r.ReadInt32()
	if err != nil {
		return nil, err
	}

	room.Name, err = p.r.ReadQString()
	if err != nil {
		return nil, err
	}

	room.IsLocked, err = p.r.ReadBool()
	if err != nil {
		return nil, err
	}

	// Special exits (version dependent)
	if err := p.readSpecialExits(room); err != nil {
		return nil, err
	}

	// Symbol
	if err := p.readRoomSymbol(room); err != nil {
		return nil, err
	}

	// Symbol color (v21+)
	if p.m.Version >= 21 {
		color, err := p.readQColor()
		if err != nil {
			return nil, err
		}
		room.SymbolColor = &color
	}

	// User data (v10+)
	if p.m.Version >= 10 {
		if err := p.readRoomUserData(room); err != nil {
			return nil, err
		}
	}

	// Custom lines (v11+)
	if p.m.Version >= 11 {
		if err := p.readRoomCustomLines(room); err != nil {
			return nil, err
		}
	}

	// Exit stubs (v13+)
	if p.m.Version >= 13 {
		if err := p.readExitStubs(room); err != nil {
			return nil, err
		}
	}

	// Exit weights and doors (v16+)
	if p.m.Version >= 16 {
		if err := p.readExitWeightsAndDoors(room); err != nil {
			return nil, err
		}
	}

	return room, nil
}

func (p *parser) readSpecialExits(room *MudletRoom) error {
	if p.m.Version >= 21 {
		// v21+: QMultiMap<QString, int>
		count, err := p.r.ReadInt32()
		if err != nil {
			return err
		}
		for i := int32(0); i < count; i++ {
			cmd, err := p.r.ReadQString()
			if err != nil {
				return err
			}
			destRoom, err := p.r.ReadInt32()
			if err != nil {
				return err
			}
			room.SpecialExits[cmd] = destRoom
		}
	} else if p.m.Version >= 6 {
		// v6-20: QMultiMap<int, QString>
		count, err := p.r.ReadInt32()
		if err != nil {
			return err
		}
		for i := int32(0); i < count; i++ {
			destRoom, err := p.r.ReadInt32()
			if err != nil {
				return err
			}
			cmd, err := p.r.ReadQString()
			if err != nil {
				return err
			}
			// Strip lock prefix ("0" or "1")
			if len(cmd) > 1 {
				cmd = cmd[1:]
			}
			room.SpecialExits[cmd] = destRoom
		}
	}
	return nil
}

func (p *parser) readRoomSymbol(room *MudletRoom) error {
	if p.m.Version >= 19 {
		var err error
		room.Symbol, err = p.r.ReadQString()
		return err
	} else if p.m.Version >= 9 {
		_, err := p.r.ReadByte()
		return err
	}
	return nil
}

func (p *parser) readRoomUserData(room *MudletRoom) error {
	count, err := p.r.ReadInt32()
	if err != nil {
		return err
	}
	for i := int32(0); i < count; i++ {
		key, err := p.r.ReadQString()
		if err != nil {
			return err
		}
		value, err := p.r.ReadQString()
		if err != nil {
			return err
		}
		room.UserData[key] = value
	}
	return nil
}

func (p *parser) readRoomCustomLines(room *MudletRoom) error {
	if p.m.Version >= 20 {
		return p.readRoomCustomLinesV20(room)
	}
	return p.readRoomCustomLinesOld(room)
}

func (p *parser) readRoomCustomLinesV20(room *MudletRoom) error {
	// customLines: QMap<QString, QList<QPointF>>
	count, err := p.r.ReadInt32()
	if err != nil {
		return err
	}
	for i := int32(0); i < count; i++ {
		dir, err := p.r.ReadQString()
		if err != nil {
			return err
		}
		pointCount, err := p.r.ReadInt32()
		if err != nil {
			return err
		}
		points := make([]Point2D, 0, pointCount)
		for j := int32(0); j < pointCount; j++ {
			x, err := p.r.ReadDouble()
			if err != nil {
				return err
			}
			y, err := p.r.ReadDouble()
			if err != nil {
				return err
			}
			points = append(points, Point2D{X: x, Y: y})
		}
		room.CustomLines[dir] = points
	}

	// customLinesArrow: QMap<QString, bool>
	count, err = p.r.ReadInt32()
	if err != nil {
		return err
	}
	for i := int32(0); i < count; i++ {
		dir, err := p.r.ReadQString()
		if err != nil {
			return err
		}
		arrow, err := p.r.ReadBool()
		if err != nil {
			return err
		}
		room.CustomLinesArrow[dir] = arrow
	}

	// customLinesColor: QMap<QString, QColor>
	count, err = p.r.ReadInt32()
	if err != nil {
		return err
	}
	for i := int32(0); i < count; i++ {
		dir, err := p.r.ReadQString()
		if err != nil {
			return err
		}
		color, err := p.readQColor()
		if err != nil {
			return err
		}
		room.CustomLinesColor[dir] = color
	}

	// customLinesStyle: QMap<QString, int>
	count, err = p.r.ReadInt32()
	if err != nil {
		return err
	}
	for i := int32(0); i < count; i++ {
		dir, err := p.r.ReadQString()
		if err != nil {
			return err
		}
		style, err := p.r.ReadInt32()
		if err != nil {
			return err
		}
		room.CustomLinesStyle[dir] = style
	}

	// Special exit locks (v21+)
	if p.m.Version >= 21 {
		count, err = p.r.ReadInt32()
		if err != nil {
			return err
		}
		room.SpecialExitLocks = make([]string, 0, count)
		for i := int32(0); i < count; i++ {
			lock, err := p.r.ReadQString()
			if err != nil {
				return err
			}
			room.SpecialExitLocks = append(room.SpecialExitLocks, lock)
		}
	}

	// exitLocks: QList<int>
	count, err = p.r.ReadInt32()
	if err != nil {
		return err
	}
	room.ExitLocks = make([]int32, 0, count)
	for i := int32(0); i < count; i++ {
		lock, err := p.r.ReadInt32()
		if err != nil {
			return err
		}
		room.ExitLocks = append(room.ExitLocks, lock)
	}

	return nil
}

func (p *parser) readRoomCustomLinesOld(room *MudletRoom) error {
	// customLines: QMap<QString, QList<QPointF>>
	count, err := p.r.ReadInt32()
	if err != nil {
		return err
	}
	for i := int32(0); i < count; i++ {
		dir, err := p.r.ReadQString()
		if err != nil {
			return err
		}
		pointCount, err := p.r.ReadInt32()
		if err != nil {
			return err
		}
		points := make([]Point2D, 0, pointCount)
		for j := int32(0); j < pointCount; j++ {
			x, err := p.r.ReadDouble()
			if err != nil {
				return err
			}
			y, err := p.r.ReadDouble()
			if err != nil {
				return err
			}
			points = append(points, Point2D{X: x, Y: y})
		}
		room.CustomLines[dir] = points
	}

	// customLinesArrow: QMap<QString, bool>
	count, err = p.r.ReadInt32()
	if err != nil {
		return err
	}
	for i := int32(0); i < count; i++ {
		dir, err := p.r.ReadQString()
		if err != nil {
			return err
		}
		arrow, err := p.r.ReadBool()
		if err != nil {
			return err
		}
		room.CustomLinesArrow[dir] = arrow
	}

	// customLinesColor: QMap<QString, QList<int>> (3 ints for RGB)
	count, err = p.r.ReadInt32()
	if err != nil {
		return err
	}
	for i := int32(0); i < count; i++ {
		dir, err := p.r.ReadQString()
		if err != nil {
			return err
		}
		rgbCount, err := p.r.ReadInt32()
		if err != nil {
			return err
		}
		var r, g, b int32
		for j := int32(0); j < rgbCount && j < 3; j++ {
			val, err := p.r.ReadInt32()
			if err != nil {
				return err
			}
			switch j {
			case 0:
				r = val
			case 1:
				g = val
			case 2:
				b = val
			}
		}
		// Skip extra values if any
		for j := int32(3); j < rgbCount; j++ {
			if _, err := p.r.ReadInt32(); err != nil {
				return err
			}
		}
		room.CustomLinesColor[dir] = Color{
			Red:   uint16(r) << 8,
			Green: uint16(g) << 8,
			Blue:  uint16(b) << 8,
			Alpha: 0xFFFF,
		}
	}

	// customLinesStyle: QMap<QString, QString>
	count, err = p.r.ReadInt32()
	if err != nil {
		return err
	}
	for i := int32(0); i < count; i++ {
		_, err := p.r.ReadQString()
		if err != nil {
			return err
		}
		// In old format, style is stored as QString
		_, err = p.r.ReadQString()
		if err != nil {
			return err
		}
	}

	// exitLocks: QList<int>
	count, err = p.r.ReadInt32()
	if err != nil {
		return err
	}
	room.ExitLocks = make([]int32, 0, count)
	for i := int32(0); i < count; i++ {
		lock, err := p.r.ReadInt32()
		if err != nil {
			return err
		}
		room.ExitLocks = append(room.ExitLocks, lock)
	}

	return nil
}

func (p *parser) readExitStubs(room *MudletRoom) error {
	count, err := p.r.ReadInt32()
	if err != nil {
		return err
	}
	room.ExitStubs = make([]int32, 0, count)
	for i := int32(0); i < count; i++ {
		stub, err := p.r.ReadInt32()
		if err != nil {
			return err
		}
		room.ExitStubs = append(room.ExitStubs, stub)
	}
	return nil
}

func (p *parser) readExitWeightsAndDoors(room *MudletRoom) error {
	// exitWeights: QMap<QString, int>
	count, err := p.r.ReadInt32()
	if err != nil {
		return err
	}
	for i := int32(0); i < count; i++ {
		dir, err := p.r.ReadQString()
		if err != nil {
			return err
		}
		weight, err := p.r.ReadInt32()
		if err != nil {
			return err
		}
		room.ExitWeights[dir] = weight
	}

	// doors: QMap<QString, int>
	count, err = p.r.ReadInt32()
	if err != nil {
		return err
	}
	for i := int32(0); i < count; i++ {
		dir, err := p.r.ReadQString()
		if err != nil {
			return err
		}
		doorType, err := p.r.ReadInt32()
		if err != nil {
			return err
		}
		room.Doors[dir] = doorType
	}

	return nil
}
