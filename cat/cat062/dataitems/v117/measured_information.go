// dataitems/cat062/measured_information.go
package v117

import (
	"bytes"
	"fmt"
	"strings"
)

// MeasuredInformation implements I062/340
// All measured data related to the last report used to update the track
type MeasuredInformation struct {
	// Subfield #1: Sensor Identification
	SensorSAC *uint8 // System Area Code
	SensorSIC *uint8 // System Identification Code

	// Subfield #2: Measured Position (polar coordinates)
	MeasuredRange   *float64 // In nautical miles
	MeasuredAzimuth *float64 // In degrees (0-360)

	// Subfield #3: Measured 3-D Height
	Measured3DHeight *float64 // In feet

	// Subfield #4: Last Measured Mode C code
	LastModeC          *float64 // In flight levels (FL)
	LastModeCValidated bool     // Whether the Mode C was validated
	LastModeCGarbled   bool     // Whether the Mode C was garbled

	// Subfield #5: Last Measured Mode 3/A code
	LastMode3A          *uint16 // Octal Mode 3/A code (0000-7777)
	LastMode3AValidated bool    // Whether the Mode 3/A was validated
	LastMode3AGarbled   bool    // Whether the Mode 3/A was garbled
	LastMode3ASmoothed  bool    // Whether the Mode 3/A was derived from local tracker

	// Subfield #6: Report Type
	ReportType      *uint8 // Type of detection (bits 8-6)
	SimulatedTarget bool   // Whether this is a simulated target
	ReportFromRamp  bool   // Whether this is from a field monitor
	TestTarget      bool   // Whether this is a test target
}

