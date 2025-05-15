// asterix/dataitem.go
package asterix

import (
	"bytes"
	"io"
)

// DataItem represents a single ASTERIX data field
type DataItem interface {
	// Encode writes the item to the provided buffer
	// Returns number of bytes written and any error
	Encode(buf *bytes.Buffer) (int, error)

	// Decode reads the item from the provided buffer
	// Returns number of bytes read and any error
	Decode(buf *bytes.Buffer) (int, error)

	// Validate checks if the item contains valid data
	// It returns nil if valid, or an error describing the validation issue
	Validate() error
}

// ItemType indicates how a data item should be processed
type ItemType uint8

const (
	Fixed      ItemType = iota + 1 // Fixed length data item
	Extended                       // Extended length data item (with FX bits)
	Explicit                       // Explicit length data item (with length indicator)
	Repetitive                     // Repetitive data item (with count indicator)
	Compound                       // Compound data item (with primary subitem)
)

// String returns a string representation of the item type
func (t ItemType) String() string {
	switch t {
	case Fixed:
		return "Fixed"
	case Extended:
		return "Extended"
	case Explicit:
		return "Explicit"
	case Repetitive:
		return "Repetitive"
	case Compound:
		return "Compound"
	default:
		return "Unknown"
	}
}

// DataField describes a field in the UAP
type DataField struct {
	FRN         uint8    // Field Reference Number
	DataItem    string   // e.g., "021/010"
	Description string   // Human-readable description
	Type        ItemType // Type of data item
	Length      uint8    // For fixed length items
	Mandatory   bool     // Whether the item is mandatory in the record
}

// FixedLengthItem represents a data item with a fixed length
type FixedLengthItem interface {
	DataItem
	FixedLength() int
}

// ExtendedLengthItem represents a data item with a variable length
// determined by FX (extension) bits
type ExtendedLengthItem interface {
	DataItem
	HasExtension() bool
}

// ExplicitLengthItem represents a data item with a length indicator
type ExplicitLengthItem interface {
	DataItem
	LengthIndicator() uint8
}

// RepetitiveItem represents a data item that repeats a variable number of times
type RepetitiveItem interface {
	DataItem
	RepetitionCount() uint8
}

// CompoundItem represents a data item that contains multiple subitems
type CompoundItem interface {
	DataItem
	SubitemCount() int
}

// DataReader provides common methods for reading data items
type DataReader interface {
	ReadByte() (byte, error)
	Read(p []byte) (n int, err error)
	ReadFull(p []byte) (n int, err error)
}

// DataWriter provides common methods for writing data items
type DataWriter interface {
	WriteByte(c byte) error
	Write(p []byte) (n int, err error)
}

// ByteReader is an io.Reader that can also read individual bytes
type ByteReader interface {
	io.Reader
	ReadByte() (byte, error)
}

// ByteWriter is an io.Writer that can also write individual bytes
type ByteWriter interface {
	io.Writer
	WriteByte(c byte) error
}

// ReadFull reads exactly len(p) bytes from r into p
// If r returns an EOF before reading len(p) bytes, ReadFull returns an error
func ReadFull(r ByteReader, p []byte) (n int, err error) {
	return io.ReadFull(r, p)
}

// ReadUint16 reads a 16-bit unsigned integer from the reader in big-endian order
func ReadUint16(r ByteReader) (uint16, error) {
	b1, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	b2, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	return uint16(b1)<<8 | uint16(b2), nil
}

// ReadUint24 reads a 24-bit unsigned integer from the reader in big-endian order
func ReadUint24(r ByteReader) (uint32, error) {
	b1, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	b2, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	b3, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	return uint32(b1)<<16 | uint32(b2)<<8 | uint32(b3), nil
}

// ReadUint32 reads a 32-bit unsigned integer from the reader in big-endian order
func ReadUint32(r ByteReader) (uint32, error) {
	b1, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	b2, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	b3, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	b4, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	return uint32(b1)<<24 | uint32(b2)<<16 | uint32(b3)<<8 | uint32(b4), nil
}

// WriteUint16 writes a 16-bit unsigned integer to the writer in big-endian order
func WriteUint16(w ByteWriter, v uint16) error {
	if err := w.WriteByte(byte(v >> 8)); err != nil {
		return err
	}
	return w.WriteByte(byte(v))
}

// WriteUint24 writes a 24-bit unsigned integer to the writer in big-endian order
func WriteUint24(w ByteWriter, v uint32) error {
	if v > 0xFFFFFF {
		v &= 0xFFFFFF // Truncate to 24 bits
	}
	if err := w.WriteByte(byte(v >> 16)); err != nil {
		return err
	}
	if err := w.WriteByte(byte(v >> 8)); err != nil {
		return err
	}
	return w.WriteByte(byte(v))
}

// WriteUint32 writes a 32-bit unsigned integer to the writer in big-endian order
func WriteUint32(w ByteWriter, v uint32) error {
	if err := w.WriteByte(byte(v >> 24)); err != nil {
		return err
	}
	if err := w.WriteByte(byte(v >> 16)); err != nil {
		return err
	}
	if err := w.WriteByte(byte(v >> 8)); err != nil {
		return err
	}
	return w.WriteByte(byte(v))
}

// BaseFixedLengthItem provides a base implementation for fixed length data items
type BaseFixedLengthItem struct {
	length int // Length in bytes
}

// FixedLength returns the fixed length of the item
func (b *BaseFixedLengthItem) FixedLength() int {
	return b.length
}

// BaseExtendedLengthItem provides a base implementation for extended length data items
type BaseExtendedLengthItem struct {
	hasExtension bool // Whether there are extensions
}

// HasExtension returns true if the item has extensions
func (b *BaseExtendedLengthItem) HasExtension() bool {
	return b.hasExtension
}

// BaseExplicitLengthItem provides a base implementation for explicit length data items
type BaseExplicitLengthItem struct {
	length uint8 // Length indicator value
}

// LengthIndicator returns the length indicator value
func (b *BaseExplicitLengthItem) LengthIndicator() uint8 {
	return b.length
}

// BaseRepetitiveItem provides a base implementation for repetitive data items
type BaseRepetitiveItem struct {
	count uint8 // Repetition count
}

// RepetitionCount returns the repetition count
func (b *BaseRepetitiveItem) RepetitionCount() uint8 {
	return b.count
}

// BaseCompoundItem provides a base implementation for compound data items
type BaseCompoundItem struct {
	subitemCount int // Number of subitems
}

// SubitemCount returns the number of subitems
func (b *BaseCompoundItem) SubitemCount() int {
	return b.subitemCount
}
