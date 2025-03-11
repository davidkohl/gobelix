// asterix/validation.go
package asterix

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// MessageValidator provides utilities for validating ASTERIX messages
type MessageValidator struct {
	// Configuration
	MaxFSPECExtensions int    // Maximum allowed FSPEC extensions
	MaxMessageSize     uint16 // Maximum allowed message size
}

// NewMessageValidator creates a new MessageValidator with default settings
func NewMessageValidator() *MessageValidator {
	return &MessageValidator{
		MaxFSPECExtensions: 8,     // Conservative - most valid messages use 1-2
		MaxMessageSize:     16384, // 16K should be enough for any valid message
	}
}

// ValidateMessageStructure performs basic structural validation of an ASTERIX message
func (v *MessageValidator) ValidateMessageStructure(data []byte) (Category, error) {
	// Check minimum length
	if len(data) < 3 {
		return 0, fmt.Errorf("%w: message too short (%d bytes)",
			ErrInvalidMessage, len(data))
	}

	// Extract category
	cat := Category(data[0])
	if !cat.IsValid() {
		return cat, fmt.Errorf("%w: %d", ErrInvalidCategory, cat)
	}

	// Validate length field
	length := binary.BigEndian.Uint16(data[1:3])
	if length < 3 {
		return cat, fmt.Errorf("%w: too small (%d)", ErrInvalidLength, length)
	}

	if length > v.MaxMessageSize {
		return cat, fmt.Errorf("%w: exceeds maximum allowed size (%d > %d)",
			ErrInvalidLength, length, v.MaxMessageSize)
	}

	if int(length) != len(data) {
		return cat, fmt.Errorf("%w: declared %d, actual %d",
			ErrInvalidLength, length, len(data))
	}

	// Must have at least one FSPEC byte
	if len(data) < 4 {
		return cat, fmt.Errorf("%w: no space for FSPEC", ErrInvalidMessage)
	}

	// Validate FSPEC structure
	return cat, v.validateFSPEC(data[3:])
}

// validateFSPEC checks if the FSPEC structure is valid
func (v *MessageValidator) validateFSPEC(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("%w: empty FSPEC", ErrInvalidFSPEC)
	}

	// Track FSPEC bytes
	var fspecBytes []byte
	hasDataBits := false

	// Process extension chain
	for i := 0; i < len(data); i++ {
		// Prevent excessive FSPEC chains
		if i >= v.MaxFSPECExtensions {
			return fmt.Errorf("%w: too many extension bytes (%d)",
				ErrInvalidFSPEC, i)
		}

		fspecByte := data[i]
		fspecBytes = append(fspecBytes, fspecByte)

		// Check if any data bits are set (bits 7-1, excluding FX bit)
		if fspecByte&0xFE != 0 {
			hasDataBits = true
		}

		// Check if extension bit is set
		if fspecByte&0x01 == 0 {
			// No more extensions
			break
		}

		// Ensure we have another byte for extension
		if i+1 >= len(data) {
			return fmt.Errorf("%w: truncated after extension bit",
				ErrInvalidFSPEC)
		}
	}

	// Ensure at least one data bit is set
	if !hasDataBits {
		return fmt.Errorf("%w: no data bits set", ErrInvalidFSPEC)
	}

	return nil
}

// CheckBuffer verifies that a buffer has enough bytes remaining
func CheckBuffer(buf *bytes.Buffer, needed int, itemName string) error {
	if buf.Len() < needed {
		return fmt.Errorf("%w for %s: need %d bytes, have %d",
			ErrBufferTooShort, itemName, needed, buf.Len())
	}
	return nil
}

// AnalyzeMessage provides detailed analysis of an ASTERIX message
// Useful for debugging and diagnostic purposes
func AnalyzeMessage(data []byte) map[string]interface{} {
	result := make(map[string]interface{})

	// Basic properties
	if len(data) < 3 {
		result["valid"] = false
		result["error"] = "message too short"
		return result
	}

	result["category"] = int(data[0])
	result["length"] = binary.BigEndian.Uint16(data[1:3])
	result["actual_length"] = len(data)

	// FSPEC analysis if present
	if len(data) > 3 {
		fspecInfo := make(map[string]interface{})
		fspecBytes := make([]byte, 0)

		// Extract FSPEC bytes
		hasExtension := true
		for i := 3; i < len(data) && hasExtension; i++ {
			if i >= len(data) {
				break
			}

			fspecByte := data[i]
			fspecBytes = append(fspecBytes, fspecByte)

			// Check extension bit
			hasExtension = (fspecByte & 0x01) != 0
		}

		fspecInfo["bytes"] = fspecBytes
		fspecInfo["byte_count"] = len(fspecBytes)

		// Count data bits
		dataBitCount := 0
		for _, b := range fspecBytes {
			// Check bits 7-1 (excluding FX bit)
			for j := 0; j < 7; j++ {
				if b&(0x80>>j) != 0 {
					dataBitCount++
				}
			}
		}

		fspecInfo["data_bit_count"] = dataBitCount

		// Check if last byte has extension
		if len(fspecBytes) > 0 {
			fspecInfo["has_truncated_extension"] =
				(fspecBytes[len(fspecBytes)-1] & 0x01) != 0
		}

		result["fspec"] = fspecInfo
	}

	// Overall assessment
	isValid := true
	var validationErrors []string

	// Check length consistency
	if len(data) < 3 || int(result["length"].(uint16)) != len(data) {
		isValid = false
		validationErrors = append(validationErrors, "length mismatch")
	}

	// Check FSPEC validity
	if fspec, ok := result["fspec"].(map[string]interface{}); ok {
		if fspec["data_bit_count"].(int) == 0 {
			isValid = false
			validationErrors = append(validationErrors, "FSPEC has no data bits")
		}

		if fspec["has_truncated_extension"].(bool) {
			isValid = false
			validationErrors = append(validationErrors, "truncated FSPEC extension")
		}
	}

	result["valid"] = isValid
	if len(validationErrors) > 0 {
		result["validation_errors"] = validationErrors
	}

	return result
}
