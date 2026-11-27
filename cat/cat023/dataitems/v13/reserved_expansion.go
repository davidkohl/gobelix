// cat/cat023/dataitems/v13/reserved_expansion.go
package v13

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// ReservedExpansion represents RE023 - Reserved Expansion Field
// Variable length field reserved for future use
type ReservedExpansion struct {
	Data []byte
}

// Decode decodes the field. LEN (the first octet) counts itself; Data
// holds LEN and the LEN-1 payload octets.
func (r *ReservedExpansion) Decode(buf *bytes.Buffer) (int, error) {
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
	r.Data = make([]byte, l)
	r.Data[0] = l
	copy(r.Data[1:], buf.Next(int(l)-1))
	return int(l), nil
}

// Encode encodes the field. An empty Data encodes as LEN = 1.
func (r *ReservedExpansion) Encode(buf *bytes.Buffer) (int, error) {
	if err := r.Validate(); err != nil {
		return 0, err
	}
	if len(r.Data) == 0 {
		return buf.Write([]byte{1})
	}
	return buf.Write(r.Data)
}

// Validate checks that Data starts with a LEN octet equal to its length.
func (r *ReservedExpansion) Validate() error {
	if len(r.Data) > 0 && int(r.Data[0]) != len(r.Data) {
		return fmt.Errorf("%w: reserved expansion LEN %d does not match %d octets", asterix.ErrInvalidLength, r.Data[0], len(r.Data))
	}
	if len(r.Data) > 255 {
		return fmt.Errorf("%w: reserved expansion longer than 255 octets", asterix.ErrInvalidLength)
	}
	return nil
}

// String returns a string representation
func (r *ReservedExpansion) String() string {
	if len(r.Data) <= 1 {
		return "ReservedExpansion[empty]"
	}
	return fmt.Sprintf("ReservedExpansion[%d bytes]", len(r.Data)-1)
}
