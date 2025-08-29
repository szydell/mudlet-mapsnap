package mapparser

import (
	"bufio"

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
	// 1) envColors QMap<int,int>
 if envCount, err := qt.ReadInt32(); err == nil && envCount >= 0 && envCount < 100000 {
		for i := int32(0); i < envCount; i++ {
			if _, err := qt.ReadInt32(); err != nil { break }
			if _, err := qt.ReadInt32(); err != nil { break }
		}
	}
	// 2) areaNames QMap<int, QString>
	if areaCount, err := qt.ReadInt32(); err == nil && areaCount >= 0 && areaCount < 100000 {
		areas := make(map[int32]*Area, areaCount)
		for i := int32(0); i < areaCount; i++ {
			id, err := qt.ReadInt32()
			if err != nil { break }
			name, err := qt.ReadQString()
			if err != nil { break }
			areas[id] = &Area{ID: id, Name: name}
		}
		if len(areas) > 0 {
			m.Areas = areas
		}
	}

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
		// Read 1-byte version
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
