package v13

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// ModeSRelatedData implements I011/380 - Mode-S / ADS-B Related Data
// Definition: Data specific to Mode-S / ADS-B.
// Format: Compound Data Item, comprising a primary subfield of two octets,
// followed by up to 11 subfields.
type ModeSRelatedData struct {
	// Presence flags
	HasMB  bool // Mode S MB data present
	HasADR bool // Aircraft Address present
	HasCOM bool // Communications capability present
	HasACT bool // Aircraft derived type present
	HasEMC bool // Emitter category present
	HasATC bool // Available technologies present

	// Subfield values
	MBData           []ModeSMBReport // BDS reports
	AircraftAddress  uint32          // 24-bit Mode S address
	CommsCapability  CommsACASStatus // Communications/ACAS capability
	AircraftType     string          // 4-character aircraft type
	EmitterCategory  uint8           // Emitter category code
	AvailableTech    AvailableTechnologies
}

// ModeSMBReport represents a single BDS report
type ModeSMBReport struct {
	Data [7]byte // 56-bit MB data
	BDS1 uint8   // BDS1 address (4 bits)
	BDS2 uint8   // BDS2 address (4 bits)
}

// CommsACASStatus represents I011/380 subfield #4
type CommsACASStatus struct {
	COM  uint8 // Communications capability (0-7)
	STAT uint8 // Flight status (0-15)
	SSC  bool  // Specific service capability
	ARC  bool  // Altitude reporting: 0=100ft, 1=25ft
	AIC  bool  // Aircraft identification capability
	B1A  bool  // BDS 1,0 bit 16
	B1B  uint8 // BDS 1,0 bits 37/40 (4 bits)
	AC   bool  // ACAS operational
	MN   bool  // Multiple navigational aids
	DC   bool  // Differential correction (0=yes, 1=no)
}

// AvailableTechnologies represents I011/380 subfield #11
type AvailableTechnologies struct {
	VDL bool // VDL Mode 4 available (0=yes, 1=no)
	MDS bool // Mode S available (0=yes, 1=no)
	UAT bool // UAT available (0=yes, 1=no)
}

