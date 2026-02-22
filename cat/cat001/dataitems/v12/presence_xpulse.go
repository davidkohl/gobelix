package v12

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// PresenceXPulse represents I001/150 - Presence of X-Pulse
// One-octet fixed length data item
type PresenceXPulse struct {
	XA bool // X-pulse received in Mode-3/A reply
	XC bool // X-pulse received in Mode-C reply
	X2 bool // X-pulse received in Mode-2 reply
}

func (p *PresenceXPulse) Decode(buf *bytes.Buffer) (int, error) {
	data, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("%w: need 1 byte for presence of X-pulse", asterix.ErrBufferTooShort)
	}

	p.XA = (data & 0x80) != 0 // bit 8
	p.XC = (data & 0x20) != 0 // bit 6
	p.X2 = (data & 0x04) != 0 // bit 3
	// bits 7, 5, 4, 2, 1 are spare

	return 1, nil
}

func (p *PresenceXPulse) Encode(buf *bytes.Buffer) (int, error) {
	data := uint8(0)
	if p.XA {
		data |= 0x80 // bit 8
	}
	if p.XC {
		data |= 0x20 // bit 6
	}
	if p.X2 {
		data |= 0x04 // bit 3
	}

	if err := buf.WriteByte(data); err != nil {
		return 0, fmt.Errorf("writing presence of X-pulse: %w", err)
	}
	return 1, nil
}

func (p *PresenceXPulse) Validate() error {
	return nil
}

func (p *PresenceXPulse) String() string {
	result := "X-Pulse:"
	if p.XA {
		result += " Mode-3/A"
	}
	if p.XC {
		result += " Mode-C"
	}
	if p.X2 {
		result += " Mode-2"
	}
	if !p.XA && !p.XC && !p.X2 {
		result += " None"
	}
	return result
}
