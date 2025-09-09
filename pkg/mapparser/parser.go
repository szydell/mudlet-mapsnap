package mapparser

import (
	"bufio"
	"errors"

	"fmt"
	"io"
	"os"
)

// ParseMapFile parses a Mudlet map file and returns a Map structure
func ParseMapFile(filename string) (m *Map, err error) {
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

	m, err = ParseMap(file)
	return m, err
}

// ParseMap parses a Mudlet map from an io.Reader
func ParseMap(reader io.Reader) (*Map, error) {
	// Wrap in bufio.Reader so we can use parseHeader
	br := bufio.NewReader(reader)

	// Initialize map structure
	m := &Map{
		Rooms:        make(map[int32]*Room),
		Areas:        make(map[int32]*Area),
		Environments: []Environment{},
		CustomLines:  []CustomLine{},
		Labels:       []Label{},
	}

	// Parse and validate header now; then try to parse areas (QMap<int, QString>)
	if err := parseHeader(br, &m.Header); err != nil {
		return nil, err
	}

	// According to Node.js reference (MudletMap serialization), the stream continues with:
	// envColors: QMap(QInt, QInt), then areaNames: QMap(QInt, QString, sorted), then other fields.
	// We'll parse envColors (ignored for now) and areaNames to populate Areas.
	qt := NewBinaryReader(br)
	// debug helper
	debugOn := os.Getenv("MAPSNAP_DEBUG") == "1"
	dbg := func(msg string) {
		if debugOn {
			fmt.Printf("[parser] @%d %s\n", qt.Position(), msg)
		}
	}
	dbg("after header")
	// 1) envColors QMap<int,int>
	if envCount, err := qt.ReadInt32(); err == nil && envCount >= 0 && envCount < 100000 {
		for i := int32(0); i < envCount; i++ {
			if _, err := qt.ReadInt32(); err != nil {
				break
			}
			if _, err := qt.ReadInt32(); err != nil {
				break
			}
		}
	}
	// 2) areaNames QMap<int, QString>
	if areaCount, err := qt.ReadInt32(); err == nil && areaCount >= 0 && areaCount < 100000 {
		areas := make(map[int32]*Area, areaCount)
		for i := int32(0); i < areaCount; i++ {
			id, err := qt.ReadInt32()
			if err != nil {
				break
			}
			name, err := qt.ReadQString()
			if err != nil {
				break
			}
			areas[id] = &Area{ID: id, Name: name}
		}
		if len(areas) > 0 {
			m.Areas = areas
		}
	}

	// 3) mCustomEnvColors QMap<int,QColor> - skip
	if cnt, err := qt.ReadInt32(); err == nil {
		for i := int32(0); i < cnt; i++ {
			if _, err := qt.ReadInt32(); err != nil {
				break
			}
			// QColor: 1 byte spec + 5x uint16
			if _, err := qt.ReadInt8(); err != nil {
				break
			}
			for j := 0; j < 5; j++ {
				if _, err := qt.ReadUInt16(); err != nil {
					break
				}
			}
		}
	}
	// 4) mpRoomDbHashToRoomId QMap<QString,QUInt> - skip
	if cnt, err := qt.ReadInt32(); err == nil {
		for i := int32(0); i < cnt; i++ {
			if _, err := qt.ReadQString(); err != nil {
				break
			}
			if _, err := qt.ReadUInt32(); err != nil {
				break
			}
		}
	}
	// 5) mUserData QMap<QString,QString> - skip
	if cnt, err := qt.ReadInt32(); err == nil {
		for i := int32(0); i < cnt; i++ {
			if _, err := qt.ReadQString(); err != nil {
				break
			}
			if _, err := qt.ReadQString(); err != nil {
				break
			}
		}
	}
	// 6) mapSymbolFont QFont - skip detailed fields
	if _, err := qt.ReadQString(); err == nil {
		_, _ = qt.ReadQString()
		_, _ = qt.ReadDouble()
		_, _ = qt.ReadInt32()
		_, _ = qt.ReadInt8()
		_, _ = qt.ReadUInt16()
		_, _ = qt.ReadByte()
		_, _ = qt.ReadInt8()
		_, _ = qt.ReadInt8()
		_, _ = qt.ReadUInt16()
		_, _ = qt.ReadInt8()
		_, _ = qt.ReadInt32()
		_, _ = qt.ReadInt32()
		_, _ = qt.ReadInt8()
		_, _ = qt.ReadInt8()
	}
	// 7) mapFontFudgeFactor (double)
	_, _ = qt.ReadDouble()
	// 8) useOnlyMapFont (bool)
	_, _ = qt.ReadBool()

	// 9) areas: MudletAreas - skip content but consume
	if cnt, err := qt.ReadInt32(); err == nil {
		for i := int32(0); i < cnt; i++ {
			// key id
			if _, err := qt.ReadInt32(); err != nil {
				break
			}
			// value MudletArea
			// QList<QUInt>
			if l, err := qt.ReadInt32(); err == nil {
				for j := int32(0); j < l; j++ {
					_, _ = qt.ReadUInt32()
				}
			}
			// QList<QInt>
			if l, err := qt.ReadInt32(); err == nil {
				for j := int32(0); j < l; j++ {
					_, _ = qt.ReadInt32()
				}
			}
			// QMultiMap<int,QPair<int,int>>
			if l, err := qt.ReadInt32(); err == nil {
				for j := int32(0); j < l; j++ {
					_, _ = qt.ReadInt32()
					_, _ = qt.ReadInt32()
					_, _ = qt.ReadInt32()
				}
			}
			// gridMode bool
			_, _ = qt.ReadBool()
			// six ints
			for k := 0; k < 6; k++ {
				_, _ = qt.ReadInt32()
			}
			// QVector (3 doubles)
			for k := 0; k < 3; k++ {
				_, _ = qt.ReadDouble()
			}
			// 4x QMap<int,int>
			for k := 0; k < 4; k++ {
				if n, err := qt.ReadInt32(); err == nil {
					for t := int32(0); t < n; t++ {
						_, _ = qt.ReadInt32()
						_, _ = qt.ReadInt32()
					}
				}
			}
			// QVector
			for k := 0; k < 3; k++ {
				_, _ = qt.ReadDouble()
			}
			// isZone bool, zoneAreaRef int
			_, _ = qt.ReadBool()
			_, _ = qt.ReadInt32()
			// userData QMap<QString,QString>
			if n, err := qt.ReadInt32(); err == nil {
				for t := int32(0); t < n; t++ {
					_, _ = qt.ReadQString()
					_, _ = qt.ReadQString()
				}
			}
		}
	}

	// 10) mRoomIdHash QMap<QString,QInt> - skip
	if cnt, err := qt.ReadInt32(); err == nil {
		for i := int32(0); i < cnt; i++ {
			_, _ = qt.ReadQString()
			_, _ = qt.ReadInt32()
		}
	}

	dbg("before labels")
	// 11) labels: MudletLabels - skip efficiently (handles embedded PNGs)
	// Even if MAPSNAP_SKIP_LABELS=1, we prefer structured skipping over heuristic scan for performance.
	if cnt, err := qt.ReadInt32(); err == nil {
		for i := int32(0); i < cnt; i++ {
			if total, err := qt.ReadInt32(); err == nil {
				_, _ = qt.ReadInt32()
				for j := int32(0); j < total; j++ { // label entries
					// Read minimal MudletLabel to skip
					_, _ = qt.ReadInt32() // id
					// pos (QVector: 3 doubles), dummy1 (1), dummy2 (1), size (QPair: 2) => total 7 doubles
					for k := 0; k < 7; k++ {
						_, _ = qt.ReadDouble()
					}
					_, _ = qt.ReadQString() // text
					// fgColor, bgColor
					_, _ = qt.ReadInt8()
					for c := 0; c < 5; c++ {
						_, _ = qt.ReadUInt16()
					}
					_, _ = qt.ReadInt8()
					for c := 0; c < 5; c++ {
						_, _ = qt.ReadUInt16()
					}
					// QPixMap: read header marker (uint32), then check the next 4 bytes for PNG magic and consume until IEND
					_, _ = qt.ReadUInt32()
					if sig, _ := qt.Peek(4); len(sig) == 4 {
						if uint32(sig[0])<<24|uint32(sig[1])<<16|uint32(sig[2])<<8|uint32(sig[3]) == 0x89504e47 {
							_ = skipPNG(qt)
						}
					}
					_, _ = qt.ReadBool()
					_, _ = qt.ReadBool()
				}
			}
		}
	} else {
		// Fallback only if we failed to read labels count
		dbg("labels count read failed, using heuristic skip to rooms")
		if err := skipToRoomsHeuristic(qt, m.Areas); err != nil {
			return nil, fmt.Errorf("skipToRoomsHeuristic: %w", err)
		}
	}

	dbg("before rooms")
	// 12) rooms: MudletRooms - parse minimal fields!
	if err := parseRooms(qt, m); err != nil {
		return nil, err
	}
	dbg("after rooms")
	return m, nil
}

