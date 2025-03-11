// dataitems/cat062/flight_plan_related_data.go
package v117

import (
	"bytes"
	"fmt"
)

// FlightPlanRelatedData implements I062/390
// All flight plan related information, provided by ground-based systems
type FlightPlanRelatedData struct {
	Data []byte
}

func (f *FlightPlanRelatedData) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0
	f.Data = nil

	// Primary subfield can be up to 3 octets, each with FX bit
	for i := 0; i < 3; i++ {
		octet := make([]byte, 1)
		n, err := buf.Read(octet)
		if err != nil {
			return bytesRead, fmt.Errorf("reading flight plan related data primary subfield octet %d: %w", i+1, err)
		}
		bytesRead += n
		f.Data = append(f.Data, octet[0])

		// Check if further extension exists
		hasExtension := (octet[0] & 0x01) != 0
		if !hasExtension {
			break
		}
	}

	// Now read subfields based on the bits set in the primary subfield

	// TAG: bit-24 (bit-8 of first byte) Subfield #1: FPPS Identification Tag
	if len(f.Data) > 0 && (f.Data[0]&0x80) != 0 {
		subfieldData := make([]byte, 2)
		n, err := buf.Read(subfieldData)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading FPPS identification tag subfield: %w", err)
		}
		bytesRead += n
		f.Data = append(f.Data, subfieldData...)
	}

	// CSN: bit-23 (bit-7 of first byte) Subfield #2: Callsign
	if len(f.Data) > 0 && (f.Data[0]&0x40) != 0 {
		subfieldData := make([]byte, 7)
		n, err := buf.Read(subfieldData)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading callsign subfield: %w", err)
		}
		bytesRead += n
		f.Data = append(f.Data, subfieldData...)
	}

	// IFI: bit-22 (bit-6 of first byte) Subfield #3: IFPS_FLIGHT_ID
	if len(f.Data) > 0 && (f.Data[0]&0x20) != 0 {
		subfieldData := make([]byte, 4)
		n, err := buf.Read(subfieldData)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading IFPS flight ID subfield: %w", err)
		}
		bytesRead += n
		f.Data = append(f.Data, subfieldData...)
	}

	// FCT: bit-21 (bit-5 of first byte) Subfield #4: Flight Category
	if len(f.Data) > 0 && (f.Data[0]&0x10) != 0 {
		subfieldData := make([]byte, 1)
		n, err := buf.Read(subfieldData)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading flight category subfield: %w", err)
		}
		bytesRead += n
		f.Data = append(f.Data, subfieldData...)
	}

	// TAC: bit-20 (bit-4 of first byte) Subfield #5: Type of Aircraft
	if len(f.Data) > 0 && (f.Data[0]&0x08) != 0 {
		subfieldData := make([]byte, 4)
		n, err := buf.Read(subfieldData)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading type of aircraft subfield: %w", err)
		}
		bytesRead += n
		f.Data = append(f.Data, subfieldData...)
	}

	// WTC: bit-19 (bit-3 of first byte) Subfield #6: Wake Turbulence Category
	if len(f.Data) > 0 && (f.Data[0]&0x04) != 0 {
		subfieldData := make([]byte, 1)
		n, err := buf.Read(subfieldData)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading wake turbulence category subfield: %w", err)
		}
		bytesRead += n
		f.Data = append(f.Data, subfieldData...)
	}

	// DEP: bit-18 (bit-2 of first byte) Subfield #7: Departure Airport
	if len(f.Data) > 0 && (f.Data[0]&0x02) != 0 {
		subfieldData := make([]byte, 4)
		n, err := buf.Read(subfieldData)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading departure airport subfield: %w", err)
		}
		bytesRead += n
		f.Data = append(f.Data, subfieldData...)
	}

	// Check second byte of primary subfield if it exists
	if len(f.Data) > 1 {
		// DST: bit-16 (bit-8 of second byte) Subfield #8: Destination Airport
		if (f.Data[1] & 0x80) != 0 {
			subfieldData := make([]byte, 4)
			n, err := buf.Read(subfieldData)
			if err != nil {
				return bytesRead + n, fmt.Errorf("reading destination airport subfield: %w", err)
			}
			bytesRead += n
			f.Data = append(f.Data, subfieldData...)
		}

		// RDS: bit-15 (bit-7 of second byte) Subfield #9: Runway Designation
		if (f.Data[1] & 0x40) != 0 {
			subfieldData := make([]byte, 3)
			n, err := buf.Read(subfieldData)
			if err != nil {
				return bytesRead + n, fmt.Errorf("reading runway designation subfield: %w", err)
			}
			bytesRead += n
			f.Data = append(f.Data, subfieldData...)
		}

		// CFL: bit-14 (bit-6 of second byte) Subfield #10: Current Cleared Flight Level
		if (f.Data[1] & 0x20) != 0 {
			subfieldData := make([]byte, 2)
			n, err := buf.Read(subfieldData)
			if err != nil {
				return bytesRead + n, fmt.Errorf("reading current cleared flight level subfield: %w", err)
			}
			bytesRead += n
			f.Data = append(f.Data, subfieldData...)
		}

		// CTL: bit-13 (bit-5 of second byte) Subfield #11: Current Control Position
		if (f.Data[1] & 0x10) != 0 {
			subfieldData := make([]byte, 2)
			n, err := buf.Read(subfieldData)
			if err != nil {
				return bytesRead + n, fmt.Errorf("reading current control position subfield: %w", err)
			}
			bytesRead += n
			f.Data = append(f.Data, subfieldData...)
		}

		// TOD: bit-12 (bit-4 of second byte) Subfield #12: Time of Departure / Arrival
		if (f.Data[1] & 0x08) != 0 {
			// First byte is repetition factor
			repByte := make([]byte, 1)
			n, err := buf.Read(repByte)
			if err != nil {
				return bytesRead + n, fmt.Errorf("reading time of departure/arrival repetition factor: %w", err)
			}
			bytesRead += n
			f.Data = append(f.Data, repByte[0])

			// Each item is 4 octets, and there are repByte[0] items
			itemCount := int(repByte[0])
			for i := 0; i < itemCount; i++ {
				itemData := make([]byte, 4)
				n, err := buf.Read(itemData)
				if err != nil {
					return bytesRead + n, fmt.Errorf("reading time of departure/arrival item %d: %w", i+1, err)
				}
				bytesRead += n
				f.Data = append(f.Data, itemData...)
			}
		}

		// AST: bit-11 (bit-3 of second byte) Subfield #13: Aircraft Stand
		if (f.Data[1] & 0x04) != 0 {
			subfieldData := make([]byte, 6)
			n, err := buf.Read(subfieldData)
			if err != nil {
				return bytesRead + n, fmt.Errorf("reading aircraft stand subfield: %w", err)
			}
			bytesRead += n
			f.Data = append(f.Data, subfieldData...)
		}

		// STS: bit-10 (bit-2 of second byte) Subfield #14: Stand Status
		if (f.Data[1] & 0x02) != 0 {
			subfieldData := make([]byte, 1)
			n, err := buf.Read(subfieldData)
			if err != nil {
				return bytesRead + n, fmt.Errorf("reading stand status subfield: %w", err)
			}
			bytesRead += n
			f.Data = append(f.Data, subfieldData...)
		}
	}

	// Check third byte of primary subfield if it exists
	if len(f.Data) > 2 {
		// STD: bit-8 (bit-8 of third byte) Subfield #15: Standard Instrument Departure
		if (f.Data[2] & 0x80) != 0 {
			subfieldData := make([]byte, 7)
			n, err := buf.Read(subfieldData)
			if err != nil {
				return bytesRead + n, fmt.Errorf("reading standard instrument departure subfield: %w", err)
			}
			bytesRead += n
			f.Data = append(f.Data, subfieldData...)
		}

		// STA: bit-7 (bit-7 of third byte) Subfield #16: Standard Instrument Arrival
		if (f.Data[2] & 0x40) != 0 {
			subfieldData := make([]byte, 7)
			n, err := buf.Read(subfieldData)
			if err != nil {
				return bytesRead + n, fmt.Errorf("reading standard instrument arrival subfield: %w", err)
			}
			bytesRead += n
			f.Data = append(f.Data, subfieldData...)
		}

		// PEM: bit-6 (bit-6 of third byte) Subfield #17: Pre-emergency Mode 3/A code
		if (f.Data[2] & 0x20) != 0 {
			subfieldData := make([]byte, 2)
			n, err := buf.Read(subfieldData)
			if err != nil {
				return bytesRead + n, fmt.Errorf("reading pre-emergency Mode 3/A code subfield: %w", err)
			}
			bytesRead += n
			f.Data = append(f.Data, subfieldData...)
		}

		// PEC: bit-5 (bit-5 of third byte) Subfield #18: Pre-emergency Callsign
		if (f.Data[2] & 0x10) != 0 {
			subfieldData := make([]byte, 7)
			n, err := buf.Read(subfieldData)
			if err != nil {
				return bytesRead + n, fmt.Errorf("reading pre-emergency callsign subfield: %w", err)
			}
			bytesRead += n
			f.Data = append(f.Data, subfieldData...)
		}
	}

	return bytesRead, nil
}

func (f *FlightPlanRelatedData) Encode(buf *bytes.Buffer) (int, error) {
	if len(f.Data) == 0 {
		// If no data, encode a minimal valid value
		return buf.Write([]byte{0})
	}
	return buf.Write(f.Data)
}

func (f *FlightPlanRelatedData) String() string {
	return fmt.Sprintf("FlightPlanRelatedData[%d bytes]", len(f.Data))
}

func (f *FlightPlanRelatedData) Validate() error {
	return nil
}
