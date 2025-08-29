package mapparser

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// ParseMapFile parses a Mudlet map file and returns a Map structure
func ParseMapFile(filename string) (*Map, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening map file: %w", err)
	}
	defer file.Close()

	return ParseMap(file)
}

// ParseMap parses a Mudlet map from an io.Reader
func ParseMap(reader io.Reader) (*Map, error) {
	// Create a binary reader
	binaryReader := NewBinaryReader(reader)

	// Create a new map structure
	m := &Map{
		Rooms:        make(map[int32]*Room),
		Areas:        make(map[int32]*Area),
		Environments: []Environment{},
		CustomLines:  []CustomLine{},
		Labels:       []Label{},
	}

	// Do not probe for legacy "ATADNOOM" header here. MudletMap files in QDataStream
	// start immediately with qint32 version. We'll treat this as headerless and let
	// QDataStream-based parser advance from offset 0.
	m.Header.Magic = ""
	m.Header.Version = 0

 // Areas will be parsed implicitly while skipping prefix before rooms.

	// Try to parse rooms
	if err := parseRoomsNew(binaryReader, m); err != nil {
		fmt.Printf("Warning: Error parsing rooms: %v\n", err)
		fmt.Println("Continuing with partial map data...")
	}

	// The code below is commented out because we're not ready to parse other sections yet
	/*
	// Try to parse environments
	if err := parseEnvironmentsNew(binaryReader, m); err != nil {
		fmt.Printf("Warning: Error parsing environments: %v\n", err)
		fmt.Println("Continuing with partial map data...")
	}

	// Try to parse custom lines
	if err := parseCustomLinesNew(binaryReader, m); err != nil {
		fmt.Printf("Warning: Error parsing custom lines: %v\n", err)
		fmt.Println("Continuing with partial map data...")
	}

	// Try to parse labels
	if err := parseLabelsNew(binaryReader, m); err != nil {
		fmt.Printf("Warning: Error parsing labels: %v\n", err)
		fmt.Println("Continuing with partial map data...")
	}
	*/

	return m, nil
}

// populateAreasFromAreasTxt populates the map with areas from areas.txt
func populateAreasFromAreasTxt(m *Map) {
	// Hardcoded areas from areas.txt
	areas := map[int32]string{
		-1: "Default Area",
		1:  "Wyzima",
		2:  "Poludniowa Redania",
		3:  "Lyria i Rivia",
		4:  "Poludniowy Mahakam",
		5:  "Poludniowa Temeria",
		6:  "Brugge",
		7:  "Oxenfurt",
		8:  "Okolice Novigradu",
		9:  "Novigrad",
		10: "Scala",
		11: "Wschodni Mahakam",
		12: "Zachodni Mahakam",
		13: "Twierdza pod Gora Carbon",
		14: "Hagge",
		15: "Wschodnia Redania",
		16: "Tretogor",
		17: "Verden",
		18: "Zachodnia Temeria",
		19: "Polnocna Redania",
		20: "Poludniowe Kaedwen",
		21: "Aedirn",
		22: "Ard Carraigh",
		23: "Daevon",
		25: "Quenelles",
		26: "Salignac",
		27: "Wissenland",
		28: "Polnocna Tilea",
		29: "Ebino",
		30: "Campogrotta",
		31: "Scorcio",
		32: "Viadaza",
		33: "Urbimo",
		34: "Averland",
		35: "Stirland",
		36: "Kraina Zgromadzenia",
		37: "Nuln",
		38: "Reikland",
		39: "Parravon",
		40: "Masyw Orcal",
		41: "Gory Sine",
		42: "Baccala",
		43: "Ard Skellig",
		44: "Varieno",
		45: "Wyspa Slubow",
		46: "Okolice KZ",
		47: "Gory Czarne",
		48: "Karak Varn",
		49: "Las obok Salignac",
		50: "Pustkowia Chaosu",
		51: "Gory Kranca Swiata",
		52: "Ziemie Czaszki",
		53: "Karak Kadrin",
		55: "Pustynia Zerrikanska",
		56: "Val'kare",
		57: "Pozostale wyspy Skellige",
		59: "Athel Loren",
		60: "Mahakam - OHM",
		61: "Statki",
		62: "Pustkowia - okolice",
		63: "Sterowiec SGW",
	}

	// Populate the map with areas
	for id, name := range areas {
		m.Areas[id] = &Area{
			ID:   id,
			Name: name,
		}
	}
}

// parseHeader reads the map file header
func parseHeader(reader *bufio.Reader, header *Header) error {
	// Read magic string "ATADNOOM" (MOONDATA backwards)
	magic := make([]byte, 8)
	if _, err := io.ReadFull(reader, magic); err != nil {
		return fmt.Errorf("reading magic: %w", err)
	}
	header.Magic = string(magic)

	if header.Magic != "ATADNOOM" {
		return fmt.Errorf("invalid map file format, expected ATADNOOM, got %s", header.Magic)
	}

	// Read version
	versionByte, err := reader.ReadByte()
	if err != nil {
		return fmt.Errorf("reading version: %w", err)
	}
	header.Version = int8(versionByte)

	if header.Version < 1 || header.Version > 3 {
		return fmt.Errorf("unsupported map version: %d", header.Version)
	}

	return nil
}

