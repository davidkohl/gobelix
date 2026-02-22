// dataitems/cat062/measured_information.go
package v120

import (
	"bytes"
	"fmt"
	"strings"
)

// MeasuredInformation implements I062/340
// All measured data related to the last report used to update the track
//
// Subfields (6 total):
// #1 (bit 8): SID - Sensor Identification (2 octets)
// #2 (bit 7): POS - Measured Position (4 octets)
// #3 (bit 6): HEI - Measured 3-D Height (2 octets)
// #4 (bit 5): MDC - Last Measured Mode C Code (2 octets)
// #5 (bit 4): MDA - Last Measured Mode 3/A Code (2 octets)
// #6 (bit 3): TYP - Report Type (1 octet)
type MeasuredInformation struct {
	// Subfield #1: Sensor Identification
	// SAC/SIC of the sensor that provided the last measurement
	SensorSAC *uint8
	SensorSIC *uint8

	// Subfield #2: Measured Position (Cartesian coordinates)
	// In meters, LSB = 0.5m for both Rho and Theta
	MeasuredPositionRho   *int16   // Rho component (radial distance)
	MeasuredPositionTheta *uint16  // Theta component (azimuth angle)

	// Subfield #3: Measured 3-D Height
	// In feet, LSB = 25 feet
	Measured3DHeight *int16

	// Subfield #4: Last Measured Mode C Code
	ModeCV        *bool    // Mode C code validation
	ModeC         *bool    // Mode C code garbled
	ModeCAltitude *float64 // Mode C altitude in feet (converted from 100ft Gray code)

	// Subfield #5: Last Measured Mode 3/A Code
	Mode3AV    *bool   // Mode 3/A code validation
	Mode3AG    *bool   // Mode 3/A code garbled
	Mode3AL    *bool   // Mode 3/A code smoothed
	Mode3ACode *uint16 // 12-bit Mode 3/A code (octal representation)

	// Subfield #6: Report Type
	ReportTypeTYP *uint8 // Type of report (5 bits)
	ReportTypeSIM *bool  // Simulated target report
	ReportTypeRAB *bool  // Test target report
	ReportTypeTST *bool  // Selected altitude available
}

// Decode decodes the Measured Information compound data item from the buffer
func (m *MeasuredInformation) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Read primary subfield (1 octet, FX bit at bit-1)
	if buf.Len() < 1 {
		return 0, fmt.Errorf("buffer too short for measured information primary subfield")
	}

	primaryByte, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading measured information primary subfield: %w", err)
	}
	bytesRead++

	// Process each subfield based on bit presence
	subfieldIndex := 0

	// Bit 8: SID - Sensor Identification
	if (primaryByte & 0x80) != 0 {
		n, err := m.decodeSensorIdentification(buf)
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("decoding sensor identification: %w", err)
		}
	}
	subfieldIndex++

	// Bit 7: POS - Measured Position
	if (primaryByte & 0x40) != 0 {
		n, err := m.decodeMeasuredPosition(buf)
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("decoding measured position: %w", err)
		}
	}
	subfieldIndex++

	// Bit 6: HEI - Measured 3-D Height
	if (primaryByte & 0x20) != 0 {
		n, err := m.decodeMeasured3DHeight(buf)
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("decoding measured 3-D height: %w", err)
		}
	}
	subfieldIndex++

	// Bit 5: MDC - Last Measured Mode C Code
	if (primaryByte & 0x10) != 0 {
		n, err := m.decodeModeC(buf)
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("decoding mode C code: %w", err)
		}
	}
	subfieldIndex++

	// Bit 4: MDA - Last Measured Mode 3/A Code
	if (primaryByte & 0x08) != 0 {
		n, err := m.decodeMode3A(buf)
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("decoding mode 3/A code: %w", err)
		}
	}
	subfieldIndex++

	// Bit 3: TYP - Report Type
	if (primaryByte & 0x04) != 0 {
		n, err := m.decodeReportType(buf)
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("decoding report type: %w", err)
		}
	}

	// Bit 2: Spare

	// Bit 1: FX - Extension indicator (not used per specification)
	if (primaryByte & 0x01) != 0 {
		// Extension not defined in specification
		return bytesRead, fmt.Errorf("unexpected extension bit in measured information")
	}

	return bytesRead, nil
}

// decodeSensorIdentification decodes Subfield #1: Sensor Identification (2 octets)
func (m *MeasuredInformation) decodeSensorIdentification(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 2 {
		return 0, fmt.Errorf("buffer too short for sensor identification")
	}

	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, err
	}

	sac := data[0]
	sic := data[1]

	m.SensorSAC = &sac
	m.SensorSIC = &sic

	return 2, nil
}

