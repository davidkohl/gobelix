package v13

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// EstimatedAccuracies implements I011/500 - Estimated Accuracies
// Definition: Overview of all important accuracies (standard deviations)
// Format: Compound Data Item, comprising a primary subfield of one octet,
// followed by subfields of predefined length.
type EstimatedAccuracies struct {
	// Presence flags
	HasAPC bool // Position (Cartesian) accuracy
	HasAPW bool // Position (WGS-84) accuracy
	HasATH bool // Height accuracy
	HasAVC bool // Velocity (Cartesian) accuracy
	HasARC bool // Rate of Climb/Descent accuracy
	HasAAC bool // Acceleration accuracy

	// Values
	APCX float64 // X-component position accuracy (m), LSB = 0.25m
	APCY float64 // Y-component position accuracy (m)
	APWLat float64 // Latitude accuracy (degrees), LSB = 180/2^31
	APWLon float64 // Longitude accuracy (degrees)
	ATH  float64 // Height accuracy (m), LSB = 0.5m
	AVCX float64 // X-velocity accuracy (m/s), LSB = 0.1 m/s
	AVCY float64 // Y-velocity accuracy (m/s)
	ARC  float64 // Rate of climb accuracy (m/s), LSB = 0.1 m/s
	AACX float64 // X-acceleration accuracy (m/s²), LSB = 0.01 m/s²
	AACY float64 // Y-acceleration accuracy (m/s²)
}

const (
	apcLSB = 0.25       // meters
	apwLSB = 180.0 / 2147483648.0 // 180/2^31 degrees
	athLSB = 0.5        // meters
	avcLSB = 0.1        // m/s
	arcLSB = 0.1        // m/s
	aacLSB = 0.01       // m/s²
)

func (e *EstimatedAccuracies) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Primary subfield
	primary, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading I011/500 primary: %w", err)
	}
	bytesRead++

	e.HasAPC = (primary & 0x80) != 0
	e.HasAPW = (primary & 0x40) != 0
	e.HasATH = (primary & 0x20) != 0
	e.HasAVC = (primary & 0x10) != 0
	e.HasARC = (primary & 0x08) != 0
	e.HasAAC = (primary & 0x04) != 0
	fx := (primary & 0x01) != 0

	// Skip extensions
	for fx {
		data, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading I011/500 ext: %w", err)
		}
		bytesRead++
		fx = (data & 0x01) != 0
	}

	// Read subfields
	if e.HasAPC {
		// 2 bytes: X and Y components, 1 byte each
		var data [2]byte
		n, err := buf.Read(data[:])
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("reading APC: %w", err)
		}
		e.APCX = float64(data[0]) * apcLSB
		e.APCY = float64(data[1]) * apcLSB
	}

	if e.HasAPW {
		// 4 bytes: Lat and Lon, 2 bytes each
		var data [4]byte
		n, err := buf.Read(data[:])
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("reading APW: %w", err)
		}
		e.APWLat = float64(binary.BigEndian.Uint16(data[0:2])) * apwLSB
		e.APWLon = float64(binary.BigEndian.Uint16(data[2:4])) * apwLSB
	}

	if e.HasATH {
		// 2 bytes
		var data [2]byte
		n, err := buf.Read(data[:])
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("reading ATH: %w", err)
		}
		e.ATH = float64(binary.BigEndian.Uint16(data[:])) * athLSB
	}

	if e.HasAVC {
		// 2 bytes
		var data [2]byte
		n, err := buf.Read(data[:])
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("reading AVC: %w", err)
		}
		e.AVCX = float64(data[0]) * avcLSB
		e.AVCY = float64(data[1]) * avcLSB
	}

	if e.HasARC {
		// 1 byte
		data, err := buf.ReadByte()
		bytesRead++
		if err != nil {
			return bytesRead, fmt.Errorf("reading ARC: %w", err)
		}
		e.ARC = float64(data) * arcLSB
	}

	if e.HasAAC {
		// 2 bytes
		var data [2]byte
		n, err := buf.Read(data[:])
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("reading AAC: %w", err)
		}
		e.AACX = float64(data[0]) * aacLSB
		e.AACY = float64(data[1]) * aacLSB
	}

	return bytesRead, nil
}

func (e *EstimatedAccuracies) Encode(buf *bytes.Buffer) (int, error) {
	bytesWritten := 0

	// Primary subfield
	var primary uint8
	if e.HasAPC {
		primary |= 0x80
	}
	if e.HasAPW {
		primary |= 0x40
	}
	if e.HasATH {
		primary |= 0x20
	}
	if e.HasAVC {
		primary |= 0x10
	}
	if e.HasARC {
		primary |= 0x08
	}
	if e.HasAAC {
		primary |= 0x04
	}
	// FX = 0

	if err := buf.WriteByte(primary); err != nil {
		return bytesWritten, fmt.Errorf("writing I011/500 primary: %w", err)
	}
	bytesWritten++

	// Write subfields
	if e.HasAPC {
		data := []byte{uint8(e.APCX / apcLSB), uint8(e.APCY / apcLSB)}
		if _, err := buf.Write(data); err != nil {
			return bytesWritten, fmt.Errorf("writing APC: %w", err)
		}
		bytesWritten += 2
	}

	if e.HasAPW {
		var data [4]byte
		binary.BigEndian.PutUint16(data[0:2], uint16(e.APWLat/apwLSB))
		binary.BigEndian.PutUint16(data[2:4], uint16(e.APWLon/apwLSB))
		if _, err := buf.Write(data[:]); err != nil {
			return bytesWritten, fmt.Errorf("writing APW: %w", err)
		}
		bytesWritten += 4
	}

	if e.HasATH {
		var data [2]byte
		binary.BigEndian.PutUint16(data[:], uint16(e.ATH/athLSB))
		if _, err := buf.Write(data[:]); err != nil {
			return bytesWritten, fmt.Errorf("writing ATH: %w", err)
		}
		bytesWritten += 2
	}

	if e.HasAVC {
		data := []byte{uint8(e.AVCX / avcLSB), uint8(e.AVCY / avcLSB)}
		if _, err := buf.Write(data); err != nil {
			return bytesWritten, fmt.Errorf("writing AVC: %w", err)
		}
		bytesWritten += 2
	}

	if e.HasARC {
		if err := buf.WriteByte(uint8(e.ARC / arcLSB)); err != nil {
			return bytesWritten, fmt.Errorf("writing ARC: %w", err)
		}
		bytesWritten++
	}

	if e.HasAAC {
		data := []byte{uint8(e.AACX / aacLSB), uint8(e.AACY / aacLSB)}
		if _, err := buf.Write(data); err != nil {
			return bytesWritten, fmt.Errorf("writing AAC: %w", err)
		}
		bytesWritten += 2
	}

	return bytesWritten, nil
}

func (e *EstimatedAccuracies) Validate() error {
	return nil
}

func (e *EstimatedAccuracies) String() string {
	parts := []string{}
	if e.HasAPC {
		parts = append(parts, fmt.Sprintf("Pos(%.2fm,%.2fm)", e.APCX, e.APCY))
	}
	if e.HasATH {
		parts = append(parts, fmt.Sprintf("Alt(%.1fm)", e.ATH))
	}
	if e.HasAVC {
		parts = append(parts, fmt.Sprintf("Vel(%.1fm/s,%.1fm/s)", e.AVCX, e.AVCY))
	}
	if e.HasARC {
		parts = append(parts, fmt.Sprintf("RoC(%.1fm/s)", e.ARC))
	}
	return fmt.Sprintf("Accuracies: %v", parts)
}
