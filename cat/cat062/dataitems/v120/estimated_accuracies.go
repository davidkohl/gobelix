// dataitems/cat062/estimated_accuracies.go
package v120

import (
	"bytes"
	"fmt"
	"strings"
)

// EstimatedAccuracies implements I062/500
// Overview of all important accuracies
//
// Subfields (8 total, 7 in first extension, 1 in second):
// #1 (bit 16): APC - Estimated Accuracy Of Track Position (Cartesian) (4 octets)
// #2 (bit 15): COV - XY Covariance Component (2 octets)
// #3 (bit 14): APW - Estimated Accuracy Of Track Position (WGS-84) (4 octets)
// #4 (bit 13): AGA - Estimated Accuracy Of Geometric Altitude (1 octet)
// #5 (bit 12): ABA - Estimated Accuracy Of Barometric Altitude (1 octet)
// #6 (bit 11): ATV - Estimated Accuracy Of Track Velocity (Cartesian) (2 octets)
// #7 (bit 10): AA - Estimated Accuracy Of Acceleration (Cartesian) (2 octets)
// #8 (bit 8): ARC - Estimated Accuracy Of Rate Of Climb/Descent (1 octet)
type EstimatedAccuracies struct {
	// Subfield #1: Estimated Accuracy Of Track Position (Cartesian)
	// APC-X and APC-Y in meters, LSB = 0.5m
	APCX *uint16 // Standard Deviation of X component
	APCY *uint16 // Standard Deviation of Y component

	// Subfield #2: XY Covariance Component
	// In meters squared, LSB = 0.5m
	XYCovariance *int16 // Covariance of X and Y components

	// Subfield #3: Estimated Accuracy Of Track Position (WGS-84)
	// Latitude and Longitude in WGS-84, LSB = 180°/2^25 (≈ 0.0000054°)
	APWLatitude  *uint16 // Standard Deviation of Latitude
	APWLongitude *uint16 // Standard Deviation of Longitude

	// Subfield #4: Estimated Accuracy Of Calculated Track Geometric Altitude
	// In feet, LSB = 6.25 feet
	GeometricAltitudeAccuracy *uint8

	// Subfield #5: Estimated Accuracy Of Calculated Track Barometric Altitude
	// In feet, LSB = 6.25 feet
	BarometricAltitudeAccuracy *uint8

	// Subfield #6: Estimated Accuracy Of Track Velocity (Cartesian)
	// Vx and Vy in meters/second, LSB = 0.25 m/s
	VelocityX *uint8 // Standard Deviation of Vx
	VelocityY *uint8 // Standard Deviation of Vy

	// Subfield #7: Estimated Accuracy Of Acceleration (Cartesian)
	// Ax and Ay in meters/second^2, LSB = 0.25 m/s^2
	AccelerationX *uint8 // Standard Deviation of Ax
	AccelerationY *uint8 // Standard Deviation of Ay

	// Subfield #8: Estimated Accuracy Of Rate Of Climb/Descent
	// In feet/minute, LSB = 6.25 ft/min
	RateOfClimbDescentAccuracy *uint8
}