// Decode parses an ASTERIX Category 062 I340 data item from the buffer
func (m *MeasuredInformation) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Read the primary subfield (FSPEC octet)
	if buf.Len() < 1 {
		return 0, fmt.Errorf("buffer too short for measured information FSPEC")
	}

	fspec, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading measured information FSPEC: %w", err)
	}
	bytesRead++

	// Check for FX extension bit
	hasExtension := (fspec & 0x01) != 0
	if hasExtension {
		// According to the spec, there are no extensions for I062/340
		return bytesRead, fmt.Errorf("unexpected FX bit set in measured information FSPEC")
	}

	// Subfield #1: Sensor Identification (if bit 8 is set)
	if (fspec & 0x80) != 0 {
		if buf.Len() < 2 {
			return bytesRead, fmt.Errorf("buffer too short for sensor identification")
		}

		data := make([]byte, 2)
		n, err := buf.Read(data)
		if err != nil || n != 2 {
			return bytesRead + n, fmt.Errorf("reading sensor identification: %w", err)
		}
		bytesRead += n

		sac := data[0]
		sic := data[1]
		m.SensorSAC = &sac
		m.SensorSIC = &sic
	}

	// Subfield #2: Measured Position (if bit 7 is set)
	if (fspec & 0x40) != 0 {
		if buf.Len() < 4 {
			return bytesRead, fmt.Errorf("buffer too short for measured position")
		}

		data := make([]byte, 4)
		n, err := buf.Read(data)
		if err != nil || n != 4 {
			return bytesRead + n, fmt.Errorf("reading measured position: %w", err)
		}
		bytesRead += n

		// Extract range (16 bits)
		rangeBits := uint16(data[0])<<8 | uint16(data[1])
		// LSB = 1/256 NM
		measuredRange := float64(rangeBits) / 256.0
		m.MeasuredRange = &measuredRange

		// Extract azimuth (16 bits)
		azimuthBits := uint16(data[2])<<8 | uint16(data[3])
		// LSB = 360/2^16 degrees
		measuredAzimuth := float64(azimuthBits) * 360.0 / 65536.0
		m.MeasuredAzimuth = &measuredAzimuth
	}

	// Subfield #3: Measured 3-D Height (if bit 6 is set)
	if (fspec & 0x20) != 0 {
		if buf.Len() < 2 {
			return bytesRead, fmt.Errorf("buffer too short for measured 3-D height")
		}

		data := make([]byte, 2)
		n, err := buf.Read(data)
		if err != nil || n != 2 {
			return bytesRead + n, fmt.Errorf("reading measured 3-D height: %w", err)
		}
		bytesRead += n

		// Extract height (16 bits)
		heightBits := uint16(data[0])<<8 | uint16(data[1])
		// LSB = 25 feet
		height := float64(heightBits) * 25.0
		m.Measured3DHeight = &height
	}

	// Subfield #4: Last Measured Mode C code (if bit 5 is set)
	if (fspec & 0x10) != 0 {
		if buf.Len() < 2 {
			return bytesRead, fmt.Errorf("buffer too short for last measured Mode C code")
		}

		data := make([]byte, 2)
		n, err := buf.Read(data)
		if err != nil || n != 2 {
			return bytesRead + n, fmt.Errorf("reading last measured Mode C code: %w", err)
		}
		bytesRead += n

		// Extract validity and garbled flags
		m.LastModeCValidated = (data[0] & 0x80) == 0 // V bit (inverted: 0 = validated)
		m.LastModeCGarbled = (data[0] & 0x40) != 0   // G bit

		// Extract Mode C code as two's complement
		var modeCValue int16
		modeCBits := uint16(data[0]&0x3F)<<8 | uint16(data[1])
		if (modeCBits & 0x2000) != 0 {
			// Negative value (two's complement)
			modeCValue = -int16(^modeCBits&0x3FFF + 1)
		} else {
			// Positive value
			modeCValue = int16(modeCBits)
		}

		// LSB = 1/4 FL
		modeC := float64(modeCValue) * 0.25
		m.LastModeC = &modeC
	}

	// Subfield #5: Last Measured Mode 3/A code (if bit 4 is set)
	if (fspec & 0x08) != 0 {
		if buf.Len() < 2 {
			return bytesRead, fmt.Errorf("buffer too short for last measured Mode 3/A code")
		}

		data := make([]byte, 2)
		n, err := buf.Read(data)
		if err != nil || n != 2 {
			return bytesRead + n, fmt.Errorf("reading last measured Mode 3/A code: %w", err)
		}
		bytesRead += n

		// Extract validity, garbled, and smoothed flags
		m.LastMode3AValidated = (data[0] & 0x80) == 0 // V bit (inverted: 0 = validated)
		m.LastMode3AGarbled = (data[0] & 0x40) != 0   // G bit
		m.LastMode3ASmoothed = (data[0] & 0x20) != 0  // L bit

		// Extract Mode 3/A code (12 bits)
		// The 12 bits represent 4 octal digits (A, B, C, D), each using 3 bits
		mode3A := uint16(data[0]&0x0F)<<8 | uint16(data[1])
		m.LastMode3A = &mode3A
	}

	// Subfield #6: Report Type (if bit 3 is set)
	if (fspec & 0x04) != 0 {
		if buf.Len() < 1 {
			return bytesRead, fmt.Errorf("buffer too short for report type")
		}

		data, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading report type: %w", err)
		}
		bytesRead++

		// Extract report type and flags
		reportType := (data & 0xE0) >> 5 // Bits 8-6
		m.ReportType = &reportType
		m.SimulatedTarget = (data & 0x10) != 0 // Bit 5
		m.ReportFromRamp = (data & 0x08) != 0  // Bit 4
		m.TestTarget = (data & 0x04) != 0      // Bit 3
		// Bits 2-1 are spare
	}

	return bytesRead, nil
}

