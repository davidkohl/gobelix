// cat/cat023/dataitems/v13/service_statistics.go
package v13

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/davidkohl/gobelix/asterix"
)

// ServiceStatistics represents I023/120 - Service Statistics
// Compound data item with optional subfields
type ServiceStatistics struct {
	// Subfield #1: TYPE - Type of report counter
	TYPE *TYPEValue

	// Subfield #2: REF - Reference from which the messages are countered
	REF *REFValue

	// Subfield #3: Counter Value (Counters)
	CV []CounterValue
}

// TYPEValue represents the TYPE subfield (1 byte)
type TYPEValue struct {
	TYPE uint8 // 8-bit type: 1=Number of unknown messages, 2=Number of too old messages, 3=Number of failed messages, etc.
}

// REFValue represents the REF subfield (1 byte)
type REFValue struct {
	REF bool // 0=From midnight, 1=From last report
}

// CounterValue represents a single counter (repetitive, 3 bytes each)
type CounterValue struct {
	Counter uint32 // 24-bit counter value
}

// Decode decodes the Service Statistics from bytes
func (s *ServiceStatistics) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Read primary subfield (1 octet with FX extension bit)
	primaryBytes := make([]byte, 0, 1)
	for {
		b, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("%w: reading service statistics primary subfield", asterix.ErrBufferTooShort)
		}
		bytesRead++
		primaryBytes = append(primaryBytes, b)

		// Check if there's an extension
		hasExtension := (b & 0x01) != 0
		if !hasExtension {
			break
		}
	}

	// Now read subfields based on the bits set in the primary subfield
	subfieldIndex := 0
	for byteIdx := 0; byteIdx < len(primaryBytes); byteIdx++ {
		// Process bits 8-2 (bit 1 is FX)
		for bitPos := 7; bitPos >= 1; bitPos-- {
			if (primaryBytes[byteIdx] & (1 << bitPos)) != 0 {
				// This subfield is present
				n, err := s.decodeSubfield(subfieldIndex, buf)
				bytesRead += n
				if err != nil {
					return bytesRead, fmt.Errorf("decoding subfield #%d: %w", subfieldIndex+1, err)
				}
			}
			subfieldIndex++
		}
	}

	return bytesRead, nil
}

// decodeSubfield decodes a specific subfield based on index
func (s *ServiceStatistics) decodeSubfield(index int, buf *bytes.Buffer) (int, error) {
	switch index {
	case 0: // #1: TYPE
		b, err := buf.ReadByte()
		if err != nil {
			return 0, fmt.Errorf("%w: reading TYPE subfield", asterix.ErrBufferTooShort)
		}
		s.TYPE = &TYPEValue{TYPE: b}
		return 1, nil

	case 1: // #2: REF
		b, err := buf.ReadByte()
		if err != nil {
			return 0, fmt.Errorf("%w: reading REF subfield", asterix.ErrBufferTooShort)
		}
		s.REF = &REFValue{REF: (b & 0x80) != 0}
		return 1, nil

	case 2: // #3: CV - Counter Values (Repetitive)
		// Read repetition factor
		rep, err := buf.ReadByte()
		if err != nil {
			return 0, fmt.Errorf("%w: reading counter value repetition factor", asterix.ErrBufferTooShort)
		}
		bytesRead := 1

		s.CV = make([]CounterValue, rep)
		for i := 0; i < int(rep); i++ {
			var counterData [3]byte
			n, err := buf.Read(counterData[:])
			if err != nil || n != 3 {
				return bytesRead + n, fmt.Errorf("%w: reading counter value %d", asterix.ErrBufferTooShort, i+1)
			}
			bytesRead += n

			// Decode 24-bit counter
			counter := uint32(counterData[0])<<16 | uint32(counterData[1])<<8 | uint32(counterData[2])
			s.CV[i] = CounterValue{Counter: counter}
		}
		return bytesRead, nil

	default:
		return 0, fmt.Errorf("unknown service statistics subfield index: %d", index)
	}
}