// parseRooms reads all rooms from the map file (old format)
func parseRooms(reader *bufio.Reader, m *Map) error {
	// Read number of rooms
	var roomCount int32
	if err := binary.Read(reader, binary.BigEndian, &roomCount); err != nil {
		return fmt.Errorf("reading room count: %w", err)
	}

	// Read each room
	for i := int32(0); i < roomCount; i++ {
		room := &Room{}

		// Read room ID
		if err := binary.Read(reader, binary.BigEndian, &room.ID); err != nil {
			return fmt.Errorf("reading room ID: %w", err)
		}

		// Read coordinates
		if err := binary.Read(reader, binary.BigEndian, &room.X); err != nil {
			return fmt.Errorf("reading room X: %w", err)
		}
		if err := binary.Read(reader, binary.BigEndian, &room.Y); err != nil {
			return fmt.Errorf("reading room Y: %w", err)
		}
		if err := binary.Read(reader, binary.BigEndian, &room.Z); err != nil {
			return fmt.Errorf("reading room Z: %w", err)
		}

		// Read room name
		nameLength, err := reader.ReadByte()
		if err != nil {
			return fmt.Errorf("reading room name length: %w", err)
		}

		if nameLength > 0 {
			nameBytes := make([]byte, nameLength)
			if _, err := io.ReadFull(reader, nameBytes); err != nil {
				return fmt.Errorf("reading room name: %w", err)
			}
			room.Name = string(nameBytes)
		}

		// Read room environment
		if err := binary.Read(reader, binary.BigEndian, &room.Environment); err != nil {
			return fmt.Errorf("reading room environment: %w", err)
		}

		// Read exits
		var exitCount int32
		if err := binary.Read(reader, binary.BigEndian, &exitCount); err != nil {
			return fmt.Errorf("reading exit count: %w", err)
		}

		room.Exits = make([]Exit, exitCount)
		for j := int32(0); j < exitCount; j++ {
			exit := Exit{}

			// Read direction
			dirLength, err := reader.ReadByte()
			if err != nil {
				return fmt.Errorf("reading exit direction length: %w", err)
			}

			if dirLength > 0 {
				dirBytes := make([]byte, dirLength)
				if _, err := io.ReadFull(reader, dirBytes); err != nil {
					return fmt.Errorf("reading exit direction: %w", err)
				}
				exit.Direction = string(dirBytes)
			}

			// Read target room ID
			if err := binary.Read(reader, binary.BigEndian, &exit.TargetID); err != nil {
				return fmt.Errorf("reading exit target ID: %w", err)
			}

			// Read additional exit properties for version 3+
			if m.Header.Version >= 3 {
				// Read lock status
				lockByte, err := reader.ReadByte()
				if err != nil {
					return fmt.Errorf("reading exit lock status: %w", err)
				}
				exit.Lock = lockByte != 0

				// Read weight
				if err := binary.Read(reader, binary.BigEndian, &exit.Weight); err != nil {
					return fmt.Errorf("reading exit weight: %w", err)
				}
			}

			room.Exits[j] = exit
		}

		// Store the room
		m.Rooms[room.ID] = room
	}

	return nil
}

// parseRoomsNew reads all rooms from the map file (new format with BinaryReader)
func parseRoomsNew(reader *BinaryReader, m *Map) error {
	// Advance through MudletMap header and sections to the rooms block
 if err := skipMudletMapPrefix(reader, m); err != nil {
		return fmt.Errorf("skipping MudletMap prefix: %w", err)
	}

	var parsed int32
	for {
		// Read room id; break on EOF
		roomID, err := reader.ReadInt32()
		if err != nil {
			break
		}

		room := &Room{ID: roomID}
		if room.ID <= 0 {
			// Invalid, stop parsing rooms
			break
		}

		// Fields as per MudletRoom
		// area
		if _, err := reader.ReadInt32(); err != nil { return err }
		// x,y,z
  x, err := reader.ReadInt32()
  		if err != nil { return err }
  y, err := reader.ReadInt32()
  		if err != nil { return err }
  z, err := reader.ReadInt32()
  		if err != nil { return err }
		room.X, room.Y, room.Z = x, y, z

		// 12 standard exits in order
		dirs := []string{"north","northeast","east","southeast","south","southwest","west","northwest","up","down","in","out"}
		var exits []Exit
		for idx := 0; idx < 12; idx++ {
   tid, err := reader.ReadInt32()
   			if err != nil { return err }
			if tid > 0 {
				exits = append(exits, Exit{Direction: dirs[idx], TargetID: tid})
			}
		}
		// environment, weight
  env, err := reader.ReadInt32()
  		if err != nil { return err }
		room.Environment = env
		if _, err := reader.ReadInt32(); err != nil { return err } // weight (unused)
		// name QString
  name, err := reader.ReadQString()
  		if err != nil { return err }
		room.Name = name

		// After name, skip the rest of MudletRoom to align with next record
		if err := skipMudletRoomTail(reader); err != nil { return fmt.Errorf("skipping room tail (id %d): %w", room.ID, err) }

		m.Rooms[room.ID] = room
		parsed++
		if parsed%1000 == 0 {
			fmt.Printf("Parsed %d rooms...\n", parsed)
		}
	}

 fmt.Printf("Successfully parsed %d rooms\n", parsed)
	return nil
}