// parseHeader reads the map file header
// Supports two formats:
// 1) Legacy placeholder format with magic "ATADNOOM" + 1-byte version
// 2) Actual Mudlet QDataStream: first value is qint32 version (e.g., 20)
func parseHeader(reader *bufio.Reader, header *Header) error {
	// Peek to check for legacy magic
	if peek, _ := reader.Peek(8); len(peek) == 8 && string(peek) == "ATADNOOM" {
		// Consume magic
		magic := make([]byte, 8)
		if _, err := io.ReadFull(reader, magic); err != nil {
			return fmt.Errorf("reading magic: %w", err)
		}
		header.Magic = string(magic)
		// Read the 1-byte version
		versionByte, err := reader.ReadByte()
		if err != nil {
			return fmt.Errorf("reading legacy version: %w", err)
		}
		header.Version = int8(versionByte)
		return nil
	}
	// Fallback: treat as Qt QDataStream and read qint32 version
	br := NewBinaryReader(reader)
	v, err := br.ReadInt32()
	if err != nil {
		return fmt.Errorf("reading Qt version: %w", err)
	}
	header.Magic = "" // QDataStream has no magic prefix
	header.Version = int8(v)
	return nil
}

// parseRooms parses the MudletRooms section (sequence of id->MudletRoom entries until EOF)
func parseRooms(qt *BinaryReader, m *Map) error {
	dirs := []string{"north", "northeast", "east", "southeast", "south", "southwest", "west", "northwest", "up", "down", "in", "out", "north2", "east2", "south2", "west2"}
	for {
		// If no more data, we are done
		if peek, err := qt.Peek(1); err != nil || len(peek) == 0 {
			break
		}
		// Read room id
		id, err := qt.ReadInt32()
		if err != nil {
			// likely EOF
			break
		}
		r := &Room{ID: id}
		// area (int32) - currently unused in our model
		_, _ = qt.ReadInt32()
		// coordinates
		r.X, _ = qt.ReadInt32()
		r.Y, _ = qt.ReadInt32()
		r.Z, _ = qt.ReadInt32()
		// 12 standard exits
		exits := make([]Exit, 0, len(dirs))
		for _, name := range dirs {
			tgt, _ := qt.ReadInt32()
			exits = append(exits, Exit{Direction: name, TargetID: tgt})
		}
		// environment, weight
		r.Environment, _ = qt.ReadInt32()
		_, _ = qt.ReadInt32()
		// name
		r.Name, _ = qt.ReadQString()
		// isLocked
		_, _ = qt.ReadBool()
		// rawSpecialExits QMultiMap<QUInt, QString>
		if n, err := qt.ReadUInt32(); err == nil {
			for i := uint32(0); i < n; i++ {
				_, _ = qt.ReadUInt32()
				_, _ = qt.ReadQString()
			}
		}
		// symbol QString (unused)
		_, _ = qt.ReadQString()
		// userData QMap<QString,QString>
		if n, err := qt.ReadUInt32(); err == nil {
			for i := uint32(0); i < n; i++ {
				_, _ = qt.ReadQString()
				_, _ = qt.ReadQString()
			}
		}
		// customLines QMap<QString, QList<QPoint>>
		if n, err := qt.ReadUInt32(); err == nil {
			for i := uint32(0); i < n; i++ {
				_, _ = qt.ReadQString()
				if l, err := qt.ReadUInt32(); err == nil {
					for j := uint32(0); j < l; j++ {
						// QPoint: two doubles
						_, _ = qt.ReadDouble()
						_, _ = qt.ReadDouble()
					}
				}
			}
		}
		// customLinesArrow QMap<QString, QBool>
		if n, err := qt.ReadUInt32(); err == nil {
			for i := uint32(0); i < n; i++ {
				_, _ = qt.ReadQString()
				_, _ = qt.ReadBool()
			}
		}
		// customLinesColor QMap<QString, QColor>
		if n, err := qt.ReadUInt32(); err == nil {
			for i := uint32(0); i < n; i++ {
				_, _ = qt.ReadQString()
				// QColor: 1 byte spec + 5x uint16
				_, _ = qt.ReadInt8()
				for c := 0; c < 5; c++ {
					_, _ = qt.ReadUInt16()
				}
			}
		}
		// customLinesStyle QMap<QString, QUInt>
		if n, err := qt.ReadUInt32(); err == nil {
			for i := uint32(0); i < n; i++ {
				_, _ = qt.ReadQString()
				_, _ = qt.ReadUInt32()
			}
		}
		// exitLocks QList<QInt>
		if l, err := qt.ReadUInt32(); err == nil {
			for i := uint32(0); i < l; i++ {
				_, _ = qt.ReadInt32()
			}
		}
		// stubs QList<QInt>
		if l, err := qt.ReadUInt32(); err == nil {
			for i := uint32(0); i < l; i++ {
				_, _ = qt.ReadInt32()
			}
		}
		// exitWeights QMap<QString, QInt>
		if n, err := qt.ReadUInt32(); err == nil {
			for i := uint32(0); i < n; i++ {
				_, _ = qt.ReadQString()
				_, _ = qt.ReadInt32()
			}
		}
		// doors QMap<QString, QInt>
		if n, err := qt.ReadUInt32(); err == nil {
			for i := uint32(0); i < n; i++ {
				_, _ = qt.ReadQString()
				_, _ = qt.ReadInt32()
			}
		}

		r.Exits = exits
		m.Rooms[r.ID] = r
	}
	return nil
}

