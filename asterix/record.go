// asterix/record.go
package asterix

import (
	"bytes"
	"fmt"
)

// Record represents a single ASTERIX record containing a collection
// of data items ordered by their Field Reference Numbers (FRN)
type Record struct {
	fspec    *FSPEC
	category Category
	items    map[string]DataItem
	uap      UAP
}

// NewRecord creates a new record for a specific category and UAP
func NewRecord(category Category, uap UAP) (*Record, error) {
	if !category.IsValid() {
		return nil, fmt.Errorf("%w: %d", ErrInvalidCategory, category)
	}
	if uap == nil {
		return nil, fmt.Errorf("%w: UAP cannot be nil", ErrInvalidMessage)
	}
	if uap.Category() != category {
		return nil, fmt.Errorf("%w: UAP category %d does not match record category %d",
			ErrInvalidMessage, uap.Category(), category)
	}

	return &Record{
		fspec:    NewFSPEC(),
		category: category,
		items:    make(map[string]DataItem),
		uap:      uap,
	}, nil
}

// SetDataItem adds or updates a data item in the record
func (r *Record) SetDataItem(id string, item DataItem) error {
	if item == nil {
		return fmt.Errorf("%w: data item cannot be nil", ErrInvalidMessage)
	}

	frn := r.uap.FRNByID(id)
	if frn == 0 {
		return fmt.Errorf("%w: %s", ErrUnknownDataItem, id)
	}

	if err := item.Validate(); err != nil {
		return fmt.Errorf("validating %s: %w", id, err)
	}

	r.items[id] = item
	r.fspec.SetFRN(frn)
	return nil
}

// GetDataItem retrieves a data item by its ID
func (r *Record) GetDataItem(id string) (DataItem, bool) {
	item, exists := r.items[id]
	return item, exists
}

// Encode writes the record to the buffer according to ASTERIX specification
func (r *Record) Encode(buf *bytes.Buffer) (int, error) {
	if err := r.validateMandatoryFields(); err != nil {
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

func (r *Record) Decode(buf *bytes.Buffer) (int, error) {
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

		item, err := r.uap.CreateDataItem(field.DataItem)
		if err != nil {
			return bytesRead, fmt.Errorf("creating %s: %w", field.DataItem, err)
		}

		n, err := item.Decode(buf)
		if err != nil {
			return bytesRead, fmt.Errorf("decoding %s at byte %d: %w",
				field.DataItem, bytesRead, err)
		}
		bytesRead += n

		r.items[field.DataItem] = item
	}

	return bytesRead, nil
}

// validateMandatoryFields ensures all mandatory fields are present
func (r *Record) validateMandatoryFields() error {
	for _, field := range r.uap.Fields() {
		if !field.Mandatory {
			continue
		}
		if _, exists := r.items[field.DataItem]; !exists {
			return fmt.Errorf("%w: %s", ErrMandatoryField, field.DataItem)
		}
	}
	return nil
}
