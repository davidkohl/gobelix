// asterix/encoder.go
package asterix

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// Encoder handles encoding of ASTERIX data
type Encoder struct {
	encoders map[Category]*CategoryEncoder
}

// CategoryEncoder holds category-specific encoding information
type CategoryEncoder struct {
	category   Category
	fieldSpecs []FieldSpec
	uap        UAP
}

// NewEncoder creates a new ASTERIX encoder
func NewEncoder(uaps ...UAP) (*Encoder, error) {
	e := &Encoder{
		encoders: make(map[Category]*CategoryEncoder),
	}

	for _, uap := range uaps {
		if uap == nil {
			return nil, fmt.Errorf("%w: UAP cannot be nil", ErrInvalidMessage)
		}

		ce, err := newCategoryEncoder(uap)
		if err != nil {
			return nil, fmt.Errorf("creating encoder for category %v: %w", uap.Category(), err)
		}
		e.encoders[uap.Category()] = ce
	}

	return e, nil
}

// newCategoryEncoder creates a new category-specific encoder
func newCategoryEncoder(uap UAP) (*CategoryEncoder, error) {
	ce := &CategoryEncoder{
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
		ce.fieldSpecs = append(ce.fieldSpecs, spec)
	}

	return ce, nil
}

// Encode writes an AsterixMessage to an io.Writer
func (e *Encoder) Encode(writer io.Writer, msg *AsterixMessage) error {
	// Validate message
	if msg == nil {
		return fmt.Errorf("%w: message cannot be nil", ErrInvalidMessage)
	}

	// Get category encoder
	ce, exists := e.encoders[msg.Category]
	if !exists {
		return fmt.Errorf("%w: %d", ErrUnknownCategory, msg.Category)
	}

	// Serialize message
	buf := new(bytes.Buffer)

	// Write category
	if err := buf.WriteByte(byte(msg.Category)); err != nil {
		return fmt.Errorf("writing category: %w", err)
	}

	// Reserve space for length
	if _, err := buf.Write([]byte{0, 0}); err != nil {
		return fmt.Errorf("reserving space for length: %w", err)
	}

	// Encode each record
	for i, record := range msg.records {
		// Create FSPEC
		fspec := NewFSPEC()
		for _, spec := range ce.fieldSpecs {
			if _, exists := record[spec.DataItem]; exists {
				if err := fspec.SetFRN(spec.FRN); err != nil {
					return fmt.Errorf("record %d: setting FRN for %s: %w",
						i, spec.DataItem, err)
				}
			}
		}

		// Write FSPEC
		if _, err := fspec.Encode(buf); err != nil {
			return fmt.Errorf("record %d: encoding FSPEC: %w", i, err)
		}

		// Write items in FRN order
		for _, spec := range ce.fieldSpecs {
			item, exists := record[spec.DataItem]
			if !exists {
				continue
			}

			if err := item.Validate(); err != nil {
				return fmt.Errorf("record %d: validating %s: %w",
					i, spec.DataItem, err)
			}

			if _, err := item.Encode(buf); err != nil {
				return fmt.Errorf("record %d: encoding %s: %w",
					i, spec.DataItem, err)
			}
		}
	}

	// Update length field
	data := buf.Bytes()
	binary.BigEndian.PutUint16(data[1:3], uint16(len(data)))

	// Write to output
	_, err := writer.Write(data)
	return err
}

// EncodeBatch encodes multiple records into a single data block
func (e *Encoder) EncodeBatch(cat Category, records []map[string]DataItem) ([]byte, error) {
	ce, exists := e.encoders[cat]
	if !exists {
		return nil, fmt.Errorf("%w: %d", ErrUnknownCategory, cat)
	}

	buf := new(bytes.Buffer)

	// Write category
	if err := buf.WriteByte(byte(ce.category)); err != nil {
		return nil, fmt.Errorf("writing category: %w", err)
	}

	// Reserve space for length
	if err := binary.Write(buf, binary.BigEndian, uint16(0)); err != nil {
		return nil, fmt.Errorf("reserving length: %w", err)
	}

	// Encode each record
	for i, items := range records {
		// Create FSPEC
		fspec := NewFSPEC()
		for _, spec := range ce.fieldSpecs {
			if _, exists := items[spec.DataItem]; exists {
				if err := fspec.SetFRN(spec.FRN); err != nil {
					return nil, fmt.Errorf("record %d: setting FRN for %s: %w",
						i, spec.DataItem, err)
				}
			}
		}

		// Write FSPEC
		if _, err := fspec.Encode(buf); err != nil {
			return nil, fmt.Errorf("record %d: encoding FSPEC: %w", i, err)
		}

		// Write items in FRN order
		for _, spec := range ce.fieldSpecs {
			item, exists := items[spec.DataItem]
			if !exists {
				continue
			}

			if err := item.Validate(); err != nil {
				return nil, fmt.Errorf("record %d: validating %s: %w",
					i, spec.DataItem, err)
			}

			if _, err := item.Encode(buf); err != nil {
				return nil, fmt.Errorf("record %d: encoding %s: %w",
					i, spec.DataItem, err)
			}
		}
	}

	// Update length
	data := buf.Bytes()
	binary.BigEndian.PutUint16(data[1:3], uint16(len(data)))

	return data, nil
}
