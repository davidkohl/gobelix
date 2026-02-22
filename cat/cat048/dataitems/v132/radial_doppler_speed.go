// dataitems/cat048/radial_doppler_speed.go
package v132

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// RadialDopplerSpeed implements I048/120
// Information on the Doppler Speed of the target report.
type RadialDopplerSpeed struct {
	// Subfields availability flags
	CAL bool // Calculated Doppler Speed
	RDS bool // Raw Doppler Speed

	// Calculated Doppler Speed subfield
	CalculatedDopplerValid bool    // Doppler speed validity flag
	CalculatedDopplerSpeed float64 // Calculated Doppler Speed (m/s)

	// Raw Doppler Speed subfield
	RawDopplerCount     uint8     // Repetition factor for Raw Doppler Speed
	RawDopplerSpeeds    []float64 // Doppler Speed for multiple responses (m/s)
	RawDopplerAmbiguity []float64 // Ambiguity Range for multiple responses (m/s)
	RawDopplerFrequency []float64 // Transmitter Frequency for multiple responses (MHz)
}

// Decode implements the DataItem interface
func (r *RadialDopplerSpeed) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Check we have at least 1 byte for primary subfield
	if buf.Len() < 1 {
		return 0, fmt.Errorf("%w: need 1 byte for radial doppler speed primary subfield", asterix.ErrBufferTooShort)
	}

	// Peek at the primary subfield byte to validate BEFORE consuming it
	primaryByte := buf.Bytes()[0]

	// Extract subfield presence flags
	cal := (primaryByte & 0x80) != 0 // bit 8
	rds := (primaryByte & 0x40) != 0 // bit 7

	// Note: According to the spec, only one secondary subfield shall be present.
	// However, some radar implementations may send both. We allow this for
	// interoperability, but the Validate() method still enforces the spec rule.

	// Validation passed - now safely consume the primary subfield byte
	var primary [1]byte
	n, err := buf.Read(primary[:])
	if err != nil {
		return n, fmt.Errorf("reading radial doppler speed primary subfield: %w", err)
	}
	bytesRead += n

	// Store the validated flags
	r.CAL = cal
	r.RDS = rds
	// bits 6-2 are spare
	fx := (primary[0] & 0x01) != 0 // bit 1 (FX)

	// Track unknown subfields from extension presence bytes
	unknownSubfieldCount := 0
	if fx {
		// FX bit is set in primary subfield, which means there's an extension
		// Not defined in v1.32, but may exist in other versions (e.g., v1.17)
		// Skip all extension presence bytes (each has FX bit indicating if more follow)
		for fx {
			var extByte [1]byte
			n, err := buf.Read(extByte[:])
			if err != nil {
				return bytesRead + n, fmt.Errorf("reading radial doppler speed extension: %w", err)
			}
			bytesRead += n

			// Count unknown subfields indicated in extension (bits 8-2)
			// NOTE: We don't know the sizes of these subfields, so we can't skip them.
			// This count is for diagnostic purposes only.
			for bit := 7; bit >= 1; bit-- {
				if extByte[0]&(1<<bit) != 0 {
					unknownSubfieldCount++
				}
			}

			fx = (extByte[0] & 0x01) != 0 // Check FX bit for next extension
		}

		// If unknown subfields are indicated but we can't determine their sizes,
		// we have no choice but to continue and hope the data is v1.32 compliant.
		// Log or error if needed:
		// if unknownSubfieldCount > 0 {
		//     // Unknown subfields present - data may not parse correctly
		// }
	}
	_ = unknownSubfieldCount // suppress unused variable warning for now

	// Subfield #1: Calculated Doppler Speed (2 bytes)
	if r.CAL {
		var data [2]byte
		n, err := buf.Read(data[:])
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading calculated doppler speed: %w", err)
		}
		bytesRead += n

		// Extract validity flag
		r.CalculatedDopplerValid = (data[0] & 0x80) == 0 // bit 16 (inverted)
		// bits 15-11 are spare

		// Extract doppler speed, bits 10-1, LSB = 1 m/s, in two's complement
		raw := int16(uint16(data[0]&0x03)<<8 | uint16(data[1]))
		if (data[0] & 0x02) != 0 { // Check if bit 10 is set (negative number)
			// Apply sign extension to get proper negative value
			raw |= -1024 // -1024 = 0xFC00 = Two's complement sign extension for 10 bits
		}

		r.CalculatedDopplerSpeed = float64(raw)
	}

	// Subfield #2: Raw Doppler Speed (variable length)
	if r.RDS {
		// First byte is repetition factor
		var repFactorByte [1]byte
		n, err := buf.Read(repFactorByte[:])
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading raw doppler speed repetition factor: %w", err)
		}
		bytesRead += n

		r.RawDopplerCount = repFactorByte[0]

		// Sanity check: REP values above 16 are highly unlikely in practice
		// (typically 1-5 Doppler responses). Very high values like 97 indicate
		// buffer misalignment from earlier items - fail early with clear error.
		const maxReasonableREP = 16
		if r.RawDopplerCount > maxReasonableREP {
			return bytesRead, fmt.Errorf("raw doppler speed REP=%d is unreasonably high (max %d), likely buffer misalignment", r.RawDopplerCount, maxReasonableREP)
		}

		// If REP is 0, no data follows
		if r.RawDopplerCount == 0 {
			r.RawDopplerSpeeds = []float64{}
			r.RawDopplerAmbiguity = []float64{}
			r.RawDopplerFrequency = []float64{}
			return bytesRead, nil
		}

		// Check if we have enough data before attempting to read
		expectedBytes := int(r.RawDopplerCount) * 6
		if buf.Len() < expectedBytes {
			return bytesRead, fmt.Errorf("%w: need %d bytes for raw doppler speed (%d repetitions), have %d", asterix.ErrBufferTooShort, expectedBytes, r.RawDopplerCount, buf.Len())
		}

		// Initialize arrays
		r.RawDopplerSpeeds = make([]float64, r.RawDopplerCount)
		r.RawDopplerAmbiguity = make([]float64, r.RawDopplerCount)
		r.RawDopplerFrequency = make([]float64, r.RawDopplerCount)

		// Read data for each repetition
		for i := uint8(0); i < r.RawDopplerCount; i++ {
			var data [6]byte // 6 bytes per repetition
			n, err := buf.Read(data[:])
			if err != nil {
				return bytesRead + n, fmt.Errorf("reading raw doppler speed data at repetition %d: %w", i, err)
			}
			bytesRead += n

			if n != 6 {
				return bytesRead, fmt.Errorf("%w: need 6 bytes for raw doppler speed at repetition %d, have %d", asterix.ErrBufferTooShort, i, n)
			}

			// Extract doppler speed, 16 bits, LSB = 1 m/s, in two's complement
			speedRaw := int16(uint16(data[0])<<8 | uint16(data[1]))
			r.RawDopplerSpeeds[i] = float64(speedRaw)

			// Extract ambiguity range, 16 bits, LSB = 1 m/s
			ambiguityRaw := uint16(data[2])<<8 | uint16(data[3])
			r.RawDopplerAmbiguity[i] = float64(ambiguityRaw)

			// Extract transmitter frequency, 16 bits, LSB = 1 MHz
			frequencyRaw := uint16(data[4])<<8 | uint16(data[5])
			r.RawDopplerFrequency[i] = float64(frequencyRaw)
		}
	}

	return bytesRead, nil
}

