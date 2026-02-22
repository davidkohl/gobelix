// dataitems/common/datasource.go
package common

import (
	"bytes"
	"fmt"
)

type DataSourceIdentifier struct {
	SAC uint8 // System Area Code
	SIC uint8 // System Identification Code
}

func (d *DataSourceIdentifier) Encode(buf *bytes.Buffer) (int, error) {
	if err := d.Validate(); err != nil {
		return 0, err
	}

	if err := buf.WriteByte(d.SAC); err != nil {
		return 0, fmt.Errorf("writing SAC: %w", err)
	}
	if err := buf.WriteByte(d.SIC); err != nil {
		return 1, fmt.Errorf("writing SIC: %w", err)
	}
	return 2, nil
}

func (d *DataSourceIdentifier) Decode(buf *bytes.Buffer) (int, error) {
	var err error
	bytesRead := 0

	d.SAC, err = buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading SAC: %w", err)
	}
	bytesRead++

	d.SIC, err = buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading SIC: %w", err)
	}
	bytesRead++

	return bytesRead, nil
}

func (d *DataSourceIdentifier) Validate() error {
	if d.SAC == 0 {
		return fmt.Errorf("SAC cannot be 0")
	}
	if d.SIC == 0 {
		return fmt.Errorf("SIC cannot be 0")
	}
	return nil
}

func (d *DataSourceIdentifier) String() string {
	return fmt.Sprintf("SAC: %d, SIC: %d", d.SAC, d.SIC)
}
