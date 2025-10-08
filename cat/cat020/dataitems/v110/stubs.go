// cat/cat020/dataitems/v110/stubs.go
// Placeholder implementations for Cat020 data items
// TODO: Implement full decoding/encoding for each item
package v110

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// Simple stub types for data items that need full implementation

type TargetReportDescriptor struct{ data []byte }
type PositionWGS84 struct{ Latitude, Longitude float64 }
type PositionCartesian struct{ X, Y float64 }
type TrackStatus struct{ data []byte }
type CalculatedTrackVelocity struct{ Vx, Vy int16 }
type FlightLevel struct{ FlightLevel int16 }
type ModeCCode struct{ V, G, Code uint16 }
type TargetAddress struct{ Address uint32 }
type TargetIdentification struct{ Callsign string }
type MeasuredHeight struct{ Height int16 }
type GeometricHeight struct{ Height int16 }
type CalculatedAcceleration struct{ Ax, Ay int8 }
type VehicleFleetIdentification struct{ Fleet uint8 }
type PreprogrammedMessage struct{ TRB, MSG uint8 }
type PositionAccuracy struct{ data []byte }
type ContributingDevices struct{ data []byte }
type BDSRegisterData struct{ data []byte }
type CommunicationsACAS struct{ data []byte }
type ACASResolutionAdvisory struct{ data []byte }
type WarningErrorConditions struct{ data []byte }
type Mode1Code struct{ Code uint8 }
type Mode2Code struct{ Code uint16 }

// Constructor functions
func NewTargetReportDescriptor() *TargetReportDescriptor { return &TargetReportDescriptor{} }
func NewPositionWGS84() *PositionWGS84 { return &PositionWGS84{} }
func NewPositionCartesian() *PositionCartesian { return &PositionCartesian{} }
func NewTrackStatus() *TrackStatus { return &TrackStatus{} }
func NewCalculatedTrackVelocity() *CalculatedTrackVelocity { return &CalculatedTrackVelocity{} }
func NewFlightLevel() *FlightLevel { return &FlightLevel{} }
func NewModeCCode() *ModeCCode { return &ModeCCode{} }
func NewTargetAddress() *TargetAddress { return &TargetAddress{} }
func NewTargetIdentification() *TargetIdentification { return &TargetIdentification{} }
func NewMeasuredHeight() *MeasuredHeight { return &MeasuredHeight{} }
func NewGeometricHeight() *GeometricHeight { return &GeometricHeight{} }
func NewCalculatedAcceleration() *CalculatedAcceleration { return &CalculatedAcceleration{} }
func NewVehicleFleetIdentification() *VehicleFleetIdentification { return &VehicleFleetIdentification{} }
func NewPreprogrammedMessage() *PreprogrammedMessage { return &PreprogrammedMessage{} }
func NewPositionAccuracy() *PositionAccuracy { return &PositionAccuracy{} }
func NewContributingDevices() *ContributingDevices { return &ContributingDevices{} }
func NewBDSRegisterData() *BDSRegisterData { return &BDSRegisterData{} }
func NewCommunicationsACAS() *CommunicationsACAS { return &CommunicationsACAS{} }
func NewACASResolutionAdvisory() *ACASResolutionAdvisory { return &ACASResolutionAdvisory{} }
func NewWarningErrorConditions() *WarningErrorConditions { return &WarningErrorConditions{} }
func NewMode1Code() *Mode1Code { return &Mode1Code{} }
func NewMode2Code() *Mode2Code { return &Mode2Code{} }

// Stub implementations - read the expected number of bytes
func (t *TargetReportDescriptor) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 1 { return 0, asterix.ErrBufferTooShort }
	b := buf.Next(1)[0]
	t.data = []byte{b}
	if b&0x01 != 0 && buf.Len() > 0 { // FX bit set
		b2 := buf.Next(1)[0]
		t.data = append(t.data, b2)
		return 2, nil
	}
	return 1, nil
}
func (t *TargetReportDescriptor) Encode(buf *bytes.Buffer) (int, error) {
	n, _ := buf.Write(t.data)
	return n, nil
}
func (t *TargetReportDescriptor) Validate() error { return nil }
func (t *TargetReportDescriptor) String() string { return fmt.Sprintf("%x", t.data) }