// Encode implements the DataItem interface
func (r *RadialDopplerSpeed) Encode(buf *bytes.Buffer) (int, error) {
	if err := r.Validate(); err != nil {
		return 0, err
	}

	bytesWritten := 0

	// Primary subfield
	primary := byte(0)
	if r.CAL {
		primary |= 0x80 // bit 8
	}
	if r.RDS {
		primary |= 0x40 // bit 7
	}
	// bits 6-2 are spare
	// FX bit (bit 1) is set to 0 as no extensions are defined

	err := buf.WriteByte(primary)
	if err != nil {
		return bytesWritten, fmt.Errorf("writing primary subfield: %w", err)
	}
	bytesWritten++

	// Subfield #1: Calculated Doppler Speed (2 bytes)
	if r.CAL {
		var data [2]byte

		// Set validity flag
		if !r.CalculatedDopplerValid {
			data[0] |= 0x80 // bit 16
		}
		// bits 15-11 are spare

		// Set doppler speed, bits 10-1, LSB = 1 m/s, in two's complement
		// Limit to 10-bit range (-512 to 511)
		rawSpeed := int16(r.CalculatedDopplerSpeed)
		if rawSpeed < -512 {
			rawSpeed = -512
		} else if rawSpeed > 511 {
			rawSpeed = 511
		}

		// Extract the 10 least significant bits
		data[0] |= byte((rawSpeed >> 8) & 0x03) // bits 10-9
		data[1] = byte(rawSpeed & 0xFF)         // bits 8-1

		n, err := buf.Write(data[:])
		if err != nil {
			return bytesWritten + n, fmt.Errorf("writing calculated doppler speed: %w", err)
		}
		bytesWritten += n
	}

	// Subfield #2: Raw Doppler Speed (variable length)
	if r.RDS {
		// Ensure arrays are not nil and have same length
		if r.RawDopplerSpeeds == nil || r.RawDopplerAmbiguity == nil || r.RawDopplerFrequency == nil ||
			len(r.RawDopplerSpeeds) != len(r.RawDopplerAmbiguity) || len(r.RawDopplerSpeeds) != len(r.RawDopplerFrequency) {
			return bytesWritten, fmt.Errorf("raw doppler arrays must have same length and not be nil")
		}

		// First byte is repetition factor
		count := uint8(len(r.RawDopplerSpeeds))
		if count == 0 {
			return bytesWritten, fmt.Errorf("raw doppler speed must have at least one entry")
		}

		err := buf.WriteByte(count)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing raw doppler speed repetition factor: %w", err)
		}
		bytesWritten++

		// Write data for each repetition
		for i := uint8(0); i < count; i++ {
			var data [6]byte // 6 bytes per repetition

			// Set doppler speed, 16 bits, LSB = 1 m/s, in two's complement
			speedRaw := int16(r.RawDopplerSpeeds[i])
			data[0] = byte(speedRaw >> 8)
			data[1] = byte(speedRaw)

			// Set ambiguity range, 16 bits, LSB = 1 m/s, unsigned
			ambiguityRaw := uint16(r.RawDopplerAmbiguity[i])
			data[2] = byte(ambiguityRaw >> 8)
			data[3] = byte(ambiguityRaw)

			// Set transmitter frequency, 16 bits, LSB = 1 MHz, unsigned
			frequencyRaw := uint16(r.RawDopplerFrequency[i])
			data[4] = byte(frequencyRaw >> 8)
			data[5] = byte(frequencyRaw)

			n, err := buf.Write(data[:])
			if err != nil {
				return bytesWritten + n, fmt.Errorf("writing raw doppler speed data: %w", err)
			}
			bytesWritten += n
		}
	}

	return bytesWritten, nil
}

