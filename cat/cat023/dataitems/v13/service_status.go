// cat/cat023/dataitems/v13/service_status.go
package v13

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// ServiceStatus represents I023/110 - Service Status (CAT023 ed 1.3
// §5.2.7). One-octet extensible item: spare(8-5) STAT(4-2) FX(1). No
// extension is defined in ed 1.3; extension octets are skipped on decode.
type ServiceStatus struct {
	// STAT is the status of the service; nil encodes as STAT = 0 (unknown).
	STAT *STATValue
}

// STATValue holds STAT (3 bits): 0 unknown, 1 failed, 2 disabled,
// 3 degraded, 4 normal, 5 initialisation.
type STATValue struct {
	STAT uint8
}

// Decode decodes the Service Status from bytes
func (s *ServiceStatus) Decode(buf *bytes.Buffer) (int, error) {
	b, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("%w: reading I023/110", asterix.ErrBufferTooShort)
	}
	n := 1
	s.STAT = &STATValue{STAT: (b >> 1) & 0x07}
	for b&0x01 != 0 {
		if b, err = buf.ReadByte(); err != nil {
			return n, fmt.Errorf("%w: reading I023/110 extension", asterix.ErrBufferTooShort)
		}
		n++
	}
	return n, nil
}

// Encode encodes the Service Status to bytes
func (s *ServiceStatus) Encode(buf *bytes.Buffer) (int, error) {
	if err := s.Validate(); err != nil {
		return 0, err
	}
	var b byte
	if s.STAT != nil {
		b = (s.STAT.STAT & 0x07) << 1
	}
	if err := buf.WriteByte(b); err != nil {
		return 0, fmt.Errorf("writing I023/110: %w", err)
	}
	return 1, nil
}

// Validate validates the Service Status
func (s *ServiceStatus) Validate() error {
	if s.STAT != nil && s.STAT.STAT > 7 {
		return fmt.Errorf("%w: STAT is 3 bits, got %d", asterix.ErrInvalidMessage, s.STAT.STAT)
	}
	return nil
}

// String returns a string representation
func (s *ServiceStatus) String() string {
	if s.STAT == nil {
		return "ServiceStatus{}"
	}

	statDesc := map[uint8]string{
		0: "Unknown",
		1: "Failed",
		2: "Disabled",
		3: "Degraded",
		4: "Normal",
		5: "Initialisation",
	}

	if desc, ok := statDesc[s.STAT.STAT]; ok {
		return fmt.Sprintf("ServiceStatus{STAT:%s}", desc)
	}

	return fmt.Sprintf("ServiceStatus{STAT:Invalid(%d)}", s.STAT.STAT)
}