// Encode encodes the Service Statistics to bytes
func (s *ServiceStatistics) Encode(buf *bytes.Buffer) (int, error) {
	if err := s.Validate(); err != nil {
		return 0, err
	}

	// Build primary subfield based on which fields are present
	primaryBytes := s.buildPrimarySubfield()
	n, err := buf.Write(primaryBytes)
	if err != nil {
		return n, fmt.Errorf("writing service statistics primary subfield: %w", err)
	}
	bytesWritten := n

	// Encode each present subfield
	encoders := []func(*bytes.Buffer) (int, error){
		s.encodeTYPE,
		s.encodeREF,
		s.encodeCV,
	}

	for _, encoder := range encoders {
		n, err := encoder(buf)
		bytesWritten += n
		if err != nil {
			return bytesWritten, err
		}
	}

	return bytesWritten, nil
}

// buildPrimarySubfield builds the FSPEC for compound data item
func (s *ServiceStatistics) buildPrimarySubfield() []byte {
	// Determine which subfields are present
	presence := []bool{
		s.TYPE != nil,
		s.REF != nil,
		len(s.CV) > 0,
	}

	// Build primary subfield byte
	var b byte
	if presence[0] {
		b |= 0x80 // bit 8
	}
	if presence[1] {
		b |= 0x40 // bit 7
	}
	if presence[2] {
		b |= 0x20 // bit 6
	}
	// FX bit is 0 (no further extensions needed for now)

	return []byte{b}
}

// Individual subfield encoders
func (s *ServiceStatistics) encodeTYPE(buf *bytes.Buffer) (int, error) {
	if s.TYPE == nil {
		return 0, nil
	}
	return buf.Write([]byte{s.TYPE.TYPE})
}

func (s *ServiceStatistics) encodeREF(buf *bytes.Buffer) (int, error) {
	if s.REF == nil {
		return 0, nil
	}
	var b byte
	if s.REF.REF {
		b = 0x80
	}
	return buf.Write([]byte{b})
}

func (s *ServiceStatistics) encodeCV(buf *bytes.Buffer) (int, error) {
	if len(s.CV) == 0 {
		return 0, nil
	}

	// Write repetition factor
	n, err := buf.Write([]byte{byte(len(s.CV))})
	if err != nil {
		return n, err
	}
	bytesWritten := n

	// Write each counter
	for i := range s.CV {
		counter := s.CV[i].Counter
		data := []byte{
			byte(counter >> 16),
			byte(counter >> 8),
			byte(counter),
		}
		n, err := buf.Write(data)
		bytesWritten += n
		if err != nil {
			return bytesWritten, err
		}
	}

	return bytesWritten, nil
}

// Validate validates the Service Statistics
func (s *ServiceStatistics) Validate() error {
	// TYPE values should be reasonable (1-255)
	if s.TYPE != nil && s.TYPE.TYPE == 0 {
		return fmt.Errorf("%w: TYPE value should not be 0", asterix.ErrInvalidMessage)
	}

	// Counter values should fit in 24 bits
	for i, cv := range s.CV {
		if cv.Counter > 0xFFFFFF {
			return fmt.Errorf("%w: counter value %d exceeds 24-bit limit: %d", asterix.ErrInvalidMessage, i, cv.Counter)
		}
	}

	return nil
}

// String returns a string representation
func (s *ServiceStatistics) String() string {
	var parts []string

	if s.TYPE != nil {
		typeDesc := map[uint8]string{
			1: "Unknown Messages",
			2: "Too Old Messages",
			3: "Failed Messages",
			4: "Total Messages",
			5: "Total Corrected Messages",
		}
		if desc, ok := typeDesc[s.TYPE.TYPE]; ok {
			parts = append(parts, fmt.Sprintf("Type:%s", desc))
		} else {
			parts = append(parts, fmt.Sprintf("Type:%d", s.TYPE.TYPE))
		}
	}

	if s.REF != nil {
		if s.REF.REF {
			parts = append(parts, "Ref:Last Report")
		} else {
			parts = append(parts, "Ref:Midnight")
		}
	}

	if len(s.CV) > 0 {
		var counters []string
		for _, cv := range s.CV {
			counters = append(counters, fmt.Sprintf("%d", cv.Counter))
		}
		parts = append(parts, fmt.Sprintf("Counters:[%s]", strings.Join(counters, ", ")))
	}

	if len(parts) == 0 {
		return "ServiceStatistics{}"
	}

	return fmt.Sprintf("ServiceStatistics{%s}", strings.Join(parts, ", "))
}
