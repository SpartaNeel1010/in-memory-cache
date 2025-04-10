package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strconv"
)

func getAllKeys()([]string, error) {
	
	file, err := os.Open(fullPath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		var ret []string 
		return ret,err
	}
	defer file.Close()

	r := bufio.NewReader(file)
	keys, err := parseRDB(r)
	return keys,err
}

func parseRDB(r *bufio.Reader) ([]string, error) {
	// Header: expect "REDIS0011" (9 bytes)
	header := make([]byte, 9)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}
	if string(header) != "REDIS0011" {
		return nil, fmt.Errorf("invalid header: %s", string(header))
	}

	var keys []string
	for {
		// Peek the next marker byte.
		markerByte, err := r.Peek(1)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		switch markerByte[0] {
		case 0xFA:
			// Metadata subsection. Consume marker and skip name/value.
			_, _ = r.ReadByte() // consume 0xFA
			_, err = readString(r) // metadata name
			if err != nil {
				return nil, err
			}
			_, err = readString(r) // metadata value
			if err != nil {
				return nil, err
			}
		case 0xFE:
			// Database subsection.
			dbKeys, err := parseDatabaseSection(r)
			if err != nil {
				return nil, err
			}
			keys = append(keys, dbKeys...)
		case 0xFF:
			// End-of-file section. Consume marker and checksum.
			_, _ = r.ReadByte() // consume 0xFF
			// Read the following 8-byte checksum.
			checksum := make([]byte, 8)
			if _, err := io.ReadFull(r, checksum); err != nil {
				return nil, err
			}
			return keys, nil
		default:
			return nil, fmt.Errorf("unexpected marker: 0x%x", markerByte[0])
		}
	}
	return keys, nil
}

// parseDatabaseSection parses one database subsection and returns all keys found.
func parseDatabaseSection(r *bufio.Reader) ([]string, error) {
	marker, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	if marker != 0xFE {
		return nil, fmt.Errorf("expected database section marker 0xFE, got 0x%x", marker)
	}

	// Read the database index (size encoded, we ignore its value here)
	_, err = readLength(r)
	if err != nil {
		return nil, err
	}

	// Next, we expect the hash table size marker (0xFB)
	marker, err = r.ReadByte()
	if err != nil {
		return nil, err
	}
	if marker != 0xFB {
		return nil, fmt.Errorf("expected hash table size marker 0xFB, got 0x%x", marker)
	}

	// Read the total key-value pair count and the expire count (both size encoded).
	totalKeys, err := readLength(r)
	if err != nil {
		return nil, err
	}
	_, err = readLength(r) // expire count; not used further here
	if err != nil {
		return nil, err
	}

	var keys []string
	// Process each key-value pair.
	for i := uint64(0); i < totalKeys; i++ {
		// Check for optional expire marker.
		peek, err := r.Peek(1)
		if err != nil {
			return nil, err
		}
		if peek[0] == 0xFC {
			// Expire in milliseconds: marker + 8-byte timestamp.
			_, _ = r.ReadByte() // consume marker
			if _, err := io.CopyN(io.Discard, r, 8); err != nil {
				return nil, err
			}
		} else if peek[0] == 0xFD {
			// Expire in seconds: marker + 4-byte timestamp.
			_, _ = r.ReadByte() // consume marker
			if _, err := io.CopyN(io.Discard, r, 4); err != nil {
				return nil, err
			}
		}

		// Read the value type (1 byte); for our purposes we ignore its value.
		_, err = r.ReadByte()
		if err != nil {
			return nil, err
		}

		// Read the key (string encoded).
		key, err := readString(r)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)

		// Skip the value (assumed to be string encoded for this challenge).
		_, err = readString(r)
		if err != nil {
			return nil, err
		}
	}
	return keys, nil
}

// readLength reads a size-encoded integer from r.
// It handles the 3 cases where the first two bits indicate how many bytes the length uses.
func readLength(r *bufio.Reader) (uint64, error) {
	b, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	typ := b >> 6
	switch typ {
	case 0: // 0b00: length is in the lower 6 bits.
		return uint64(b & 0x3F), nil
	case 1: // 0b01: length is in the next 14 bits.
		b2, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		return uint64((b & 0x3F)<<8 | b2), nil
	case 2: // 0b10: length is encoded in the next 4 bytes (big-endian).
		buf := make([]byte, 4)
		if _, err := io.ReadFull(r, buf); err != nil {
			return 0, err
		}
		return uint64(binary.BigEndian.Uint32(buf)), nil
	case 3:
		// 0b11 indicates a special string encoding; this function expects a plain length.
		return 0, fmt.Errorf("special encoding encountered in length")
	default:
		return 0, fmt.Errorf("invalid length encoding")
	}
}

