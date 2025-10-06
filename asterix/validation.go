// asterix/validation.go
package asterix

import (
	"fmt"
	"math"
)

// ValidationResult represents the result of a validation operation
type ValidationResult struct {
	Valid   bool
	Errors  []error
	Details map[string]string
}

// NewValidationResult creates a new validation result
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Valid:   true,
		Errors:  make([]error, 0),
		Details: make(map[string]string),
	}
}

// AddError adds an error to the validation result
func (v *ValidationResult) AddError(err error) {
	v.Valid = false
	v.Errors = append(v.Errors, err)
}

// AddDetail adds a detail to the validation result
func (v *ValidationResult) AddDetail(key, value string) {
	v.Details[key] = value
}

// Error implements the error interface
func (v *ValidationResult) Error() string {
	if len(v.Errors) == 0 {
		return "no validation errors"
	}
	return fmt.Sprintf("%d validation errors: %v", len(v.Errors), v.Errors)
}

// Common validation functions for various types of data

// ValidateUint checks if a uint value is within the specified bounds
func ValidateUint(name string, value uint, min, max uint) error {
	if value < min || value > max {
		return NewValidationError(name, "value", value,
			fmt.Sprintf("out of range [%d, %d]", min, max))
	}
	return nil
}

// ValidateUint8 checks if a uint8 value is within the specified bounds
func ValidateUint8(name string, value uint8, min, max uint8) error {
	if value < min || value > max {
		return NewValidationError(name, "value", value,
			fmt.Sprintf("out of range [%d, %d]", min, max))
	}
	return nil
}

// ValidateUint16 checks if a uint16 value is within the specified bounds
func ValidateUint16(name string, value uint16, min, max uint16) error {
	if value < min || value > max {
		return NewValidationError(name, "value", value,
			fmt.Sprintf("out of range [%d, %d]", min, max))
	}
	return nil
}

// ValidateUint32 checks if a uint32 value is within the specified bounds
func ValidateUint32(name string, value uint32, min, max uint32) error {
	if value < min || value > max {
		return NewValidationError(name, "value", value,
			fmt.Sprintf("out of range [%d, %d]", min, max))
	}
	return nil
}

// ValidateInt checks if an int value is within the specified bounds
func ValidateInt(name string, value int, min, max int) error {
	if value < min || value > max {
		return NewValidationError(name, "value", value,
			fmt.Sprintf("out of range [%d, %d]", min, max))
	}
	return nil
}

// ValidateInt8 checks if an int8 value is within the specified bounds
func ValidateInt8(name string, value int8, min, max int8) error {
	if value < min || value > max {
		return NewValidationError(name, "value", value,
			fmt.Sprintf("out of range [%d, %d]", min, max))
	}
	return nil
}

// ValidateInt16 checks if an int16 value is within the specified bounds
func ValidateInt16(name string, value int16, min, max int16) error {
	if value < min || value > max {
		return NewValidationError(name, "value", value,
			fmt.Sprintf("out of range [%d, %d]", min, max))
	}
	return nil
}

// ValidateInt32 checks if an int32 value is within the specified bounds
func ValidateInt32(name string, value int32, min, max int32) error {
	if value < min || value > max {
		return NewValidationError(name, "value", value,
			fmt.Sprintf("out of range [%d, %d]", min, max))
	}
	return nil
}

// ValidateFloat32 checks if a float32 value is within the specified bounds
func ValidateFloat32(name string, value float32, min, max float32) error {
	if value < min || value > max || math.IsNaN(float64(value)) {
		return NewValidationError(name, "value", value,
			fmt.Sprintf("out of range [%f, %f] or NaN", min, max))
	}
	return nil
}

// ValidateFloat64 checks if a float64 value is within the specified bounds
func ValidateFloat64(name string, value float64, min, max float64) error {
	if value < min || value > max || math.IsNaN(value) {
		return NewValidationError(name, "value", value,
			fmt.Sprintf("out of range [%f, %f] or NaN", min, max))
	}
	return nil
}

// ValidateEnum checks if a value is one of the allowed values
func ValidateEnum(name string, value interface{}, allowedValues []interface{}) error {
	for _, allowed := range allowedValues {
		if value == allowed {
			return nil
		}
	}
	return NewValidationError(name, "value", value, "not an allowed value")
}

// ValidateString checks if a string value has a length within the specified bounds
func ValidateString(name string, value string, minLength, maxLength int) error {
	if len(value) < minLength || len(value) > maxLength {
		return NewValidationError(name, "length", len(value),
			fmt.Sprintf("out of range [%d, %d]", minLength, maxLength))
	}
	return nil
}

// ValidateSliceLength checks if a slice has a length within the specified bounds
func ValidateSliceLength(name string, value interface{}, minLength, maxLength int) error {
	var length int
	switch v := value.(type) {
	case []byte:
		length = len(v)
	case []int:
		length = len(v)
	case []uint:
		length = len(v)
	case []string:
		length = len(v)
	default:
		return NewValidationError(name, "type", fmt.Sprintf("%T", value),
			"not a slice type")
	}

	if length < minLength || length > maxLength {
		return NewValidationError(name, "length", length,
			fmt.Sprintf("out of range [%d, %d]", minLength, maxLength))
	}
	return nil
}

// ValidateLatitude checks if a latitude value is valid (-90 to 90 degrees)
func ValidateLatitude(name string, value float64) error {
	return ValidateFloat64(name, value, -90.0, 90.0)
}

