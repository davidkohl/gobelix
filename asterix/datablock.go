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
	category  Category  // Category of this data block
	records   []*Record // Records within this data block
	uap       UAP       // User Application Profile
	blockable bool      // Whether this data block supports blocking
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
		category:  category,
		records:   make([]*Record, 0),
		uap:       uap,
		blockable: category.IsBlockable(),
	}, nil
}

// AddRecord adds a record to the data block
func (db *DataBlock) AddRecord(record *Record) error {
	if record == nil {
		return fmt.Errorf("%w: record cannot be nil", ErrInvalidMessage)
	}
	if record.Category() != db.category {
		return fmt.Errorf("%w: record category %d does not match block category %d",
			ErrInvalidCategory, record.Category(), db.category)
	}

	db.records = append(db.records, record)
	return nil
}

// Records returns all records in the data block
func (db *DataBlock) Records() []*Record {
	// Return a copy to prevent modification
	records := make([]*Record, len(db.records))
	copy(records, db.records)
	return records
}

// Encode serializes the data block according to ASTERIX specification
func (db *DataBlock) Encode() ([]byte, error) {
	return db.EncodeWithBuffer(nil)
}

// EncodeWithBuffer serializes the data block using the provided buffer
func (db *DataBlock) EncodeWithBuffer(buf *bytes.Buffer) ([]byte, error) {
	if buf == nil {
		buf = new(bytes.Buffer)
	} else {
		buf.Reset()
	}

	// Write category
	if err := buf.WriteByte(byte(db.category)); err != nil {
		return nil, fmt.Errorf("writing category: %w", err)
	}

	// Reserve space for length (2 bytes)
	if err := binary.Write(buf, binary.BigEndian, uint16(0)); err != nil {
		return nil, fmt.Errorf("reserving length: %w", err)
	}

	// If the category doesn't support blocking or there's only one record,
	// encode as a single record (no blocking)
	if !db.blockable || len(db.records) == 1 {
		for i, record := range db.records {
			_, err := record.Encode(buf)
			if err != nil {
				return nil, fmt.Errorf("encoding record %d: %w", i, err)
			}
		}
	} else {
		// Use blocking for multiple records
		for i, record := range db.records {
			// Pre-encode the record to a temporary buffer
			recordBuf := new(bytes.Buffer)
			_, err := record.Encode(recordBuf)
			if err != nil {
				return nil, fmt.Errorf("encoding record %d: %w", i, err)
			}

			// Write the encoded record to the main buffer
			_, err = buf.Write(recordBuf.Bytes())
			if err != nil {
				return nil, fmt.Errorf("writing record %d: %w", i, err)
			}
		}
	}

	// Update length
	data := buf.Bytes()
	binary.BigEndian.PutUint16(data[1:3], uint16(len(data)))

	return data, nil
}

// Decode parses an ASTERIX data block from bytes
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

// DecodeFrom decodes a data block from a reader
func (db *DataBlock) DecodeFrom(r io.Reader) error {
	// Read category and length (3 bytes)
	header := make([]byte, 3)
	if _, err := io.ReadFull(r, header); err != nil {
		return fmt.Errorf("reading header: %w", err)
	}

	// Verify category
	cat := Category(header[0])
	if cat != db.category {
		return fmt.Errorf("%w: expected %d, got %d",
			ErrInvalidCategory, db.category, cat)
	}

	// Get length
	length := binary.BigEndian.Uint16(header[1:3])
	if length < 3 {
		return fmt.Errorf("%w: length too small (%d)", ErrInvalidLength, length)
	}

	// Read the rest of the message
	data := make([]byte, length)
	copy(data[:3], header)
	if _, err := io.ReadFull(r, data[3:]); err != nil {
		return fmt.Errorf("reading message body: %w", err)
	}

	// Decode the complete message
	return db.Decode(data)
}

// EncodeTo encodes a data block to a writer
func (db *DataBlock) EncodeTo(w io.Writer) error {
	data, err := db.Encode()
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// Category returns the category of this data block
func (db *DataBlock) Category() Category {
	return db.category
}

// UAP returns the UAP used by this data block
func (db *DataBlock) UAP() UAP {
	return db.uap
}

// Blockable returns whether this data block supports blocking
func (db *DataBlock) Blockable() bool {
	return db.blockable
}

// SetBlockable sets whether this data block supports blocking
func (db *DataBlock) SetBlockable(blockable bool) {
	db.blockable = blockable
}

// Clear removes all records from the data block
func (db *DataBlock) Clear() {
	db.records = db.records[:0]
}

// RecordCount returns the number of records in the data block
func (db *DataBlock) RecordCount() int {
	return len(db.records)
}

// EstimateSize estimates the encoded size of this data block in bytes
func (db *DataBlock) EstimateSize() int {
	size := 3 // CAT + LEN fields

	// Add estimated size of each record
	for _, record := range db.records {
		size += record.EstimateSize()
	}

	return size
}

// EncodeRecord creates a new record for this data block, encodes the provided data items into it,
// and adds it to the block (a helper function for common usage)
func (db *DataBlock) EncodeRecord(items map[string]DataItem) error {
	record, err := NewRecord(db.category, db.uap)
	if err != nil {
		return fmt.Errorf("creating record: %w", err)
	}

	// Add each data item to the record
	for id, item := range items {
		if err := record.SetDataItem(id, item); err != nil {
			return fmt.Errorf("setting data item %s: %w", id, err)
		}
	}

	// Add the record to the data block
	return db.AddRecord(record)
}

// IsASRS (All Same Record Structure) checks if all records have the same FSPEC structure
// This can be useful for optimizing encoding/decoding
func (db *DataBlock) IsASRS() bool {
	if len(db.records) <= 1 {
		return true
	}

	firstFSPEC := db.records[0].FSPEC()
	firstSize := firstFSPEC.Size()
	firstData := make([]byte, firstSize)
	firstFSPEC.EncodeToBytes(firstData, 0)

	for i := 1; i < len(db.records); i++ {
		fspec := db.records[i].FSPEC()
		if fspec.Size() != firstSize {
			return false
		}

		data := make([]byte, fspec.Size())
		fspec.EncodeToBytes(data, 0)

		// Compare byte by byte
		for j := 0; j < firstSize; j++ {
			if data[j] != firstData[j] {
				return false
			}
		}
	}

	return true
}

// Clone creates a deep copy of this data block
func (db *DataBlock) Clone() (*DataBlock, error) {
	clone, err := NewDataBlock(db.category, db.uap)
	if err != nil {
		return nil, err
	}

	clone.blockable = db.blockable

	// Clone each record
	for _, record := range db.records {
		recordClone, err := record.Clone()
		if err != nil {
			return nil, fmt.Errorf("cloning record: %w", err)
		}
		if err := clone.AddRecord(recordClone); err != nil {
			return nil, fmt.Errorf("adding cloned record: %w", err)
		}
	}

	return clone, nil
}