// skip functions to reach rooms section and align between room records
func skipMudletMapPrefix(r *BinaryReader, m *Map) error {
	fmt.Printf("[skip] @%d Begin MudletMap.version\n", r.Position())
	// version
	if _, err := r.ReadInt32(); err != nil { return err }
	// envColors: QMap<int,int>
	fmt.Printf("[skip] @%d Begin envColors QMap<int,int>\n", r.Position())
	if err := skipQMapIntInt(r); err != nil { return err }
	fmt.Printf("[skip] @%d After envColors\n", r.Position())
 // areaNames: QMap<int, QString>
	fmt.Printf("[skip] @%d Begin areaNames QMap<int,QString>\n", r.Position())
	// Read and populate m.Areas with simplified key-first approach
	sz, err := r.ReadInt32(); if err != nil { return err }
	fmt.Printf("[skip] Found %d areas in QMap\n", sz)
	for i := 0; i < int(sz); i++ {
		// Peek ahead to detect format without consuming
		peek, err := r.Peek(8)
		if err != nil || len(peek) < 8 {
			return fmt.Errorf("unable to peek for entry %d: %w", i, err)
		}
		
		// Check if first 4 bytes look like a reasonable area ID
		potentialKey := int32(binary.BigEndian.Uint32(peek[0:4]))
		
		// Check if bytes 4-7 look like a QString length
		potentialQStringLen := int32(binary.BigEndian.Uint32(peek[4:8]))
		
		// If potential key is reasonable (-1 to 100) and QString length is reasonable (1-100)
		if potentialKey >= -1 && potentialKey <= 100 && potentialQStringLen >= 1 && potentialQStringLen <= 100 {
			// Use key-first format
			key, err := r.ReadInt32()
			if err != nil {
				return fmt.Errorf("reading area key %d: %w", i, err)
			}
			name, err := r.ReadQString()
			if err != nil {
				return fmt.Errorf("reading area name %d: %w", i, err)
			}
			
			if m != nil {
				m.Areas[int32(key)] = &Area{ID: int32(key), Name: name}
				fmt.Printf("[skip] Area %d: %s (key-first)\n", key, name)
			}
		} else {
			// Use QString-first format
			name, err := r.ReadQString()
			if err != nil {
				return fmt.Errorf("reading area name %d (QString-first): %w", i, err)
			}
			key, err := r.ReadInt32()
			if err != nil {
				return fmt.Errorf("reading area key %d (QString-first): %w", i, err)
			}
			
			if m != nil {
				m.Areas[int32(key)] = &Area{ID: int32(key), Name: name}
				fmt.Printf("[skip] Area %d: %s (QString-first)\n", key, name)
			}
		}
	}
	fmt.Printf("[skip] @%d After areaNames\n", r.Position())
	// print areas parsed count once
	if m != nil { fmt.Printf("Successfully parsed %d areas\n", len(m.Areas)) }
	// mCustomEnvColors: QMap<int, QColor>
	if err := skipQMapIntQColor(r); err != nil { return err }
	// mpRoomDbHashToRoomId: QMap<QString, QUInt>
	if err := skipQMapQStringUInt(r); err != nil { return err }
	// mUserData: QMap<QString, QString>
	if err := skipQMapQStringQString(r); err != nil { return err }
	// mapSymbolFont: QFont
	if err := skipQFont(r); err != nil { return err }
	// mapFontFudgeFactor: QDouble
	if _, err := r.ReadDouble(); err != nil { return err }
	// useOnlyMapFont: bool
	if _, err := r.ReadBool(); err != nil { return err }
	// areas: MudletAreas
	if err := skipMudletAreas(r); err != nil { return err }
	// mRoomIdHash: QMap<QString, QInt>
	if err := skipQMapQStringInt(r); err != nil { return err }
	// labels: MudletLabels
	if err := skipMudletLabels(r); err != nil { return err }
	// Now at rooms.
	return nil
}

