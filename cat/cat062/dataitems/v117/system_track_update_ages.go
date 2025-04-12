// dataitems/cat062/system_track_update_ages.go
package v117

import (
	"bytes"
	"fmt"
	"strings"
)

// SystemTrackUpdateAges implements I062/290
// Ages of the last plot/local track/target report update for each sensor type
type SystemTrackUpdateAges struct {
	// Track age since first occurrence
	TrackAge *float64 // In seconds

	// Ages of the last detection used to update the track for each sensor type
	// All ages in seconds, measured from the Time of Track Information (I062/070)
	PSRAge       *float64 // Primary Surveillance Radar age
	SSRAge       *float64 // Secondary Surveillance Radar age
	ModeS_Age    *float64 // Mode S age
	ADSC_Age     *float64 // ADS-C age
	ADSB_ES_Age  *float64 // ADS-B Extended Squitter age
	ADSB_VDL_Age *float64 // ADS-B VDL Mode 4 age
	ADSB_UAT_Age *float64 // ADS-B UAT age
	LoopAge      *float64 // Loop age
	MLTAge       *float64 // Multilateration age

	// Raw data for reporting purposes
	rawData []byte
}

// Decode parses an ASTERIX Category 062 I290 data item from the buffer
func (s *SystemTrackUpdateAges) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0
	s.rawData = nil

	// Read FSPEC bytes (primary subfield)
	if buf.Len() < 1 {
		return 0, fmt.Errorf("buffer too short for primary subfield")
	}

	fspec1, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading primary subfield: %w", err)
	}
	bytesRead++
	s.rawData = append(s.rawData, fspec1)

	// Check if we have a second FSPEC byte
	hasSecondFSPEC := (fspec1 & 0x01) != 0
	var fspec2 byte
	if hasSecondFSPEC {
		if buf.Len() < 1 {
			return bytesRead, fmt.Errorf("buffer too short for second FSPEC byte")
		}
		fspec2, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading second FSPEC byte: %w", err)
		}
		bytesRead++
		s.rawData = append(s.rawData, fspec2)
	}

	// Process the first FSPEC byte
	// FRN 1 (bit 8): Track age
	if (fspec1 & 0x80) != 0 {
		if buf.Len() < 1 {
			return bytesRead, fmt.Errorf("buffer too short for track age")
		}
		data := make([]byte, 1)
		n, err := buf.Read(data)
		if err != nil || n != 1 {
			return bytesRead + n, fmt.Errorf("reading track age: %w", err)
		}
		bytesRead += n
		s.rawData = append(s.rawData, data...)

		age := float64(data[0]) * 0.25 // LSB = 1/4 second
		s.TrackAge = &age
	}

	// FRN 2 (bit 7): PSR age
	if (fspec1 & 0x40) != 0 {
		if buf.Len() < 1 {
			return bytesRead, fmt.Errorf("buffer too short for PSR age")
		}
		data := make([]byte, 1)
		n, err := buf.Read(data)
		if err != nil || n != 1 {
			return bytesRead + n, fmt.Errorf("reading PSR age: %w", err)
		}
		bytesRead += n
		s.rawData = append(s.rawData, data...)

		age := float64(data[0]) * 0.25 // LSB = 1/4 second
		s.PSRAge = &age
	}

	// FRN 3 (bit 6): SSR age
	if (fspec1 & 0x20) != 0 {
		if buf.Len() < 1 {
			return bytesRead, fmt.Errorf("buffer too short for SSR age")
		}
		data := make([]byte, 1)
		n, err := buf.Read(data)
		if err != nil || n != 1 {
			return bytesRead + n, fmt.Errorf("reading SSR age: %w", err)
		}
		bytesRead += n
		s.rawData = append(s.rawData, data...)

		age := float64(data[0]) * 0.25 // LSB = 1/4 second
		s.SSRAge = &age
	}

	// FRN 4 (bit 5): Mode S age
	if (fspec1 & 0x10) != 0 {
		if buf.Len() < 1 {
			return bytesRead, fmt.Errorf("buffer too short for Mode S age")
		}
		data := make([]byte, 1)
		n, err := buf.Read(data)
		if err != nil || n != 1 {
			return bytesRead + n, fmt.Errorf("reading Mode S age: %w", err)
		}
		bytesRead += n
		s.rawData = append(s.rawData, data...)

		age := float64(data[0]) * 0.25 // LSB = 1/4 second
		s.ModeS_Age = &age
	}

	// FRN 5 (bit 4): ADS-C age
	if (fspec1 & 0x08) != 0 {
		if buf.Len() < 2 {
			return bytesRead, fmt.Errorf("buffer too short for ADS-C age")
		}
		data := make([]byte, 2)
		n, err := buf.Read(data)
		if err != nil || n != 2 {
			return bytesRead + n, fmt.Errorf("reading ADS-C age: %w", err)
		}
		bytesRead += n
		s.rawData = append(s.rawData, data...)

		// ADS-C age is a 16-bit value (2 bytes)
		age := float64(uint16(data[0])<<8|uint16(data[1])) * 0.25 // LSB = 1/4 second
		s.ADSC_Age = &age
	}

	// FRN 6 (bit 3): ADS-B Extended Squitter age
	if (fspec1 & 0x04) != 0 {
		if buf.Len() < 1 {
			return bytesRead, fmt.Errorf("buffer too short for ADS-B ES age")
		}
		data := make([]byte, 1)
		n, err := buf.Read(data)
		if err != nil || n != 1 {
			return bytesRead + n, fmt.Errorf("reading ADS-B ES age: %w", err)
		}
		bytesRead += n
		s.rawData = append(s.rawData, data...)

		age := float64(data[0]) * 0.25 // LSB = 1/4 second
		s.ADSB_ES_Age = &age
	}

	// FRN 7 (bit 2): ADS-B VDL Mode 4 age
	if (fspec1 & 0x02) != 0 {
		if buf.Len() < 1 {
			return bytesRead, fmt.Errorf("buffer too short for ADS-B VDL age")
		}
		data := make([]byte, 1)
		n, err := buf.Read(data)
		if err != nil || n != 1 {
			return bytesRead + n, fmt.Errorf("reading ADS-B VDL age: %w", err)
		}
		bytesRead += n
		s.rawData = append(s.rawData, data...)

		age := float64(data[0]) * 0.25 // LSB = 1/4 second
		s.ADSB_VDL_Age = &age
	}

	// Process the second FSPEC byte if present
	if hasSecondFSPEC {
		// FRN 8 (bit 8 of second byte): ADS-B UAT age
		if (fspec2 & 0x80) != 0 {
			if buf.Len() < 1 {
				return bytesRead, fmt.Errorf("buffer too short for ADS-B UAT age")
			}
			data := make([]byte, 1)
			n, err := buf.Read(data)
			if err != nil || n != 1 {
				return bytesRead + n, fmt.Errorf("reading ADS-B UAT age: %w", err)
			}
			bytesRead += n
			s.rawData = append(s.rawData, data...)

			age := float64(data[0]) * 0.25 // LSB = 1/4 second
			s.ADSB_UAT_Age = &age
		}

		// FRN 9 (bit 7 of second byte): Loop age
		if (fspec2 & 0x40) != 0 {
			if buf.Len() < 1 {
				return bytesRead, fmt.Errorf("buffer too short for Loop age")
			}
			data := make([]byte, 1)
			n, err := buf.Read(data)
			if err != nil || n != 1 {
				return bytesRead + n, fmt.Errorf("reading Loop age: %w", err)
			}
			bytesRead += n
			s.rawData = append(s.rawData, data...)

			age := float64(data[0]) * 0.25 // LSB = 1/4 second
			s.LoopAge = &age
		}

		// FRN 10 (bit 6 of second byte): Multilateration age
		if (fspec2 & 0x20) != 0 {
			if buf.Len() < 1 {
				return bytesRead, fmt.Errorf("buffer too short for MLT age")
			}
			data := make([]byte, 1)
			n, err := buf.Read(data)
			if err != nil || n != 1 {
				return bytesRead + n, fmt.Errorf("reading MLT age: %w", err)
			}
			bytesRead += n
			s.rawData = append(s.rawData, data...)

			age := float64(data[0]) * 0.25 // LSB = 1/4 second
			s.MLTAge = &age
		}

		// The rest of the second FSPEC byte (bits 5-2) are spare bits
		// And bit 1 is the FX bit, which we would process if it's set to 1
		// But according to spec there are no more extensions, so we ignore it
	}

	return bytesRead, nil
}

