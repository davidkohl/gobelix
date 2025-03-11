// dataitems/cat062/aircraft_derived_data.go
package v120

import (
	"bytes"
	"fmt"
	"io"
)

// AircraftDerivedData implements I062/380
// This is a stub implementation that simply reads and skips the data without parsing
type AircraftDerivedData struct {
	// Raw bytes of the data for now
	Data []byte
}

func (a *AircraftDerivedData) Decode(buf *bytes.Buffer) (int, error) {
	// For compound items, first byte is a primary subfield defining presence of subsequent fields
	primarySubfield := make([]byte, 1)
	n, err := buf.Read(primarySubfield)
	if err != nil {
		return n, fmt.Errorf("reading aircraft derived data primary subfield: %w", err)
	}

	// Store the first byte
	a.Data = append(a.Data, primarySubfield...)
	bytesRead := n

	// Determine if there are more bytes to read based on the FX bits
	hasExtension := (primarySubfield[0] & 0x01) != 0
	currentByte := primarySubfield[0]

	// Continue reading extension octets as long as FX bit is set
	for hasExtension {
		octet := make([]byte, 1)
		n, err := buf.Read(octet)
		if err != nil {
			if err == io.EOF {
				break
			}
			return bytesRead + n, fmt.Errorf("reading aircraft derived data extension: %w", err)
		}

		a.Data = append(a.Data, octet[0])
		bytesRead += n

		currentByte = octet[0]
		hasExtension = (currentByte & 0x01) != 0
	}

	// Now read each subfield based on bits set in the primary subfield and extensions
	// This is a stub implementation, so we'll just read up to 32 potential subfields
	// and assume each is present if the corresponding bit is set
	for i := 0; i < 32; i++ {
		byteIndex := i / 8
		bitPosition := 7 - (i % 8)

		// If we've run out of bytes to check, stop
		if byteIndex >= len(a.Data) {
			break
		}

		// Skip the FX bit as those are handled above
		if bitPosition == 0 {
			continue
		}

		// Check if the bit is set, indicating presence of this subfield
		if (a.Data[byteIndex] & (1 << bitPosition)) != 0 {
			// In a real implementation, we would determine the size of each subfield
			// and read the appropriate number of bytes. For now, we'll just
			// make a simplifying assumption that each subfield is 1-8 bytes
			// and read a reasonable portion of the buffer.
			subfieldData := make([]byte, 8)
			n, err := buf.Read(subfieldData)
			if err != nil {
				if err == io.EOF {
					break
				}
				return bytesRead + n, fmt.Errorf("reading aircraft derived data subfield %d: %w", i, err)
			}

			a.Data = append(a.Data, subfieldData[:n]...)
			bytesRead += n
		}
	}

	return bytesRead, nil
}

func (a *AircraftDerivedData) Encode(buf *bytes.Buffer) (int, error) {
	// For now, just write back the raw bytes if we have them
	if len(a.Data) > 0 {
		return buf.Write(a.Data)
	}

	// Otherwise return a minimal valid structure (empty FSPEC with no fields)
	return buf.Write([]byte{0})
}

func (a *AircraftDerivedData) String() string {
	return fmt.Sprintf("AircraftDerivedData[%d bytes]", len(a.Data))
}

func (a *AircraftDerivedData) Validate() error {
	return nil
}
