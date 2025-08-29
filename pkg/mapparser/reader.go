package mapparser

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"unicode/utf16"
)

// BinaryReader provides helper methods for reading binary data
type BinaryReader struct {
	reader *bufio.Reader
	pos    int // approximate position for debugging; not strictly maintained
}

// Position returns current approximate byte offset from the start
func (br *BinaryReader) Position() int {
	return br.pos
}

// NewBinaryReader creates a new BinaryReader
func NewBinaryReader(reader io.Reader) *BinaryReader {
	return &BinaryReader{
		reader: bufio.NewReader(reader),
	}
}

// ReadByte reads a single byte
func (br *BinaryReader) ReadByte() (byte, error) {
	b, err := br.reader.ReadByte()
	if err == nil {
		br.pos++
	}
	return b, err
}

// ReadInt8 reads an int8
func (br *BinaryReader) ReadInt8() (int8, error) {
	b, err := br.ReadByte()
	if err != nil {
		return 0, err
	}
	return int8(b), nil
}

// ReadInt32 reads an int32 in big endian format
func (br *BinaryReader) ReadInt32() (int32, error) {
	var value int32
	err := binary.Read(br.reader, binary.BigEndian, &value)
	if err != nil {
		return 0, err
	}
	br.pos += 4
	return value, nil
}

// ReadString reads a length-prefixed string
func (br *BinaryReader) ReadString() (string, error) {
	// Read string length (1 byte)
	length, err := br.ReadByte()
	if err != nil {
		return "", fmt.Errorf("reading string length: %w", err)
	}

	// If length is 0, return empty string
	if length == 0 {
		return "", nil
	}

	// Read string data
	data := make([]byte, length)
	if _, err := io.ReadFull(br.reader, data); err != nil {
		return "", fmt.Errorf("reading string data: %w", err)
	}

	return string(data), nil
}

// ReadQString reads a Qt QString from QDataStream (Qt 5.x semantics)
// Format: qint32 byteLength; -1 means empty string; otherwise byteLength is number of BYTES of UTF-16BE data
// Then follows `byteLength` bytes (must be divisible by 2) representing 16-bit QChars (UTF-16BE)
func (br *BinaryReader) ReadQString() (string, error) {
	// In Qt5 QDataStream, QString is serialized as quint32 byte length (or 0xFFFFFFFF for null),
	// followed by that many bytes of UTF-16BE data.
	var n uint32
	if err := binary.Read(br.reader, binary.BigEndian, &n); err != nil {
		return "", fmt.Errorf("reading QString length: %w", err)
	}
	br.pos += 4
	if n == 0xFFFFFFFF {
		return "", nil
	}
	if n%2 != 0 || n > 10000000 {
		return "", fmt.Errorf("invalid QString byte length: %d", n)
	}
	units := make([]uint16, int(n/2))
	if err := binary.Read(br.reader, binary.BigEndian, &units); err != nil {
		return "", fmt.Errorf("reading QString data: %w", err)
	}
	br.pos += int(n)
	return string(utf16.Decode(units)), nil
}

// ReadBool reads a boolean value (1 byte, 0 = false, non-zero = true)
func (br *BinaryReader) ReadBool() (bool, error) {
	b, err := br.ReadByte()
	if err != nil {
		return false, err
	}
	return b != 0, nil
}

// ReadUInt16 reads an unsigned 16-bit integer in big endian
func (br *BinaryReader) ReadUInt16() (uint16, error) {
	var value uint16
	err := binary.Read(br.reader, binary.BigEndian, &value)
	if err != nil {
		return 0, err
	}
	br.pos += 2
	return value, nil
}

// ReadUInt32 reads an unsigned 32-bit integer in big endian
func (br *BinaryReader) ReadUInt32() (uint32, error) {
	var value uint32
	err := binary.Read(br.reader, binary.BigEndian, &value)
	if err != nil {
		return 0, err
	}
	br.pos += 4
	return value, nil
}

// ReadDouble reads an IEEE754 float64 in big endian
func (br *BinaryReader) ReadDouble() (float64, error) {
	var bits uint64
	err := binary.Read(br.reader, binary.BigEndian, &bits)
	if err != nil {
		return 0, err
	}
	br.pos += 8
	return math.Float64frombits(bits), nil
}

// Skip n bytes
// Peek returns the next n bytes without advancing the reader
func (br *BinaryReader) Peek(n int) ([]byte, error) {
	return br.reader.Peek(n)
}

func (br *BinaryReader) Skip(n int) error {
	buf := make([]byte, n)
	_, err := io.ReadFull(br.reader, buf)
	if err == nil {
		br.pos += n
	}
	return err
}
