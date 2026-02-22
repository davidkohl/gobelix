package v13

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// AlertMessages implements I011/600 - Alert Messages
// Definition: Alert involving the targets indicated in I011/605.
// Format: Three-octet fixed length Data Item.
type AlertMessages struct {
	ACK         bool  // Acknowledged (0=yes, 1=no)
	SVR         uint8 // Severity (0=end, 1=pre-alarm, 2=severe)
	AlertType   uint8 // Alert type code
	AlertNumber uint8 // Alert number
}

// Severity constants
const (
	AlertSeverityEnd      uint8 = 0 // End of alert
	AlertSeverityPreAlarm uint8 = 1 // Pre-alarm
	AlertSeveritySevere   uint8 = 2 // Severe alert
)

func (a *AlertMessages) Decode(buf *bytes.Buffer) (int, error) {
	var data [3]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading alert messages: %w", err)
	}
	if n != 3 {
		return n, fmt.Errorf("alert messages: expected 3 bytes, got %d", n)
	}

	a.ACK = (data[0] & 0x80) != 0
	a.SVR = (data[0] >> 5) & 0x03
	// bits 5-1 are spare
	a.AlertType = data[1]
	a.AlertNumber = data[2]

	return 3, nil
}

func (a *AlertMessages) Encode(buf *bytes.Buffer) (int, error) {
	var data [3]byte

	if a.ACK {
		data[0] |= 0x80
	}
	data[0] |= (a.SVR & 0x03) << 5
	data[1] = a.AlertType
	data[2] = a.AlertNumber

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing alert messages: %w", err)
	}
	return n, nil
}

func (a *AlertMessages) Validate() error {
	if a.SVR > 2 {
		return fmt.Errorf("severity out of range: %d (expected 0-2)", a.SVR)
	}
	return nil
}

func (a *AlertMessages) String() string {
	ack := "Acknowledged"
	if a.ACK {
		ack = "Not Acknowledged"
	}
	sev := ""
	switch a.SVR {
	case AlertSeverityEnd:
		sev = "End"
	case AlertSeverityPreAlarm:
		sev = "Pre-alarm"
	case AlertSeveritySevere:
		sev = "Severe"
	}
	return fmt.Sprintf("Alert: Type=%d, Number=%d, %s, %s", a.AlertType, a.AlertNumber, sev, ack)
}

// TracksInAlert implements I011/605 - Tracks in Alert
// Definition: List of track numbers of the targets concerned by the alert described in I011/600.
// Format: Repetitive Data Item starting with a one-octet Field Repetition Indicator (REP)
// followed by two-octet track numbers.
type TracksInAlert struct {
	TrackNumbers []uint16 // List of track numbers (0-4095 each)
}

func (t *TracksInAlert) Decode(buf *bytes.Buffer) (int, error) {
	rep, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading tracks in alert REP: %w", err)
	}
	bytesRead := 1

	t.TrackNumbers = make([]uint16, rep)
	for i := 0; i < int(rep); i++ {
		var data [2]byte
		n, err := buf.Read(data[:])
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("reading track number: %w", err)
		}

		raw := binary.BigEndian.Uint16(data[:])
		t.TrackNumbers[i] = raw & 0x0FFF
	}

	return bytesRead, nil
}

func (t *TracksInAlert) Encode(buf *bytes.Buffer) (int, error) {
	if err := buf.WriteByte(uint8(len(t.TrackNumbers))); err != nil {
		return 0, fmt.Errorf("writing tracks in alert REP: %w", err)
	}
	bytesWritten := 1

	for _, tn := range t.TrackNumbers {
		var data [2]byte
		binary.BigEndian.PutUint16(data[:], tn&0x0FFF)
		n, err := buf.Write(data[:])
		bytesWritten += n
		if err != nil {
			return bytesWritten, fmt.Errorf("writing track number: %w", err)
		}
	}

	return bytesWritten, nil
}

func (t *TracksInAlert) Validate() error {
	for _, tn := range t.TrackNumbers {
		if tn > 4095 {
			return fmt.Errorf("track number out of range: %d (max 4095)", tn)
		}
	}
	return nil
}

func (t *TracksInAlert) String() string {
	return fmt.Sprintf("Tracks in Alert: %v", t.TrackNumbers)
}

// HoldbarStatus implements I011/610 - Holdbar Status
// Definition: Status of up to sixteen banks of twelve indicators.
// Format: Repetitive Data Item starting with a one-octet Field Repetition Indicator (REP)
// followed by two-octet banks/indicators.
type HoldbarStatus struct {
	Banks []HoldbarBank
}

type HoldbarBank struct {
	BankNumber uint8  // Bank number (4 bits)
	Indicators uint16 // 12 indicator bits
}

func (h *HoldbarStatus) Decode(buf *bytes.Buffer) (int, error) {
	rep, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading holdbar status REP: %w", err)
	}
	bytesRead := 1

	h.Banks = make([]HoldbarBank, rep)
	for i := 0; i < int(rep); i++ {
		var data [2]byte
		n, err := buf.Read(data[:])
		bytesRead += n
		if err != nil {
			return bytesRead, fmt.Errorf("reading holdbar bank: %w", err)
		}

		raw := binary.BigEndian.Uint16(data[:])
		h.Banks[i].BankNumber = uint8((raw >> 12) & 0x0F)
		h.Banks[i].Indicators = raw & 0x0FFF
	}

	return bytesRead, nil
}

func (h *HoldbarStatus) Encode(buf *bytes.Buffer) (int, error) {
	if err := buf.WriteByte(uint8(len(h.Banks))); err != nil {
		return 0, fmt.Errorf("writing holdbar status REP: %w", err)
	}
	bytesWritten := 1

	for _, bank := range h.Banks {
		var data [2]byte
		raw := (uint16(bank.BankNumber) << 12) | (bank.Indicators & 0x0FFF)
		binary.BigEndian.PutUint16(data[:], raw)
		n, err := buf.Write(data[:])
		bytesWritten += n
		if err != nil {
			return bytesWritten, fmt.Errorf("writing holdbar bank: %w", err)
		}
	}

	return bytesWritten, nil
}

func (h *HoldbarStatus) Validate() error {
	for _, bank := range h.Banks {
		if bank.BankNumber > 15 {
			return fmt.Errorf("bank number out of range: %d (max 15)", bank.BankNumber)
		}
	}
	return nil
}

func (h *HoldbarStatus) String() string {
	parts := []string{}
	for _, bank := range h.Banks {
		parts = append(parts, fmt.Sprintf("Bank%d=%03X", bank.BankNumber, bank.Indicators))
	}
	return fmt.Sprintf("Holdbar Status: %v", parts)
}
