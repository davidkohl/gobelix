// dataitems/cat021/air_speed.go
package v26

import (
	"bytes"
	"fmt"
	"math"
)

// AirSpeed implements I021/150
type AirSpeed struct {
	Id     string
	IsMach bool    // True if Mach number, false if IAS
	Speed  float64 // Speed in NM/s if IAS, or Mach if Mach number
}

func (a *AirSpeed) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading air speed: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("insufficient data for air speed: got %d bytes, want 2", n)
	}

	a.IsMach = (data[0] & 0x80) != 0
	raw := uint16(data[0]&0x7F)<<8 | uint16(data[1])

	if a.IsMach {
		a.Speed = float64(raw) * 0.001 // LSB = 0.001 Mach
	} else {
		// Convert to knots directly
		a.Speed = float64(raw) * math.Pow(2, -14) * 3600 // LSB = 2^-14 NM/s * 3600 = knots
	}

	return n, a.Validate()
}

func (a *AirSpeed) Encode(buf *bytes.Buffer) (int, error) {
	if err := a.Validate(); err != nil {
		return 0, err
	}

	var raw uint16
	if a.IsMach {
		raw = uint16(math.Round(a.Speed * 1000)) // Convert Mach to raw value
	} else {
		// Convert from knots to raw value
		raw = uint16(math.Round((a.Speed / 3600) * math.Pow(2, 14))) // Convert knots to raw value
	}

	data := make([]byte, 2)
	if a.IsMach {
		data[0] |= 0x80
	}
	data[0] |= byte(raw >> 8)
	data[1] = byte(raw)

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing air speed: %w", err)
	}
	return n, nil
}

func (a *AirSpeed) Validate() error {
	if a.IsMach {
		if a.Speed < 0 || a.Speed > 4.096 {
			return fmt.Errorf("mach number out of valid range [0,4.096]: %f", a.Speed)
		}
	} else {
		if a.Speed < 0 || a.Speed >= 7200 {
			return fmt.Errorf("IAS out of valid range [0,7200) knots: %f", a.Speed)
		}
	}
	return nil
}

func (a *AirSpeed) String() string {
	if a.IsMach {
		return fmt.Sprintf("%.3f Mach", a.Speed)
	}
	return fmt.Sprintf("%.1f kts", a.Speed*3600) // Convert NM/s to knots
}
