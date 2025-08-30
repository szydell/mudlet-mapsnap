package main

import (
	"fmt"
	"os"

	"github.com/szydell/arkadia-mapsnap/pkg/mapparser"
)

// ExamineQt walks through MudletMap (Qt QDataStream) structure and logs offsets and sizes.
func ExamineQt(filename string) error {
	f, err := os.Open(filename)
	if err != nil { return err }
	defer f.Close()

	r := mapparser.NewBinaryReader(f)
	log := func(section string) {
		fmt.Printf("@%d: %s\n", r.Position(), section)
	}

	// version (qint32)
	log("Begin MudletMap.version (qint32)")
	if _, err := r.ReadInt32(); err != nil { return fmt.Errorf("version: %w", err) }
	log("After version")

	// envColors: QMap<int,int>
	log("Begin envColors QMap<int,int>")
	if err := exQtQMapIntInt(r); err != nil { return fmt.Errorf("envColors: %w", err) }
	log("After envColors")

	// areaNames: QMap<int, QString>
	log("Begin areaNames QMap<int,QString>")
	if err := exQtQMapIntQString(r); err != nil { return fmt.Errorf("areaNames: %w", err) }
	log("After areaNames")

	// mCustomEnvColors: QMap<int,QColor>
	log("Begin mCustomEnvColors QMap<int,QColor>")
	if err := exQtQMapIntQColor(r); err != nil { return fmt.Errorf("mCustomEnvColors: %w", err) }
	log("After mCustomEnvColors")

	// mpRoomDbHashToRoomId: QMap<QString,QUInt>
	log("Begin mpRoomDbHashToRoomId QMap<QString,QUInt>")
	if err := exQtQMapQStringUInt(r); err != nil { return fmt.Errorf("mpRoomDbHashToRoomId: %w", err) }
	log("After mpRoomDbHashToRoomId")

	// mUserData: QMap<QString,QString>
	log("Begin mUserData QMap<QString,QString>")
	if err := exQtQMapQStringQString(r); err != nil { return fmt.Errorf("mUserData: %w", err) }
	log("After mUserData")

	// mapSymbolFont: QFont
	log("Begin mapSymbolFont QFont")
	if err := exQtQFont(r); err != nil { return fmt.Errorf("mapSymbolFont: %w", err) }
	log("After mapSymbolFont")

	// mapFontFudgeFactor: double
	log("Begin mapFontFudgeFactor (double)")
	if _, err := r.ReadDouble(); err != nil { return fmt.Errorf("mapFontFudgeFactor: %w", err) }
	log("After mapFontFudgeFactor")

	// useOnlyMapFont: bool
	log("Begin useOnlyMapFont (bool)")
	if _, err := r.ReadBool(); err != nil { return fmt.Errorf("useOnlyMapFont: %w", err) }
	log("After useOnlyMapFont")

	// areas: MudletAreas
	log("Begin areas MudletAreas")
	if err := exQtMudletAreas(r); err != nil { return fmt.Errorf("areas: %w", err) }
	log("After areas")

	// mRoomIdHash: QMap<QString,QInt>
	log("Begin mRoomIdHash QMap<QString,QInt>")
	if err := exQtQMapQStringInt(r); err != nil { return fmt.Errorf("mRoomIdHash: %w", err) }
	log("After mRoomIdHash")

	// labels: MudletLabels
	log("Begin labels MudletLabels")
	if err := exQtMudletLabels(r); err != nil { return fmt.Errorf("labels: %w", err) }
	log("After labels")

	fmt.Printf("Rooms section should start around offset %d\n", r.Position())
	return nil
}

