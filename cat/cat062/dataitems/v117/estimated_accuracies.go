// dataitems/cat062/estimated_accuracies.go
package v117

import (
	"bytes"
	"fmt"
	"strings"
)

// EstimatedAccuracies implements I062/500
// Contains the estimated accuracy for various parameters of the track
type EstimatedAccuracies struct {
	// Subfield #1: Estimated Accuracy Of Track Position (Cartesian)
	// Standard deviation in meters
	PositionAccuracyX *float64 // X component accuracy
	PositionAccuracyY *float64 // Y component accuracy

	// Subfield #2: XY Covariance Component
	// XY covariance component in two's complement form
	Covariance *float64

	// Subfield #3: Estimated Accuracy Of Track Position (WGS-84)
	// Standard deviation in degrees
	PositionAccuracyLat *float64 // Latitude component accuracy
	PositionAccuracyLon *float64 // Longitude component accuracy

	// Subfield #4: Estimated Accuracy Of Calculated Track Geometric Altitude
	// Standard deviation in feet
	GeometricAltitudeAccuracy *float64

	// Subfield #5: Estimated Accuracy Of Calculated Track Barometric Altitude
	// Standard deviation in flight levels
	BarometricAltitudeAccuracy *float64

	// Subfield #6: Estimated Accuracy Of Track Velocity (Cartesian)
	// Standard deviation in meters per second
	VelocityAccuracyX *float64 // X component accuracy
	VelocityAccuracyY *float64 // Y component accuracy

	// Subfield #7: Estimated Accuracy Of Acceleration (Cartesian)
	// Standard deviation in meters per second squared
	AccelerationAccuracyX *float64 // X component accuracy
	AccelerationAccuracyY *float64 // Y component accuracy

	// Subfield #8: Estimated Accuracy Of Rate Of Climb/Descent
	// Standard deviation in feet per minute
	RateOfClimbAccuracy *float64
}

