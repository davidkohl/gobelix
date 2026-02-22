package v12

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// TrackQuality represents I001/210 - Track Quality
// Variable length data item with FX extension mechanism
// Quality indicator with application-dependent bit signification
type TrackQuality struct {
	Quality []uint8 // Quality indicators (7 bits per octet, MSB first)
}

func (t *TrackQuality) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0
	t.Quality = make([]uint8, 0, 1)

	hasFX := true
	for hasFX {
		data, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("%w: incomplete track quality", asterix.ErrBufferTooShort)
		}
		bytesRead++

		// Extract 7-bit quality value (bits 8-2)
		qualityValue := (data >> 1) & 0x7F
		t.Quality = append(t.Quality, qualityValue)

		// Check FX bit for extension
		hasFX = (data & 0x01) != 0
	}

	return bytesRead, nil
}

func (t *TrackQuality) Encode(buf *bytes.Buffer) (int, error) {
	if len(t.Quality) == 0 {
		return 0, fmt.Errorf("%w: track quality must have at least one octet", asterix.ErrInvalidMessage)
	}

	bytesWritten := 0
	for i, qualityValue := range t.Quality {
		octet := (qualityValue & 0x7F) << 1 // bits 8-2

		// Set FX bit if not the last octet
		if i < len(t.Quality)-1 {
			octet |= 0x01 // FX bit
		}

		if err := buf.WriteByte(octet); err != nil {
			return bytesWritten, fmt.Errorf("writing track quality: %w", err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

func (t *TrackQuality) Validate() error {
	if len(t.Quality) == 0 {
		return fmt.Errorf("%w: track quality must have at least one octet", asterix.ErrInvalidMessage)
	}
	for i, q := range t.Quality {
		if q > 127 {
			return fmt.Errorf("%w: quality value at index %d out of range [0,127]: %d", asterix.ErrInvalidMessage, i, q)
		}
	}
	return nil
}

func (t *TrackQuality) String() string {
	if len(t.Quality) == 0 {
		return "Track Quality: None"
	}
	if len(t.Quality) == 1 {
		return fmt.Sprintf("Track Quality: %d", t.Quality[0])
	}
	return fmt.Sprintf("Track Quality: %v", t.Quality)
}
