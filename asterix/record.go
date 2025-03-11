// asterix/record.go
package asterix

import (
	"bytes"
	"fmt"
	"io"
)

// Record represents a single ASTERIX record
type Record struct {
	category Category
	fspec    *FSPEC
	items    map[string]DataItem
	uap      UAP
}

// NewRecord creates a new record for a specific category
func NewRecord(cat Category, uap UAP) (*Record, error) {
	if !cat.IsValid() {
		return nil, fmt.Errorf("%w: %d", ErrInvalidCategory, cat)
	}
	if uap == nil {
		return nil, fmt.Errorf("%w: UAP cannot be nil", ErrInvalidMessage)
	}
	if uap.Category() != cat {
		return nil, fmt.Errorf("%w: UAP category %d does not match record category %d",
			ErrInvalidMessage, uap.Category(), cat)
	}

	return &Record{
		category: cat,
		fspec:    NewFSPEC(),
		items:    make(map[string]DataItem),
		uap:      uap,
	}, nil
}

// SetDataItem adds or updates a data item
func (r *Record) SetDataItem(id string, item DataItem) error {
	if item == nil {
		return fmt.Errorf("%w: data item cannot be nil", ErrInvalidMessage)
	}

	// Find FRN for this item
	var frn uint8
	for _, field := range r.uap.Fields() {
		if field.DataItem == id {
			frn = field.FRN
			break
		}
	}

	if frn == 0 {
		return fmt.Errorf("%w: %s", ErrUnknownDataItem, id)
	}

	if err := item.Validate(); err != nil {
		return fmt.Errorf("validating %s: %w", id, err)
	}

	r.items[id] = item
	return r.fspec.SetFRN(frn)
}

// GetDataItem retrieves a data item by its ID
func (r *Record) GetDataItem(id string) (DataItem, string, bool) {
	item, exists := r.items[id]
	return item, fmt.Sprintf("%T", item), exists
}

// Encode writes the record to a buffer
func (r *Record) Encode(buf *bytes.Buffer) (int, error) {
	if err := r.uap.Validate(r.items); err != nil {
		return 0, err
	}

	bytesWritten := 0

	// Write FSPEC
	n, err := r.fspec.Encode(buf)
	if err != nil {
		return bytesWritten, fmt.Errorf("encoding FSPEC: %w", err)
	}
	bytesWritten += n

	// Write items in FRN order
	for _, field := range r.uap.Fields() {
		if !r.fspec.GetFRN(field.FRN) {
			continue
		}

		item, exists := r.items[field.DataItem]
		if !exists {
			return bytesWritten, fmt.Errorf("%w: %s marked in FSPEC but not present",
				ErrInvalidMessage, field.DataItem)
		}

		n, err := item.Encode(buf)
		if err != nil {
			return bytesWritten, fmt.Errorf("encoding %s: %w", field.DataItem, err)
		}
		bytesWritten += n
	}

	return bytesWritten, nil
}

// Decode reads a record from a buffer
func (r *Record) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() == 0 {
		return 0, io.EOF
	}

	bytesRead := 0

	// Read FSPEC
	n, err := r.fspec.Decode(buf)
	if err != nil {
		return bytesRead, fmt.Errorf("decoding FSPEC: %w", err)
	}
	bytesRead += n

	// Clear existing items
	r.items = make(map[string]DataItem)

	// Read items based on FSPEC
	for _, field := range r.uap.Fields() {
		if !r.fspec.GetFRN(field.FRN) {
			continue
		}

		// Check if we have enough bytes for fixed-length items
		if field.Type == Fixed && buf.Len() < int(field.Length) {
			return bytesRead, fmt.Errorf("buffer too short for %s: need %d bytes, have %d",
				field.DataItem, field.Length, buf.Len())
		}

		item, err := r.uap.CreateDataItem(field.DataItem)
		if err != nil {
			if field.Type == Fixed {
				// For fixed length items, we can skip unknown ones
				if buf.Len() < int(field.Length) {
					return bytesRead, fmt.Errorf("buffer too short to skip %s: need %d bytes, have %d",
						field.DataItem, field.Length, buf.Len())
				}
				buf.Next(int(field.Length))
				bytesRead += int(field.Length)
				continue
			}
			return bytesRead, fmt.Errorf("creating %s: %w", field.DataItem, err)
		}

		n, err := item.Decode(buf)
		if err != nil {
			return bytesRead, fmt.Errorf("decoding %s: %w", field.DataItem, err)
		}
		bytesRead += n

		r.items[field.DataItem] = item
	}

	return bytesRead, r.uap.Validate(r.items)
}