// Encode serializes the measured information into the buffer
func (m *MeasuredInformation) Encode(buf *bytes.Buffer) (int, error) {
	bytesWritten := 0

	// Determine which subfields are present
	hasSensor := m.SensorSAC != nil && m.SensorSIC != nil
	hasPosition := m.MeasuredRange != nil && m.MeasuredAzimuth != nil
	hasHeight := m.Measured3DHeight != nil
	hasModeC := m.LastModeC != nil
	hasMode3A := m.LastMode3A != nil
	hasReportType := m.ReportType != nil

	// Build FSPEC
	fspec := byte(0)
	if hasSensor {
		fspec |= 0x80 // Bit 8: Sensor Identification
	}
	if hasPosition {
		fspec |= 0x40 // Bit 7: Measured Position
	}
	if hasHeight {
		fspec |= 0x20 // Bit 6: Measured 3-D Height
	}
	if hasModeC {
		fspec |= 0x10 // Bit 5: Last Measured Mode C code
	}
	if hasMode3A {
		fspec |= 0x08 // Bit 4: Last Measured Mode 3/A code
	}
	if hasReportType {
		fspec |= 0x04 // Bit 3: Report Type
	}
	// Bit 2 is spare
	// Bit 1 (FX) is not set - no extension

	// Write FSPEC
	err := buf.WriteByte(fspec)
	if err != nil {
		return 0, fmt.Errorf("writing measured information FSPEC: %w", err)
	}
	bytesWritten++

	// Write Subfield #1: Sensor Identification
	if hasSensor {
		data := []byte{*m.SensorSAC, *m.SensorSIC}
		n, err := buf.Write(data)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing sensor identification: %w", err)
		}
		bytesWritten += n
	}

	// Write Subfield #2: Measured Position
	if hasPosition {
		// Convert range to binary (1/256 NM resolution)
		rangeBits := uint16(*m.MeasuredRange * 256.0)
		if *m.MeasuredRange >= 256.0 {
			rangeBits = 0xFFFF // Maximum value (256 NM)
		}

		// Convert azimuth to binary (360/2^16 degrees resolution)
		azimuthBits := uint16(*m.MeasuredAzimuth*65536.0/360.0) & 0xFFFF

		data := []byte{
			byte(rangeBits >> 8),
			byte(rangeBits),
			byte(azimuthBits >> 8),
			byte(azimuthBits),
		}
		n, err := buf.Write(data)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing measured position: %w", err)
		}
		bytesWritten += n
	}

	// Write Subfield #3: Measured 3-D Height
	if hasHeight {
		// Convert height to binary (25 feet resolution)
		heightBits := uint16(*m.Measured3DHeight / 25.0)

		data := []byte{
			byte(heightBits >> 8),
			byte(heightBits),
		}
		n, err := buf.Write(data)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing measured 3-D height: %w", err)
		}
		bytesWritten += n
	}

	// Write Subfield #4: Last Measured Mode C code
	if hasModeC {
		// Convert to 1/4 FL resolution
		modeCValue := int16(*m.LastModeC * 4.0)

		// Set first byte with flags and high bits
		firstByte := byte(0)
		if !m.LastModeCValidated {
			firstByte |= 0x80 // V bit (1 = not validated)
		}
		if m.LastModeCGarbled {
			firstByte |= 0x40 // G bit
		}

		// Handle two's complement for negative values
		var modeCBits uint16
		if modeCValue < 0 {
			// Two's complement for negative values
			modeCBits = uint16(^(-modeCValue) + 1)
			// Clear the sign bit and set the two's complement sign bit
			modeCBits = (modeCBits & 0x1FFF) | 0x2000
		} else {
			modeCBits = uint16(modeCValue)
		}

		// Add the high 6 bits to the first byte
		firstByte |= byte((modeCBits >> 8) & 0x3F)

		data := []byte{
			firstByte,
			byte(modeCBits),
		}
		n, err := buf.Write(data)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing last measured Mode C code: %w", err)
		}
		bytesWritten += n
	}

	// Write Subfield #5: Last Measured Mode 3/A code
	if hasMode3A {
		// Set first byte with flags and high bits
		firstByte := byte(0)
		if !m.LastMode3AValidated {
			firstByte |= 0x80 // V bit (1 = not validated)
		}
		if m.LastMode3AGarbled {
			firstByte |= 0x40 // G bit
		}
		if m.LastMode3ASmoothed {
			firstByte |= 0x20 // L bit
		}

		// Add the high 4 bits of the Mode 3/A code
		firstByte |= byte((*m.LastMode3A >> 8) & 0x0F)

		data := []byte{
			firstByte,
			byte(*m.LastMode3A),
		}
		n, err := buf.Write(data)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing last measured Mode 3/A code: %w", err)
		}
		bytesWritten += n
	}

	// Write Subfield #6: Report Type
	if hasReportType {
		reportTypeByte := byte((*m.ReportType & 0x07) << 5) // Bits 8-6: Report Type

		if m.SimulatedTarget {
			reportTypeByte |= 0x10 // Bit 5: SIM
		}
		if m.ReportFromRamp {
			reportTypeByte |= 0x08 // Bit 4: RAB
		}
		if m.TestTarget {
			reportTypeByte |= 0x04 // Bit 3: TST
		}
		// Bits 2-1 are spare (set to 0)

		err := buf.WriteByte(reportTypeByte)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing report type: %w", err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

// String returns a human-readable representation of the measured information
func (m *MeasuredInformation) String() string {
	parts := []string{}

	if m.SensorSAC != nil && m.SensorSIC != nil {
		parts = append(parts, fmt.Sprintf("Sensor: %d/%d", *m.SensorSAC, *m.SensorSIC))
	}

	if m.MeasuredRange != nil && m.MeasuredAzimuth != nil {
		parts = append(parts, fmt.Sprintf("Pos: %.2f NM / %.2f°", *m.MeasuredRange, *m.MeasuredAzimuth))
	}

	if m.Measured3DHeight != nil {
		parts = append(parts, fmt.Sprintf("Height: %.0f ft", *m.Measured3DHeight))
	}

	if m.LastModeC != nil {
		validStr := ""
		if !m.LastModeCValidated {
			validStr = "[not validated]"
		}
		if m.LastModeCGarbled {
			validStr += "[garbled]"
		}
		parts = append(parts, fmt.Sprintf("Mode C: FL %.2f %s", *m.LastModeC, validStr))
	}

	if m.LastMode3A != nil {
		// Convert 12-bit Mode 3/A code to octal representation
		a := (*m.LastMode3A >> 9) & 0x7
		b := (*m.LastMode3A >> 6) & 0x7
		c := (*m.LastMode3A >> 3) & 0x7
		d := *m.LastMode3A & 0x7

		validStr := ""
		if !m.LastMode3AValidated {
			validStr = "[not validated]"
		}
		if m.LastMode3AGarbled {
			validStr += "[garbled]"
		}
		if m.LastMode3ASmoothed {
			validStr += "[smoothed]"
		}

		parts = append(parts, fmt.Sprintf("Mode 3/A: %o%o%o%o %s", a, b, c, d, validStr))
	}

	if m.ReportType != nil {
		// Map report type values to strings
		reportTypes := []string{
			"No detection",
			"Single PSR",
			"Single SSR",
			"SSR+PSR",
			"Mode S All-Call",
			"Mode S Roll-Call",
			"Mode S All-Call+PSR",
			"Mode S Roll-Call+PSR",
		}

		typeStr := "Unknown"
		if int(*m.ReportType) < len(reportTypes) {
			typeStr = reportTypes[*m.ReportType]
		}

		flags := ""
		if m.SimulatedTarget {
			flags += "[simulated]"
		}
		if m.ReportFromRamp {
			flags += "[field monitor]"
		}
		if m.TestTarget {
			flags += "[test]"
		}

		parts = append(parts, fmt.Sprintf("Type: %s %s", typeStr, flags))
	}

	if len(parts) == 0 {
		return "MeasuredInformation[empty]"
	}

	return fmt.Sprintf("MeasuredInformation[%s]", strings.Join(parts, ", "))
}

// Validate performs validation on the measured information
func (m *MeasuredInformation) Validate() error {
	// Check range
	if m.MeasuredRange != nil && (*m.MeasuredRange < 0 || *m.MeasuredRange > 256) {
		return fmt.Errorf("measured range out of range [0,256]: %.2f NM", *m.MeasuredRange)
	}

	// Check azimuth
	if m.MeasuredAzimuth != nil && (*m.MeasuredAzimuth < 0 || *m.MeasuredAzimuth >= 360) {
		return fmt.Errorf("measured azimuth out of range [0,360): %.2f°", *m.MeasuredAzimuth)
	}

	// Check Mode C code
	if m.LastModeC != nil && (*m.LastModeC < -12 || *m.LastModeC > 1270) {
		return fmt.Errorf("Mode C flight level out of range [-12,1270]: %.2f", *m.LastModeC)
	}

	// Check Mode 3/A code
	if m.LastMode3A != nil && *m.LastMode3A > 0x0FFF {
		return fmt.Errorf("Mode 3/A code exceeds 12-bit limit: %04X", *m.LastMode3A)
	}

	// Check report type
	if m.ReportType != nil && *m.ReportType > 7 {
		return fmt.Errorf("report type out of range [0,7]: %d", *m.ReportType)
	}

	return nil
}

// formatMode3A formats a Mode 3/A code as an octal string (4 digits)
func formatMode3A(code uint16) string {
	// Extract each octal digit
	a := (code >> 9) & 0x7
	b := (code >> 6) & 0x7
	c := (code >> 3) & 0x7
	d := code & 0x7

	return fmt.Sprintf("%o%o%o%o", a, b, c, d)
}
