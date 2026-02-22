// cat/cat020/dataitems/v15/contributing_devices.go
package v15

import (
	"bytes"
	"fmt"
)

// ContributingDevices implements I020/400 - Contributing Devices
// This is a repetitive data item where each repetition is 1 octet
type ContributingDevices struct {
	Devices []uint8 // Each device is 1 byte
}

// Decode reads the repetitive device data
func (c *ContributingDevices) Decode(buf *bytes.Buffer) (int, error) {
	// Read REP byte
	rep, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading I020/400 REP: %w", err)
	}
	bytesRead := 1

	// Read repetitions (each is 1 byte)
	c.Devices = make([]uint8, rep)
	for i := 0; i < int(rep); i++ {
		device, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading I020/400 device %d: %w", i, err)
		}
		c.Devices[i] = device
		bytesRead++
	}

	return bytesRead, nil
}

// Encode writes the repetitive device data
func (c *ContributingDevices) Encode(buf *bytes.Buffer) (int, error) {
	if len(c.Devices) > 255 {
		return 0, fmt.Errorf("I020/400: too many devices (%d > 255)", len(c.Devices))
	}

	// Write REP byte
	err := buf.WriteByte(byte(len(c.Devices)))
	if err != nil {
		return 0, fmt.Errorf("writing I020/400 REP: %w", err)
	}
	bytesWritten := 1

	// Write each device
	for i, device := range c.Devices {
		err := buf.WriteByte(device)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing I020/400 device %d: %w", i, err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

// Validate implements the DataItem interface
func (c *ContributingDevices) Validate() error {
	return nil
}

// String returns a string representation
func (c *ContributingDevices) String() string {
	if len(c.Devices) == 0 {
		return "Contributing Devices: (none)"
	}
	return fmt.Sprintf("Contributing Devices: %d devices", len(c.Devices))
}