// Decode decodes the Estimated Accuracies compound data item from the buffer
func (e *EstimatedAccuracies) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Read primary subfield (with possible extension)
	primaryBytes := make([]byte, 0, 2)

	// Read first byte
	if buf.Len() < 1 {
		return 0, fmt.Errorf("buffer too short for estimated accuracies primary subfield")
	}

	firstByte, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading estimated accuracies primary subfield: %w", err)
	}
	primaryBytes = append(primaryBytes, firstByte)
	bytesRead++

	// Check FX bit (bit 1) for extension
	if (firstByte & 0x01) != 0 {
		// Read extension byte
		if buf.Len() < 1 {
			return bytesRead, fmt.Errorf("buffer too short for estimated accuracies extension")
		}

		secondByte, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading estimated accuracies extension: %w", err)
		}
		primaryBytes = append(primaryBytes, secondByte)
		bytesRead++
	}

	// Process subfields from first primary byte
	subfieldIndex := 0

	// Bit 8 (bit 16): APC - Estimated Accuracy Of Track Position (Cartesian)
	if (primaryBytes[0] & 0x80) != 0 {
		n, err := e.decodeAPC(buf)
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("decoding APC: %w", err)
		}
	}
	subfieldIndex++

	// Bit 7 (bit 15): COV - XY Covariance Component
	if (primaryBytes[0] & 0x40) != 0 {
		n, err := e.decodeCOV(buf)
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("decoding COV: %w", err)
		}
	}
	subfieldIndex++

	// Bit 6 (bit 14): APW - Estimated Accuracy Of Track Position (WGS-84)
	if (primaryBytes[0] & 0x20) != 0 {
		n, err := e.decodeAPW(buf)
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("decoding APW: %w", err)
		}
	}
	subfieldIndex++

	// Bit 5 (bit 13): AGA - Estimated Accuracy Of Geometric Altitude
	if (primaryBytes[0] & 0x10) != 0 {
		n, err := e.decodeAGA(buf)
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("decoding AGA: %w", err)
		}
	}
	subfieldIndex++

	// Bit 4 (bit 12): ABA - Estimated Accuracy Of Barometric Altitude
	if (primaryBytes[0] & 0x08) != 0 {
		n, err := e.decodeABA(buf)
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("decoding ABA: %w", err)
		}
	}
	subfieldIndex++

	// Bit 3 (bit 11): ATV - Estimated Accuracy Of Track Velocity
	if (primaryBytes[0] & 0x04) != 0 {
		n, err := e.decodeATV(buf)
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("decoding ATV: %w", err)
		}
	}
	subfieldIndex++

	// Bit 2 (bit 10): AA - Estimated Accuracy Of Acceleration
	if (primaryBytes[0] & 0x02) != 0 {
		n, err := e.decodeAA(buf)
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("decoding AA: %w", err)
		}
	}
	subfieldIndex++

	// Process subfields from second primary byte (if present)
	if len(primaryBytes) > 1 {
		// Bit 8 (bit 8 of second byte): ARC - Estimated Accuracy Of Rate Of Climb/Descent
		if (primaryBytes[1] & 0x80) != 0 {
			n, err := e.decodeARC(buf)
			bytesRead += n
			if err != nil {
				return bytesRead, fmt.Errorf("decoding ARC: %w", err)
			}
		}
		subfieldIndex++

		// Bit 7-2: Spare
		// Bit 1: FX (should be 0, no further extension defined)
		if (primaryBytes[1] & 0x01) != 0 {
			return bytesRead, fmt.Errorf("unexpected second extension bit in estimated accuracies")
		}
	}

	return bytesRead, nil
}

// decodeAPC decodes Subfield #1: Estimated Accuracy Of Track Position (Cartesian) (4 octets)
// APC-X (2 octets) and APC-Y (2 octets), LSB = 0.5m
func (e *EstimatedAccuracies) decodeAPC(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 4 {
		return 0, fmt.Errorf("buffer too short for APC")
	}

	data := make([]byte, 4)
	n, err := buf.Read(data)
	if err != nil {
		return n, err
	}

	// APC-X: unsigned 16-bit, LSB = 0.5m
	apcX := uint16(data[0])<<8 | uint16(data[1])
	e.APCX = &apcX

	// APC-Y: unsigned 16-bit, LSB = 0.5m
	apcY := uint16(data[2])<<8 | uint16(data[3])
	e.APCY = &apcY

	return 4, nil
}

// decodeCOV decodes Subfield #2: XY Covariance Component (2 octets)
// Signed 16-bit, LSB = 0.5m
func (e *EstimatedAccuracies) decodeCOV(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 2 {
		return 0, fmt.Errorf("buffer too short for COV")
	}

	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil {
		return n, err
	}

	// Signed 16-bit, LSB = 0.5m
	cov := int16(data[0])<<8 | int16(data[1])
	e.XYCovariance = &cov

	return 2, nil
}

