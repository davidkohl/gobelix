// dataitems/cat062/system_track_update_ages.go
package v117

import (
	"bytes"
	"fmt"
)

// SystemTrackUpdateAges implements I062/290
// Ages of the last plot/local track/target report update for each sensor type
type SystemTrackUpdateAges struct {
	Data []byte
}

func (s *SystemTrackUpdateAges) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0
	s.Data = nil

	// Read primary subfield (first octet)
	if buf.Len() < 1 {
		return 0, fmt.Errorf("buffer too short for FSPEC")
	}

	primaryByte, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading primary subfield: %w", err)
	}
	bytesRead++
	s.Data = append(s.Data, primaryByte)

	// Check if extension exists (FX bit)
	hasExtension := (primaryByte & 0x01) != 0

	// Read second octet if extension exists
	if hasExtension {
		if buf.Len() < 1 {
			return bytesRead, fmt.Errorf("buffer too short for extended primary subfield")
		}

		secondaryByte, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading extended primary subfield: %w", err)
		}
		bytesRead++
		s.Data = append(s.Data, secondaryByte)
	}

	// Define subfield sizes
	subfieldSizes := map[int]int{
		1:  1, // Track age
		2:  1, // PSR age
		3:  1, // SSR age
		4:  1, // Mode S age
		5:  2, // ADS-C age (2 octets)
		6:  1, // ADS-B Extended Squitter age
		7:  1, // ADS-B VDL Mode 4 age
		8:  1, // ADS-B UAT age
		9:  1, // Loop age
		10: 1, // Multilateration age
	}

	// Check which subfields are present based on FSPEC bits
	// First octet (bits 16-10)
	for i := 0; i < 7; i++ {
		bitPos := 7 - i
		if primaryByte&(1<<bitPos) != 0 {
			frn := i + 1
			size := subfieldSizes[frn]

			if buf.Len() < size {
				return bytesRead, fmt.Errorf("buffer too short for subfield %d: need %d bytes, have %d",
					frn, size, buf.Len())
			}

			data := make([]byte, size)
			n, err := buf.Read(data)
			if err != nil || n != size {
				return bytesRead + n, fmt.Errorf("reading subfield %d: %w", frn, err)
			}
			bytesRead += n
			s.Data = append(s.Data, data...)
		}
	}

	// Second octet if present (bits 8-2)
	if hasExtension {
		secondaryByte := s.Data[1]
		for i := 0; i < 7; i++ {
			bitPos := 7 - i
			if secondaryByte&(1<<bitPos) != 0 {
				frn := i + 8 // FRNs 8-14 (ignoring the FX bit)
				size := subfieldSizes[frn]

				if buf.Len() < size {
					return bytesRead, fmt.Errorf("buffer too short for subfield %d: need %d bytes, have %d",
						frn, size, buf.Len())
				}

				data := make([]byte, size)
				n, err := buf.Read(data)
				if err != nil || n != size {
					return bytesRead + n, fmt.Errorf("reading subfield %d: %w", frn, err)
				}
				bytesRead += n
				s.Data = append(s.Data, data...)
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

func (s *SystemTrackUpdateAges) Validate() error {
	// Basic validation - at least one FSPEC byte should be present
	if len(s.Data) == 0 {
		return fmt.Errorf("empty data")
	}
	return nil
}