func (p *PositionWGS84) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 8 { return 0, asterix.ErrBufferTooShort }
	data := buf.Next(8)
	lat := int32(data[0])<<24 | int32(data[1])<<16 | int32(data[2])<<8 | int32(data[3])
	lon := int32(data[4])<<24 | int32(data[5])<<16 | int32(data[6])<<8 | int32(data[7])
	// LSB = 180/2^25 degrees per spec §5.2.4
	// For 32-bit signed: value * 180 / 2^31 = value * 180 / 2147483648
	p.Latitude = float64(lat) * 180.0 / 2147483648.0
	p.Longitude = float64(lon) * 180.0 / 2147483648.0
	return 8, nil
}
func (p *PositionWGS84) Encode(buf *bytes.Buffer) (int, error) {
	lat := int32(p.Latitude * 2147483648.0 / 180.0)
	lon := int32(p.Longitude * 2147483648.0 / 180.0)
	data := []byte{
		byte(lat >> 24), byte(lat >> 16), byte(lat >> 8), byte(lat),
		byte(lon >> 24), byte(lon >> 16), byte(lon >> 8), byte(lon),
	}
	n, _ := buf.Write(data)
	return n, nil
}
func (p *PositionWGS84) Validate() error { return nil }
func (p *PositionWGS84) String() string { return fmt.Sprintf("%.6f°, %.6f°", p.Latitude, p.Longitude) }

func (p *PositionCartesian) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 6 { return 0, asterix.ErrBufferTooShort }
	data := buf.Next(6)
	x := int32(data[0])<<16 | int32(data[1])<<8 | int32(data[2])
	if x&0x800000 != 0 { x |= ^0xFFFFFF } // Sign extend
	y := int32(data[3])<<16 | int32(data[4])<<8 | int32(data[5])
	if y&0x800000 != 0 { y |= ^0xFFFFFF }
	p.X = float64(x) * 0.5  // LSB = 0.5 m per spec §5.2.5
	p.Y = float64(y) * 0.5
	return 6, nil
}
func (p *PositionCartesian) Encode(buf *bytes.Buffer) (int, error) {
	x := int32(p.X * 2.0) & 0xFFFFFF  // LSB = 0.5 m, so divide by 0.5 (multiply by 2)
	y := int32(p.Y * 2.0) & 0xFFFFFF
	data := []byte{
		byte(x >> 16), byte(x >> 8), byte(x),
		byte(y >> 16), byte(y >> 8), byte(y),
	}
	n, _ := buf.Write(data)
	return n, nil
}
func (p *PositionCartesian) Validate() error { return nil }
func (p *PositionCartesian) String() string { return fmt.Sprintf("X=%.2fm, Y=%.2fm", p.X, p.Y) }

// Add more stub implementations for remaining items
func (t *TrackStatus) Decode(buf *bytes.Buffer) (int, error) {
	// I020/170 is Extended type: read octets until FX bit is 0
	if buf.Len() < 1 { return 0, asterix.ErrBufferTooShort }
	t.data = []byte{}
	bytesRead := 0
	for {
		if buf.Len() < 1 { return bytesRead, asterix.ErrBufferTooShort }
		b := buf.Next(1)[0]
		t.data = append(t.data, b)
		bytesRead++
		if b&0x01 == 0 { // FX bit is 0, end of extensions
			break
		}
	}
	return bytesRead, nil
}
func (t *TrackStatus) Encode(buf *bytes.Buffer) (int, error) { n, _ := buf.Write(t.data); return n, nil }
func (t *TrackStatus) Validate() error { return nil }
func (t *TrackStatus) String() string { return fmt.Sprintf("%x", t.data) }

func (v *CalculatedTrackVelocity) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 4 { return 0, asterix.ErrBufferTooShort }
	data := buf.Next(4)
	v.Vx = int16(data[0])<<8 | int16(data[1])
	v.Vy = int16(data[2])<<8 | int16(data[3])
	return 4, nil
}
func (v *CalculatedTrackVelocity) Encode(buf *bytes.Buffer) (int, error) {
	data := []byte{byte(v.Vx >> 8), byte(v.Vx), byte(v.Vy >> 8), byte(v.Vy)}
	n, _ := buf.Write(data)
	return n, nil
}
func (v *CalculatedTrackVelocity) Validate() error { return nil }
func (v *CalculatedTrackVelocity) String() string { return fmt.Sprintf("Vx=%d, Vy=%d", v.Vx, v.Vy) }

// Add minimal stubs for remaining items to make it compile
func (f *FlightLevel) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 2 { return 0, asterix.ErrBufferTooShort }
	data := buf.Next(2)
	f.FlightLevel = int16(data[0])<<8 | int16(data[1])
	return 2, nil
}
func (f *FlightLevel) Encode(buf *bytes.Buffer) (int, error) {
	data := []byte{byte(f.FlightLevel >> 8), byte(f.FlightLevel)}
	n, _ := buf.Write(data)
	return n, nil
}
func (f *FlightLevel) Validate() error { return nil }
func (f *FlightLevel) String() string { return fmt.Sprintf("FL%d", f.FlightLevel/4) }

// Implement remaining stubs with minimal functionality
func decodeFixed(buf *bytes.Buffer, size int) ([]byte, error) {
	if buf.Len() < size { return nil, asterix.ErrBufferTooShort }
	return buf.Next(size), nil
}

