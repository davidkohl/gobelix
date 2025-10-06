package v12

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// RadarPlotCharacteristics represents I001/130 - Radar Plot Characteristics
// Extended item with SRL, SRR, SAM, PRL, PAM, RPD, APD
type RadarPlotCharacteristics struct {
	SRL uint8 // SSR plot runlength
	SRR uint8 // Number of received replies for M(SSR)
	SAM int8  // Amplitude of M(SSR) reply
	PRL uint8 // Primary plot runlength
	PAM int8  // Amplitude of Primary plot
	RPD int8  // Difference in range between PSR and SSR
	APD int8  // Difference in azimuth between PSR and SSR
}

func (r *RadarPlotCharacteristics) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	if buf.Len() < 1 {
		return 0, fmt.Errorf("%w: need at least 1 byte for radar plot characteristics", asterix.ErrBufferTooShort)
	}

	data := buf.Next(1)
	bytesRead++

	// First octet: spare (bit 8), SRL (bits 7-2), FX (bit 1)
	r.SRL = (data[0] >> 1) & 0x3F

	hasFX := (data[0] & 0x01) != 0

	if hasFX && buf.Len() >= 1 {
		data = buf.Next(1)
		bytesRead++
		// Second octet: spare (bit 8), SRR (bits 7-2), FX (bit 1)
		r.SRR = (data[0] >> 1) & 0x3F
		hasFX = (data[0] & 0x01) != 0
	}

	if hasFX && buf.Len() >= 1 {
		data = buf.Next(1)
		bytesRead++
		// Third octet: SAM (bits 8-2, signed), FX (bit 1)
		sam := int16(data[0] >> 1)
		if sam >= 64 {
			sam = sam - 128
		}
		r.SAM = int8(sam)
		hasFX = (data[0] & 0x01) != 0
	}

	if hasFX && buf.Len() >= 1 {
		data = buf.Next(1)
		bytesRead++
		// Fourth octet: PRL (bits 8-2), FX (bit 1)
		r.PRL = (data[0] >> 1) & 0x7F
		hasFX = (data[0] & 0x01) != 0
	}

	if hasFX && buf.Len() >= 1 {
		data = buf.Next(1)
		bytesRead++
		// Fifth octet: PAM (bits 8-2, signed), FX (bit 1)
		pam := int16(data[0] >> 1)
		if pam >= 64 {
			pam = pam - 128
		}
		r.PAM = int8(pam)
		hasFX = (data[0] & 0x01) != 0
	}

	if hasFX && buf.Len() >= 1 {
		data = buf.Next(1)
		bytesRead++
		// Sixth octet: RPD (bits 8-2, signed), FX (bit 1)
		rpd := int16(data[0] >> 1)
		if rpd >= 64 {
			rpd = rpd - 128
		}
		r.RPD = int8(rpd)
		hasFX = (data[0] & 0x01) != 0
	}

	if hasFX && buf.Len() >= 1 {
		data = buf.Next(1)
		bytesRead++
		// Seventh octet: APD (bits 8-2, signed), FX (bit 1)
		apd := int16(data[0] >> 1)
		if apd >= 64 {
			apd = apd - 128
		}
		r.APD = int8(apd)
		hasFX = (data[0] & 0x01) != 0
	}

	// Handle any remaining extensions
	for hasFX && buf.Len() >= 1 {
		data = buf.Next(1)
		bytesRead++
		hasFX = (data[0] & 0x01) != 0
	}

	return bytesRead, nil
}

func (r *RadarPlotCharacteristics) Encode(buf *bytes.Buffer) (int, error) {
	bytesWritten := 0

	// Determine which octets are needed
	needOctet7 := r.APD != 0
	needOctet6 := needOctet7 || r.RPD != 0
	needOctet5 := needOctet6 || r.PAM != 0
	needOctet4 := needOctet5 || r.PRL != 0
	needOctet3 := needOctet4 || r.SAM != 0
	needOctet2 := needOctet3 || r.SRR != 0

	// First octet
	octet1 := (r.SRL & 0x3F) << 1
	if needOctet2 {
		octet1 |= 0x01
	}
	buf.WriteByte(octet1)
	bytesWritten++

	if needOctet2 {
		octet2 := (r.SRR & 0x3F) << 1
		if needOctet3 {
			octet2 |= 0x01
		}
		buf.WriteByte(octet2)
		bytesWritten++
	}

	if needOctet3 {
		octet3 := (uint8(r.SAM) & 0x7F) << 1
		if needOctet4 {
			octet3 |= 0x01
		}
		buf.WriteByte(octet3)
		bytesWritten++
	}

	if needOctet4 {
		octet4 := (r.PRL & 0x7F) << 1
		if needOctet5 {
			octet4 |= 0x01
		}
		buf.WriteByte(octet4)
		bytesWritten++
	}

	if needOctet5 {
		octet5 := (uint8(r.PAM) & 0x7F) << 1
		if needOctet6 {
			octet5 |= 0x01
		}
		buf.WriteByte(octet5)
		bytesWritten++
	}

	if needOctet6 {
		octet6 := (uint8(r.RPD) & 0x7F) << 1
		if needOctet7 {
			octet6 |= 0x01
		}
		buf.WriteByte(octet6)
		bytesWritten++
	}

	if needOctet7 {
		octet7 := (uint8(r.APD) & 0x7F) << 1
		// No FX bit (last octet)
		buf.WriteByte(octet7)
		bytesWritten++
	}

	return bytesWritten, nil
}

func (r *RadarPlotCharacteristics) String() string {
	return fmt.Sprintf("SRL=%d SRR=%d SAM=%d PRL=%d PAM=%d RPD=%d APD=%d",
		r.SRL, r.SRR, r.SAM, r.PRL, r.PAM, r.RPD, r.APD)
}

func (r *RadarPlotCharacteristics) Validate() error {
	return nil
}
