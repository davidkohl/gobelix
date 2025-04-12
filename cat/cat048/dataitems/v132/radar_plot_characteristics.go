// dataitems/cat048/radar_plot_characteristics.go
package v132

import (
	"bytes"
	"fmt"
	"math"
)

// RadarPlotCharacteristics implements I048/130
// Additional information on the quality of the target report.
type RadarPlotCharacteristics struct {
	// Subfields availability flags
	SRL bool // SSR plot runlength
	SRR bool // Number of received replies for SSR
	SAM bool // Amplitude of SSR reply
	PRL bool // PSR plot runlength
	PAM bool // PSR amplitude
	RPD bool // Difference in range between PSR and SSR plot
	APD bool // Difference in azimuth between PSR and SSR plot

	// Subfield values
	SSRRunLength      float64 // SSR plot runlength (degrees)
	SSRReplyCount     uint8   // Number of received replies for SSR
	SSRAmplitude      int8    // Amplitude of SSR reply (dBm)
	PSRRunLength      float64 // PSR plot runlength (degrees)
	PSRAmplitude      int8    // PSR amplitude (dBm)
	RangeDifference   float64 // Range difference PSR-SSR (NM)
	AzimuthDifference float64 // Azimuth difference PSR-SSR (degrees)
}

// Decode implements the DataItem interface
func (r *RadarPlotCharacteristics) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Primary subfield
	primary := make([]byte, 1)
	n, err := buf.Read(primary)
	if err != nil {
		return n, fmt.Errorf("reading radar plot characteristics primary subfield: %w", err)
	}
	bytesRead += n

	// Extract subfield presence flags
	r.SRL = (primary[0] & 0x80) != 0 // bit 8
	r.SRR = (primary[0] & 0x40) != 0 // bit 7
	r.SAM = (primary[0] & 0x20) != 0 // bit 6
	r.PRL = (primary[0] & 0x10) != 0 // bit 5
	r.PAM = (primary[0] & 0x08) != 0 // bit 4
	r.RPD = (primary[0] & 0x04) != 0 // bit 3
	r.APD = (primary[0] & 0x02) != 0 // bit 2
	fx := (primary[0] & 0x01) != 0   // bit 1 (FX)

	if fx {
		// FX bit is set in primary subfield, which means there's an extension
		// Not defined in the specification yet, just skip for now
		return bytesRead, fmt.Errorf("FX bit set in primary subfield, but extensions are not defined in the specification")
	}

	// Read all the subfields that are present

	// Subfield #1: SSR Plot Runlength
	if r.SRL {
		data := make([]byte, 1)
		n, err := buf.Read(data)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading SSR plot runlength: %w", err)
		}
		bytesRead += n

		// LSB = 360/2^13 degrees = 0.044 degrees
		r.SSRRunLength = float64(data[0]) * (360.0 / 8192.0)
	}

	// Subfield #2: Number of Received Replies for SSR
	if r.SRR {
		data := make([]byte, 1)
		n, err := buf.Read(data)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading SSR reply count: %w", err)
		}
		bytesRead += n

		r.SSRReplyCount = data[0]
	}

	// Subfield #3: Amplitude of SSR Reply
	if r.SAM {
		data := make([]byte, 1)
		n, err := buf.Read(data)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading SSR amplitude: %w", err)
		}
		bytesRead += n

		// Amplitude in dBm, two's complement
		r.SSRAmplitude = int8(data[0])
	}

	// Subfield #4: PSR Plot Runlength
	if r.PRL {
		data := make([]byte, 1)
		n, err := buf.Read(data)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading PSR plot runlength: %w", err)
		}
		bytesRead += n

		// LSB = 360/2^13 degrees = 0.044 degrees
		r.PSRRunLength = float64(data[0]) * (360.0 / 8192.0)
	}

	// Subfield #5: PSR Amplitude
	if r.PAM {
		data := make([]byte, 1)
		n, err := buf.Read(data)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading PSR amplitude: %w", err)
		}
		bytesRead += n

		// Amplitude in dBm, two's complement
		r.PSRAmplitude = int8(data[0])
	}

	// Subfield #6: Difference in Range between PSR and SSR plot
	if r.RPD {
		data := make([]byte, 1)
		n, err := buf.Read(data)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading range difference: %w", err)
		}
		bytesRead += n

		// Range in 1/256 NM, two's complement
		r.RangeDifference = float64(int8(data[0])) / 256.0
	}

	// Subfield #7: Difference in Azimuth between PSR and SSR plot
	if r.APD {
		data := make([]byte, 1)
		n, err := buf.Read(data)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading azimuth difference: %w", err)
		}
		bytesRead += n

		// Azimuth in 360/2^14 degrees, two's complement
		r.AzimuthDifference = float64(int8(data[0])) * (360.0 / 16384.0)
	}

	return bytesRead, nil
}

