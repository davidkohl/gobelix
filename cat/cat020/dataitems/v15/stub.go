// cat/cat020/dataitems/v15/stub.go
package v15

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// FixedStub is a stub for fixed-length data items
type FixedStub struct {
	ID     string
	Length int
	Data   []byte
}

// NewFixedStub creates a new fixed-length stub
func NewFixedStub(id string, length int) *FixedStub {
	return &FixedStub{
		ID:     id,
		Length: length,
	}
}

// Decode reads the fixed number of bytes
func (s *FixedStub) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, s.Length)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading stub %s: %w", s.ID, err)
	}
	if n != s.Length {
		return n, fmt.Errorf("stub %s: expected %d bytes, got %d", s.ID, s.Length, n)
	}

	s.Data = data
	return n, nil
}

// Encode writes the raw bytes
func (s *FixedStub) Encode(buf *bytes.Buffer) (int, error) {
	if len(s.Data) == 0 {
		return 0, fmt.Errorf("stub %s has no data to encode", s.ID)
	}

	n, err := buf.Write(s.Data)
	if err != nil {
		return n, fmt.Errorf("writing stub %s: %w", s.ID, err)
	}
	return n, nil
}

// Validate implements the DataItem interface
func (s *FixedStub) Validate() error {
	return nil
}

// String returns a string representation
func (s *FixedStub) String() string {
	return fmt.Sprintf("%s (stub, %d bytes)", s.ID, len(s.Data))
}

// ExtendedStub is a stub for extended data items (with FX bit)
type ExtendedStub struct {
	ID   string
	Data []byte
}

// NewExtendedStub creates a new extended stub
func NewExtendedStub(id string) *ExtendedStub {
	return &ExtendedStub{ID: id}
}

// Decode reads octets until FX bit is 0
func (s *ExtendedStub) Decode(buf *bytes.Buffer) (int, error) {
	var data []byte
	bytesRead := 0

	for {
		b, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading extended stub %s: %w", s.ID, err)
		}
		data = append(data, b)
		bytesRead++

		// Check FX bit (LSB, bit 0)
		if b&0x01 == 0 {
			break
		}
	}

	s.Data = data
	return bytesRead, nil
}

// Encode writes the raw bytes
func (s *ExtendedStub) Encode(buf *bytes.Buffer) (int, error) {
	if len(s.Data) == 0 {
		return 0, fmt.Errorf("extended stub %s has no data to encode", s.ID)
	}

	n, err := buf.Write(s.Data)
	if err != nil {
		return n, fmt.Errorf("writing extended stub %s: %w", s.ID, err)
	}
	return n, nil
}

// Validate implements the DataItem interface
func (s *ExtendedStub) Validate() error {
	return nil
}

// String returns a string representation
func (s *ExtendedStub) String() string {
	return fmt.Sprintf("%s (extended stub, %d bytes)", s.ID, len(s.Data))
}

// RepetitiveStub is a stub for repetitive data items
type RepetitiveStub struct {
	ID          string
	ItemLength  int
	Data        []byte
	Repetitions uint8
}

// NewRepetitiveStub creates a new repetitive stub
func NewRepetitiveStub(id string, itemLength int) *RepetitiveStub {
	return &RepetitiveStub{
		ID:         id,
		ItemLength: itemLength,
	}
}

// Decode reads REP + repetitions
func (s *RepetitiveStub) Decode(buf *bytes.Buffer) (int, error) {
	// Read REP byte
	rep, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading repetitive stub %s REP: %w", s.ID, err)
	}
	s.Repetitions = rep
	bytesRead := 1

	// Read repetitions
	dataLen := int(rep) * s.ItemLength
	data := make([]byte, dataLen)
	n, err := buf.Read(data)
	if err != nil {
		return bytesRead + n, fmt.Errorf("reading repetitive stub %s data: %w", s.ID, err)
	}
	if n != dataLen {
		return bytesRead + n, fmt.Errorf("repetitive stub %s: expected %d bytes, got %d", s.ID, dataLen, n)
	}

	s.Data = data
	return bytesRead + n, nil
}

// Encode writes REP + raw bytes
func (s *RepetitiveStub) Encode(buf *bytes.Buffer) (int, error) {
	// Write REP byte
	err := buf.WriteByte(s.Repetitions)
	if err != nil {
		return 0, fmt.Errorf("writing repetitive stub %s REP: %w", s.ID, err)
	}
	bytesWritten := 1

	// Write data
	n, err := buf.Write(s.Data)
	if err != nil {
		return bytesWritten + n, fmt.Errorf("writing repetitive stub %s data: %w", s.ID, err)
	}
	return bytesWritten + n, nil
}

