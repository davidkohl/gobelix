// dataitems/cat021/time_reception_position.go
package v26

import (
	"bytes"
	"fmt"
	"math"

	"github.com/davidkohl/gobelix/asterix"
)

// TimeOfMessageReceptionPosition implements I021/073
// Time of reception of the latest position squitter in the Ground Station,
// in the form of elapsed time since last midnight, expressed as UTC.
type TimeOfMessageReceptionPosition struct {
	Time float64 // Time in seconds since midnight
}

func (t *TimeOfMessageReceptionPosition) Decode(buf *bytes.Buffer) (int, error) {
	var data [3]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading time of message reception position: %w", err)
	}
	if n != 3 {
		return n, fmt.Errorf("%w: need 3 bytes for time of message reception position, have %d", asterix.ErrBufferTooShort, n)
	}

	counts := uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2])
	t.Time = float64(counts) / 128.0 // LSB = 1/128 seconds

	return n, t.Validate()
}

func (t *TimeOfMessageReceptionPosition) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	counts := uint32(math.Round(t.Time * 128.0))

	var b [3]byte
	b[0] = byte(counts >> 16)
	b[1] = byte(counts >> 8)
	b[2] = byte(counts)

	n, err := buf.Write(b[:])
	if err != nil {
		return n, fmt.Errorf("writing time of message reception position: %w", err)
	}
	return n, nil
}

func (t *TimeOfMessageReceptionPosition) Validate() error {
	if t.Time < 0 || t.Time >= 86400 {
		return fmt.Errorf("time out of valid range [0,86400): %f", t.Time)
	}
	return nil
}

func (t *TimeOfMessageReceptionPosition) String() string {
	return fmt.Sprintf("%s (%.3fs since midnight)", secondsToTimeString(t.Time), t.Time)
}

// secondsToTimeString converts seconds since midnight to HH:MM:SS.mmm format
func secondsToTimeString(seconds float64) string {
	totalSeconds := int(seconds)
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	secs := totalSeconds % 60
	millis := int((seconds - float64(totalSeconds)) * 1000)
	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, secs, millis)
}
