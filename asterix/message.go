// asterix/message.go
package asterix

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// DecodedMessage represents a complete decoded ASTERIX message
type AsterixMessage struct {
	Category   Category              // ASTERIX category (e.g. Cat062)
	records    []map[string]DataItem // Decoded records within this message
	RawMessage []byte                // Original binary data
	Timestamp  time.Time             // When the message was decoded
	Source     string                // Optional source identifier
	uap        UAP
}

// NewAsterixMessage creates a new empty ASTERIX message of the specified category
func NewAsterixMessage(category Category) *AsterixMessage {
	return &AsterixMessage{
		Category:  category,
		records:   make([]map[string]DataItem, 0),
		Timestamp: time.Now(),
	}
}

// AddRecord adds a new record to the message
func (m *AsterixMessage) AddRecord(record map[string]DataItem) {
	m.records = append(m.records, record)
	// Mark RawMessage as potentially stale
	m.RawMessage = nil
}

func (m *AsterixMessage) GetDataItemFromRecord(id string, rid int) (DataItem, string, bool) {
	if rid > len(m.records) {
		return nil, "", false
	}

	targetRecord := m.records[rid]

	if val, ok := targetRecord[id]; ok {
		return val, fmt.Sprintf("%T", val), true
	}
	return nil, "", false
}

func (m *AsterixMessage) GetRecordCount() int {
	return len(m.records)
}

// String returns a string representation of the decoded message
func (m *AsterixMessage) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("ASTERIX %s Message (%d bytes, %d records)\n",
		m.Category.String(), len(m.RawMessage), len(m.records)))
	sb.WriteString(fmt.Sprintf("Timestamp: %s\n", m.Timestamp.Format(time.RFC3339Nano)))

	if m.Source != "" {
		sb.WriteString(fmt.Sprintf("Source: %s\n", m.Source))
	}

	// For each record
	for i, record := range m.records {
		sb.WriteString(fmt.Sprintf("Record #%d:\n", i+1))

		if m.uap != nil {
			fields := m.uap.Fields()

			// Sort fields by FRN
			sort.Slice(fields, func(i, j int) bool {
				return fields[i].FRN < fields[j].FRN
			})

			// Print data items in FRN order
			for _, field := range fields {
				if item, exists := record[field.DataItem]; exists {
					sb.WriteString(fmt.Sprintf("%2d  %s (%s): %v\n",
						field.FRN, field.DataItem, field.Description, item))
				}
			}
		} else {
			// Fallback if UAP not available - sort by data item ID
			keys := make([]string, 0, len(record))
			for k := range record {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				sb.WriteString(fmt.Sprintf("  %s: %v\n", k, record[k]))
			}
		}
	}

	return sb.String()
}