// Validate implements the DataItem interface
func (s *RepetitiveStub) Validate() error {
	return nil
}

// String returns a string representation
func (s *RepetitiveStub) String() string {
	return fmt.Sprintf("%s (repetitive stub, %d reps, %d bytes)", s.ID, s.Repetitions, len(s.Data))
}

// CompoundStub is a stub for compound data items
type CompoundStub struct {
	ID   string
	Data []byte
}

// NewCompoundStub creates a new compound stub
func NewCompoundStub(id string) *CompoundStub {
	return &CompoundStub{ID: id}
}

// Decode reads primary subfield + subfields (simplified - just reads 1 byte for now)
func (s *CompoundStub) Decode(buf *bytes.Buffer) (int, error) {
	// For now, just read 1 byte (primary subfield)
	// Real compound items need to parse the primary subfield to know which subfields follow
	b, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading compound stub %s: %w", s.ID, err)
	}
	s.Data = []byte{b}
	return 1, nil
}

// Encode writes the raw bytes
func (s *CompoundStub) Encode(buf *bytes.Buffer) (int, error) {
	if len(s.Data) == 0 {
		return 0, fmt.Errorf("compound stub %s has no data to encode", s.ID)
	}

	n, err := buf.Write(s.Data)
	if err != nil {
		return n, fmt.Errorf("writing compound stub %s: %w", s.ID, err)
	}
	return n, nil
}

// Validate implements the DataItem interface
func (s *CompoundStub) Validate() error {
	return nil
}

// String returns a string representation
func (s *CompoundStub) String() string {
	return fmt.Sprintf("%s (compound stub, %d bytes)", s.ID, len(s.Data))
}

// ExplicitStub is a stub for explicit-length data items (RE/SP)
type ExplicitStub struct {
	ID   string
	Data []byte
}

// NewExplicitStub creates a new explicit stub
func NewExplicitStub(id string) *ExplicitStub {
	return &ExplicitStub{ID: id}
}

// Decode reads length byte + data
// Note: The length byte value INCLUDES the length byte itself per ASTERIX spec
func (s *ExplicitStub) Decode(buf *bytes.Buffer) (int, error) {
	// Read length byte
	length, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading explicit stub %s length: %w", s.ID, err)
	}

	if length == 0 {
		return 1, fmt.Errorf("explicit stub %s: invalid length 0", s.ID)
	}

	// Length includes the length byte itself, so data is (length - 1) bytes
	dataLen := int(length) - 1
	data := make([]byte, dataLen)
	n, err := buf.Read(data)
	if err != nil {
		return 1 + n, fmt.Errorf("reading explicit stub %s data: %w", s.ID, err)
	}
	if n != dataLen {
		return 1 + n, fmt.Errorf("explicit stub %s: expected %d bytes, got %d", s.ID, dataLen, n)
	}

	s.Data = data
	return int(length), nil // Return total length including length byte
}

// Encode writes length byte + raw bytes
// Note: The length byte value INCLUDES the length byte itself per ASTERIX spec
func (s *ExplicitStub) Encode(buf *bytes.Buffer) (int, error) {
	totalLen := len(s.Data) + 1 // +1 for length byte itself
	if totalLen > 255 {
		return 0, fmt.Errorf("explicit stub %s: total length too long (%d bytes)", s.ID, totalLen)
	}

	// Write length byte (includes itself)
	err := buf.WriteByte(byte(totalLen))
	if err != nil {
		return 0, fmt.Errorf("writing explicit stub %s length: %w", s.ID, err)
	}

	// Write data
	n, err := buf.Write(s.Data)
	if err != nil {
		return 1 + n, fmt.Errorf("writing explicit stub %s data: %w", s.ID, err)
	}
	return 1 + n, nil
}

// Validate implements the DataItem interface
func (s *ExplicitStub) Validate() error {
	return nil
}

// String returns a string representation
func (s *ExplicitStub) String() string {
	return fmt.Sprintf("%s (explicit stub, %d bytes)", s.ID, len(s.Data))
}

// Stub is an alias for FixedStub for backward compatibility
type Stub = FixedStub

// Helper function to check type satisfaction (will fail at compile time if not satisfied)
var (
	_ asterix.DataItem = (*FixedStub)(nil)
	_ asterix.DataItem = (*ExtendedStub)(nil)
	_ asterix.DataItem = (*RepetitiveStub)(nil)
	_ asterix.DataItem = (*CompoundStub)(nil)
	_ asterix.DataItem = (*ExplicitStub)(nil)
)