func (m *ModeCCode) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 4 { return 0, asterix.ErrBufferTooShort }
	data := buf.Next(4)
	// Octet 1: V, G, spare bits
	m.V = uint16(data[0]>>7) & 0x01
	m.G = uint16(data[0]>>6) & 0x01
	// Octets 1-2: Mode-C reply in Gray notation (bits 28-17)
	// Octets 3-4: Quality bits (bits 12-1)
	m.Code = uint16(data[0]&0x3F)<<8 | uint16(data[1])
	return 4, nil
}
func (m *ModeCCode) Encode(buf *bytes.Buffer) (int, error) {
	data := []byte{
		byte(m.V<<7) | byte(m.G<<6) | byte((m.Code>>8)&0x3F),
		byte(m.Code),
		0x00, 0x00, // Quality bits (not implemented)
	}
	buf.Write(data)
	return 4, nil
}
func (m *ModeCCode) Validate() error { return nil }
func (m *ModeCCode) String() string { return fmt.Sprintf("ModeC V=%d G=%d Code=%04o", m.V, m.G, m.Code) }

func (t *TargetAddress) Decode(buf *bytes.Buffer) (int, error) { data, err := decodeFixed(buf, 3); if err == nil { t.Address = uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2]) }; return 3, err }
func (t *TargetAddress) Encode(buf *bytes.Buffer) (int, error) { buf.Write([]byte{byte(t.Address >> 16), byte(t.Address >> 8), byte(t.Address)}); return 3, nil }
func (t *TargetAddress) Validate() error { return nil }
func (t *TargetAddress) String() string { return fmt.Sprintf("%06X", t.Address) }

func (t *TargetIdentification) Decode(buf *bytes.Buffer) (int, error) { _, err := decodeFixed(buf, 7); return 7, err }
func (t *TargetIdentification) Encode(buf *bytes.Buffer) (int, error) { buf.Write(make([]byte, 7)); return 7, nil }
func (t *TargetIdentification) Validate() error { return nil }
func (t *TargetIdentification) String() string { return t.Callsign }

func (m *MeasuredHeight) Decode(buf *bytes.Buffer) (int, error) { _, err := decodeFixed(buf, 2); return 2, err }
func (m *MeasuredHeight) Encode(buf *bytes.Buffer) (int, error) { buf.Write([]byte{0, 0}); return 2, nil }
func (m *MeasuredHeight) Validate() error { return nil }
func (m *MeasuredHeight) String() string { return fmt.Sprintf("%dm", m.Height) }

func (g *GeometricHeight) Decode(buf *bytes.Buffer) (int, error) { _, err := decodeFixed(buf, 2); return 2, err }
func (g *GeometricHeight) Encode(buf *bytes.Buffer) (int, error) { buf.Write([]byte{0, 0}); return 2, nil }
func (g *GeometricHeight) Validate() error { return nil }
func (g *GeometricHeight) String() string { return fmt.Sprintf("%dm", g.Height) }

func (c *CalculatedAcceleration) Decode(buf *bytes.Buffer) (int, error) { _, err := decodeFixed(buf, 2); return 2, err }
func (c *CalculatedAcceleration) Encode(buf *bytes.Buffer) (int, error) { buf.Write([]byte{0, 0}); return 2, nil }
func (c *CalculatedAcceleration) Validate() error { return nil }
func (c *CalculatedAcceleration) String() string { return fmt.Sprintf("Ax=%d, Ay=%d", c.Ax, c.Ay) }

func (v *VehicleFleetIdentification) Decode(buf *bytes.Buffer) (int, error) { _, err := decodeFixed(buf, 1); return 1, err }
func (v *VehicleFleetIdentification) Encode(buf *bytes.Buffer) (int, error) { buf.WriteByte(v.Fleet); return 1, nil }
func (v *VehicleFleetIdentification) Validate() error { return nil }
func (v *VehicleFleetIdentification) String() string { return fmt.Sprintf("Fleet=%d", v.Fleet) }

func (p *PreprogrammedMessage) Decode(buf *bytes.Buffer) (int, error) { _, err := decodeFixed(buf, 1); return 1, err }
func (p *PreprogrammedMessage) Encode(buf *bytes.Buffer) (int, error) { buf.WriteByte(0); return 1, nil }
func (p *PreprogrammedMessage) Validate() error { return nil }
func (p *PreprogrammedMessage) String() string { return fmt.Sprintf("TRB=%d MSG=%d", p.TRB, p.MSG) }