// skipPNG scans forward from the current position to find the PNG IEND chunk marker and consumes up to and including it.
func skipPNG(qt *BinaryReader) error {
	needle := []byte{0x49, 0x45, 0x4e, 0x44} // 'IEND'
	buf := make([]byte, 4)
	// initialize window
	for {
		peek, err := qt.Peek(4)
		if err != nil || len(peek) < 4 {
			return err
		}
		copy(buf, peek)
		if buf[0] == needle[0] && buf[1] == needle[1] && buf[2] == needle[2] && buf[3] == needle[3] {
			// consume 'IEND' + 4-byte CRC to land after PNG
			if err := qt.Skip(8); err != nil {
				return err
			}
			return nil
		}
		// advance by one byte and continue
		if _, err := qt.ReadByte(); err != nil {
			return err
		}
	}
}

// skipToRoomsHeuristic advances the reader until it finds a pair of int32 values where the second (area) exists in areas map.
// This is a best-effort fallback to jump over labels when label parsing is problematic.
func skipToRoomsHeuristic(qt *BinaryReader, areas map[int32]*Area) error {
	// Build a fast lookup of area ids
	valid := make(map[int32]struct{}, len(areas))
	for id := range areas {
		valid[id] = struct{}{}
	}
	// Sliding window over bytes, checking 8-byte sequences as two int32
	for {
		peek, err := qt.Peek(8)
		if err != nil || len(peek) < 8 {
			return fmt.Errorf("EOF before rooms")
		}
		id := int32(uint32(peek[0])<<24 | uint32(peek[1])<<16 | uint32(peek[2])<<8 | uint32(peek[3]))
		area := int32(uint32(peek[4])<<24 | uint32(peek[5])<<16 | uint32(peek[6])<<8 | uint32(peek[7]))
		if id > 0 {
			if _, ok := valid[area]; ok {
				// Found a plausible start of a room entry. Do not consume bytes; let parseRooms proceed from here.
				return nil
			}
		}
		// advance by one byte
		if _, err := qt.ReadByte(); err != nil {
			return err
		}
	}
}
