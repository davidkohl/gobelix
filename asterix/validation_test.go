// asterix/validation_test.go
package asterix

import (
	"fmt"
	"math"
	"testing"
)

func TestValidationResult(t *testing.T) {
	// Create a new validation result
	vr := NewValidationResult()
	if !vr.Valid {
		t.Error("New validation result should be valid")
	}
	if len(vr.Errors) != 0 {
		t.Errorf("New validation result should have no errors, got %d", len(vr.Errors))
	}

	// Add an error and check validity
	vr.AddError(fmt.Errorf("test error"))
	if vr.Valid {
		t.Error("Validation result with error should not be valid")
	}
	if len(vr.Errors) != 1 {
		t.Errorf("Validation result should have 1 error, got %d", len(vr.Errors))
	}

	// Add a detail
	vr.AddDetail("test_key", "test_value")
	if val, ok := vr.Details["test_key"]; !ok || val != "test_value" {
		t.Errorf("Detail not added correctly, got %v", vr.Details)
	}

	// Test error message
	errMsg := vr.Error()
	if errMsg != "1 validation errors: [test error]" {
		t.Errorf("Error message incorrect, got %s", errMsg)
	}
}

func TestValidateUint(t *testing.T) {
	// Test valid case
	err := ValidateUint("test", 5, 1, 10)
	if err != nil {
		t.Errorf("ValidateUint should pass for 5 in range [1, 10], got %v", err)
	}

	// Test invalid case - below minimum
	err = ValidateUint("test", 0, 1, 10)
	if err == nil {
		t.Error("ValidateUint should fail for 0 in range [1, 10]")
	}
	if !IsValidationError(err) {
		t.Errorf("Error should be validation error, got %T", err)
	}

	// Test invalid case - above maximum
	err = ValidateUint("test", 11, 1, 10)
	if err == nil {
		t.Error("ValidateUint should fail for 11 in range [1, 10]")
	}
}

func TestValidateFloat64(t *testing.T) {
	// Test valid case
	err := ValidateFloat64("test", 5.5, 1.0, 10.0)
	if err != nil {
		t.Errorf("ValidateFloat64 should pass for 5.5 in range [1.0, 10.0], got %v", err)
	}

	// Test invalid case - below minimum
	err = ValidateFloat64("test", 0.5, 1.0, 10.0)
	if err == nil {
		t.Error("ValidateFloat64 should fail for 0.5 in range [1.0, 10.0]")
	}

	// Test invalid case - above maximum
	err = ValidateFloat64("test", 10.5, 1.0, 10.0)
	if err == nil {
		t.Error("ValidateFloat64 should fail for 10.5 in range [1.0, 10.0]")
	}

	// Test invalid case - NaN
	err = ValidateFloat64("test", math.NaN(), 1.0, 10.0)
	if err == nil {
		t.Error("ValidateFloat64 should fail for NaN")
	}
}

func TestValidateString(t *testing.T) {
	// Test valid case
	err := ValidateString("test", "hello", 3, 10)
	if err != nil {
		t.Errorf("ValidateString should pass for 'hello' with length range [3, 10], got %v", err)
	}

	// Test invalid case - too short
	err = ValidateString("test", "hi", 3, 10)
	if err == nil {
		t.Error("ValidateString should fail for 'hi' with length range [3, 10]")
	}

	// Test invalid case - too long
	err = ValidateString("test", "hello world!", 3, 10)
	if err == nil {
		t.Error("ValidateString should fail for 'hello world!' with length range [3, 10]")
	}
}

func TestValidateEnum(t *testing.T) {
	// Test valid case
	allowed := []interface{}{1, 2, 3, "test"}
	err := ValidateEnum("test", 2, allowed)
	if err != nil {
		t.Errorf("ValidateEnum should pass for 2 in [1, 2, 3, \"test\"], got %v", err)
	}

	// Test valid case - string
	err = ValidateEnum("test", "test", allowed)
	if err != nil {
		t.Errorf("ValidateEnum should pass for \"test\" in [1, 2, 3, \"test\"], got %v", err)
	}

	// Test invalid case
	err = ValidateEnum("test", 4, allowed)
	if err == nil {
		t.Error("ValidateEnum should fail for 4 in [1, 2, 3, \"test\"]")
	}
}