// decodeMeasuredPosition decodes Subfield #2: Measured Position (4 octets)
// Rho (2 octets): Calculated range in polar coordinates, LSB = 0.5m
// Theta (2 octets): Calculated azimuth in polar coordinates, LSB = 360°/2^16
func (m *MeasuredInformation) decodeMeasuredPosition(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 4 {
		return 0, fmt.Errorf("buffer too short for measured position")
	}

	var data [4]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, err
	}

	// Rho: signed 16-bit, LSB = 0.5m
	rho := int16(data[0])<<8 | int16(data[1])

	// Theta: unsigned 16-bit, LSB = 360°/2^16
	theta := uint16(data[2])<<8 | uint16(data[3])

	m.MeasuredPositionRho = &rho
	m.MeasuredPositionTheta = &theta

	return 4, nil
}

// decodeMeasured3DHeight decodes Subfield #3: Measured 3-D Height (2 octets)
// In feet, LSB = 25 feet, signed
func (m *MeasuredInformation) decodeMeasured3DHeight(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 2 {
		return 0, fmt.Errorf("buffer too short for measured 3-D height")
	}

	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, err
	}

	// Signed 16-bit, LSB = 25 feet
	height := int16(data[0])<<8 | int16(data[1])

	m.Measured3DHeight = &height

	return 2, nil
}

// decodeModeC decodes Subfield #4: Last Measured Mode C Code (2 octets)
//
// Octet 1:
// - Bit 16: V (code validated)
// - Bit 15: G (code garbled)
// - Bits 14-1: Mode C altitude (14-bit two's complement, LSB = 1/4 FL = 25 ft)
func (m *MeasuredInformation) decodeModeC(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 2 {
		return 0, fmt.Errorf("buffer too short for mode C code")
	}

	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, err
	}

	// Bit 16: V - code validated
	v := (data[0] & 0x80) != 0
	m.ModeCV = &v

	// Bit 15: G - code garbled
	g := (data[0] & 0x40) != 0
	m.ModeC = &g

	// Bits 14-1: Mode C altitude (14-bit two's complement, LSB = 1/4 FL = 25 ft)
	// Extract 14-bit value
	modeCBits := uint16(data[0]&0x3F)<<8 | uint16(data[1])

	// Convert from 14-bit two's complement
	var modeCValue int16
	if (modeCBits & 0x2000) != 0 {
		// Negative value - sign extend from 14 bits
		modeCValue = int16(modeCBits | 0xC000)
	} else {
		modeCValue = int16(modeCBits)
	}

	// LSB = 1/4 FL = 25 feet, store as flight level
	altitude := float64(modeCValue) * 0.25 * 100.0 // Convert FL to feet
	m.ModeCAltitude = &altitude

	return 2, nil
}

// decodeMode3A decodes Subfield #5: Last Measured Mode 3/A Code (2 octets)
//
// Octet 1:
// - Bit 16: V (code validated)
// - Bit 15: G (code garbled)
// - Bit 14: L (code smoothed)
// - Bit 13: Spare
// - Bits 12-1: 12-bit Mode 3/A code (octal representation)
func (m *MeasuredInformation) decodeMode3A(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 2 {
		return 0, fmt.Errorf("buffer too short for mode 3/A code")
	}

	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, err
	}

	// Bit 16: V - code validated
	v := (data[0] & 0x80) != 0
	m.Mode3AV = &v

	// Bit 15: G - code garbled
	g := (data[0] & 0x40) != 0
	m.Mode3AG = &g

	// Bit 14: L - code smoothed
	l := (data[0] & 0x20) != 0
	m.Mode3AL = &l

	// Bit 13: Spare

	// Bits 12-1: 12-bit Mode 3/A code
	code := (uint16(data[0]&0x0F) << 8) | uint16(data[1])
	m.Mode3ACode = &code

	return 2, nil
}

// decodeReportType decodes Subfield #6: Report Type (1 octet)
//
// - Bits 8-4: TYP (Type of report, 5 bits)
// - Bit 3: SIM (Simulated target report)
// - Bit 2: RAB (Report from field monitor)
// - Bit 1: TST (Test target report)
func (m *MeasuredInformation) decodeReportType(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 1 {
		return 0, fmt.Errorf("buffer too short for report type")
	}

	data, err := buf.ReadByte()
	if err != nil {
		return 0, err
	}

	// Bits 8-4: TYP (5 bits)
	typ := (data & 0xF8) >> 3
	m.ReportTypeTYP = &typ

	// Bit 3: SIM
	sim := (data & 0x04) != 0
	m.ReportTypeSIM = &sim

	// Bit 2: RAB
	rab := (data & 0x02) != 0
	m.ReportTypeRAB = &rab

	// Bit 1: TST
	tst := (data & 0x01) != 0
	m.ReportTypeTST = &tst

	return 1, nil
}

