// dataitems/cat062/system_track_update_ages.go
package v120

import (
	"bytes"
	"fmt"
)

// SystemTrackUpdateAges implements I062/290
// Compound data item with a primary subfield indicating presence of subfields
type SystemTrackUpdateAges struct {
	Data []byte
}

func (s *SystemTrackUpdateAges) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0
	s.Data = nil

	// Read primary subfield (first byte - may be extended)
	primaryByte := make([]byte, 1)
	n, err := buf.Read(primaryByte)
	if err != nil {
		return n, fmt.Errorf("reading system track update ages primary subfield: %w", err)
	}
	bytesRead += n
	s.Data = append(s.Data, primaryByte[0])

	// Check if extended primary subfield
	hasExtension := (primaryByte[0] & 0x01) != 0

	// Read second byte of primary subfield if extension bit is set
	if hasExtension {
		secondByte := make([]byte, 1)
		n, err := buf.Read(secondByte)
		if err != nil {
			return bytesRead, fmt.Errorf("reading system track update ages extension: %w", err)
		}
		bytesRead += n
		s.Data = append(s.Data, secondByte[0])

		// Note: According to the spec, there's no further extensions of the primary subfield
	}

	// Now read each subfield based on bits set in the primary subfield
	// First byte - check bits 8-2 (bits 16-10 if counting from 1)
	for i := 7; i > 0; i-- {
		if (primaryByte[0] & (1 << i)) != 0 {
			// Each subfield is 1 octet except subfield #5 which is 2 octets
			subfieldSize := 1
			if i == 4 { // Bit 12 = Subfield #5
				subfieldSize = 2
			}

			subfieldData := make([]byte, subfieldSize)
			n, err := buf.Read(subfieldData)
			if err != nil || n != subfieldSize {
				return bytesRead + n, fmt.Errorf("reading system track update ages subfield: %w", err)
			}
			bytesRead += n
			s.Data = append(s.Data, subfieldData...)
		}
	}

	// If we have second primary byte, check bits 8-2 (skip FX bit)
	if hasExtension {
		for i := 7; i > 0; i-- {
			if (s.Data[1] & (1 << i)) != 0 {
				// All remaining subfields are 1 octet
				subfieldData := make([]byte, 1)
				n, err := buf.Read(subfieldData)
				if err != nil || n != 1 {
					return bytesRead + n, fmt.Errorf("reading system track update ages extended subfield: %w", err)
				}
				bytesRead += n
				s.Data = append(s.Data, subfieldData...)
			}
		}
	}

	return bytesRead, nil
}

func (s *SystemTrackUpdateAges) Encode(buf *bytes.Buffer) (int, error) {
	if len(s.Data) == 0 {
		// If no data, encode a minimal valid value (empty FSPEC)
		return buf.Write([]byte{0})
	}
	return buf.Write(s.Data)
}

func (s *SystemTrackUpdateAges) String() string {
	return fmt.Sprintf("SystemTrackUpdateAges[%d bytes]", len(s.Data))
}

// SystemTrackUpdateAges Validate method
func (s *SystemTrackUpdateAges) Validate() error {
	// Stub implementation
	return nil
}