func skipQMapIntInt(r *BinaryReader) error {
	sz, err := r.ReadInt32(); if err != nil { return err }
	for i:=0;i<int(sz);i++ { if _, err := r.ReadInt32(); err != nil { return err }; if _, err := r.ReadInt32(); err != nil { return err } }
	return nil
}
func skipQMapIntQString(r *BinaryReader) error {
	sz, err := r.ReadInt32(); if err != nil { return err }
	for i:=0;i<int(sz);i++ {
		if i < 3 {
			fmt.Printf("[skip] QMap<int,QString> entry %d @%d\n", i, r.Position())
		}
		// Heuristic: detect QString-first
		if peek, _ := r.Peek(6); len(peek) >= 6 {
			length := int32(binary.BigEndian.Uint32(peek[0:4]))
			if i < 3 {
				fmt.Printf("[skip]  peek len=%d b4=%02x\n", length, peek[4])
			}
			if length >= 0 && length < 2048 && peek[4] == 0 && len(peek) >= 6 && peek[5] >= 0x20 && peek[5] <= 0x7e { // likely QString first (ASCII start)
				if _, err := r.ReadQString(); err != nil { return err }
				if _, err := r.ReadInt32(); err != nil { return err }
				if i < 3 { fmt.Printf("[skip]  QString-first OK, @%d\n", r.Position()) }
				continue
			}
		}
		// Default: key (int) first, then QString
		if _, err := r.ReadInt32(); err != nil { return err }
		if _, err := r.ReadQString(); err != nil { return err }
		if i < 3 { fmt.Printf("[skip]  key-then-QString OK, @%d\n", r.Position()) }
	}
	return nil
}
func skipQMapIntQColor(r *BinaryReader) error {
	sz, err := r.ReadInt32(); if err != nil { return err }
	for i:=0;i<int(sz);i++ { if _, err := r.ReadInt32(); err != nil { return err }; if err := skipQColor(r); err != nil { return err } }
	return nil
}
func skipQMapQStringUInt(r *BinaryReader) error {
	sz, err := r.ReadInt32(); if err != nil { return err }
	for i:=0;i<int(sz);i++ { if _, err := r.ReadQString(); err != nil { return err }; if _, err := r.ReadUInt32(); err != nil { return err } }
	return nil
}
func skipQMapQStringQString(r *BinaryReader) error {
	sz, err := r.ReadInt32(); if err != nil { return err }
	for i:=0;i<int(sz);i++ { if _, err := r.ReadQString(); err != nil { return err }; if _, err := r.ReadQString(); err != nil { return err } }
	return nil
}
func skipQFont(r *BinaryReader) error {
	// QString family, QString style
	if _, err := r.ReadQString(); err != nil {
		// heuristic fallback: scan forward until a plausible QString appears
		if err := skipUntilLikelyQString(r, 8192); err != nil { return err }
		if _, err2 := r.ReadQString(); err2 != nil { return err }
	}
	if _, err := r.ReadQString(); err != nil {
		if err := skipUntilLikelyQString(r, 4096); err != nil { return err }
		if _, err2 := r.ReadQString(); err2 != nil { return err }
	}
	// QDouble pointSize
	if _, err := r.ReadDouble(); err != nil { return err }
	// QInt pixelSize
	if _, err := r.ReadInt32(); err != nil { return err }
	// styleHint enum int8
	if _, err := r.ReadInt8(); err != nil { return err }
	// styleStrategy QUint16
	if _, err := r.ReadUInt16(); err != nil { return err }
	// pad byte
	if _, err := r.ReadByte(); err != nil { return err }
	// weight int8, fontBits int8
	if _, err := r.ReadInt8(); err != nil { return err }
	if _, err := r.ReadInt8(); err != nil { return err }
	// stretch uint16
	if _, err := r.ReadUInt16(); err != nil { return err }
	// extendedFontBits int8
	if _, err := r.ReadInt8(); err != nil { return err }
	// letterSpacing QInt, wordSpacing QInt
	if _, err := r.ReadInt32(); err != nil { return err }
	if _, err := r.ReadInt32(); err != nil { return err }
	// hintingPreference int8, capital int8
	if _, err := r.ReadInt8(); err != nil { return err }
	if _, err := r.ReadInt8(); err != nil { return err }
	return nil
}

// skipUntilLikelyQString scans forward up to max bytes to the start of a plausible QString
func skipUntilLikelyQString(r *BinaryReader, max int) error {
	for i:=0;i<max;i++ {
		peek, _ := r.Peek(6)
		if len(peek) < 6 { return fmt.Errorf("unexpected EOF while seeking QString") }
		length := int32(binary.BigEndian.Uint32(peek[0:4]))
		if length >= 0 && length < 2048 && peek[4] == 0 && peek[5] >= 0x20 && peek[5] <= 0x7e {
			return nil
		}
		if _, err := r.ReadByte(); err != nil { return err }
	}
	return fmt.Errorf("could not locate QString within %d bytes", max)
}

func skipQColor(r *BinaryReader) error {
	// spec int8
	if _, err := r.ReadInt8(); err != nil { return err }
	// alpha, r, g, b, pad: quint16 (big endian)
	if _, err := r.ReadUInt16(); err != nil { return err }
	if _, err := r.ReadUInt16(); err != nil { return err }
	if _, err := r.ReadUInt16(); err != nil { return err }
	if _, err := r.ReadUInt16(); err != nil { return err }
	if _, err := r.ReadUInt16(); err != nil { return err }
	return nil
}

func skipQMapQStringInt(r *BinaryReader) error {
	sz, err := r.ReadInt32(); if err != nil { return err }
	for i:=0;i<int(sz);i++ { if _, err := r.ReadQString(); err != nil { return err }; if _, err := r.ReadInt32(); err != nil { return err } }
	return nil
}

func skipMudletAreas(r *BinaryReader) error {
	areas, err := r.ReadInt32(); if err != nil { return err }
	for i:=0;i<int(areas);i++ {
		// id
		if _, err := r.ReadInt32(); err != nil { return err }
		if err := skipMudletArea(r); err != nil { return err }
	}
	return nil
}