func (m *ModeSRelatedData) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Read primary subfield (2 octets with FX bits)
	octet1, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading I011/380 primary: %w", err)
	}
	bytesRead++

	m.HasMB = (octet1 & 0x80) != 0
	m.HasADR = (octet1 & 0x40) != 0
	// bit 6 is always 0 (subfield #3 never sent)
	m.HasCOM = (octet1 & 0x10) != 0
	// bits 5-3 are always 0 (subfields #5-7 never sent)
	fx := (octet1 & 0x01) != 0

	if fx {
		octet2, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading I011/380 primary ext: %w", err)
		}
		bytesRead++

		m.HasACT = (octet2 & 0x80) != 0
		m.HasEMC = (octet2 & 0x40) != 0
		// bit 6 is always 0 (subfield #10 never sent)
		m.HasATC = (octet2 & 0x10) != 0
		// bits 4-2 are spare
		fx = (octet2 & 0x01) != 0

		// Skip additional extensions
		for fx {
			data, err := buf.ReadByte()
			if err != nil {
				return bytesRead, fmt.Errorf("reading I011/380 ext: %w", err)
			}
			bytesRead++
			fx = (data & 0x01) != 0
		}
	}

	// Read subfields in order
	if m.HasMB {
		// Repetitive: REP (1 byte) + N * 8 bytes
		rep, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading MB REP: %w", err)
		}
		bytesRead++

		m.MBData = make([]ModeSMBReport, rep)
		for i := 0; i < int(rep); i++ {
			var data [8]byte
			n, err := buf.Read(data[:])
			if err != nil {
				return bytesRead + n, fmt.Errorf("reading MB data: %w", err)
			}
			bytesRead += 8

			copy(m.MBData[i].Data[:], data[0:7])
			m.MBData[i].BDS1 = (data[7] >> 4) & 0x0F
			m.MBData[i].BDS2 = data[7] & 0x0F
		}
	}

	if m.HasADR {
		// 3 bytes aircraft address
		var data [3]byte
		n, err := buf.Read(data[:])
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading aircraft address: %w", err)
		}
		bytesRead += 3

		m.AircraftAddress = uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2])
	}

	if m.HasCOM {
		// 3 bytes communications capability
		var data [3]byte
		n, err := buf.Read(data[:])
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading COM capability: %w", err)
		}
		bytesRead += 3

		m.CommsCapability.COM = (data[0] >> 5) & 0x07
		m.CommsCapability.STAT = ((data[0] & 0x1E) >> 1)
		// bit 0 of data[0] is spare
		m.CommsCapability.SSC = (data[1] & 0x80) != 0
		m.CommsCapability.ARC = (data[1] & 0x40) != 0
		m.CommsCapability.AIC = (data[1] & 0x20) != 0
		m.CommsCapability.B1A = (data[1] & 0x10) != 0
		m.CommsCapability.B1B = data[1] & 0x0F
		m.CommsCapability.AC = (data[2] & 0x80) != 0
		m.CommsCapability.MN = (data[2] & 0x40) != 0
		m.CommsCapability.DC = (data[2] & 0x20) != 0
	}

	if m.HasACT {
		// 4 bytes ASCII aircraft type
		var data [4]byte
		n, err := buf.Read(data[:])
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading aircraft type: %w", err)
		}
		bytesRead += 4
		m.AircraftType = string(bytes.TrimRight(data[:], " \x00"))
	}

	if m.HasEMC {
		data, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading emitter category: %w", err)
		}
		bytesRead++
		m.EmitterCategory = data
	}

	if m.HasATC {
		data, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading available tech: %w", err)
		}
		bytesRead++
		m.AvailableTech.VDL = (data & 0x80) != 0
		m.AvailableTech.MDS = (data & 0x40) != 0
		m.AvailableTech.UAT = (data & 0x20) != 0
	}

	return bytesRead, nil
}