// ValidateLongitude checks if a longitude value is valid (-180 to 180 degrees)
func ValidateLongitude(name string, value float64) error {
	return ValidateFloat64(name, value, -180.0, 180.0)
}

// ValidateAltitude checks if an altitude value is valid (in feet)
func ValidateAltitude(name string, value float64) error {
	// Generic range for altitude: -1500 feet (below sea level) to 100,000 feet
	return ValidateFloat64(name, value, -1500.0, 100000.0)
}

// ValidateHeading checks if a heading value is valid (0 to 360 degrees)
func ValidateHeading(name string, value float64) error {
	// Heading is typically 0 to 360 degrees
	if value < 0.0 {
		// Normalize negative values
		value += 360.0 * (1.0 + math.Floor(-value/360.0))
	}
	if value >= 360.0 {
		// Normalize values >= 360
		value -= 360.0 * math.Floor(value/360.0)
	}
	return ValidateFloat64(name, value, 0.0, 360.0)
}

// ValidateSpeed checks if a speed value is valid (in knots)
func ValidateSpeed(name string, value float64) error {
	// Generic range for speed: 0 to Mach 3 (approximately 2000 knots)
	return ValidateFloat64(name, value, 0.0, 2000.0)
}

// ValidateVerticalRate checks if a vertical rate value is valid (in feet per minute)
func ValidateVerticalRate(name string, value float64) error {
	// Generic range for vertical rate: -8000 to +8000 feet per minute
	return ValidateFloat64(name, value, -8000.0, 8000.0)
}

// ValidateDataSourceIdentifier validates a Data Source Identifier
func ValidateDataSourceIdentifier(name string, sac, sic uint8) error {
	if err := ValidateUint8(name+".SAC", sac, 1, 255); err != nil {
		return err
	}
	if err := ValidateUint8(name+".SIC", sic, 1, 255); err != nil {
		return err
	}
	return nil
}

// ValidateTimeOfDay validates a Time of Day value (in seconds since midnight)
func ValidateTimeOfDay(name string, seconds float64) error {
	// Valid range: 0 to 86400 seconds (24 hours)
	return ValidateFloat64(name, seconds, 0.0, 86400.0)
}

// ValidateTargetAddress validates a 24-bit ICAO aircraft address
func ValidateTargetAddress(name string, address uint32) error {
	// Valid range: 0x000001 to 0xFFFFFF (24 bits)
	return ValidateUint32(name, address, 1, 0xFFFFFF)
}

// ValidateTrackNumber validates a track number
func ValidateTrackNumber(name string, trackNumber uint16) error {
	// Valid range depends on the system, but a generic range is 0 to 4095
	return ValidateUint16(name, trackNumber, 0, 4095)
}

// ValidateCallsign validates an aircraft callsign
func ValidateCallsign(name string, callsign string) error {
	// Callsigns are typically 2-7 characters
	return ValidateString(name, callsign, 2, 7)
}

// ValidateFieldPresence checks if all required fields are present
// DEPRECATED: This function has a design flaw. Use ValidateRequiredFields instead.
// The map should contain ONLY required fields as keys, with true=present, false=missing
func ValidateFieldPresence(requiredFields map[string]bool) error {
	for field, present := range requiredFields {
		if !present {
			return fmt.Errorf("%w: %s", ErrMandatoryField, field)
		}
	}
	return nil
}

// ValidateRequiredFields checks if all fields in the required list are present
// requiredFields: list of field names that must be present
// presentFields: map of field name -> presence status
func ValidateRequiredFields(requiredFields []string, presentFields map[string]bool) error {
	for _, field := range requiredFields {
		if !presentFields[field] {
			return fmt.Errorf("%w: %s", ErrMandatoryField, field)
		}
	}
	return nil
}

// ValidateConditionalField checks if a field is present when a condition is true
func ValidateConditionalField(condition bool, fieldPresent bool, fieldName string) error {
	if condition && !fieldPresent {
		return fmt.Errorf("%w: %s required when condition is true", ErrMandatoryField, fieldName)
	}
	return nil
}

// ValidateBitField validates a bit field value
func ValidateBitField(name string, value uint, validBits uint) error {
	if value > (1<<validBits)-1 {
		return NewValidationError(name, "value", value,
			fmt.Sprintf("exceeds maximum bit field value for %d bits", validBits))
	}
	return nil
}

// MultiErr combines multiple errors into a single error
type MultiErr struct {
	Errors []error
}

// NewMultiErr creates a new multi-error
func NewMultiErr() *MultiErr {
	return &MultiErr{
		Errors: make([]error, 0),
	}
}

// Add adds an error to the multi-error
func (m *MultiErr) Add(err error) {
	if err != nil {
		m.Errors = append(m.Errors, err)
	}
}

// Error implements the error interface
func (m *MultiErr) Error() string {
	if len(m.Errors) == 0 {
		return "no errors"
	}

	if len(m.Errors) == 1 {
		return m.Errors[0].Error()
	}

	result := fmt.Sprintf("%d errors: [", len(m.Errors))
	for i, err := range m.Errors {
		if i > 0 {
			result += ", "
		}
		result += err.Error()
	}
	result += "]"
	return result
}

// HasErrors returns true if there are any errors
func (m *MultiErr) HasErrors() bool {
	return len(m.Errors) > 0
}

// AsError returns nil if no errors, otherwise returns the multi-error
func (m *MultiErr) AsError() error {
	if !m.HasErrors() {
		return nil
	}
	return m
}
