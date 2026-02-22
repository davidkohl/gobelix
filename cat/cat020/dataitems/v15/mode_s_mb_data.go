// cat/cat020/dataitems/v15/mode_s_mb_data.go
package v15

import (
	"bytes"
	"fmt"
	"math"
)

// BDSRegister represents a single BDS register with address and data
type BDSRegister struct {
	BDS1 uint8  // BDS Register Address 1 (bits 8-5)
	BDS2 uint8  // BDS Register Address 2 (bits 4-1)
	Data []byte // 7 bytes (56 bits) of BDS data
}

// ModeSMBData implements I020/250 - Mode S MB Data (BDS Register Data)
// Per CAT020 spec v1.11 pages 31-32:
// Repetitive data item: REP + n*(7 bytes data + 1 byte BDS address)
type ModeSMBData struct {
	Registers []BDSRegister
}

// Decode reads the repetitive BDS register data
func (m *ModeSMBData) Decode(buf *bytes.Buffer) (int, error) {
	// Read REP byte
	rep, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading I020/250 REP: %w", err)
	}
	bytesRead := 1

	// Read repetitions (each is 8 bytes: 7 data + 1 address)
	m.Registers = make([]BDSRegister, rep)
	for i := 0; i < int(rep); i++ {
		var data [8]byte
		n, err := buf.Read(data[:])
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading I020/250 BDS register %d: %w", i, err)
		}
		if n != 8 {
			return bytesRead + n, fmt.Errorf("I020/250 BDS register %d: expected 8 bytes, got %d", i, n)
		}

		// Extract BDS address from last byte
		m.Registers[i] = BDSRegister{
			BDS1: (data[7] >> 4) & 0x0F, // bits 8-5
			BDS2: data[7] & 0x0F,        // bits 4-1
			Data: data[0:7],             // 7 bytes of BDS data
		}
		bytesRead += 8
	}

	return bytesRead, nil
}

