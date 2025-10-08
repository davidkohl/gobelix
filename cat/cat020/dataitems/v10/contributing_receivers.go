// cat/cat020/dataitems/v10/contributing_receivers.go
package v10

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/davidkohl/gobelix/asterix"
)

// ContributingReceivers represents I020/400 - Contributing Receivers
// Repetitive data item: 1+1+ octets
// Overview of Receiver Units which have contributed to the Target Detection
type ContributingReceivers struct {
	Receivers []byte // Each byte represents 8 receiver units (bit set = contributed)
}

// NewContributingReceivers creates a new Contributing Receivers data item
func NewContributingReceivers() *ContributingReceivers {
	return &ContributingReceivers{}
}

// Decode decodes the Contributing Receivers from bytes
func (c *ContributingReceivers) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 1 {
		return 0, fmt.Errorf("%w: need at least 1 byte, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	bytesRead := 0

	// Read repetition factor
	rep := buf.Next(1)
	bytesRead++
	repCount := int(rep[0])

	if repCount == 0 {
		return bytesRead, nil
	}

	if buf.Len() < repCount {
		return bytesRead, fmt.Errorf("%w: need %d bytes for receivers, have %d", asterix.ErrBufferTooShort, repCount, buf.Len())
	}

	// Read receiver bytes
	c.Receivers = make([]byte, repCount)
	n, err := buf.Read(c.Receivers)
	bytesRead += n
	if err != nil {
		return bytesRead, fmt.Errorf("reading receiver bytes: %w", err)
	}

	return bytesRead, nil
}

// Encode encodes the Contributing Receivers to bytes
func (c *ContributingReceivers) Encode(buf *bytes.Buffer) (int, error) {
	if err := c.Validate(); err != nil {
		return 0, err
	}

	bytesWritten := 0

	// Write repetition factor
	repCount := byte(len(c.Receivers))
	if err := buf.WriteByte(repCount); err != nil {
		return bytesWritten, fmt.Errorf("writing repetition factor: %w", err)
	}
	bytesWritten++

	if repCount == 0 {
		return bytesWritten, nil
	}

	// Write receiver bytes
	n, err := buf.Write(c.Receivers)
	bytesWritten += n
	if err != nil {
		return bytesWritten, fmt.Errorf("writing receiver bytes: %w", err)
	}

	return bytesWritten, nil
}

// Validate validates the Contributing Receivers
func (c *ContributingReceivers) Validate() error {
	if len(c.Receivers) > 255 {
		return fmt.Errorf("%w: too many receiver bytes, max 255, got %d", asterix.ErrInvalidMessage, len(c.Receivers))
	}
	return nil
}

// String returns a string representation
func (c *ContributingReceivers) String() string {
	if len(c.Receivers) == 0 {
		return "No receivers"
	}

	var contributing []int
	for byteIdx, b := range c.Receivers {
		for bitIdx := 0; bitIdx < 8; bitIdx++ {
			if b&(1<<uint(7-bitIdx)) != 0 {
				receiverNum := byteIdx*8 + bitIdx + 1
				contributing = append(contributing, receiverNum)
			}
		}
	}

	if len(contributing) == 0 {
		return "No contributing receivers"
	}

	var parts []string
	for _, r := range contributing {
		parts = append(parts, fmt.Sprintf("RU%d", r))
	}

	return strings.Join(parts, ", ")
}
