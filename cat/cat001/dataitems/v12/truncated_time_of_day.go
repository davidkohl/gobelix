package v12

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// TruncatedTimeOfDay represents I001/141 - Truncated Time of Day
// 2 bytes, LSB = 1/128 second
type TruncatedTimeOfDay struct {
	Time float64 // Seconds (LSB = 1/128 s)
}

func (t *TruncatedTimeOfDay) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil || n != 2 {
		return n, fmt.Errorf("%w: need 2 bytes for truncated time of day, have %d", asterix.ErrBufferTooShort, n)
	}

	// 16-bit unsigned value, LSB = 1/128 second
	raw := uint16(data[0])<<8 | uint16(data[1])
	t.Time = float64(raw) / 128.0

	return 2, nil
}

func (t *TruncatedTimeOfDay) Encode(buf *bytes.Buffer) (int, error) {
	// Convert to 1/128 second units
	raw := uint16(t.Time * 128.0)

	buf.WriteByte(byte(raw >> 8))
	buf.WriteByte(byte(raw & 0xFF))

	return 2, nil
}

func (t *TruncatedTimeOfDay) String() string {
	return fmt.Sprintf("%.3fs", t.Time)
}

func (t *TruncatedTimeOfDay) Validate() error {
	return nil
}
