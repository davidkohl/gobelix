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
// Structured sub-items: BPS (REQ 274/309), SelH (REQ 311), NAV, GAO (REQ
// 275), TNH and SGV (§2.7, decoded into the SGV* fields and also kept as
// RawSGV). STA and MES are compound (FX-extended, MES military-specific) and
// are decoded only as raw bytes into their Raw* fields; String() renders them
// as hex rather than silently dropping them.
//
// A pre-framed payload (INCLUDING the LEN octet) can instead be carried
// verbatim in Data; Data is ignored by Encode when any structured field is
// set.
type ReservedExpansion struct {
	HasBPS bool
	BPSHpa float64 // barometric pressure setting in hPa (e.g. 1013.2)

	HasSelH      bool
	SelHDeg      float64 // selected heading, degrees
	SelHMagnetic bool    // HRD: true = Magnetic North
	SelHValid    bool    // Stat: heading data available and valid

	HasNAV          bool
	NAVAutopilot    bool // AP
	NAVVNAV         bool // VN
	NAVAltitudeHold bool // AH
	NAVApproach     bool // AM
	NAVMCPPopulated bool // MFM#VAL: MCP/FCU mode bits (AP/VN/AH/AM) actively populated

	HasGAO           bool
	GAORight         bool // bit-8: direction, false = left of centreline, true = right
	GAOLateralM      int  // lateral axis offset, metres (LSB 2m)
	GAOLongitudinalM int  // longitudinal axis offset, metres (LSB 2m)

	HasTNH bool
	TNHDeg float64 // true north heading, degrees (LSB 360/2^16)

	// SGV, Surface Ground Vector (REF §2.7), sent by aircraft on the
	// surface instead of I021/160.  Octets 1-2: bit-16 STP, bit-15 HTS,
	// bit-14 HTT, bit-13 HRD, bits-12/2 GSS (LSB 0.125 kt), bit-1 FX; first
	// extension: bits-8/2 HGT (LSB 360/128°), bit-1 FX.  Further extensions
	// are kept in RawSGV only.  Decode fills both RawSGV and the fields;
	// Encode writes RawSGV when set, else builds the item from the fields.
	HasSGV        bool
	SGVStopped    bool    // STP: aircraft has stopped (else moving)
	SGVValid      bool    // HTS: heading/ground track data valid
	SGVTrack      bool    // HTT: HGT is the ground track (else heading)
	SGVMagnetic   bool    // HRD: HGT referenced to magnetic north (else true north)
	SGVSpeedKt    float64 // GSS: ground speed, kt (LSB 0.125 kt, 0-255.875)
	SGVHasHeading bool    // first extension present (HGT sent)
	SGVHeadingDeg float64 // HGT: heading or ground track, degrees (LSB 360/128)
	RawSGV        []byte  // the item as sent, primary subfield + extensions

	HasSTA bool
	RawSTA []byte // Aircraft Status, compound/FX-extended

	HasMES bool
	RawMES []byte // Military Extended Squitter, compound/FX-extended

	Data []byte
}

func (r *ReservedExpansion) structured() bool {
	return r.HasBPS || r.HasSelH || r.HasNAV || r.HasGAO || r.HasTNH ||
		r.HasSGV || r.HasSTA || r.HasMES
}

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
	if r.HasNAV {
		ind |= 0x20
		length++
	}
	if r.HasGAO {
		ind |= 0x10
		length++
	}
	var sgv []byte
	if r.HasSGV {
		sgv = r.RawSGV
		if len(sgv) == 0 {
			sgv = r.encodeSGV()
		}
		ind |= 0x08
		length += len(sgv)
	}
	if r.HasSTA {
		ind |= 0x04
		length += len(r.RawSTA)
	}
	if r.HasTNH {
		ind |= 0x02
		length += 2
	}
	if r.HasMES {
		ind |= 0x01
		length += len(r.RawMES)
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
	if r.HasNAV {
		// bit-8 AP, bit-7 VN, bit-6 AH, bit-5 AM, bit-4 MFM#EP(=1), bit-3 MFM#VAL; bits 2/1 spare.
		var v byte
		if r.NAVAutopilot {
			v |= 0x80
		}
		if r.NAVVNAV {
			v |= 0x40
		}
		if r.NAVAltitudeHold {
			v |= 0x20
		}
		if r.NAVApproach {
			v |= 0x10
		}
		v |= 0x08 // MFM#EP: element populated
		if r.NAVMCPPopulated {
			v |= 0x04
		}
		buf.WriteByte(v)
	}
	if r.HasGAO {
		// bits-8/6 lateral (LSB 2m), bits-5/1 longitudinal (LSB 2m); bit-8 direction.
		lat := clampGAO(r.GAOLateralM / 2)
		lon := clampGAO(r.GAOLongitudinalM / 2)
		v := (lat & 0x03) << 5
		if r.GAORight {
			v |= 0x80
		}
		v |= lon & 0x1F
		buf.WriteByte(byte(v))
	}
	if r.HasSGV {
		buf.Write(sgv)
	}
	if r.HasSTA {
		buf.Write(r.RawSTA)
	}
	if r.HasTNH {
		// LSB = 360/2^16 degrees.
		v := uint16(math.Round(r.TNHDeg / (360.0 / 65536.0)))
		buf.WriteByte(byte(v >> 8))
		buf.WriteByte(byte(v))
	}
	if r.HasMES {
		buf.Write(r.RawMES)
	}
	return buf.Len() - start, nil
}

