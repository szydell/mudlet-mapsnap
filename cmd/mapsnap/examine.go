package main

import (
	"fmt"
	"os"

	"github.com/szydell/arkadia-mapsnap/pkg/mapparser"
)

// ExamineFile examines a binary map file and walks through its Qt/MudletMap structure,
// logging offsets and sizes for each section.
func ExamineFile(filename string) error {
	// Get file info for size
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

	r := mapparser.NewBinaryReader(f)
	log := func(section string) {
		fmt.Printf("@%d: %s\n", r.Position(), section)
	}

	// version (qint32)
	log("Begin MudletMap.version (qint32)")
	version, err := r.ReadInt32()
	if err != nil {
		return fmt.Errorf("version: %w", err)
	}
	fmt.Printf("    version = %d\n", version)
	log("After version")

	// envColors: QMap<int,int>
	log("Begin envColors QMap<int,int>")
	if err := examineQMapIntInt(r); err != nil {
		return fmt.Errorf("envColors: %w", err)
	}
	log("After envColors")

	// areaNames: QMap<int, QString>
	log("Begin areaNames QMap<int,QString>")
	if err := examineQMapIntQString(r); err != nil {
		return fmt.Errorf("areaNames: %w", err)
	}
	log("After areaNames")

	// mCustomEnvColors: QMap<int,QColor>
	log("Begin mCustomEnvColors QMap<int,QColor>")
	if err := examineQMapIntQColor(r); err != nil {
		return fmt.Errorf("mCustomEnvColors: %w", err)
	}
	log("After mCustomEnvColors")

	// mpRoomDbHashToRoomId: QMap<QString,QUInt>
	log("Begin mpRoomDbHashToRoomId QMap<QString,QUInt>")
	if err := examineQMapQStringUInt(r); err != nil {
		return fmt.Errorf("mpRoomDbHashToRoomId: %w", err)
	}
	log("After mpRoomDbHashToRoomId")

	// mUserData: QMap<QString,QString>
	log("Begin mUserData QMap<QString,QString>")
	if err := examineQMapQStringQString(r); err != nil {
		return fmt.Errorf("mUserData: %w", err)
	}
	log("After mUserData")

	// mapSymbolFont: QFont
	log("Begin mapSymbolFont QFont")
	if err := examineQFont(r); err != nil {
		return fmt.Errorf("mapSymbolFont: %w", err)
	}
	log("After mapSymbolFont")

	// mapFontFudgeFactor: double
	log("Begin mapFontFudgeFactor (double)")
	fudge, err := r.ReadDouble()
	if err != nil {
		return fmt.Errorf("mapFontFudgeFactor: %w", err)
	}
	fmt.Printf("    mapFontFudgeFactor = %f\n", fudge)
	log("After mapFontFudgeFactor")

	// useOnlyMapFont: bool
	log("Begin useOnlyMapFont (bool)")
	useOnly, err := r.ReadBool()
	if err != nil {
		return fmt.Errorf("useOnlyMapFont: %w", err)
	}
	fmt.Printf("    useOnlyMapFont = %v\n", useOnly)
	log("After useOnlyMapFont")

	// areas: MudletAreas
	log("Begin areas MudletAreas")
	if err := examineMudletAreas(r); err != nil {
		return fmt.Errorf("areas: %w", err)
	}
	log("After areas")

	// mRoomIdHash: QMap<QString,QInt>
	log("Begin mRoomIdHash QMap<QString,QInt>")
	if err := examineQMapQStringInt(r); err != nil {
		return fmt.Errorf("mRoomIdHash: %w", err)
	}
	log("After mRoomIdHash")

	// labels: MudletLabels
	log("Begin labels MudletLabels")
	if err := examineMudletLabels(r); err != nil {
		return fmt.Errorf("labels: %w", err)
	}
	log("After labels")

	fmt.Printf("\nRooms section should start around offset %d\n", r.Position())
	return nil
}

func examineQMapIntInt(r *mapparser.BinaryReader) error {
	sz, err := r.ReadInt32()
	if err != nil {
		return err
	}
	fmt.Printf("    count = %d\n", sz)
	for i := 0; i < int(sz); i++ {
		if _, err := r.ReadInt32(); err != nil {
			return err
		}
		if _, err := r.ReadInt32(); err != nil {
			return err
		}
	}
	return nil
}

func examineQMapIntQString(r *mapparser.BinaryReader) error {
	sz, err := r.ReadInt32()
	if err != nil {
		return err
	}
	fmt.Printf("    count = %d\n", sz)
	for i := 0; i < int(sz); i++ {
		fmt.Printf("    entry %d @%d begin\n", i, r.Position())
		key, err := r.ReadInt32()
		if err != nil {
			return err
		}
		peek, _ := r.Peek(8)
		if len(peek) >= 8 {
			fmt.Printf("      key=%d next8=%02x %02x %02x %02x %02x %02x %02x %02x @%d\n",
				key, peek[0], peek[1], peek[2], peek[3], peek[4], peek[5], peek[6], peek[7], r.Position())
		}
		// Manually read QString for instrumentation
		lenPeek, _ := r.Peek(4)
		var byteLen uint32 = uint32(lenPeek[0])<<24 | uint32(lenPeek[1])<<16 | uint32(lenPeek[2])<<8 | uint32(lenPeek[3])
		fmt.Printf("      QString byteLen (peek)=%d @%d\n", byteLen, r.Position())
		str, err := r.ReadQString()
		if err != nil {
			return err
		}
		fmt.Printf("      QString value='%s' after QString @%d\n", str, r.Position())
	}
	return nil
}