// Encode writes the repetitive BDS register data
func (m *ModeSMBData) Encode(buf *bytes.Buffer) (int, error) {
	if len(m.Registers) > 255 {
		return 0, fmt.Errorf("I020/250: too many BDS registers (%d > 255)", len(m.Registers))
	}

	// Write REP byte
	err := buf.WriteByte(byte(len(m.Registers)))
	if err != nil {
		return 0, fmt.Errorf("writing I020/250 REP: %w", err)
	}
	bytesWritten := 1

	// Write each BDS register
	for i, reg := range m.Registers {
		if len(reg.Data) != 7 {
			return bytesWritten, fmt.Errorf("I020/250 BDS register %d: invalid data length %d (must be 7)", i, len(reg.Data))
		}

		// Write 7 bytes of data
		n, err := buf.Write(reg.Data)
		if err != nil {
			return bytesWritten + n, fmt.Errorf("writing I020/250 BDS register %d data: %w", i, err)
		}
		bytesWritten += n

		// Write BDS address byte
		addrByte := (reg.BDS1&0x0F)<<4 | (reg.BDS2 & 0x0F)
		err = buf.WriteByte(addrByte)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing I020/250 BDS register %d address: %w", i, err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

// Validate implements the DataItem interface
func (m *ModeSMBData) Validate() error {
	for i, reg := range m.Registers {
		if len(reg.Data) != 7 {
			return fmt.Errorf("I020/250 BDS register %d: invalid data length %d (must be 7)", i, len(reg.Data))
		}
	}
	return nil
}

// String returns a string representation with decoded BDS data
func (m *ModeSMBData) String() string {
	if len(m.Registers) == 0 {
		return "(none)"
	}

	if len(m.Registers) == 1 {
		reg := m.Registers[0]
		decoded := reg.DecodeContent()
		if decoded != "" {
			return decoded
		}
		// Show BDS address and raw hex for unknown registers
		return fmt.Sprintf("BDS %d,%d: %02X %02X %02X %02X %02X %02X %02X",
			reg.BDS1, reg.BDS2,
			reg.Data[0], reg.Data[1], reg.Data[2], reg.Data[3],
			reg.Data[4], reg.Data[5], reg.Data[6])
	}

	// Multiple registers
	result := fmt.Sprintf("%d registers:", len(m.Registers))
	for _, reg := range m.Registers {
		decoded := reg.DecodeContent()
		if decoded != "" {
			result += fmt.Sprintf("\n        %s", decoded)
		} else {
			// Show BDS address and raw hex for unknown registers
			result += fmt.Sprintf("\n        BDS %d,%d: %02X %02X %02X %02X %02X %02X %02X",
				reg.BDS1, reg.BDS2,
				reg.Data[0], reg.Data[1], reg.Data[2], reg.Data[3],
				reg.Data[4], reg.Data[5], reg.Data[6])
		}
	}

	return result
}

// DecodeContent attempts to decode the BDS register content
func (r *BDSRegister) DecodeContent() string {
	switch {
	case r.BDS1 == 1 && r.BDS2 == 7:
		return r.decodeBDS17()
	case r.BDS1 == 2 && r.BDS2 == 0:
		return r.decodeBDS20()
	case r.BDS1 == 4 && r.BDS2 == 0:
		return r.decodeBDS40()
	case r.BDS1 == 5 && r.BDS2 == 0:
		return r.decodeBDS50()
	case r.BDS1 == 6 && r.BDS2 == 0:
		return r.decodeBDS60()
	default:
		return ""
	}
}

// decodeBDS17 decodes BDS 1,7 - Common Usage GICB Capability Report
func (r *BDSRegister) decodeBDS17() string {
	if len(r.Data) != 7 {
		return ""
	}

	// BDS 1,7 indicates which BDS registers the aircraft can provide
	// Bits 1-56 map to BDS capabilities
	var caps []string

	// Common BDS registers to check
	if r.Data[1]&0x10 != 0 {
		caps = append(caps, "2,0") // Aircraft ID
	}
	if r.Data[2]&0x02 != 0 {
		caps = append(caps, "4,0") // Selected Intention
	}
	if r.Data[2]&0x01 != 0 {
		caps = append(caps, "5,0") // Track/Turn
	}
	if r.Data[3]&0x80 != 0 {
		caps = append(caps, "6,0") // Heading/Speed
	}

	if len(caps) == 0 {
		return fmt.Sprintf("GICB Capability: %02X %02X %02X %02X %02X %02X %02X",
			r.Data[0], r.Data[1], r.Data[2], r.Data[3], r.Data[4], r.Data[5], r.Data[6])
	}

	return fmt.Sprintf("GICB Caps: %v", caps)
}

// decodeBDS20 decodes BDS 2,0 - Aircraft Identification
func (r *BDSRegister) decodeBDS20() string {
	if len(r.Data) != 7 {
		return ""
	}

	// BDS 2,0 format (56 bits):
	// Bits 1-8: Reserved (TC - Type Code, typically 1-4 for aircraft categories)
	// Bits 9-14: Character 1 (6 bits, ICAO encoding)
	// Bits 15-20: Character 2
	// Bits 21-26: Character 3
	// Bits 27-32: Character 4
	// Bits 33-38: Character 5
	// Bits 39-44: Character 6
	// Bits 45-50: Character 7
	// Bits 51-56: Character 8

	var chars [8]byte

	// Character 1: bits 9-14 (data[1] bits 7-2)
	chars[0] = (r.Data[1] & 0xFC) >> 2
	// Character 2: bits 15-20 (data[1] bits 1-0 + data[2] bits 7-4)
	chars[1] = ((r.Data[1] & 0x03) << 4) | ((r.Data[2] & 0xF0) >> 4)
	// Character 3: bits 21-26 (data[2] bits 3-0 + data[3] bits 7-6)
	chars[2] = ((r.Data[2] & 0x0F) << 2) | ((r.Data[3] & 0xC0) >> 6)
	// Character 4: bits 27-32 (data[3] bits 5-0)
	chars[3] = r.Data[3] & 0x3F
	// Character 5: bits 33-38 (data[4] bits 7-2)
	chars[4] = (r.Data[4] & 0xFC) >> 2
	// Character 6: bits 39-44 (data[4] bits 1-0 + data[5] bits 7-4)
	chars[5] = ((r.Data[4] & 0x03) << 4) | ((r.Data[5] & 0xF0) >> 4)
	// Character 7: bits 45-50 (data[5] bits 3-0 + data[6] bits 7-6)
	chars[6] = ((r.Data[5] & 0x0F) << 2) | ((r.Data[6] & 0xC0) >> 6)
	// Character 8: bits 51-56 (data[6] bits 5-0)
	chars[7] = r.Data[6] & 0x3F

	// Convert to string using ICAO 6-bit encoding
	var result [8]byte
	for i, c := range chars {
		result[i] = icao6BitToChar(c)
	}

	// Trim trailing spaces
	callsign := string(result[:])
	for len(callsign) > 0 && callsign[len(callsign)-1] == ' ' {
		callsign = callsign[:len(callsign)-1]
	}

	// Show raw hex for debugging if callsign contains invalid chars
	hasInvalid := false
	for _, c := range callsign {
		if c == '?' {
			hasInvalid = true
			break
		}
	}

	if hasInvalid {
		return fmt.Sprintf("Callsign: \"%s\" (raw: %02X %02X %02X %02X %02X %02X %02X)",
			callsign, r.Data[0], r.Data[1], r.Data[2], r.Data[3], r.Data[4], r.Data[5], r.Data[6])
	}

	return fmt.Sprintf("Callsign: %s", callsign)
}

// icao6BitToChar converts ICAO 6-bit encoded character to ASCII
// Per ICAO Annex 10 Vol IV / Doc 9871:
// 0x00       = space (filler)
// 0x01-0x1A  = A-Z (1-26)
// 0x1B-0x1F  = reserved (27-31)
// 0x20       = space (32)
// 0x21-0x2F  = reserved (33-47)
// 0x30-0x39  = 0-9 (48-57)
// 0x3A-0x3F  = reserved (58-63)
func icao6BitToChar(b byte) byte {
	switch {
	case b == 0x00:
		return ' '
	case b >= 0x01 && b <= 0x1A: // 1-26 = A-Z
		return 'A' + b - 1
	case b == 0x20: // 32 = space
		return ' '
	case b >= 0x30 && b <= 0x39: // 48-57 = 0-9
		return '0' + b - 0x30
	default:
		return '?'
	}
}

// decodeBDS40 decodes BDS 4,0 - Selected vertical intention
func (r *BDSRegister) decodeBDS40() string {
	if len(r.Data) != 7 {
		return ""
	}

	var parts []string

	// MCP/FCU Selected Altitude (bits 1-12)
	if (r.Data[0] & 0x80) != 0 { // Status bit 1
		alt := int(r.Data[0]&0x7F)<<5 | int(r.Data[1]&0xF8)>>3
		if alt&0x800 != 0 { // Sign extend
			alt = alt - 0x1000
		}
		altitude := alt * 16
		if altitude >= -1000 && altitude <= 65000 {
			parts = append(parts, fmt.Sprintf("MCP Alt: %d ft", altitude))
		}
	}

	// FMS Selected Altitude (bits 14-25)
	if (r.Data[1] & 0x04) != 0 { // Status bit 14
		alt := int(r.Data[1]&0x03)<<10 | int(r.Data[2])<<2 | int(r.Data[3]&0xC0)>>6
		if alt&0x800 != 0 { // Sign extend
			alt = alt - 0x1000
		}
		altitude := alt * 16
		if altitude >= -1000 && altitude <= 65000 {
			parts = append(parts, fmt.Sprintf("FMS Alt: %d ft", altitude))
		}
	}

	// Barometric Pressure Setting (bits 27-38)
	if (r.Data[3] & 0x20) != 0 { // Status bit 27
		press := int(r.Data[3]&0x1F)<<7 | int(r.Data[4]&0xFE)>>1
		pressure := float64(press)*0.1 + 800.0
		if pressure >= 800.0 && pressure <= 1100.0 {
			parts = append(parts, fmt.Sprintf("QNH: %.1f mb", pressure))
		}
	}

	if len(parts) == 0 {
		return "Selected Vertical Intention (no data)"
	}
	return fmt.Sprintf("Selected Vertical Intention: %s", joinParts(parts))
}

// decodeBDS50 decodes BDS 5,0 - Track and turn report
func (r *BDSRegister) decodeBDS50() string {
	if len(r.Data) != 7 {
		return ""
	}

	var parts []string

	// Roll Angle (bits 1-10)
	if (r.Data[0] & 0x80) != 0 { // Status bit 1
		roll := int(r.Data[0]&0x7F)<<3 | int(r.Data[1]&0xE0)>>5
		if roll&0x200 != 0 { // Sign extend
			roll = roll - 0x400
		}
		rollAngle := float64(roll) * 45.0 / 256.0
		if math.Abs(rollAngle) <= 50.0 {
			parts = append(parts, fmt.Sprintf("Roll: %.1f°", rollAngle))
		}
	}

	// True Track Angle (bits 12-21)
	if (r.Data[1] & 0x10) != 0 { // Status bit 12
		track := int(r.Data[1]&0x0F)<<6 | int(r.Data[2]&0xFC)>>2
		trackAngle := float64(track) * 90.0 / 512.0
		if trackAngle >= 0.0 && trackAngle <= 360.0 {
			parts = append(parts, fmt.Sprintf("Track: %.1f°", trackAngle))
		}
	}

	// Ground Speed (bits 23-32)
	if (r.Data[2] & 0x02) != 0 { // Status bit 23
		gs := int(r.Data[2]&0x01)<<9 | int(r.Data[3])<<1 | int(r.Data[4]&0x80)>>7
		groundSpeed := float64(gs) * 2.0
		if groundSpeed >= 0.0 && groundSpeed <= 750.0 {
			parts = append(parts, fmt.Sprintf("GS: %.0f kt", groundSpeed))
		}
	}

	// Track Angle Rate (bits 34-42)
	if (r.Data[4] & 0x40) != 0 { // Status bit 34
		tar := int(r.Data[4]&0x3F)<<3 | int(r.Data[5]&0xE0)>>5
		if tar&0x100 != 0 { // Sign extend
			tar = tar - 0x200
		}
		trackAngleRate := float64(tar) * 8.0 / 256.0
		if math.Abs(trackAngleRate) <= 16.0 {
			parts = append(parts, fmt.Sprintf("TAR: %.2f°/s", trackAngleRate))
		}
	}

	// True Airspeed (bits 44-53)
	if (r.Data[5] & 0x10) != 0 { // Status bit 44
		tas := int(r.Data[5]&0x0F)<<6 | int(r.Data[6]&0xFC)>>2
		trueAirspeed := float64(tas) * 2.0
		if trueAirspeed >= 0.0 && trueAirspeed <= 750.0 {
			parts = append(parts, fmt.Sprintf("TAS: %.0f kt", trueAirspeed))
		}
	}

	if len(parts) == 0 {
		return "Track and Turn Report (no data)"
	}
	return fmt.Sprintf("Track/Turn: %s", joinParts(parts))
}

// decodeBDS60 decodes BDS 6,0 - Heading and speed report
func (r *BDSRegister) decodeBDS60() string {
	if len(r.Data) != 7 {
		return ""
	}

	var parts []string

	// Magnetic Heading (bits 1-11)
	if (r.Data[0] & 0x80) != 0 { // Status bit 1
		hdg := int(r.Data[0]&0x7F)<<4 | int(r.Data[1]&0xF0)>>4
		heading := float64(hdg) * 90.0 / 512.0
		if heading >= 0.0 && heading <= 360.0 {
			parts = append(parts, fmt.Sprintf("HDG: %.1f°", heading))
		}
	}

	// Indicated Airspeed (bits 13-22)
	if (r.Data[1] & 0x08) != 0 { // Status bit 13
		ias := int(r.Data[1]&0x07)<<7 | int(r.Data[2]&0xFE)>>1
		airspeed := float64(ias)
		if airspeed >= 0.0 && airspeed <= 700.0 {
			parts = append(parts, fmt.Sprintf("IAS: %.0f kt", airspeed))
		}
	}

	// Mach Number (bits 24-33)
	if (r.Data[2] & 0x01) != 0 { // Status bit 24
		mach := int(r.Data[3])<<2 | int(r.Data[4]&0xC0)>>6
		machNumber := float64(mach) * 0.008
		if machNumber >= 0.0 && machNumber <= 1.0 {
			parts = append(parts, fmt.Sprintf("Mach: %.3f", machNumber))
		}
	}

	// Barometric Altitude Rate (bits 35-44)
	if (r.Data[4] & 0x20) != 0 { // Status bit 35
		baro := int(r.Data[4]&0x1F)<<5 | int(r.Data[5]&0xF8)>>3
		if baro&0x200 != 0 { // Sign extend
			baro = baro - 0x400
		}
		baroRate := baro * 32
		if baroRate >= -10000 && baroRate <= 10000 {
			sign := ""
			if baroRate > 0 {
				sign = "+"
			}
			parts = append(parts, fmt.Sprintf("BaroRate: %s%d ft/min", sign, baroRate))
		}
	}

	if len(parts) == 0 {
		return "Heading and Speed Report (no data)"
	}
	return fmt.Sprintf("Hdg/Speed: %s", joinParts(parts))
}

func joinParts(parts []string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += ", "
		}
		result += p
	}
	return result
}
