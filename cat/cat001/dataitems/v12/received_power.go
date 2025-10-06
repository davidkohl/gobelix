package v12

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// ReceivedPower represents I001/131 - Received Power
type ReceivedPower struct {
	Power int8 // Received power in dBm (signed 8-bit)
}

func (r *ReceivedPower) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 1 {
		return 0, fmt.Errorf("%w: need 1 byte for received power, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	data := buf.Next(1)
	r.Power = int8(data[0])

	return 1, nil
}

func (r *ReceivedPower) Encode(buf *bytes.Buffer) (int, error) {
	buf.WriteByte(byte(r.Power))
	return 1, nil
}

func (r *ReceivedPower) String() string {
	return fmt.Sprintf("%d dBm", r.Power)
}

func (r *ReceivedPower) Validate() error {
	return nil
}
