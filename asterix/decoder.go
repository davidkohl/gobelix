// asterix/decoder.go
package asterix

import (
	"bytes"
	"fmt"
)

// Decoder handles decoding of ASTERIX data
type Decoder struct {
	// map of known UAPs by category
	uaps map[Category]UAP
}

// NewDecoder creates a new ASTERIX decoder with registered UAPs
func NewDecoder(uaps ...UAP) (*Decoder, error) {
	d := &Decoder{
		uaps: make(map[Category]UAP),
	}

	// Register provided UAPs
	for _, uap := range uaps {
		if uap == nil {
			return nil, fmt.Errorf("%w: UAP cannot be nil", ErrInvalidMessage)
		}
		d.uaps[uap.Category()] = uap
	}

	return d, nil
}

// Decode takes raw ASTERIX data and returns the decoded data items
func (d *Decoder) Decode(data []byte) ([]map[string]DataItem, error) {
	if len(data) < 3 {
		return nil, fmt.Errorf("%w: data too short", ErrInvalidMessage)
	}

	// Read category
	cat := Category(data[0])

	// Find appropriate UAP
	uap, exists := d.uaps[cat]
	if !exists {
		return nil, fmt.Errorf("%w: %d", ErrUnknownCategory, cat)
	}

	// Read length
	length := uint16(data[1])<<8 | uint16(data[2])
	if int(length) != len(data) {
		return nil, fmt.Errorf("%w: expected %d, got %d", ErrInvalidLength, length, len(data))
	}

	buf := bytes.NewBuffer(data[3:]) // Skip CAT and length
	var results []map[string]DataItem

	// Decode records until buffer is exhausted
	for buf.Len() > 0 {
		items, err := d.decodeRecord(buf, uap)
		if err != nil {
			return nil, err
		}
		results = append(results, items)
	}

	return results, nil
}

func (d *Decoder) decodeRecord(buf *bytes.Buffer, uap UAP) (map[string]DataItem, error) {
	fspec := NewFSPEC()
	n, err := fspec.Decode(buf)
	_ = n
	if err != nil {
		return nil, fmt.Errorf("decoding FSPEC: %w", err)
	}

	items := make(map[string]DataItem)

	// Decode items based on FSPEC
	for _, field := range uap.Fields() {
		if !fspec.GetFRN(field.FRN) {
			continue
		}

		item, err := uap.CreateDataItem(field.DataItem)
		if err != nil {
			// Skip unknown data item based on its type
			if err := d.skipField(buf, field); err != nil {
				return nil, fmt.Errorf("skipping field %s: %w", field.DataItem, err)
			}
			continue
		}

		_, err = item.Decode(buf)
		if err != nil {
			return nil, fmt.Errorf("decoding %s: %w", field.DataItem, err)
		}

		items[field.DataItem] = item
	}

	return items, nil
}

func (d *Decoder) skipField(buf *bytes.Buffer, field DataField) error {
	switch field.Type {
	case Fixed:
		if buf.Len() < int(field.Length) {
			return fmt.Errorf("buffer too short for fixed field")
		}
		buf.Next(int(field.Length))
		return nil

	case Extended:
		for {
			b, err := buf.ReadByte()
			if err != nil {
				return fmt.Errorf("reading extended field byte: %w", err)
			}
			if b&0x01 == 0 {
				return nil
			}
		}

	case Repetitive:
		rep, err := buf.ReadByte()
		if err != nil {
			return fmt.Errorf("reading repetition factor: %w", err)
		}
		length := int(rep) * int(field.Length)
		if buf.Len() < length {
			return fmt.Errorf("buffer too short for repetitive field")
		}
		buf.Next(length)
		return nil

	case Compound:
		// Read primary field
		b, err := buf.ReadByte()
		_ = b
		if err != nil {
			return fmt.Errorf("reading compound primary field: %w", err)
		}
		// For now we'll just indicate we can't handle compound fields
		return fmt.Errorf("compound field skipping not implemented")

	default:
		return fmt.Errorf("unknown field type: %v", field.Type)
	}
}
