// dataitems/cat062/aircraft_derived_data.go
package v117

import (
	"bytes"
	"fmt"
)

// AircraftDerivedData implements I062/380
// Data derived directly by the aircraft
type AircraftDerivedData struct {
	Data []byte
}

func (a *AircraftDerivedData) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0
	a.Data = nil

	// Read FSPEC bytes (variable number based on extension bits)
	var fspecBytes []byte
	var fspecByteCount int

	for {
		if buf.Len() < 1 {
			return bytesRead, fmt.Errorf("buffer too short for FSPEC byte")
		}

		fspecByte, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading FSPEC byte: %w", err)
		}
		bytesRead++
		fspecBytes = append(fspecBytes, fspecByte)
		fspecByteCount++

		// Check if we need to continue reading FSPEC (FX bit)
		if fspecByte&0x01 == 0 {
			break
		}

		// Safety check - no valid ASTERIX message needs more than 4 FSPEC bytes
		if fspecByteCount >= 4 {
			return bytesRead, fmt.Errorf("too many FSPEC bytes")
		}
	}

	// Store FSPEC bytes
	a.Data = fspecBytes

	// Now we need to determine which subfields are present and read them
	// This is a mapping of FRNs to data sizes in bytes
	subfieldSizes := map[int]int{
		1:  3, // Target Address
		2:  6, // Target Identification
		3:  2, // Magnetic Heading
		4:  2, // Indicated Airspeed/Mach
		5:  2, // True Airspeed
		6:  2, // Selected Altitude
		7:  2, // Final State Selected Altitude
		10: 2, // Communications/ACAS Capability
		11: 2, // Status reported by ADS-B
		12: 7, // ACAS Resolution Advisory Report
		13: 2, // Barometric Vertical Rate
		14: 2, // Geometric Vertical Rate
		15: 2, // Roll Angle
		16: 2, // Track Angle Rate
		17: 2, // Track Angle
		18: 2, // Ground Speed
		19: 1, // Velocity Uncertainty
		20: 8, // Meteorological Data
		21: 1, // Emitter Category
		22: 6, // Position
		23: 2, // Geometric Altitude
		24: 1, // Position Uncertainty
		26: 2, // Indicated Airspeed
		27: 2, // Mach Number
		28: 2, // Barometric Pressure Setting
	}

	// Special variable-length subfields
	variableSizeSubfields := map[int]bool{
		8:  true, // Trajectory Intent Status
		9:  true, // Trajectory Intent Data
		25: true, // Mode S MB Data
	}

	// Process each bit in the FSPEC
	for byteIdx, fspecByte := range fspecBytes {
		// Process 7 bits in each FSPEC byte (bit 0 is FX bit)
		for bitPos := 7; bitPos > 0; bitPos-- {
			if fspecByte&(1<<bitPos) == 0 {
				continue // Bit not set, no subfield
			}

			// Calculate FRN from byte and bit position
			frn := 7*byteIdx + (8 - bitPos)

			// Special handling for variable-length subfields
			if variableSizeSubfields[frn] {
				switch frn {
				case 8: // Trajectory Intent Status
					if buf.Len() < 1 {
						return bytesRead, fmt.Errorf("buffer too short for subfield %d", frn)
					}

					b, err := buf.ReadByte()
					if err != nil {
						return bytesRead, fmt.Errorf("reading subfield %d: %w", frn, err)
					}
					bytesRead++
					a.Data = append(a.Data, b)

					// Handle extension if present
					if b&0x01 != 0 {
						return bytesRead, fmt.Errorf("subfield %d extension not supported", frn)
					}

				case 9: // Trajectory Intent Data
					if buf.Len() < 1 {
						return bytesRead, fmt.Errorf("buffer too short for subfield %d repetition", frn)
					}

					// Read repetition factor
					repByte, err := buf.ReadByte()
					if err != nil {
						return bytesRead, fmt.Errorf("reading subfield %d repetition: %w", frn, err)
					}
					bytesRead++
					a.Data = append(a.Data, repByte)

					// Each point is 15 octets
					requiredBytes := int(repByte) * 15
					if buf.Len() < requiredBytes {
						return bytesRead, fmt.Errorf("buffer too short for subfield %d: need %d bytes, have %d",
							frn, requiredBytes, buf.Len())
					}

					// Read the data
					data := make([]byte, requiredBytes)
					n, err := buf.Read(data)
					if err != nil || n != requiredBytes {
						return bytesRead + n, fmt.Errorf("reading subfield %d data: %w", frn, err)
					}
					bytesRead += n
					a.Data = append(a.Data, data...)

				case 25: // Mode S MB Data
					if buf.Len() < 1 {
						return bytesRead, fmt.Errorf("buffer too short for subfield %d repetition", frn)
					}

					// Read repetition factor
					repByte, err := buf.ReadByte()
					if err != nil {
						return bytesRead, fmt.Errorf("reading subfield %d repetition: %w", frn, err)
					}
					bytesRead++
					a.Data = append(a.Data, repByte)

					// Each entry is 8 octets
					requiredBytes := int(repByte) * 8
					if buf.Len() < requiredBytes {
						return bytesRead, fmt.Errorf("buffer too short for subfield %d: need %d bytes, have %d",
							frn, requiredBytes, buf.Len())
					}

					// Read the data
					data := make([]byte, requiredBytes)
					n, err := buf.Read(data)
					if err != nil || n != requiredBytes {
						return bytesRead + n, fmt.Errorf("reading subfield %d data: %w", frn, err)
					}
					bytesRead += n
					a.Data = append(a.Data, data...)
				}
			} else if size, ok := subfieldSizes[frn]; ok {
				// Fixed-size subfield
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
				a.Data = append(a.Data, data...)
			} else {
				// Unknown subfield - this shouldn't happen with proper FSPEC processing
				return bytesRead, fmt.Errorf("unknown subfield FRN %d", frn)
			}
		}
	}

	return bytesRead, nil
}

func (a *AircraftDerivedData) Encode(buf *bytes.Buffer) (int, error) {
	if len(a.Data) == 0 {
		// If no data, encode a minimal valid structure (empty FSPEC)
		return buf.Write([]byte{0})
	}
	return buf.Write(a.Data)
}

func (a *AircraftDerivedData) String() string {
	return fmt.Sprintf("AircraftDerivedData[%d bytes]", len(a.Data))
}

func (a *AircraftDerivedData) Validate() error {
	// Basic validation - at least one FSPEC byte should be present
	if len(a.Data) == 0 {
		return fmt.Errorf("empty data")
	}
	return nil
}