// decodeAPW decodes Subfield #3: Estimated Accuracy Of Track Position (WGS-84) (4 octets)
// APW-LAT (2 octets) and APW-LON (2 octets), LSB = 180°/2^25
func (e *EstimatedAccuracies) decodeAPW(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 4 {
		return 0, fmt.Errorf("buffer too short for APW")
	}

	data := make([]byte, 4)
	n, err := buf.Read(data)
	if err != nil {
		return n, err
	}

	// APW-LAT: unsigned 16-bit, LSB = 180°/2^25
	apwLat := uint16(data[0])<<8 | uint16(data[1])
	e.APWLatitude = &apwLat

	// APW-LON: unsigned 16-bit, LSB = 180°/2^25
	apwLon := uint16(data[2])<<8 | uint16(data[3])
	e.APWLongitude = &apwLon

	return 4, nil
}

// decodeAGA decodes Subfield #4: Estimated Accuracy Of Geometric Altitude (1 octet)
// Unsigned 8-bit, LSB = 6.25 feet
func (e *EstimatedAccuracies) decodeAGA(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 1 {
		return 0, fmt.Errorf("buffer too short for AGA")
	}

	data, err := buf.ReadByte()
	if err != nil {
		return 0, err
	}

	e.GeometricAltitudeAccuracy = &data

	return 1, nil
}

// decodeABA decodes Subfield #5: Estimated Accuracy Of Barometric Altitude (1 octet)
// Unsigned 8-bit, LSB = 6.25 feet
func (e *EstimatedAccuracies) decodeABA(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 1 {
		return 0, fmt.Errorf("buffer too short for ABA")
	}

	data, err := buf.ReadByte()
	if err != nil {
		return 0, err
	}

	e.BarometricAltitudeAccuracy = &data

	return 1, nil
}

// decodeATV decodes Subfield #6: Estimated Accuracy Of Track Velocity (2 octets)
// Vx (1 octet) and Vy (1 octet), LSB = 0.25 m/s
func (e *EstimatedAccuracies) decodeATV(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 2 {
		return 0, fmt.Errorf("buffer too short for ATV")
	}

	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil {
		return n, err
	}

	// Vx: unsigned 8-bit, LSB = 0.25 m/s
	vx := data[0]
	e.VelocityX = &vx

	// Vy: unsigned 8-bit, LSB = 0.25 m/s
	vy := data[1]
	e.VelocityY = &vy

	return 2, nil
}

// decodeAA decodes Subfield #7: Estimated Accuracy Of Acceleration (2 octets)
// Ax (1 octet) and Ay (1 octet), LSB = 0.25 m/s^2
func (e *EstimatedAccuracies) decodeAA(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 2 {
		return 0, fmt.Errorf("buffer too short for AA")
	}

	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil {
		return n, err
	}

	// Ax: unsigned 8-bit, LSB = 0.25 m/s^2
	ax := data[0]
	e.AccelerationX = &ax

	// Ay: unsigned 8-bit, LSB = 0.25 m/s^2
	ay := data[1]
	e.AccelerationY = &ay

	return 2, nil
}

// decodeARC decodes Subfield #8: Estimated Accuracy Of Rate Of Climb/Descent (1 octet)
// Unsigned 8-bit, LSB = 6.25 ft/min
func (e *EstimatedAccuracies) decodeARC(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 1 {
		return 0, fmt.Errorf("buffer too short for ARC")
	}

	data, err := buf.ReadByte()
	if err != nil {
		return 0, err
	}

	e.RateOfClimbDescentAccuracy = &data

	return 1, nil
}

