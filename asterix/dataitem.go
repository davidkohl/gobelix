// asterix/dataitem.go
package asterix

import "bytes"

// DataItem represents a single ASTERIX data field
type DataItem interface {
	// Encode writes the item to the buffer
	// Returns number of bytes written and any error
	Encode(buf *bytes.Buffer) (int, error)

	// Decode reads the item from the buffer
	// Returns number of bytes read and any error
	Decode(buf *bytes.Buffer) (int, error)

	// Validate checks if the item contains valid data
	Validate() error
}

// ItemType indicates how a data item should be processed
type ItemType uint8

const (
	Fixed ItemType = iota + 1
	Extended
	Repetitive
	Compound
)

// DataField describes a field in the UAP
type DataField struct {
	FRN         uint8  // Field Reference Number
	DataItem    string // e.g., "021/010"
	Description string
	Type        ItemType
	Length      uint8 // For fixed length items
	Mandatory   bool
}