// readString reads a string-encoded value from r.
// It first reads a size (or special encoding indicator) then reads the corresponding bytes.
func readString(r *bufio.Reader) (string, error) {
	b, err := r.ReadByte()
	if err != nil {
		return "", err
	}
	typ := b >> 6
	if typ != 3 {
		// Normal string: unread the byte and use readLength.
		if err := r.UnreadByte(); err != nil {
			return "", err
		}
		length, err := readLength(r)
		if err != nil {
			return "", err
		}
		buf := make([]byte, length)
		if _, err := io.ReadFull(r, buf); err != nil {
			return "", err
		}
		return string(buf), nil
	}

	// Special encoded string.
	specialType := b & 0x3F
	switch specialType {
	case 0: // 8-bit integer encoded.
		i, err := r.ReadByte()
		if err != nil {
			return "", err
		}
		return strconv.Itoa(int(i)), nil
	case 1: // 16-bit integer encoded, little-endian.
		buf := make([]byte, 2)
		if _, err := io.ReadFull(r, buf); err != nil {
			return "", err
		}
		val := int(binary.LittleEndian.Uint16(buf))
		return strconv.Itoa(val), nil
	case 2: // 32-bit integer encoded, little-endian.
		buf := make([]byte, 4)
		if _, err := io.ReadFull(r, buf); err != nil {
			return "", err
		}
		val := int(binary.LittleEndian.Uint32(buf))
		return strconv.Itoa(val), nil
	default:
		return "", fmt.Errorf("unsupported special string encoding type: %d", specialType)
	}
}



func loadDB() {
	// Read header: expect "REDIS0011" (9 bytes)
	file, err := os.Open(fullPath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return 
	}
	defer file.Close()

	r := bufio.NewReader(file)
	header := make([]byte, 9)
	if _, err := io.ReadFull(r, header); err != nil {
		return 
	}
	if string(header) != "REDIS0011" {
		fmt.Printf("invalid header: %s", string(header))
		return 
	}

	

	// Process the remainder of the file.
	for {
		marker, err := r.Peek(1)
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Printf("Error: %s", err)
			return
		}

		switch marker[0] {
		case 0xFA:
			// Metadata subsection: skip metadata name/value.
			_, _ = r.ReadByte() // consume marker 0xFA
			_, err := readString(r) // metadata name
			if err != nil {
				fmt.Printf("Error: %s", err)
				return
			}
			_, err = readString(r) // metadata value
			if err != nil {
				fmt.Printf("Error: %s", err)
				return
			}
		case 0xFE:
			// Database subsection.
			if err := loadDatabaseSection(r); err != nil {
				fmt.Printf("Error: %s", err)
				return
			}
		case 0xFF:
			// End-of-file section: consume marker and checksum.
			_, _ = r.ReadByte() // consume 0xFF
			checksum := make([]byte, 8)
			if _, err := io.ReadFull(r, checksum); err != nil {
				fmt.Printf("Error: %s", err)
				return
			}
			return
		default:
			fmt.Printf("Error: %s", err)
			return
		}
	}

}

// loadDatabaseSection processes one database subsection and fills DB and expTime.
func loadDatabaseSection(r *bufio.Reader) error {
	// Consume the database section marker (0xFE).
	marker, err := r.ReadByte()
	if err != nil {
		return err
	}
	if marker != 0xFE {
		return fmt.Errorf("expected database marker 0xFE, got 0x%x", marker)
	}

	// Read the database index (size encoded; we ignore its value).
	_, err = readLength(r)
	if err != nil {
		return err
	}

	// Expect the hash table size marker (0xFB).
	marker, err = r.ReadByte()
	if err != nil {
		return err
	}
	if marker != 0xFB {
		return fmt.Errorf("expected hash table size marker 0xFB, got 0x%x", marker)
	}

	// Read the total key-value pair count and the expire count.
	totalKeys, err := readLength(r)
	if err != nil {
		return err
	}
	// The expire count (number of keys with expiry) is not used further.
	_, err = readLength(r)
	if err != nil {
		return err
	}

	for i := uint64(0); i < totalKeys; i++ {
		var hasExpire bool
		var expireTs uint64

		// Check for optional expire marker.
		peek, err := r.Peek(1)
		if err != nil {
			return err
		}
		if peek[0] == 0xFC {
			// Marker 0xFC: expire timestamp in milliseconds.
			_, _ = r.ReadByte() // consume marker
			buf := make([]byte, 8)
			if _, err := io.ReadFull(r, buf); err != nil {
				return err
			}
			expireTs = binary.LittleEndian.Uint64(buf)
			hasExpire = true
		} else if peek[0] == 0xFD {
			// Marker 0xFD: expire timestamp in seconds.
			_, _ = r.ReadByte() // consume marker
			buf := make([]byte, 4)
			if _, err := io.ReadFull(r, buf); err != nil {
				return err
			}
			expireTs = uint64(binary.LittleEndian.Uint32(buf))
			hasExpire = true
		}

		// Read the value type (1 byte); its value is ignored.
		_, err = r.ReadByte()
		if err != nil {
			return err
		}

		// Read the key (string encoded).
		key, err := readString(r)
		if err != nil {
			return err
		}

		// Read the value (assumed to be string encoded).
		value, err := readString(r)
		if err != nil {
			return err
		}

		DB[key] = value
		if hasExpire {
			expTime[key] = int(expireTs)
		}
	}

	return nil
}