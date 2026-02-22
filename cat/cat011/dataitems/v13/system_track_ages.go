package v13

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// SystemTrackUpdateAges implements I011/290 - System Track Update Ages
// Definition: Ages of the last plot/local track, or the last valid mode-A/mode-C,
// used to update the system track.
// Format: Compound Data Item, comprising a primary subfield of two octets,
// followed by up to twelve subfields.
type SystemTrackUpdateAges struct {
	// Presence flags
	HasPSR bool // PSR age present
	HasSSR bool // SSR age present
	HasMDA bool // Mode A age present
	HasMFL bool // Measured Flight Level age present
	HasMDS bool // Mode S age present
	HasADS bool // ADS age present
	HasADB bool // ADS-B age present
	HasMD1 bool // Mode 1 age present
	HasMD2 bool // Mode 2 age present
	HasLOP bool // Loop age present
	HasTRK bool // Track age present
	HasMUL bool // Multilateration age present

	// Age values in seconds (LSB = 0.25s for most, except ADS which is 2 bytes)
	PSRAge float64 // Age of last PSR report
	SSRAge float64 // Age of last SSR report
	MDAAge float64 // Age of last Mode A report
	MFLAge float64 // Age of last Mode C report
	MDSAge float64 // Age of last Mode S report
	ADSAge float64 // Age of last ADS report (2 bytes, max >4 hours)
	ADBAge float64 // Age of last ADS-B report
	MD1Age float64 // Age of last Mode 1 report
	MD2Age float64 // Age of last Mode 2 report
	LOPAge float64 // Age of last loop detection
	TRKAge float64 // Actual track age since first occurrence
	MULAge float64 // Age of last multilateration detection
}

const ageLSB = 0.25 // seconds

func (s *SystemTrackUpdateAges) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Read primary subfield (2 octets with FX bits)
	octet1, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading I011/290 primary: %w", err)
	}
	bytesRead++

	s.HasPSR = (octet1 & 0x80) != 0
	s.HasSSR = (octet1 & 0x40) != 0
	s.HasMDA = (octet1 & 0x20) != 0
	s.HasMFL = (octet1 & 0x10) != 0
	s.HasMDS = (octet1 & 0x08) != 0
	s.HasADS = (octet1 & 0x04) != 0
	s.HasADB = (octet1 & 0x02) != 0
	fx := (octet1 & 0x01) != 0

	if fx {
		octet2, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading I011/290 primary ext: %w", err)
		}
		bytesRead++

		s.HasMD1 = (octet2 & 0x80) != 0
		s.HasMD2 = (octet2 & 0x40) != 0
		s.HasLOP = (octet2 & 0x20) != 0
		s.HasTRK = (octet2 & 0x10) != 0
		s.HasMUL = (octet2 & 0x08) != 0
		// bits 3-2 are spare
		// Skip additional extensions
		fx = (octet2 & 0x01) != 0
		for fx {
			data, err := buf.ReadByte()
			if err != nil {
				return bytesRead, fmt.Errorf("reading I011/290 primary ext: %w", err)
			}
			bytesRead++
			fx = (data & 0x01) != 0
		}
	}

	// Read subfields in order
	if s.HasPSR {
		data, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading PSR age: %w", err)
		}
		bytesRead++
		s.PSRAge = float64(data) * ageLSB
	}

	if s.HasSSR {
		data, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading SSR age: %w", err)
		}
		bytesRead++
		s.SSRAge = float64(data) * ageLSB
	}

	if s.HasMDA {
		data, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading MDA age: %w", err)
		}
		bytesRead++
		s.MDAAge = float64(data) * ageLSB
	}

	if s.HasMFL {
		data, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading MFL age: %w", err)
		}
		bytesRead++
		s.MFLAge = float64(data) * ageLSB
	}

	if s.HasMDS {
		data, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading MDS age: %w", err)
		}
		bytesRead++
		s.MDSAge = float64(data) * ageLSB
	}

	if s.HasADS {
		// ADS age is 2 bytes
		var data [2]byte
		n, err := buf.Read(data[:])
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading ADS age: %w", err)
		}
		bytesRead += 2
		s.ADSAge = float64(binary.BigEndian.Uint16(data[:])) * ageLSB
	}

	if s.HasADB {
		data, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading ADB age: %w", err)
		}
		bytesRead++
		s.ADBAge = float64(data) * ageLSB
	}

	if s.HasMD1 {
		data, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading MD1 age: %w", err)
		}
		bytesRead++
		s.MD1Age = float64(data) * ageLSB
	}

	if s.HasMD2 {
		data, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading MD2 age: %w", err)
		}
		bytesRead++
		s.MD2Age = float64(data) * ageLSB
	}

	if s.HasLOP {
		data, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading LOP age: %w", err)
		}
		bytesRead++
		s.LOPAge = float64(data) * ageLSB
	}

	if s.HasTRK {
		data, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading TRK age: %w", err)
		}
		bytesRead++
		s.TRKAge = float64(data) * ageLSB
	}

	if s.HasMUL {
		data, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading MUL age: %w", err)
		}
		bytesRead++
		s.MULAge = float64(data) * ageLSB
	}

	return bytesRead, nil
}