// encodeSGV builds the SGV item from the SGV* fields: two octets, plus the
// HGT extension when SGVHasHeading.
func (r *ReservedExpansion) encodeSGV() []byte {
	gss := int(math.Round(r.SGVSpeedKt / 0.125))
	if gss < 0 {
		gss = 0
	}
	if gss > 0x7FF {
		gss = 0x7FF
	}
	b0 := byte(gss >> 7 & 0x0F)
	if r.SGVStopped {
		b0 |= 0x80
	}
	if r.SGVValid {
		b0 |= 0x40
	}
	if r.SGVTrack {
		b0 |= 0x20
	}
	if r.SGVMagnetic {
		b0 |= 0x10
	}
	b1 := byte(gss&0x7F) << 1
	if !r.SGVHasHeading {
		return []byte{b0, b1}
	}
	h := math.Mod(r.SGVHeadingDeg, 360)
	if h < 0 {
		h += 360
	}
	hgt := int(math.Round(h/(360.0/128))) & 0x7F // 360° wraps to 0
	return []byte{b0, b1 | 0x01, byte(hgt) << 1}
}

// decodeSGV fills the SGV* fields from an item framed by readFXItem (at
// least two octets, extensions as announced by FX).
func (r *ReservedExpansion) decodeSGV(b []byte) {
	r.SGVStopped = b[0]&0x80 != 0
	r.SGVValid = b[0]&0x40 != 0
	r.SGVTrack = b[0]&0x20 != 0
	r.SGVMagnetic = b[0]&0x10 != 0
	r.SGVSpeedKt = float64(int(b[0]&0x0F)<<7|int(b[1]>>1)) * 0.125
	if len(b) > 2 {
		r.SGVHasHeading = true
		r.SGVHeadingDeg = float64(b[2]>>1) * (360.0 / 128)
	}
}

func clampGAO(v int) int {
	if v < 0 {
		return 0
	}
	return v
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

	*r = ReservedExpansion{Data: r.Data}
	if len(rest) == 0 {
		return 1 + m, nil
	}

	ind := rest[0]
	body := rest[1:]
	consume := func(n int) ([]byte, error) {
		if len(body) < n {
			return nil, fmt.Errorf("reserved expansion: need %d bytes, have %d", n, len(body))
		}
		sub := body[:n]
		body = body[n:]
		return sub, nil
	}

	if ind&0x80 != 0 { // BPS
		b, err := consume(2)
		if err != nil {
			return 1 + m, err
		}
		v := int(b[0]&0x0F)<<8 | int(b[1])
		r.HasBPS = true
		r.BPSHpa = 800 + float64(v)*0.1
	}
	if ind&0x40 != 0 { // SelH
		b, err := consume(2)
		if err != nil {
			return 1 + m, err
		}
		r.HasSelH = true
		r.SelHMagnetic = b[0]&0x08 != 0
		r.SelHValid = b[0]&0x04 != 0
		v := int(b[0]&0x03)<<8 | int(b[1])
		r.SelHDeg = float64(v) * 0.703125
	}
	if ind&0x20 != 0 { // NAV
		b, err := consume(1)
		if err != nil {
			return 1 + m, err
		}
		r.HasNAV = true
		r.NAVAutopilot = b[0]&0x80 != 0
		r.NAVVNAV = b[0]&0x40 != 0
		r.NAVAltitudeHold = b[0]&0x20 != 0
		r.NAVApproach = b[0]&0x10 != 0
		r.NAVMCPPopulated = b[0]&0x04 != 0
	}
	if ind&0x10 != 0 { // GAO
		b, err := consume(1)
		if err != nil {
			return 1 + m, err
		}
		r.HasGAO = true
		r.GAORight = b[0]&0x80 != 0
		r.GAOLateralM = int((b[0]>>5)&0x03) * 2
		r.GAOLongitudinalM = int(b[0]&0x1F) * 2
	}
	if ind&0x08 != 0 { // SGV: variable length, FX-extended primary + extensions.
		sgv, n, err := readFXItem(body, 2)
		if err != nil {
			return 1 + m, fmt.Errorf("reserved expansion SGV: %w", err)
		}
		r.HasSGV = true
		r.RawSGV = sgv
		r.decodeSGV(sgv)
		body = body[n:]
	}
	if ind&0x04 != 0 { // STA: compound, FX-extended primary + extensions.
		sta, n, err := readFXItem(body, 1)
		if err != nil {
			return 1 + m, fmt.Errorf("reserved expansion STA: %w", err)
		}
		r.HasSTA = true
		r.RawSTA = sta
		body = body[n:]
	}
	if ind&0x02 != 0 { // TNH
		b, err := consume(2)
		if err != nil {
			return 1 + m, err
		}
		r.HasTNH = true
		v := int(b[0])<<8 | int(b[1])
		r.TNHDeg = float64(v) * (360.0 / 65536.0)
	}
	if ind&0x01 != 0 { // MES: compound, FX-extended primary + subfields.
		mes, n, err := readFXItem(body, 1)
		if err != nil {
			return 1 + m, fmt.Errorf("reserved expansion MES: %w", err)
		}
		r.HasMES = true
		r.RawMES = mes
		body = body[n:]
	}

	return 1 + m, nil
}

