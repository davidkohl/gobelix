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
	data, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("%w: need 1 byte for received power", asterix.ErrBufferTooShort)
	}

	r.Power = int8(data)

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
