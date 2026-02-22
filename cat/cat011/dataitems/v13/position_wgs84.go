package v13

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

// PositionWGS84 implements I011/041 - Position in WGS-84 Co-ordinates
// Definition: Position of a target in WGS-84 Co-ordinates.
// Format: Eight-octet fixed length Data Item
// Latitude and Longitude are in two's complement, LSB = 180/2^31 degrees
type PositionWGS84 struct {
	Latitude  float64 // In degrees, range -90 to +90
	Longitude float64 // In degrees, range -180 to +180
}

const wgs84LSB = 180.0 / math.MaxInt32 // 180 / 2^31

func (p *PositionWGS84) Decode(buf *bytes.Buffer) (int, error) {
	var data [8]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading position WGS-84: %w", err)
	}
	if n != 8 {
		return n, fmt.Errorf("position WGS-84: expected 8 bytes, got %d", n)
	}

	// Latitude: bits 64-33 (first 4 bytes), two's complement
	latRaw := int32(binary.BigEndian.Uint32(data[0:4]))
	p.Latitude = float64(latRaw) * wgs84LSB

	// Longitude: bits 32-1 (last 4 bytes), two's complement
	lonRaw := int32(binary.BigEndian.Uint32(data[4:8]))
	p.Longitude = float64(lonRaw) * wgs84LSB

	return 8, nil
}

func (p *PositionWGS84) Encode(buf *bytes.Buffer) (int, error) {
	if err := p.Validate(); err != nil {
		return 0, err
	}

	var data [8]byte

	// Latitude to raw
	latRaw := int32(math.Round(p.Latitude / wgs84LSB))
	binary.BigEndian.PutUint32(data[0:4], uint32(latRaw))

	// Longitude to raw
	lonRaw := int32(math.Round(p.Longitude / wgs84LSB))
	binary.BigEndian.PutUint32(data[4:8], uint32(lonRaw))

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing position WGS-84: %w", err)
	}
	return n, nil
}

func (p *PositionWGS84) Validate() error {
	if p.Latitude < -90 || p.Latitude > 90 {
		return fmt.Errorf("latitude out of range: %f (expected -90 to 90)", p.Latitude)
	}
	if p.Longitude < -180 || p.Longitude >= 180 {
		return fmt.Errorf("longitude out of range: %f (expected -180 to <180)", p.Longitude)
	}
	return nil
}

func (p *PositionWGS84) String() string {
	latDir := "N"
	if p.Latitude < 0 {
		latDir = "S"
	}
	lonDir := "E"
	if p.Longitude < 0 {
		lonDir = "W"
	}
	return fmt.Sprintf("Position: %.6f°%s, %.6f°%s", math.Abs(p.Latitude), latDir, math.Abs(p.Longitude), lonDir)
}