func TestValidateSliceLength(t *testing.T) {
	// Test valid case - byte slice
	err := ValidateSliceLength("test", []byte{1, 2, 3}, 1, 5)
	if err != nil {
		t.Errorf("ValidateSliceLength should pass for []byte{1, 2, 3} with length range [1, 5], got %v", err)
	}

	// Test valid case - int slice
	err = ValidateSliceLength("test", []int{1, 2, 3}, 1, 5)
	if err != nil {
		t.Errorf("ValidateSliceLength should pass for []int{1, 2, 3} with length range [1, 5], got %v", err)
	}

	// Test invalid case - too short
	err = ValidateSliceLength("test", []byte{}, 1, 5)
	if err == nil {
		t.Error("ValidateSliceLength should fail for [] with length range [1, 5]")
	}

	// Test invalid case - too long
	err = ValidateSliceLength("test", []int{1, 2, 3, 4, 5, 6}, 1, 5)
	if err == nil {
		t.Error("ValidateSliceLength should fail for [1, 2, 3, 4, 5, 6] with length range [1, 5]")
	}

	// Test invalid case - not a slice
	err = ValidateSliceLength("test", 123, 1, 5)
	if err == nil {
		t.Error("ValidateSliceLength should fail for non-slice type")
	}
}

func TestSpecializedValidations(t *testing.T) {
	t.Run("ValidateLatitude", func(t *testing.T) {
		// Test valid cases
		validValues := []float64{0.0, 90.0, -90.0, 45.123, -45.123}
		for _, val := range validValues {
			err := ValidateLatitude("lat", val)
			if err != nil {
				t.Errorf("ValidateLatitude should pass for %f, got %v", val, err)
			}
		}

		// Test invalid cases
		invalidValues := []float64{91.0, -91.0, 100.0, -100.0}
		for _, val := range invalidValues {
			err := ValidateLatitude("lat", val)
			if err == nil {
				t.Errorf("ValidateLatitude should fail for %f", val)
			}
		}
	})

	t.Run("ValidateLongitude", func(t *testing.T) {
		// Test valid cases
		validValues := []float64{0.0, 180.0, -180.0, 45.123, -45.123}
		for _, val := range validValues {
			err := ValidateLongitude("lon", val)
			if err != nil {
				t.Errorf("ValidateLongitude should pass for %f, got %v", val, err)
			}
		}

		// Test invalid cases
		invalidValues := []float64{181.0, -181.0, 200.0, -200.0}
		for _, val := range invalidValues {
			err := ValidateLongitude("lon", val)
			if err == nil {
				t.Errorf("ValidateLongitude should fail for %f", val)
			}
		}
	})

	t.Run("ValidateHeading", func(t *testing.T) {
		// Test valid cases
		validValues := []float64{0.0, 180.0, 359.9, 45.123}
		for _, val := range validValues {
			err := ValidateHeading("hdg", val)
			if err != nil {
				t.Errorf("ValidateHeading should pass for %f, got %v", val, err)
			}
		}

		// Heading normalization should handle values outside 0-360
		err := ValidateHeading("hdg", 370.0)
		if err != nil {
			t.Errorf("ValidateHeading should normalize 370.0 to 10.0, got %v", err)
		}

		err = ValidateHeading("hdg", -10.0)
		if err != nil {
			t.Errorf("ValidateHeading should normalize -10.0 to 350.0, got %v", err)
		}
	})
}

func TestDataSourceIdentifierValidation(t *testing.T) {
	// Test valid case
	err := ValidateDataSourceIdentifier("dsi", 1, 1)
	if err != nil {
		t.Errorf("ValidateDataSourceIdentifier should pass for SAC=1, SIC=1, got %v", err)
	}

	// Test invalid SAC
	err = ValidateDataSourceIdentifier("dsi", 0, 1)
	if err == nil {
		t.Error("ValidateDataSourceIdentifier should fail for SAC=0")
	}

	// Test invalid SIC
	err = ValidateDataSourceIdentifier("dsi", 1, 0)
	if err == nil {
		t.Error("ValidateDataSourceIdentifier should fail for SIC=0")
	}
}

func TestTimeOfDayValidation(t *testing.T) {
	// Test valid cases
	validTimes := []float64{0.0, 43200.0, 86399.9}
	for _, val := range validTimes {
		err := ValidateTimeOfDay("tod", val)
		if err != nil {
			t.Errorf("ValidateTimeOfDay should pass for %f, got %v", val, err)
		}
	}

	// Test invalid cases
	invalidTimes := []float64{-1.0, 86400.1, 90000.0}
	for _, val := range invalidTimes {
		err := ValidateTimeOfDay("tod", val)
		if err == nil {
			t.Errorf("ValidateTimeOfDay should fail for %f", val)
		}
	}
}

