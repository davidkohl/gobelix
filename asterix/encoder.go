// asterix/encoder.go
package asterix

import (
	"bytes"
	"encoding/binary"
	"fmt"
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

// Encode takes an ASTERIX category and data items and returns encoded data
func (e *Encoder) Encode(cat Category, items map[string]DataItem) ([]byte, error) {
	ce, exists := e.encoders[cat]
	if !exists {
		return nil, fmt.Errorf("%w: %d", ErrUnknownCategory, cat)
	}

	return ce.encode(items)
}

// encode handles category-specific encoding
func (ce *CategoryEncoder) encode(items map[string]DataItem) ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write category
	if err := buf.WriteByte(byte(ce.category)); err != nil {
		return nil, fmt.Errorf("writing category: %w", err)
	}

	// Reserve space for length
	if err := binary.Write(buf, binary.BigEndian, uint16(0)); err != nil {
		return nil, fmt.Errorf("reserving length: %w", err)
	}

	// Create FSPEC
	fspec := NewFSPEC()
	for _, spec := range ce.fieldSpecs {
		if _, exists := items[spec.DataItem]; exists {
			if err := fspec.SetFRN(spec.FRN); err != nil {
				return nil, fmt.Errorf("setting FRN for %s: %w", spec.DataItem, err)
			}
		}
	}

	// Write FSPEC
	if _, err := fspec.Encode(buf); err != nil {
		return nil, fmt.Errorf("encoding FSPEC: %w", err)
	}

	// Write items in FRN order
	for _, spec := range ce.fieldSpecs {
		item, exists := items[spec.DataItem]
		if !exists {
			continue
		}

		if err := item.Validate(); err != nil {
			return nil, fmt.Errorf("validating %s: %w", spec.DataItem, err)
		}

		if _, err := item.Encode(buf); err != nil {
			return nil, fmt.Errorf("encoding %s: %w", spec.DataItem, err)
		}
	}

	// Update length
	data := buf.Bytes()
	binary.BigEndian.PutUint16(data[1:3], uint16(len(data)))

	return data, nil
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
