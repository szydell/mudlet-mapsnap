package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// ExamineFile examines a binary file and prints its structure
func ExamineFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	// Read first 1024 bytes
	data := make([]byte, 1024)
	n, err := file.Read(data)
	if err != nil && err != io.EOF {
		return fmt.Errorf("reading file: %w", err)
	}
	data = data[:n]

	fmt.Printf("File size: %d bytes\n", n)
	fmt.Println("First 32 bytes as hex:")
	for i := 0; i < 32 && i < n; i++ {
		fmt.Printf("%02x ", data[i])
		if (i+1)%8 == 0 {
			fmt.Println()
		}
	}
	fmt.Println()

	// Try to interpret the first few values as different types
	if n >= 4 {
		fmt.Printf("First 4 bytes as int32 (big endian): %d\n", binary.BigEndian.Uint32(data[:4]))
		fmt.Printf("First 4 bytes as int32 (little endian): %d\n", binary.LittleEndian.Uint32(data[:4]))
	}

	// Look for UTF-16 strings
	fmt.Println("\nPossible UTF-16 strings:")

	// Look for UTF-16BE strings (first byte is 0, second byte is ASCII)
	fmt.Println("UTF-16BE strings (common in Java and network protocols):")
	for i := 0; i < n-20; i += 2 {
		if data[i] == 0 && data[i+1] >= 32 && data[i+1] <= 126 {
			// Found potential start of UTF-16BE string
			start := i
			str := ""

			for j := i; j < n-1; j += 2 {
				if data[j] == 0 && data[j+1] >= 32 && data[j+1] <= 126 {
					str += string(data[j+1])
				} else {
					break
				}
			}

			// Only report strings of reasonable length
			if len(str) >= 4 {
				fmt.Printf("Offset %d: UTF-16BE String: %s\n", start, str)
				// Skip ahead to avoid duplicate detections
				i = start + len(str)*2 - 2
			}
		}
	}

	// Look for UTF-16LE strings (first byte is ASCII, second byte is 0)
	fmt.Println("\nUTF-16LE strings (common in Windows):")
	for i := 0; i < n-20; i += 2 {
		if data[i] >= 32 && data[i] <= 126 && data[i+1] == 0 {
			// Found potential start of UTF-16LE string
			start := i
			str := ""

			for j := i; j < n-1; j += 2 {
				if data[j] >= 32 && data[j] <= 126 && data[j+1] == 0 {
					str += string(data[j])
				} else {
					break
				}
			}

			// Only report strings of reasonable length
			if len(str) >= 4 {
				fmt.Printf("Offset %d: UTF-16LE String: %s\n", start, str)
				// Skip ahead to avoid duplicate detections
				i = start + len(str)*2 - 2
			}
		}
	}

	// Look for length-prefixed strings (4-byte length followed by UTF-16BE)
	fmt.Println("\nLength-prefixed UTF-16BE strings:")
	for i := 0; i < n-8; i += 4 {
		// Check for a potential length prefix (4 bytes)
		if i+4 < n {
			length := int(binary.BigEndian.Uint32(data[i:i+4]))
			// Only consider reasonable string lengths
			if length > 0 && length < 100 && i+4+length*2 <= n {
				str := ""
				valid := true

				for j := 0; j < length; j++ {
					if i+4+j*2+1 < n {
						// Check for UTF-16BE pattern (first byte is 0 for ASCII range)
						if data[i+4+j*2] == 0 && data[i+4+j*2+1] >= 32 && data[i+4+j*2+1] <= 126 {
							str += string(data[i+4+j*2+1])
						} else {
							valid = false
							break
						}
					}
				}

				if valid && len(str) == length {
					fmt.Printf("Offset %d: Length prefix %d, String: %s\n", i, length, str)
				}
			}
		}
	}

	// Try to find ASCII strings
	fmt.Println("\nPossible ASCII strings:")
	start := -1
	for i := 0; i < n; i++ {
		if data[i] >= 32 && data[i] <= 126 {
			if start == -1 {
				start = i
			}
		} else {
			if start != -1 && i-start >= 4 {
				fmt.Printf("Offset %d: %s\n", start, string(data[start:i]))
			}
			start = -1
		}
	}

	return nil
}
