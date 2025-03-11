// dataitems/cat021/time_reception_position_high.go
package v26

import (
	"bytes"
	"fmt"
	"math"
)

// FSIType represents the Full Second Indication values
type FSIType uint8

const (
	FSISameSecond    FSIType = 0 // TOMRp whole seconds = (I021/073) Whole seconds
	FSIOneSecondMore FSIType = 1 // TOMRp whole seconds = (I021/073) Whole seconds + 1
	FSIOneSecondLess FSIType = 2 // TOMRp whole seconds = (I021/073) Whole seconds - 1
	FSIReserved      FSIType = 3 // Reserved
)

// TimeOfMessageReceptionPositionHigh implements I021/074
// High precision variant showing fraction of the second for position reception time
type TimeOfMessageReceptionPositionHigh struct {
	FSI            FSIType // Full Second Indication
	FractionalTime float64 // Fractional part of the time of message reception
}

func (t *TimeOfMessageReceptionPositionHigh) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 4)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading high precision position time: %w", err)
	}
	if n != 4 {
		return n, fmt.Errorf("insufficient data: got %d bytes, want 4", n)
	}

	t.FSI = FSIType((data[0] >> 6) & 0x03)
	counts := uint32(data[0]&0x3F)<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])
	t.FractionalTime = float64(counts) / float64(1<<30) // LSB = 2^-30 seconds

	return n, t.Validate()
}

func (t *TimeOfMessageReceptionPositionHigh) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	counts := uint32(math.Round(t.FractionalTime * float64(1<<30)))

	b := make([]byte, 4)
	b[0] = byte(uint8(t.FSI)<<6) | byte(counts>>24)
	b[1] = byte(counts >> 16)
	b[2] = byte(counts >> 8)
	b[3] = byte(counts)

	n, err := buf.Write(b)
	if err != nil {
		return n, fmt.Errorf("writing high precision position time: %w", err)
	}
	return n, nil
}

func (t *TimeOfMessageReceptionPositionHigh) Validate() error {
	if t.FSI > FSIReserved {
		return fmt.Errorf("invalid FSI value: %d", t.FSI)
	}
	if t.FractionalTime < 0 || t.FractionalTime >= 1 {
		return fmt.Errorf("fractional time out of valid range [0,1): %f", t.FractionalTime)
	}
	return nil
}

func (t *TimeOfMessageReceptionPositionHigh) String() string {
	return fmt.Sprintf("FSI: %v - Fraction: %v", t.FSI, t.FractionalTime)
}