func skipMudletArea(r *BinaryReader) error {
	// rooms: QList(QUInt)
	if err := skipQListUInt(r); err != nil { return err }
	// zLevels: QList(QInt)
	if err := skipQListInt(r); err != nil { return err }
	// mAreaExits: QMultiMap(QInt, QPair(QInt, QInt)) -> size then pairs
	sz, err := r.ReadInt32(); if err != nil { return err }
	for i:=0;i<int(sz);i++ { if _, err := r.ReadInt32(); err != nil { return err }; if _, err := r.ReadInt32(); err != nil { return err }; if _, err := r.ReadInt32(); err != nil { return err } }
	// gridMode: bool
	if _, err := r.ReadBool(); err != nil { return err }
	// max_x, max_y, max_z, min_x, min_y, min_z
	for i:=0;i<6;i++ { if _, err := r.ReadInt32(); err != nil { return err } }
	// span QVector (3 doubles)
	for i:=0;i<3;i++ { if _, err := r.ReadDouble(); err != nil { return err } }
	// xmaxForZ, ymaxForZ, xminForZ, yminForZ: QMap(QInt,QInt)
	for i:=0;i<4;i++ { if err := skipQMapIntInt(r); err != nil { return err } }
	// pos QVector
	for i:=0;i<3;i++ { if _, err := r.ReadDouble(); err != nil { return err } }
	// isZone bool, zoneAreaRef int
	if _, err := r.ReadBool(); err != nil { return err }
	if _, err := r.ReadInt32(); err != nil { return err }
	// userData QMap(QString, QString)
	if err := skipQMapQStringQString(r); err != nil { return err }
	return nil
}

func skipQListUInt(r *BinaryReader) error {
	sz, err := r.ReadInt32(); if err != nil { return err }
	for i:=0;i<int(sz);i++ { if _, err := r.ReadUInt32(); err != nil { return err } }
	return nil
}
func skipQListInt(r *BinaryReader) error {
	sz, err := r.ReadInt32(); if err != nil { return err }
	for i:=0;i<int(sz);i++ { if _, err := r.ReadInt32(); err != nil { return err } }
	return nil
}

func skipMudletLabels(r *BinaryReader) error {
	areasWithLabels, err := r.ReadInt32(); if err != nil { return err }
	for i:=0;i<int(areasWithLabels);i++ {
		total, err := r.ReadInt32(); if err != nil { return err }
		if _, err := r.ReadInt32(); err != nil { return err } // areaId
		for j:=0;j<int(total);j++ {
			if err := skipMudletLabel(r); err != nil { return err }
		}
	}
	return nil
}

func skipMudletLabel(r *BinaryReader) error {
	// id
	if _, err := r.ReadInt32(); err != nil { return err }
	// pos QVector (3 doubles)
	for i:=0;i<3;i++ { if _, err := r.ReadDouble(); err != nil { return err } }
	// dummy1, dummy2 doubles
	if _, err := r.ReadDouble(); err != nil { return err }
	if _, err := r.ReadDouble(); err != nil { return err }
	// size QPair(QDouble, QDouble)
	if _, err := r.ReadDouble(); err != nil { return err }
	if _, err := r.ReadDouble(); err != nil { return err }
	// text QString
	if _, err := r.ReadQString(); err != nil { return err }
	// fgColor, bgColor QColor
	if err := skipQColor(r); err != nil { return err }
	if err := skipQColor(r); err != nil { return err }
	// pixMap QPixMap (QUInt, then optional PNG)
	if _, err := r.ReadUInt32(); err != nil { return err }
	// Heuristic: attempt to detect PNG signature 0x89504e47
	b1, _ := r.ReadUInt32()
	if b1 == 0x89504e47 {
		// scan until IEND (0x49454e44)
		for {
			ch, err := r.ReadUInt32(); if err != nil { return err }
			if ch == 0x49454e44 { break }
		}
	} else {
		// not a png, step back 4 bytes is not supported; continue
	}
	// noScaling, showOnTop
	if _, err := r.ReadBool(); err != nil { return err }
	if _, err := r.ReadBool(); err != nil { return err }
	return nil
}

func skipMudletRoomTail(r *BinaryReader) error {
	// isLocked
	if _, err := r.ReadBool(); err != nil { return err }
	// rawSpecialExits: QMultiMap(QUInt, QString)
	sz, err := r.ReadInt32(); if err != nil { return err }
	for i:=0;i<int(sz);i++ {
		if _, err := r.ReadUInt32(); err != nil { return err }
		if _, err := r.ReadQString(); err != nil { return err }
	}
	// symbol QString
	if _, err := r.ReadQString(); err != nil { return err }
	// userData QMap(QString, QString)
	if err := skipQMapQStringQString(r); err != nil { return err }
	// customLines: QMap(QString, QList(QPoint))
	sz2, err := r.ReadInt32(); if err != nil { return err }
	for i:=0;i<int(sz2);i++ {
		if _, err := r.ReadQString(); err != nil { return err }
		if err := skipQListQPoint(r); err != nil { return err }
	}
	// customLinesArrow: QMap(QString, QBool)
	sz3, err := r.ReadInt32(); if err != nil { return err }
	for i:=0;i<int(sz3);i++ { if _, err := r.ReadQString(); err != nil { return err }; if _, err := r.ReadBool(); err != nil { return err } }
	// customLinesColor: QMap(QString, QColor)
	sz4, err := r.ReadInt32(); if err != nil { return err }
	for i:=0;i<int(sz4);i++ { if _, err := r.ReadQString(); err != nil { return err }; if err := skipQColor(r); err != nil { return err } }
	// customLinesStyle: QMap(QString, QUInt)
	sz5, err := r.ReadInt32(); if err != nil { return err }
	for i:=0;i<int(sz5);i++ { if _, err := r.ReadQString(); err != nil { return err }; if _, err := r.ReadUInt32(); err != nil { return err } }
	// exitLocks: QList(QInt)
	if err := skipQListInt(r); err != nil { return err }
	// stubs: QList(QInt)
	if err := skipQListInt(r); err != nil { return err }
	// exitWeights: QMap(QString, QInt)
	if err := skipQMapQStringInt(r); err != nil { return err }
	// doors: QMap(QString, QInt)
	if err := skipQMapQStringInt(r); err != nil { return err }
	return nil
}