func exQtQMapIntInt(r *mapparser.BinaryReader) error {
	sz, err := r.ReadInt32(); if err != nil { return err }
	for i:=0;i<int(sz);i++ { if _, err := r.ReadInt32(); err != nil { return err }; if _, err := r.ReadInt32(); err != nil { return err } }
	return nil
}
func exQtQMapIntQString(r *mapparser.BinaryReader) error {
	// Read count (QUInt)
	sz, err := r.ReadInt32(); if err != nil { return err }
	fmt.Printf("    areaNames count=%d\n", sz)
	for i:=0;i<int(sz);i++ {
		fmt.Printf("    entry %d @%d begin\n", i, r.Position())
		key, err := r.ReadInt32(); if err != nil { return err }
		peek, _ := r.Peek(8)
		if len(peek) >= 8 {
			fmt.Printf("      key=%d next8=%02x %02x %02x %02x %02x %02x %02x %02x @%d\n", key, peek[0],peek[1],peek[2],peek[3],peek[4],peek[5],peek[6],peek[7], r.Position())
		}
		// Manually read QString for instrumentation
		// Read quint32 byte length
		lenPeek, _ := r.Peek(4)
		var byteLen uint32 = uint32(lenPeek[0])<<24 | uint32(lenPeek[1])<<16 | uint32(lenPeek[2])<<8 | uint32(lenPeek[3])
		fmt.Printf("      QString byteLen (peek)=%d @%d\n", byteLen, r.Position())
		str, err := r.ReadQString(); if err != nil { return err }
		fmt.Printf("      QString value='%s' after QString @%d\n", str, r.Position())
	}
	return nil
}

