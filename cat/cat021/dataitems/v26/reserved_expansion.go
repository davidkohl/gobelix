// dataitems/cat021/reserved_expansion.go
package v26

import (
	"bytes"
	"fmt"
	"math"
)

// ReservedExpansion implements the CAT021 Reserved Expansion Field
// (EUROCONTROL ASTERIX Part 12a, CAT021 REF, Edition 1.5). Framing:
//
//	[LEN] total REF length in octets INCLUDING the LEN octet itself
//	[Items Indicator] presence bits BPS/SelH/NAV/GAO/SGV/STA/TNH/MES
//	[present sub-items in indicator-bit order]
//
// Structured sub-items are the ones ED-129B requires a ground system to
// source: BPS (REQ 274/309), SelH (REQ 311), GAO (REQ 275). A pre-framed
// payload (INCLUDING the LEN octet) can instead be carried verbatim in Data;
// Data is ignored when any structured field is set.
type ReservedExpansion struct {
	HasBPS bool
	BPSHpa float64 // barometric pressure setting in hPa (e.g. 1013.2)

	HasSelH      bool
	SelHDeg      float64 // selected heading, degrees
	SelHMagnetic bool    // HRD: true = Magnetic North
	SelHValid    bool    // Stat: heading data available and valid

	HasGAO bool
	GAO    uint8 // raw GPS Antenna Offset byte

	Data []byte
}

func (r *ReservedExpansion) structured() bool { return r.HasBPS || r.HasSelH || r.HasGAO }

func (r *ReservedExpansion) Encode(buf *bytes.Buffer) (int, error) {
	if !r.structured() {
		if len(r.Data) == 0 {
			return buf.Write([]byte{0})
		}
		return buf.Write(r.Data)
	}
	var ind byte
	length := 2 // LEN octet + Items Indicator octet
	if r.HasBPS {
		ind |= 0x80
		length += 2
	}
	if r.HasSelH {
		ind |= 0x40
		length += 2
	}
	if r.HasGAO {
		ind |= 0x10
		length++
	}
	start := buf.Len()
	buf.WriteByte(byte(length))
	buf.WriteByte(ind)
	if r.HasBPS {
		// BPS = (setting − 800) hPa, LSB 0.1 hPa in bits 12/1; bits 16/13 spare.
		v := int(math.Round((r.BPSHpa - 800) / 0.1))
		if v < 0 {
			v = 0
		}
		if v > 0x0FFF {
			v = 0x0FFF
		}
		buf.WriteByte(byte(v >> 8))
		buf.WriteByte(byte(v))
	}
	if r.HasSelH {
		// bit-12 HRD, bit-11 Stat, bits 10/1 SelH (LSB 0.703125°); 16/13 spare.
		h := int(math.Round(r.SelHDeg / 0.703125))
		if h < 0 {
			h = 0
		}
		if h > 0x3FF {
			h = 0x3FF
		}
		v := uint16(h) & 0x3FF
		if r.SelHMagnetic {
			v |= 1 << 11
		}
		if r.SelHValid {
			v |= 1 << 10
		}
		buf.WriteByte(byte(v >> 8))
		buf.WriteByte(byte(v))
	}
	if r.HasGAO {
		buf.WriteByte(r.GAO)
	}
	return buf.Len() - start, nil
}

func (r *ReservedExpansion) Decode(buf *bytes.Buffer) (int, error) {
	// LEN counts the whole REF INCLUDING the LEN octet — read LEN-1 further
	// bytes (the old passthrough read LEN more, over-reading by one byte).
	lenByte, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading reserved expansion length: %w", err)
	}
	if lenByte < 1 {
		return 1, fmt.Errorf("reserved expansion length %d invalid", lenByte)
	}
	rest := make([]byte, int(lenByte)-1)
	m, err := buf.Read(rest)
	if err != nil || m != len(rest) {
		return 1 + m, fmt.Errorf("reserved expansion truncated: want %d bytes, got %d", len(rest), m)
	}
	r.Data = append([]byte{lenByte}, rest...)
	return 1 + m, nil
}

func (r *ReservedExpansion) Validate() error { return nil }

func (r *ReservedExpansion) String() string {
	if r.structured() {
		return fmt.Sprintf("REF{BPS:%v SelH:%v GAO:%v}", r.HasBPS, r.HasSelH, r.HasGAO)
	}
	return fmt.Sprintf("REF{raw %d bytes}", len(r.Data))
}
