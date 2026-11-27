// cat062/dataitems/v120/special_purpose.go
package v120

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// SpecialPurpose implements "SP062", the special purpose field. Data holds the
// whole explicit item: the LEN octet, which counts itself, and the LEN-1
// octets that follow it.
type SpecialPurpose struct {
	Data []byte
}

// Decode reads LEN and the LEN-1 octets after it.
func (s *SpecialPurpose) Decode(buf *bytes.Buffer) (int, error) {
	l, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("%w: reading special purpose length", asterix.ErrBufferTooShort)
	}
	if l == 0 {
		return 1, fmt.Errorf("%w: special purpose LEN is 0 (it counts itself)", asterix.ErrInvalidLength)
	}
	if buf.Len() < int(l)-1 {
		return 1, fmt.Errorf("%w: special purpose LEN %d needs %d octets, have %d", asterix.ErrBufferTooShort, l, int(l)-1, buf.Len())
	}
	s.Data = make([]byte, l)
	s.Data[0] = l
	copy(s.Data[1:], buf.Next(int(l)-1))
	return int(l), nil
}

// Encode writes Data. An empty Data encodes as LEN = 1.
func (s *SpecialPurpose) Encode(buf *bytes.Buffer) (int, error) {
	if err := s.Validate(); err != nil {
		return 0, err
	}
	if len(s.Data) == 0 {
		return buf.Write([]byte{1})
	}
	return buf.Write(s.Data)
}

// String returns a string representation.
func (s *SpecialPurpose) String() string {
	if len(s.Data) <= 1 {
		return "SpecialPurpose[empty]"
	}
	return fmt.Sprintf("SpecialPurpose[%d bytes]", len(s.Data)-1)
}

// Validate checks that Data starts with a LEN octet equal to its length.
func (s *SpecialPurpose) Validate() error {
	if len(s.Data) > 255 {
		return fmt.Errorf("%w: special purpose longer than 255 octets", asterix.ErrInvalidLength)
	}
	if len(s.Data) > 0 && int(s.Data[0]) != len(s.Data) {
		return fmt.Errorf("%w: special purpose LEN %d does not match %d octets", asterix.ErrInvalidLength, s.Data[0], len(s.Data))
	}
	return nil
}