func (s *SystemTrackUpdateAges) Encode(buf *bytes.Buffer) (int, error) {
	bytesWritten := 0

	// Build primary subfield
	needSecondOctet := s.HasMD1 || s.HasMD2 || s.HasLOP || s.HasTRK || s.HasMUL

	var octet1 uint8
	if s.HasPSR {
		octet1 |= 0x80
	}
	if s.HasSSR {
		octet1 |= 0x40
	}
	if s.HasMDA {
		octet1 |= 0x20
	}
	if s.HasMFL {
		octet1 |= 0x10
	}
	if s.HasMDS {
		octet1 |= 0x08
	}
	if s.HasADS {
		octet1 |= 0x04
	}
	if s.HasADB {
		octet1 |= 0x02
	}
	if needSecondOctet {
		octet1 |= 0x01 // FX
	}

	if err := buf.WriteByte(octet1); err != nil {
		return bytesWritten, fmt.Errorf("writing I011/290 primary: %w", err)
	}
	bytesWritten++

	if needSecondOctet {
		var octet2 uint8
		if s.HasMD1 {
			octet2 |= 0x80
		}
		if s.HasMD2 {
			octet2 |= 0x40
		}
		if s.HasLOP {
			octet2 |= 0x20
		}
		if s.HasTRK {
			octet2 |= 0x10
		}
		if s.HasMUL {
			octet2 |= 0x08
		}
		// FX = 0

		if err := buf.WriteByte(octet2); err != nil {
			return bytesWritten, fmt.Errorf("writing I011/290 primary ext: %w", err)
		}
		bytesWritten++
	}

	// Write subfields
	if s.HasPSR {
		if err := buf.WriteByte(uint8(s.PSRAge / ageLSB)); err != nil {
			return bytesWritten, fmt.Errorf("writing PSR age: %w", err)
		}
		bytesWritten++
	}

	if s.HasSSR {
		if err := buf.WriteByte(uint8(s.SSRAge / ageLSB)); err != nil {
			return bytesWritten, fmt.Errorf("writing SSR age: %w", err)
		}
		bytesWritten++
	}

	if s.HasMDA {
		if err := buf.WriteByte(uint8(s.MDAAge / ageLSB)); err != nil {
			return bytesWritten, fmt.Errorf("writing MDA age: %w", err)
		}
		bytesWritten++
	}

	if s.HasMFL {
		if err := buf.WriteByte(uint8(s.MFLAge / ageLSB)); err != nil {
			return bytesWritten, fmt.Errorf("writing MFL age: %w", err)
		}
		bytesWritten++
	}

	if s.HasMDS {
		if err := buf.WriteByte(uint8(s.MDSAge / ageLSB)); err != nil {
			return bytesWritten, fmt.Errorf("writing MDS age: %w", err)
		}
		bytesWritten++
	}

	if s.HasADS {
		var data [2]byte
		binary.BigEndian.PutUint16(data[:], uint16(s.ADSAge/ageLSB))
		if _, err := buf.Write(data[:]); err != nil {
			return bytesWritten, fmt.Errorf("writing ADS age: %w", err)
		}
		bytesWritten += 2
	}

	if s.HasADB {
		if err := buf.WriteByte(uint8(s.ADBAge / ageLSB)); err != nil {
			return bytesWritten, fmt.Errorf("writing ADB age: %w", err)
		}
		bytesWritten++
	}

	if s.HasMD1 {
		if err := buf.WriteByte(uint8(s.MD1Age / ageLSB)); err != nil {
			return bytesWritten, fmt.Errorf("writing MD1 age: %w", err)
		}
		bytesWritten++
	}

	if s.HasMD2 {
		if err := buf.WriteByte(uint8(s.MD2Age / ageLSB)); err != nil {
			return bytesWritten, fmt.Errorf("writing MD2 age: %w", err)
		}
		bytesWritten++
	}

	if s.HasLOP {
		if err := buf.WriteByte(uint8(s.LOPAge / ageLSB)); err != nil {
			return bytesWritten, fmt.Errorf("writing LOP age: %w", err)
		}
		bytesWritten++
	}

	if s.HasTRK {
		if err := buf.WriteByte(uint8(s.TRKAge / ageLSB)); err != nil {
			return bytesWritten, fmt.Errorf("writing TRK age: %w", err)
		}
		bytesWritten++
	}

	if s.HasMUL {
		if err := buf.WriteByte(uint8(s.MULAge / ageLSB)); err != nil {
			return bytesWritten, fmt.Errorf("writing MUL age: %w", err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

func (s *SystemTrackUpdateAges) Validate() error {
	return nil
}

func (s *SystemTrackUpdateAges) String() string {
	ages := []string{}
	if s.HasPSR {
		ages = append(ages, fmt.Sprintf("PSR=%.2fs", s.PSRAge))
	}
	if s.HasSSR {
		ages = append(ages, fmt.Sprintf("SSR=%.2fs", s.SSRAge))
	}
	if s.HasMDA {
		ages = append(ages, fmt.Sprintf("MDA=%.2fs", s.MDAAge))
	}
	if s.HasMFL {
		ages = append(ages, fmt.Sprintf("MFL=%.2fs", s.MFLAge))
	}
	if s.HasMDS {
		ages = append(ages, fmt.Sprintf("MDS=%.2fs", s.MDSAge))
	}
	if s.HasADS {
		ages = append(ages, fmt.Sprintf("ADS=%.2fs", s.ADSAge))
	}
	if s.HasADB {
		ages = append(ages, fmt.Sprintf("ADB=%.2fs", s.ADBAge))
	}
	if s.HasMD1 {
		ages = append(ages, fmt.Sprintf("MD1=%.2fs", s.MD1Age))
	}
	if s.HasMD2 {
		ages = append(ages, fmt.Sprintf("MD2=%.2fs", s.MD2Age))
	}
	if s.HasLOP {
		ages = append(ages, fmt.Sprintf("LOP=%.2fs", s.LOPAge))
	}
	if s.HasTRK {
		ages = append(ages, fmt.Sprintf("TRK=%.2fs", s.TRKAge))
	}
	if s.HasMUL {
		ages = append(ages, fmt.Sprintf("MUL=%.2fs", s.MULAge))
	}
	return fmt.Sprintf("Track Ages: %v", ages)
}