func exQtQMapIntQColor(r *mapparser.BinaryReader) error {
	sz, err := r.ReadInt32(); if err != nil { return err }
	for i:=0;i<int(sz);i++ { if _, err := r.ReadInt32(); err != nil { return err }; if err := exQtQColor(r); err != nil { return err } }
	return nil
}
func exQtQMapQStringUInt(r *mapparser.BinaryReader) error {
	sz, err := r.ReadInt32(); if err != nil { return err }
	for i:=0;i<int(sz);i++ { if _, err := r.ReadQString(); err != nil { return err }; if _, err := r.ReadUInt32(); err != nil { return err } }
	return nil
}
func exQtQMapQStringQString(r *mapparser.BinaryReader) error {
	sz, err := r.ReadInt32(); if err != nil { return err }
	for i:=0;i<int(sz);i++ { if _, err := r.ReadQString(); err != nil { return err }; if _, err := r.ReadQString(); err != nil { return err } }
	return nil
}
func exQtQMapQStringInt(r *mapparser.BinaryReader) error {
	sz, err := r.ReadInt32(); if err != nil { return err }
	for i:=0;i<int(sz);i++ { if _, err := r.ReadQString(); err != nil { return err }; if _, err := r.ReadInt32(); err != nil { return err } }
	return nil
}
func exQtQColor(r *mapparser.BinaryReader) error {
	if _, err := r.ReadInt8(); err != nil { return err }
	for i:=0;i<5;i++ { if _, err := r.ReadUInt16(); err != nil { return err } }
	return nil
}
func exQtQFont(r *mapparser.BinaryReader) error {
	if _, err := r.ReadQString(); err != nil { return err }
	if _, err := r.ReadQString(); err != nil { return err }
	if _, err := r.ReadDouble(); err != nil { return err }
	if _, err := r.ReadInt32(); err != nil { return err }
	if _, err := r.ReadInt8(); err != nil { return err }
	if _, err := r.ReadUInt16(); err != nil { return err }
	if _, err := r.ReadByte(); err != nil { return err }
	if _, err := r.ReadInt8(); err != nil { return err }
	if _, err := r.ReadInt8(); err != nil { return err }
	if _, err := r.ReadUInt16(); err != nil { return err }
	if _, err := r.ReadInt8(); err != nil { return err }
	if _, err := r.ReadInt32(); err != nil { return err }
	if _, err := r.ReadInt32(); err != nil { return err }
	if _, err := r.ReadInt8(); err != nil { return err }
	if _, err := r.ReadInt8(); err != nil { return err }
	return nil
}
func exQtQListUInt(r *mapparser.BinaryReader) error {
	sz, err := r.ReadInt32(); if err != nil { return err }
	for i:=0;i<int(sz);i++ { if _, err := r.ReadUInt32(); err != nil { return err } }
	return nil
}
func exQtQListInt(r *mapparser.BinaryReader) error {
	sz, err := r.ReadInt32(); if err != nil { return err }
	for i:=0;i<int(sz);i++ { if _, err := r.ReadInt32(); err != nil { return err } }
	return nil
}
func exQtQVector(r *mapparser.BinaryReader) error {
	for i:=0;i<3;i++ { if _, err := r.ReadDouble(); err != nil { return err } }
	return nil
}
func exQtMudletArea(r *mapparser.BinaryReader) error {
	if err := exQtQListUInt(r); err != nil { return err }
	if err := exQtQListInt(r); err != nil { return err }
	sz, err := r.ReadInt32(); if err != nil { return err }
	for i:=0;i<int(sz);i++ { if _, err := r.ReadInt32(); err != nil { return err }; if _, err := r.ReadInt32(); err != nil { return err }; if _, err := r.ReadInt32(); err != nil { return err } }
	if _, err := r.ReadBool(); err != nil { return err }
	for i:=0;i<6;i++ { if _, err := r.ReadInt32(); err != nil { return err } }
	if err := exQtQVector(r); err != nil { return err }
	for i:=0;i<4;i++ { if err := exQtQMapIntInt(r); err != nil { return err } }
	if err := exQtQVector(r); err != nil { return err }
	if _, err := r.ReadBool(); err != nil { return err }
	if _, err := r.ReadInt32(); err != nil { return err }
	if err := exQtQMapQStringQString(r); err != nil { return err }
	return nil
}
func exQtMudletAreas(r *mapparser.BinaryReader) error {
	sz, err := r.ReadInt32(); if err != nil { return err }
	for i:=0;i<int(sz);i++ { if _, err := r.ReadInt32(); err != nil { return err }; if err := exQtMudletArea(r); err != nil { return err } }
	return nil
}
var exQtLabelDebugCount int
func exQtMudletLabel(r *mapparser.BinaryReader) error {
	if _, err := r.ReadInt32(); err != nil { return err }
	for i:=0;i<3;i++ { if _, err := r.ReadDouble(); err != nil { return err } }
	// dummy1, dummy2
 for i:=0;i<2;i++ { if _, err := r.ReadDouble(); err != nil { return err } }
	// size: QPair<double,double>
	for i:=0;i<2;i++ { if _, err := r.ReadDouble(); err != nil { return err } }
	if exQtLabelDebugCount < 2 {
		if peek, _ := r.Peek(8); len(peek) == 8 {
			fmt.Printf("      pre-QString next8=%02x %02x %02x %02x %02x %02x %02x %02x @%d\n", peek[0],peek[1],peek[2],peek[3],peek[4],peek[5],peek[6],peek[7], r.Position())
		}
	}
	str, err := r.ReadQString(); if err != nil { return err }
	if exQtLabelDebugCount < 2 { fmt.Printf("      label text='%s' @%d\n", str, r.Position()) }
	exQtLabelDebugCount++
	if err := exQtQColor(r); err != nil { return err }
	if err := exQtQColor(r); err != nil { return err }
	// QPixMap: header marker (uint32 already acts as presence/size), then maybe PNG magic in next 4 bytes
	_, _ = r.ReadUInt32()
	if sig, _ := r.Peek(4); len(sig) == 4 {
		if uint32(sig[0])<<24|uint32(sig[1])<<16|uint32(sig[2])<<8|uint32(sig[3]) == 0x89504e47 {
			if err := exQtSkipPNG(r); err != nil { return err }
		}
	}
	if _, err := r.ReadBool(); err != nil { return err }
	if _, err := r.ReadBool(); err != nil { return err }
	if exQtLabelDebugCount <= 3 {
		if peek, _ := r.Peek(8); len(peek) == 8 {
			fmt.Printf("      after-label peek next8=%02x %02x %02x %02x %02x %02x %02x %02x @%d\n", peek[0],peek[1],peek[2],peek[3],peek[4],peek[5],peek[6],peek[7], r.Position())
		}
	}
	return nil
}
// exQtSkipPNG scans until it sees the PNG IEND chunk marker and consumes it.
func exQtSkipPNG(r *mapparser.BinaryReader) error {
	needle := []byte{0x49, 0x45, 0x4e, 0x44} // 'I','E','N','D'
	for {
		peek, err := r.Peek(4)
		if err != nil || len(peek) < 4 { return err }
		if peek[0]==needle[0] && peek[1]==needle[1] && peek[2]==needle[2] && peek[3]==needle[3] {
			// consume 'IEND' + 4-byte CRC
			if err := r.Skip(8); err != nil { return err }
			return nil
		}
		if _, err := r.ReadByte(); err != nil { return err }
	}
}
func exQtMudletLabels(r *mapparser.BinaryReader) error {
	sz, err := r.ReadInt32(); if err != nil { return err }
	for i:=0;i<int(sz);i++ {
		total, err := r.ReadInt32(); if err != nil { return err }
		if _, err := r.ReadInt32(); err != nil { return err }
		for j:=0;j<int(total);j++ { if err := exQtMudletLabel(r); err != nil { return err } }
	}
	return nil
}