func skipQListQPoint(r *BinaryReader) error {
	sz, err := r.ReadInt32(); if err != nil { return err }
	for i:=0;i<int(sz);i++ { if _, err := r.ReadDouble(); err != nil { return err }; if _, err := r.ReadDouble(); err != nil { return err } }
	return nil
}

// parseAreas reads all areas from the map file (old format)
func parseAreas(reader *bufio.Reader, m *Map) error {
	// Read number of areas
	var areaCount int32
	if err := binary.Read(reader, binary.BigEndian, &areaCount); err != nil {
		return fmt.Errorf("reading area count: %w", err)
	}

	// Read each area
	for i := int32(0); i < areaCount; i++ {
		area := &Area{}

		// Read area ID
		if err := binary.Read(reader, binary.BigEndian, &area.ID); err != nil {
			return fmt.Errorf("reading area ID: %w", err)
		}

		// Read area name
		nameLength, err := reader.ReadByte()
		if err != nil {
			return fmt.Errorf("reading area name length: %w", err)
		}

		if nameLength > 0 {
			nameBytes := make([]byte, nameLength)
			if _, err := io.ReadFull(reader, nameBytes); err != nil {
				return fmt.Errorf("reading area name: %w", err)
			}
			area.Name = string(nameBytes)
		}

		// Store the area
		m.Areas[area.ID] = area
	}

	return nil
}

// parseAreasNew reads areaNames QMap<int,QString> at the start of MudletMap and fills m.Areas
func parseAreasNew(r *BinaryReader, m *Map) error {
	// We need to walk through version and envColors to areaNames, read it, then rewind reader for later parsers.
	// To keep minimal changes, we'll do a shallow pass with a separate reader copy by reopening the underlying file is not possible here.
	// So, we parse areaNames first using a snapshot of position, then reset by returning an errorless state and let parseRoomsNew call skipMudletMapPrefix again.
	start := r.Position()
	// Use a small helper that works on a throwaway clone by buffering the prefix we consume, then put it back by tracking only positions.
	// Since we cannot unread, we only run this function before parseRoomsNew, and parseRoomsNew assumes fresh reader at offset 0.
	if start != 0 {
		// If not at start, don't attempt. Keep areas empty.
		return nil
	}
	// version
	if _, err := r.ReadInt32(); err != nil { return err }
	// envColors
	if err := skipQMapIntInt(r); err != nil { return err }
	// areaNames count
	sz, err := r.ReadInt32()
	if err != nil { return err }
	for i:=0;i<int(sz);i++ {
		key, err := r.ReadInt32(); if err != nil { return err }
		name, err := r.ReadQString(); if err != nil { return err }
		m.Areas[int32(key)] = &Area{ID:int32(key), Name:name}
	}
	return nil
}

// parseEnvironments reads all environments from the map file (old format)
func parseEnvironments(reader *bufio.Reader, m *Map) error {
	// Read number of environments
	var envCount int32
	if err := binary.Read(reader, binary.BigEndian, &envCount); err != nil {
		return fmt.Errorf("reading environment count: %w", err)
	}

	// Read each environment
	m.Environments = make([]Environment, envCount)
	for i := int32(0); i < envCount; i++ {
		env := Environment{}

		// Read environment name
		nameLength, err := reader.ReadByte()
		if err != nil {
			return fmt.Errorf("reading environment name length: %w", err)
		}

		if nameLength > 0 {
			nameBytes := make([]byte, nameLength)
			if _, err := io.ReadFull(reader, nameBytes); err != nil {
				return fmt.Errorf("reading environment name: %w", err)
			}
			env.Name = string(nameBytes)
		}

		// Read environment color
		if err := binary.Read(reader, binary.BigEndian, &env.Color); err != nil {
			return fmt.Errorf("reading environment color: %w", err)
		}

		m.Environments[i] = env
	}

	return nil
}

// parseEnvironmentsNew reads all environments from the map file (new format with BinaryReader)
func parseEnvironmentsNew(reader *BinaryReader, m *Map) error {
	// Read number of environments
	envCount, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("reading environment count: %w", err)
	}

	// Validate environment count to prevent excessive processing
	const maxEnvCount = 1000 // A reasonable maximum number of environments
	if envCount < 0 || envCount > maxEnvCount {
		return fmt.Errorf("invalid environment count: %d (must be between 0 and %d)", envCount, maxEnvCount)
	}

	fmt.Printf("Found %d environments\n", envCount)

	// Read each environment
	m.Environments = make([]Environment, envCount)
	for i := int32(0); i < envCount; i++ {
		env := Environment{}

		// Read environment name (UTF-16BE string)
  env.Name, err = reader.ReadQString()
		if err != nil {
			return fmt.Errorf("reading environment name for environment %d: %w", i, err)
		}

		// Read environment color
		env.Color, err = reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("reading environment color for environment %d: %w", i, err)
		}

		m.Environments[i] = env
		fmt.Printf("Environment %d: %s\n", i, env.Name)
	}

	return nil
}

