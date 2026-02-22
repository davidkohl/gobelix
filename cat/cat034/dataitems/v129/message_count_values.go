package v129

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// MessageCountValue represents a single message count entry
type MessageCountValue struct {
	TYP     uint8  // Type of message counter (5 bits, values 0-20)
	Counter uint16 // 11-bit counter value
}

// MessageCountValues represents I034/070 - Message Count Values
// Repetitive data item with REP indicator
type MessageCountValues struct {
	Values []MessageCountValue
}

func (m *MessageCountValues) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Read REP (repetition factor)
	rep, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("%w: reading message count REP", asterix.ErrBufferTooShort)
	}
	bytesRead++

	// If REP is 0, no values to read
	if rep == 0 {
		m.Values = []MessageCountValue{}
		return bytesRead, nil
	}

	// Initialize values slice
	m.Values = make([]MessageCountValue, rep)

	// Read each message count value (2 bytes each)
	for i := 0; i < int(rep); i++ {
		var data [2]byte
		n, err := buf.Read(data[:])
		if err != nil || n != 2 {
			return bytesRead + n, fmt.Errorf("%w: need 2 bytes for message count value, have %d", asterix.ErrBufferTooShort, n)
		}
		bytesRead += n

		// Extract fields
		// TYP: bits 16-12 (5 bits)
		m.Values[i].TYP = (data[0] >> 3) & 0x1F
		// COUNTER: bits 11-1 (11 bits)
		m.Values[i].Counter = uint16(data[0]&0x07)<<8 | uint16(data[1])
	}

	return bytesRead, nil
}

func (m *MessageCountValues) Encode(buf *bytes.Buffer) (int, error) {
	if err := m.Validate(); err != nil {
		return 0, err
	}

	bytesWritten := 0

	// Write REP
	rep := byte(len(m.Values))
	if err := buf.WriteByte(rep); err != nil {
		return bytesWritten, fmt.Errorf("writing message count REP: %w", err)
	}
	bytesWritten++

	// Write each message count value
	for _, val := range m.Values {
		var data [2]byte

		// Build first byte: TYP (bits 16-12), Counter high 3 bits (bits 11-9)
		data[0] = (val.TYP & 0x1F) << 3
		data[0] |= byte((val.Counter >> 8) & 0x07)

		// Second byte: Counter low 8 bits
		data[1] = byte(val.Counter & 0xFF)

		n, err := buf.Write(data[:])
		if err != nil {
			return bytesWritten + n, fmt.Errorf("writing message count value: %w", err)
		}
		bytesWritten += n
	}

	return bytesWritten, nil
}

func (m *MessageCountValues) Validate() error {
	if len(m.Values) > 255 {
		return fmt.Errorf("%w: too many message count values: %d", asterix.ErrInvalidMessage, len(m.Values))
	}
	for i, val := range m.Values {
		if val.TYP > 31 {
			return fmt.Errorf("%w: message type at index %d out of range: %d", asterix.ErrInvalidMessage, i, val.TYP)
		}
		if val.Counter > 2047 {
			return fmt.Errorf("%w: counter at index %d out of range: %d", asterix.ErrInvalidMessage, i, val.Counter)
		}
	}
	return nil
}

func (m *MessageCountValues) String() string {
	if len(m.Values) == 0 {
		return "Message Counts: None"
	}

	result := fmt.Sprintf("Message Counts (%d):", len(m.Values))
	for i, val := range m.Values {
		typeName := getMessageTypeName(val.TYP)
		result += fmt.Sprintf("\n  [%d] %s: %d", i, typeName, val.Counter)
	}
	return result
}

func getMessageTypeName(typ uint8) string {
	switch typ {
	case 0:
		return "No detection (misses)"
	case 1:
		return "Single PSR"
	case 2:
		return "Single SSR (Non-Mode S)"
	case 3:
		return "SSR+PSR (Non-Mode S)"
	case 4:
		return "Single All-Call (Mode S)"
	case 5:
		return "Single Roll-Call (Mode S)"
	case 6:
		return "All-Call+PSR (Mode S)"
	case 7:
		return "Roll-Call+PSR (Mode S)"
	case 8:
		return "Filter: Weather"
	case 9:
		return "Filter: Jamming Strobe"
	case 10:
		return "Filter: PSR"
	case 11:
		return "Filter: SSR/Mode S"
	case 12:
		return "Filter: SSR/Mode S+PSR"
	case 13:
		return "Filter: Enhanced Surveillance"
	case 14:
		return "Filter: PSR+Enhanced Surveillance"
	case 15:
		return "Filter: PSR+Enhanced Surveillance+SSR/Mode S (not in API)"
	case 16:
		return "Filter: PSR+Enhanced Surveillance+all SSR/Mode S"
	case 17:
		return "Re-Interrogations (per sector)"
	case 18:
		return "BDS Swap and wrong DF replies (per sector)"
	case 19:
		return "Mode A/C FRUIT (per sector)"
	case 20:
		return "Mode S FRUIT (per sector)"
	default:
		return fmt.Sprintf("Unknown Type %d", typ)
	}
}
