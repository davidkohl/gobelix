// cat/cat020/dataitems/v15/acas_resolution_advisory.go
package v15

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// ACASResolutionAdvisory implements I020/260
// ACAS Resolution Advisory Report (7 octets)
type ACASResolutionAdvisory struct {
	TYP uint8    // Type of threat (0-3)
	STYP uint8   // Subtype (0-3)
	ARA [14]bool // Active Resolution Advisories (14 bits)
	RAC uint8   // RA Complement (0-15)
	RAT bool    // RA Terminated
	MTE bool    // Multiple Threat Encounter
	TTI uint8   // Threat Type Indicator (0-3)
	TID [4]byte // Threat Identity Data (4 characters, 6-bit IA5)
}

func (a *ACASResolutionAdvisory) Encode(buf *bytes.Buffer) (int, error) {
	if err := a.Validate(); err != nil {
		return 0, err
	}

	data := make([]byte, 7)

	// Octet 1: TYP (2 bits), STYP (2 bits), ARA bits 13-10 (4 bits)
	data[0] = (a.TYP & 0x03) << 6
	data[0] |= (a.STYP & 0x03) << 4
	if a.ARA[13] {
		data[0] |= 0x08
	}
	if a.ARA[12] {
		data[0] |= 0x04
	}
	if a.ARA[11] {
		data[0] |= 0x02
	}
	if a.ARA[10] {
		data[0] |= 0x01
	}

	// Octet 2: ARA bits 9-2 (8 bits)
	for i := 9; i >= 2; i-- {
		if a.ARA[i] {
			data[1] |= 1 << (i - 2)
		}
	}

	// Octet 3: ARA bits 1-0 (2 bits), RAC (4 bits), RAT (1 bit), MTE (1 bit)
	if a.ARA[1] {
		data[2] |= 0x80
	}
	if a.ARA[0] {
		data[2] |= 0x40
	}
	data[2] |= (a.RAC & 0x0F) << 2
	if a.RAT {
		data[2] |= 0x02
	}
	if a.MTE {
		data[2] |= 0x01
	}

	// Octet 4: TTI (2 bits), spare (6 bits)
	data[3] = (a.TTI & 0x03) << 6

	// Octets 5-7: Threat Identity Data (TID) - 3 octets for 4 characters (6-bit each)
	// First character
	data[4] = (a.TID[0] & 0x3F) << 2
	// Second character (split across octets 5-6)
	data[4] |= (a.TID[1] & 0x30) >> 4
	data[5] = (a.TID[1] & 0x0F) << 4
	// Third character (split across octets 6-7)
	data[5] |= (a.TID[2] & 0x3C) >> 2
	data[6] = (a.TID[2] & 0x03) << 6
	// Fourth character
	data[6] |= a.TID[3] & 0x3F

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing ACAS resolution advisory: %w", err)
	}
	return n, nil
}

func (a *ACASResolutionAdvisory) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 7)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading ACAS resolution advisory: %w", err)
	}
	if n != 7 {
		return n, fmt.Errorf("%w: need 7 bytes for ACAS RA, have %d", asterix.ErrBufferTooShort, n)
	}

	// Octet 1
	a.TYP = (data[0] >> 6) & 0x03
	a.STYP = (data[0] >> 4) & 0x03
	a.ARA[13] = (data[0] & 0x08) != 0
	a.ARA[12] = (data[0] & 0x04) != 0
	a.ARA[11] = (data[0] & 0x02) != 0
	a.ARA[10] = (data[0] & 0x01) != 0

	// Octet 2
	for i := 9; i >= 2; i-- {
		a.ARA[i] = (data[1] & (1 << (i - 2))) != 0
	}

	// Octet 3
	a.ARA[1] = (data[2] & 0x80) != 0
	a.ARA[0] = (data[2] & 0x40) != 0
	a.RAC = (data[2] >> 2) & 0x0F
	a.RAT = (data[2] & 0x02) != 0
	a.MTE = (data[2] & 0x01) != 0

	// Octet 4
	a.TTI = (data[3] >> 6) & 0x03

	// Octets 5-7: Decode TID
	a.TID[0] = (data[4] >> 2) & 0x3F
	a.TID[1] = ((data[4] & 0x03) << 4) | ((data[5] >> 4) & 0x0F)
	a.TID[2] = ((data[5] & 0x0F) << 2) | ((data[6] >> 6) & 0x03)
	a.TID[3] = data[6] & 0x3F

	return n, a.Validate()
}

func (a *ACASResolutionAdvisory) Validate() error {
	if a.TYP > 3 {
		return fmt.Errorf("invalid threat type: %d", a.TYP)
	}
	if a.STYP > 3 {
		return fmt.Errorf("invalid subtype: %d", a.STYP)
	}
	if a.RAC > 15 {
		return fmt.Errorf("invalid RA complement: %d", a.RAC)
	}
	if a.TTI > 3 {
		return fmt.Errorf("invalid threat type indicator: %d", a.TTI)
	}
	return nil
}

func (a *ACASResolutionAdvisory) String() string {
	threatType := ""
	switch a.TYP {
	case 0:
		threatType = "No ID"
	case 1:
		threatType = "Single threat"
	case 2:
		threatType = "Multiple threat"
	case 3:
		threatType = "Reserved"
	}

	return fmt.Sprintf("Type: %s, RAT: %t, MTE: %t", threatType, a.RAT, a.MTE)
}
