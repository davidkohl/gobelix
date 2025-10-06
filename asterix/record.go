// asterix/record.go
package asterix

import (
	"bytes"
	"fmt"
	"io"
)

// Record represents a single ASTERIX record.
//
// Thread Safety: Record is NOT safe for concurrent use.
// Each Record instance should be accessed by only one goroutine at a time,
// or protected by external synchronization.
// Methods that modify the Record (SetDataItem, Reset) should not be called
// concurrently with any other methods.
type Record struct {
	category Category            // Category of this record
	fspec    *FSPEC              // Field Specification
	items    map[string]DataItem // Data items indexed by their reference (e.g., "I021/010")
	uap      UAP                 // User Application Profile defining the structure
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
		return WrapError(err, "validating %s", id)
	}

	r.items[id] = item
	return r.fspec.SetFRN(frn)
}

// GetDataItem retrieves a data item by its ID
func (r *Record) GetDataItem(id string) (DataItem, bool) {
	item, exists := r.items[id]
	return item, exists
}

// Encode writes the record to a buffer
func (r *Record) Encode(buf *bytes.Buffer) (int, error) {
	// Validate the record against UAP rules
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
			return bytesWritten, NewEncodingError(
				r.category,
				field.DataItem,
				fmt.Sprintf("FRN %d marked in FSPEC but item not present", field.FRN),
				ErrInvalidMessage,
			)
		}

		n, err := item.Encode(buf)
		if err != nil {
			return bytesWritten, NewEncodingError(
				r.category,
				field.DataItem,
				"failed to encode item",
				err,
			)
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
		return bytesRead, NewDecodeError(
			r.category,
			"",
			"decoding FSPEC",
			err,
		)
	}
	bytesRead += n

	// Clear existing items
	r.items = make(map[string]DataItem)

	// Read items based on FSPEC
	for _, field := range r.uap.Fields() {
		if !r.fspec.GetFRN(field.FRN) {
			continue
		}

		// Create data item
		item, err := r.uap.CreateDataItem(field.DataItem)
		if err != nil {
			// For fixed length items, skip over the bytes if we can't create the item
			// This is intentional for forward compatibility - unknown fields are silently skipped
			// to allow newer protocol versions to be partially parsed by older decoders.
			//
			// NOTE: In production, you may want to log this at DEBUG level:
			//   logger.Debug("Skipping unknown field", "field", field.DataItem, "bytes", field.Length)
			if field.Type == Fixed {
				if buf.Len() < int(field.Length) {
					return bytesRead, NewDecodeError(
						r.category,
						field.DataItem,
						fmt.Sprintf("buffer too short to skip field: need %d bytes, have %d", field.Length, buf.Len()),
						ErrBufferTooShort,
					)
				}
				buf.Next(int(field.Length))
				bytesRead += int(field.Length)
				continue
			}
			return bytesRead, NewDecodeError(
				r.category,
				field.DataItem,
				"creating data item",
				err,
			)
		}

		// Decode the item
		n, err := item.Decode(buf)
		if err != nil {
			return bytesRead, NewDecodeError(
				r.category,
				field.DataItem,
				"decoding data item",
				err,
			).WithPosition(bytesRead, buf.Len()+bytesRead)
		}
		bytesRead += n

		// Store the item
		r.items[field.DataItem] = item
	}

	// Validate the record
	if err := r.uap.Validate(r.items); err != nil {
		return bytesRead, err
	}

	return bytesRead, nil
}

// Category returns the category of this record
func (r *Record) Category() Category {
	return r.category
}

// UAP returns the UAP used by this record
func (r *Record) UAP() UAP {
	return r.uap
}

// FSPEC returns the field specification of this record
func (r *Record) FSPEC() *FSPEC {
	return r.fspec
}

// Items returns a map of all data items in this record
// WARNING: The returned map contains pointers to the original data items.
// Modifying the items will affect the record. This is intentional for performance.
// If you need a deep copy, use Clone() instead.
func (r *Record) Items() map[string]DataItem {
	return r.items
}

// ItemCount returns the number of data items in this record
func (r *Record) ItemCount() int {
	return len(r.items)
}

// HasDataItem checks if a specific data item is present
func (r *Record) HasDataItem(id string) bool {
	_, exists := r.items[id]
	return exists
}

// EstimateSize estimates the encoded size of this record in bytes
func (r *Record) EstimateSize() int {
	size := r.fspec.Size() // FSPEC size

	// Add size of each data item
	for _, field := range r.uap.Fields() {
		if !r.fspec.GetFRN(field.FRN) {
			continue
		}

		if field.Type == Fixed {
			size += int(field.Length)
			continue
		}

		// For non-fixed items, we need a rough estimate
		// This is just an estimate, the actual size may differ
		size += 4 // Reasonable default size
	}

	return size
}

// Clone creates a deep copy of this record
//
// PERFORMANCE NOTE: This uses encode/decode for cloning, which is slower than
// field-by-field copy but guarantees correctness. To optimize, each DataItem
// implementation would need its own Clone() method. For most use cases, the
// current implementation is sufficient. If you need high-performance cloning,
// consider caching decoded records or using copy-on-write patterns.
func (r *Record) Clone() (*Record, error) {
	newRecord, err := NewRecord(r.category, r.uap)
	if err != nil {
		return nil, err
	}

	// Encode to buffer and decode to create a deep copy
	// This ensures all data items are properly deep copied
	buf := new(bytes.Buffer)
	_, err = r.Encode(buf)
	if err != nil {
		return nil, err
	}

	_, err = newRecord.Decode(buf)
	if err != nil {
		return nil, err
	}

	return newRecord, nil
}

// Reset clears all data items but keeps the category and UAP
func (r *Record) Reset() {
	r.fspec = NewFSPEC()
	r.items = make(map[string]DataItem)
}

// Validate checks if the record is valid according to the UAP
func (r *Record) Validate() error {
	return r.uap.Validate(r.items)
}
