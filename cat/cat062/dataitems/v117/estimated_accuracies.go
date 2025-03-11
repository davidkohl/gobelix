// dataitems/cat062/estimated_accuracies.go
package v117

import (
	"bytes"
	"fmt"
)

// EstimatedAccuracies implements I062/500
// Overview of all important accuracies
type EstimatedAccuracies struct {
	Data []byte
}

func (e *EstimatedAccuracies) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0
	e.Data = nil

	// Primary subfield - 1 or 2 octets depending on FX bit
	primaryByte := make([]byte, 1)
	n, err := buf.Read(primaryByte)
	if err != nil {
		return n, fmt.Errorf("reading estimated accuracies primary subfield: %w", err)
	}
	bytesRead += n
	e.Data = append(e.Data, primaryByte[0])

	// Check for primary subfield extension
	hasExtension := (primaryByte[0] & 0x01) != 0

	// Read second primary subfield byte if extension bit is set
	if hasExtension {
		secondByte := make([]byte, 1)
		n, err := buf.Read(secondByte)
		if err != nil {
			return bytesRead, fmt.Errorf("reading estimated accuracies primary extension: %w", err)
		}
		bytesRead += n
		e.Data = append(e.Data, secondByte[0])
	}

	// Process each subfield based on bits in the primary subfield
	// APC: bit-16 (bit-8 of first byte) Subfield #1: Estimated Accuracy Of Track Position (Cartesian)
	if (e.Data[0] & 0x80) != 0 {
		subfieldData := make([]byte, 4)
		n, err := buf.Read(subfieldData)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading APC subfield: %w", err)
		}
		bytesRead += n
		e.Data = append(e.Data, subfieldData...)
	}

	// COV: bit-15 (bit-7 of first byte) Subfield #2: XY Covariance
	if (e.Data[0] & 0x40) != 0 {
		subfieldData := make([]byte, 2)
		n, err := buf.Read(subfieldData)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading COV subfield: %w", err)
		}
		bytesRead += n
		e.Data = append(e.Data, subfieldData...)
	}

	// APW: bit-14 (bit-6 of first byte) Subfield #3: Estimated Accuracy Of Track Position (WGS-84)
	if (e.Data[0] & 0x20) != 0 {
		subfieldData := make([]byte, 4)
		n, err := buf.Read(subfieldData)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading APW subfield: %w", err)
		}
		bytesRead += n
		e.Data = append(e.Data, subfieldData...)
	}

	// AGA: bit-13 (bit-5 of first byte) Subfield #4: Estimated Accuracy Of Calculated Track Geometric Altitude
	if (e.Data[0] & 0x10) != 0 {
		subfieldData := make([]byte, 1)
		n, err := buf.Read(subfieldData)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading AGA subfield: %w", err)
		}
		bytesRead += n
		e.Data = append(e.Data, subfieldData...)
	}

	// ABA: bit-12 (bit-4 of first byte) Subfield #5: Estimated Accuracy Of Calculated Track Barometric Altitude
	if (e.Data[0] & 0x08) != 0 {
		subfieldData := make([]byte, 1)
		n, err := buf.Read(subfieldData)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading ABA subfield: %w", err)
		}
		bytesRead += n
		e.Data = append(e.Data, subfieldData...)
	}

	// ATV: bit-11 (bit-3 of first byte) Subfield #6: Estimated Accuracy Of Track Velocity (Cartesian)
	if (e.Data[0] & 0x04) != 0 {
		subfieldData := make([]byte, 2)
		n, err := buf.Read(subfieldData)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading ATV subfield: %w", err)
		}
		bytesRead += n
		e.Data = append(e.Data, subfieldData...)
	}

	// AA: bit-10 (bit-2 of first byte) Subfield #7: Estimated Accuracy Of Acceleration (Cartesian)
	if (e.Data[0] & 0x02) != 0 {
		subfieldData := make([]byte, 2)
		n, err := buf.Read(subfieldData)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading AA subfield: %w", err)
		}
		bytesRead += n
		e.Data = append(e.Data, subfieldData...)
	}

	// If we have a second primary byte, check its bits
	if hasExtension && len(e.Data) > 1 {
		// ARC: bit-8 (bit-8 of second byte) Subfield #8: Estimated Accuracy Of Rate Of Climb/Descent
		if (e.Data[1] & 0x80) != 0 {
			subfieldData := make([]byte, 1)
			n, err := buf.Read(subfieldData)
			if err != nil {
				return bytesRead + n, fmt.Errorf("reading ARC subfield: %w", err)
			}
			bytesRead += n
			e.Data = append(e.Data, subfieldData...)
		}
	}

	return bytesRead, nil
}

func (e *EstimatedAccuracies) Encode(buf *bytes.Buffer) (int, error) {
	if len(e.Data) == 0 {
		// If no data, encode a minimal valid value
		return buf.Write([]byte{0})
	}
	return buf.Write(e.Data)
}

func (e *EstimatedAccuracies) String() string {
	return fmt.Sprintf("EstimatedAccuracies[%d bytes]", len(e.Data))
}

func (e *EstimatedAccuracies) Validate() error {
	return nil
}