// Decode parses an ASTERIX Category 062 I500 data item from the buffer
func (e *EstimatedAccuracies) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Read the primary subfield (FSPEC)
	if buf.Len() < 1 {
		return 0, fmt.Errorf("buffer too short for estimated accuracies FSPEC")
	}

	fspec1, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading estimated accuracies FSPEC: %w", err)
	}
	bytesRead++

	// Check for FX extension bit
	hasSecondFSPEC := (fspec1 & 0x01) != 0
	var fspec2 byte
	if hasSecondFSPEC {
		if buf.Len() < 1 {
			return bytesRead, fmt.Errorf("buffer too short for second FSPEC byte")
		}
		fspec2, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading second FSPEC byte: %w", err)
		}
		bytesRead++

		// Check that the second FSPEC doesn't have extension bit set
		if (fspec2 & 0x01) != 0 {
			return bytesRead, fmt.Errorf("unexpected extension in second FSPEC byte")
		}
	}

	// Subfield #1: Estimated Accuracy Of Track Position (Cartesian)
	if (fspec1 & 0x80) != 0 {
		if buf.Len() < 4 {
			return bytesRead, fmt.Errorf("buffer too short for position accuracy")
		}

		data := make([]byte, 4)
		n, err := buf.Read(data)
		if err != nil || n != 4 {
			return bytesRead + n, fmt.Errorf("reading position accuracy: %w", err)
		}
		bytesRead += n

		// X component (first 2 bytes)
		xAccBits := uint16(data[0])<<8 | uint16(data[1])
		xAcc := float64(xAccBits) * 0.5 // LSB = 0.5m
		e.PositionAccuracyX = &xAcc

		// Y component (last 2 bytes)
		yAccBits := uint16(data[2])<<8 | uint16(data[3])
		yAcc := float64(yAccBits) * 0.5 // LSB = 0.5m
		e.PositionAccuracyY = &yAcc
	}

	// Subfield #2: XY Covariance
	if (fspec1 & 0x40) != 0 {
		if buf.Len() < 2 {
			return bytesRead, fmt.Errorf("buffer too short for XY covariance")
		}

		data := make([]byte, 2)
		n, err := buf.Read(data)
		if err != nil || n != 2 {
			return bytesRead + n, fmt.Errorf("reading XY covariance: %w", err)
		}
		bytesRead += n

		// Extract covariance as two's complement
		covBits := uint16(data[0])<<8 | uint16(data[1])
		var covValue int16
		if (covBits & 0x8000) != 0 {
			// Negative value (two's complement)
			covValue = -int16(^covBits + 1)
		} else {
			// Positive value
			covValue = int16(covBits)
		}

		cov := float64(covValue) * 0.5 // LSB = 0.5m
		e.Covariance = &cov
	}

	// Subfield #3: Estimated Accuracy Of Track Position (WGS-84)
	if (fspec1 & 0x20) != 0 {
		if buf.Len() < 4 {
			return bytesRead, fmt.Errorf("buffer too short for WGS-84 position accuracy")
		}

		data := make([]byte, 4)
		n, err := buf.Read(data)
		if err != nil || n != 4 {
			return bytesRead + n, fmt.Errorf("reading WGS-84 position accuracy: %w", err)
		}
		bytesRead += n

		// Latitude component (first 2 bytes)
		latAccBits := uint16(data[0])<<8 | uint16(data[1])
		latAcc := float64(latAccBits) * 180.0 / float64(1<<25) // LSB = 180/2^25 degrees
		e.PositionAccuracyLat = &latAcc

		// Longitude component (last 2 bytes)
		lonAccBits := uint16(data[2])<<8 | uint16(data[3])
		lonAcc := float64(lonAccBits) * 180.0 / float64(1<<25) // LSB = 180/2^25 degrees
		e.PositionAccuracyLon = &lonAcc
	}

	// Subfield #4: Estimated Accuracy Of Calculated Track Geometric Altitude
	if (fspec1 & 0x10) != 0 {
		if buf.Len() < 1 {
			return bytesRead, fmt.Errorf("buffer too short for geometric altitude accuracy")
		}

		data, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading geometric altitude accuracy: %w", err)
		}
		bytesRead++

		// Convert to feet
		altAcc := float64(data) * 6.25 // LSB = 6.25 feet
		e.GeometricAltitudeAccuracy = &altAcc
	}

	// Subfield #5: Estimated Accuracy Of Calculated Track Barometric Altitude
	if (fspec1 & 0x08) != 0 {
		if buf.Len() < 1 {
			return bytesRead, fmt.Errorf("buffer too short for barometric altitude accuracy")
		}

		data, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading barometric altitude accuracy: %w", err)
		}
		bytesRead++

		// Convert to flight levels
		altAcc := float64(data) * 0.25 // LSB = 1/4 FL
		e.BarometricAltitudeAccuracy = &altAcc
	}

	// Subfield #6: Estimated Accuracy Of Track Velocity (Cartesian)
	if (fspec1 & 0x04) != 0 {
		if buf.Len() < 2 {
			return bytesRead, fmt.Errorf("buffer too short for velocity accuracy")
		}

		data := make([]byte, 2)
		n, err := buf.Read(data)
		if err != nil || n != 2 {
			return bytesRead + n, fmt.Errorf("reading velocity accuracy: %w", err)
		}
		bytesRead += n

		// X component (first byte)
		xAcc := float64(data[0]) * 0.25 // LSB = 0.25m/s
		e.VelocityAccuracyX = &xAcc

		// Y component (second byte)
		yAcc := float64(data[1]) * 0.25 // LSB = 0.25m/s
		e.VelocityAccuracyY = &yAcc
	}

	// Subfield #7: Estimated Accuracy Of Acceleration (Cartesian)
	if (fspec1 & 0x02) != 0 {
		if buf.Len() < 2 {
			return bytesRead, fmt.Errorf("buffer too short for acceleration accuracy")
		}

		data := make([]byte, 2)
		n, err := buf.Read(data)
		if err != nil || n != 2 {
			return bytesRead + n, fmt.Errorf("reading acceleration accuracy: %w", err)
		}
		bytesRead += n

		// X component (first byte)
		xAcc := float64(data[0]) * 0.25 // LSB = 0.25m/s²
		e.AccelerationAccuracyX = &xAcc

		// Y component (second byte)
		yAcc := float64(data[1]) * 0.25 // LSB = 0.25m/s²
		e.AccelerationAccuracyY = &yAcc
	}

	// The second FSPEC byte contains one subfield
	if hasSecondFSPEC {
		// Subfield #8: Estimated Accuracy Of Rate Of Climb/Descent
		if (fspec2 & 0x80) != 0 {
			if buf.Len() < 1 {
				return bytesRead, fmt.Errorf("buffer too short for rate of climb accuracy")
			}

			data, err := buf.ReadByte()
			if err != nil {
				return bytesRead, fmt.Errorf("reading rate of climb accuracy: %w", err)
			}
			bytesRead++

			// Convert to feet per minute
			rocAcc := float64(data) * 6.25 // LSB = 6.25 feet/minute
			e.RateOfClimbAccuracy = &rocAcc
		}
	}

	return bytesRead, nil
}