// Encode encodes the Estimated Accuracies compound data item to the buffer
func (e *EstimatedAccuracies) Encode(buf *bytes.Buffer) (int, error) {
	bytesWritten := 0

	// Build primary subfield
	var primaryByte1 byte = 0x00
	var primaryByte2 byte = 0x00
	needsExtension := false

	if e.APCX != nil && e.APCY != nil {
		primaryByte1 |= 0x80 // Bit 8: APC
	}
	if e.XYCovariance != nil {
		primaryByte1 |= 0x40 // Bit 7: COV
	}
	if e.APWLatitude != nil && e.APWLongitude != nil {
		primaryByte1 |= 0x20 // Bit 6: APW
	}
	if e.GeometricAltitudeAccuracy != nil {
		primaryByte1 |= 0x10 // Bit 5: AGA
	}
	if e.BarometricAltitudeAccuracy != nil {
		primaryByte1 |= 0x08 // Bit 4: ABA
	}
	if e.VelocityX != nil && e.VelocityY != nil {
		primaryByte1 |= 0x04 // Bit 3: ATV
	}
	if e.AccelerationX != nil && e.AccelerationY != nil {
		primaryByte1 |= 0x02 // Bit 2: AA
	}

	// Check if we need second byte
	if e.RateOfClimbDescentAccuracy != nil {
		needsExtension = true
		primaryByte2 |= 0x80 // Bit 8: ARC
	}

	// Set FX bit if extension needed
	if needsExtension {
		primaryByte1 |= 0x01 // Bit 1: FX
	}

	// Write primary subfield
	err := buf.WriteByte(primaryByte1)
	if err != nil {
		return 0, err
	}
	bytesWritten++

	if needsExtension {
		err := buf.WriteByte(primaryByte2)
		if err != nil {
			return bytesWritten, err
		}
		bytesWritten++
	}

	// Encode subfields
	if (primaryByte1 & 0x80) != 0 {
		n, err := e.encodeAPC(buf)
		bytesWritten += n
		if err != nil {
			return bytesWritten, err
		}
	}

	if (primaryByte1 & 0x40) != 0 {
		n, err := e.encodeCOV(buf)
		bytesWritten += n
		if err != nil {
			return bytesWritten, err
		}
	}

	if (primaryByte1 & 0x20) != 0 {
		n, err := e.encodeAPW(buf)
		bytesWritten += n
		if err != nil {
			return bytesWritten, err
		}
	}

	if (primaryByte1 & 0x10) != 0 {
		n, err := e.encodeAGA(buf)
		bytesWritten += n
		if err != nil {
			return bytesWritten, err
		}
	}

	if (primaryByte1 & 0x08) != 0 {
		n, err := e.encodeABA(buf)
		bytesWritten += n
		if err != nil {
			return bytesWritten, err
		}
	}

	if (primaryByte1 & 0x04) != 0 {
		n, err := e.encodeATV(buf)
		bytesWritten += n
		if err != nil {
			return bytesWritten, err
		}
	}

	if (primaryByte1 & 0x02) != 0 {
		n, err := e.encodeAA(buf)
		bytesWritten += n
		if err != nil {
			return bytesWritten, err
		}
	}

	if needsExtension && (primaryByte2&0x80) != 0 {
		n, err := e.encodeARC(buf)
		bytesWritten += n
		if err != nil {
			return bytesWritten, err
		}
	}

	return bytesWritten, nil
}

// encodeAPC encodes Subfield #1
func (e *EstimatedAccuracies) encodeAPC(buf *bytes.Buffer) (int, error) {
	data := []byte{
		byte(*e.APCX >> 8),
		byte(*e.APCX),
		byte(*e.APCY >> 8),
		byte(*e.APCY),
	}
	return buf.Write(data)
}

// encodeCOV encodes Subfield #2
func (e *EstimatedAccuracies) encodeCOV(buf *bytes.Buffer) (int, error) {
	data := []byte{
		byte(*e.XYCovariance >> 8),
		byte(*e.XYCovariance),
	}
	return buf.Write(data)
}

// encodeAPW encodes Subfield #3
func (e *EstimatedAccuracies) encodeAPW(buf *bytes.Buffer) (int, error) {
	data := []byte{
		byte(*e.APWLatitude >> 8),
		byte(*e.APWLatitude),
		byte(*e.APWLongitude >> 8),
		byte(*e.APWLongitude),
	}
	return buf.Write(data)
}