// Encode encodes the Measured Information compound data item to the buffer
func (m *MeasuredInformation) Encode(buf *bytes.Buffer) (int, error) {
	bytesWritten := 0

	// Build primary subfield
	var primaryByte byte = 0x00

	if m.SensorSAC != nil && m.SensorSIC != nil {
		primaryByte |= 0x80 // Bit 8: SID
	}
	if m.MeasuredPositionRho != nil && m.MeasuredPositionTheta != nil {
		primaryByte |= 0x40 // Bit 7: POS
	}
	if m.Measured3DHeight != nil {
		primaryByte |= 0x20 // Bit 6: HEI
	}
	if m.ModeCV != nil || m.ModeC != nil || m.ModeCAltitude != nil {
		primaryByte |= 0x10 // Bit 5: MDC
	}
	if m.Mode3AV != nil || m.Mode3AG != nil || m.Mode3AL != nil || m.Mode3ACode != nil {
		primaryByte |= 0x08 // Bit 4: MDA
	}
	if m.ReportTypeTYP != nil {
		primaryByte |= 0x04 // Bit 3: TYP
	}
	// Bit 2: Spare
	// Bit 1: FX (always 0, no extension)

	err := buf.WriteByte(primaryByte)
	if err != nil {
		return 0, err
	}
	bytesWritten++

	// Encode subfields
	if (primaryByte & 0x80) != 0 {
		n, err := m.encodeSensorIdentification(buf)
		bytesWritten += n
		if err != nil {
			return bytesWritten, err
		}
	}

	if (primaryByte & 0x40) != 0 {
		n, err := m.encodeMeasuredPosition(buf)
		bytesWritten += n
		if err != nil {
			return bytesWritten, err
		}
	}

	if (primaryByte & 0x20) != 0 {
		n, err := m.encodeMeasured3DHeight(buf)
		bytesWritten += n
		if err != nil {
			return bytesWritten, err
		}
	}

	if (primaryByte & 0x10) != 0 {
		n, err := m.encodeModeC(buf)
		bytesWritten += n
		if err != nil {
			return bytesWritten, err
		}
	}

	if (primaryByte & 0x08) != 0 {
		n, err := m.encodeMode3A(buf)
		bytesWritten += n
		if err != nil {
			return bytesWritten, err
		}
	}

	if (primaryByte & 0x04) != 0 {
		n, err := m.encodeReportType(buf)
		bytesWritten += n
		if err != nil {
			return bytesWritten, err
		}
	}

	return bytesWritten, nil
}

// encodeSensorIdentification encodes Subfield #1
func (m *MeasuredInformation) encodeSensorIdentification(buf *bytes.Buffer) (int, error) {
	data := []byte{
		*m.SensorSAC,
		*m.SensorSIC,
	}
	return buf.Write(data)
}

// encodeMeasuredPosition encodes Subfield #2
func (m *MeasuredInformation) encodeMeasuredPosition(buf *bytes.Buffer) (int, error) {
	data := []byte{
		byte(*m.MeasuredPositionRho >> 8),
		byte(*m.MeasuredPositionRho),
		byte(*m.MeasuredPositionTheta >> 8),
		byte(*m.MeasuredPositionTheta),
	}
	return buf.Write(data)
}

// encodeMeasured3DHeight encodes Subfield #3
func (m *MeasuredInformation) encodeMeasured3DHeight(buf *bytes.Buffer) (int, error) {
	data := []byte{
		byte(*m.Measured3DHeight >> 8),
		byte(*m.Measured3DHeight),
	}
	return buf.Write(data)
}

// encodeModeC encodes Subfield #4
func (m *MeasuredInformation) encodeModeC(buf *bytes.Buffer) (int, error) {
	var byte1, byte2 byte

	if m.ModeCV != nil && *m.ModeCV {
		byte1 |= 0x80
	}
	if m.ModeC != nil && *m.ModeC {
		byte1 |= 0x40
	}

	if m.ModeCAltitude != nil {
		// Convert altitude from feet to 1/4 FL units (LSB = 25 ft)
		modeCValue := int16(*m.ModeCAltitude / 100.0 / 0.25)

		// Pack into 14 bits (already in two's complement form)
		modeCBits := uint16(modeCValue) & 0x3FFF

		byte1 |= byte((modeCBits >> 8) & 0x3F)
		byte2 = byte(modeCBits)
	}

	data := []byte{byte1, byte2}
	return buf.Write(data)
}

