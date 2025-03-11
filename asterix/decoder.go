// asterix/decoder.go
package asterix

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

// Decoder handles decoding of ASTERIX data
type Decoder struct {
	decoders map[Category]*CategoryDecoder
}

// CategoryDecoder holds pre-compiled information for decoding a specific category
type CategoryDecoder struct {
	category   Category
	fieldSpecs []FieldSpec
	uap        UAP
}

// FieldSpec contains pre-compiled field information
type FieldSpec struct {
	FRN      uint8
	DataItem string
	Type     ItemType
	Length   uint8
}

// NewDecoder creates a decoder with the provided UAPs
func NewDecoder(uaps ...UAP) (*Decoder, error) {
	d := &Decoder{
		decoders: make(map[Category]*CategoryDecoder),
	}

	for _, uap := range uaps {
		if uap == nil {
			return nil, fmt.Errorf("%w: UAP cannot be nil", ErrInvalidMessage)
		}

		cd, err := newCategoryDecoder(uap)
		if err != nil {
			return nil, fmt.Errorf("creating decoder for category %v: %w", uap.Category(), err)
		}
		d.decoders[uap.Category()] = cd
	}

	return d, nil
}

// newCategoryDecoder creates a new category-specific decoder
func newCategoryDecoder(uap UAP) (*CategoryDecoder, error) {
	cd := &CategoryDecoder{
		category: uap.Category(),
		uap:      uap,
	}

	// Pre-compile field specifications
	for _, field := range uap.Fields() {
		spec := FieldSpec{
			FRN:      field.FRN,
			DataItem: field.DataItem,
			Type:     field.Type,
			Length:   field.Length,
		}
		cd.fieldSpecs = append(cd.fieldSpecs, spec)
	}

	return cd, nil
}

// Decode processes raw ASTERIX data
func (d *Decoder) Decode(data []byte) (*AsterixMessage, error) {
	if len(data) < 3 {
		return nil, fmt.Errorf("%w: data too short", ErrInvalidMessage)
	}

	cat := Category(data[0])
	cd, exists := d.decoders[cat]
	if !exists {
		return nil, fmt.Errorf("%w: %d", ErrUnknownCategory, cat)
	}

	// Check length
	length := binary.BigEndian.Uint16(data[1:3])
	if int(length) != len(data) {
		return nil, fmt.Errorf("%w: expected %d, got %d",
			ErrInvalidLength, length, len(data))
	}

	// Create the message structure
	msg := &AsterixMessage{
		Category:   cat,
		RawMessage: data,
		Timestamp:  time.Now(),
		uap:        cd.uap,
	}

	// Decode records
	records, err := cd.decode(bytes.NewBuffer(data[3:])) // Skip CAT/LEN
	if err != nil {
		return nil, fmt.Errorf("decoding records: %w", err)
	}

	// Store records
	msg.records = records

	return msg, nil
}

// decode processes data for a specific category
func (cd *CategoryDecoder) decode(buf *bytes.Buffer) ([]map[string]DataItem, error) {
	var results []map[string]DataItem

	for buf.Len() > 0 {
		// Check for at least one byte for FSPEC
		if buf.Len() < 1 {
			break // End of data reached
		}

		items, err := cd.decodeRecord(buf)
		if err != nil {
			// Handle EOF while processing the last record
			if err == io.EOF && buf.Len() == 0 {
				break
			}
			return nil, err
		}
		results = append(results, items)
	}

	return results, nil
}

// decodeRecord processes a single ASTERIX record
func (cd *CategoryDecoder) decodeRecord(buf *bytes.Buffer) (map[string]DataItem, error) {
	if buf.Len() == 0 {
		return nil, io.EOF
	}

	// Read FSPEC
	fspec := NewFSPEC()
	if _, err := fspec.Decode(buf); err != nil {
		return nil, fmt.Errorf("decoding FSPEC: %w", err)
	}

	// Decode fields using pre-compiled specs
	items := make(map[string]DataItem)
	for _, spec := range cd.fieldSpecs {
		if !fspec.GetFRN(spec.FRN) {
			continue
		}

		// For fixed length items, check if we have enough bytes
		if spec.Type == Fixed && buf.Len() < int(spec.Length) {
			return nil, fmt.Errorf("buffer too short for %s: need %d bytes, have %d",
				spec.DataItem, spec.Length, buf.Len())
		}

		item, err := cd.uap.CreateDataItem(spec.DataItem)
		if err != nil {
			if spec.Type == Fixed {
				// For fixed length items, we can skip even unknown ones
				if buf.Len() < int(spec.Length) {
					return nil, fmt.Errorf("buffer too short to skip %s: need %d bytes, have %d",
						spec.DataItem, spec.Length, buf.Len())
				}
				buf.Next(int(spec.Length))
				continue
			}
			return nil, fmt.Errorf("creating item %s: %w", spec.DataItem, err)
		}

		if _, err := item.Decode(buf); err != nil {
			return nil, fmt.Errorf("decoding %s: %w", spec.DataItem, err)
		}

		items[spec.DataItem] = item
	}

	return items, nil
}