// readFXItem consumes an FX-extended item from data: a primary subfield of
// primaryOctets octets whose LSB is the Field Extension bit, followed by
// zero or more one-octet extensions as long as each extension's LSB is set.
// It returns the raw bytes consumed (primary + extensions) and their count.
func readFXItem(data []byte, primaryOctets int) ([]byte, int, error) {
	if len(data) < primaryOctets {
		return nil, 0, fmt.Errorf("need %d bytes for primary subfield, have %d", primaryOctets, len(data))
	}
	n := primaryOctets
	for data[n-1]&0x01 != 0 {
		if len(data) < n+1 {
			return nil, 0, fmt.Errorf("FX bit set but no extension octet available")
		}
		n++
	}
	return data[:n], n, nil
}

func (r *ReservedExpansion) Validate() error { return nil }

func (r *ReservedExpansion) String() string {
	if !r.structured() {
		return fmt.Sprintf("REF{raw % X}", r.Data)
	}
	var parts []string
	if r.HasBPS {
		parts = append(parts, fmt.Sprintf("BPS:%.1fhPa", r.BPSHpa))
	}
	if r.HasSelH {
		ref := "True"
		if r.SelHMagnetic {
			ref = "Mag"
		}
		valid := "invalid"
		if r.SelHValid {
			valid = "valid"
		}
		parts = append(parts, fmt.Sprintf("SelH:%.3f°(%s,%s)", r.SelHDeg, ref, valid))
	}
	if r.HasNAV {
		parts = append(parts, fmt.Sprintf("NAV{AP:%v VN:%v AH:%v AM:%v MCPPop:%v}",
			r.NAVAutopilot, r.NAVVNAV, r.NAVAltitudeHold, r.NAVApproach, r.NAVMCPPopulated))
	}
	if r.HasGAO {
		side := "left"
		if r.GAORight {
			side = "right"
		}
		parts = append(parts, fmt.Sprintf("GAO{lat:%dm lon:%dm %s}", r.GAOLateralM, r.GAOLongitudinalM, side))
	}
	if r.HasSGV {
		sgv := fmt.Sprintf("SGV{GS:%.3fkt stopped:%v valid:%v", r.SGVSpeedKt, r.SGVStopped, r.SGVValid)
		if r.SGVHasHeading {
			kind, ref := "hdg", "True"
			if r.SGVTrack {
				kind = "trk"
			}
			if r.SGVMagnetic {
				ref = "Mag"
			}
			sgv += fmt.Sprintf(" %s:%.4f°(%s)", kind, r.SGVHeadingDeg, ref)
		}
		parts = append(parts, sgv+"}")
	}
	if r.HasSTA {
		parts = append(parts, fmt.Sprintf("STA:% X", r.RawSTA))
	}
	if r.HasTNH {
		parts = append(parts, fmt.Sprintf("TNH:%.4f°", r.TNHDeg))
	}
	if r.HasMES {
		parts = append(parts, fmt.Sprintf("MES:% X", r.RawMES))
	}
	out := "REF{"
	for i, p := range parts {
		if i > 0 {
			out += " "
		}
		out += p
	}
	return out + "}"
}
