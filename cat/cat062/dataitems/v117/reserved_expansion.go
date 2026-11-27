// cat062/dataitems/v117/reserved_expansion.go
package v117

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// ReservedExpansion implements "RE062", the reserved expansion field. Data holds the
// whole explicit item: the LEN octet, which counts itself, and the LEN-1
// octets that follow it.
type ReservedExpansion struct {
	Data []byte
}

// Decode reads LEN and the LEN-1 octets after it.
func (s *ReservedExpansion) Decode(buf *bytes.Buffer) (int, error) {
	l, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("%w: reading reserved expansion length", asterix.ErrBufferTooShort)
	}
	if l == 0 {
		return 1, fmt.Errorf("%w: reserved expansion LEN is 0 (it counts itself)", asterix.ErrInvalidLength)
	}
	if buf.Len() < int(l)-1 {
		return 1, fmt.Errorf("%w: reserved expansion LEN %d needs %d octets, have %d", asterix.ErrBufferTooShort, l, int(l)-1, buf.Len())
	}
	s.Data = make([]byte, l)
	s.Data[0] = l
	copy(s.Data[1:], buf.Next(int(l)-1))
	return int(l), nil
}

// Encode writes Data. An empty Data encodes as LEN = 1.
func (s *ReservedExpansion) Encode(buf *bytes.Buffer) (int, error) {
	if err := s.Validate(); err != nil {
		return 0, err
	}
	if len(s.Data) == 0 {
		return buf.Write([]byte{1})
	}
	return buf.Write(s.Data)
}

// String returns a string representation.
func (s *ReservedExpansion) String() string {
	if len(s.Data) <= 1 {
		return "ReservedExpansion[empty]"
	}
	return fmt.Sprintf("ReservedExpansion[%d bytes]", len(s.Data)-1)
}

// Validate checks that Data starts with a LEN octet equal to its length.
func (s *ReservedExpansion) Validate() error {
	if len(s.Data) > 255 {
		return fmt.Errorf("%w: reserved expansion longer than 255 octets", asterix.ErrInvalidLength)
	}
	if len(s.Data) > 0 && int(s.Data[0]) != len(s.Data) {
		return fmt.Errorf("%w: reserved expansion LEN %d does not match %d octets", asterix.ErrInvalidLength, s.Data[0], len(s.Data))
	}
	return nil
}
