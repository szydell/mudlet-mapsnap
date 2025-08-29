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
		if _, err := r.ReadQString(); err != nil { return err }
		fmt.Printf("      after QString @%d\n", r.Position())
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
func exQtMudletLabel(r *mapparser.BinaryReader) error {
	if _, err := r.ReadInt32(); err != nil { return err }
	for i:=0;i<3;i++ { if _, err := r.ReadDouble(); err != nil { return err } }
	if _, err := r.ReadDouble(); err != nil { return err }
	if _, err := r.ReadDouble(); err != nil { return err }
	if _, err := r.ReadDouble(); err != nil { return err }
	if _, err := r.ReadDouble(); err != nil { return err }
	if _, err := r.ReadQString(); err != nil { return err }
	if err := exQtQColor(r); err != nil { return err }
	if err := exQtQColor(r); err != nil { return err }
	if _, err := r.ReadUInt32(); err != nil { return err }
	b1, _ := r.ReadUInt32(); if b1 == 0x89504e47 { for { ch, err := r.ReadUInt32(); if err != nil { return err }; if ch == 0x49454e44 { break } } }
	if _, err := r.ReadBool(); err != nil { return err }
	if _, err := r.ReadBool(); err != nil { return err }
	return nil
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
