// dataitems/cat062/track_data_ages.go
package v120

import (
	"bytes"
	"fmt"
)

// TrackDataAges implements I062/295
// Compound data item with a primary subfield of up to five octets
type TrackDataAges struct {
	Data []byte
}

func (t *TrackDataAges) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0
	t.Data = nil

	// Primary subfield can be up to 5 octets, each with FX bit
	for i := 0; i < 5; i++ {
		octet := make([]byte, 1)
		n, err := buf.Read(octet)
		if err != nil {
			return bytesRead, fmt.Errorf("reading track data ages primary subfield octet %d: %w", i+1, err)
		}
		bytesRead += n
		t.Data = append(t.Data, octet[0])

		// Check if further extension exists
		hasExtension := (octet[0] & 0x01) != 0
		if !hasExtension {
			break
		}
	}

	// Now read each subfield based on bits set in the primary subfield
	// This is an extensive subfield with up to 40 bits (5 octets * 8 bits)
	// We'll check all bits except FX bits (bits 1, 9, 17, 25, 33)
	for byteIdx := 0; byteIdx < len(t.Data); byteIdx++ {
		octet := t.Data[byteIdx]
		for bitIdx := 7; bitIdx > 0; bitIdx-- { // Skip bit 0 (FX)
			if (octet & (1 << bitIdx)) != 0 {
				// All subfields are 1 octet
				subfieldData := make([]byte, 1)
				n, err := buf.Read(subfieldData)
				if err != nil || n != 1 {
					return bytesRead + n, fmt.Errorf("reading track data ages subfield: %w", err)
				}
				bytesRead += n
				t.Data = append(t.Data, subfieldData...)
			}
		}
	}

	return bytesRead, nil
}

func (t *TrackDataAges) Encode(buf *bytes.Buffer) (int, error) {
	if len(t.Data) == 0 {
		// If no data, encode a minimal valid value (empty FSPEC)
		return buf.Write([]byte{0})
	}
	return buf.Write(t.Data)
}

func (t *TrackDataAges) String() string {
	return fmt.Sprintf("TrackDataAges[%d bytes]", len(t.Data))
}

func (t *TrackDataAges) Validate() error {
	return nil
}