func (p *PositionAccuracy) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 1 { return 0, asterix.ErrBufferTooShort }
	fspec := buf.Next(1)[0]
	p.data = []byte{fspec}
	count := 0
	for i := 7; i >= 1; i-- {
		if fspec&(1<<i) != 0 { count++ }
	}
	if buf.Len() < count { return 1, asterix.ErrBufferTooShort }
	p.data = append(p.data, buf.Next(count)...)
	return 1 + count, nil
}
func (p *PositionAccuracy) Encode(buf *bytes.Buffer) (int, error) { n, _ := buf.Write(p.data); return n, nil }
func (p *PositionAccuracy) Validate() error { return nil }
func (p *PositionAccuracy) String() string { return "PosAccuracy" }

func (c *ContributingDevices) Decode(buf *bytes.Buffer) (int, error) {
	// I020/400 is repetitive: 1 octet REP + (1 octet per device)
	if buf.Len() < 1 { return 0, asterix.ErrBufferTooShort }
	rep := buf.Next(1)[0]
	bytesNeeded := int(rep)
	if buf.Len() < bytesNeeded { return 1, asterix.ErrBufferTooShort }
	c.data = buf.Next(bytesNeeded)
	return 1 + bytesNeeded, nil
}
func (c *ContributingDevices) Encode(buf *bytes.Buffer) (int, error) { buf.WriteByte(byte(len(c.data))); buf.Write(c.data); return 1 + len(c.data), nil }
func (c *ContributingDevices) Validate() error { return nil }
func (c *ContributingDevices) String() string { return "ContribDevices" }

func (b *BDSRegisterData) Decode(buf *bytes.Buffer) (int, error) {
	// I020/250 is repetitive: 1 octet REP + (8 octets per BDS register)
	if buf.Len() < 1 { return 0, asterix.ErrBufferTooShort }
	rep := buf.Next(1)[0]
	bytesNeeded := int(rep) * 8
	if buf.Len() < bytesNeeded { return 1, asterix.ErrBufferTooShort }
	b.data = buf.Next(bytesNeeded)
	return 1 + bytesNeeded, nil
}
func (b *BDSRegisterData) Encode(buf *bytes.Buffer) (int, error) { buf.WriteByte(byte(len(b.data) / 8)); buf.Write(b.data); return 1 + len(b.data), nil }
func (b *BDSRegisterData) Validate() error { return nil }
func (b *BDSRegisterData) String() string { return "BDS" }

func (c *CommunicationsACAS) Decode(buf *bytes.Buffer) (int, error) { _, err := decodeFixed(buf, 2); return 2, err }
func (c *CommunicationsACAS) Encode(buf *bytes.Buffer) (int, error) { buf.Write([]byte{0, 0}); return 2, nil }
func (c *CommunicationsACAS) Validate() error { return nil }
func (c *CommunicationsACAS) String() string { return "COM/ACAS" }

func (a *ACASResolutionAdvisory) Decode(buf *bytes.Buffer) (int, error) { _, err := decodeFixed(buf, 7); return 7, err }
func (a *ACASResolutionAdvisory) Encode(buf *bytes.Buffer) (int, error) { buf.Write(make([]byte, 7)); return 7, nil }
func (a *ACASResolutionAdvisory) Validate() error { return nil }
func (a *ACASResolutionAdvisory) String() string { return "ACAS RA" }

func (w *WarningErrorConditions) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 1 { return 0, asterix.ErrBufferTooShort }
	b := buf.Next(1)[0]
	w.data = []byte{b}
	if b&0x01 != 0 && buf.Len() > 0 {
		b2 := buf.Next(1)[0]
		w.data = append(w.data, b2)
		return 2, nil
	}
	return 1, nil
}
func (w *WarningErrorConditions) Encode(buf *bytes.Buffer) (int, error) { n, _ := buf.Write(w.data); return n, nil }
func (w *WarningErrorConditions) Validate() error { return nil }
func (w *WarningErrorConditions) String() string { return "Warning/Error" }

func (m *Mode1Code) Decode(buf *bytes.Buffer) (int, error) { data, err := decodeFixed(buf, 1); if err == nil { m.Code = uint8(data[0]) & 0x1F }; return 1, err }
func (m *Mode1Code) Encode(buf *bytes.Buffer) (int, error) { buf.WriteByte(byte(m.Code) & 0x1F); return 1, nil }
func (m *Mode1Code) Validate() error { return nil }
func (m *Mode1Code) String() string { return fmt.Sprintf("Mode1: %o", m.Code) }

func (m *Mode2Code) Decode(buf *bytes.Buffer) (int, error) { data, err := decodeFixed(buf, 2); if err == nil { m.Code = uint16(data[0])<<8 | uint16(data[1]) }; return 2, err }
func (m *Mode2Code) Encode(buf *bytes.Buffer) (int, error) { buf.Write([]byte{byte(m.Code >> 8), byte(m.Code)}); return 2, nil }
func (m *Mode2Code) Validate() error { return nil }
func (m *Mode2Code) String() string { return fmt.Sprintf("Mode2: %o", m.Code) }
