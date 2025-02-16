// asterix/encoder.go
package asterix

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Encoder handles encoding of ASTERIX data
type Encoder struct{}

// NewEncoder creates a new ASTERIX encoder
func NewEncoder() *Encoder {
	return &Encoder{}
}

// Encode takes a UAP and data items and returns encoded ASTERIX data
func (e *Encoder) Encode(uap UAP, items map[string]DataItem) ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write category
	if err := buf.WriteByte(byte(uap.Category())); err != nil {
		return nil, fmt.Errorf("writing category: %w", err)
	}

	// Reserve space for length
	if err := binary.Write(buf, binary.BigEndian, uint16(0)); err != nil {
		return nil, fmt.Errorf("reserving length: %w", err)
	}

	// Create FSPEC based on present items
	fspec := NewFSPEC()
	for id := range items {
		frn := uap.FRNByID(id)
		if frn == 0 {
			return nil, fmt.Errorf("%w: %s", ErrUnknownDataItem, id)
		}
		fspec.SetFRN(frn)
	}

	// Write FSPEC
	if _, err := fspec.Encode(buf); err != nil {
		return nil, fmt.Errorf("encoding FSPEC: %w", err)
	}

	// Write items in FRN order
	for _, field := range uap.Fields() {
		if !fspec.GetFRN(field.FRN) {
			continue
		}

		item, exists := items[field.DataItem]
		if !exists {
			return nil, fmt.Errorf("%w: %s marked in FSPEC but not present",
				ErrInvalidMessage, field.DataItem)
		}

		if _, err := item.Encode(buf); err != nil {
			return nil, fmt.Errorf("encoding %s: %w", field.DataItem, err)
		}
	}

	// Update length
	data := buf.Bytes()
	binary.BigEndian.PutUint16(data[1:3], uint16(len(data)))

	return data, nil
}
