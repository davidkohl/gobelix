// asterix/datablock.go
package asterix

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// DataBlock represents a complete ASTERIX message
type DataBlock struct {
	category Category
	records  []*Record
	uap      UAP
}

// NewDataBlock creates a new ASTERIX data block
func NewDataBlock(category Category, uap UAP) (*DataBlock, error) {
	if !category.IsValid() {
		return nil, fmt.Errorf("%w: %d", ErrInvalidCategory, category)
	}
	if uap == nil {
		return nil, fmt.Errorf("%w: UAP cannot be nil", ErrInvalidMessage)
	}
	if uap.Category() != category {
		return nil, fmt.Errorf("%w: UAP category %d does not match block category %d",
			ErrInvalidCategory, uap.Category(), category)
	}

	return &DataBlock{
		category: category,
		records:  make([]*Record, 0),
		uap:      uap,
	}, nil
}

// AddRecord adds a record to the data block
func (db *DataBlock) AddRecord(record *Record) error {
	if record == nil {
		return fmt.Errorf("%w: record cannot be nil", ErrInvalidMessage)
	}
	if record.category != db.category {
		return fmt.Errorf("%w: record category %d does not match block category %d",
			ErrInvalidCategory, record.category, db.category)
	}

	db.records = append(db.records, record)
	return nil
}

// Records returns all records in the data block
func (db *DataBlock) Records() []*Record {
	return db.records
}

// Encode serializes the data block according to ASTERIX specification
func (db *DataBlock) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write category
	if err := buf.WriteByte(byte(db.category)); err != nil {
		return nil, fmt.Errorf("writing category: %w", err)
	}

	// Reserve space for length
	if err := binary.Write(buf, binary.BigEndian, uint16(0)); err != nil {
		return nil, fmt.Errorf("reserving length: %w", err)
	}

	// Encode all records
	for i, record := range db.records {
		_, err := record.Encode(buf)
		if err != nil {
			return nil, fmt.Errorf("encoding record %d: %w", i, err)
		}
	}

	// Update length
	data := buf.Bytes()
	binary.BigEndian.PutUint16(data[1:3], uint16(len(data)))

	return data, nil
}

// Decode parses a complete ASTERIX data block
func (db *DataBlock) Decode(data []byte) error {
	if len(data) < 3 {
		return fmt.Errorf("%w: data too short", ErrInvalidMessage)
	}

	// Verify category
	cat := Category(data[0])
	if cat != db.category {
		return fmt.Errorf("%w: expected %d, got %d",
			ErrInvalidCategory, db.category, cat)
	}

	// Check length
	length := binary.BigEndian.Uint16(data[1:3])
	if int(length) != len(data) {
		return fmt.Errorf("%w: expected %d, got %d",
			ErrInvalidLength, length, len(data))
	}

	// Clear existing records
	db.records = db.records[:0]

	// Read records
	buf := bytes.NewBuffer(data[3:]) // Skip CAT/LEN
	for buf.Len() > 0 {
		// Check if there's enough data for at least an FSPEC byte
		if buf.Len() == 0 {
			break // End of buffer, stop processing
		}

		record, err := NewRecord(db.category, db.uap)
		if err != nil {
			return fmt.Errorf("creating record: %w", err)
		}

		// Try to decode the record
		_, err = record.Decode(buf)
		if err != nil {
			// If we hit EOF while processing the last record, we can ignore it
			if err == io.EOF && buf.Len() == 0 {
				break
			}
			return fmt.Errorf("decoding record: %w", err)
		}

		db.records = append(db.records, record)
	}

	return nil
}

// Clear removes all records from the data block
func (db *DataBlock) Clear() {
	db.records = db.records[:0]
}

// Length returns the number of records in the data block
func (db *DataBlock) Length() int {
	return len(db.records)
}
