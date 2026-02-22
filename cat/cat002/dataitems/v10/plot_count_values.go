package v10

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// PlotCountValue represents a single plot count entry
type PlotCountValue struct {
	A       bool   // Aerial identification (0=antenna 1, 1=antenna 2)
	Ident   uint8  // Plot category identification (1=primary, 2=SSR, 3=combined)
	Counter uint16 // 10-bit counter value
}

// PlotCountValues represents I002/070 - Plot Count Values
// Repetitive data item with REP indicator
type PlotCountValues struct {
	Values []PlotCountValue
}

func (p *PlotCountValues) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Read REP (repetition factor)
	rep, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("%w: reading plot count REP", asterix.ErrBufferTooShort)
	}
	bytesRead++

	// Initialize values slice
	p.Values = make([]PlotCountValue, rep)

	// Read each plot count value (2 bytes each)
	for i := 0; i < int(rep); i++ {
		var data [2]byte
		n, err := buf.Read(data[:])
		if err != nil || n != 2 {
			return bytesRead + n, fmt.Errorf("%w: need 2 bytes for plot count value, have %d", asterix.ErrBufferTooShort, n)
		}
		bytesRead += n

		// Extract fields
		p.Values[i].A = (data[0] & 0x80) != 0              // bit 16
		p.Values[i].Ident = (data[0] >> 3) & 0x1F          // bits 15-11
		p.Values[i].Counter = uint16(data[0]&0x03)<<8 | uint16(data[1]) // bits 10-1
	}

	return bytesRead, nil
}

func (p *PlotCountValues) Encode(buf *bytes.Buffer) (int, error) {
	if err := p.Validate(); err != nil {
		return 0, err
	}

	bytesWritten := 0

	// Write REP
	rep := byte(len(p.Values))
	if err := buf.WriteByte(rep); err != nil {
		return bytesWritten, fmt.Errorf("writing plot count REP: %w", err)
	}
	bytesWritten++

	// Write each plot count value
	for _, val := range p.Values {
		var data [2]byte

		// Build first byte: A (bit 16), Ident (bits 15-11), Counter high 2 bits (bits 10-9)
		if val.A {
			data[0] |= 0x80
		}
		data[0] |= (val.Ident & 0x1F) << 3
		data[0] |= byte((val.Counter >> 8) & 0x03)

		// Second byte: Counter low 8 bits
		data[1] = byte(val.Counter & 0xFF)

		n, err := buf.Write(data[:])
		if err != nil {
			return bytesWritten + n, fmt.Errorf("writing plot count value: %w", err)
		}
		bytesWritten += n
	}

	return bytesWritten, nil
}

func (p *PlotCountValues) Validate() error {
	if len(p.Values) > 255 {
		return fmt.Errorf("%w: too many plot count values: %d", asterix.ErrInvalidMessage, len(p.Values))
	}
	for i, val := range p.Values {
		if val.Ident > 31 {
			return fmt.Errorf("%w: plot category ident at index %d out of range: %d", asterix.ErrInvalidMessage, i, val.Ident)
		}
		if val.Counter > 1023 {
			return fmt.Errorf("%w: counter at index %d out of range: %d", asterix.ErrInvalidMessage, i, val.Counter)
		}
	}
	return nil
}

func (p *PlotCountValues) String() string {
	if len(p.Values) == 0 {
		return "Plot Counts: None"
	}

	result := fmt.Sprintf("Plot Counts (%d):", len(p.Values))
	for i, val := range p.Values {
		antenna := "ANT1"
		if val.A {
			antenna = "ANT2"
		}
		category := "Unknown"
		switch val.Ident {
		case 1:
			category = "Primary"
		case 2:
			category = "SSR"
		case 3:
			category = "Combined"
		}
		result += fmt.Sprintf("\n  [%d] %s %s: %d", i, antenna, category, val.Counter)
	}
	return result
}
