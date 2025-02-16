// asterix/errors.go
package asterix

import "fmt"

// Core ASTERIX errors
var (
	ErrInvalidMessage  = fmt.Errorf("invalid ASTERIX message")
	ErrInvalidLength   = fmt.Errorf("invalid length")
	ErrInvalidFSPEC    = fmt.Errorf("invalid FSPEC")
	ErrMandatoryField  = fmt.Errorf("mandatory field missing")
	ErrInvalidCategory = fmt.Errorf("invalid category")
	ErrUnknownDataItem = fmt.Errorf("unknown data item")
	ErrInvalidField    = fmt.Errorf("invalid field value")
	ErrUAPNotDefined   = fmt.Errorf("UAP not defined for category")
	ErrFRNOutOfRange   = fmt.Errorf("FRN out of range")
	ErrBufferTooShort  = fmt.Errorf("buffer too short")
	ErrInvalidDataType = fmt.Errorf("invalid data type")
	ErrUnknownCategory = fmt.Errorf("unknown category")
)

// ValidationError provides detailed context for validation failures
type ValidationError struct {
	DataItem string
	Field    string
	Value    any
	Reason   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error in %s.%s: %v - %s",
		e.DataItem, e.Field, e.Value, e.Reason)
}

func (e *ValidationError) Unwrap() error {
	return ErrInvalidField
}