// parseCustomLines reads custom lines from the map file (old format, version 2+)
func parseCustomLines(reader *bufio.Reader, m *Map) error {
	// Read number of custom lines
	var lineCount int32
	if err := binary.Read(reader, binary.BigEndian, &lineCount); err != nil {
		return fmt.Errorf("reading custom line count: %w", err)
	}

	// Read each custom line
	m.CustomLines = make([]CustomLine, lineCount)
	for i := int32(0); i < lineCount; i++ {
		line := CustomLine{}

		// Read coordinates
		if err := binary.Read(reader, binary.BigEndian, &line.X1); err != nil {
			return fmt.Errorf("reading line X1: %w", err)
		}
		if err := binary.Read(reader, binary.BigEndian, &line.Y1); err != nil {
			return fmt.Errorf("reading line Y1: %w", err)
		}
		if err := binary.Read(reader, binary.BigEndian, &line.Z1); err != nil {
			return fmt.Errorf("reading line Z1: %w", err)
		}
		if err := binary.Read(reader, binary.BigEndian, &line.X2); err != nil {
			return fmt.Errorf("reading line X2: %w", err)
		}
		if err := binary.Read(reader, binary.BigEndian, &line.Y2); err != nil {
			return fmt.Errorf("reading line Y2: %w", err)
		}
		if err := binary.Read(reader, binary.BigEndian, &line.Z2); err != nil {
			return fmt.Errorf("reading line Z2: %w", err)
		}

		// Read line properties
		if err := binary.Read(reader, binary.BigEndian, &line.Color); err != nil {
			return fmt.Errorf("reading line color: %w", err)
		}

		styleByte, err := reader.ReadByte()
		if err != nil {
			return fmt.Errorf("reading line style: %w", err)
		}
		line.Style = int8(styleByte)

		widthByte, err := reader.ReadByte()
		if err != nil {
			return fmt.Errorf("reading line width: %w", err)
		}
		line.Width = int8(widthByte)

		m.CustomLines[i] = line
	}

	return nil
}

// parseCustomLinesNew reads custom lines from the map file (new format with BinaryReader, version 2+)
func parseCustomLinesNew(reader *BinaryReader, m *Map) error {
	// Read number of custom lines
	lineCount, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("reading custom line count: %w", err)
	}

	// Validate line count to prevent excessive processing
	const maxLineCount = 10000 // A reasonable maximum number of custom lines
	if lineCount < 0 || lineCount > maxLineCount {
		return fmt.Errorf("invalid custom line count: %d (must be between 0 and %d)", lineCount, maxLineCount)
	}

	fmt.Printf("Found %d custom lines\n", lineCount)

	// Read each custom line
	m.CustomLines = make([]CustomLine, lineCount)
	for i := int32(0); i < lineCount; i++ {
		line := CustomLine{}

		// Read coordinates
		line.X1, err = reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("reading line X1 for line %d: %w", i, err)
		}
		line.Y1, err = reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("reading line Y1 for line %d: %w", i, err)
		}
		line.Z1, err = reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("reading line Z1 for line %d: %w", i, err)
		}
		line.X2, err = reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("reading line X2 for line %d: %w", i, err)
		}
		line.Y2, err = reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("reading line Y2 for line %d: %w", i, err)
		}
		line.Z2, err = reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("reading line Z2 for line %d: %w", i, err)
		}

		// Read line properties
		line.Color, err = reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("reading line color for line %d: %w", i, err)
		}

		styleByte, err := reader.ReadInt8()
		if err != nil {
			return fmt.Errorf("reading line style for line %d: %w", i, err)
		}
		line.Style = styleByte

		widthByte, err := reader.ReadInt8()
		if err != nil {
			return fmt.Errorf("reading line width for line %d: %w", i, err)
		}
		line.Width = widthByte

		m.CustomLines[i] = line

		// Print progress every 1000 lines
		if i > 0 && i%1000 == 0 {
			fmt.Printf("Parsed %d custom lines...\n", i)
		}
	}

	return nil
}

// parseLabels reads labels from the map file (old format, version 3+)
func parseLabels(reader *bufio.Reader, m *Map) error {
	// Read number of labels
	var labelCount int32
	if err := binary.Read(reader, binary.BigEndian, &labelCount); err != nil {
		return fmt.Errorf("reading label count: %w", err)
	}

	// Read each label
	m.Labels = make([]Label, labelCount)
	for i := int32(0); i < labelCount; i++ {
		label := Label{}

		// Read coordinates
		if err := binary.Read(reader, binary.BigEndian, &label.X); err != nil {
			return fmt.Errorf("reading label X: %w", err)
		}
		if err := binary.Read(reader, binary.BigEndian, &label.Y); err != nil {
			return fmt.Errorf("reading label Y: %w", err)
		}
		if err := binary.Read(reader, binary.BigEndian, &label.Z); err != nil {
			return fmt.Errorf("reading label Z: %w", err)
		}

		// Read text
		textLength, err := reader.ReadByte()
		if err != nil {
			return fmt.Errorf("reading label text length: %w", err)
		}

		if textLength > 0 {
			textBytes := make([]byte, textLength)
			if _, err := io.ReadFull(reader, textBytes); err != nil {
				return fmt.Errorf("reading label text: %w", err)
			}
			label.Text = string(textBytes)
		}

		// Read label properties
		if err := binary.Read(reader, binary.BigEndian, &label.Color); err != nil {
			return fmt.Errorf("reading label color: %w", err)
		}

		sizeByte, err := reader.ReadByte()
		if err != nil {
			return fmt.Errorf("reading label size: %w", err)
		}
		label.Size = int8(sizeByte)

		bgByte, err := reader.ReadByte()
		if err != nil {
			return fmt.Errorf("reading label background flag: %w", err)
		}
		label.ShowBackground = bgByte != 0

		m.Labels[i] = label
	}

	return nil
}

