package v10

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// SectorNumber represents I002/020 - Sector Number
type SectorNumber struct {
	SectorNumber float64 // Azimuth in degrees (LSB = 360/256 degrees)
}

func (s *SectorNumber) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 1 {
		return 0, fmt.Errorf("%w: need 1 byte for sector number, have %d", asterix.ErrBufferTooShort, buf.Len())
	}
	data := buf.Next(1)
	s.SectorNumber = float64(data[0]) * (360.0 / 256.0)
	return 1, nil
}

func (s *SectorNumber) Encode(buf *bytes.Buffer) (int, error) {
	sectorByte := uint8(s.SectorNumber * 256.0 / 360.0)
	buf.WriteByte(sectorByte)
	return 1, nil
}

func (s *SectorNumber) Validate() error {
	return nil
}

func (s *SectorNumber) String() string {
	return fmt.Sprintf("%.2f°", s.SectorNumber)
}