func examineQMapIntQColor(r *mapparser.BinaryReader) error {
	sz, err := r.ReadInt32()
	if err != nil {
		return err
	}
	fmt.Printf("    count = %d\n", sz)
	for i := 0; i < int(sz); i++ {
		if _, err := r.ReadInt32(); err != nil {
			return err
		}
		if err := examineQColor(r); err != nil {
			return err
		}
	}
	return nil
}

func examineQMapQStringUInt(r *mapparser.BinaryReader) error {
	sz, err := r.ReadInt32()
	if err != nil {
		return err
	}
	fmt.Printf("    count = %d\n", sz)
	for i := 0; i < int(sz); i++ {
		if _, err := r.ReadQString(); err != nil {
			return err
		}
		if _, err := r.ReadUInt32(); err != nil {
			return err
		}
	}
	return nil
}

func examineQMapQStringQString(r *mapparser.BinaryReader) error {
	sz, err := r.ReadInt32()
	if err != nil {
		return err
	}
	fmt.Printf("    count = %d\n", sz)
	for i := 0; i < int(sz); i++ {
		if _, err := r.ReadQString(); err != nil {
			return err
		}
		if _, err := r.ReadQString(); err != nil {
			return err
		}
	}
	return nil
}

func examineQMapQStringInt(r *mapparser.BinaryReader) error {
	sz, err := r.ReadInt32()
	if err != nil {
		return err
	}
	fmt.Printf("    count = %d\n", sz)
	for i := 0; i < int(sz); i++ {
		if _, err := r.ReadQString(); err != nil {
			return err
		}
		if _, err := r.ReadInt32(); err != nil {
			return err
		}
	}
	return nil
}

func examineQColor(r *mapparser.BinaryReader) error {
	if _, err := r.ReadInt8(); err != nil {
		return err
	}
	for i := 0; i < 5; i++ {
		if _, err := r.ReadUInt16(); err != nil {
			return err
		}
	}
	return nil
}

func examineQFont(r *mapparser.BinaryReader) error {
	if _, err := r.ReadQString(); err != nil {
		return err
	}
	if _, err := r.ReadQString(); err != nil {
		return err
	}
	if _, err := r.ReadDouble(); err != nil {
		return err
	}
	if _, err := r.ReadInt32(); err != nil {
		return err
	}
	if _, err := r.ReadInt8(); err != nil {
		return err
	}
	if _, err := r.ReadUInt16(); err != nil {
		return err
	}
	if _, err := r.ReadByte(); err != nil {
		return err
	}
	if _, err := r.ReadInt8(); err != nil {
		return err
	}
	if _, err := r.ReadInt8(); err != nil {
		return err
	}
	if _, err := r.ReadUInt16(); err != nil {
		return err
	}
	if _, err := r.ReadInt8(); err != nil {
		return err
	}
	if _, err := r.ReadInt32(); err != nil {
		return err
	}
	if _, err := r.ReadInt32(); err != nil {
		return err
	}
	if _, err := r.ReadInt8(); err != nil {
		return err
	}
	if _, err := r.ReadInt8(); err != nil {
		return err
	}
	return nil
}

func examineQListUInt(r *mapparser.BinaryReader) error {
	sz, err := r.ReadInt32()
	if err != nil {
		return err
	}
	for i := 0; i < int(sz); i++ {
		if _, err := r.ReadUInt32(); err != nil {
			return err
		}
	}
	return nil
}

func examineQListInt(r *mapparser.BinaryReader) error {
	sz, err := r.ReadInt32()
	if err != nil {
		return err
	}
	for i := 0; i < int(sz); i++ {
		if _, err := r.ReadInt32(); err != nil {
			return err
		}
	}
	return nil
}

func examineQVector(r *mapparser.BinaryReader) error {
	for i := 0; i < 3; i++ {
		if _, err := r.ReadDouble(); err != nil {
			return err
		}
	}
	return nil
}