// Encode serializes the estimated accuracies into the buffer
func (e *EstimatedAccuracies) Encode(buf *bytes.Buffer) (int, error) {
	bytesWritten := 0

	// Determine which subfields are present
	hasPosition := e.PositionAccuracyX != nil && e.PositionAccuracyY != nil
	hasCovariance := e.Covariance != nil
	hasWGS84Position := e.PositionAccuracyLat != nil && e.PositionAccuracyLon != nil
	hasGeoAltitude := e.GeometricAltitudeAccuracy != nil
	hasBaroAltitude := e.BarometricAltitudeAccuracy != nil
	hasVelocity := e.VelocityAccuracyX != nil && e.VelocityAccuracyY != nil
	hasAcceleration := e.AccelerationAccuracyX != nil && e.AccelerationAccuracyY != nil
	hasRateOfClimb := e.RateOfClimbAccuracy != nil

	// Need second FSPEC byte?
	needSecondByte := hasRateOfClimb

	// First FSPEC byte
	fspec1 := byte(0)
	if hasPosition {
		fspec1 |= 0x80 // Bit 8: Position Accuracy
	}
	if hasCovariance {
		fspec1 |= 0x40 // Bit 7: Covariance
	}
	if hasWGS84Position {
		fspec1 |= 0x20 // Bit 6: WGS84 Position Accuracy
	}
	if hasGeoAltitude {
		fspec1 |= 0x10 // Bit 5: Geometric Altitude Accuracy
	}
	if hasBaroAltitude {
		fspec1 |= 0x08 // Bit 4: Barometric Altitude Accuracy
	}
	if hasVelocity {
		fspec1 |= 0x04 // Bit 3: Velocity Accuracy
	}
	if hasAcceleration {
		fspec1 |= 0x02 // Bit 2: Acceleration Accuracy
	}
	if needSecondByte {
		fspec1 |= 0x01 // Bit 1: FX
	}

	// Write first FSPEC byte
	err := buf.WriteByte(fspec1)
	if err != nil {
		return 0, fmt.Errorf("writing first FSPEC byte: %w", err)
	}
	bytesWritten++

	// Second FSPEC byte if needed
	if needSecondByte {
		fspec2 := byte(0)
		if hasRateOfClimb {
			fspec2 |= 0x80 // Bit 8: Rate of Climb Accuracy
		}
		// Bits 7-2 are spare
		// Bit 1 (FX) = 0, no extension

		// Write second FSPEC byte
		err := buf.WriteByte(fspec2)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing second FSPEC byte: %w", err)
		}
		bytesWritten++
	}

	// Subfield #1: Estimated Accuracy Of Track Position (Cartesian)
	if hasPosition {
		// Convert to binary (0.5m resolution)
		xAccBits := uint16(*e.PositionAccuracyX / 0.5)
		yAccBits := uint16(*e.PositionAccuracyY / 0.5)

		data := []byte{
			byte(xAccBits >> 8),
			byte(xAccBits),
			byte(yAccBits >> 8),
			byte(yAccBits),
		}
		n, err := buf.Write(data)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing position accuracy: %w", err)
		}
		bytesWritten += n
	}

	// Subfield #2: XY Covariance
	if hasCovariance {
		// Convert to binary (0.5m resolution)
		var covBits uint16
		covValue := int16(*e.Covariance / 0.5)
		if covValue < 0 {
			// Two's complement for negative values
			covBits = uint16(^(-covValue) + 1)
		} else {
			covBits = uint16(covValue)
		}

		data := []byte{
			byte(covBits >> 8),
			byte(covBits),
		}
		n, err := buf.Write(data)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing covariance: %w", err)
		}
		bytesWritten += n
	}

	// Subfield #3: Estimated Accuracy Of Track Position (WGS-84)
	if hasWGS84Position {
		// Convert to binary (180/2^25 degrees resolution)
		latAccBits := uint16(*e.PositionAccuracyLat * float64(1<<25) / 180.0)
		lonAccBits := uint16(*e.PositionAccuracyLon * float64(1<<25) / 180.0)

		data := []byte{
			byte(latAccBits >> 8),
			byte(latAccBits),
			byte(lonAccBits >> 8),
			byte(lonAccBits),
		}
		n, err := buf.Write(data)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing WGS-84 position accuracy: %w", err)
		}
		bytesWritten += n
	}

	// Subfield #4: Estimated Accuracy Of Calculated Track Geometric Altitude
	if hasGeoAltitude {
		// Convert to binary (6.25 feet resolution)
		altAccBits := uint8(*e.GeometricAltitudeAccuracy / 6.25)

		err := buf.WriteByte(altAccBits)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing geometric altitude accuracy: %w", err)
		}
		bytesWritten++
	}

	// Subfield #5: Estimated Accuracy Of Calculated Track Barometric Altitude
	if hasBaroAltitude {
		// Convert to binary (1/4 FL resolution)
		altAccBits := uint8(*e.BarometricAltitudeAccuracy / 0.25)

		err := buf.WriteByte(altAccBits)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing barometric altitude accuracy: %w", err)
		}
		bytesWritten++
	}

	// Subfield #6: Estimated Accuracy Of Track Velocity (Cartesian)
	if hasVelocity {
		// Convert to binary (0.25 m/s resolution)
		xAccBits := uint8(*e.VelocityAccuracyX / 0.25)
		yAccBits := uint8(*e.VelocityAccuracyY / 0.25)

		data := []byte{xAccBits, yAccBits}
		n, err := buf.Write(data)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing velocity accuracy: %w", err)
		}
		bytesWritten += n
	}

	// Subfield #7: Estimated Accuracy Of Acceleration (Cartesian)
	if hasAcceleration {
		// Convert to binary (0.25 m/s² resolution)
		xAccBits := uint8(*e.AccelerationAccuracyX / 0.25)
		yAccBits := uint8(*e.AccelerationAccuracyY / 0.25)

		data := []byte{xAccBits, yAccBits}
		n, err := buf.Write(data)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing acceleration accuracy: %w", err)
		}
		bytesWritten += n
	}

	// Subfield #8: Estimated Accuracy Of Rate Of Climb/Descent
	if hasRateOfClimb {
		// Convert to binary (6.25 feet/minute resolution)
		rocAccBits := uint8(*e.RateOfClimbAccuracy / 6.25)

		err := buf.WriteByte(rocAccBits)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing rate of climb accuracy: %w", err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

// String returns a human-readable representation of the estimated accuracies
func (e *EstimatedAccuracies) String() string {
	parts := []string{}

	if e.PositionAccuracyX != nil && e.PositionAccuracyY != nil {
		parts = append(parts, fmt.Sprintf("PosXY: %.1fm/%.1fm", *e.PositionAccuracyX, *e.PositionAccuracyY))
	}

	if e.Covariance != nil {
		parts = append(parts, fmt.Sprintf("Cov: %.1fm", *e.Covariance))
	}

	if e.PositionAccuracyLat != nil && e.PositionAccuracyLon != nil {
		// Convert to more readable format (meters at the equator)
		// 1 degree latitude ≈ 111km, 1 degree longitude at equator ≈ 111km
		latMeter := *e.PositionAccuracyLat * 111000
		lonMeter := *e.PositionAccuracyLon * 111000
		parts = append(parts, fmt.Sprintf("PosLatLon: %.1fm/%.1fm", latMeter, lonMeter))
	}

	if e.GeometricAltitudeAccuracy != nil {
		parts = append(parts, fmt.Sprintf("GeoAlt: %.1fft", *e.GeometricAltitudeAccuracy))
	}

	if e.BarometricAltitudeAccuracy != nil {
		parts = append(parts, fmt.Sprintf("BaroAlt: FL%.2f", *e.BarometricAltitudeAccuracy))
	}

	if e.VelocityAccuracyX != nil && e.VelocityAccuracyY != nil {
		parts = append(parts, fmt.Sprintf("Vel: %.2fm/s", max(*e.VelocityAccuracyX, *e.VelocityAccuracyY)))
	}

	if e.AccelerationAccuracyX != nil && e.AccelerationAccuracyY != nil {
		parts = append(parts, fmt.Sprintf("Acc: %.2fm/s²", max(*e.AccelerationAccuracyX, *e.AccelerationAccuracyY)))
	}

	if e.RateOfClimbAccuracy != nil {
		parts = append(parts, fmt.Sprintf("ROC: %.1fft/min", *e.RateOfClimbAccuracy))
	}

	if len(parts) == 0 {
		return "EstimatedAccuracies[empty]"
	}

	return fmt.Sprintf("EstimatedAccuracies[%s]", strings.Join(parts, ", "))
}

// Validate performs validation on the estimated accuracies
func (e *EstimatedAccuracies) Validate() error {
	// All values should be positive
	if e.PositionAccuracyX != nil && *e.PositionAccuracyX < 0 {
		return fmt.Errorf("position accuracy X cannot be negative: %.2f", *e.PositionAccuracyX)
	}

	if e.PositionAccuracyY != nil && *e.PositionAccuracyY < 0 {
		return fmt.Errorf("position accuracy Y cannot be negative: %.2f", *e.PositionAccuracyY)
	}

	// Covariance can be positive or negative

	if e.PositionAccuracyLat != nil && *e.PositionAccuracyLat < 0 {
		return fmt.Errorf("position accuracy latitude cannot be negative: %.8f", *e.PositionAccuracyLat)
	}

	if e.PositionAccuracyLon != nil && *e.PositionAccuracyLon < 0 {
		return fmt.Errorf("position accuracy longitude cannot be negative: %.8f", *e.PositionAccuracyLon)
	}

	if e.GeometricAltitudeAccuracy != nil && *e.GeometricAltitudeAccuracy < 0 {
		return fmt.Errorf("geometric altitude accuracy cannot be negative: %.2f", *e.GeometricAltitudeAccuracy)
	}

	if e.BarometricAltitudeAccuracy != nil && *e.BarometricAltitudeAccuracy < 0 {
		return fmt.Errorf("barometric altitude accuracy cannot be negative: %.2f", *e.BarometricAltitudeAccuracy)
	}

	if e.VelocityAccuracyX != nil && *e.VelocityAccuracyX < 0 {
		return fmt.Errorf("velocity accuracy X cannot be negative: %.2f", *e.VelocityAccuracyX)
	}

	if e.VelocityAccuracyY != nil && *e.VelocityAccuracyY < 0 {
		return fmt.Errorf("velocity accuracy Y cannot be negative: %.2f", *e.VelocityAccuracyY)
	}

	if e.AccelerationAccuracyX != nil && *e.AccelerationAccuracyX < 0 {
		return fmt.Errorf("acceleration accuracy X cannot be negative: %.2f", *e.AccelerationAccuracyX)
	}

	if e.AccelerationAccuracyY != nil && *e.AccelerationAccuracyY < 0 {
		return fmt.Errorf("acceleration accuracy Y cannot be negative: %.2f", *e.AccelerationAccuracyY)
	}

	if e.RateOfClimbAccuracy != nil && *e.RateOfClimbAccuracy < 0 {
		return fmt.Errorf("rate of climb accuracy cannot be negative: %.2f", *e.RateOfClimbAccuracy)
	}

	return nil
}

// max returns the maximum of two float64 values
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