// encodeAGA encodes Subfield #4
func (e *EstimatedAccuracies) encodeAGA(buf *bytes.Buffer) (int, error) {
	return buf.Write([]byte{*e.GeometricAltitudeAccuracy})
}

// encodeABA encodes Subfield #5
func (e *EstimatedAccuracies) encodeABA(buf *bytes.Buffer) (int, error) {
	return buf.Write([]byte{*e.BarometricAltitudeAccuracy})
}

// encodeATV encodes Subfield #6
func (e *EstimatedAccuracies) encodeATV(buf *bytes.Buffer) (int, error) {
	data := []byte{
		*e.VelocityX,
		*e.VelocityY,
	}
	return buf.Write(data)
}

// encodeAA encodes Subfield #7
func (e *EstimatedAccuracies) encodeAA(buf *bytes.Buffer) (int, error) {
	data := []byte{
		*e.AccelerationX,
		*e.AccelerationY,
	}
	return buf.Write(data)
}

// encodeARC encodes Subfield #8
func (e *EstimatedAccuracies) encodeARC(buf *bytes.Buffer) (int, error) {
	return buf.Write([]byte{*e.RateOfClimbDescentAccuracy})
}

// String returns a human-readable representation
func (e *EstimatedAccuracies) String() string {
	var parts []string

	if e.APCX != nil && e.APCY != nil {
		apcxMeters := float64(*e.APCX) * 0.5
		apcyMeters := float64(*e.APCY) * 0.5
		parts = append(parts, fmt.Sprintf("PosCart=(%.1f,%.1f)m", apcxMeters, apcyMeters))
	}

	if e.XYCovariance != nil {
		covMeters := float64(*e.XYCovariance) * 0.5
		parts = append(parts, fmt.Sprintf("Cov=%.1fm", covMeters))
	}

	if e.APWLatitude != nil && e.APWLongitude != nil {
		latDeg := float64(*e.APWLatitude) * 180.0 / 33554432.0 // 2^25
		lonDeg := float64(*e.APWLongitude) * 180.0 / 33554432.0
		parts = append(parts, fmt.Sprintf("PosWGS84=(%.6f°,%.6f°)", latDeg, lonDeg))
	}

	if e.GeometricAltitudeAccuracy != nil {
		geoAltFeet := float64(*e.GeometricAltitudeAccuracy) * 6.25
		parts = append(parts, fmt.Sprintf("GeoAlt=%.1f ft", geoAltFeet))
	}

	if e.BarometricAltitudeAccuracy != nil {
		baroAltFeet := float64(*e.BarometricAltitudeAccuracy) * 6.25
		parts = append(parts, fmt.Sprintf("BaroAlt=%.1f ft", baroAltFeet))
	}

	if e.VelocityX != nil && e.VelocityY != nil {
		vxMs := float64(*e.VelocityX) * 0.25
		vyMs := float64(*e.VelocityY) * 0.25
		parts = append(parts, fmt.Sprintf("Vel=(%.2f,%.2f)m/s", vxMs, vyMs))
	}

	if e.AccelerationX != nil && e.AccelerationY != nil {
		axMs2 := float64(*e.AccelerationX) * 0.25
		ayMs2 := float64(*e.AccelerationY) * 0.25
		parts = append(parts, fmt.Sprintf("Acc=(%.2f,%.2f)m/s²", axMs2, ayMs2))
	}

	if e.RateOfClimbDescentAccuracy != nil {
		rocFtMin := float64(*e.RateOfClimbDescentAccuracy) * 6.25
		parts = append(parts, fmt.Sprintf("RoC=%.1f ft/min", rocFtMin))
	}

	if len(parts) == 0 {
		return "EstimatedAccuracies{empty}"
	}

	return fmt.Sprintf("EstimatedAccuracies{%s}", strings.Join(parts, ", "))
}

// Validate performs validation checks on the data
func (e *EstimatedAccuracies) Validate() error {
	// All fields are unsigned or have no special constraints
	// No validation needed for this data item
	return nil
}