func examineMudletArea(r *mapparser.BinaryReader) error {
	if err := examineQListUInt(r); err != nil {
		return err
	}
	if err := examineQListInt(r); err != nil {
		return err
	}
	sz, err := r.ReadInt32()
	if err != nil {
		return err
	}
	for i := 0; i < int(sz); i++ {
		if _, err := r.ReadInt32(); err != nil {
			return err
		}
		if _, err := r.ReadInt32(); err != nil {
			return err
		}
		if _, err := r.ReadInt32(); err != nil {
			return err
		}
	}
	if _, err := r.ReadBool(); err != nil {
		return err
	}
	for i := 0; i < 6; i++ {
		if _, err := r.ReadInt32(); err != nil {
			return err
		}
	}
	if err := examineQVector(r); err != nil {
		return err
	}
	for i := 0; i < 4; i++ {
		if err := examineQMapIntInt(r); err != nil {
			return err
		}
	}
	if err := examineQVector(r); err != nil {
		return err
	}
	if _, err := r.ReadBool(); err != nil {
		return err
	}
	if _, err := r.ReadInt32(); err != nil {
		return err
	}
	if err := examineQMapQStringQString(r); err != nil {
		return err
	}
	return nil
}

func examineMudletAreas(r *mapparser.BinaryReader) error {
	sz, err := r.ReadInt32()
	if err != nil {
		return err
	}
	fmt.Printf("    count = %d\n", sz)
	for i := 0; i < int(sz); i++ {
		areaID, err := r.ReadInt32()
		if err != nil {
			return err
		}
		fmt.Printf("    area %d (id=%d) @%d\n", i, areaID, r.Position())
		if err := examineMudletArea(r); err != nil {
			return err
		}
	}
	return nil
}

var labelDebugCount int

func examineMudletLabel(r *mapparser.BinaryReader) error {
	if _, err := r.ReadInt32(); err != nil {
		return err
	}
	for i := 0; i < 3; i++ {
		if _, err := r.ReadDouble(); err != nil {
			return err
		}
	}
	// dummy1, dummy2
	for i := 0; i < 2; i++ {
		if _, err := r.ReadDouble(); err != nil {
			return err
		}
	}
	// size: QPair<double,double>
	for i := 0; i < 2; i++ {
		if _, err := r.ReadDouble(); err != nil {
			return err
		}
	}
	if labelDebugCount < 3 {
		if peek, _ := r.Peek(8); len(peek) == 8 {
			fmt.Printf("      pre-QString next8=%02x %02x %02x %02x %02x %02x %02x %02x @%d\n",
				peek[0], peek[1], peek[2], peek[3], peek[4], peek[5], peek[6], peek[7], r.Position())
		}
	}
	str, err := r.ReadQString()
	if err != nil {
		return err
	}
	if labelDebugCount < 3 {
		fmt.Printf("      label text='%s' @%d\n", str, r.Position())
	}
	labelDebugCount++
	if err := examineQColor(r); err != nil {
		return err
	}
	if err := examineQColor(r); err != nil {
		return err
	}
	// QPixMap: header marker (uint32 already acts as presence/size), then maybe PNG magic in next 4 bytes
	_, _ = r.ReadUInt32()
	if sig, _ := r.Peek(4); len(sig) == 4 {
		if uint32(sig[0])<<24|uint32(sig[1])<<16|uint32(sig[2])<<8|uint32(sig[3]) == 0x89504e47 {
			if err := examineSkipPNG(r); err != nil {
				return err
			}
		}
	}
	if _, err := r.ReadBool(); err != nil {
		return err
	}
	if _, err := r.ReadBool(); err != nil {
		return err
	}
	if labelDebugCount <= 3 {
		if peek, _ := r.Peek(8); len(peek) == 8 {
			fmt.Printf("      after-label peek next8=%02x %02x %02x %02x %02x %02x %02x %02x @%d\n",
				peek[0], peek[1], peek[2], peek[3], peek[4], peek[5], peek[6], peek[7], r.Position())
		}
	}
	return nil
}

// examineSkipPNG scans until it sees the PNG IEND chunk marker and consumes it.
func examineSkipPNG(r *mapparser.BinaryReader) error {
	needle := []byte{0x49, 0x45, 0x4e, 0x44} // 'I','E','N','D'
	for {
		peek, err := r.Peek(4)
		if err != nil || len(peek) < 4 {
			return err
		}
		if peek[0] == needle[0] && peek[1] == needle[1] && peek[2] == needle[2] && peek[3] == needle[3] {
			// consume 'IEND' + 4-byte CRC
			if err := r.Skip(8); err != nil {
				return err
			}
			return nil
		}
		if _, err := r.ReadByte(); err != nil {
			return err
		}
	}
}

func examineMudletLabels(r *mapparser.BinaryReader) error {
	sz, err := r.ReadInt32()
	if err != nil {
		return err
	}
	fmt.Printf("    count = %d (number of areas with labels)\n", sz)
	for i := 0; i < int(sz); i++ {
		total, err := r.ReadInt32()
		if err != nil {
			return err
		}
		areaID, err := r.ReadInt32()
		if err != nil {
			return err
		}
		fmt.Printf("    area %d: areaID=%d, labels=%d @%d\n", i, areaID, total, r.Position())
		for j := 0; j < int(total); j++ {
			if err := examineMudletLabel(r); err != nil {
				return err
			}
		}
	}
	return nil
}