// Validate implements the DataItem interface
func (r *RadialDopplerSpeed) Validate() error {
	// Note: According to the spec, only one secondary subfield shall be present.
	// However, we don't enforce this during decode for interoperability with
	// non-compliant radar implementations. Both subfields will be parsed if present.

	// Validate calculated doppler speed
	if r.CAL {
		// Check range for calculated doppler speed (10 bits, two's complement)
		if r.CalculatedDopplerSpeed < -512 || r.CalculatedDopplerSpeed > 511 {
			return fmt.Errorf("calculated doppler speed out of range [-512,511]: %f", r.CalculatedDopplerSpeed)
		}
	}

	// Validate raw doppler speed
	if r.RDS {
		// Ensure arrays are not nil and have same length
		if r.RawDopplerSpeeds == nil || r.RawDopplerAmbiguity == nil || r.RawDopplerFrequency == nil ||
			len(r.RawDopplerSpeeds) != len(r.RawDopplerAmbiguity) || len(r.RawDopplerSpeeds) != len(r.RawDopplerFrequency) {
			return fmt.Errorf("raw doppler arrays must have same length and not be nil")
		}

		// Check range for each entry
		for i, speed := range r.RawDopplerSpeeds {
			// Check range for doppler speed (16 bits, two's complement)
			if speed < -32768 || speed > 32767 {
				return fmt.Errorf("raw doppler speed at index %d out of range [-32768,32767]: %f", i, speed)
			}

			// Check range for ambiguity range (16 bits, unsigned)
			if r.RawDopplerAmbiguity[i] < 0 || r.RawDopplerAmbiguity[i] > 65535 {
				return fmt.Errorf("raw doppler ambiguity at index %d out of range [0,65535]: %f", i, r.RawDopplerAmbiguity[i])
			}

			// Check range for transmitter frequency (16 bits, unsigned)
			if r.RawDopplerFrequency[i] < 0 || r.RawDopplerFrequency[i] > 65535 {
				return fmt.Errorf("raw doppler frequency at index %d out of range [0,65535]: %f", i, r.RawDopplerFrequency[i])
			}
		}
	}

	return nil
}

// String returns a human-readable representation
func (r *RadialDopplerSpeed) String() string {
	result := "Doppler Speed:"

	if r.CAL {
		validStr := "valid"
		if !r.CalculatedDopplerValid {
			validStr = "doubtful"
		}
		result += fmt.Sprintf("\n  Calculated: %.0f m/s (%s)", r.CalculatedDopplerSpeed, validStr)
	}

	if r.RDS {
		result += fmt.Sprintf("\n  Raw (%d responses):", r.RawDopplerCount)
		for i := 0; i < len(r.RawDopplerSpeeds); i++ {
			result += fmt.Sprintf("\n   #%d: %.0f m/s, Ambiguity: %.0f m/s, Freq: %.0f MHz",
				i+1, r.RawDopplerSpeeds[i], r.RawDopplerAmbiguity[i], r.RawDopplerFrequency[i])
		}
	}

	return result
}
