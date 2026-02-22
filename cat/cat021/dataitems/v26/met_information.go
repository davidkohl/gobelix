// dataitems/cat021/met_information.go
package v26

import (
	"bytes"
	"fmt"
	"math"
	"strings"

	"github.com/davidkohl/gobelix/asterix"
)

// MetInformation implements I021/220
// Met Information (Compound)
type MetInformation struct {
	// Subfield #1: Wind Speed
	WindSpeed *uint16 // Wind speed in knots, LSB = 1 knot

	// Subfield #2: Wind Direction
	WindDirection *uint16 // Wind direction in degrees, LSB = 1°

	// Subfield #3: Temperature
	Temperature *float64 // Temperature in Celsius, LSB = 0.25°C

	// Subfield #4: Turbulence
	Turbulence *uint8 // Turbulence level (0-3)
}

func (m *MetInformation) Encode(buf *bytes.Buffer) (int, error) {
	if err := m.Validate(); err != nil {
		return 0, err
	}

	bytesWritten := 0

	// Determine which subfields are present
	var primaryField uint8
	if m.WindSpeed != nil {
		primaryField |= 0x80 // bit 8
	}
	if m.WindDirection != nil {
		primaryField |= 0x40 // bit 7
	}
	if m.Temperature != nil {
		primaryField |= 0x20 // bit 6
	}
	if m.Turbulence != nil {
		primaryField |= 0x10 // bit 5
	}
	// Bits 4-2 are spare
	// Bit 1 is FX (no extension)

	// Write primary field
	if err := buf.WriteByte(primaryField); err != nil {
		return bytesWritten, fmt.Errorf("writing primary field: %w", err)
	}
	bytesWritten++

	// Write Wind Speed if present (2 octets)
	if m.WindSpeed != nil {
		var data [2]byte
		data[0] = byte(*m.WindSpeed >> 8)
		data[1] = byte(*m.WindSpeed)
		n, err := buf.Write(data[:])
		if err != nil {
			return bytesWritten, fmt.Errorf("writing wind speed: %w", err)
		}
		bytesWritten += n
	}

	// Write Wind Direction if present (2 octets)
	if m.WindDirection != nil {
		var data [2]byte
		data[0] = byte(*m.WindDirection >> 8)
		data[1] = byte(*m.WindDirection)
		n, err := buf.Write(data[:])
		if err != nil {
			return bytesWritten, fmt.Errorf("writing wind direction: %w", err)
		}
		bytesWritten += n
	}

	// Write Temperature if present (2 octets)
	if m.Temperature != nil {
		// Convert to raw value: LSB = 0.25°C
		raw := int16(math.Round(*m.Temperature / 0.25))
		var data [2]byte
		data[0] = byte(raw >> 8)
		data[1] = byte(raw)
		n, err := buf.Write(data[:])
		if err != nil {
			return bytesWritten, fmt.Errorf("writing temperature: %w", err)
		}
		bytesWritten += n
	}

	// Write Turbulence if present (1 octet)
	if m.Turbulence != nil {
		if err := buf.WriteByte(*m.Turbulence); err != nil {
			return bytesWritten, fmt.Errorf("writing turbulence: %w", err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

func (m *MetInformation) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Read primary field
	primaryField, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading primary field: %w", err)
	}
	bytesRead++

	hasWindSpeed := (primaryField & 0x80) != 0
	hasWindDirection := (primaryField & 0x40) != 0
	hasTemperature := (primaryField & 0x20) != 0
	hasTurbulence := (primaryField & 0x10) != 0

	// Read Wind Speed if present (2 octets)
	if hasWindSpeed {
		var data [2]byte
		n, err := buf.Read(data[:])
		if err != nil {
			return bytesRead, fmt.Errorf("reading wind speed: %w", err)
		}
		if n != 2 {
			return bytesRead, fmt.Errorf("%w: need 2 bytes for wind speed, have %d", asterix.ErrBufferTooShort, n)
		}
		ws := uint16(data[0])<<8 | uint16(data[1])
		m.WindSpeed = &ws
		bytesRead += n
	}

	// Read Wind Direction if present (2 octets)
	if hasWindDirection {
		var data [2]byte
		n, err := buf.Read(data[:])
		if err != nil {
			return bytesRead, fmt.Errorf("reading wind direction: %w", err)
		}
		if n != 2 {
			return bytesRead, fmt.Errorf("%w: need 2 bytes for wind direction, have %d", asterix.ErrBufferTooShort, n)
		}
		wd := uint16(data[0])<<8 | uint16(data[1])
		m.WindDirection = &wd
		bytesRead += n
	}

	// Read Temperature if present (2 octets)
	if hasTemperature {
		var data [2]byte
		n, err := buf.Read(data[:])
		if err != nil {
			return bytesRead, fmt.Errorf("reading temperature: %w", err)
		}
		if n != 2 {
			return bytesRead, fmt.Errorf("%w: need 2 bytes for temperature, have %d", asterix.ErrBufferTooShort, n)
		}
		raw := int16(uint16(data[0])<<8 | uint16(data[1]))
		temp := float64(raw) * 0.25
		m.Temperature = &temp
		bytesRead += n
	}

	// Read Turbulence if present (1 octet)
	if hasTurbulence {
		turb, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading turbulence: %w", err)
		}
		m.Turbulence = &turb
		bytesRead++
	}

	return bytesRead, m.Validate()
}

func (m *MetInformation) Validate() error {
	if m.WindDirection != nil && *m.WindDirection >= 360 {
		return fmt.Errorf("wind direction out of range [0,360): %d", *m.WindDirection)
	}
	if m.Turbulence != nil && *m.Turbulence > 3 {
		return fmt.Errorf("turbulence level out of range [0,3]: %d", *m.Turbulence)
	}
	return nil
}

func (m *MetInformation) String() string {
	var parts []string

	if m.WindSpeed != nil {
		parts = append(parts, fmt.Sprintf("Wind: %dkts", *m.WindSpeed))
	}
	if m.WindDirection != nil {
		parts = append(parts, fmt.Sprintf("from %d°", *m.WindDirection))
	}
	if m.Temperature != nil {
		parts = append(parts, fmt.Sprintf("Temp: %.1f°C", *m.Temperature))
	}
	if m.Turbulence != nil {
		turbStr := ""
		switch *m.Turbulence {
		case 0:
			turbStr = "Nil"
		case 1:
			turbStr = "Light"
		case 2:
			turbStr = "Moderate"
		case 3:
			turbStr = "Severe"
		}
		parts = append(parts, fmt.Sprintf("Turb: %s", turbStr))
	}

	if len(parts) == 0 {
		return "No met info"
	}

	return strings.Join(parts, ", ")
}