func TestValidateFieldPresence(t *testing.T) {
	// Test all required fields present
	fields := map[string]bool{
		"field1": true,  // Required and present
		"field2": true,  // Required and present
		"field3": false, // Not required
	}
	err := ValidateFieldPresence(fields)
	if err != nil {
		t.Errorf("ValidateFieldPresence should pass when all required fields are present, got %v", err)
	}

	// Test missing required field
	fields = map[string]bool{
		"field1": false, // Required but missing
		"field2": true,  // Required and present
		"field3": false, // Not required
	}
	err = ValidateFieldPresence(fields)
	if err == nil {
		t.Error("ValidateFieldPresence should fail when a required field is missing")
	}
	if !IsMandatoryFieldMissing(err) {
		t.Errorf("Error should be mandatory field missing, got %T", err)
	}
}

func TestValidateConditionalField(t *testing.T) {
	// Test condition true, field present
	err := ValidateConditionalField(true, true, "field1")
	if err != nil {
		t.Errorf("ValidateConditionalField should pass when condition is true and field is present, got %v", err)
	}

	// Test condition true, field not present
	err = ValidateConditionalField(true, false, "field1")
	if err == nil {
		t.Error("ValidateConditionalField should fail when condition is true and field is not present")
	}

	// Test condition false, field not present
	err = ValidateConditionalField(false, false, "field1")
	if err != nil {
		t.Errorf("ValidateConditionalField should pass when condition is false, got %v", err)
	}

	// Test condition false, field present
	err = ValidateConditionalField(false, true, "field1")
	if err != nil {
		t.Errorf("ValidateConditionalField should pass when condition is false, got %v", err)
	}
}

func TestMultiErr(t *testing.T) {
	// Create a new multi-error
	me := NewMultiErr()
	if me.HasErrors() {
		t.Error("New multi-error should have no errors")
	}

	// Test AsError() with no errors
	err := me.AsError()
	if err != nil {
		t.Errorf("AsError() should return nil for empty multi-error, got %v", err)
	}

	// Add an error
	me.Add(fmt.Errorf("error 1"))
	if !me.HasErrors() {
		t.Error("Multi-error should have errors after adding one")
	}
	if len(me.Errors) != 1 {
		t.Errorf("Multi-error should have 1 error, got %d", len(me.Errors))
	}

	// Test error message with one error
	errMsg := me.Error()
	if errMsg != "error 1" {
		t.Errorf("Error message incorrect for one error, got %s", errMsg)
	}

	// Add another error
	me.Add(fmt.Errorf("error 2"))
	if len(me.Errors) != 2 {
		t.Errorf("Multi-error should have 2 errors, got %d", len(me.Errors))
	}

	// Test error message with multiple errors
	errMsg = me.Error()
	if errMsg != "2 errors: [error 1, error 2]" {
		t.Errorf("Error message incorrect for multiple errors, got %s", errMsg)
	}

	// Test AsError() with errors
	err = me.AsError()
	if err == nil {
		t.Error("AsError() should not return nil for multi-error with errors")
	}
	if err.Error() != "2 errors: [error 1, error 2]" {
		t.Errorf("AsError() returned incorrect error message, got %s", err.Error())
	}

	// Test adding nil error
	me.Add(nil)
	if len(me.Errors) != 2 {
		t.Errorf("Adding nil error should not increase error count, got %d", len(me.Errors))
	}
}

func TestValidateBitField(t *testing.T) {
	// Test valid case - value within bit field range
	err := ValidateBitField("test", 15, 4) // 15 (1111) is valid for 4 bits
	if err != nil {
		t.Errorf("ValidateBitField should pass for 15 with 4 bits, got %v", err)
	}

	// Test invalid case - value exceeds bit field range
	err = ValidateBitField("test", 16, 4) // 16 (10000) is invalid for 4 bits
	if err == nil {
		t.Error("ValidateBitField should fail for 16 with 4 bits")
	}

	// Test boundary case - maximum value for bit field
	err = ValidateBitField("test", 255, 8) // 255 (11111111) is max value for 8 bits
	if err != nil {
		t.Errorf("ValidateBitField should pass for 255 with 8 bits, got %v", err)
	}

	// Test boundary case - value exceeds maximum for bit field
	err = ValidateBitField("test", 256, 8) // 256 (100000000) exceeds 8 bits
	if err == nil {
		t.Error("ValidateBitField should fail for 256 with 8 bits")
	}
}
