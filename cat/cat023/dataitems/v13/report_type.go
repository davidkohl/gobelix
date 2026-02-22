// cat/cat023/dataitems/v13/report_type.go
package v13

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// ReportType represents I023/000 - Report Type
// Fixed length: 1 byte
type ReportType struct {
	ReportType uint8 // 1=Ground Station Status, 2=Service Status, 3=Service Statistics
}

// Decode decodes the Report Type from bytes
func (r *ReportType) Decode(buf *bytes.Buffer) (int, error) {
	data, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("%w: need 1 byte for report type", asterix.ErrBufferTooShort)
	}

	r.ReportType = data

	return 1, nil
}

// Encode encodes the Report Type to bytes
func (r *ReportType) Encode(buf *bytes.Buffer) (int, error) {
	if err := r.Validate(); err != nil {
		return 0, err
	}

	if err := buf.WriteByte(r.ReportType); err != nil {
		return 0, fmt.Errorf("writing report type: %w", err)
	}

	return 1, nil
}

// Validate validates the Report Type
func (r *ReportType) Validate() error {
	if r.ReportType < 1 || r.ReportType > 3 {
		return fmt.Errorf("%w: report type must be 1-3, got %d", asterix.ErrInvalidMessage, r.ReportType)
	}
	return nil
}

// String returns a string representation
func (r *ReportType) String() string {
	types := map[uint8]string{
		1: "Ground Station Status",
		2: "Service Status",
		3: "Service Statistics",
	}
	if name, ok := types[r.ReportType]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(%d)", r.ReportType)
}