// parseLabelsNew reads labels from the map file (new format with BinaryReader, version 3+)
func parseLabelsNew(reader *BinaryReader, m *Map) error {
	// Read number of labels
	labelCount, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("reading label count: %w", err)
	}

	// Validate label count to prevent excessive processing
	const maxLabelCount = 10000 // A reasonable maximum number of labels
	if labelCount < 0 || labelCount > maxLabelCount {
		return fmt.Errorf("invalid label count: %d (must be between 0 and %d)", labelCount, maxLabelCount)
	}

	fmt.Printf("Found %d labels\n", labelCount)

	// Read each label
	m.Labels = make([]Label, labelCount)
	for i := int32(0); i < labelCount; i++ {
		label := Label{}

		// Read coordinates
		label.X, err = reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("reading label X for label %d: %w", i, err)
		}
		label.Y, err = reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("reading label Y for label %d: %w", i, err)
		}
		label.Z, err = reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("reading label Z for label %d: %w", i, err)
		}

		// Read text (UTF-16BE string)
  label.Text, err = reader.ReadQString()
		if err != nil {
			return fmt.Errorf("reading label text for label %d: %w", i, err)
		}

		// Read label properties
		label.Color, err = reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("reading label color for label %d: %w", i, err)
		}

		sizeByte, err := reader.ReadInt8()
		if err != nil {
			return fmt.Errorf("reading label size for label %d: %w", i, err)
		}
		label.Size = sizeByte

		bgBool, err := reader.ReadBool()
		if err != nil {
			return fmt.Errorf("reading label background flag for label %d: %w", i, err)
		}
		label.ShowBackground = bgBool

		m.Labels[i] = label

		// Print progress every 1000 labels
		if i > 0 && i%1000 == 0 {
			fmt.Printf("Parsed %d labels...\n", i)
		}
	}

	return nil
}

// ValidateMap checks the map for integrity issues
func ValidateMap(m *Map) []ValidationError {
	var errors []ValidationError

	// Check for rooms with invalid exits
	for roomID, room := range m.Rooms {
		for _, exit := range room.Exits {
			if exit.TargetID > 0 {
				if _, exists := m.Rooms[exit.TargetID]; !exists {
					errors = append(errors, ValidationError{
						Type:    "InvalidExit",
						Message: fmt.Sprintf("Room %d has exit to non-existent room %d", roomID, exit.TargetID),
						RoomID:  roomID,
					})
				}
			}
		}
	}

	// Check for rooms with invalid environments
	for roomID, room := range m.Rooms {
		if room.Environment >= int32(len(m.Environments)) {
			errors = append(errors, ValidationError{
				Type:    "InvalidEnvironment",
				Message: fmt.Sprintf("Room %d has invalid environment ID %d", roomID, room.Environment),
				RoomID:  roomID,
			})
		}
	}

	return errors
}

// ExportToJSON exports the map to a JSON file
func ExportToJSON(m *Map, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("creating JSON file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(m); err != nil {
		return fmt.Errorf("encoding map to JSON: %w", err)
	}

	return nil
}

// GetMapStats returns statistics about the map
func GetMapStats(m *Map) MapStats {
	stats := MapStats{
		TotalRooms:        len(m.Rooms),
		TotalAreas:        len(m.Areas),
		TotalEnvironments: len(m.Environments),
		ZLevels:           []int32{},
	}

	// Find bounding box and Z levels
	zLevelsMap := make(map[int32]bool)

	if len(m.Rooms) > 0 {
		// Initialize with first room
		var firstRoom *Room
		for _, room := range m.Rooms {
			firstRoom = room
			break
		}

		stats.BoundingBox = BoundingBox{
			MinX: firstRoom.X,
			MaxX: firstRoom.X,
			MinY: firstRoom.Y,
			MaxY: firstRoom.Y,
			MinZ: firstRoom.Z,
			MaxZ: firstRoom.Z,
		}

		// Track all Z levels and update bounding box
		for _, room := range m.Rooms {
			// Update bounding box
			if room.X < stats.BoundingBox.MinX {
				stats.BoundingBox.MinX = room.X
			}
			if room.X > stats.BoundingBox.MaxX {
				stats.BoundingBox.MaxX = room.X
			}
			if room.Y < stats.BoundingBox.MinY {
				stats.BoundingBox.MinY = room.Y
			}
			if room.Y > stats.BoundingBox.MaxY {
				stats.BoundingBox.MaxY = room.Y
			}
			if room.Z < stats.BoundingBox.MinZ {
				stats.BoundingBox.MinZ = room.Z
			}
			if room.Z > stats.BoundingBox.MaxZ {
				stats.BoundingBox.MaxZ = room.Z
			}

			// Track Z level
			zLevelsMap[room.Z] = true
		}
	}

	// Convert Z levels map to slice
	for z := range zLevelsMap {
		stats.ZLevels = append(stats.ZLevels, z)
	}

	return stats
}
