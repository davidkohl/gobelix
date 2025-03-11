// dataitems/cat062/track_data_ages.go
package v117

import (
	"bytes"
	"fmt"
)

// TrackDataAges implements I062/295
// Ages of the data provided
type TrackDataAges struct {
	Data []byte
}

func (t *TrackDataAges) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0
	t.Data = nil

	// Primary subfield can be up to 5 octets, each with FX bit
	var primaryBytes []byte

	for i := 0; i < 5; i++ {
		if buf.Len() < 1 {
			if i == 0 {
				return 0, fmt.Errorf("buffer too short for primary subfield")
			}
			// We've read at least one octet, so this is not a critical error
			break
		}

		octet, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading primary subfield octet %d: %w", i+1, err)
		}
		bytesRead++
		primaryBytes = append(primaryBytes, octet)

		// Store in the data
		t.Data = append(t.Data, octet)

		// Check if we need to continue (FX bit)
		if octet&0x01 == 0 {
			break
		}
	}

	if len(primaryBytes) == 0 {
		// No primary bytes read, return with what we've got
		return bytesRead, nil
	}

	// Now process each bit in the primary subfield to determine which subfields are present
	// Each subfield is 1 octet
	for _, octet := range primaryBytes {
		// Check each bit except the FX bit (bit-1)
		for bitPos := 7; bitPos > 0; bitPos-- {
			if octet&(1<<bitPos) == 0 {
				continue // Bit not set
			}

			// Check buffer before reading
			if buf.Len() < 1 {
				// Instead of returning an error, we can return what we've successfully read
				// This allows for partial decoding of messages
				return bytesRead, nil
			}

			// All subfields are 1 octet in size
			subfieldByte, err := buf.ReadByte()
			if err != nil {
				// Return what we've read so far instead of failing completely
				return bytesRead, nil
			}
			bytesRead++
			t.Data = append(t.Data, subfieldByte)
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
	// Basic validation - at least one FSPEC byte should be present
	if len(t.Data) == 0 {
		return fmt.Errorf("empty data")
	}
	return nil
}
