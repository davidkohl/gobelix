// dataitems/cat021/mode_3a_code.go
package v26

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// Mode3ACode implements I021/070
// Mode 3/A code in octal representation
type Mode3ACode struct {
	Code  string // Transponder code as a string in octal (e.g., "7777")
	Valid bool   // Code validation
}

func (m *Mode3ACode) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading Mode 3/A code: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("insufficient data for Mode 3/A code: got %d bytes, want 2", n)
	}

	// First 4 bits are reserved and should be 0
	if (data[0] & 0xF0) != 0 {
		return n, fmt.Errorf("invalid Mode 3/A code format: reserved bits not 0")
	}

	// Extract the 12-bit code
	codeValue := uint16(data[0])<<8 | uint16(data[1])

	// Convert to octal and format as a 4-digit string
	m.Code = fmt.Sprintf("%04o", codeValue)
	m.Valid = true

	return n, m.Validate()
}

func (m *Mode3ACode) Encode(buf *bytes.Buffer) (int, error) {
	if err := m.Validate(); err != nil {
		return 0, err
	}

	// Clean the code and convert to uint16
	codeValue, err := m.getCodeValue()
	if err != nil {
		return 0, err
	}

	data := make([]byte, 2)

	// First byte: First 4 bits are 0, next 4 bits are high bits of code
	data[0] = byte(0) // First 4 bits reserved as 0
	data[0] |= byte((codeValue >> 8) & 0x0F)

	// Second byte: Low 8 bits of code
	data[1] = byte(codeValue & 0xFF)

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing Mode 3/A code: %w", err)
	}

	return n, nil
}

// getCodeValue converts the string code to a uint16 value
func (m *Mode3ACode) getCodeValue() (uint16, error) {
	// Clean the code first
	cleanCode := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '7' {
			return r
		}
		return -1
	}, m.Code)

	if len(cleanCode) == 0 {
		return 0, fmt.Errorf("empty Mode 3/A code string")
	}

	// Parse as octal number
	code, err := strconv.ParseUint(cleanCode, 8, 16)
	if err != nil {
		return 0, fmt.Errorf("invalid octal Mode 3/A code: %w", err)
	}

	return uint16(code), nil
}

func (m *Mode3ACode) Validate() error {
	if m.Code == "" {
		return fmt.Errorf("Mode 3/A code is empty")
	}

	// Clean and check the code
	cleanCode := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '7' {
			return r
		}
		return -1
	}, m.Code)

	if len(cleanCode) == 0 {
		return fmt.Errorf("Mode 3/A code contains no valid octal digits")
	}

	// Parse as octal number to check validity
	code, err := strconv.ParseUint(cleanCode, 8, 16)
	if err != nil {
		return fmt.Errorf("invalid octal Mode 3/A code: %w", err)
	}

	// The code should fit in 12 bits
	if code > 0x0FFF {
		return fmt.Errorf("Mode 3/A code value out of valid range: %d", code)
	}

	return nil
}

// String returns a human-readable representation of the Mode 3/A code
func (m *Mode3ACode) String() string {
	if m.Code == "" {
		return "(empty code)"
	}

	// Clean the code
	cleanCode := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '7' {
			return r
		}
		return -1
	}, m.Code)

	// Format with 4 digits (padding with leading zeros if needed)
	paddedCode := fmt.Sprintf("%04s", cleanCode)
	if len(paddedCode) > 4 {
		paddedCode = paddedCode[len(paddedCode)-4:] // Take last 4 digits if too long
	}

	// Insert spaces between digits for better readability
	formatted := ""
	for i, ch := range paddedCode {
		if i > 0 {
			formatted += " "
		}
		formatted += string(ch)
	}

	if !m.Valid {
		return formatted + " (invalid)"
	}
	return formatted
}