// encodeMode3A encodes Subfield #5
func (m *MeasuredInformation) encodeMode3A(buf *bytes.Buffer) (int, error) {
	var byte1, byte2 byte

	if m.Mode3AV != nil && *m.Mode3AV {
		byte1 |= 0x80
	}
	if m.Mode3AG != nil && *m.Mode3AG {
		byte1 |= 0x40
	}
	if m.Mode3AL != nil && *m.Mode3AL {
		byte1 |= 0x20
	}

	if m.Mode3ACode != nil {
		// Pack 12-bit code
		byte1 |= byte((*m.Mode3ACode >> 8) & 0x0F)
		byte2 = byte(*m.Mode3ACode)
	}

	data := []byte{byte1, byte2}
	return buf.Write(data)
}

// encodeReportType encodes Subfield #6
func (m *MeasuredInformation) encodeReportType(buf *bytes.Buffer) (int, error) {
	var data byte

	if m.ReportTypeTYP != nil {
		data |= (*m.ReportTypeTYP & 0x1F) << 3
	}
	if m.ReportTypeSIM != nil && *m.ReportTypeSIM {
		data |= 0x04
	}
	if m.ReportTypeRAB != nil && *m.ReportTypeRAB {
		data |= 0x02
	}
	if m.ReportTypeTST != nil && *m.ReportTypeTST {
		data |= 0x01
	}

	return buf.Write([]byte{data})
}

// String returns a human-readable representation
func (m *MeasuredInformation) String() string {
	var parts []string

	if m.SensorSAC != nil && m.SensorSIC != nil {
		parts = append(parts, fmt.Sprintf("Sensor=%d/%d", *m.SensorSAC, *m.SensorSIC))
	}

	if m.MeasuredPositionRho != nil && m.MeasuredPositionTheta != nil {
		rhoMeters := float64(*m.MeasuredPositionRho) * 0.5
		thetaDegrees := float64(*m.MeasuredPositionTheta) * 360.0 / 65536.0
		parts = append(parts, fmt.Sprintf("Pos=%.1fm@%.2f°", rhoMeters, thetaDegrees))
	}

	if m.Measured3DHeight != nil {
		heightFeet := float64(*m.Measured3DHeight) * 25.0
		parts = append(parts, fmt.Sprintf("Height=%.0f ft", heightFeet))
	}

	if m.ModeCAltitude != nil {
		var flags []string
		if m.ModeCV != nil && *m.ModeCV {
			flags = append(flags, "V")
		}
		if m.ModeC != nil && *m.ModeC {
			flags = append(flags, "G")
		}
		flagStr := ""
		if len(flags) > 0 {
			flagStr = fmt.Sprintf("[%s]", strings.Join(flags, ","))
		}
		parts = append(parts, fmt.Sprintf("ModeC=%.0f ft%s", *m.ModeCAltitude, flagStr))
	}

	if m.Mode3ACode != nil {
		var flags []string
		if m.Mode3AV != nil && *m.Mode3AV {
			flags = append(flags, "V")
		}
		if m.Mode3AG != nil && *m.Mode3AG {
			flags = append(flags, "G")
		}
		if m.Mode3AL != nil && *m.Mode3AL {
			flags = append(flags, "L")
		}
		flagStr := ""
		if len(flags) > 0 {
			flagStr = fmt.Sprintf("[%s]", strings.Join(flags, ","))
		}
		parts = append(parts, fmt.Sprintf("Mode3A=%04o%s", *m.Mode3ACode, flagStr))
	}

	if m.ReportTypeTYP != nil {
		var flags []string
		if m.ReportTypeSIM != nil && *m.ReportTypeSIM {
			flags = append(flags, "SIM")
		}
		if m.ReportTypeRAB != nil && *m.ReportTypeRAB {
			flags = append(flags, "RAB")
		}
		if m.ReportTypeTST != nil && *m.ReportTypeTST {
			flags = append(flags, "TST")
		}
		flagStr := ""
		if len(flags) > 0 {
			flagStr = fmt.Sprintf("[%s]", strings.Join(flags, ","))
		}
		parts = append(parts, fmt.Sprintf("Type=%d%s", *m.ReportTypeTYP, flagStr))
	}

	if len(parts) == 0 {
		return "MeasuredInformation{empty}"
	}

	return fmt.Sprintf("MeasuredInformation{%s}", strings.Join(parts, ", "))
}

// Validate performs validation checks on the data
func (m *MeasuredInformation) Validate() error {
	// Validate report type if present
	if m.ReportTypeTYP != nil {
		if *m.ReportTypeTYP > 31 {
			return fmt.Errorf("report type TYP out of range [0,31]: %d", *m.ReportTypeTYP)
		}
	}

	// Validate Mode 3/A code if present
	if m.Mode3ACode != nil {
		if *m.Mode3ACode > 0x0FFF {
			return fmt.Errorf("mode 3/A code out of range [0,4095]: %d", *m.Mode3ACode)
		}
	}

	return nil
}
