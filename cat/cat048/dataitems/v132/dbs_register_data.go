// dataitems/cat048/bds_register_data.go
package v132

import (
	"bytes"
	"fmt"
)

// BDSRegister represents a single BDS register with its data and identification
type BDSRegister struct {
	Data []byte // 56-bit message (7 bytes)
	BDS1 uint8  // BDS Register Address 1 (4 bits)
	BDS2 uint8  // BDS Register Address 2 (4 bits)
}

// BDSRegisterData implements I048/250
// BDS Register Data as extracted from the aircraft transponder
type BDSRegisterData struct {
	Registers []BDSRegister
}

// Decode implements the DataItem interface
func (b *BDSRegisterData) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0
	b.Registers = nil

	// Read repetition factor
	repByte := make([]byte, 1)
	n, err := buf.Read(repByte)
	if err != nil {
		return bytesRead, fmt.Errorf("reading BDS register repetition: %w", err)
	}
	bytesRead += n

	repetitions := int(repByte[0])
	if repetitions == 0 {
		return bytesRead, nil // No registers
	}

	// Read each register
	for i := 0; i < repetitions; i++ {
		register := BDSRegister{
			Data: make([]byte, 7),
		}

		// Read 7 bytes of register data
		m, err := buf.Read(register.Data)
		if err != nil {
			return bytesRead + m, fmt.Errorf("reading BDS register data: %w", err)
		}
		bytesRead += m

		if m != 7 {
			return bytesRead, fmt.Errorf("insufficient data for BDS register: got %d bytes, want 7", m)
		}

		// Read BDS1/BDS2
		bdsBytes := make([]byte, 1)
		p, err := buf.Read(bdsBytes)
		if err != nil {
			return bytesRead + p, fmt.Errorf("reading BDS register code: %w", err)
		}
		bytesRead += p

		register.BDS1 = (bdsBytes[0] >> 4) & 0x0F // bits 8-5
		register.BDS2 = bdsBytes[0] & 0x0F        // bits 4-1

		b.Registers = append(b.Registers, register)
	}

	return bytesRead, nil
}

// Encode implements the DataItem interface
func (b *BDSRegisterData) Encode(buf *bytes.Buffer) (int, error) {
	if err := b.Validate(); err != nil {
		return 0, err
	}

	bytesWritten := 0

	// Write repetition factor
	rep := byte(len(b.Registers))
	err := buf.WriteByte(rep)
	if err != nil {
		return bytesWritten, fmt.Errorf("writing BDS register repetition: %w", err)
	}
	bytesWritten++

	// Write each register
	for i, register := range b.Registers {
		// Write 7 bytes of register data
		n, err := buf.Write(register.Data)
		if err != nil {
			return bytesWritten + n, fmt.Errorf("writing BDS register data: %w", err)
		}
		bytesWritten += n

		// Write BDS1/BDS2
		bdsCode := byte(register.BDS1<<4 | register.BDS2)
		err = buf.WriteByte(bdsCode)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing BDS register code for register %d: %w", i, err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

// Validate implements the DataItem interface
func (b *BDSRegisterData) Validate() error {
	if len(b.Registers) > 255 {
		return fmt.Errorf("too many BDS registers: %d (max 255)", len(b.Registers))
	}

	for i, register := range b.Registers {
		if len(register.Data) != 7 {
			return fmt.Errorf("invalid data length for BDS register %d: %d bytes (should be 7)", i, len(register.Data))
		}
		if register.BDS1 > 15 || register.BDS2 > 15 {
			return fmt.Errorf("invalid BDS code for register %d: %X,%X (both should be 0-15)", i, register.BDS1, register.BDS2)
		}
	}

	return nil
}

// String returns a human-readable representation
func (b *BDSRegisterData) String() string {
	if len(b.Registers) == 0 {
		return "No BDS Registers"
	}

	result := fmt.Sprintf("%d BDS Register(s)", len(b.Registers))
	for i, register := range b.Registers {
		if register.BDS1 == 0 && register.BDS2 == 0 {
			// Comm-B broadcast
			result += fmt.Sprintf("\n  #%d: Comm-B Broadcast", i+1)
		} else {
			// Try to decode known BDS registers
			decoded := ""
			switch {
			case register.BDS1 == 4 && register.BDS2 == 0:
				if data, err := DecodeBDS40(register.Data); err == nil {
					decoded = "\n  " + indentString(data.String(), "  ")
				}
			case register.BDS1 == 5 && register.BDS2 == 0:
				if data, err := DecodeBDS50(register.Data); err == nil {
					decoded = "\n  " + indentString(data.String(), "  ")
				}
			case register.BDS1 == 6 && register.BDS2 == 0:
				if data, err := DecodeBDS60(register.Data); err == nil {
					decoded = "\n  " + indentString(data.String(), "  ")
				}
			}

			if decoded != "" {
				result += decoded
			} else {
				result += fmt.Sprintf("\n  #%d: BDS %X,%X (raw data)", i+1, register.BDS1, register.BDS2)
			}
		}
	}

	return result
}

// indentString adds indentation to all lines in a string
func indentString(s string, indent string) string {
	lines := []string{}
	current := ""
	for i, c := range s {
		if c == '\n' {
			lines = append(lines, indent+current)
			current = ""
		} else {
			current += string(c)
		}
		if i == len(s)-1 && current != "" {
			lines = append(lines, indent+current)
		}
	}
	result := ""
	for i, line := range lines {
		result += line
		if i < len(lines)-1 {
			result += "\n"
		}
	}
	return result
}

// AddRegister adds a new BDS register
func (b *BDSRegisterData) AddRegister(data []byte, bds1, bds2 uint8) error {
	// Validate inputs
	if len(data) != 7 {
		return fmt.Errorf("BDS register data must be 7 bytes (56 bits)")
	}
	if bds1 > 15 || bds2 > 15 {
		return fmt.Errorf("BDS codes must be 0-15, got %X,%X", bds1, bds2)
	}

	// Create and add register
	register := BDSRegister{
		Data: make([]byte, 7),
		BDS1: bds1,
		BDS2: bds2,
	}
	copy(register.Data, data)
	b.Registers = append(b.Registers, register)

	return nil
}

// GetBDSRegister returns the BDS register with the specified codes, if available
func (b *BDSRegisterData) GetBDSRegister(bds1, bds2 uint8) ([]byte, bool) {
	for _, register := range b.Registers {
		if register.BDS1 == bds1 && register.BDS2 == bds2 {
			return register.Data, true
		}
	}
	return nil, false
}