// Encode serializes the System Track Update Ages into the buffer
func (s *SystemTrackUpdateAges) Encode(buf *bytes.Buffer) (int, error) {
	// If we have raw data, just send it back
	if len(s.rawData) > 0 {
		return buf.Write(s.rawData)
	}

	// We need to build the FSPEC based on which fields are present
	bytesWritten := 0

	// Determine which fields are present
	hasTrackAge := s.TrackAge != nil
	hasPSRAge := s.PSRAge != nil
	hasSSRAge := s.SSRAge != nil
	hasModeS := s.ModeS_Age != nil
	hasADSC := s.ADSC_Age != nil
	hasADSB_ES := s.ADSB_ES_Age != nil
	hasADSB_VDL := s.ADSB_VDL_Age != nil
	hasADSB_UAT := s.ADSB_UAT_Age != nil
	hasLoop := s.LoopAge != nil
	hasMLT := s.MLTAge != nil

	// Need second FSPEC byte?
	needSecondByte := hasADSB_UAT || hasLoop || hasMLT

	// First FSPEC byte
	fspec1 := byte(0)
	if hasTrackAge {
		fspec1 |= 0x80 // bit 8: Track age
	}
	if hasPSRAge {
		fspec1 |= 0x40 // bit 7: PSR age
	}
	if hasSSRAge {
		fspec1 |= 0x20 // bit 6: SSR age
	}
	if hasModeS {
		fspec1 |= 0x10 // bit 5: Mode S age
	}
	if hasADSC {
		fspec1 |= 0x08 // bit 4: ADS-C age
	}
	if hasADSB_ES {
		fspec1 |= 0x04 // bit 3: ADS-B ES age
	}
	if hasADSB_VDL {
		fspec1 |= 0x02 // bit 2: ADS-B VDL age
	}
	if needSecondByte {
		fspec1 |= 0x01 // bit 1: FX
	}

	// Write first FSPEC byte
	err := buf.WriteByte(fspec1)
	if err != nil {
		return bytesWritten, fmt.Errorf("writing first FSPEC byte: %w", err)
	}
	bytesWritten++

	// Second FSPEC byte if needed
	if needSecondByte {
		fspec2 := byte(0)
		if hasADSB_UAT {
			fspec2 |= 0x80 // bit 8: ADS-B UAT age
		}
		if hasLoop {
			fspec2 |= 0x40 // bit 7: Loop age
		}
		if hasMLT {
			fspec2 |= 0x20 // bit 6: MLT age
		}
		// No extension
		// fspec2 |= 0x00 // bit 1: FX = 0

		// Write second FSPEC byte
		err := buf.WriteByte(fspec2)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing second FSPEC byte: %w", err)
		}
		bytesWritten++
	}

	// Write the data for each present field

	// FRN 1: Track age
	if hasTrackAge {
		// Convert to 1/4 seconds, max value is 63.75s (255 * 0.25)
		val := uint8(min(*s.TrackAge/0.25, 255))
		err := buf.WriteByte(val)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing track age: %w", err)
		}
		bytesWritten++
	}

	// FRN 2: PSR age
	if hasPSRAge {
		val := uint8(min(*s.PSRAge/0.25, 255))
		err := buf.WriteByte(val)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing PSR age: %w", err)
		}
		bytesWritten++
	}

	// FRN 3: SSR age
	if hasSSRAge {
		val := uint8(min(*s.SSRAge/0.25, 255))
		err := buf.WriteByte(val)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing SSR age: %w", err)
		}
		bytesWritten++
	}

	// FRN 4: Mode S age
	if hasModeS {
		val := uint8(min(*s.ModeS_Age/0.25, 255))
		err := buf.WriteByte(val)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing Mode S age: %w", err)
		}
		bytesWritten++
	}

	// FRN 5: ADS-C age (16-bit value)
	if hasADSC {
		// Max value is 16383.75s (65535 * 0.25)
		val := uint16(min(*s.ADSC_Age/0.25, 65535))
		data := []byte{byte(val >> 8), byte(val)}
		n, err := buf.Write(data)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing ADS-C age: %w", err)
		}
		bytesWritten += n
	}

	// FRN 6: ADS-B ES age
	if hasADSB_ES {
		val := uint8(min(*s.ADSB_ES_Age/0.25, 255))
		err := buf.WriteByte(val)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing ADS-B ES age: %w", err)
		}
		bytesWritten++
	}

	// FRN 7: ADS-B VDL age
	if hasADSB_VDL {
		val := uint8(min(*s.ADSB_VDL_Age/0.25, 255))
		err := buf.WriteByte(val)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing ADS-B VDL age: %w", err)
		}
		bytesWritten++
	}

	// FRN 8: ADS-B UAT age
	if hasADSB_UAT {
		val := uint8(min(*s.ADSB_UAT_Age/0.25, 255))
		err := buf.WriteByte(val)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing ADS-B UAT age: %w", err)
		}
		bytesWritten++
	}

	// FRN 9: Loop age
	if hasLoop {
		val := uint8(min(*s.LoopAge/0.25, 255))
		err := buf.WriteByte(val)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing Loop age: %w", err)
		}
		bytesWritten++
	}

	// FRN 10: MLT age
	if hasMLT {
		val := uint8(min(*s.MLTAge/0.25, 255))
		err := buf.WriteByte(val)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing MLT age: %w", err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

// String returns a human-readable representation of the System Track Update Ages
func (s *SystemTrackUpdateAges) String() string {
	parts := []string{}

	if s.TrackAge != nil {
		parts = append(parts, fmt.Sprintf("TRK: %.2fs", *s.TrackAge))
	}
	if s.PSRAge != nil {
		parts = append(parts, fmt.Sprintf("PSR: %.2fs", *s.PSRAge))
	}
	if s.SSRAge != nil {
		parts = append(parts, fmt.Sprintf("SSR: %.2fs", *s.SSRAge))
	}
	if s.ModeS_Age != nil {
		parts = append(parts, fmt.Sprintf("MDS: %.2fs", *s.ModeS_Age))
	}
	if s.ADSC_Age != nil {
		parts = append(parts, fmt.Sprintf("ADS-C: %.2fs", *s.ADSC_Age))
	}
	if s.ADSB_ES_Age != nil {
		parts = append(parts, fmt.Sprintf("ADS-B ES: %.2fs", *s.ADSB_ES_Age))
	}
	if s.ADSB_VDL_Age != nil {
		parts = append(parts, fmt.Sprintf("ADS-B VDL: %.2fs", *s.ADSB_VDL_Age))
	}
	if s.ADSB_UAT_Age != nil {
		parts = append(parts, fmt.Sprintf("ADS-B UAT: %.2fs", *s.ADSB_UAT_Age))
	}
	if s.LoopAge != nil {
		parts = append(parts, fmt.Sprintf("Loop: %.2fs", *s.LoopAge))
	}
	if s.MLTAge != nil {
		parts = append(parts, fmt.Sprintf("MLT: %.2fs", *s.MLTAge))
	}

	if len(parts) == 0 {
		return "SystemTrackUpdateAges[empty]"
	}

	return fmt.Sprintf("SystemTrackUpdateAges[%s]", strings.Join(parts, ", "))
}

// Validate performs validation on the System Track Update Ages
func (s *SystemTrackUpdateAges) Validate() error {
	// Check that ages are within valid ranges
	if s.TrackAge != nil && (*s.TrackAge < 0 || *s.TrackAge > 63.75) {
		return fmt.Errorf("track age out of range [0,63.75]: %.2f", *s.TrackAge)
	}
	if s.PSRAge != nil && (*s.PSRAge < 0 || *s.PSRAge > 63.75) {
		return fmt.Errorf("PSR age out of range [0,63.75]: %.2f", *s.PSRAge)
	}
	if s.SSRAge != nil && (*s.SSRAge < 0 || *s.SSRAge > 63.75) {
		return fmt.Errorf("SSR age out of range [0,63.75]: %.2f", *s.SSRAge)
	}
	if s.ModeS_Age != nil && (*s.ModeS_Age < 0 || *s.ModeS_Age > 63.75) {
		return fmt.Errorf("Mode S age out of range [0,63.75]: %.2f", *s.ModeS_Age)
	}
	if s.ADSC_Age != nil && (*s.ADSC_Age < 0 || *s.ADSC_Age > 16383.75) {
		return fmt.Errorf("ADS-C age out of range [0,16383.75]: %.2f", *s.ADSC_Age)
	}
	if s.ADSB_ES_Age != nil && (*s.ADSB_ES_Age < 0 || *s.ADSB_ES_Age > 63.75) {
		return fmt.Errorf("ADS-B ES age out of range [0,63.75]: %.2f", *s.ADSB_ES_Age)
	}
	if s.ADSB_VDL_Age != nil && (*s.ADSB_VDL_Age < 0 || *s.ADSB_VDL_Age > 63.75) {
		return fmt.Errorf("ADS-B VDL age out of range [0,63.75]: %.2f", *s.ADSB_VDL_Age)
	}
	if s.ADSB_UAT_Age != nil && (*s.ADSB_UAT_Age < 0 || *s.ADSB_UAT_Age > 63.75) {
		return fmt.Errorf("ADS-B UAT age out of range [0,63.75]: %.2f", *s.ADSB_UAT_Age)
	}
	if s.LoopAge != nil && (*s.LoopAge < 0 || *s.LoopAge > 63.75) {
		return fmt.Errorf("Loop age out of range [0,63.75]: %.2f", *s.LoopAge)
	}
	if s.MLTAge != nil && (*s.MLTAge < 0 || *s.MLTAge > 63.75) {
		return fmt.Errorf("MLT age out of range [0,63.75]: %.2f", *s.MLTAge)
	}

	return nil
}
