// dataitems/cat021/mode_s_mb_data.go
package v26

import (
	"bytes"
	"fmt"
)

// ModeSMBDataRecord represents a single Mode S MB Data record
type ModeSMBDataRecord struct {
	MBData [8]byte // 56-bit Mode S Comm-B message
	BDS1   uint8   // BDS register 1 (4 bits)
	BDS2   uint8   // BDS register 2 (4 bits)
}

// ModeSMBData implements I021/250
// Mode S MB Data (Repetitive, 8 octets per record)
type ModeSMBData struct {
	Records []ModeSMBDataRecord
}

func (m *ModeSMBData) Encode(buf *bytes.Buffer) (int, error) {
	if err := m.Validate(); err != nil {
		return 0, err
	}

	bytesWritten := 0

	// Write repetition factor (number of records)
	rep := uint8(len(m.Records))
	if rep > 255 {
		return 0, fmt.Errorf("too many Mode S MB records: %d (max 255)", len(m.Records))
	}

	if err := buf.WriteByte(rep); err != nil {
		return bytesWritten, fmt.Errorf("writing repetition factor: %w", err)
	}
	bytesWritten++

	// Write each record
	for i, record := range m.Records {
		// Write 56-bit MB data (7 octets)
		n, err := buf.Write(record.MBData[:7])
		if err != nil {
			return bytesWritten, fmt.Errorf("writing MB data for record %d: %w", i, err)
		}
		bytesWritten += n

		// Write BDS register (1 octet: BDS1 in upper 4 bits, BDS2 in lower 4 bits)
		bds := (record.BDS1 << 4) | (record.BDS2 & 0x0F)
		if err := buf.WriteByte(bds); err != nil {
			return bytesWritten, fmt.Errorf("writing BDS register for record %d: %w", i, err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

func (m *ModeSMBData) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Read repetition factor
	rep, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading repetition factor: %w", err)
	}
	bytesRead++

	if rep == 0 {
		m.Records = []ModeSMBDataRecord{}
		return bytesRead, nil
	}

	// Read each record
	m.Records = make([]ModeSMBDataRecord, rep)
	for i := 0; i < int(rep); i++ {
		// Read 56-bit MB data (7 octets)
		mbData := make([]byte, 7)
		n, err := buf.Read(mbData)
		if err != nil {
			return bytesRead, fmt.Errorf("reading MB data for record %d: %w", i, err)
		}
		if n != 7 {
			return bytesRead, fmt.Errorf("insufficient MB data for record %d: got %d bytes, want 7", i, n)
		}
		copy(m.Records[i].MBData[:7], mbData)
		bytesRead += n

		// Read BDS register (1 octet)
		bds, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading BDS register for record %d: %w", i, err)
		}
		m.Records[i].BDS1 = (bds >> 4) & 0x0F
		m.Records[i].BDS2 = bds & 0x0F
		bytesRead++
	}

	return bytesRead, m.Validate()
}

func (m *ModeSMBData) Validate() error {
	for i, record := range m.Records {
		if record.BDS1 > 15 {
			return fmt.Errorf("invalid BDS1 for record %d: %d", i, record.BDS1)
		}
		if record.BDS2 > 15 {
			return fmt.Errorf("invalid BDS2 for record %d: %d", i, record.BDS2)
		}
	}
	return nil
}

func (m *ModeSMBData) String() string {
	if len(m.Records) == 0 {
		return "No MB data"
	}
	return fmt.Sprintf("%d Mode S MB record(s)", len(m.Records))
}