// Encode implements the DataItem interface
func (r *RadarPlotCharacteristics) Encode(buf *bytes.Buffer) (int, error) {
	bytesWritten := 0

	// Primary subfield
	primary := byte(0)
	if r.SRL {
		primary |= 0x80 // bit 8
	}
	if r.SRR {
		primary |= 0x40 // bit 7
	}
	if r.SAM {
		primary |= 0x20 // bit 6
	}
	if r.PRL {
		primary |= 0x10 // bit 5
	}
	if r.PAM {
		primary |= 0x08 // bit 4
	}
	if r.RPD {
		primary |= 0x04 // bit 3
	}
	if r.APD {
		primary |= 0x02 // bit 2
	}
	// FX bit (bit 1) is set to 0 as no extensions are defined

	err := buf.WriteByte(primary)
	if err != nil {
		return bytesWritten, fmt.Errorf("writing primary subfield: %w", err)
	}
	bytesWritten++

	// Write all the subfields that are present

	// Subfield #1: SSR Plot Runlength
	if r.SRL {
		// Convert to raw value
		rawValue := uint8(math.Round(r.SSRRunLength * (8192.0 / 360.0)))

		err := buf.WriteByte(rawValue)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing SSR plot runlength: %w", err)
		}
		bytesWritten++
	}

	// Subfield #2: Number of Received Replies for SSR
	if r.SRR {
		err := buf.WriteByte(r.SSRReplyCount)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing SSR reply count: %w", err)
		}
		bytesWritten++
	}

	// Subfield #3: Amplitude of SSR Reply
	if r.SAM {
		err := buf.WriteByte(byte(r.SSRAmplitude))
		if err != nil {
			return bytesWritten, fmt.Errorf("writing SSR amplitude: %w", err)
		}
		bytesWritten++
	}

	// Subfield #4: PSR Plot Runlength
	if r.PRL {
		// Convert to raw value
		rawValue := uint8(math.Round(r.PSRRunLength * (8192.0 / 360.0)))

		err := buf.WriteByte(rawValue)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing PSR plot runlength: %w", err)
		}
		bytesWritten++
	}

	// Subfield #5: PSR Amplitude
	if r.PAM {
		err := buf.WriteByte(byte(r.PSRAmplitude))
		if err != nil {
			return bytesWritten, fmt.Errorf("writing PSR amplitude: %w", err)
		}
		bytesWritten++
	}

	// Subfield #6: Difference in Range between PSR and SSR plot
	if r.RPD {
		// Convert to raw value
		rawValue := int8(math.Round(r.RangeDifference * 256.0))

		err := buf.WriteByte(byte(rawValue))
		if err != nil {
			return bytesWritten, fmt.Errorf("writing range difference: %w", err)
		}
		bytesWritten++
	}

	// Subfield #7: Difference in Azimuth between PSR and SSR plot
	if r.APD {
		// Convert to raw value
		rawValue := int8(math.Round(r.AzimuthDifference * (16384.0 / 360.0)))

		err := buf.WriteByte(byte(rawValue))
		if err != nil {
			return bytesWritten, fmt.Errorf("writing azimuth difference: %w", err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

// Validate implements the DataItem interface
func (r *RadarPlotCharacteristics) Validate() error {
	// Check ranges for all values
	if r.SRL && (r.SSRRunLength < 0 || r.SSRRunLength > 11.21) {
		return fmt.Errorf("SSR runlength out of range [0,11.21]: %f", r.SSRRunLength)
	}

	if r.PRL && (r.PSRRunLength < 0 || r.PSRRunLength > 11.21) {
		return fmt.Errorf("PSR runlength out of range [0,11.21]: %f", r.PSRRunLength)
	}

	if r.RPD && (r.RangeDifference < -0.5 || r.RangeDifference > 0.5) {
		return fmt.Errorf("range difference out of range [-0.5,0.5]: %f", r.RangeDifference)
	}

	if r.APD && (r.AzimuthDifference < -2.8125 || r.AzimuthDifference > 2.8125) {
		return fmt.Errorf("azimuth difference out of range [-2.8125,2.8125]: %f", r.AzimuthDifference)
	}

	return nil
}

// String returns a human-readable representation
func (r *RadarPlotCharacteristics) String() string {
	result := "Plot Characteristics:"

	if r.SRL {
		result += fmt.Sprintf("\n  SSR Runlength: %.3f°", r.SSRRunLength)
	}

	if r.SRR {
		result += fmt.Sprintf("\n  SSR Replies: %d", r.SSRReplyCount)
	}

	if r.SAM {
		result += fmt.Sprintf("\n  SSR Amplitude: %d dBm", r.SSRAmplitude)
	}

	if r.PRL {
		result += fmt.Sprintf("\n  PSR Runlength: %.3f°", r.PSRRunLength)
	}

	if r.PAM {
		result += fmt.Sprintf("\n  PSR Amplitude: %d dBm", r.PSRAmplitude)
	}

	if r.RPD {
		result += fmt.Sprintf("\n  Range Diff: %.4f NM", r.RangeDifference)
	}

	if r.APD {
		result += fmt.Sprintf("\n  Azimuth Diff: %.4f°", r.AzimuthDifference)
	}

	return result
}
