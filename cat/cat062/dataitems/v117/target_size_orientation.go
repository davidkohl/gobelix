// dataitems/cat062/target_size_orientation.go
package v117

import (
	"bytes"
	"fmt"
)

// TargetSizeOrientation implements I062/270
// Variable length data item comprising a first part of one octet,
// followed by one-octet extents as necessary
type TargetSizeOrientation struct {
	Data []byte
}

func (t *TargetSizeOrientation) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0
	t.Data = nil

	// Read first byte (length)
	firstByte := make([]byte, 1)
	n, err := buf.Read(firstByte)
	if err != nil {
		return n, fmt.Errorf("reading target size orientation first byte: %w", err)
	}
	bytesRead += n
	t.Data = append(t.Data, firstByte[0])

	// Check for extension
	hasExtension := (firstByte[0] & 0x01) != 0

	// Read first extension if present
	if hasExtension {
		extByte := make([]byte, 1)
		n, err := buf.Read(extByte)
		if err != nil {
			return bytesRead, fmt.Errorf("reading target size orientation first extension: %w", err)
		}
		bytesRead += n
		t.Data = append(t.Data, extByte[0])

		// Check if second extension exists
		hasExtension = (extByte[0] & 0x01) != 0

		// Read second extension if present
		if hasExtension {
			extByte := make([]byte, 1)
			n, err := buf.Read(extByte)
			if err != nil {
				return bytesRead, fmt.Errorf("reading target size orientation second extension: %w", err)
			}
			bytesRead += n
			t.Data = append(t.Data, extByte[0])

			// There are no further extensions according to the spec
		}
	}

	return bytesRead, nil
}

func (t *TargetSizeOrientation) Encode(buf *bytes.Buffer) (int, error) {
	if len(t.Data) == 0 {
		// If no data, encode a minimal valid value
		return buf.Write([]byte{0})
	}
	return buf.Write(t.Data)
}

func (t *TargetSizeOrientation) String() string {
	return fmt.Sprintf("TargetSizeOrientation[%d bytes]", len(t.Data))
}

// TargetSizeOrientation Validate method
func (t *TargetSizeOrientation) Validate() error {
	// Stub implementation
	return nil
}
