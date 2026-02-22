package v13

import (
	"bytes"
	"fmt"
)

// MessageType implements I011/000 - Message Type
// Definition: This Data Item allows for a more convenient handling of the
// messages at the receiver side by further defining the type of transaction.
type MessageType struct {
	Type uint8 // Message Type (1-7)
}

// Message type constants
const (
	MessageTypeTargetReport          uint8 = 1 // Target reports, flight plan data and basic alerts
	MessageTypeManualAttachment      uint8 = 2 // Manual attachment of flight plan to track
	MessageTypeManualDetachment      uint8 = 3 // Manual detachment of flight plan to track
	MessageTypeFlightPlanInsertion   uint8 = 4 // Insertion of flight plan data
	MessageTypeFlightPlanSuppression uint8 = 5 // Suppression of flight plan data
	MessageTypeFlightPlanModification uint8 = 6 // Modification of flight plan data
	MessageTypeHoldbarStatus         uint8 = 7 // Holdbar status
)

func (m *MessageType) Decode(buf *bytes.Buffer) (int, error) {
	data, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading message type: %w", err)
	}

	m.Type = data
	return 1, nil
}

func (m *MessageType) Encode(buf *bytes.Buffer) (int, error) {
	if err := buf.WriteByte(m.Type); err != nil {
		return 0, fmt.Errorf("writing message type: %w", err)
	}
	return 1, nil
}

func (m *MessageType) Validate() error {
	if m.Type < 1 || m.Type > 7 {
		return fmt.Errorf("message type out of range: %d (expected 1-7)", m.Type)
	}
	return nil
}

func (m *MessageType) String() string {
	typeStr := ""
	switch m.Type {
	case MessageTypeTargetReport:
		typeStr = "Target Report"
	case MessageTypeManualAttachment:
		typeStr = "Manual Attachment"
	case MessageTypeManualDetachment:
		typeStr = "Manual Detachment"
	case MessageTypeFlightPlanInsertion:
		typeStr = "Flight Plan Insertion"
	case MessageTypeFlightPlanSuppression:
		typeStr = "Flight Plan Suppression"
	case MessageTypeFlightPlanModification:
		typeStr = "Flight Plan Modification"
	case MessageTypeHoldbarStatus:
		typeStr = "Holdbar Status"
	default:
		typeStr = "Unknown"
	}
	return fmt.Sprintf("Message Type: %s (%d)", typeStr, m.Type)
}
