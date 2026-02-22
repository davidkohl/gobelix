// cat/cat023/dataitems/v13/data_source_identifier.go
package v13

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// DataSourceIdentifier represents I023/010 - Data Source Identifier
// Fixed length: 2 bytes (SAC/SIC)
type DataSourceIdentifier struct {
	SAC uint8 // System Area Code
	SIC uint8 // System Identification Code
}

// Decode decodes the Data Source Identifier from bytes
func (d *DataSourceIdentifier) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("%w: reading data source identifier", asterix.ErrBufferTooShort)
	}
	if n != 2 {
		return n, fmt.Errorf("%w: need 2 bytes for data source identifier, got %d", asterix.ErrBufferTooShort, n)
	}

	d.SAC = data[0]
	d.SIC = data[1]

	return n, nil
}

// Encode encodes the Data Source Identifier to bytes
func (d *DataSourceIdentifier) Encode(buf *bytes.Buffer) (int, error) {
	data := []byte{d.SAC, d.SIC}
	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing data source identifier: %w", err)
	}
	return n, nil
}

// Validate validates the Data Source Identifier
func (d *DataSourceIdentifier) Validate() error {
	// No specific validation needed - all uint8 values are valid
	return nil
}

// String returns a string representation
func (d *DataSourceIdentifier) String() string {
	return fmt.Sprintf("SAC:%d SIC:%d", d.SAC, d.SIC)
}
