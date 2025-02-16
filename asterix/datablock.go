// asterix/datablock.go
package asterix

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// DataBlock represents a complete ASTERIX message
type DataBlock struct {
	category Category
	records  []*Record
}

// NewDataBlock creates a new ASTERIX data block
func NewDataBlock(category Category) (*DataBlock, error) {
	if !category.IsValid() {
		return nil, fmt.Errorf("%w: %d", ErrInvalidCategory, category)
	}

	return &DataBlock{
		category: category,
		records:  make([]*Record, 0),
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

	// Write category (1 octet)
	if err := buf.WriteByte(byte(db.category)); err != nil {
		return nil, fmt.Errorf("writing category: %w", err)
	}
	bytesWritten := 1

	// Reserve space for length (2 octets)
	if err := binary.Write(buf, binary.BigEndian, uint16(0)); err != nil {
		return nil, fmt.Errorf("reserving length: %w", err)
	}
	bytesWritten += 2

	// Encode all records
	for i, record := range db.records {
		n, err := record.Encode(buf)
		if err != nil {
			return nil, fmt.Errorf("encoding record %d: %w", i, err)
		}
		bytesWritten += n
	}

	// Update length field
	data := buf.Bytes()
	binary.BigEndian.PutUint16(data[1:3], uint16(bytesWritten))

	return data, nil
}

func (db *DataBlock) Decode(data []byte) error {
	if len(data) < 3 {
		return fmt.Errorf("%w: data too short", ErrInvalidMessage)
	}

	buf := bytes.NewBuffer(data)
	bytesRead := 0

	// Read category
	cat, err := buf.ReadByte()
	if err != nil {
		return fmt.Errorf("reading category: %w", err)
	}
	bytesRead++

	if Category(cat) != db.category {
		return fmt.Errorf("%w: expected %d, got %d",
			ErrInvalidCategory, db.category, cat)
	}

	// Read length
	var length uint16
	if err := binary.Read(buf, binary.BigEndian, &length); err != nil {
		return fmt.Errorf("reading length: %w", err)
	}
	bytesRead += 2

	if int(length) != len(data) {
		return fmt.Errorf("%w: expected %d, got %d",
			ErrInvalidLength, length, len(data))
	}

	// Clear existing records
	db.records = db.records[:0]

	// Read records until all bytes are consumed
	for bytesRead < int(length) {
		record, err := NewRecord(db.category, db.records[0].uap)
		if err != nil {
			return fmt.Errorf("creating record at byte %d: %w", bytesRead, err)
		}

		n, err := record.Decode(buf)
		if err != nil {
			return fmt.Errorf("decoding record at byte %d: %w", bytesRead, err)
		}
		bytesRead += n

		db.records = append(db.records, record)
	}

	return nil
}