func (m *ModeSRelatedData) Encode(buf *bytes.Buffer) (int, error) {
	bytesWritten := 0

	// Build primary subfield
	needSecondOctet := m.HasACT || m.HasEMC || m.HasATC

	var octet1 uint8
	if m.HasMB {
		octet1 |= 0x80
	}
	if m.HasADR {
		octet1 |= 0x40
	}
	// bit 6 always 0
	if m.HasCOM {
		octet1 |= 0x10
	}
	// bits 5-3 always 0
	if needSecondOctet {
		octet1 |= 0x01 // FX
	}

	if err := buf.WriteByte(octet1); err != nil {
		return bytesWritten, fmt.Errorf("writing I011/380 primary: %w", err)
	}
	bytesWritten++

	if needSecondOctet {
		var octet2 uint8
		if m.HasACT {
			octet2 |= 0x80
		}
		if m.HasEMC {
			octet2 |= 0x40
		}
		// bit 6 always 0
		if m.HasATC {
			octet2 |= 0x10
		}
		// FX = 0

		if err := buf.WriteByte(octet2); err != nil {
			return bytesWritten, fmt.Errorf("writing I011/380 ext: %w", err)
		}
		bytesWritten++
	}

	// Write subfields
	if m.HasMB {
		if err := buf.WriteByte(uint8(len(m.MBData))); err != nil {
			return bytesWritten, fmt.Errorf("writing MB REP: %w", err)
		}
		bytesWritten++

		for _, report := range m.MBData {
			var data [8]byte
			copy(data[0:7], report.Data[:])
			data[7] = (report.BDS1 << 4) | (report.BDS2 & 0x0F)
			if _, err := buf.Write(data[:]); err != nil {
				return bytesWritten, fmt.Errorf("writing MB data: %w", err)
			}
			bytesWritten += 8
		}
	}

	if m.HasADR {
		data := []byte{
			byte(m.AircraftAddress >> 16),
			byte(m.AircraftAddress >> 8),
			byte(m.AircraftAddress),
		}
		if _, err := buf.Write(data); err != nil {
			return bytesWritten, fmt.Errorf("writing aircraft address: %w", err)
		}
		bytesWritten += 3
	}

	if m.HasCOM {
		var data [3]byte
		data[0] = (m.CommsCapability.COM << 5) | (m.CommsCapability.STAT << 1)
		data[1] = m.CommsCapability.B1B
		if m.CommsCapability.SSC {
			data[1] |= 0x80
		}
		if m.CommsCapability.ARC {
			data[1] |= 0x40
		}
		if m.CommsCapability.AIC {
			data[1] |= 0x20
		}
		if m.CommsCapability.B1A {
			data[1] |= 0x10
		}
		if m.CommsCapability.AC {
			data[2] |= 0x80
		}
		if m.CommsCapability.MN {
			data[2] |= 0x40
		}
		if m.CommsCapability.DC {
			data[2] |= 0x20
		}
		if _, err := buf.Write(data[:]); err != nil {
			return bytesWritten, fmt.Errorf("writing COM capability: %w", err)
		}
		bytesWritten += 3
	}

	if m.HasACT {
		act := m.AircraftType
		for len(act) < 4 {
			act += " "
		}
		if _, err := buf.Write([]byte(act[:4])); err != nil {
			return bytesWritten, fmt.Errorf("writing aircraft type: %w", err)
		}
		bytesWritten += 4
	}

	if m.HasEMC {
		if err := buf.WriteByte(m.EmitterCategory); err != nil {
			return bytesWritten, fmt.Errorf("writing emitter category: %w", err)
		}
		bytesWritten++
	}

	if m.HasATC {
		var data uint8
		if m.AvailableTech.VDL {
			data |= 0x80
		}
		if m.AvailableTech.MDS {
			data |= 0x40
		}
		if m.AvailableTech.UAT {
			data |= 0x20
		}
		if err := buf.WriteByte(data); err != nil {
			return bytesWritten, fmt.Errorf("writing available tech: %w", err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

func (m *ModeSRelatedData) Validate() error {
	return nil
}

func (m *ModeSRelatedData) String() string {
	parts := []string{}
	if m.HasADR {
		parts = append(parts, fmt.Sprintf("ICAO=%06X", m.AircraftAddress))
	}
	if m.HasMB {
		parts = append(parts, fmt.Sprintf("MB=%d reports", len(m.MBData)))
	}
	if m.HasACT {
		parts = append(parts, fmt.Sprintf("Type=%s", m.AircraftType))
	}
	if m.HasEMC {
		parts = append(parts, fmt.Sprintf("ECAT=%d", m.EmitterCategory))
	}
	return fmt.Sprintf("Mode-S Data: %v", parts)
}

// TrackNumber implements I011/161 - Track Number
// Definition: Identification of a fusion track (single track number)
// Format: Two-octet fixed length Data Item
// Bit 16: spare (0)
// Bits 12-1: Fusion Track Number (0-4095)
type TrackNumber struct {
	Number uint16 // Track number (0-4095)
}

func (t *TrackNumber) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading track number: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("track number: expected 2 bytes, got %d", n)
	}

	raw := binary.BigEndian.Uint16(data[:])
	t.Number = raw & 0x0FFF

	return 2, nil
}

func (t *TrackNumber) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	var data [2]byte
	binary.BigEndian.PutUint16(data[:], t.Number&0x0FFF)

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing track number: %w", err)
	}
	return n, nil
}

func (t *TrackNumber) Validate() error {
	if t.Number > 4095 {
		return fmt.Errorf("track number out of range: %d (max 4095)", t.Number)
	}
	return nil
}

func (t *TrackNumber) String() string {
	return fmt.Sprintf("Track Number: %d", t.Number)
}
